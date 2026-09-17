package finance

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubBudgetRepo struct {
	budgets []*Budget
	saved   *Budget
}

func (r *stubBudgetRepo) Save(ctx context.Context, budget *Budget) error {
	if budget.ID().Value() == 0 {
		budget.AssignID(NewBudgetID(len(r.budgets) + 1))
	}
	r.saved = budget
	found := false
	for i, existing := range r.budgets {
		if existing.ID().Value() == budget.ID().Value() {
			r.budgets[i] = budget
			found = true
			break
		}
	}
	if !found {
		r.budgets = append(r.budgets, budget)
	}
	return nil
}

func (r *stubBudgetRepo) FindByID(ctx context.Context, id BudgetID) (*Budget, error) {
	for _, budget := range r.budgets {
		if budget.ID().Value() == id.Value() {
			return budget, nil
		}
	}
	return nil, errors.New("not found")
}

func (r *stubBudgetRepo) FindByUserID(ctx context.Context, userID UserID) ([]*Budget, error) {
	var result []*Budget
	for _, budget := range r.budgets {
		if budget.UserID().Value() == userID.Value() {
			result = append(result, budget)
		}
	}
	return result, nil
}

func (r *stubBudgetRepo) FindByUserIDAndCategory(ctx context.Context, userID UserID, categoryID CategoryID) ([]*Budget, error) {
	var result []*Budget
	for _, budget := range r.budgets {
		if budget.UserID().Value() == userID.Value() && budget.CategoryID().Value() == categoryID.Value() {
			result = append(result, budget)
		}
	}
	return result, nil
}

func (r *stubBudgetRepo) FindActiveByUserID(ctx context.Context, userID UserID) ([]*Budget, error) {
	return r.FindByUserID(ctx, userID)
}

func (r *stubBudgetRepo) Delete(ctx context.Context, id BudgetID) error {
	return nil
}

func (r *stubBudgetRepo) DeleteByIDAndUserID(ctx context.Context, id BudgetID, userID UserID) error {
	for i, budget := range r.budgets {
		if budget.ID().Value() == id.Value() && budget.UserID().Value() == userID.Value() {
			r.budgets = append(r.budgets[:i], r.budgets[i+1:]...)
			return nil
		}
	}
	return errors.New("not found")
}

func (r *stubBudgetRepo) GetTotalCount(ctx context.Context) (int, error) {
	return len(r.budgets), nil
}

func (r *stubBudgetRepo) GetCountByDateRange(ctx context.Context, startDate, endDate time.Time) (int, error) {
	return len(r.budgets), nil
}

type stubCategoryRepo struct {
	categories map[int]*Category
}

func (r *stubCategoryRepo) Save(ctx context.Context, category *Category) error {
	return nil
}

func (r *stubCategoryRepo) FindByID(ctx context.Context, id CategoryID) (*Category, error) {
	category, ok := r.categories[id.Value()]
	if !ok {
		return nil, errors.New("not found")
	}
	return category, nil
}

func (r *stubCategoryRepo) FindByUserID(ctx context.Context, userID UserID) ([]*Category, error) {
	return nil, nil
}

func (r *stubCategoryRepo) FindByUserIDAndType(ctx context.Context, userID UserID, categoryType CategoryType) ([]*Category, error) {
	return nil, nil
}

func (r *stubCategoryRepo) FindDefaultCategories(ctx context.Context) ([]*Category, error) {
	return nil, nil
}

func (r *stubCategoryRepo) Delete(ctx context.Context, id CategoryID) error {
	return nil
}

func (r *stubCategoryRepo) ExistsByID(ctx context.Context, id CategoryID) (bool, error) {
	_, ok := r.categories[id.Value()]
	return ok, nil
}

func newExpenseCategory(id int) *Category {
	category, _ := NewCategory(NewCategoryID(id), nil, "Food", "#fff", true, CategoryTypeExpense)
	return category
}

func newIncomeCategory(id int) *Category {
	category, _ := NewCategory(NewCategoryID(id), nil, "Salary", "#fff", true, CategoryTypeIncome)
	return category
}

