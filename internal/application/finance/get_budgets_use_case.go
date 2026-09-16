package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
)

// GetBudgetsResponse represents the response for getting budgets
type GetBudgetsResponse struct {
	Budgets []BudgetResponse `json:"budgets"`
}

// GetBudgetsUseCase handles getting budgets for a user
type GetBudgetsUseCase struct {
	budgetService      *finance.BudgetService
	categoryService    *finance.CategoryService
	transactionService *finance.TransactionService
}

// NewGetBudgetsUseCase creates a new get budgets use case
func NewGetBudgetsUseCase(
	budgetService *finance.BudgetService,
	categoryService *finance.CategoryService,
	transactionService *finance.TransactionService,
) *GetBudgetsUseCase {
	return &GetBudgetsUseCase{
		budgetService:      budgetService,
		categoryService:    categoryService,
		transactionService: transactionService,
	}
}

// Execute executes the get budgets use case
func (uc *GetBudgetsUseCase) Execute(ctx context.Context, userID int) (*GetBudgetsResponse, error) {
	budgets, err := uc.budgetService.GetBudgetsByUser(ctx, finance.NewUserID(userID))
	if err != nil {
		return nil, err
	}

	budgetResponses := make([]BudgetResponse, len(budgets))
	for i, budget := range budgets {
		budgetResponses[i] = toBudgetResponse(ctx, budget, uc.categoryService, uc.transactionService)
	}

	return &GetBudgetsResponse{
		Budgets: budgetResponses,
	}, nil
}
