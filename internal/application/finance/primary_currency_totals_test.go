package finance

import (
	"context"
	"testing"
	"time"

	"panda-pocket/internal/domain/entitlement"
	domainFinance "panda-pocket/internal/domain/finance"
)

type analyticsTxRepo struct {
	items []*domainFinance.Transaction
}

func (r *analyticsTxRepo) Save(ctx context.Context, transaction *domainFinance.Transaction) error {
	return nil
}
func (r *analyticsTxRepo) FindByID(ctx context.Context, id domainFinance.TransactionID) (*domainFinance.Transaction, error) {
	return nil, nil
}
func (r *analyticsTxRepo) FindByIDAndType(ctx context.Context, id domainFinance.TransactionID, transactionType domainFinance.TransactionType) (*domainFinance.Transaction, error) {
	return nil, nil
}
func (r *analyticsTxRepo) FindByUserID(ctx context.Context, userID domainFinance.UserID) ([]*domainFinance.Transaction, error) {
	return r.items, nil
}
func (r *analyticsTxRepo) FindByUserIDAndDateRange(ctx context.Context, userID domainFinance.UserID, startDate, endDate time.Time) ([]*domainFinance.Transaction, error) {
	return r.items, nil
}
func (r *analyticsTxRepo) FindByUserIDAndCategory(ctx context.Context, userID domainFinance.UserID, categoryID domainFinance.CategoryID) ([]*domainFinance.Transaction, error) {
	return nil, nil
}
func (r *analyticsTxRepo) FindByUserIDWithFilters(ctx context.Context, userID domainFinance.UserID, filters domainFinance.TransactionFilters) ([]*domainFinance.Transaction, int64, error) {
	return r.items, int64(len(r.items)), nil
}
func (r *analyticsTxRepo) Delete(ctx context.Context, id domainFinance.TransactionID) error {
	return nil
}
func (r *analyticsTxRepo) DeleteByIDAndType(ctx context.Context, id domainFinance.TransactionID, transactionType domainFinance.TransactionType) error {
	return nil
}
func (r *analyticsTxRepo) GetTotalCount(ctx context.Context) (int, error) {
	return len(r.items), nil
}
func (r *analyticsTxRepo) GetTotalExpenses(ctx context.Context) (float64, error) {
	return 0, nil
}
func (r *analyticsTxRepo) GetTotalIncome(ctx context.Context) (float64, error) {
	return 0, nil
}

type analyticsCurrencyRepo struct {
	primaryID int
}

func (s *analyticsCurrencyRepo) Save(ctx context.Context, currency *domainFinance.Currency) error {
	return nil
}
func (s *analyticsCurrencyRepo) FindByID(ctx context.Context, id domainFinance.CurrencyID) (*domainFinance.Currency, error) {
	code := "AUD"
	symbol := "A$"
	if id.Value() == 1 {
		code = "IDR"
		symbol = "Rp"
	}
	c, err := domainFinance.NewCurrency(id, nil, code, code, symbol, true)
	if err != nil {
		return nil, err
	}
	return c, nil
}
func (s *analyticsCurrencyRepo) FindByUserID(ctx context.Context, userID domainFinance.UserID) ([]*domainFinance.Currency, error) {
	return nil, nil
}
func (s *analyticsCurrencyRepo) FindSystemCurrencies(ctx context.Context) ([]*domainFinance.Currency, error) {
	c, _ := s.FindByID(ctx, domainFinance.NewCurrencyID(s.primaryID))
	return []*domainFinance.Currency{c}, nil
}
func (s *analyticsCurrencyRepo) Delete(ctx context.Context, id domainFinance.CurrencyID) error {
	return nil
}
func (s *analyticsCurrencyRepo) ExistsByID(ctx context.Context, id domainFinance.CurrencyID) (bool, error) {
	return true, nil
}
func (s *analyticsCurrencyRepo) ExistsByCodeAndUserID(ctx context.Context, code string, userID domainFinance.UserID) (bool, error) {
	return false, nil
}
func (s *analyticsCurrencyRepo) SetUserDefaultCurrency(ctx context.Context, userID domainFinance.UserID, currencyID domainFinance.CurrencyID) error {
	return nil
}
func (s *analyticsCurrencyRepo) GetUserDefaultCurrency(ctx context.Context, userID domainFinance.UserID) (*domainFinance.Currency, error) {
	return s.FindByID(ctx, domainFinance.NewCurrencyID(s.primaryID))
}

type analyticsCategoryRepo struct{}

func (s *analyticsCategoryRepo) Save(ctx context.Context, category *domainFinance.Category) error {
	return nil
}
func (s *analyticsCategoryRepo) FindByID(ctx context.Context, id domainFinance.CategoryID) (*domainFinance.Category, error) {
	return domainFinance.NewCategory(id, nil, "Food", "#fff", true, domainFinance.CategoryTypeExpense)
}
func (s *analyticsCategoryRepo) FindByUserID(ctx context.Context, userID domainFinance.UserID) ([]*domainFinance.Category, error) {
	return nil, nil
}
func (s *analyticsCategoryRepo) FindByUserIDAndType(ctx context.Context, userID domainFinance.UserID, categoryType domainFinance.CategoryType) ([]*domainFinance.Category, error) {
	return nil, nil
}
func (s *analyticsCategoryRepo) FindDefaultCategories(ctx context.Context) ([]*domainFinance.Category, error) {
	return nil, nil
}
func (s *analyticsCategoryRepo) Delete(ctx context.Context, id domainFinance.CategoryID) error {
	return nil
}
func (s *analyticsCategoryRepo) ExistsByID(ctx context.Context, id domainFinance.CategoryID) (bool, error) {
	return true, nil
}

