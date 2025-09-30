package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
	"time"
)

// GetBudgetsResponse represents the response for getting budgets
type GetBudgetsResponse struct {
	Budgets []BudgetResponse `json:"budgets"`
}

// BudgetResponse represents a budget in the response
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

// BudgetReport represents budget tracking information
type BudgetReport struct {
	IsOnTrack      bool    `json:"is_on_track"`
	TotalSpent     float64 `json:"total_spent"`
	Remaining      float64 `json:"remaining"`
	PercentageUsed float64 `json:"percentage_used"`
}

// GetBudgetsUseCase handles getting budgets for a user
type GetBudgetsUseCase struct {
	budgetService      *finance.BudgetService
	categoryService    *finance.CategoryService
	transactionService *finance.TransactionService
}

// NewGetBudgetsUseCase creates a new get budgets use case
func NewGetBudgetsUseCase(budgetService *finance.BudgetService, categoryService *finance.CategoryService, transactionService *finance.TransactionService) *GetBudgetsUseCase {
	return &GetBudgetsUseCase{
		budgetService:      budgetService,
		categoryService:    categoryService,
		transactionService: transactionService,
	}
}

// Execute executes the get budgets use case
func (uc *GetBudgetsUseCase) Execute(ctx context.Context, userID int) (*GetBudgetsResponse, error) {
	// Get budgets
	budgets, err := uc.budgetService.GetBudgetsByUser(ctx, finance.NewUserID(userID))
	if err != nil {
		return nil, err
	}

	// Convert to response format
	budgetResponses := make([]BudgetResponse, len(budgets))
	for i, budget := range budgets {
		// Fetch category information
		category, err := uc.categoryService.GetCategoryByID(ctx, budget.CategoryID())
		var categoryResponse *CategoryResponse
		if err == nil {
			categoryResponse = &CategoryResponse{
				ID:    category.ID().Value(),
				Name:  category.Name(),
				Color: category.Color(),
				Type:  string(category.Type()),
			}
		}

		// Calculate budget report
		report, err := uc.calculateBudgetReport(ctx, budget)
		if err != nil {
			// If report calculation fails, continue without report
			report = nil
		}

		budgetResponses[i] = BudgetResponse{
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

	return &GetBudgetsResponse{
		Budgets: budgetResponses,
	}, nil
}

// calculateBudgetReport calculates the budget report for a given budget
func (uc *GetBudgetsUseCase) calculateBudgetReport(ctx context.Context, budget *finance.Budget) (*BudgetReport, error) {
	// Get transactions for this budget's category within the budget period
	transactions, err := uc.transactionService.GetTransactionsByUserAndDateRange(
		ctx,
		budget.UserID(),
		budget.StartDate(),
		budget.EndDate(),
	)
	if err != nil {
		return nil, err
	}

	// Filter transactions by category and type (only expenses)
	var totalSpent float64
	for _, transaction := range transactions {
		if transaction.CategoryID().Value() == budget.CategoryID().Value() &&
			transaction.Type() == finance.TransactionTypeExpense {
			totalSpent += transaction.Amount().Amount()
		}
	}

	// Calculate report metrics
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
