package finance

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type memTransactionRepo struct {
	byID map[int]*Transaction
}

func (r *memTransactionRepo) Save(ctx context.Context, transaction *Transaction) error {
	r.byID[transaction.ID().Value()] = transaction
	return nil
}
func (r *memTransactionRepo) FindByID(ctx context.Context, id TransactionID) (*Transaction, error) {
	tx, ok := r.byID[id.Value()]
	if !ok {
		return nil, errors.New("not found")
	}
	return tx, nil
}
func (r *memTransactionRepo) FindByIDAndType(ctx context.Context, id TransactionID, transactionType TransactionType) (*Transaction, error) {
	tx, err := r.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if tx.Type() != transactionType {
		return nil, errors.New("not found")
	}
	return tx, nil
}
func (r *memTransactionRepo) FindByUserID(ctx context.Context, userID UserID) ([]*Transaction, error) {
	return nil, nil
}
func (r *memTransactionRepo) FindByUserIDAndDateRange(ctx context.Context, userID UserID, startDate, endDate time.Time) ([]*Transaction, error) {
	return nil, nil
}
func (r *memTransactionRepo) FindByUserIDAndCategory(ctx context.Context, userID UserID, categoryID CategoryID) ([]*Transaction, error) {
	return nil, nil
}
func (r *memTransactionRepo) FindByUserIDWithFilters(ctx context.Context, userID UserID, filters TransactionFilters) ([]*Transaction, int64, error) {
	return nil, 0, nil
}
func (r *memTransactionRepo) Delete(ctx context.Context, id TransactionID) error { return nil }
func (r *memTransactionRepo) DeleteByIDAndType(ctx context.Context, id TransactionID, transactionType TransactionType) error {
	return nil
}
func (r *memTransactionRepo) GetTotalCount(ctx context.Context) (int, error)         { return 0, nil }
func (r *memTransactionRepo) GetTotalExpenses(ctx context.Context) (float64, error)  { return 0, nil }
func (r *memTransactionRepo) GetTotalIncome(ctx context.Context) (float64, error)    { return 0, nil }

type memCategoryRepo struct {
	byID map[int]*Category
}

func (r *memCategoryRepo) Save(ctx context.Context, category *Category) error { return nil }
func (r *memCategoryRepo) FindByID(ctx context.Context, id CategoryID) (*Category, error) {
	c, ok := r.byID[id.Value()]
	if !ok {
		return nil, errors.New("category not found")
	}
	return c, nil
}
func (r *memCategoryRepo) FindByUserID(ctx context.Context, userID UserID) ([]*Category, error) {
	return nil, nil
}
func (r *memCategoryRepo) FindByUserIDAndType(ctx context.Context, userID UserID, categoryType CategoryType) ([]*Category, error) {
	return nil, nil
}
func (r *memCategoryRepo) FindDefaultCategories(ctx context.Context) ([]*Category, error) {
	return nil, nil
}
func (r *memCategoryRepo) Delete(ctx context.Context, id CategoryID) error { return nil }
func (r *memCategoryRepo) ExistsByID(ctx context.Context, id CategoryID) (bool, error) {
	_, ok := r.byID[id.Value()]
	return ok, nil
}

