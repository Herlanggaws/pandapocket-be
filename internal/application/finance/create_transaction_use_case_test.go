package finance

import (
	"context"
	"errors"
	"testing"
	"time"

	"panda-pocket/internal/domain/entitlement"
	domainFinance "panda-pocket/internal/domain/finance"
	"panda-pocket/internal/infrastructure/analytics"
)

func TestNormalizeExpenseSource(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"foto_struk", "foto_struk"},
		{"tanya_ai", "tanya_ai"},
		{"manual", "manual"},
		{"", "manual"},
		{"receipt", "manual"},
		{"FOTO_STRUK", "manual"},
		{"something_else", "manual"},
	}
	for _, tc := range cases {
		if got := normalizeExpenseSource(tc.in); got != tc.want {
			t.Fatalf("normalizeExpenseSource(%q)=%q, want %q", tc.in, got, tc.want)
		}
	}
}

type createTxRepo struct {
	items []*domainFinance.Transaction
	next  int
}

func (r *createTxRepo) Save(ctx context.Context, transaction *domainFinance.Transaction) error {
	if transaction.ID().Value() == 0 {
		r.next++
		transaction.AssignID(domainFinance.NewTransactionID(r.next))
	}
	r.items = append(r.items, transaction)
	return nil
}

func (r *createTxRepo) FindByID(ctx context.Context, id domainFinance.TransactionID) (*domainFinance.Transaction, error) {
	return nil, errors.New("not found")
}
func (r *createTxRepo) FindByIDAndType(ctx context.Context, id domainFinance.TransactionID, transactionType domainFinance.TransactionType) (*domainFinance.Transaction, error) {
	return nil, errors.New("not found")
}
func (r *createTxRepo) FindByUserID(ctx context.Context, userID domainFinance.UserID) ([]*domainFinance.Transaction, error) {
	return r.items, nil
}
func (r *createTxRepo) FindByUserIDAndDateRange(ctx context.Context, userID domainFinance.UserID, startDate, endDate time.Time) ([]*domainFinance.Transaction, error) {
	return nil, nil
}
func (r *createTxRepo) FindByUserIDAndCategory(ctx context.Context, userID domainFinance.UserID, categoryID domainFinance.CategoryID) ([]*domainFinance.Transaction, error) {
	return nil, nil
}
func (r *createTxRepo) FindByUserIDWithFilters(ctx context.Context, userID domainFinance.UserID, filters domainFinance.TransactionFilters) ([]*domainFinance.Transaction, int64, error) {
	var matched []*domainFinance.Transaction
	for _, tx := range r.items {
		if tx.UserID().Value() != userID.Value() {
			continue
		}
		if filters.TransactionType != nil && tx.Type() != *filters.TransactionType {
			continue
		}
		if filters.StartDate != nil && tx.CreatedAt().Before(*filters.StartDate) {
			continue
		}
		if filters.EndDate != nil && tx.CreatedAt().After(*filters.EndDate) {
			continue
		}
		matched = append(matched, tx)
	}
	return matched, int64(len(matched)), nil
}
func (r *createTxRepo) Delete(ctx context.Context, id domainFinance.TransactionID) error { return nil }
func (r *createTxRepo) DeleteByIDAndType(ctx context.Context, id domainFinance.TransactionID, transactionType domainFinance.TransactionType) error {
	return nil
}
func (r *createTxRepo) GetTotalCount(ctx context.Context) (int, error)        { return 0, nil }
func (r *createTxRepo) GetTotalExpenses(ctx context.Context) (float64, error) { return 0, nil }
func (r *createTxRepo) GetTotalIncome(ctx context.Context) (float64, error)   { return 0, nil }

type createTxCategoryRepo struct {
	cat *domainFinance.Category
}

func (r *createTxCategoryRepo) Save(ctx context.Context, category *domainFinance.Category) error {
	return nil
}
func (r *createTxCategoryRepo) FindByID(ctx context.Context, id domainFinance.CategoryID) (*domainFinance.Category, error) {
	if r.cat != nil && r.cat.ID().Value() == id.Value() {
		return r.cat, nil
	}
	return nil, errors.New("category not found")
}
func (r *createTxCategoryRepo) FindByUserID(ctx context.Context, userID domainFinance.UserID) ([]*domainFinance.Category, error) {
	return nil, nil
}
func (r *createTxCategoryRepo) FindByUserIDAndType(ctx context.Context, userID domainFinance.UserID, categoryType domainFinance.CategoryType) ([]*domainFinance.Category, error) {
	return nil, nil
}
func (r *createTxCategoryRepo) FindDefaultCategories(ctx context.Context) ([]*domainFinance.Category, error) {
	return nil, nil
}
func (r *createTxCategoryRepo) Delete(ctx context.Context, id domainFinance.CategoryID) error {
	return nil
}
func (r *createTxCategoryRepo) ExistsByID(ctx context.Context, id domainFinance.CategoryID) (bool, error) {
	return r.cat != nil && r.cat.ID().Value() == id.Value(), nil
}

