package finance

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	domainFinance "panda-pocket/internal/domain/finance"
	domainIdentity "panda-pocket/internal/domain/identity"
)

type memRecurringRepo struct {
	mu    sync.Mutex
	items map[int]*domainFinance.RecurringTransaction
	next  int
}

func newMemRecurringRepo() *memRecurringRepo {
	return &memRecurringRepo{items: map[int]*domainFinance.RecurringTransaction{}, next: 1}
}

func (r *memRecurringRepo) Save(ctx context.Context, rt *domainFinance.RecurringTransaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if rt.ID().Value() == 0 {
		rt.AssignID(domainFinance.NewRecurringTransactionID(r.next))
		r.next++
	}
	r.items[rt.ID().Value()] = rt
	return nil
}

func (r *memRecurringRepo) FindByID(ctx context.Context, id domainFinance.RecurringTransactionID) (*domainFinance.RecurringTransaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	rt, ok := r.items[id.Value()]
	if !ok {
		return nil, errors.New("not found")
	}
	return rt, nil
}

func (r *memRecurringRepo) FindByUserID(ctx context.Context, userID domainFinance.UserID) ([]*domainFinance.RecurringTransaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*domainFinance.RecurringTransaction, 0)
	for _, rt := range r.items {
		if rt.UserID().Value() == userID.Value() {
			out = append(out, rt)
		}
	}
	return out, nil
}

func (r *memRecurringRepo) FindActiveByUserID(ctx context.Context, userID domainFinance.UserID) ([]*domainFinance.RecurringTransaction, error) {
	return r.FindByUserID(ctx, userID)
}

func (r *memRecurringRepo) FindDueTransactions(ctx context.Context) ([]*domainFinance.RecurringTransaction, error) {
	return nil, nil
}

func (r *memRecurringRepo) Delete(ctx context.Context, id domainFinance.RecurringTransactionID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.items, id.Value())
	return nil
}

type memPendingRepo struct {
	mu    sync.Mutex
	items map[int]*domainFinance.PendingTransaction
	next  int
}

func newMemPendingRepo() *memPendingRepo {
	return &memPendingRepo{items: map[int]*domainFinance.PendingTransaction{}, next: 1}
}

func (r *memPendingRepo) Save(ctx context.Context, pending *domainFinance.PendingTransaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if pending.ID().Value() == 0 {
		for _, existing := range r.items {
			if existing.RecurringTransactionID().Value() == pending.RecurringTransactionID().Value() &&
				existing.DueDate().Equal(pending.DueDate()) {
				return errors.New("duplicate recurring due date")
			}
		}
		pending.AssignID(domainFinance.NewPendingTransactionID(r.next))
		r.next++
	}
	r.items[pending.ID().Value()] = pending
	return nil
}

func (r *memPendingRepo) FindByID(ctx context.Context, id domainFinance.PendingTransactionID) (*domainFinance.PendingTransaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	pt, ok := r.items[id.Value()]
	if !ok {
		return nil, errors.New("pending transaction not found")
	}
	return pt, nil
}

func (r *memPendingRepo) FindOpenByUserID(ctx context.Context, userID domainFinance.UserID) ([]*domainFinance.PendingTransaction, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]*domainFinance.PendingTransaction, 0)
	for _, pt := range r.items {
		if pt.UserID().Value() == userID.Value() && pt.IsOpen() {
			out = append(out, pt)
		}
	}
	return out, nil
}

func (r *memPendingRepo) ExistsByRecurringAndDueDate(ctx context.Context, recurringID domainFinance.RecurringTransactionID, dueDate time.Time) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	due := dueDate.Truncate(24 * time.Hour)
	for _, pt := range r.items {
		if pt.RecurringTransactionID().Value() == recurringID.Value() && pt.DueDate().Equal(due) {
			return true, nil
		}
	}
	return false, nil
}

func (r *memPendingRepo) openCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	n := 0
	for _, pt := range r.items {
		if pt.IsOpen() {
			n++
		}
	}
	return n
}

func (r *memPendingRepo) totalCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.items)
}

type memPrefsRepo struct{}

