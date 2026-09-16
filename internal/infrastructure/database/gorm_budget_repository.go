package database

import (
	"context"
	"errors"
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

func (r *GormBudgetRepository) toDomain(budgetModel Budget) (*finance.Budget, error) {
	amount, err := finance.NewMoney(budgetModel.Amount, finance.NewCurrencyID(int(budgetModel.CurrencyID)))
	if err != nil {
		return nil, err
	}

	return finance.ReconstituteBudget(
		finance.NewBudgetID(int(budgetModel.ID)),
		finance.NewUserID(int(budgetModel.UserID)),
		finance.NewCategoryID(int(budgetModel.CategoryID)),
		amount,
		finance.BudgetPeriod(budgetModel.Period),
		budgetModel.StartDate,
		budgetModel.EndDate,
		budgetModel.CreatedAt,
	)
}

// Save saves a budget to the database
func (r *GormBudgetRepository) Save(ctx context.Context, budget *finance.Budget) error {
	budgetModel := &Budget{
		UserID:     uint(budget.UserID().Value()),
		CategoryID: uint(budget.CategoryID().Value()),
		CurrencyID: uint(budget.Amount().Currency().Value()),
		Amount:     budget.Amount().Amount(),
		Period:     string(budget.Period()),
		StartDate:  budget.StartDate(),
		EndDate:    budget.EndDate(),
	}

	if budget.ID().Value() != 0 {
		budgetModel.ID = uint(budget.ID().Value())
		updates := map[string]interface{}{
			"category_id": budgetModel.CategoryID,
			"currency_id": budgetModel.CurrencyID,
			"amount":      budgetModel.Amount,
			"period":      budgetModel.Period,
			"start_date":  budgetModel.StartDate,
			"end_date":    budgetModel.EndDate,
			"updated_at":  time.Now(),
		}
		return r.db.WithContext(ctx).Model(&Budget{}).Where("id = ?", budgetModel.ID).Updates(updates).Error
	}

	if err := r.db.WithContext(ctx).Create(budgetModel).Error; err != nil {
		return err
	}

	budget.AssignID(finance.NewBudgetID(int(budgetModel.ID)))
	return nil
}

// FindByID finds a budget by ID
func (r *GormBudgetRepository) FindByID(ctx context.Context, id finance.BudgetID) (*finance.Budget, error) {
	var budgetModel Budget

	err := r.db.WithContext(ctx).First(&budgetModel, id.Value()).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}

	return r.toDomain(budgetModel)
}

// FindByUserID finds all budgets for a user
func (r *GormBudgetRepository) FindByUserID(ctx context.Context, userID finance.UserID) ([]*finance.Budget, error) {
	var budgetModels []Budget

	err := r.db.WithContext(ctx).Where("user_id = ?", userID.Value()).Find(&budgetModels).Error
	if err != nil {
		return nil, err
	}

	budgets := make([]*finance.Budget, 0, len(budgetModels))
	for _, model := range budgetModels {
		budget, err := r.toDomain(model)
		if err != nil {
			return nil, err
		}
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

	budgets := make([]*finance.Budget, 0, len(budgetModels))
	for _, model := range budgetModels {
		budget, err := r.toDomain(model)
		if err != nil {
			return nil, err
		}
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

	budgets := make([]*finance.Budget, 0, len(budgetModels))
	for _, model := range budgetModels {
		budget, err := r.toDomain(model)
		if err != nil {
			return nil, err
		}
		budgets = append(budgets, budget)
	}

	return budgets, nil
}

// Delete deletes a budget by ID and user ID
func (r *GormBudgetRepository) Delete(ctx context.Context, id finance.BudgetID) error {
	return r.db.WithContext(ctx).Delete(&Budget{}, id.Value()).Error
}

// DeleteByIDAndUserID deletes a budget scoped to the owning user
func (r *GormBudgetRepository) DeleteByIDAndUserID(ctx context.Context, id finance.BudgetID, userID finance.UserID) error {
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id.Value(), userID.Value()).Delete(&Budget{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
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

// GetTotalCount gets the total count of budgets
func (r *GormBudgetRepository) GetTotalCount(ctx context.Context) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Budget{}).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

// GetCountByDateRange gets the count of budgets created within the date range
func (r *GormBudgetRepository) GetCountByDateRange(ctx context.Context, startDate, endDate time.Time) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Budget{}).
		Where("created_at >= ? AND created_at <= ?", startDate, endDate).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return int(count), nil
}
