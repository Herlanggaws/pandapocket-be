package billing

import (
	"context"
	"testing"
	"time"

	domainBilling "panda-pocket/internal/domain/billing"
)

type memSubscriptionRepo struct {
	byUser map[int]*domainBilling.Subscription
	nextID int
}

func newMemSubscriptionRepo() *memSubscriptionRepo {
	return &memSubscriptionRepo{byUser: make(map[int]*domainBilling.Subscription), nextID: 1}
}

func (r *memSubscriptionRepo) Save(_ context.Context, sub *domainBilling.Subscription) error {
	if sub.ID().Value() == 0 {
		sub.AssignID(domainBilling.NewSubscriptionID(r.nextID))
		r.nextID++
	}
	r.byUser[sub.UserID()] = sub
	return nil
}

func (r *memSubscriptionRepo) FindByUserID(_ context.Context, userID int) (*domainBilling.Subscription, error) {
	sub, ok := r.byUser[userID]
	if !ok {
		return nil, domainBilling.ErrNotFound
	}
	return sub, nil
}

func TestGetSubscriptionCreatesFreeWhenMissing(t *testing.T) {
	repo := newMemSubscriptionRepo()
	uc := NewGetSubscriptionUseCase(repo)

	resp, err := uc.Execute(context.Background(), 42)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Plan != "free" || resp.Status != "expired" || resp.IsPro {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if _, err := repo.FindByUserID(context.Background(), 42); err != nil {
		t.Fatal("expected Free row to be persisted")
	}
}

func TestGetSubscriptionReturnsExisting(t *testing.T) {
	repo := newMemSubscriptionRepo()
	sub, err := domainBilling.NewFreeSubscription(7)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(context.Background(), sub); err != nil {
		t.Fatal(err)
	}

	uc := NewGetSubscriptionUseCase(repo)
	resp, err := uc.Execute(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Plan != "free" || resp.IsPro {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if len(repo.byUser) != 1 {
		t.Fatalf("expected no duplicate rows, got %d", len(repo.byUser))
	}
}

func TestGetSubscriptionNormalizesExpiredTrial(t *testing.T) {
	repo := newMemSubscriptionRepo()
	now := time.Now().UTC()
	past := now.Add(-time.Hour)
	sub := domainBilling.ReconstituteSubscription(
		domainBilling.NewSubscriptionID(0), 9, domainBilling.PlanFree, nil, domainBilling.StatusTrialing,
		&past, nil, nil, nil, domainBilling.CustomerRef(9), false, now, now,
	)
	if err := repo.Save(context.Background(), sub); err != nil {
		t.Fatal(err)
	}

	uc := NewGetSubscriptionUseCase(repo)
	resp, err := uc.Execute(context.Background(), 9)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != "expired" || resp.IsPro {
		t.Fatalf("expected expired free, got %+v", resp)
	}
	if resp.TrialEndsAt == nil {
		t.Fatal("trial_ends_at must remain after normalize")
	}
	stored, err := repo.FindByUserID(context.Background(), 9)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status() != domainBilling.StatusExpired {
		t.Fatalf("expected persisted expired, got %s", stored.Status())
	}
}

func TestGetSubscriptionKeepsActiveTrial(t *testing.T) {
	repo := newMemSubscriptionRepo()
	sub, err := domainBilling.NewTrialSubscription(11)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(context.Background(), sub); err != nil {
		t.Fatal(err)
	}

	uc := NewGetSubscriptionUseCase(repo)
	resp, err := uc.Execute(context.Background(), 11)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != "trialing" || !resp.IsPro || resp.TrialEndsAt == nil {
		t.Fatalf("expected active trial: %+v", resp)
	}
}
