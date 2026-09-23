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

func TestUnlockFullIncluded(t *testing.T) {
	b := NewCreditBalance(1, true)
	b.UnlockFullIncluded()
	if b.IncludedUnlocked != IncludedGrant() {
		t.Fatalf("unlocked=%d", b.IncludedUnlocked)
	}
}
