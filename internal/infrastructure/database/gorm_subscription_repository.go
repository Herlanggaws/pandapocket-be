package database

import (
	"context"
	"errors"
	"panda-pocket/internal/domain/billing"

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
	}
	if sub.ID().Value() != 0 {
		model.ID = uint(sub.ID().Value())
	}
	return model
}

func (r *GormSubscriptionRepository) Save(ctx context.Context, sub *billing.Subscription) error {
	model := r.toModel(sub)
	if sub.ID().Value() != 0 {
		return r.db.WithContext(ctx).Save(model).Error
	}
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	sub.AssignID(billing.NewSubscriptionID(int(model.ID)))
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
