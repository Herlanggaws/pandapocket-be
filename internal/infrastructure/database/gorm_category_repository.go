package database

import (
	"context"
	"panda-pocket/internal/domain/finance"

	"gorm.io/gorm"
)

// GormCategoryRepository implements the CategoryRepository interface using GORM
type GormCategoryRepository struct {
	db *gorm.DB
}

// NewGormCategoryRepository creates a new GORM category repository
func NewGormCategoryRepository(db *gorm.DB) *GormCategoryRepository {
	return &GormCategoryRepository{db: db}
}

// Save saves a category to the database
func (r *GormCategoryRepository) Save(ctx context.Context, category *finance.Category) error {
	// Convert domain category to GORM model
	categoryModel := &Category{
		Name:         category.Name(),
		Color:        category.Color(),
		IsDefault:    category.IsDefault(),
		CategoryType: string(category.Type()),
	}

	if category.ID().Value() != 0 {
		categoryModel.ID = uint(category.ID().Value())
	}

	if category.UserID().Value() != 0 {
		userID := uint(category.UserID().Value())
		categoryModel.UserID = &userID
	}

	// Save using GORM
	if err := r.db.WithContext(ctx).Save(categoryModel).Error; err != nil {
		return err
	}

	return nil
}

// FindByID finds a category by ID
func (r *GormCategoryRepository) FindByID(ctx context.Context, id finance.CategoryID) (*finance.Category, error) {
	var categoryModel Category

	err := r.db.WithContext(ctx).First(&categoryModel, id.Value()).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}

	// Convert GORM model to domain category
	categoryID := finance.NewCategoryID(int(categoryModel.ID))
	var userID *finance.UserID
	if categoryModel.UserID != nil {
		userIDVal := finance.NewUserID(int(*categoryModel.UserID))
		userID = &userIDVal
	}
	categoryType := finance.CategoryType(categoryModel.CategoryType)

	category, err := finance.NewCategory(
		categoryID,
		userID,
		categoryModel.Name,
		categoryModel.Color,
		categoryModel.IsDefault,
		categoryType,
	)
	if err != nil {
		return nil, err
	}

	return category, nil
}

// FindByUserID finds all categories for a user
func (r *GormCategoryRepository) FindByUserID(ctx context.Context, userID finance.UserID) ([]*finance.Category, error) {
	var categoryModels []Category

	err := r.db.WithContext(ctx).Where("user_id = ? OR user_id IS NULL", userID.Value()).Find(&categoryModels).Error
	if err != nil {
		return nil, err
	}

	// Convert GORM models to domain categories
	var categories []*finance.Category
	for _, model := range categoryModels {
		categoryID := finance.NewCategoryID(int(model.ID))
		var userID *finance.UserID
		if model.UserID != nil {
			userIDVal := finance.NewUserID(int(*model.UserID))
			userID = &userIDVal
		}
		categoryType := finance.CategoryType(model.CategoryType)

		category, err := finance.NewCategory(
			categoryID,
			userID,
			model.Name,
			model.Color,
			model.IsDefault,
			categoryType,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// FindByUserIDAndType finds categories by user ID and type
func (r *GormCategoryRepository) FindByUserIDAndType(ctx context.Context, userID finance.UserID, categoryType finance.CategoryType) ([]*finance.Category, error) {
	var categoryModels []Category

	err := r.db.WithContext(ctx).Where("(user_id = ? OR user_id IS NULL) AND category_type = ?", userID.Value(), string(categoryType)).Find(&categoryModels).Error
	if err != nil {
		return nil, err
	}

	// Convert GORM models to domain categories
	var categories []*finance.Category
	for _, model := range categoryModels {
		categoryID := finance.NewCategoryID(int(model.ID))
		var userID *finance.UserID
		if model.UserID != nil {
			userIDVal := finance.NewUserID(int(*model.UserID))
			userID = &userIDVal
		}

		category, err := finance.NewCategory(
			categoryID,
			userID,
			model.Name,
			model.Color,
			model.IsDefault,
			categoryType,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}

// Delete deletes a category by ID
func (r *GormCategoryRepository) Delete(ctx context.Context, id finance.CategoryID) error {
	return r.db.WithContext(ctx).Delete(&Category{}, id.Value()).Error
}

// ExistsByID checks if a category exists with the given ID
func (r *GormCategoryRepository) ExistsByID(ctx context.Context, id finance.CategoryID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Category{}).Where("id = ?", id.Value()).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// FindDefaultCategories finds all default categories
func (r *GormCategoryRepository) FindDefaultCategories(ctx context.Context) ([]*finance.Category, error) {
	var categoryModels []Category

	err := r.db.WithContext(ctx).Where("is_default = ?", true).Find(&categoryModels).Error
	if err != nil {
		return nil, err
	}

	// Convert GORM models to domain categories
	var categories []*finance.Category
	for _, model := range categoryModels {
		categoryID := finance.NewCategoryID(int(model.ID))
		var userID *finance.UserID
		if model.UserID != nil {
			userIDVal := finance.NewUserID(int(*model.UserID))
			userID = &userIDVal
		}
		categoryType := finance.CategoryType(model.CategoryType)

		category, err := finance.NewCategory(
			categoryID,
			userID,
			model.Name,
			model.Color,
			model.IsDefault,
			categoryType,
		)
		if err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, nil
}