type createTxCurrencyRepo struct {
	cur *domainFinance.Currency
}

func (r *createTxCurrencyRepo) Save(ctx context.Context, currency *domainFinance.Currency) error {
	return nil
}
func (r *createTxCurrencyRepo) FindByID(ctx context.Context, id domainFinance.CurrencyID) (*domainFinance.Currency, error) {
	if r.cur != nil && r.cur.ID().Value() == id.Value() {
		return r.cur, nil
	}
	return nil, errors.New("currency not found")
}
func (r *createTxCurrencyRepo) FindByUserID(ctx context.Context, userID domainFinance.UserID) ([]*domainFinance.Currency, error) {
	return nil, nil
}
func (r *createTxCurrencyRepo) FindSystemCurrencies(ctx context.Context) ([]*domainFinance.Currency, error) {
	return nil, nil
}
func (r *createTxCurrencyRepo) Delete(ctx context.Context, id domainFinance.CurrencyID) error {
	return nil
}
func (r *createTxCurrencyRepo) ExistsByID(ctx context.Context, id domainFinance.CurrencyID) (bool, error) {
	return r.cur != nil && r.cur.ID().Value() == id.Value(), nil
}
func (r *createTxCurrencyRepo) ExistsByCodeAndUserID(ctx context.Context, code string, userID domainFinance.UserID) (bool, error) {
	return false, nil
}
func (r *createTxCurrencyRepo) SetUserDefaultCurrency(ctx context.Context, userID domainFinance.UserID, currencyID domainFinance.CurrencyID) error {
	return nil
}
func (r *createTxCurrencyRepo) GetUserDefaultCurrency(ctx context.Context, userID domainFinance.UserID) (*domainFinance.Currency, error) {
	return r.cur, nil
}

type createTxWalletRepo struct {
	wallet *domainFinance.Wallet
}

func (r *createTxWalletRepo) Save(ctx context.Context, wallet *domainFinance.Wallet) error {
	return nil
}
func (r *createTxWalletRepo) FindByID(ctx context.Context, id domainFinance.WalletID) (*domainFinance.Wallet, error) {
	if r.wallet != nil && r.wallet.ID().Value() == id.Value() {
		return r.wallet, nil
	}
	return nil, errors.New("wallet not found")
}
func (r *createTxWalletRepo) FindByUserID(ctx context.Context, userID domainFinance.UserID, includeArchived bool) ([]*domainFinance.Wallet, error) {
	if r.wallet != nil {
		return []*domainFinance.Wallet{r.wallet}, nil
	}
	return nil, nil
}
func (r *createTxWalletRepo) FindDefaultByUserID(ctx context.Context, userID domainFinance.UserID) (*domainFinance.Wallet, error) {
	if r.wallet != nil {
		return r.wallet, nil
	}
	return nil, errors.New("default wallet not found")
}
func (r *createTxWalletRepo) CountActiveByUserID(ctx context.Context, userID domainFinance.UserID) (int64, error) {
	return 1, nil
}
func (r *createTxWalletRepo) ClearDefaultForUser(ctx context.Context, userID domainFinance.UserID) error {
	return nil
}
func (r *createTxWalletRepo) HasTransactions(ctx context.Context, id domainFinance.WalletID) (bool, error) {
	return false, nil
}
func (r *createTxWalletRepo) GetBalanceBreakdown(ctx context.Context, id domainFinance.WalletID) (domainFinance.WalletBalanceBreakdown, error) {
	return domainFinance.WalletBalanceBreakdown{}, nil
}
func (r *createTxWalletRepo) AlignPendingCurrency(ctx context.Context, id domainFinance.WalletID, currencyID domainFinance.CurrencyID) error {
	return nil
}
func (r *createTxWalletRepo) AlignRecurringCurrency(ctx context.Context, id domainFinance.WalletID, currencyID domainFinance.CurrencyID) error {
	return nil
}

