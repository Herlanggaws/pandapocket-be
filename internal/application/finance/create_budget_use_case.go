package finance

import (
	"context"
	"errors"
	"panda-pocket/internal/domain/entitlement"
	"panda-pocket/internal/domain/finance"
	"time"
)

// CreateBudgetRequest represents the request for creating a budget
type CreateBudgetRequest struct {
	CategoryID int      `json:"category_id" binding:"required"`
	Amount     float64  `json:"amount"`
	LimitType  string   `json:"limit_type"`
	Percent    *float64 `json:"percent"`
	Period     string   `json:"period" binding:"required,oneof=weekly monthly yearly"`
	StartDate  string   `json:"start_date" binding:"required"`
}

// CreateBudgetUseCase handles budget creation
type CreateBudgetUseCase struct {
	budgetService      *finance.BudgetService
	currencyService    *finance.CurrencyService
	categoryService    *finance.CategoryService
	transactionService *finance.TransactionService
	entitlements       entitlement.Checker
}

// NewCreateBudgetUseCase creates a new create budget use case
func NewCreateBudgetUseCase(
	budgetService *finance.BudgetService,
	currencyService *finance.CurrencyService,
	categoryService *finance.CategoryService,
	transactionService *finance.TransactionService,
	entitlements entitlement.Checker,
) *CreateBudgetUseCase {
	return &CreateBudgetUseCase{
		budgetService:      budgetService,
		currencyService:    currencyService,
		categoryService:    categoryService,
		transactionService: transactionService,
		entitlements:       entitlements,
	}
}

// Execute executes the create budget use case
func (uc *CreateBudgetUseCase) Execute(ctx context.Context, userID int, req CreateBudgetRequest) (*BudgetResponse, error) {
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, err
	}

	active, err := uc.budgetService.GetActiveBudgetsByUser(ctx, finance.NewUserID(userID))
	if err != nil {
		return nil, err
	}
	if err := entitlement.EnforceCreateLimit(
		ctx, uc.entitlements, userID,
		entitlement.FeatureBudgets, len(active), entitlement.FreeActiveBudgets,
	); err != nil {
		return nil, err
	}

	limitType, err := finance.ParseBudgetLimitType(req.LimitType)
	if err != nil {
		return nil, err
	}

	currency, err := uc.currencyService.GetPrimaryCurrency(ctx, finance.NewUserID(userID))
	if err != nil {
		return nil, err
	}

	amountValue := req.Amount
	if limitType == finance.BudgetLimitPercent {
		amountValue = 0
	}

	money, err := finance.NewMoney(amountValue, currency.ID())
	if err != nil {
		return nil, err
	}

	if limitType == finance.BudgetLimitFixed && req.Amount <= 0 {
		return nil, errors.New("budget amount must be positive")
	}

	budget, err := uc.budgetService.CreateBudget(
		ctx,
		finance.NewUserID(userID),
		finance.NewCategoryID(req.CategoryID),
		money,
		limitType,
		req.Percent,
		finance.BudgetPeriod(req.Period),
		startDate,
	)
	if err != nil {
		return nil, err
	}

	response := toBudgetResponse(ctx, budget, uc.categoryService, uc.transactionService)
	return &response, nil
}
