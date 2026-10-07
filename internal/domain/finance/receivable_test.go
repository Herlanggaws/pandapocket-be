package finance

import (
	"testing"
	"time"
)

func TestReceivableApplyAndReverseCollection(t *testing.T) {
	principal := 1000.0
	r, err := NewReceivable(
		NewUserID(1),
		"Friend",
		ReceivableTypePersonalLoan,
		NewCurrencyID(1),
		1000,
		"",
		nil,
		ReceivableDetails{OriginalPrincipal: &principal},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.ApplyCollection(400); err != nil {
		t.Fatal(err)
	}
	if r.CurrentBalance() != 600 {
		t.Fatalf("balance = %v", r.CurrentBalance())
	}
	pct := r.CollectionProgressPercent()
	if pct == nil || *pct < 39.9 || *pct > 40.1 {
		t.Fatalf("progress = %v", pct)
	}
	if err := r.ReverseCollection(400); err != nil {
		t.Fatal(err)
	}
	if r.CurrentBalance() != 1000 {
		t.Fatalf("restored balance = %v", r.CurrentBalance())
	}
}

func TestReceivableCollectionRejectsOverpay(t *testing.T) {
	r, err := NewReceivable(
		NewUserID(1),
		"Invoice",
		ReceivableTypeInvoice,
		NewCurrencyID(1),
		100,
		"",
		nil,
		ReceivableDetails{},
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.ApplyCollection(150); err == nil {
		t.Fatal("expected over-collection error")
	}
}

func TestParseReceivableType(t *testing.T) {
	if _, err := ParseReceivableType("personal_loan"); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseReceivableType("loan"); err == nil {
		t.Fatal("expected invalid type")
	}
	c, err := NewReceivableCollection(NewReceivableID(1), NewUserID(1), 50, time.Now(), nil, "note")
	if err != nil {
		t.Fatal(err)
	}
	if c.Amount() != 50 {
		t.Fatalf("amount = %v", c.Amount())
	}
}
