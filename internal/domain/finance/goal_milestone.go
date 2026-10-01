package finance

import (
	"context"
	"time"
)

type GoalMilestoneRecord struct {
	Percent    int
	Celebrated bool
}

type GoalMilestoneRepository interface {
	ListByGoal(ctx context.Context, goalID int) ([]GoalMilestoneRecord, error)
	Insert(ctx context.Context, userID, goalID, percent int, reachedAt time.Time) error
	DeletePercents(ctx context.Context, goalID int, percents []int) error
	MarkCelebrated(ctx context.Context, userID, goalID, percent int, at time.Time) error
}
