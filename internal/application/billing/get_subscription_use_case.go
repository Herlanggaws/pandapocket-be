package billing

import (
	"context"
	"errors"
	"time"

	domainBilling "panda-pocket/internal/domain/billing"
)

type SubscriptionResponse struct {
	Plan              string  `json:"plan"`
	Status            string  `json:"status"`
	BillingInterval   *string `json:"billing_interval"`
	TrialEndsAt       *string `json:"trial_ends_at"`
	CurrentPeriodEnd  *string `json:"current_period_end"`
	GraceEndsAt       *string `json:"grace_ends_at"`
	CancelAtPeriodEnd bool    `json:"cancel_at_period_end"`
	IsPro             bool    `json:"is_pro"`
}

type GetSubscriptionUseCase struct {
	repo domainBilling.SubscriptionRepository
}

func NewGetSubscriptionUseCase(repo domainBilling.SubscriptionRepository) *GetSubscriptionUseCase {
	return &GetSubscriptionUseCase{repo: repo}
}

func (uc *GetSubscriptionUseCase) Execute(ctx context.Context, userID int) (*SubscriptionResponse, error) {
	now := time.Now().UTC()
	sub, err := uc.repo.FindByUserID(ctx, userID)
	if err != nil {
		if !errors.Is(err, domainBilling.ErrNotFound) {
			return nil, err
		}
		sub, err = domainBilling.NewFreeSubscription(userID)
		if err != nil {
			return nil, err
		}
		if err := uc.repo.Save(ctx, sub); err != nil {
			return nil, err
		}
		return toSubscriptionResponse(sub, now), nil
	}

	if sub.ExpireTrialIfNeeded(now) {
		if err := uc.repo.Save(ctx, sub); err != nil {
			return nil, err
		}
	}
	return toSubscriptionResponse(sub, now), nil
}

func toSubscriptionResponse(sub *domainBilling.Subscription, now time.Time) *SubscriptionResponse {
	var interval *string
	if sub.BillingInterval() != nil {
		value := sub.BillingInterval().String()
		interval = &value
	}

	return &SubscriptionResponse{
		Plan:              sub.Plan().String(),
		Status:            sub.Status().String(),
		BillingInterval:   interval,
		TrialEndsAt:       formatTimePtr(sub.TrialEndsAt()),
		CurrentPeriodEnd:  formatTimePtr(sub.CurrentPeriodEnd()),
		GraceEndsAt:       formatTimePtr(sub.GraceEndsAt()),
		CancelAtPeriodEnd: sub.CancelAtPeriodEnd(),
		IsPro:             sub.IsPro(now),
	}
}

func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	value := t.UTC().Format(time.RFC3339)
	return &value
}
