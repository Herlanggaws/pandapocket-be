package database

import (
	"context"
	"time"

	"panda-pocket/internal/domain/finance"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormGoalMilestoneRepository struct {
	db *gorm.DB
}

func NewGormGoalMilestoneRepository(db *gorm.DB) *GormGoalMilestoneRepository {
	return &GormGoalMilestoneRepository{db: db}
}

func (r *GormGoalMilestoneRepository) ListByGoal(ctx context.Context, goalID int) ([]finance.GoalMilestoneRecord, error) {
	var models []GoalMilestone
	if err := r.db.WithContext(ctx).Where("goal_id = ?", goalID).Order("percent ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	records := make([]finance.GoalMilestoneRecord, 0, len(models))
	for _, model := range models {
		records = append(records, finance.GoalMilestoneRecord{
			Percent:    model.Percent,
			Celebrated: model.CelebratedAt != nil,
		})
	}
	return records, nil
}

func (r *GormGoalMilestoneRepository) Insert(ctx context.Context, userID, goalID, percent int, reachedAt time.Time) error {
	model := GoalMilestone{
		UserID:    uint(userID),
		GoalID:    uint(goalID),
		Percent:   percent,
		ReachedAt: reachedAt,
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "goal_id"}, {Name: "percent"}},
		DoNothing: true,
	}).Create(&model).Error
}

func (r *GormGoalMilestoneRepository) DeletePercents(ctx context.Context, goalID int, percents []int) error {
	if len(percents) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Where("goal_id = ? AND percent IN ?", goalID, percents).
		Delete(&GoalMilestone{}).Error
}

func (r *GormGoalMilestoneRepository) MarkCelebrated(ctx context.Context, userID, goalID, percent int, at time.Time) error {
	return r.db.WithContext(ctx).Model(&GoalMilestone{}).
		Where("user_id = ? AND goal_id = ? AND percent = ? AND celebrated_at IS NULL", userID, goalID, percent).
		Updates(map[string]any{"celebrated_at": at, "updated_at": at}).Error
}
