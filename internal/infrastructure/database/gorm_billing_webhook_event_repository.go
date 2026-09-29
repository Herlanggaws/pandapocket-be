package database

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GormBillingWebhookEventRepository struct {
	db *gorm.DB
}

func NewGormBillingWebhookEventRepository(db *gorm.DB) *GormBillingWebhookEventRepository {
	return &GormBillingWebhookEventRepository{db: db}
}

func (r *GormBillingWebhookEventRepository) Exists(ctx context.Context, eventID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&BillingWebhookEvent{}).Where("event_id = ?", eventID).Count(&count).Error
	return count > 0, err
}

// TryInsert stores the event. inserted is false when event_id already exists.
func (r *GormBillingWebhookEventRepository) TryInsert(ctx context.Context, eventID, payload string) (bool, error) {
	model := BillingWebhookEvent{
		EventID:    eventID,
		Payload:    payload,
		ReceivedAt: time.Now().UTC(),
	}
	result := r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&model)
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 {
		return false, nil
	}
	return true, nil
}