func (m *memPrefsRepo) FindByUserID(ctx context.Context, userID domainIdentity.UserID) (*domainIdentity.UserPreferences, error) {
	return nil, errors.New("not found")
}

func (m *memPrefsRepo) Save(ctx context.Context, prefs *domainIdentity.UserPreferences) error {
	return nil
}

type memTxRepo struct {
	mu    sync.Mutex
	saved []*domainFinance.Transaction
}

func (r *memTxRepo) Save(ctx context.Context, transaction *domainFinance.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.saved = append(r.saved, transaction)
	return nil
}
func (r *memTxRepo) FindByID(ctx context.Context, id domainFinance.TransactionID) (*domainFinance.Transaction, error) {
	return nil, errors.New("not implemented")
}
func (r *memTxRepo) FindByIDAndType(ctx context.Context, id domainFinance.TransactionID, transactionType domainFinance.TransactionType) (*domainFinance.Transaction, error) {
	return nil, errors.New("not implemented")
}
func (r *memTxRepo) FindByUserID(ctx context.Context, userID domainFinance.UserID) ([]*domainFinance.Transaction, error) {
	return nil, nil
}
func (r *memTxRepo) FindByUserIDAndDateRange(ctx context.Context, userID domainFinance.UserID, startDate, endDate time.Time) ([]*domainFinance.Transaction, error) {
	return nil, nil
}
func (r *memTxRepo) FindByUserIDAndCategory(ctx context.Context, userID domainFinance.UserID, categoryID domainFinance.CategoryID) ([]*domainFinance.Transaction, error) {
	return nil, nil
}
func (r *memTxRepo) FindByUserIDWithFilters(ctx context.Context, userID domainFinance.UserID, filters domainFinance.TransactionFilters) ([]*domainFinance.Transaction, int64, error) {
	return nil, 0, nil
}
func (r *memTxRepo) Delete(ctx context.Context, id domainFinance.TransactionID) error {
	return nil
}
func (r *memTxRepo) DeleteByIDAndType(ctx context.Context, id domainFinance.TransactionID, transactionType domainFinance.TransactionType) error {
	return nil
}
func (r *memTxRepo) GetTotalCount(ctx context.Context) (int, error)        { return 0, nil }
func (r *memTxRepo) GetTotalExpenses(ctx context.Context) (float64, error) { return 0, nil }
func (r *memTxRepo) GetTotalIncome(ctx context.Context) (float64, error)   { return 0, nil }
func (r *memTxRepo) GetCountByDateRange(ctx context.Context, startDate, endDate time.Time) (int, error) {
	return 0, nil
}
func (r *memTxRepo) GetExpensesByDateRange(ctx context.Context, startDate, endDate time.Time) (float64, error) {
	return 0, nil
}
func (r *memTxRepo) GetIncomesByDateRange(ctx context.Context, startDate, endDate time.Time) (float64, error) {
	return 0, nil
}

type stubCategoryRepo struct{}

func (s *stubCategoryRepo) Save(ctx context.Context, category *domainFinance.Category) error {
	return nil
}
func (s *stubCategoryRepo) FindByID(ctx context.Context, id domainFinance.CategoryID) (*domainFinance.Category, error) {
	return domainFinance.NewCategory(id, nil, "Test", "#000", true, domainFinance.CategoryTypeExpense)
}
func (s *stubCategoryRepo) FindByUserID(ctx context.Context, userID domainFinance.UserID) ([]*domainFinance.Category, error) {
	return nil, nil
}
func (s *stubCategoryRepo) FindByUserIDAndType(ctx context.Context, userID domainFinance.UserID, categoryType domainFinance.CategoryType) ([]*domainFinance.Category, error) {
	return nil, nil
}
func (s *stubCategoryRepo) FindDefaultCategories(ctx context.Context) ([]*domainFinance.Category, error) {
	return nil, nil
}
func (s *stubCategoryRepo) Delete(ctx context.Context, id domainFinance.CategoryID) error {
	return nil
}
func (s *stubCategoryRepo) ExistsByID(ctx context.Context, id domainFinance.CategoryID) (bool, error) {
	return true, nil
}

