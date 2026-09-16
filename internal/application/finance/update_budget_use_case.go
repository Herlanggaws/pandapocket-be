package finance

import (
	"context"
	"errors"
	"panda-pocket/internal/domain/finance"
	"strconv"
	"time"
)

// UpdateBudgetRequest represents the request for updating a budget
type UpdateBudgetRequest struct {
	CategoryID int     `json:"category_id" binding:"required"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Period     string  `json:"period" binding:"required,oneof=weekly monthly yearly"`
	StartDate  string  `json:"start_date" binding:"required"`
	EndDate    string  `json:"end_date" binding:"required"`
}

// UpdateBudgetUseCase handles budget updates
type UpdateBudgetUseCase struct {
	budgetService      *finance.BudgetService
	categoryService    *finance.CategoryService
	transactionService *finance.TransactionService
}

// NewUpdateBudgetUseCase creates a new update budget use case
func NewUpdateBudgetUseCase(
	budgetService *finance.BudgetService,
	categoryService *finance.CategoryService,
	transactionService *finance.TransactionService,
) *UpdateBudgetUseCase {
	return &UpdateBudgetUseCase{
		budgetService:      budgetService,
		categoryService:    categoryService,
		transactionService: transactionService,
	}
}

// Execute updates a budget
func (uc *UpdateBudgetUseCase) Execute(
	ctx context.Context,
	budgetIDStr string,
	userID int,
	req UpdateBudgetRequest,
) (*BudgetResponse, error) {
	budgetIDInt, err := strconv.Atoi(budgetIDStr)
	if err != nil {
		return nil, err
	}

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, err
	}

	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return nil, err
	}

	if endDate.Before(startDate) {
		return nil, errors.New("end date must be on or after start date")
	}

	budgetID := finance.NewBudgetID(budgetIDInt)
	userIDDomain := finance.NewUserID(userID)

	existingBudget, err := uc.budgetService.GetBudgetByID(ctx, budgetID)
	if err != nil {
		return nil, errors.New("budget not found")
	}
	if existingBudget.UserID().Value() != userID {
		return nil, errors.New("budget not found")
	}

	amountDomain, err := finance.NewMoney(req.Amount, existingBudget.Amount().Currency())
	if err != nil {
		return nil, err
	}

	updatedBudget, err := uc.budgetService.UpdateBudget(
		ctx,
		budgetID,
		userIDDomain,
		finance.NewCategoryID(req.CategoryID),
		amountDomain,
		finance.BudgetPeriod(req.Period),
		startDate,
		endDate,
	)
	if err != nil {
		return nil, err
	}

	response := toBudgetResponse(ctx, updatedBudget, uc.categoryService, uc.transactionService)
	return &response, nil
}
