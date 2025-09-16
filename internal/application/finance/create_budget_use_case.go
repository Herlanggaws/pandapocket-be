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

// CreateBudgetResponse represents the response for creating a budget
type CreateBudgetResponse struct {
	ID         int     `json:"id"`
	UserID     int     `json:"user_id"`
	CategoryID int     `json:"category_id"`
	Amount     float64 `json:"amount"`
	Period     string  `json:"period"`
	StartDate  string  `json:"start_date"`
	EndDate    string  `json:"end_date"`
	CreatedAt  string  `json:"created_at"`
}

// CreateBudgetUseCase handles budget creation
type CreateBudgetUseCase struct {
	budgetService   *finance.BudgetService
	currencyService *finance.CurrencyService
}

// NewCreateBudgetUseCase creates a new create budget use case
func NewCreateBudgetUseCase(budgetService *finance.BudgetService, currencyService *finance.CurrencyService) *CreateBudgetUseCase {
	return &CreateBudgetUseCase{
		budgetService:   budgetService,
		currencyService: currencyService,
	}
}

// Execute executes the create budget use case
func (uc *CreateBudgetUseCase) Execute(ctx context.Context, userID int, req CreateBudgetRequest) (*CreateBudgetResponse, error) {
	// Parse start date
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return nil, err
	}

	// Get primary currency for the user
	currency, err := uc.currencyService.GetPrimaryCurrency(ctx, finance.NewUserID(userID))
	if err != nil {
		return nil, err
	}

	// Create money object
	money, err := finance.NewMoney(req.Amount, currency.ID())
	if err != nil {
		return nil, err
	}

	// Create budget
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

	// Convert to response format
	return &CreateBudgetResponse{
		ID:         budget.ID().Value(),
		UserID:     budget.UserID().Value(),
		CategoryID: budget.CategoryID().Value(),
		Amount:     budget.Amount().Amount(),
		Period:     string(budget.Period()),
		StartDate:  budget.StartDate().Format("2006-01-02"),
		EndDate:    budget.EndDate().Format("2006-01-02"),
		CreatedAt:  budget.CreatedAt().Format(time.RFC3339),
	}, nil
}
