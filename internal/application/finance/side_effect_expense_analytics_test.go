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

type sideEffectReceivableRepo struct {
	items  map[int]*domainFinance.Receivable
	nextID int
}

func (r *sideEffectReceivableRepo) Save(_ context.Context, receivable *domainFinance.Receivable) error {
	if r.items == nil {
		r.items = map[int]*domainFinance.Receivable{}
	}
	if receivable.ID().Value() == 0 {
		r.nextID++
		receivable.AssignID(domainFinance.NewReceivableID(r.nextID))
	}
	r.items[receivable.ID().Value()] = receivable
	return nil
}

func (r *sideEffectReceivableRepo) FindByID(_ context.Context, id domainFinance.ReceivableID) (*domainFinance.Receivable, error) {
	item, ok := r.items[id.Value()]
	if !ok {
		return nil, errors.New("not found")
	}
	return item, nil
}

func (r *sideEffectReceivableRepo) FindByUserID(context.Context, domainFinance.UserID, bool) ([]*domainFinance.Receivable, error) {
	out := make([]*domainFinance.Receivable, 0, len(r.items))
	for _, item := range r.items {
		out = append(out, item)
	}
	return out, nil
}

func (r *sideEffectReceivableRepo) FindByCreateExpenseID(context.Context, int) (*domainFinance.Receivable, error) {
	return nil, errors.New("not found")
}

type sideEffectCollectionRepo struct{}

func (r *sideEffectCollectionRepo) Save(context.Context, *domainFinance.ReceivableCollection) error {
	return nil
}
func (r *sideEffectCollectionRepo) FindByReceivableID(context.Context, domainFinance.ReceivableID) ([]*domainFinance.ReceivableCollection, error) {
	return nil, nil
}
func (r *sideEffectCollectionRepo) FindByIncomeID(context.Context, int) (*domainFinance.ReceivableCollection, error) {
	return nil, errors.New("not found")
}
func (r *sideEffectCollectionRepo) Delete(context.Context, domainFinance.ReceivableCollectionID) error {
	return nil
}

type sideEffectLiabilityRepo struct {
	byID map[int]*domainFinance.Liability
}

func (r *sideEffectLiabilityRepo) Save(_ context.Context, liability *domainFinance.Liability) error {
	if r.byID == nil {
		r.byID = map[int]*domainFinance.Liability{}
	}
	r.byID[liability.ID().Value()] = liability
	return nil
}

func (r *sideEffectLiabilityRepo) FindByID(_ context.Context, id domainFinance.LiabilityID) (*domainFinance.Liability, error) {
	item, ok := r.byID[id.Value()]
	if !ok {
		return nil, errors.New("not found")
	}
	return item, nil
}

func (r *sideEffectLiabilityRepo) FindByUserID(context.Context, domainFinance.UserID, bool) ([]*domainFinance.Liability, error) {
	return nil, nil
}

type sideEffectPaymentRepo struct {
	nextID int
}

func (r *sideEffectPaymentRepo) Save(_ context.Context, payment *domainFinance.LiabilityPayment) error {
	if payment.ID().Value() == 0 {
		r.nextID++
		payment.AssignID(domainFinance.NewLiabilityPaymentID(r.nextID))
	}
	return nil
}
func (r *sideEffectPaymentRepo) FindByLiabilityID(context.Context, domainFinance.LiabilityID) ([]*domainFinance.LiabilityPayment, error) {
	return nil, nil
}
func (r *sideEffectPaymentRepo) FindByExpenseID(context.Context, int) (*domainFinance.LiabilityPayment, error) {
	return nil, errors.New("not found")
}
func (r *sideEffectPaymentRepo) Delete(context.Context, domainFinance.LiabilityPaymentID) error {
	return nil
}

func newSideEffectCreateTxStack(t *testing.T) (*CreateTransactionUseCase, *domainFinance.CategoryService) {
	t.Helper()
	idr, err := domainFinance.NewCurrency(domainFinance.NewCurrencyID(1), nil, "IDR", "Rupiah", "Rp", true)
	if err != nil {
		t.Fatalf("currency: %v", err)
	}
	cat, err := domainFinance.NewCategory(domainFinance.NewCategoryID(1), nil, "Debt", "#000", true, domainFinance.CategoryTypeExpense)
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
	createTx := NewCreateTransactionUseCase(txService, walletService, entitlement.StaticChecker{Pro: true})
	categoryService := domainFinance.NewCategoryService(catRepo, nil)
	return createTx, categoryService
}