func newCreateTransactionUseCase(t *testing.T) *CreateTransactionUseCase {
	t.Helper()
	idr, err := domainFinance.NewCurrency(domainFinance.NewCurrencyID(1), nil, "IDR", "Rupiah", "Rp", true)
	if err != nil {
		t.Fatalf("currency: %v", err)
	}
	cat, err := domainFinance.NewCategory(domainFinance.NewCategoryID(1), nil, "Food", "#000", true, domainFinance.CategoryTypeExpense)
	if err != nil {
		t.Fatalf("category: %v", err)
	}
	wallet := domainFinance.ReconstituteWallet(
		domainFinance.NewWalletID(1),
		domainFinance.NewUserID(1),
		"Cash",
		domainFinance.WalletTypeCash,
		domainFinance.NewCurrencyID(1),
		0,
		true,
		false,
		time.Now(),
	)

	txRepo := &createTxRepo{}
	catRepo := &createTxCategoryRepo{cat: cat}
	curRepo := &createTxCurrencyRepo{cur: idr}
	walletRepo := &createTxWalletRepo{wallet: wallet}

	txService := domainFinance.NewTransactionService(txRepo, catRepo, curRepo, walletRepo)
	walletService := domainFinance.NewWalletService(walletRepo, curRepo)
	return NewCreateTransactionUseCase(txService, walletService, entitlement.StaticChecker{Pro: true})
}

type analyticsCapture struct {
	event string
	props map[string]interface{}
}

func collectAnalyticsCaptures(t *testing.T) *[]analyticsCapture {
	t.Helper()
	captures := &[]analyticsCapture{}
	analytics.SetTestCapture(func(userID int, event string, properties map[string]interface{}) {
		*captures = append(*captures, analyticsCapture{event: event, props: properties})
	})
	t.Cleanup(func() { analytics.SetTestCapture(nil) })
	return captures
}

func eventSources(captures []analyticsCapture, event string) []string {
	var sources []string
	for _, c := range captures {
		if c.event != event {
			continue
		}
		src, _ := c.props["source"].(string)
		sources = append(sources, src)
	}
	return sources
}

func TestCreateTransaction_FirstExpenseSource(t *testing.T) {
	cases := []struct {
		name       string
		source     string
		wantSource string
	}{
		{"foto_struk", "foto_struk", "foto_struk"},
		{"tanya_ai", "tanya_ai", "tanya_ai"},
		{"empty defaults to manual", "", "manual"},
		{"unknown defaults to manual", "receipt_ocr", "manual"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			captures := collectAnalyticsCaptures(t)

			uc := newCreateTransactionUseCase(t)
			_, err := uc.Execute(context.Background(), 1, CreateTransactionRequest{
				CategoryID: 1,
				Amount:     10000,
				Date:       "2024-01-15",
				Type:       "expense",
				Source:     tc.source,
			})
			if err != nil {
				t.Fatalf("Execute: %v", err)
			}

			expenseAdded := eventSources(*captures, "expense_added")
			firstExpense := eventSources(*captures, "first_expense_added")
			if len(expenseAdded) != 1 || expenseAdded[0] != tc.wantSource {
				t.Fatalf("expense_added=%v, want [%q]", expenseAdded, tc.wantSource)
			}
			if len(firstExpense) != 1 || firstExpense[0] != tc.wantSource {
				t.Fatalf("first_expense_added=%v, want [%q]", firstExpense, tc.wantSource)
			}
			for _, c := range *captures {
				if _, ok := c.props["amount"]; ok {
					t.Fatal("amount must not be in analytics properties")
				}
				if _, ok := c.props["description"]; ok {
					t.Fatal("description must not be in analytics properties")
				}
			}
		})
	}
}

func TestCreateTransaction_SecondExpenseOnlyExpenseAdded(t *testing.T) {
	uc := newCreateTransactionUseCase(t)
	_, err := uc.Execute(context.Background(), 1, CreateTransactionRequest{
		CategoryID: 1,
		Amount:     10000,
		Date:       "2024-01-15",
		Type:       "expense",
		Source:     "manual",
	})
	if err != nil {
		t.Fatalf("first Execute: %v", err)
	}

	captures := collectAnalyticsCaptures(t)
	_, err = uc.Execute(context.Background(), 1, CreateTransactionRequest{
		CategoryID: 1,
		Amount:     20000,
		Date:       "2024-01-16",
		Type:       "expense",
		Source:     "manual",
	})
	if err != nil {
		t.Fatalf("second Execute: %v", err)
	}

	if got := eventSources(*captures, "expense_added"); len(got) != 1 || got[0] != "manual" {
		t.Fatalf("expense_added=%v, want [manual]", got)
	}
	if got := eventSources(*captures, "first_expense_added"); len(got) != 0 {
		t.Fatalf("first_expense_added=%v, want none", got)
	}
}

