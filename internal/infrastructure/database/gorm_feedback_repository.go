package database

import (
	"context"
	"panda-pocket/internal/domain/feedback"

	"gorm.io/gorm"
)

type GormFeedbackRepository struct {
	db *gorm.DB
}

func NewGormFeedbackRepository(db *gorm.DB) *GormFeedbackRepository {
	return &GormFeedbackRepository{db: db}
}

func (r *GormFeedbackRepository) Create(ctx context.Context, item *feedback.Feedback) error {
	model := &UserFeedback{
		UserID:   uint(item.UserID()),
		Category: item.Category().String(),
		Message:  item.Message(),
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	item.AssignID(feedback.NewFeedbackID(int(model.ID)))
	return nil
}
