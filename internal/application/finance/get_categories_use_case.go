package finance

import (
	"context"
	"errors"
	"panda-pocket/internal/domain/finance"
)

// GetCategoriesResponse represents the response for getting categories
type GetCategoriesResponse struct {
	Categories []CategoryResponse `json:"categories"`
}

// CategoryResponse represents a category in the response
type CategoryResponse struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	Type  string `json:"type"`
}

// GetCategoriesUseCase handles getting categories for a user
type GetCategoriesUseCase struct {
	categoryService *finance.CategoryService
}

// NewGetCategoriesUseCase creates a new get categories use case
func NewGetCategoriesUseCase(categoryService *finance.CategoryService) *GetCategoriesUseCase {
	return &GetCategoriesUseCase{
		categoryService: categoryService,
	}
}

// Execute executes the get categories use case
func (uc *GetCategoriesUseCase) Execute(ctx context.Context, userID int, categoryType string) (*GetCategoriesResponse, error) {
	var categories []*finance.Category
	var err error
	
	if categoryType != "" {
		// Get categories by type
		var financeCategoryType finance.CategoryType
		switch categoryType {
		case "expense":
			financeCategoryType = finance.CategoryTypeExpense
		case "income":
			financeCategoryType = finance.CategoryTypeIncome
		default:
			return nil, errors.New("invalid category type")
		}
		
		categories, err = uc.categoryService.GetCategoriesByUserAndType(ctx, finance.NewUserID(userID), financeCategoryType)
	} else {
		// Get all categories
		categories, err = uc.categoryService.GetCategoriesByUser(ctx, finance.NewUserID(userID))
	}
	
	if err != nil {
		return nil, err
	}
	
	// Convert to response format
	categoryResponses := make([]CategoryResponse, len(categories))
	for i, category := range categories {
		categoryResponses[i] = CategoryResponse{
			ID:    category.ID().Value(),
			Name:  category.Name(),
			Color: category.Color(),
			Type:  string(category.Type()),
		}
	}
	
	return &GetCategoriesResponse{
		Categories: categoryResponses,
	}, nil
}
