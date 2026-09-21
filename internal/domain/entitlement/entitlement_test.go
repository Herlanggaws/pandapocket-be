package entitlement

import (
	"context"
	"errors"
	"testing"
	"time"

	"panda-pocket/internal/domain/billing"
)

type memSubRepo struct {
	byUser map[int]*billing.Subscription
}

func (r *memSubRepo) Save(_ context.Context, sub *billing.Subscription) error {
	r.byUser[sub.UserID()] = sub
	return nil
}

func (r *memSubRepo) FindByUserID(_ context.Context, userID int) (*billing.Subscription, error) {
	sub, ok := r.byUser[userID]
	if !ok {
		return nil, billing.ErrNotFound
	}
	return sub, nil
}

func (r *memSubRepo) ListAll(_ context.Context) ([]*billing.Subscription, error) {
	out := make([]*billing.Subscription, 0, len(r.byUser))
	for _, sub := range r.byUser {
		out = append(out, sub)
	}
	return out, nil
}

func TestCheckFreeLimit(t *testing.T) {
	if err := CheckFreeLimit(FeatureBudgets, 2, 3); err != nil {
		t.Fatalf("expected under limit ok, got %v", err)
	}
	err := CheckFreeLimit(FeatureBudgets, 3, 3)
	if err == nil {
		t.Fatal("expected limit exceeded")
	}
	var le *LimitExceeded
	if !errors.As(err, &le) {
		t.Fatalf("expected LimitExceeded, got %T", err)
	}
	if le.Feature != FeatureBudgets || le.Limit != 3 || le.Used != 3 {
		t.Fatalf("unexpected fields: %+v", le)
	}
	if !errors.Is(err, ErrPremiumRequired) {
		t.Fatal("expected errors.Is ErrPremiumRequired")
	}
}

func TestEnforceCreateLimit(t *testing.T) {
	ctx := context.Background()
	if err := EnforceCreateLimit(ctx, StaticChecker{Pro: true}, 1, FeatureTransactions, 100, FreeTransactionsPerMonth); err != nil {
		t.Fatalf("pro should skip limits: %v", err)
	}
	if err := EnforceCreateLimit(ctx, StaticChecker{Pro: false}, 1, FeatureTransactions, 49, FreeTransactionsPerMonth); err != nil {
		t.Fatalf("free under limit should pass: %v", err)
	}
	if err := EnforceCreateLimit(ctx, StaticChecker{Pro: false}, 1, FeatureTransactions, 50, FreeTransactionsPerMonth); err == nil {
		t.Fatal("free at limit should fail")
	}
	if err := EnforceCreateLimit(ctx, StaticChecker{Pro: false}, 1, FeatureRecurring, 0, 0); err == nil {
		t.Fatal("recurring free should fail")
	}
	if err := EnforceCreateLimit(ctx, StaticChecker{Pro: false}, 1, FeatureWallets, 1, FreeWallets); err == nil {
		t.Fatal("free at wallet limit should fail")
	}
	if err := EnforceCreateLimit(ctx, StaticChecker{Pro: false}, 1, FeatureWallets, 0, FreeWallets); err != nil {
		t.Fatalf("free under wallet limit should pass: %v", err)
	}
	if err := EnforceCreateLimit(ctx, StaticChecker{Pro: true}, 1, FeatureWallets, 5, FreeWallets); err != nil {
		t.Fatalf("pro should skip wallet limits: %v", err)
	}
	if err := EnforceCreateLimit(ctx, StaticChecker{Pro: false}, 1, FeatureAssets, 1, FreeAssets); err == nil {
		t.Fatal("free at assets limit should fail")
	}
	if err := EnforceCreateLimit(ctx, StaticChecker{Pro: false}, 1, FeatureDebts, 1, FreeDebts); err == nil {
		t.Fatal("free at debts limit should fail")
	}
	bypassCtx := WithEntitlementBypass(ctx)
	if err := EnforceCreateLimit(bypassCtx, StaticChecker{Pro: false}, 1, FeatureRecurring, 0, 0); err != nil {
		t.Fatalf("bypass should allow free recurring seed: %v", err)
	}
}

func TestSubscriptionChecker(t *testing.T) {
	now := time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC)
	future := now.Add(24 * time.Hour)
	repo := &memSubRepo{byUser: map[int]*billing.Subscription{}}
	checker := NewSubscriptionChecker(repo)
	checker.now = func() time.Time { return now }

	isPro, err := checker.IsPro(context.Background(), 99)
	if err != nil || isPro {
		t.Fatalf("missing sub should be free, got isPro=%v err=%v", isPro, err)
	}

	free, _ := billing.NewFreeSubscription(1)
	_ = repo.Save(context.Background(), free)
	isPro, err = checker.IsPro(context.Background(), 1)
	if err != nil || isPro {
		t.Fatalf("free sub should not be pro, got isPro=%v err=%v", isPro, err)
	}

	pro := billing.ReconstituteSubscription(
		billing.NewSubscriptionID(2), 2, billing.PlanPro, nil, billing.StatusActive,
		nil, &future, nil, nil, billing.CustomerRef(2), false, now, now,
	)
	_ = repo.Save(context.Background(), pro)
	isPro, err = checker.IsPro(context.Background(), 2)
	if err != nil || !isPro {
		t.Fatalf("active pro should be pro, got isPro=%v err=%v", isPro, err)
	}
}
