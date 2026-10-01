package finance

import (
	"context"
	"errors"
	"time"

	domainFinance "panda-pocket/internal/domain/finance"
)

var milestoneThresholds = []int{25, 50, 75, 100}

var ErrMilestoneNotReached = errors.New("milestone not reached")

type milestoneRow struct {
	Percent    int
	Celebrated bool
}

type milestonePlan struct {
	Insert       []int
	Delete       []int
	Uncelebrated []int
}

func planGoalMilestones(existing []milestoneRow, progress float64) milestonePlan {
	met := map[int]bool{}
	for _, threshold := range milestoneThresholds {
		if progress >= float64(threshold) {
			met[threshold] = true
		}
	}
	have := map[int]milestoneRow{}
	for _, row := range existing {
		have[row.Percent] = row
	}

	plan := milestonePlan{}
	for _, threshold := range milestoneThresholds {
		row, exists := have[threshold]
		switch {
		case met[threshold] && !exists:
			plan.Insert = append(plan.Insert, threshold)
			plan.Uncelebrated = append(plan.Uncelebrated, threshold)
		case met[threshold] && exists && !row.Celebrated:
			plan.Uncelebrated = append(plan.Uncelebrated, threshold)
		case !met[threshold] && exists:
			plan.Delete = append(plan.Delete, threshold)
		}
	}
	return plan
}

type GoalMilestoneSync struct {
	repo domainFinance.GoalMilestoneRepository
	now  func() time.Time
}

func NewGoalMilestoneSync(repo domainFinance.GoalMilestoneRepository) *GoalMilestoneSync {
	return &GoalMilestoneSync{repo: repo, now: time.Now}
}

func (s *GoalMilestoneSync) Sync(ctx context.Context, userID, goalID int, progress float64) ([]int, error) {
	if s == nil || s.repo == nil || goalID == 0 {
		return []int{}, nil
	}
	existing, err := s.repo.ListByGoal(ctx, goalID)
	if err != nil {
		return nil, err
	}
	plan := planGoalMilestones(toMilestoneRows(existing), progress)
	if len(plan.Delete) > 0 {
		if err := s.repo.DeletePercents(ctx, goalID, plan.Delete); err != nil {
			return nil, err
		}
	}
	reachedAt := s.now()
	for _, percent := range plan.Insert {
		if err := s.repo.Insert(ctx, userID, goalID, percent, reachedAt); err != nil {
			return nil, err
		}
	}
	if plan.Uncelebrated == nil {
		return []int{}, nil
	}
	return plan.Uncelebrated, nil
}

func (s *GoalMilestoneSync) Celebrate(ctx context.Context, userID, goalID, percent int) error {
	if !isMilestonePercent(percent) {
		return ErrMilestoneNotReached
	}
	rows, err := s.repo.ListByGoal(ctx, goalID)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.Percent != percent {
			continue
		}
		if row.Celebrated {
			return nil
		}
		return s.repo.MarkCelebrated(ctx, userID, goalID, percent, s.now())
	}
	return ErrMilestoneNotReached
}

type CelebrateGoalMilestoneUseCase struct {
	goals *domainFinance.GoalService
	sync  *GoalMilestoneSync
}

func NewCelebrateGoalMilestoneUseCase(
	goals *domainFinance.GoalService,
	sync *GoalMilestoneSync,
) *CelebrateGoalMilestoneUseCase {
	return &CelebrateGoalMilestoneUseCase{goals: goals, sync: sync}
}

func (uc *CelebrateGoalMilestoneUseCase) Execute(ctx context.Context, userID, goalID, percent int) error {
	if _, err := uc.goals.GetGoalForUser(ctx, domainFinance.NewUserID(userID), domainFinance.NewGoalID(goalID)); err != nil {
		return err
	}
	return uc.sync.Celebrate(ctx, userID, goalID, percent)
}

func toMilestoneRows(records []domainFinance.GoalMilestoneRecord) []milestoneRow {
	rows := make([]milestoneRow, 0, len(records))
	for _, record := range records {
		rows = append(rows, milestoneRow{Percent: record.Percent, Celebrated: record.Celebrated})
	}
	return rows
}

func isMilestonePercent(percent int) bool {
	for _, threshold := range milestoneThresholds {
		if threshold == percent {
			return true
		}
	}
	return false
}

type goalMilestoneSyncer interface {
	Sync(ctx context.Context, userID, goalID int, progress float64) ([]int, error)
}
