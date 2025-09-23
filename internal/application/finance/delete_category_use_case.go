package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
)

// DeleteCategoryUseCase handles category deletion
type DeleteCategoryUseCase struct {
	categoryService *finance.CategoryService
}

// NewDeleteCategoryUseCase creates a new delete category use case
func NewDeleteCategoryUseCase(categoryService *finance.CategoryService) *DeleteCategoryUseCase {
	return &DeleteCategoryUseCase{
		categoryService: categoryService,
	}
}

// Execute executes the delete category use case
func (uc *DeleteCategoryUseCase) Execute(ctx context.Context, userID int, categoryID int) error {
	return uc.categoryService.DeleteCategory(
		ctx,
		finance.NewCategoryID(categoryID),
		finance.NewUserID(userID),
	)
}
