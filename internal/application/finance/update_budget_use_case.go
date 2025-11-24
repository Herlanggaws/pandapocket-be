package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
	"strconv"
	"time"
)

// UpdateBudgetResponse represents the response for updating a budget
type UpdateBudgetResponse struct {
	Amount    float64          `json:"amount"`
	Period    string           `json:"period"`
	StartDate string           `json:"start_date"`
	EndDate   string           `json:"end_date"`
	Category  *CategoryResponse `json:"category"`
}

// UpdateBudgetUseCase handles budget updates
type UpdateBudgetUseCase struct {
	budgetService   *finance.BudgetService
	categoryService *finance.CategoryService
}

// NewUpdateBudgetUseCase creates a new update budget use case
func NewUpdateBudgetUseCase(budgetService *finance.BudgetService, categoryService *finance.CategoryService) *UpdateBudgetUseCase {
	return &UpdateBudgetUseCase{
		budgetService:   budgetService,
		categoryService: categoryService,
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
) (*UpdateBudgetResponse, error) {
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

	// Fetch category information
	category, err := uc.categoryService.GetCategoryByID(ctx, updatedBudget.CategoryID())
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
	return &UpdateBudgetResponse{
		Amount:    updatedBudget.Amount().Amount(),
		Period:    string(updatedBudget.Period()),
		StartDate: updatedBudget.StartDate().Format("2006-01-02"),
		EndDate:   updatedBudget.EndDate().Format("2006-01-02"),
		Category:  categoryResponse,
	}, nil
}
