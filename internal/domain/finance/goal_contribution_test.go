package finance

import (
	"testing"
	"time"
)

func TestApplyContributionIncreasesProgress(t *testing.T) {
	g, err := NewFinancialGoal(NewUserID(1), "Trip", 1000, NewCurrencyID(1), 200, time.Now().AddDate(0, 1, 0))
	if err != nil {
		t.Fatal(err)
	}
	if err := g.ApplyContribution(300); err != nil {
		t.Fatalf("ApplyContribution: %v", err)
	}
	if g.CurrentAmount() != 500 {
		t.Fatalf("want 500, got %v", g.CurrentAmount())
	}
	if g.ProgressPercent() != 50 {
		t.Fatalf("want 50%%, got %v", g.ProgressPercent())
	}
}

func TestReverseContributionRestoresAndReactivates(t *testing.T) {
	g, err := NewFinancialGoal(NewUserID(1), "Trip", 1000, NewCurrencyID(1), 1000, time.Now().AddDate(0, 1, 0))
	if err != nil {
		t.Fatal(err)
	}
	if g.Status() != GoalStatusCompleted {
		t.Fatalf("want completed, got %s", g.Status())
	}
	if err := g.ReverseContribution(400); err != nil {
		t.Fatalf("ReverseContribution: %v", err)
	}
	if g.CurrentAmount() != 600 {
		t.Fatalf("want 600, got %v", g.CurrentAmount())
	}
	if g.Status() != GoalStatusActive {
		t.Fatalf("want active after reverse, got %s", g.Status())
	}
}

func TestReverseContributionFloorsAtZero(t *testing.T) {
	g := &FinancialGoal{currentAmount: 50, targetAmount: 100, status: GoalStatusActive}
	if err := g.ReverseContribution(80); err != nil {
		t.Fatal(err)
	}
	if g.CurrentAmount() != 0 {
		t.Fatalf("want 0, got %v", g.CurrentAmount())
	}
}

func TestNewGoalContributionRequiresExactlyOneRef(t *testing.T) {
	expenseID := 1
	incomeID := 2
	_, err := NewGoalContribution(NewGoalID(1), NewUserID(1), 100, time.Now(), &expenseID, &incomeID, nil, "")
	if err == nil {
		t.Fatal("expected error for two refs")
	}
	_, err = NewGoalContribution(NewGoalID(1), NewUserID(1), 100, time.Now(), nil, nil, nil, "")
	if err == nil {
		t.Fatal("expected error for zero refs")
	}
	c, err := NewGoalContribution(NewGoalID(1), NewUserID(1), 100, time.Now(), &expenseID, nil, nil, "note")
	if err != nil {
		t.Fatal(err)
	}
	if !c.BumpsCurrentAmount() {
		t.Fatal("expense contribution should bump current amount")
	}
}
