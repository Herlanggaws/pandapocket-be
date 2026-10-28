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

func TestUpdateWallet_SameCurrency(t *testing.T) {
	amount, err := NewMoney(50, NewCurrencyID(1))
	if err != nil {
		t.Fatalf("new money error: %v", err)
	}
	tx := NewTransaction(
		NewTransactionID(1),
		NewUserID(1),
		NewWalletID(1),
		NewCategoryID(1),
		NewCurrencyID(1),
		amount,
		"kopi",
		time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		TransactionTypeExpense,
	)

	if err := tx.UpdateWallet(NewWalletID(2), NewCurrencyID(1)); err != nil {
		t.Fatalf("UpdateWallet: %v", err)
	}
	if tx.WalletID().Value() != 2 {
		t.Fatalf("expected wallet 2, got %d", tx.WalletID().Value())
	}
	if tx.CurrencyID().Value() != 1 {
		t.Fatalf("expected currency 1, got %d", tx.CurrencyID().Value())
	}
	if tx.Amount().Amount() != 50 {
		t.Fatalf("expected amount 50, got %v", tx.Amount().Amount())
	}
}

func TestUpdateWallet_CrossCurrency(t *testing.T) {
	amount, err := NewMoney(50, NewCurrencyID(1))
	if err != nil {
		t.Fatalf("new money error: %v", err)
	}
	tx := NewTransaction(
		NewTransactionID(1),
		NewUserID(1),
		NewWalletID(1),
		NewCategoryID(1),
		NewCurrencyID(1),
		amount,
		"kopi",
		time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		TransactionTypeExpense,
	)

	if err := tx.UpdateWallet(NewWalletID(3), NewCurrencyID(2)); err != nil {
		t.Fatalf("UpdateWallet: %v", err)
	}
	if tx.WalletID().Value() != 3 {
		t.Fatalf("expected wallet 3, got %d", tx.WalletID().Value())
	}
	if tx.CurrencyID().Value() != 2 {
		t.Fatalf("expected currency 2, got %d", tx.CurrencyID().Value())
	}
	if tx.Amount().Currency().Value() != 2 {
		t.Fatalf("expected money currency 2, got %d", tx.Amount().Currency().Value())
	}
}
