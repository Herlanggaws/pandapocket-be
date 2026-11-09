package finance

import (
	"testing"
	"time"
)

func TestCalendarPeriodWindowWeeklyMondaySunday(t *testing.T) {
	// Wednesday 2026-11-11 → Mon 9 Nov – Sun 15 Nov
	now := time.Date(2026, 11, 11, 15, 0, 0, 0, time.UTC)
	start, end := CalendarPeriodWindow(BudgetPeriodWeekly, now)
	if !start.Equal(time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("weekly start: got %s", start.Format("2006-01-02"))
	}
	if !end.Equal(time.Date(2026, 11, 15, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("weekly end: got %s", end.Format("2006-01-02"))
	}
}

func TestCalendarPeriodWindowMonthly(t *testing.T) {
	now := time.Date(2026, 11, 9, 12, 0, 0, 0, time.UTC)
	start, end := CalendarPeriodWindow(BudgetPeriodMonthly, now)
	if !start.Equal(time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("monthly start: got %s", start.Format("2006-01-02"))
	}
	if !end.Equal(time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("monthly end: got %s", end.Format("2006-01-02"))
	}
}

func TestCalendarPeriodWindowYearly(t *testing.T) {
	now := time.Date(2026, 11, 9, 12, 0, 0, 0, time.UTC)
	start, end := CalendarPeriodWindow(BudgetPeriodYearly, now)
	if !start.Equal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("yearly start: got %s", start.Format("2006-01-02"))
	}
	if !end.Equal(time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("yearly end: got %s", end.Format("2006-01-02"))
	}
}

func TestEffectiveBudgetWindowIntersectsLongMonthly(t *testing.T) {
	amount, _ := NewMoney(5_000_000, NewCurrencyID(1))
	budget, err := ReconstituteBudget(
		NewBudgetID(1),
		NewUserID(1),
		NewCategoryID(1),
		amount,
		BudgetLimitFixed,
		nil,
		BudgetPeriodMonthly,
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2027, 11, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("budget: %v", err)
	}

	now := time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC)
	start, end := EffectiveBudgetWindow(budget, now)
	if !start.Equal(time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected Nov 1, got %s", start.Format("2006-01-02"))
	}
	if !end.Equal(time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("expected Nov 30, got %s", end.Format("2006-01-02"))
	}
}

func TestEffectiveBudgetWindowFallbackWhenExpired(t *testing.T) {
	amount, _ := NewMoney(100, NewCurrencyID(1))
	budget, err := ReconstituteBudget(
		NewBudgetID(1),
		NewUserID(1),
		NewCategoryID(1),
		amount,
		BudgetLimitFixed,
		nil,
		BudgetPeriodMonthly,
		time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("budget: %v", err)
	}

	now := time.Date(2026, 11, 9, 0, 0, 0, 0, time.UTC)
	start, end := EffectiveBudgetWindow(budget, now)
	if !start.Equal(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("fallback start: got %s", start.Format("2006-01-02"))
	}
	if !end.Equal(time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)) {
		t.Fatalf("fallback end: got %s", end.Format("2006-01-02"))
	}
}
