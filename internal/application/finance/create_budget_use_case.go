package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
	"time"
)

// CreateBudgetRequest represents the request for creating a budget
type CreateBudgetRequest struct {
	CategoryID int     `json:"category_id" binding:"required"`
	Amount     float64 `json:"amount" binding:"required,gt=0"`
	Period     string  `json:"period" binding:"required,oneof=weekly monthly yearly"`
	StartDate  string  `json:"start_date" binding:"required"`
}

// CreateBudgetUseCase handles budget creation
type CreateBudgetUseCase struct {
	budgetService      *finance.BudgetService
	currencyService    *finance.CurrencyService
	categoryService    *finance.CategoryService
	transactionService *finance.TransactionService
}

// NewCreateBudgetUseCase creates a new create budget use case
func NewCreateBudgetUseCase(
	budgetService *finance.BudgetService,
	currencyService *finance.CurrencyService,
	categoryService *finance.CategoryService,
	transactionService *finance.TransactionService,
) *CreateBudgetUseCase {
	return &CreateBudgetUseCase{
		budgetService:      budgetService,
		currencyService:    currencyService,
		categoryService:    categoryService,
		transactionService: transactionService,
	}
}

// Execute executes the create budget use case
func (uc *CreateBudgetUseCase) Execute(ctx context.Context, userID int, req CreateBudgetRequest) (*BudgetResponse, error) {
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, err
	}

	currency, err := uc.currencyService.GetPrimaryCurrency(ctx, finance.NewUserID(userID))
	if err != nil {
		return nil, err
	}

	money, err := finance.NewMoney(req.Amount, currency.ID())
	if err != nil {
		return nil, err
	}

	budget, err := uc.budgetService.CreateBudget(
		ctx,
		finance.NewUserID(userID),
		finance.NewCategoryID(req.CategoryID),
		money,
		finance.BudgetPeriod(req.Period),
		startDate,
	)
	if err != nil {
		return nil, err
	}

	response := toBudgetResponse(ctx, budget, uc.categoryService, uc.transactionService)
	return &response, nil
}