func TestConfirmPending_FirstExpenseAnalytics(t *testing.T) {
	pendingRepo := newMemPendingRepo()
	txRepo := &memTxRepo{}
	amount, _ := domainFinance.NewMoney(99, domainFinance.NewCurrencyID(1))
	pt, err := domainFinance.NewPendingTransaction(
		domainFinance.NewUserID(1),
		domainFinance.NewWalletID(1),
		domainFinance.NewRecurringTransactionID(5),
		time.Now(),
		amount,
		"Gym",
		domainFinance.TransactionTypeExpense,
		domainFinance.NewCategoryID(1),
		domainFinance.NewCurrencyID(1),
	)
	if err != nil {
		t.Fatal(err)
	}
	_ = pendingRepo.Save(context.Background(), pt)

	captures := collectAnalyticsCaptures(t)
	txService := domainFinance.NewTransactionService(txRepo, &stubCategoryRepo{}, &stubCurrencyRepo{}, &stubWalletRepo{})
	uc := NewConfirmPendingTransactionUseCase(pendingRepo, txService)
	_, err = uc.Execute(context.Background(), 1, pt.ID().Value())
	if err != nil {
		t.Fatal(err)
	}

	if got := eventSources(*captures, "expense_added"); len(got) != 1 || got[0] != "recurring" {
		t.Fatalf("expense_added=%v, want [recurring]", got)
	}
	if got := eventSources(*captures, "first_expense_added"); len(got) != 1 || got[0] != "recurring" {
		t.Fatalf("first_expense_added=%v, want [recurring]", got)
	}
}

func TestCreateReceivable_DoesNotDoubleCountExpenseAdded(t *testing.T) {
	createTx, categoryService := newSideEffectCreateTxStack(t)
	receivableRepo := &sideEffectReceivableRepo{}
	uc := NewCreateReceivableUseCase(
		domainFinance.NewReceivableService(receivableRepo, &sideEffectCollectionRepo{}),
		createTx,
		categoryService,
		entitlement.StaticChecker{Pro: true},
	)

	captures := collectAnalyticsCaptures(t)
	_, err := uc.Execute(context.Background(), 1, CreateReceivableRequest{
		Name:           "Friend loan",
		Type:           "personal_loan",
		CurrencyID:     1,
		CurrentBalance: 50000,
		CreateExpense:  true,
		CategoryID:     intPtr(1),
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	expenseAdded := eventSources(*captures, "expense_added")
	if len(expenseAdded) != 1 || expenseAdded[0] != "receivable" {
		t.Fatalf("expense_added=%v, want [receivable] once", expenseAdded)
	}
	firstExpense := eventSources(*captures, "first_expense_added")
	if len(firstExpense) != 1 || firstExpense[0] != "receivable" {
		t.Fatalf("first_expense_added=%v, want [receivable] once", firstExpense)
	}
}

func TestRecordLiabilityPayment_DoesNotDoubleCountExpenseAdded(t *testing.T) {
	createTx, categoryService := newSideEffectCreateTxStack(t)
	principal := 100000.0
	liability := domainFinance.ReconstituteLiability(
		domainFinance.NewLiabilityID(10),
		domainFinance.NewUserID(1),
		"Bank loan",
		domainFinance.LiabilityTypeLoan,
		domainFinance.NewCurrencyID(1),
		100000,
		"",
		false,
		nil,
		time.Now(),
		domainFinance.LiabilityDebtDetails{OriginalPrincipal: &principal, MinimumPayment: 10000},
	)
	liabilityRepo := &sideEffectLiabilityRepo{byID: map[int]*domainFinance.Liability{10: liability}}
	uc := NewRecordLiabilityPaymentUseCase(
		domainFinance.NewLiabilityService(liabilityRepo, &sideEffectPaymentRepo{}),
		createTx,
		categoryService,
	)

	captures := collectAnalyticsCaptures(t)
	_, err := uc.Execute(context.Background(), 1, 10, RecordLiabilityPaymentRequest{
		Amount:        10000,
		PaidAt:        "2024-06-01",
		CreateExpense: true,
		CategoryID:    intPtr(1),
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	expenseAdded := eventSources(*captures, "expense_added")
	if len(expenseAdded) != 1 || expenseAdded[0] != "debt" {
		t.Fatalf("expense_added=%v, want [debt] once", expenseAdded)
	}
	firstExpense := eventSources(*captures, "first_expense_added")
	if len(firstExpense) != 1 || firstExpense[0] != "debt" {
		t.Fatalf("first_expense_added=%v, want [debt] once", firstExpense)
	}
}
