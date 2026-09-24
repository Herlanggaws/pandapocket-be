package ai

import "testing"

func TestCreditBalanceAddPurchaseAccumulates(t *testing.T) {
	b := NewCreditBalance(1, true)
	if b.PurchasedRemaining != 0 {
		t.Fatalf("expected 0 purchased, got %d", b.PurchasedRemaining)
	}
	b.AddPurchase(50)
	b.AddPurchase(150)
	b.AddPurchase(50)
	if b.PurchasedRemaining != 250 {
		t.Fatalf("expected accumulated 250, got %d", b.PurchasedRemaining)
	}
	if b.Available() != b.IncludedRemaining()+250 {
		t.Fatalf("available should include purchased; got %d", b.Available())
	}
}
