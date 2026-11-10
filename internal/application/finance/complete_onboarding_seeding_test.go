package finance

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"panda-pocket/internal/domain/finance"
	domainIdentity "panda-pocket/internal/domain/identity"
)

type stubOnboardingPrefsRepo struct {
	prefs *domainIdentity.UserPreferences
	saved *domainIdentity.UserPreferences
}

func (r *stubOnboardingPrefsRepo) FindByUserID(ctx context.Context, userID domainIdentity.UserID) (*domainIdentity.UserPreferences, error) {
	return r.prefs, nil
}

func (r *stubOnboardingPrefsRepo) Save(ctx context.Context, prefs *domainIdentity.UserPreferences) error {
	r.saved = prefs
	return nil
}

type stubOnboardingWallet struct{}

func (s *stubOnboardingWallet) EnsureDefaultWallet(
	ctx context.Context,
	userID finance.UserID,
	currencyID finance.CurrencyID,
) (*finance.Wallet, bool, error) {
	wallet := finance.ReconstituteWallet(
		finance.NewWalletID(1),
		userID,
		"Cash",
		finance.WalletTypeCash,
		currencyID,
		0,
		true,
		false,
		time.Now(),
	)
	return wallet, false, nil
}

type stubOnboardingCategories struct {
	income  *finance.Category
	expense *finance.Category
}

func (s *stubOnboardingCategories) GetCategoriesByUserAndType(
	ctx context.Context,
	userID finance.UserID,
	categoryType finance.CategoryType,
) ([]*finance.Category, error) {
	switch categoryType {
	case finance.CategoryTypeIncome:
		return []*finance.Category{s.income}, nil
	case finance.CategoryTypeExpense:
		return []*finance.Category{s.expense}, nil
	default:
		return nil, nil
	}
}

type spyRecurringCreator struct {
	calls []CreateRecurringTransactionRequest
}

func (s *spyRecurringCreator) Execute(
	ctx context.Context,
	userID int,
	req CreateRecurringTransactionRequest,
) (*RecurringTransactionResponse, error) {
	s.calls = append(s.calls, req)
	return &RecurringTransactionResponse{ID: len(s.calls), Type: req.Type, CategoryID: req.CategoryID}, nil
}

type spyBudgetCreator struct {
	calls []CreateBudgetRequest
}

func (s *spyBudgetCreator) Execute(ctx context.Context, userID int, req CreateBudgetRequest) (*BudgetResponse, error) {
	s.calls = append(s.calls, req)
	return &BudgetResponse{ID: len(s.calls), Amount: req.Amount}, nil
}

type spyDueEnqueuer struct {
	calls int
}

func (s *spyDueEnqueuer) Execute(ctx context.Context, userID int) error {
	s.calls++
	return nil
}

func newTestOnboardingCategory(id int, name string, categoryType finance.CategoryType) *finance.Category {
	cat, err := finance.NewCategory(finance.NewCategoryID(id), nil, name, "#000", true, categoryType)
	if err != nil {
		panic(err)
	}
	return cat
}

func TestCompleteOnboardingDoesNotSeedExpensePending(t *testing.T) {
	prefs := domainIdentity.NewUserPreferences(
		domainIdentity.NewUserID(42),
		1,
		true,
		true,
		true,
		true,
		json.RawMessage(`{}`),
	)

	recurring := &spyRecurringCreator{}
	budget := &spyBudgetCreator{}
	enqueue := &spyDueEnqueuer{}
	prefsRepo := &stubOnboardingPrefsRepo{prefs: prefs}

	uc := NewCompleteOnboardingUseCase(
		prefsRepo,
		&stubOnboardingWallet{},
		nil,
		&stubOnboardingCategories{
			income:  newTestOnboardingCategory(1, "Salary", finance.CategoryTypeIncome),
			expense: newTestOnboardingCategory(8, "Other", finance.CategoryTypeExpense),
		},
		recurring,
		enqueue,
		budget,
		nil,
		nil,
	)

	_, err := uc.Execute(context.Background(), 42, CompleteOnboardingRequest{
		Goals:          []string{"budget"},
		MonthlyIncome:  5_000_000,
		MonthlyExpense: 4_000_000,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(recurring.calls) != 1 {
		t.Fatalf("recurring calls=%d, want 1 (income only)", len(recurring.calls))
	}
	if recurring.calls[0].Type != "income" {
		t.Fatalf("recurring type=%q, want income", recurring.calls[0].Type)
	}
	if recurring.calls[0].Description != "Onboarding monthly income" {
		t.Fatalf("recurring description=%q", recurring.calls[0].Description)
	}
	for _, call := range recurring.calls {
		if call.Type == "expense" || call.Description == "Onboarding monthly expense estimate" {
			t.Fatalf("must not seed expense recurring: %+v", call)
		}
	}

	if len(budget.calls) != 1 {
		t.Fatalf("budget calls=%d, want 1", len(budget.calls))
	}
	if budget.calls[0].Amount != 4_000_000 || budget.calls[0].CategoryID != 8 {
		t.Fatalf("budget call=%+v", budget.calls[0])
	}

	if enqueue.calls != 1 {
		t.Fatalf("enqueue calls=%d, want 1 (income pending seeded)", enqueue.calls)
	}

	if prefsRepo.saved == nil {
		t.Fatal("expected preferences saved")
	}
	var onboarding map[string]interface{}
	if err := json.Unmarshal(prefsRepo.saved.Onboarding(), &onboarding); err != nil {
		t.Fatalf("unmarshal onboarding: %v", err)
	}
	if onboarding["monthly_expense"] != float64(4_000_000) {
		t.Fatalf("monthly_expense not stored: %v", onboarding["monthly_expense"])
	}
	if onboarding["monthly_income"] != float64(5_000_000) {
		t.Fatalf("monthly_income not stored: %v", onboarding["monthly_income"])
	}
}

func TestCompleteOnboardingExpenseOnlyStillCreatesBudgetWithoutEnqueue(t *testing.T) {
	prefs := domainIdentity.NewUserPreferences(
		domainIdentity.NewUserID(7),
		1,
		true,
		true,
		true,
		true,
		json.RawMessage(`{}`),
	)

	recurring := &spyRecurringCreator{}
	budget := &spyBudgetCreator{}
	enqueue := &spyDueEnqueuer{}

	uc := NewCompleteOnboardingUseCase(
		&stubOnboardingPrefsRepo{prefs: prefs},
		&stubOnboardingWallet{},
		nil,
		&stubOnboardingCategories{
			income:  newTestOnboardingCategory(1, "Salary", finance.CategoryTypeIncome),
			expense: newTestOnboardingCategory(8, "Other", finance.CategoryTypeExpense),
		},
		recurring,
		enqueue,
		budget,
		nil,
		nil,
	)

	_, err := uc.Execute(context.Background(), 7, CompleteOnboardingRequest{
		Goals:          []string{"track"},
		MonthlyIncome:  0,
		MonthlyExpense: 1_000_000,
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(recurring.calls) != 0 {
		t.Fatalf("expected no recurring seeds, got %+v", recurring.calls)
	}
	if len(budget.calls) != 1 {
		t.Fatalf("budget calls=%d, want 1", len(budget.calls))
	}
	if enqueue.calls != 0 {
		t.Fatalf("enqueue calls=%d, want 0 when only expense estimate (no pending)", enqueue.calls)
	}
}
