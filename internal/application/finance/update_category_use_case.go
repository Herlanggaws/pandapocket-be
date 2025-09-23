package finance

import (
	"context"
	"errors"
	"panda-pocket/internal/domain/finance"
)

// UpdateCategoryRequest represents the request to update a category
type UpdateCategoryRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
	Type  string `json:"type" binding:"required,oneof=expense income"`
}

// UpdateCategoryResponse represents the response after updating a category
type UpdateCategoryResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	Type  string `json:"type"`
}

// UpdateCategoryUseCase handles category updates
type UpdateCategoryUseCase struct {
	categoryService *finance.CategoryService
}

// NewUpdateCategoryUseCase creates a new update category use case
func NewUpdateCategoryUseCase(categoryService *finance.CategoryService) *UpdateCategoryUseCase {
	return &UpdateCategoryUseCase{
		categoryService: categoryService,
	}
}

// Execute executes the update category use case
func (uc *UpdateCategoryUseCase) Execute(ctx context.Context, userID int, categoryID int, req UpdateCategoryRequest) (*UpdateCategoryResponse, error) {
	// Validate category type
	var categoryType finance.CategoryType
	switch req.Type {
	case "expense":
		categoryType = finance.CategoryTypeExpense
	case "income":
		categoryType = finance.CategoryTypeIncome
	default:
		return nil, errors.New("invalid category type")
	}

	// Update category
	err := uc.categoryService.UpdateCategory(
		ctx,
		finance.NewCategoryID(categoryID),
		finance.NewUserID(userID),
		req.Name,
		req.Color,
		categoryType,
	)
	if err != nil {
		return nil, err
	}

	// Get updated category to return
	category, err := uc.categoryService.GetCategoryByID(ctx, finance.NewCategoryID(categoryID))
	if err != nil {
		return nil, err
	}

	return &UpdateCategoryResponse{
		ID:    category.ID().Value(),
		Name:  category.Name(),
		Color: category.Color(),
		Type:  string(category.Type()),
	}, nil
}

