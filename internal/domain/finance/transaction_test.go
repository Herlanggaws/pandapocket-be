package finance

import (
	"testing"
	"time"
)

func TestTransactionAssignID(t *testing.T) {
	amount, err := NewMoney(100, NewCurrencyID(1))
	if err != nil {
		t.Fatalf("new money error: %v", err)
	}

	transaction := NewTransaction(
		TransactionID{},
		NewUserID(1),
		NewWalletID(1),
		NewCategoryID(1),
		NewCurrencyID(1),
		amount,
		"Debt payment",
		time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		TransactionTypeExpense,
	)

	transaction.AssignID(NewTransactionID(99))
	if transaction.ID().Value() != 99 {
		t.Fatalf("expected assigned id 99, got %d", transaction.ID().Value())
	}
}
