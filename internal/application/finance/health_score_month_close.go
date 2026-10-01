package finance

import (
	"context"
	"math"
	"time"

	"panda-pocket/internal/domain/entitlement"
	domainFinance "panda-pocket/internal/domain/finance"
)

type currentMonthScorer interface {
	Execute(ctx context.Context, userID int) (*HealthScoreResponse, error)
}

type MonthCloseMonth struct {
	YearMonth  string                `json:"year_month"`
	Score      int                   `json:"score"`
	Components HealthScoreComponents `json:"components"`
}

type MonthCloseDelta struct {
	BudgetAdherence float64 `json:"budget_adherence"`
	Cashflow        float64 `json:"cashflow"`
	Coverage        float64 `json:"coverage"`
}

type HealthScoreMonthCloseResponse struct {
	Current         MonthCloseMonth  `json:"current"`
	Previous        *MonthCloseMonth `json:"previous"`
	ScoreDelta      *int             `json:"score_delta"`
	ComponentDelta  *MonthCloseDelta `json:"component_delta"`
	MovedMost       string           `json:"moved_most"`
	BestMonth       *MonthCloseMonth `json:"best_month"`
	BestMonthLocked bool             `json:"best_month_locked"`
}

type GetHealthScoreMonthCloseUseCase struct {
	scorer       currentMonthScorer
	snapshots    domainFinance.HealthScoreSnapshotRepository
	entitlements entitlement.Checker
}

func NewGetHealthScoreMonthCloseUseCase(
	scorer currentMonthScorer,
	snapshots domainFinance.HealthScoreSnapshotRepository,
	entitlements entitlement.Checker,
) *GetHealthScoreMonthCloseUseCase {
	return &GetHealthScoreMonthCloseUseCase{
		scorer:       scorer,
		snapshots:    snapshots,
		entitlements: entitlements,
	}
}

func (uc *GetHealthScoreMonthCloseUseCase) Execute(ctx context.Context, userID int) (*HealthScoreMonthCloseResponse, error) {
	current, err := uc.scorer.Execute(ctx, userID)
	if err != nil {
		return nil, err
	}
	isPro, err := uc.entitlements.IsPro(ctx, userID)
	if err != nil {
		return nil, err
	}

	currentMonth := monthFromScore(current)
	response := &HealthScoreMonthCloseResponse{
		Current:         currentMonth,
		BestMonthLocked: !isPro,
	}
	if err := uc.attachPrevious(ctx, userID, current.YearMonth, currentMonth, response); err != nil {
		return nil, err
	}
	if !isPro {
		return response, nil
	}
	best, err := uc.bestMonth(ctx, userID, currentMonth)
	if err != nil {
		return nil, err
	}
	response.BestMonth = &best
	return response, nil
}

func (uc *GetHealthScoreMonthCloseUseCase) attachPrevious(
	ctx context.Context,
	userID int,
	yearMonth string,
	current MonthCloseMonth,
	response *HealthScoreMonthCloseResponse,
) error {
	previousKey, err := previousYearMonth(yearMonth)
	if err != nil {
		return err
	}
	snapshot, err := uc.snapshots.FindByYearMonth(ctx, domainFinance.NewUserID(userID), previousKey)
	if err != nil || snapshot == nil {
		return err
	}
	previous := monthFromSnapshot(snapshot)
	delta := current.Score - previous.Score
	components := MonthCloseDelta{
		BudgetAdherence: round2(current.Components.BudgetAdherence - previous.Components.BudgetAdherence),
		Cashflow:        round2(current.Components.Cashflow - previous.Components.Cashflow),
		Coverage:        round2(current.Components.Coverage - previous.Components.Coverage),
	}
	response.Previous = &previous
	response.ScoreDelta = &delta
	response.ComponentDelta = &components
	response.MovedMost = movedMost(components)
	return nil
}

func (uc *GetHealthScoreMonthCloseUseCase) bestMonth(
	ctx context.Context,
	userID int,
	current MonthCloseMonth,
) (MonthCloseMonth, error) {
	history, err := uc.snapshots.FindByUserID(ctx, domainFinance.NewUserID(userID), 12)
	if err != nil {
		return MonthCloseMonth{}, err
	}
	best := current
	for _, snapshot := range history {
		month := monthFromSnapshot(snapshot)
		if month.YearMonth == current.YearMonth {
			continue
		}
		if month.Score > best.Score || (month.Score == best.Score && month.YearMonth > best.YearMonth) {
			best = month
		}
	}
	return best, nil
}

func monthFromScore(score *HealthScoreResponse) MonthCloseMonth {
	return MonthCloseMonth{
		YearMonth:  score.YearMonth,
		Score:      score.Score,
		Components: score.Components,
	}
}

func monthFromSnapshot(snapshot *domainFinance.HealthScoreSnapshot) MonthCloseMonth {
	return MonthCloseMonth{
		YearMonth: snapshot.YearMonth(),
		Score:     snapshot.Score(),
		Components: HealthScoreComponents{
			BudgetAdherence: snapshot.BudgetAdherence(),
			Cashflow:        snapshot.Cashflow(),
			Coverage:        snapshot.Coverage(),
		},
	}
}

func previousYearMonth(yearMonth string) (string, error) {
	parsed, err := time.Parse("2006-01", yearMonth)
	if err != nil {
		return "", err
	}
	return parsed.AddDate(0, -1, 0).Format("2006-01"), nil
}

func movedMost(delta MonthCloseDelta) string {
	ranked := []struct {
		name string
		abs  float64
	}{
		{"budget_adherence", math.Abs(delta.BudgetAdherence)},
		{"cashflow", math.Abs(delta.Cashflow)},
		{"coverage", math.Abs(delta.Coverage)},
	}
	best := ranked[0]
	for _, item := range ranked[1:] {
		if item.abs > best.abs {
			best = item
		}
	}
	return best.name
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}