func TestCreateReceivable_ExpenseAddedAnalytics(t *testing.T) {
	createTx, categoryService := newSideEffectCreateTxStack(t)
	receivableRepo := &sideEffectReceivableRepo{}
	uc := NewCreateReceivableUseCase(
		domainFinance.NewReceivableService(receivableRepo, &sideEffectCollectionRepo{}),
		createTx,
		categoryService,
		entitlement.StaticChecker{Pro: true},
	)

	t.Run("captures when create_expense", func(t *testing.T) {
		var captures []struct {
			event string
			props map[string]interface{}
		}
		analytics.SetTestCapture(func(userID int, event string, properties map[string]interface{}) {
			captures = append(captures, struct {
				event string
				props map[string]interface{}
			}{event, properties})
		})
		t.Cleanup(func() { analytics.SetTestCapture(nil) })

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

		found := false
		for _, c := range captures {
			if c.event != "expense_added" {
				continue
			}
			found = true
			if got, _ := c.props["source"].(string); got != "receivable" {
				t.Fatalf("source=%q, want receivable", got)
			}
			if got, _ := c.props["category_id"].(int); got != 1 {
				t.Fatalf("category_id=%v, want 1", c.props["category_id"])
			}
			if _, ok := c.props["amount"]; ok {
				t.Fatal("amount must not be in analytics properties")
			}
			if _, ok := c.props["notes"]; ok {
				t.Fatal("notes must not be in analytics properties")
			}
			if _, ok := c.props["name"]; ok {
				t.Fatal("name must not be in analytics properties")
			}
		}
		if !found {
			t.Fatal("expected expense_added capture")
		}
	})

	t.Run("skips when create_expense false", func(t *testing.T) {
		calledExpenseAdded := false
		analytics.SetTestCapture(func(userID int, event string, properties map[string]interface{}) {
			if event == "expense_added" {
				calledExpenseAdded = true
			}
		})
		t.Cleanup(func() { analytics.SetTestCapture(nil) })

		_, err := uc.Execute(context.Background(), 1, CreateReceivableRequest{
			Name:           "Invoice only",
			Type:           "invoice",
			CurrencyID:     1,
			CurrentBalance: 25000,
			CreateExpense:  false,
		})
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if calledExpenseAdded {
			t.Fatal("expected no expense_added when create_expense is false")
		}
	})
}

func TestRecordLiabilityPayment_ExpenseAddedAnalytics(t *testing.T) {
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

	t.Run("captures when create_expense", func(t *testing.T) {
		var captures []struct {
			event string
			props map[string]interface{}
		}
		analytics.SetTestCapture(func(userID int, event string, properties map[string]interface{}) {
			captures = append(captures, struct {
				event string
				props map[string]interface{}
			}{event, properties})
		})
		t.Cleanup(func() { analytics.SetTestCapture(nil) })

		_, err := uc.Execute(context.Background(), 1, 10, RecordLiabilityPaymentRequest{
			Amount:        10000,
			PaidAt:        "2024-06-01",
			CreateExpense: true,
			CategoryID:    intPtr(1),
		})
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}

		found := false
		for _, c := range captures {
			if c.event != "expense_added" {
				continue
			}
			found = true
			if got, _ := c.props["source"].(string); got != "debt" {
				t.Fatalf("source=%q, want debt", got)
			}
			if got, _ := c.props["category_id"].(int); got != 1 {
				t.Fatalf("category_id=%v, want 1", c.props["category_id"])
			}
			if _, ok := c.props["amount"]; ok {
				t.Fatal("amount must not be in analytics properties")
			}
			if _, ok := c.props["note"]; ok {
				t.Fatal("note must not be in analytics properties")
			}
		}
		if !found {
			t.Fatal("expected expense_added capture")
		}
	})

	t.Run("skips when create_expense false", func(t *testing.T) {
		fresh := domainFinance.ReconstituteLiability(
			domainFinance.NewLiabilityID(11),
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
		liabilityRepo.byID[11] = fresh

		calledExpenseAdded := false
		analytics.SetTestCapture(func(userID int, event string, properties map[string]interface{}) {
			if event == "expense_added" {
				calledExpenseAdded = true
			}
		})
		t.Cleanup(func() { analytics.SetTestCapture(nil) })

		_, err := uc.Execute(context.Background(), 1, 11, RecordLiabilityPaymentRequest{
			Amount:        5000,
			PaidAt:        "2024-06-02",
			CreateExpense: false,
		})
		if err != nil {
			t.Fatalf("Execute: %v", err)
		}
		if calledExpenseAdded {
			t.Fatal("expected no expense_added when create_expense is false")
		}
	})
}

func intPtr(v int) *int { return &v }
