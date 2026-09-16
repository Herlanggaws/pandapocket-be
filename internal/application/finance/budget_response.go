package finance

import (
	"context"
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
	ID        int               `json:"id"`
	UserID    int               `json:"user_id"`
	Amount    float64           `json:"amount"`
	Period    string            `json:"period"`
	StartDate string            `json:"start_date"`
	EndDate   string            `json:"end_date"`
	CreatedAt string            `json:"created_at"`
	Category  *CategoryResponse `json:"category,omitempty"`
	Report    *BudgetReport     `json:"report,omitempty"`
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
	budgetAmount := budget.Amount().Amount()
	remaining := budgetAmount - totalSpent
	percentageUsed := (totalSpent / budgetAmount) * 100
	isOnTrack := totalSpent <= budgetAmount

	return &BudgetReport{
		IsOnTrack:      isOnTrack,
		TotalSpent:     totalSpent,
		Remaining:      remaining,
		PercentageUsed: percentageUsed,
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

	var report *BudgetReport
	if transactionService != nil {
		report, err = calculateBudgetReport(ctx, transactionService, budget)
		if err != nil {
			report = nil
		}
	}

	return BudgetResponse{
		ID:        budget.ID().Value(),
		UserID:    budget.UserID().Value(),
		Amount:    budget.Amount().Amount(),
		Period:    string(budget.Period()),
		StartDate: budget.StartDate().Format("2006-01-02"),
		EndDate:   budget.EndDate().Format("2006-01-02"),
		CreatedAt: budget.CreatedAt().Format(time.RFC3339),
		Category:  categoryResponse,
		Report:    report,
	}
}