func newUpdateTxFixture(t *testing.T) (
	*TransactionService,
	*memTransactionRepo,
	*memWalletRepo,
	*Transaction,
) {
	t.Helper()

	idr, err := NewCurrency(NewCurrencyID(1), nil, "IDR", "Rupiah", "Rp", true)
	if err != nil {
		t.Fatalf("currency: %v", err)
	}
	usd, err := NewCurrency(NewCurrencyID(2), nil, "USD", "US Dollar", "$", true)
	if err != nil {
		t.Fatalf("currency: %v", err)
	}
	cat, err := NewCategory(NewCategoryID(1), nil, "Food", "#000", true, CategoryTypeExpense)
	if err != nil {
		t.Fatalf("category: %v", err)
	}

	walletA := ReconstituteWallet(
		NewWalletID(1), NewUserID(1), "Cash", WalletTypeCash, NewCurrencyID(1), 0, true, false, time.Now(),
	)
	walletB := ReconstituteWallet(
		NewWalletID(2), NewUserID(1), "Bank", WalletTypeBank, NewCurrencyID(1), 0, false, false, time.Now(),
	)
	walletUSD := ReconstituteWallet(
		NewWalletID(3), NewUserID(1), "USD Cash", WalletTypeCash, NewCurrencyID(2), 0, false, false, time.Now(),
	)
	otherUserWallet := ReconstituteWallet(
		NewWalletID(4), NewUserID(99), "Other", WalletTypeCash, NewCurrencyID(1), 0, true, false, time.Now(),
	)
	archived := ReconstituteWallet(
		NewWalletID(5), NewUserID(1), "Old", WalletTypeBank, NewCurrencyID(1), 0, false, true, time.Now(),
	)

	amount, err := NewMoney(100, NewCurrencyID(1))
	if err != nil {
		t.Fatalf("money: %v", err)
	}
	tx := NewTransaction(
		NewTransactionID(10),
		NewUserID(1),
		NewWalletID(1),
		NewCategoryID(1),
		NewCurrencyID(1),
		amount,
		"lunch",
		time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		TransactionTypeExpense,
	)

	txRepo := &memTransactionRepo{byID: map[int]*Transaction{10: tx}}
	walletRepo := newMemWalletRepo(walletA, walletB, walletUSD, otherUserWallet, archived)
	catRepo := &memCategoryRepo{byID: map[int]*Category{1: cat}}
	curRepo := &memCurrencyRepo{byID: map[int]*Currency{1: idr, 2: usd}}

	svc := NewTransactionService(txRepo, catRepo, curRepo, walletRepo)
	return svc, txRepo, walletRepo, tx
}

func TestUpdateTransaction_ChangeWalletSameCurrency(t *testing.T) {
	svc, _, _, _ := newUpdateTxFixture(t)

	updated, err := svc.UpdateTransaction(
		context.Background(),
		NewTransactionID(10),
		NewUserID(1),
		NewWalletID(2),
		NewCategoryID(1),
		120,
		"lunch updated",
		time.Date(2024, 6, 2, 0, 0, 0, 0, time.UTC),
		TransactionTypeExpense,
	)
	if err != nil {
		t.Fatalf("UpdateTransaction: %v", err)
	}
	if updated.WalletID().Value() != 2 {
		t.Fatalf("expected wallet 2, got %d", updated.WalletID().Value())
	}
	if updated.CurrencyID().Value() != 1 {
		t.Fatalf("expected currency 1, got %d", updated.CurrencyID().Value())
	}
	if updated.Amount().Amount() != 120 {
		t.Fatalf("expected amount 120, got %v", updated.Amount().Amount())
	}
}

func TestUpdateTransaction_ChangeWalletCrossCurrency(t *testing.T) {
	svc, _, _, _ := newUpdateTxFixture(t)

	updated, err := svc.UpdateTransaction(
		context.Background(),
		NewTransactionID(10),
		NewUserID(1),
		NewWalletID(3),
		NewCategoryID(1),
		100,
		"lunch",
		time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		TransactionTypeExpense,
	)
	if err != nil {
		t.Fatalf("UpdateTransaction: %v", err)
	}
	if updated.WalletID().Value() != 3 {
		t.Fatalf("expected wallet 3, got %d", updated.WalletID().Value())
	}
	if updated.CurrencyID().Value() != 2 {
		t.Fatalf("expected currency 2, got %d", updated.CurrencyID().Value())
	}
}

func TestUpdateTransaction_RejectOtherUserWallet(t *testing.T) {
	svc, _, _, _ := newUpdateTxFixture(t)

	_, err := svc.UpdateTransaction(
		context.Background(),
		NewTransactionID(10),
		NewUserID(1),
		NewWalletID(4),
		NewCategoryID(1),
		100,
		"lunch",
		time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		TransactionTypeExpense,
	)
	if err == nil || !strings.Contains(err.Error(), "access denied to wallet") {
		t.Fatalf("expected access denied to wallet, got %v", err)
	}
}

func TestUpdateTransaction_RejectArchivedWallet(t *testing.T) {
	svc, _, _, _ := newUpdateTxFixture(t)

	_, err := svc.UpdateTransaction(
		context.Background(),
		NewTransactionID(10),
		NewUserID(1),
		NewWalletID(5),
		NewCategoryID(1),
		100,
		"lunch",
		time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC),
		TransactionTypeExpense,
	)
	if err == nil || !strings.Contains(err.Error(), "archived") {
		t.Fatalf("expected archived wallet error, got %v", err)
	}
}
