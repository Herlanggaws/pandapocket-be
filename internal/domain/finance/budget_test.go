package finance

import (
	"testing"
	"time"
)

func TestCalculateBudgetEndDateInclusiveWindows(t *testing.T) {
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	weekly, err := calculateBudgetEndDate(BudgetPeriodWeekly, start)
	if err != nil {
		t.Fatalf("weekly end date error: %v", err)
	}
	if !weekly.Equal(time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected weekly end 2024-01-07, got %s", weekly.Format("2006-01-02"))
	}

	monthly, err := calculateBudgetEndDate(BudgetPeriodMonthly, start)
	if err != nil {
		t.Fatalf("monthly end date error: %v", err)
	}
	if !monthly.Equal(time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected monthly end 2024-01-31, got %s", monthly.Format("2006-01-02"))
	}

	yearly, err := calculateBudgetEndDate(BudgetPeriodYearly, start)
	if err != nil {
		t.Fatalf("yearly end date error: %v", err)
	}
	if !yearly.Equal(time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected yearly end 2024-12-31, got %s", yearly.Format("2006-01-02"))
	}
}

func TestReconstituteBudgetPreservesPersistedFields(t *testing.T) {
	amount, err := NewMoney(100, NewCurrencyID(2))
	if err != nil {
		t.Fatalf("money error: %v", err)
	}

	createdAt := time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC)
	startDate := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)

	budget, err := ReconstituteBudget(
		NewBudgetID(42),
		NewUserID(7),
		NewCategoryID(3),
		amount,
		BudgetLimitFixed,
		nil,
		BudgetPeriodMonthly,
		startDate,
		endDate,
		createdAt,
	)
	if err != nil {
		t.Fatalf("reconstitute error: %v", err)
	}

	if budget.ID().Value() != 42 {
		t.Fatalf("expected id 42, got %d", budget.ID().Value())
	}
	if !budget.CreatedAt().Equal(createdAt) {
		t.Fatalf("created_at was not preserved")
	}
	if !budget.EndDate().Equal(endDate) {
		t.Fatalf("end_date was not preserved")
	}
	if budget.Amount().Currency().Value() != 2 {
		t.Fatalf("currency was not preserved")
	}
}

func TestBudgetAssignID(t *testing.T) {
	amount, _ := NewMoney(50, NewCurrencyID(1))
	budget, err := NewBudget(
		BudgetID{},
		NewUserID(1),
		NewCategoryID(1),
		amount,
		BudgetLimitFixed,
		nil,
		BudgetPeriodWeekly,
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("new budget error: %v", err)
	}

	budget.AssignID(NewBudgetID(99))
	if budget.ID().Value() != 99 {
		t.Fatalf("expected assigned id 99, got %d", budget.ID().Value())
	}
}

func TestBudgetOverlapsWith(t *testing.T) {
	amount, _ := NewMoney(100, NewCurrencyID(1))
	budget, _ := NewBudget(
		NewBudgetID(1),
		NewUserID(1),
		NewCategoryID(1),
		amount,
		BudgetLimitFixed,
		nil,
		BudgetPeriodMonthly,
		time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	)

	if !budget.OverlapsWith(
		time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC),
	) {
		t.Fatal("expected overlapping ranges to overlap")
	}

	if budget.OverlapsWith(
		time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2024, 2, 28, 0, 0, 0, 0, time.UTC),
	) {
		t.Fatal("expected non-overlapping ranges to not overlap")
	}
}

func TestUpdateEndDateRejectsBeforeStart(t *testing.T) {
	amount, _ := NewMoney(100, NewCurrencyID(1))
	budget, _ := NewBudget(
		NewBudgetID(1),
		NewUserID(1),
		NewCategoryID(1),
		amount,
		BudgetLimitFixed,
		nil,
		BudgetPeriodMonthly,
		time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
	)

	err := budget.UpdateEndDate(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	if err == nil {
		t.Fatal("expected end date before start to fail")
	}
}
