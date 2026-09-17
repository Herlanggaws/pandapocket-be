package finance

import (
	"context"
	"math"
	"panda-pocket/internal/domain/finance"
	"time"
)

// BudgetReport represents budget tracking information
type BudgetReport struct {
	IsOnTrack      bool    `json:"is_on_track"`
	TotalSpent     float64 `json:"total_spent"`
	Remaining      float64 `json:"remaining"`
	PercentageUsed float64 `json:"percentage_used"`
}

// BudgetResponse represents a budget in API responses
type BudgetResponse struct {
	ID              int               `json:"id"`
	UserID          int               `json:"user_id"`
	Amount          float64           `json:"amount"`
	LimitType       string            `json:"limit_type"`
	Percent         *float64          `json:"percent,omitempty"`
	EffectiveAmount float64           `json:"effective_amount"`
	Period          string            `json:"period"`
	StartDate       string            `json:"start_date"`
	EndDate         string            `json:"end_date"`
	CreatedAt       string            `json:"created_at"`
	Category        *CategoryResponse `json:"category,omitempty"`
	Report          *BudgetReport     `json:"report,omitempty"`
}

func buildCategoryResponse(category *finance.Category) *CategoryResponse {
	if category == nil {
		return nil
	}
	return &CategoryResponse{
		ID:    category.ID().Value(),
		Name:  category.Name(),
		Color: category.Color(),
		Type:  string(category.Type()),
	}
}

func sumExpensesForBudget(transactions []*finance.Transaction, budget *finance.Budget) float64 {
	budgetCurrencyID := budget.Amount().Currency().Value()
	var totalSpent float64
	for _, transaction := range transactions {
		if transaction.CategoryID().Value() == budget.CategoryID().Value() &&
			transaction.Type() == finance.TransactionTypeExpense &&
			transaction.CurrencyID().Value() == budgetCurrencyID {
			totalSpent += transaction.Amount().Amount()
		}
	}
	return totalSpent
}

func sumIncomeForBudget(transactions []*finance.Transaction, budget *finance.Budget) float64 {
	budgetCurrencyID := budget.Amount().Currency().Value()
	var totalIncome float64
	for _, transaction := range transactions {
		if transaction.Type() == finance.TransactionTypeIncome &&
			transaction.CurrencyID().Value() == budgetCurrencyID {
			totalIncome += transaction.Amount().Amount()
		}
	}
	return totalIncome
}

func effectiveBudgetAmount(budget *finance.Budget, periodIncome float64) float64 {
	if budget.LimitType() == finance.BudgetLimitPercent {
		if budget.Percent() == nil || periodIncome <= 0 {
			return 0
		}
		return periodIncome * (*budget.Percent()) / 100
	}
	return budget.Amount().Amount()
}

func calculateBudgetReport(
	ctx context.Context,
	transactionService *finance.TransactionService,
	budget *finance.Budget,
) (*BudgetReport, error) {
	transactions, err := transactionService.GetTransactionsByUserAndDateRange(
		ctx,
		budget.UserID(),
		budget.StartDate(),
		budget.EndDate(),
	)
	if err != nil {
		return nil, err
	}

	totalSpent := sumExpensesForBudget(transactions, budget)
	periodIncome := sumIncomeForBudget(transactions, budget)
	budgetAmount := effectiveBudgetAmount(budget, periodIncome)

	var remaining float64
	var percentageUsed float64
	isOnTrack := true

	if budgetAmount > 0 {
		remaining = budgetAmount - totalSpent
		percentageUsed = (totalSpent / budgetAmount) * 100
		isOnTrack = totalSpent <= budgetAmount
	} else {
		remaining = -totalSpent
		if totalSpent > 0 {
			percentageUsed = 100
			isOnTrack = false
		} else {
			percentageUsed = 0
			isOnTrack = true
		}
	}

	return &BudgetReport{
		IsOnTrack:      isOnTrack,
		TotalSpent:     totalSpent,
		Remaining:      remaining,
		PercentageUsed: math.Round(percentageUsed*100) / 100,
	}, nil
}

func toBudgetResponse(
	ctx context.Context,
	budget *finance.Budget,
	categoryService *finance.CategoryService,
	transactionService *finance.TransactionService,
) BudgetResponse {
	var categoryResponse *CategoryResponse
	category, err := categoryService.GetCategoryByID(ctx, budget.CategoryID())
	if err == nil {
		categoryResponse = buildCategoryResponse(category)
	}

	effectiveAmount := budget.Amount().Amount()
	var report *BudgetReport
	if transactionService != nil {
		transactions, txErr := transactionService.GetTransactionsByUserAndDateRange(
			ctx,
			budget.UserID(),
			budget.StartDate(),
			budget.EndDate(),
		)
		if txErr == nil {
			periodIncome := sumIncomeForBudget(transactions, budget)
			effectiveAmount = effectiveBudgetAmount(budget, periodIncome)
		}

		report, err = calculateBudgetReport(ctx, transactionService, budget)
		if err != nil {
			report = nil
		}
	} else if budget.LimitType() == finance.BudgetLimitPercent {
		effectiveAmount = 0
	}

	return BudgetResponse{
		ID:              budget.ID().Value(),
		UserID:          budget.UserID().Value(),
		Amount:          budget.Amount().Amount(),
		LimitType:       string(budget.LimitType()),
		Percent:         budget.Percent(),
		EffectiveAmount: effectiveAmount,
		Period:          string(budget.Period()),
		StartDate:       budget.StartDate().Format("2006-01-02"),
		EndDate:         budget.EndDate().Format("2006-01-02"),
		CreatedAt:       budget.CreatedAt().Format(time.RFC3339),
		Category:        categoryResponse,
		Report:          report,
	}
}
