package ai

import "testing"

func TestCreditBalanceSpendOrder(t *testing.T) {
	b := NewCreditBalance(1, true)
	if b.Available() != TrialIncludedCap() {
		t.Fatalf("trial available=%d", b.Available())
	}
	if err := b.Spend(); err != nil {
		t.Fatal(err)
	}
	if b.IncludedUsed != 1 || b.PurchasedRemaining != 0 {
		t.Fatalf("expected included spend, got %+v", b)
	}
	b.AddPurchase(2)
	b.IncludedUsed = b.IncludedUnlocked
	if err := b.Spend(); err != nil {
		t.Fatal(err)
	}
	if b.PurchasedRemaining != 1 {
		t.Fatalf("expected purchased spend, remaining=%d", b.PurchasedRemaining)
	}
}

func TestUnlockFullIncludedPreservesUsage(t *testing.T) {
	b := NewCreditBalance(1, true)
	if err := b.Spend(); err != nil {
		t.Fatal(err)
	}
	b.AddPurchase(50)
	used, purchased := b.IncludedUsed, b.PurchasedRemaining

	b.UnlockFullIncluded()
	b.UnlockFullIncluded()

	if b.IncludedUnlocked != IncludedGrant() {
		t.Fatalf("unlocked=%d", b.IncludedUnlocked)
	}
	if b.IncludedGranted != IncludedGrant() {
		t.Fatalf("granted=%d", b.IncludedGranted)
	}
	if b.IncludedUsed != used || b.PurchasedRemaining != purchased {
		t.Fatalf("usage reset: used=%d purchased=%d", b.IncludedUsed, b.PurchasedRemaining)
	}
}

func TestSyncUnlockKeepsTrialCap(t *testing.T) {
	b := NewCreditBalance(1, true)
	if err := b.Spend(); err != nil {
		t.Fatal(err)
	}
	b.AddPurchase(10)
	used, purchased := b.IncludedUsed, b.PurchasedRemaining

	SyncUnlockForSubscription(b, true)
	SyncUnlockForSubscription(b, true)

	if b.IncludedUnlocked != TrialIncludedCap() {
		t.Fatalf("trial unlocked=%d", b.IncludedUnlocked)
	}
	if b.IncludedUsed != used || b.PurchasedRemaining != purchased {
		t.Fatalf("trial usage reset: used=%d purchased=%d", b.IncludedUsed, b.PurchasedRemaining)
	}
}

func TestSyncUnlockOpensPaidGrantWithoutReset(t *testing.T) {
	b := NewCreditBalance(1, true)
	for range 3 {
		if err := b.Spend(); err != nil {
			t.Fatal(err)
		}
	}
	b.AddPurchase(20)
	used, purchased := b.IncludedUsed, b.PurchasedRemaining

	SyncUnlockForSubscription(b, false)
	SyncUnlockForSubscription(b, false)

	if b.IncludedUnlocked != IncludedGrant() {
		t.Fatalf("paid unlocked=%d", b.IncludedUnlocked)
	}
	if b.IncludedUsed != used || b.PurchasedRemaining != purchased {
		t.Fatalf("paid usage reset: used=%d purchased=%d", b.IncludedUsed, b.PurchasedRemaining)
	}
	if b.Available() != (IncludedGrant()-used)+purchased {
		t.Fatalf("available=%d", b.Available())
	}
}