type stubCurrencyRepo struct{}

func (s *stubCurrencyRepo) Save(ctx context.Context, currency *domainFinance.Currency) error {
	return nil
}
func (s *stubCurrencyRepo) FindByID(ctx context.Context, id domainFinance.CurrencyID) (*domainFinance.Currency, error) {
	return domainFinance.NewCurrency(id, nil, "IDR", "Rupiah", "Rp", true)
}
func (s *stubCurrencyRepo) FindByUserID(ctx context.Context, userID domainFinance.UserID) ([]*domainFinance.Currency, error) {
	return nil, nil
}
func (s *stubCurrencyRepo) FindDefaultCurrencies(ctx context.Context) ([]*domainFinance.Currency, error) {
	return nil, nil
}
func (s *stubCurrencyRepo) Delete(ctx context.Context, id domainFinance.CurrencyID) error {
	return nil
}
func (s *stubCurrencyRepo) ExistsByID(ctx context.Context, id domainFinance.CurrencyID) (bool, error) {
	return true, nil
}
func (s *stubCurrencyRepo) ExistsByCodeAndUserID(ctx context.Context, code string, userID domainFinance.UserID) (bool, error) {
	return false, nil
}
func (s *stubCurrencyRepo) SetUserDefaultCurrency(ctx context.Context, userID domainFinance.UserID, currencyID domainFinance.CurrencyID) error {
	return nil
}
func (s *stubCurrencyRepo) GetUserDefaultCurrency(ctx context.Context, userID domainFinance.UserID) (*domainFinance.Currency, error) {
	return s.FindByID(ctx, domainFinance.NewCurrencyID(1))
}


type stubWalletRepo struct{}

func (s *stubWalletRepo) Save(ctx context.Context, wallet *domainFinance.Wallet) error { return nil }
func (s *stubWalletRepo) FindByID(ctx context.Context, id domainFinance.WalletID) (*domainFinance.Wallet, error) {
	return domainFinance.ReconstituteWallet(
		domainFinance.NewWalletID(1),
		domainFinance.NewUserID(1),
		"Cash",
		domainFinance.WalletTypeCash,
		domainFinance.NewCurrencyID(1),
		0,
		true,
		false,
		time.Now(),
	), nil
}
func (s *stubWalletRepo) FindByUserID(ctx context.Context, userID domainFinance.UserID, includeArchived bool) ([]*domainFinance.Wallet, error) {
	w, _ := s.FindByID(ctx, domainFinance.NewWalletID(1))
	return []*domainFinance.Wallet{w}, nil
}
func (s *stubWalletRepo) FindDefaultByUserID(ctx context.Context, userID domainFinance.UserID) (*domainFinance.Wallet, error) {
	return s.FindByID(ctx, domainFinance.NewWalletID(1))
}
func (s *stubWalletRepo) CountActiveByUserID(ctx context.Context, userID domainFinance.UserID) (int64, error) {
	return 1, nil
}
func (s *stubWalletRepo) ClearDefaultForUser(ctx context.Context, userID domainFinance.UserID) error {
	return nil
}
func (s *stubWalletRepo) HasTransactions(ctx context.Context, id domainFinance.WalletID) (bool, error) {
	return false, nil
}
func (s *stubWalletRepo) GetBalanceBreakdown(ctx context.Context, id domainFinance.WalletID) (domainFinance.WalletBalanceBreakdown, error) {
	return domainFinance.WalletBalanceBreakdown{}, nil
}
func (s *stubWalletRepo) AlignPendingCurrency(ctx context.Context, id domainFinance.WalletID, currencyID domainFinance.CurrencyID) error {
	return nil
}

func (s *stubWalletRepo) AlignRecurringCurrency(ctx context.Context, id domainFinance.WalletID, currencyID domainFinance.CurrencyID) error {
	return nil
}

