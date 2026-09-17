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
}

type GetHealthScoreUseCase struct {
	budgetService      *finance.BudgetService
	categoryService    *finance.CategoryService
	transactionService *finance.TransactionService
	analyticsUseCase   *GetAnalyticsUseCase
}

func NewGetHealthScoreUseCase(
	budgetService *finance.BudgetService,
	categoryService *finance.CategoryService,
	transactionService *finance.TransactionService,
	analyticsUseCase *GetAnalyticsUseCase,
) *GetHealthScoreUseCase {
	return &GetHealthScoreUseCase{
		budgetService:      budgetService,
		categoryService:    categoryService,
		transactionService: transactionService,
		analyticsUseCase:   analyticsUseCase,
	}
}

func (uc *GetHealthScoreUseCase) Execute(ctx context.Context, userID int) (*HealthScoreResponse, error) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

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

	score := int(math.Round(0.5*adherence + 0.3*cashflow + 0.2*coverage))

	return &HealthScoreResponse{
		Score: score,
		Components: HealthScoreComponents{
			BudgetAdherence: math.Round(adherence*100) / 100,
			Cashflow:        math.Round(cashflow*100) / 100,
			Coverage:        coverage,
		},
		Period: "monthly",
	}, nil
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
