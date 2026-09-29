package ai

import (
	"context"
	"testing"
	"time"

	domainAI "panda-pocket/internal/domain/ai"
	domainBilling "panda-pocket/internal/domain/billing"
)

type fixedSubs struct {
	sub *domainBilling.Subscription
}

func (f fixedSubs) Save(context.Context, *domainBilling.Subscription) error { return nil }

func (f fixedSubs) FindByUserID(context.Context, int) (*domainBilling.Subscription, error) {
	if f.sub == nil {
		return nil, domainBilling.ErrNotFound
	}
	return f.sub, nil
}

func (f fixedSubs) ListAll(context.Context) ([]*domainBilling.Subscription, error) {
	if f.sub == nil {
		return nil, nil
	}
	return []*domainBilling.Subscription{f.sub}, nil
}

func TestEnsureTrialKeepsCapAndUsage(t *testing.T) {
	sub, err := domainBilling.NewTrialSubscription(7)
	if err != nil {
		t.Fatal(err)
	}
	balance := domainAI.NewCreditBalance(7, true)
	if err := balance.Spend(); err != nil {
		t.Fatal(err)
	}
	balance.AddPurchase(20)
	repo := &memCredits{balance: balance}
	svc := NewCreditService(repo, fixedSubs{sub: sub})

	got, trialing, err := svc.Ensure(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if !trialing {
		t.Fatal("expected trialing")
	}
	if got.IncludedUnlocked != domainAI.TrialIncludedCap() {
		t.Fatalf("unlocked=%d", got.IncludedUnlocked)
	}
	if got.IncludedUsed != 1 || got.PurchasedRemaining != 20 {
		t.Fatalf("usage changed: used=%d purchased=%d", got.IncludedUsed, got.PurchasedRemaining)
	}

	again, _, err := svc.Ensure(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if again.IncludedUsed != 1 || again.PurchasedRemaining != 20 || again.IncludedUnlocked != domainAI.TrialIncludedCap() {
		t.Fatalf("second ensure reset balance: %+v", again)
	}
}

func TestEnsureActiveUnlocksGrantWithoutReset(t *testing.T) {
	sub, err := domainBilling.NewTrialSubscription(7)
	if err != nil {
		t.Fatal(err)
	}
	if err := sub.ActivatePro(domainBilling.IntervalMonthly, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	balance := domainAI.NewCreditBalance(7, true)
	for i := 0; i < 4; i++ {
		if err := balance.Spend(); err != nil {
			t.Fatal(err)
		}
	}
	balance.AddPurchase(50)
	repo := &memCredits{balance: balance}
	svc := NewCreditService(repo, fixedSubs{sub: sub})

	got, trialing, err := svc.Ensure(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if trialing {
		t.Fatal("active subscription must not count as trial")
	}
	if got.IncludedUnlocked != domainAI.IncludedGrant() {
		t.Fatalf("unlocked=%d", got.IncludedUnlocked)
	}
	if got.IncludedUsed != 4 || got.PurchasedRemaining != 50 {
		t.Fatalf("usage reset: used=%d purchased=%d", got.IncludedUsed, got.PurchasedRemaining)
	}
	if got.Available() != (domainAI.IncludedGrant()-4)+50 {
		t.Fatalf("available=%d", got.Available())
	}

	again, _, err := svc.Ensure(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if again.IncludedUnlocked != domainAI.IncludedGrant() || again.IncludedUsed != 4 || again.PurchasedRemaining != 50 {
		t.Fatalf("second ensure reset balance: %+v", again)
	}
}
