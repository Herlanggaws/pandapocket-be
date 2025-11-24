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
	Amount    float64           `json:"amount"`
	Period    string            `json:"period"`
	StartDate string            `json:"start_date"`
	EndDate   string            `json:"end_date"`
	Category  *CategoryResponse `json:"category"`
}

// CreateBudgetUseCase handles budget creation
type CreateBudgetUseCase struct {
	budgetService   *finance.BudgetService
	currencyService *finance.CurrencyService
	categoryService *finance.CategoryService
}

// NewCreateBudgetUseCase creates a new create budget use case
func NewCreateBudgetUseCase(budgetService *finance.BudgetService, currencyService *finance.CurrencyService, categoryService *finance.CategoryService) *CreateBudgetUseCase {
	return &CreateBudgetUseCase{
		budgetService:   budgetService,
		currencyService: currencyService,
		categoryService: categoryService,
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

	// Convert to response format
	return &CreateBudgetResponse{
		Amount:    budget.Amount().Amount(),
		Period:    string(budget.Period()),
		StartDate: budget.StartDate().Format("2006-01-02"),
		EndDate:   budget.EndDate().Format("2006-01-02"),
		Category:  categoryResponse,
	}, nil
}
