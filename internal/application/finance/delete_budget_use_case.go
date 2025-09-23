package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
	"strconv"
)

// DeleteBudgetUseCase handles budget deletion
type DeleteBudgetUseCase struct {
	budgetService *finance.BudgetService
}

// NewDeleteBudgetUseCase creates a new delete budget use case
func NewDeleteBudgetUseCase(budgetService *finance.BudgetService) *DeleteBudgetUseCase {
	return &DeleteBudgetUseCase{
		budgetService: budgetService,
	}
}

// Execute deletes a budget
func (uc *DeleteBudgetUseCase) Execute(ctx context.Context, budgetIDStr string, userID int) error {
	// Parse budget ID
	budgetIDInt, err := strconv.Atoi(budgetIDStr)
	if err != nil {
		return err
	}

	// Convert to domain types
	budgetID := finance.NewBudgetID(budgetIDInt)
	userIDDomain := finance.NewUserID(userID)

	// Delete budget
	return uc.budgetService.DeleteBudget(ctx, budgetID, userIDDomain)
}
