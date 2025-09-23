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
) error {
	// Parse budget ID
	budgetIDInt, err := strconv.Atoi(budgetIDStr)
	if err != nil {
		return err
	}

	// Parse start date
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return err
	}

	// Convert to domain types
	budgetID := finance.NewBudgetID(budgetIDInt)
	userIDDomain := finance.NewUserID(userID)
	amountDomain, err := finance.NewMoney(amount, finance.NewCurrencyID(1)) // Default currency
	if err != nil {
		return err
	}
	period := finance.BudgetPeriod(periodStr)

	// Update budget
	return uc.budgetService.UpdateBudget(
		ctx,
		budgetID,
		userIDDomain,
		amountDomain,
		period,
		startDate,
	)
}
