package finance

import (
	"context"
	"errors"
	"panda-pocket/internal/domain/finance"
)

// CreateCategoryRequest represents the request to create a category
type CreateCategoryRequest struct {
	Name  string `json:"name" binding:"required"`
	Color string `json:"color"`
	Type  string `json:"type" binding:"required,oneof=expense income"`
}

// CreateCategoryResponse represents the response after creating a category
type CreateCategoryResponse struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	Type      string `json:"type"`
	IsDefault bool   `json:"is_default"`
}

// CreateCategoryUseCase handles category creation
type CreateCategoryUseCase struct {
	categoryService *finance.CategoryService
}

// NewCreateCategoryUseCase creates a new create category use case
func NewCreateCategoryUseCase(categoryService *finance.CategoryService) *CreateCategoryUseCase {
	return &CreateCategoryUseCase{
		categoryService: categoryService,
	}
}

// Execute executes the create category use case
func (uc *CreateCategoryUseCase) Execute(ctx context.Context, userID int, req CreateCategoryRequest) (*CreateCategoryResponse, error) {
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

	// Create category
	category, err := uc.categoryService.CreateCategory(
		ctx,
		finance.NewUserID(userID),
		req.Name,
		req.Color,
		categoryType,
	)
	if err != nil {
		return nil, err
	}

	return &CreateCategoryResponse{
		ID:        category.ID().Value(),
		Name:      category.Name(),
		Color:     category.Color(),
		Type:      string(category.Type()),
		IsDefault: category.IsDefault(),
	}, nil
}
