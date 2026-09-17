package finance

import (
	"testing"
	"time"
)

func TestNewPendingTransaction(t *testing.T) {
	amount, err := NewMoney(100000, NewCurrencyID(1))
	if err != nil {
		t.Fatal(err)
	}
	due := time.Date(2024, 6, 15, 12, 30, 0, 0, time.UTC)
	pt, err := NewPendingTransaction(
		NewUserID(1),
		NewRecurringTransactionID(10),
		due,
		amount,
		"Rent",
		TransactionTypeExpense,
		NewCategoryID(2),
		NewCurrencyID(1),
	)
	if err != nil {
		t.Fatal(err)
	}
	if pt.Status() != PendingStatusPending {
		t.Fatalf("status: got %s", pt.Status())
	}
	if !pt.DueDate().Equal(time.Date(2024, 6, 15, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("due date not truncated: %v", pt.DueDate())
	}
	if !pt.IsOpen() {
		t.Fatal("expected open")
	}
}

func TestPendingTransactionConfirm(t *testing.T) {
	amount, _ := NewMoney(50, NewCurrencyID(1))
	pt, err := NewPendingTransaction(
		NewUserID(1),
		NewRecurringTransactionID(1),
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		amount,
		"",
		TransactionTypeIncome,
		NewCategoryID(1),
		NewCurrencyID(1),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := pt.Confirm(); err != nil {
		t.Fatal(err)
	}
	if pt.Status() != PendingStatusConfirmed {
		t.Fatalf("got %s", pt.Status())
	}
	if pt.ResolvedAt() == nil {
		t.Fatal("expected resolved_at")
	}
	if err := pt.Confirm(); err == nil {
		t.Fatal("expected error on second confirm")
	}
	if err := pt.Reject(); err == nil {
		t.Fatal("expected error rejecting confirmed")
	}
}

func TestPendingTransactionReject(t *testing.T) {
	amount, _ := NewMoney(50, NewCurrencyID(1))
	pt, err := NewPendingTransaction(
		NewUserID(1),
		NewRecurringTransactionID(1),
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		amount,
		"",
		TransactionTypeExpense,
		NewCategoryID(1),
		NewCurrencyID(1),
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := pt.Reject(); err != nil {
		t.Fatal(err)
	}
	if pt.Status() != PendingStatusRejected {
		t.Fatalf("got %s", pt.Status())
	}
	if pt.IsOpen() {
		t.Fatal("expected closed")
	}
}
