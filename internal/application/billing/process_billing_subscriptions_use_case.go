package billing

import (
	"context"
	"time"

	domainBilling "panda-pocket/internal/domain/billing"
	"panda-pocket/internal/infrastructure/analytics"
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
		wasActive := sub.Status() == domainBilling.StatusActive
		if !sub.ApplyBillingTransitions(now) {
			continue
		}
		if err := uc.repo.Save(ctx, sub); err != nil {
			return updated, err
		}
		if wasActive && sub.Status() == domainBilling.StatusPastDue {
			analytics.Capture(sub.UserID(), "payment_failed", map[string]interface{}{
				"provider":   "doit",
				"error_code": "period_unpaid",
			})
		}
		updated++
	}
	return updated, nil
}
