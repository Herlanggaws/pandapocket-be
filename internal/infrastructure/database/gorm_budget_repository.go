package database

import (
	"context"
	"panda-pocket/internal/domain/finance"
	"time"

	"gorm.io/gorm"
)

// GormBudgetRepository implements the BudgetRepository interface using GORM
type GormBudgetRepository struct {
	db *gorm.DB
}

// NewGormBudgetRepository creates a new GORM budget repository
func NewGormBudgetRepository(db *gorm.DB) *GormBudgetRepository {
	return &GormBudgetRepository{db: db}
}

// Save saves a budget to the database
func (r *GormBudgetRepository) Save(ctx context.Context, budget *finance.Budget) error {
	// Convert domain budget to GORM model
	budgetModel := &Budget{
		UserID:     uint(budget.UserID().Value()),
		CategoryID: uint(budget.CategoryID().Value()),
		Amount:     budget.Amount().Amount(),
		Period:     string(budget.Period()),
		StartDate:  budget.StartDate(),
		EndDate:    budget.EndDate(),
	}

	if budget.ID().Value() != 0 {
		budgetModel.ID = uint(budget.ID().Value())
	}

	// Save using GORM
	if err := r.db.WithContext(ctx).Save(budgetModel).Error; err != nil {
		return err
	}

	return nil
}

// FindByID finds a budget by ID
func (r *GormBudgetRepository) FindByID(ctx context.Context, id finance.BudgetID) (*finance.Budget, error) {
	var budgetModel Budget

	err := r.db.WithContext(ctx).First(&budgetModel, id.Value()).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}

	// Convert GORM model to domain budget
	budgetID := finance.NewBudgetID(int(budgetModel.ID))
	userID := finance.NewUserID(int(budgetModel.UserID))
	categoryID := finance.NewCategoryID(int(budgetModel.CategoryID))
	amount, _ := finance.NewMoney(budgetModel.Amount, finance.NewCurrencyID(1)) // Default currency ID
	period := finance.BudgetPeriod(budgetModel.Period)

	budget, _ := finance.NewBudget(
		budgetID,
		userID,
		categoryID,
		amount,
		period,
		budgetModel.StartDate,
	)
	// Set the actual end date from database instead of calculated one
	budget.UpdateEndDate(budgetModel.EndDate)

	return budget, nil
}

// FindByUserID finds all budgets for a user
func (r *GormBudgetRepository) FindByUserID(ctx context.Context, userID finance.UserID) ([]*finance.Budget, error) {
	var budgetModels []Budget

	err := r.db.WithContext(ctx).Where("user_id = ?", userID.Value()).Find(&budgetModels).Error
	if err != nil {
		return nil, err
	}

	// Convert GORM models to domain budgets
	var budgets []*finance.Budget
	for _, model := range budgetModels {
		budgetID := finance.NewBudgetID(int(model.ID))
		userIDVO := finance.NewUserID(int(model.UserID))
		categoryID := finance.NewCategoryID(int(model.CategoryID))
		amount, _ := finance.NewMoney(model.Amount, finance.NewCurrencyID(1)) // Default currency ID
		period := finance.BudgetPeriod(model.Period)

		budget, _ := finance.NewBudget(
			budgetID,
			userIDVO,
			categoryID,
			amount,
			period,
			model.StartDate,
		)
		// Set the actual end date from database instead of calculated one
		budget.UpdateEndDate(model.EndDate)
		budgets = append(budgets, budget)
	}

	return budgets, nil
}

// FindByUserIDAndCategory finds budgets by user ID and category
func (r *GormBudgetRepository) FindByUserIDAndCategory(ctx context.Context, userID finance.UserID, categoryID finance.CategoryID) ([]*finance.Budget, error) {
	var budgetModels []Budget

	err := r.db.WithContext(ctx).Where("user_id = ? AND category_id = ?", userID.Value(), categoryID.Value()).Find(&budgetModels).Error
	if err != nil {
		return nil, err
	}

	// Convert GORM models to domain budgets
	var budgets []*finance.Budget
	for _, model := range budgetModels {
		budgetID := finance.NewBudgetID(int(model.ID))
		userIDVO := finance.NewUserID(int(model.UserID))
		categoryIDVO := finance.NewCategoryID(int(model.CategoryID))
		amount, _ := finance.NewMoney(model.Amount, finance.NewCurrencyID(1)) // Default currency ID
		period := finance.BudgetPeriod(model.Period)

		budget, _ := finance.NewBudget(
			budgetID,
			userIDVO,
			categoryIDVO,
			amount,
			period,
			model.StartDate,
		)
		// Set the actual end date from database instead of calculated one
		budget.UpdateEndDate(model.EndDate)
		budgets = append(budgets, budget)
	}

	return budgets, nil
}

// FindActiveByUserID finds active budgets for a user
func (r *GormBudgetRepository) FindActiveByUserID(ctx context.Context, userID finance.UserID) ([]*finance.Budget, error) {
	var budgetModels []Budget
	now := time.Now()

	err := r.db.WithContext(ctx).Where("user_id = ? AND start_date <= ? AND end_date >= ?", userID.Value(), now, now).Find(&budgetModels).Error
	if err != nil {
		return nil, err
	}

	// Convert GORM models to domain budgets
	var budgets []*finance.Budget
	for _, model := range budgetModels {
		budgetID := finance.NewBudgetID(int(model.ID))
		userIDVO := finance.NewUserID(int(model.UserID))
		categoryID := finance.NewCategoryID(int(model.CategoryID))
		amount, _ := finance.NewMoney(model.Amount, finance.NewCurrencyID(1)) // Default currency ID
		period := finance.BudgetPeriod(model.Period)

		budget, _ := finance.NewBudget(
			budgetID,
			userIDVO,
			categoryID,
			amount,
			period,
			model.StartDate,
		)
		// Set the actual end date from database instead of calculated one
		budget.UpdateEndDate(model.EndDate)
		budgets = append(budgets, budget)
	}

	return budgets, nil
}

// Delete deletes a budget by ID
func (r *GormBudgetRepository) Delete(ctx context.Context, id finance.BudgetID) error {
	return r.db.WithContext(ctx).Delete(&Budget{}, id.Value()).Error
}

// ExistsByID checks if a budget exists with the given ID
func (r *GormBudgetRepository) ExistsByID(ctx context.Context, id finance.BudgetID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Budget{}).Where("id = ?", id.Value()).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