func TestBudgetServiceCreateRejectsIncomeCategory(t *testing.T) {
	budgetRepo := &stubBudgetRepo{}
	categoryRepo := &stubCategoryRepo{
		categories: map[int]*Category{
			2: newIncomeCategory(2),
		},
	}
	service := NewBudgetService(budgetRepo, categoryRepo)
	amount, _ := NewMoney(100, NewCurrencyID(1))

	_, err := service.CreateBudget(
		context.Background(),
		NewUserID(1),
		NewCategoryID(2),
		amount,
		BudgetLimitFixed,
		nil,
		BudgetPeriodMonthly,
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	)
	if err == nil {
		t.Fatal("expected income category to be rejected")
	}
}

func TestBudgetServiceCreateRejectsOverlap(t *testing.T) {
	amount, _ := NewMoney(100, NewCurrencyID(1))
	existing, _ := NewBudget(
		NewBudgetID(1),
		NewUserID(1),
		NewCategoryID(1),
		amount,
		BudgetLimitFixed,
		nil,
		BudgetPeriodMonthly,
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	)

	budgetRepo := &stubBudgetRepo{budgets: []*Budget{existing}}
	categoryRepo := &stubCategoryRepo{
		categories: map[int]*Category{
			1: newExpenseCategory(1),
		},
	}
	service := NewBudgetService(budgetRepo, categoryRepo)

	_, err := service.CreateBudget(
		context.Background(),
		NewUserID(1),
		NewCategoryID(1),
		amount,
		BudgetLimitFixed,
		nil,
		BudgetPeriodMonthly,
		time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
	)
	if err == nil {
		t.Fatal("expected overlapping budget to be rejected")
	}
}

func TestBudgetServiceCreateAssignsID(t *testing.T) {
	budgetRepo := &stubBudgetRepo{}
	categoryRepo := &stubCategoryRepo{
		categories: map[int]*Category{
			1: newExpenseCategory(1),
		},
	}
	service := NewBudgetService(budgetRepo, categoryRepo)
	amount, _ := NewMoney(100, NewCurrencyID(1))

	budget, err := service.CreateBudget(
		context.Background(),
		NewUserID(1),
		NewCategoryID(1),
		amount,
		BudgetLimitFixed,
		nil,
		BudgetPeriodMonthly,
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("create error: %v", err)
	}
	if budget.ID().Value() == 0 {
		t.Fatal("expected created budget to have assigned id")
	}
}

func TestBudgetServiceDeleteReturnsNotFoundForOtherUser(t *testing.T) {
	amount, _ := NewMoney(100, NewCurrencyID(1))
	existing, _ := NewBudget(
		NewBudgetID(1),
		NewUserID(1),
		NewCategoryID(1),
		amount,
		BudgetLimitFixed,
		nil,
		BudgetPeriodMonthly,
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	)
	budgetRepo := &stubBudgetRepo{budgets: []*Budget{existing}}
	service := NewBudgetService(budgetRepo, &stubCategoryRepo{categories: map[int]*Category{}})

	err := service.DeleteBudget(context.Background(), NewBudgetID(1), NewUserID(2))
	if err == nil || err.Error() != "budget not found" {
		t.Fatalf("expected budget not found, got %v", err)
	}
}

func TestCategoryServiceDeleteBlockedByBudgets(t *testing.T) {
	userID := NewUserID(1)
	amount, _ := NewMoney(100, NewCurrencyID(1))
	existing, _ := NewBudget(
		NewBudgetID(1),
		userID,
		NewCategoryID(5),
		amount,
		BudgetLimitFixed,
		nil,
		BudgetPeriodMonthly,
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	)

	owned, _ := NewCategory(NewCategoryID(5), &userID, "Food", "#fff", false, CategoryTypeExpense)
	categoryRepo := &stubCategoryRepo{categories: map[int]*Category{5: owned}}
	budgetRepo := &stubBudgetRepo{budgets: []*Budget{existing}}
	service := NewCategoryService(categoryRepo, budgetRepo)

	err := service.DeleteCategory(context.Background(), NewCategoryID(5), userID)
	if err == nil {
		t.Fatal("expected delete to be blocked by existing budgets")
	}
}
