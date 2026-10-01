package finance

import (
	"context"
	"testing"

	"panda-pocket/internal/domain/entitlement"
	domainFinance "panda-pocket/internal/domain/finance"
)

type stubMonthScorer struct {
	score *HealthScoreResponse
}

func (s stubMonthScorer) Execute(context.Context, int) (*HealthScoreResponse, error) {
	return s.score, nil
}

type stubSnapshotRepo struct {
	byMonth map[string]*domainFinance.HealthScoreSnapshot
	history []*domainFinance.HealthScoreSnapshot
}

func (r stubSnapshotRepo) Upsert(context.Context, *domainFinance.HealthScoreSnapshot) error {
	return nil
}

func (r stubSnapshotRepo) FindByUserID(context.Context, domainFinance.UserID, int) ([]*domainFinance.HealthScoreSnapshot, error) {
	return r.history, nil
}

func (r stubSnapshotRepo) FindByYearMonth(_ context.Context, _ domainFinance.UserID, yearMonth string) (*domainFinance.HealthScoreSnapshot, error) {
	return r.byMonth[yearMonth], nil
}

func monthCloseFixture() (*HealthScoreResponse, stubSnapshotRepo) {
	current := &HealthScoreResponse{
		Score: 70,
		Components: HealthScoreComponents{
			BudgetAdherence: 80,
			Cashflow:        60,
			Coverage:        40,
		},
		YearMonth: "2026-10",
	}
	previous := domainFinance.NewHealthScoreSnapshot(
		domainFinance.NewUserID(1),
		"2026-09",
		60,
		50,
		60,
		40,
	)
	best := domainFinance.NewHealthScoreSnapshot(
		domainFinance.NewUserID(1),
		"2026-08",
		90,
		90,
		90,
		90,
	)
	repo := stubSnapshotRepo{
		byMonth: map[string]*domainFinance.HealthScoreSnapshot{"2026-09": previous},
		history: []*domainFinance.HealthScoreSnapshot{previous, best},
	}
	return current, repo
}

func TestMonthCloseDeltaAndLockedBestMonth(t *testing.T) {
	current, repo := monthCloseFixture()
	uc := NewGetHealthScoreMonthCloseUseCase(stubMonthScorer{score: current}, repo, entitlement.StaticChecker{Pro: false})
	response, err := uc.Execute(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if response.Previous == nil || response.ScoreDelta == nil || *response.ScoreDelta != 10 {
		t.Fatalf("delta %+v", response)
	}
	if response.MovedMost != "budget_adherence" || response.ComponentDelta.BudgetAdherence != 30 {
		t.Fatalf("moved %+v", response.ComponentDelta)
	}
	if response.BestMonth != nil || !response.BestMonthLocked {
		t.Fatalf("free best month %+v locked=%v", response.BestMonth, response.BestMonthLocked)
	}
}

func TestMonthCloseBestMonthForPro(t *testing.T) {
	current, repo := monthCloseFixture()
	uc := NewGetHealthScoreMonthCloseUseCase(stubMonthScorer{score: current}, repo, entitlement.StaticChecker{Pro: true})
	response, err := uc.Execute(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if response.BestMonth == nil || response.BestMonth.YearMonth != "2026-08" || response.BestMonth.Score != 90 {
		t.Fatalf("best %+v", response.BestMonth)
	}
	if response.BestMonthLocked {
		t.Fatal("pro best month should be unlocked")
	}
}
