package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
	"strconv"
	"time"
)

// UpdateBudgetUseCase handles budget updates
type UpdateBudgetUseCase struct {
	budgetService *finance.BudgetService
}

// NewUpdateBudgetUseCase creates a new update budget use case
func NewUpdateBudgetUseCase(budgetService *finance.BudgetService) *UpdateBudgetUseCase {
	return &UpdateBudgetUseCase{
		budgetService: budgetService,
	}
}

// Execute updates a budget
func (uc *UpdateBudgetUseCase) Execute(
	ctx context.Context,
	budgetIDStr string,
	userID int,
	categoryIDStr string,
	amount float64,
	periodStr string,
	startDateStr string,
	endDateStr string,
) (*finance.Budget, error) {
	// Parse budget ID
	budgetIDInt, err := strconv.Atoi(budgetIDStr)
	if err != nil {
		return nil, err
	}

	// Parse category ID
	categoryIDInt, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		return nil, err
	}

	// Parse start date
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return nil, err
	}

	// Parse end date
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return nil, err
	}

	// Convert to domain types
	budgetID := finance.NewBudgetID(budgetIDInt)
	userIDDomain := finance.NewUserID(userID)

	// Load existing budget to preserve its currency when updating amount
	existingBudget, err := uc.budgetService.GetBudgetByID(ctx, budgetID)
	if err != nil {
		return nil, err
	}

	amountDomain, err := finance.NewMoney(amount, existingBudget.Amount().Currency())
	if err != nil {
		return nil, err
	}
	period := finance.BudgetPeriod(periodStr)

	// Update budget
	updatedBudget, err := uc.budgetService.UpdateBudget(
		ctx,
		budgetID,
		userIDDomain,
		finance.NewCategoryID(categoryIDInt),
		amountDomain,
		period,
		startDate,
		endDate,
	)
	if err != nil {
		return nil, err
	}

	return updatedBudget, nil
}