func mustDueRecurring(t *testing.T) *domainFinance.RecurringTransaction {
	t.Helper()
	amount, err := domainFinance.NewMoney(150000, domainFinance.NewCurrencyID(1))
	if err != nil {
		t.Fatal(err)
	}
	today := time.Now().Truncate(24 * time.Hour)
	rt, err := domainFinance.NewRecurringTransaction(
		domainFinance.NewUserID(1),
		domainFinance.NewWalletID(1),
		domainFinance.NewCategoryID(1),
		domainFinance.NewCurrencyID(1),
		amount,
		"Subscription",
		domainFinance.FrequencyDaily,
		domainFinance.TransactionTypeExpense,
		domainFinance.RecurringSchedule{},
		today,
	)
	if err != nil {
		t.Fatal(err)
	}
	return rt
}

func TestEnqueueDueRecurringCreatesPendingAndAdvances(t *testing.T) {
	recurringRepo := newMemRecurringRepo()
	pendingRepo := newMemPendingRepo()
	rt := mustDueRecurring(t)
	if err := recurringRepo.Save(context.Background(), rt); err != nil {
		t.Fatal(err)
	}
	originalDue := rt.NextDueDate()

	uc := NewEnqueueDueRecurringUseCase(recurringRepo, pendingRepo, &memPrefsRepo{}, nil)
	if err := uc.Execute(context.Background(), 1); err != nil {
		t.Fatal(err)
	}

	if pendingRepo.openCount() != 1 {
		t.Fatalf("expected 1 pending, got %d", pendingRepo.openCount())
	}
	updated, _ := recurringRepo.FindByID(context.Background(), rt.ID())
	if !updated.NextDueDate().After(originalDue) {
		t.Fatalf("expected next due advanced past %v, got %v", originalDue, updated.NextDueDate())
	}
}

func TestEnqueueDueRecurringDoesNotDuplicate(t *testing.T) {
	recurringRepo := newMemRecurringRepo()
	pendingRepo := newMemPendingRepo()
	rt := mustDueRecurring(t)
	_ = recurringRepo.Save(context.Background(), rt)

	uc := NewEnqueueDueRecurringUseCase(recurringRepo, pendingRepo, &memPrefsRepo{}, nil)
	if err := uc.Execute(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if err := uc.Execute(context.Background(), 1); err != nil {
		t.Fatal(err)
	}
	if pendingRepo.totalCount() < 1 {
		t.Fatal("expected at least one pending row")
	}
	if pendingRepo.openCount() != 1 {
		t.Fatalf("expected 1 open pending after second enqueue, got %d (total %d)", pendingRepo.openCount(), pendingRepo.totalCount())
	}
}

func TestConfirmPendingCreatesTransaction(t *testing.T) {
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

	txService := domainFinance.NewTransactionService(txRepo, &stubCategoryRepo{}, &stubCurrencyRepo{}, &stubWalletRepo{})
	uc := NewConfirmPendingTransactionUseCase(pendingRepo, txService)
	resp, err := uc.Execute(context.Background(), 1, pt.ID().Value())
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != "confirmed" {
		t.Fatalf("got status %s", resp.Status)
	}
	if len(txRepo.saved) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(txRepo.saved))
	}
	if pendingRepo.openCount() != 0 {
		t.Fatal("expected no open pending")
	}
}

func TestRejectPendingDoesNotCreateTransaction(t *testing.T) {
	pendingRepo := newMemPendingRepo()
	amount, _ := domainFinance.NewMoney(99, domainFinance.NewCurrencyID(1))
	pt, err := domainFinance.NewPendingTransaction(
		domainFinance.NewUserID(1),
		domainFinance.NewWalletID(1),
		domainFinance.NewRecurringTransactionID(5),
		time.Now(),
		amount,
		"Skip me",
		domainFinance.TransactionTypeExpense,
		domainFinance.NewCategoryID(1),
		domainFinance.NewCurrencyID(1),
	)
	if err != nil {
		t.Fatal(err)
	}
	_ = pendingRepo.Save(context.Background(), pt)

	uc := NewRejectPendingTransactionUseCase(pendingRepo)
	resp, err := uc.Execute(context.Background(), 1, pt.ID().Value())
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != "rejected" {
		t.Fatalf("got status %s", resp.Status)
	}
	if pendingRepo.openCount() != 0 {
		t.Fatal("expected no open pending")
	}
}