func mustMoney(t *testing.T, amount float64, currencyID int) domainFinance.Money {
	t.Helper()
	m, err := domainFinance.NewMoney(amount, domainFinance.NewCurrencyID(currencyID))
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func TestGetAnalyticsUseCase_FiltersToPrimaryCurrency(t *testing.T) {
	now := time.Now()
	idrExpense := domainFinance.NewTransaction(
		domainFinance.NewTransactionID(1),
		domainFinance.NewUserID(1),
		domainFinance.NewWalletID(1),
		domainFinance.NewCategoryID(1),
		domainFinance.NewCurrencyID(1),
		mustMoney(t, 100_000, 1),
		"IDR lunch",
		now,
		domainFinance.TransactionTypeExpense,
	)
	audExpense := domainFinance.NewTransaction(
		domainFinance.NewTransactionID(2),
		domainFinance.NewUserID(1),
		domainFinance.NewWalletID(2),
		domainFinance.NewCategoryID(1),
		domainFinance.NewCurrencyID(21),
		mustMoney(t, 50, 21),
		"AUD coffee",
		now,
		domainFinance.TransactionTypeExpense,
	)
	idrIncome := domainFinance.NewTransaction(
		domainFinance.NewTransactionID(3),
		domainFinance.NewUserID(1),
		domainFinance.NewWalletID(1),
		domainFinance.NewCategoryID(1),
		domainFinance.NewCurrencyID(1),
		mustMoney(t, 1_000_000, 1),
		"IDR salary",
		now,
		domainFinance.TransactionTypeIncome,
	)

	txRepo := &analyticsTxRepo{items: []*domainFinance.Transaction{idrExpense, audExpense, idrIncome}}
	currencyRepo := &analyticsCurrencyRepo{primaryID: 1}
	txService := domainFinance.NewTransactionService(txRepo, &analyticsCategoryRepo{}, currencyRepo, nil)
	catService := domainFinance.NewCategoryService(&analyticsCategoryRepo{}, nil)
	currencyService := domainFinance.NewCurrencyService(currencyRepo)

	uc := NewGetAnalyticsUseCase(txService, catService, currencyService, entitlement.StaticChecker{Pro: true})
	resp, err := uc.Execute(context.Background(), 1, GetAnalyticsRequest{Period: "monthly"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CurrencyID != 1 {
		t.Fatalf("expected primary currency_id=1, got %d", resp.CurrencyID)
	}
	if resp.TotalSpent != 100_000 {
		t.Fatalf("expected total_spent=100000 (IDR only), got %v", resp.TotalSpent)
	}
	if resp.TotalIncome != 1_000_000 {
		t.Fatalf("expected total_income=1000000 (IDR only), got %v", resp.TotalIncome)
	}
	if resp.ExcludedTransactionCount != 1 {
		t.Fatalf("expected excluded_transaction_count=1, got %d", resp.ExcludedTransactionCount)
	}
	if resp.TransactionCount != 2 {
		t.Fatalf("expected transaction_count=2 (primary only), got %d", resp.TransactionCount)
	}
}

func TestGetAllTransactionsUseCase_TotalsPrimaryCurrencyOnly(t *testing.T) {
	now := time.Now()
	idrExpense := domainFinance.NewTransaction(
		domainFinance.NewTransactionID(1),
		domainFinance.NewUserID(1),
		domainFinance.NewWalletID(1),
		domainFinance.NewCategoryID(1),
		domainFinance.NewCurrencyID(1),
		mustMoney(t, 200_000, 1),
		"IDR",
		now,
		domainFinance.TransactionTypeExpense,
	)
	audIncome := domainFinance.NewTransaction(
		domainFinance.NewTransactionID(2),
		domainFinance.NewUserID(1),
		domainFinance.NewWalletID(2),
		domainFinance.NewCategoryID(1),
		domainFinance.NewCurrencyID(21),
		mustMoney(t, 1000, 21),
		"AUD",
		now,
		domainFinance.TransactionTypeIncome,
	)

	txRepo := &analyticsTxRepo{items: []*domainFinance.Transaction{idrExpense, audIncome}}
	currencyRepo := &analyticsCurrencyRepo{primaryID: 1}
	txService := domainFinance.NewTransactionService(txRepo, &analyticsCategoryRepo{}, currencyRepo, nil)
	catService := domainFinance.NewCategoryService(&analyticsCategoryRepo{}, nil)
	currencyService := domainFinance.NewCurrencyService(currencyRepo)

	uc := NewGetAllTransactionsUseCase(txService, catService, currencyService)
	resp, err := uc.Execute(context.Background(), 1, GetAllTransactionsRequest{Page: 1, Limit: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Transactions) != 2 {
		t.Fatalf("expected both rows in list, got %d", len(resp.Transactions))
	}
	if resp.TotalExpenses != 200_000 {
		t.Fatalf("expected total_expenses=200000, got %v", resp.TotalExpenses)
	}
	if resp.TotalIncomes != 0 {
		t.Fatalf("expected total_incomes=0 (AUD excluded), got %v", resp.TotalIncomes)
	}
	if resp.CurrencyID != 1 {
		t.Fatalf("expected currency_id=1, got %d", resp.CurrencyID)
	}
}
