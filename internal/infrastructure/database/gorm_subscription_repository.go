package database

import (
	"context"
	"errors"
	"panda-pocket/internal/domain/billing"
	"time"

	"gorm.io/gorm"
)

type GormSubscriptionRepository struct {
	db *gorm.DB
}

func NewGormSubscriptionRepository(db *gorm.DB) *GormSubscriptionRepository {
	return &GormSubscriptionRepository{db: db}
}

func (r *GormSubscriptionRepository) toDomain(model Subscription) *billing.Subscription {
	var interval *billing.BillingInterval
	if model.BillingInterval != nil {
		value := billing.BillingInterval(*model.BillingInterval)
		interval = &value
	}

	return billing.ReconstituteSubscription(
		billing.NewSubscriptionID(int(model.ID)),
		int(model.UserID),
		billing.Plan(model.Plan),
		interval,
		billing.Status(model.Status),
		model.TrialEndsAt,
		model.CurrentPeriodEnd,
		model.GraceEndsAt,
		model.DoitSubscriptionID,
		model.DoitCustomerRef,
		model.CancelAtPeriodEnd,
		model.CreatedAt,
		model.UpdatedAt,
	)
}

func (r *GormSubscriptionRepository) toModel(sub *billing.Subscription) *Subscription {
	var interval *string
	if sub.BillingInterval() != nil {
		value := sub.BillingInterval().String()
		interval = &value
	}

	createdAt := sub.CreatedAt()
	if createdAt.IsZero() {
		createdAt = time.Now().UTC()
	}
	updatedAt := sub.UpdatedAt()
	if updatedAt.IsZero() {
		updatedAt = time.Now().UTC()
	}

	model := &Subscription{
		UserID:             uint(sub.UserID()),
		Plan:               sub.Plan().String(),
		BillingInterval:    interval,
		Status:             sub.Status().String(),
		TrialEndsAt:        sub.TrialEndsAt(),
		CurrentPeriodEnd:   sub.CurrentPeriodEnd(),
		GraceEndsAt:        sub.GraceEndsAt(),
		DoitSubscriptionID: sub.DoitSubscriptionID(),
		DoitCustomerRef:    sub.DoitCustomerRef(),
		CancelAtPeriodEnd:  sub.CancelAtPeriodEnd(),
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
	}
	if sub.ID().Value() != 0 {
		model.ID = uint(sub.ID().Value())
	}
	return model
}

func (r *GormSubscriptionRepository) Save(ctx context.Context, sub *billing.Subscription) error {
	model := r.toModel(sub)
	if sub.ID().Value() == 0 {
		if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
			return err
		}
		sub.AssignID(billing.NewSubscriptionID(int(model.ID)))
		return nil
	}

	// Targeted update so Save never overwrites created_at with a zero value.
	updates := map[string]interface{}{
		"user_id":              model.UserID,
		"plan":                 model.Plan,
		"billing_interval":     model.BillingInterval,
		"status":               model.Status,
		"trial_ends_at":        model.TrialEndsAt,
		"current_period_end":   model.CurrentPeriodEnd,
		"grace_ends_at":        model.GraceEndsAt,
		"doit_subscription_id": model.DoitSubscriptionID,
		"doit_customer_ref":    model.DoitCustomerRef,
		"cancel_at_period_end": model.CancelAtPeriodEnd,
		"updated_at":           time.Now().UTC(),
	}
	result := r.db.WithContext(ctx).Model(&Subscription{}).Where("id = ?", model.ID).Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormSubscriptionRepository) FindByUserID(ctx context.Context, userID int) (*billing.Subscription, error) {
	var model Subscription
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, billing.ErrNotFound
		}
		return nil, err
	}
	return r.toDomain(model), nil
}

func (r *GormSubscriptionRepository) ListAll(ctx context.Context) ([]*billing.Subscription, error) {
	var models []Subscription
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]*billing.Subscription, 0, len(models))
	for _, model := range models {
		result = append(result, r.toDomain(model))
	}
	return result, nil
}
