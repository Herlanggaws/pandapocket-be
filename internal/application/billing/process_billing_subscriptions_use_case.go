package billing

import (
	"context"
	"time"

	domainBilling "panda-pocket/internal/domain/billing"
)

type ProcessBillingSubscriptionsUseCase struct {
	repo domainBilling.SubscriptionRepository
}

func NewProcessBillingSubscriptionsUseCase(repo domainBilling.SubscriptionRepository) *ProcessBillingSubscriptionsUseCase {
	return &ProcessBillingSubscriptionsUseCase{repo: repo}
}

func (uc *ProcessBillingSubscriptionsUseCase) Execute(ctx context.Context) (int, error) {
	now := time.Now().UTC()
	subs, err := uc.repo.ListAll(ctx)
	if err != nil {
		return 0, err
	}
	updated := 0
	for _, sub := range subs {
		if !sub.ApplyBillingTransitions(now) {
			continue
		}
		if err := uc.repo.Save(ctx, sub); err != nil {
			return updated, err
		}
		updated++
	}
	return updated, nil
}
