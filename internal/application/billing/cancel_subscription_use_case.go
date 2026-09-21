package billing

import (
	"context"
	"errors"
	"time"

	domainBilling "panda-pocket/internal/domain/billing"
)

type CancelSubscriptionUseCase struct {
	repo domainBilling.SubscriptionRepository
}

func NewCancelSubscriptionUseCase(repo domainBilling.SubscriptionRepository) *CancelSubscriptionUseCase {
	return &CancelSubscriptionUseCase{repo: repo}
}

func (uc *CancelSubscriptionUseCase) Execute(ctx context.Context, userID int) (*SubscriptionResponse, error) {
	now := time.Now().UTC()
	sub, err := uc.repo.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, domainBilling.ErrNotFound) {
			return nil, errors.New("subscription not found")
		}
		return nil, err
	}
	if err := sub.ScheduleCancelAtPeriodEnd(now); err != nil {
		return nil, err
	}
	if err := uc.repo.Save(ctx, sub); err != nil {
		return nil, err
	}
	return toSubscriptionResponse(sub, now), nil
}
