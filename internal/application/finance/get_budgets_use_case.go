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
}

// GetBudgetsUseCase handles getting budgets for a user
type GetBudgetsUseCase struct {
	budgetService   *finance.BudgetService
	categoryService *finance.CategoryService
}

// NewGetBudgetsUseCase creates a new get budgets use case
func NewGetBudgetsUseCase(budgetService *finance.BudgetService, categoryService *finance.CategoryService) *GetBudgetsUseCase {
	return &GetBudgetsUseCase{
		budgetService:   budgetService,
		categoryService: categoryService,
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

		budgetResponses[i] = BudgetResponse{
			ID:        budget.ID().Value(),
			UserID:    budget.UserID().Value(),
			Amount:    budget.Amount().Amount(),
			Period:    string(budget.Period()),
			StartDate: budget.StartDate().Format("2006-01-02"),
			EndDate:   budget.EndDate().Format("2006-01-02"),
			CreatedAt: budget.CreatedAt().Format(time.RFC3339),
			Category:  categoryResponse,
		}
	}

	return &GetBudgetsResponse{
		Budgets: budgetResponses,
	}, nil
}
