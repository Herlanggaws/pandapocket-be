package finance

import (
	"context"
	"math"
	"panda-pocket/internal/domain/finance"
	"time"
)

type HealthScoreComponents struct {
	BudgetAdherence float64 `json:"budget_adherence"`
	Cashflow        float64 `json:"cashflow"`
	Coverage        float64 `json:"coverage"`
}

type HealthScoreResponse struct {
	Score      int                   `json:"score"`
	Components HealthScoreComponents `json:"components"`
	Period     string                `json:"period"`
	AsOf       string                `json:"as_of,omitempty"`
	YearMonth  string                `json:"year_month,omitempty"`
}

type HealthScoreHistoryItem struct {
	YearMonth       string  `json:"year_month"`
	Score           int     `json:"score"`
	BudgetAdherence float64 `json:"budget_adherence"`
	Cashflow        float64 `json:"cashflow"`
	Coverage        float64 `json:"coverage"`
	ComputedAt      string  `json:"computed_at"`
}

type GetHealthScoreUseCase struct {
	budgetService      *finance.BudgetService
	categoryService    *finance.CategoryService
	transactionService *finance.TransactionService
	analyticsUseCase   *GetAnalyticsUseCase
	snapshotRepo       finance.HealthScoreSnapshotRepository
}

func NewGetHealthScoreUseCase(
	budgetService *finance.BudgetService,
	categoryService *finance.CategoryService,
	transactionService *finance.TransactionService,
	analyticsUseCase *GetAnalyticsUseCase,
	snapshotRepo finance.HealthScoreSnapshotRepository,
) *GetHealthScoreUseCase {
	return &GetHealthScoreUseCase{
		budgetService:      budgetService,
		categoryService:    categoryService,
		transactionService: transactionService,
		analyticsUseCase:   analyticsUseCase,
		snapshotRepo:       snapshotRepo,
	}
}

func (uc *GetHealthScoreUseCase) Execute(ctx context.Context, userID int) (*HealthScoreResponse, error) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	yearMonth := now.Format("2006-01")

	budgets, err := uc.budgetService.GetBudgetsByUser(ctx, finance.NewUserID(userID))
	if err != nil {
		return nil, err
	}

	var overlapping []*finance.Budget
	for _, budget := range budgets {
		if budget.OverlapsWith(monthStart, monthEnd) {
			overlapping = append(overlapping, budget)
		}
	}

	adherence := 50.0
	if len(overlapping) > 0 {
		var sum float64
		var counted int
		for _, budget := range overlapping {
			report, reportErr := calculateBudgetReport(ctx, uc.transactionService, budget)
			if reportErr != nil || report == nil {
				continue
			}
			component := math.Max(0, 100-math.Min(report.PercentageUsed, 150))
			if component > 100 {
				component = 100
			}
			sum += component
			counted++
		}
		if counted > 0 {
			adherence = sum / float64(counted)
		}
	}

	cashflow := 50.0
	if uc.analyticsUseCase != nil {
		analytics, analyticsErr := uc.analyticsUseCase.Execute(ctx, userID, GetAnalyticsRequest{Period: "monthly"})
		if analyticsErr == nil && analytics != nil {
			if analytics.TotalIncome > 0 {
				cashflow = clampFloat(100*(1-analytics.TotalSpent/analytics.TotalIncome), 0, 100)
			}
		}
	}

	coverage := 40.0
	if len(overlapping) >= 1 {
		coverage = 100
	}

	adherence = math.Round(adherence*100) / 100
	cashflow = math.Round(cashflow*100) / 100
	score := int(math.Round(0.5*adherence + 0.3*cashflow + 0.2*coverage))

	asOf := now.Format(time.RFC3339)
	if uc.snapshotRepo != nil {
		snapshot := finance.NewHealthScoreSnapshot(
			finance.NewUserID(userID),
			yearMonth,
			score,
			adherence,
			cashflow,
			coverage,
		)
		_ = uc.snapshotRepo.Upsert(ctx, snapshot)
		asOf = snapshot.ComputedAt().Format(time.RFC3339)
	}

	return &HealthScoreResponse{
		Score: score,
		Components: HealthScoreComponents{
			BudgetAdherence: adherence,
			Cashflow:        cashflow,
			Coverage:        coverage,
		},
		Period:    "monthly",
		AsOf:      asOf,
		YearMonth: yearMonth,
	}, nil
}

type GetHealthScoreHistoryUseCase struct {
	snapshotRepo finance.HealthScoreSnapshotRepository
}

func NewGetHealthScoreHistoryUseCase(snapshotRepo finance.HealthScoreSnapshotRepository) *GetHealthScoreHistoryUseCase {
	return &GetHealthScoreHistoryUseCase{snapshotRepo: snapshotRepo}
}

func (uc *GetHealthScoreHistoryUseCase) Execute(ctx context.Context, userID int, limit int) ([]HealthScoreHistoryItem, error) {
	if limit <= 0 || limit > 24 {
		limit = 12
	}
	snapshots, err := uc.snapshotRepo.FindByUserID(ctx, finance.NewUserID(userID), limit)
	if err != nil {
		return nil, err
	}
	result := make([]HealthScoreHistoryItem, 0, len(snapshots))
	for _, s := range snapshots {
		result = append(result, HealthScoreHistoryItem{
			YearMonth:       s.YearMonth(),
			Score:           s.Score(),
			BudgetAdherence: s.BudgetAdherence(),
			Cashflow:        s.Cashflow(),
			Coverage:        s.Coverage(),
			ComputedAt:      s.ComputedAt().Format(time.RFC3339),
		})
	}
	return result, nil
}

func clampFloat(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
