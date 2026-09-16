package database

import (
	"context"
	"errors"
	"panda-pocket/internal/domain/finance"
	"time"

	"gorm.io/gorm"
)

type GormRecurringTransactionRepository struct {
	db *gorm.DB
}

func NewGormRecurringTransactionRepository(db *gorm.DB) *GormRecurringTransactionRepository {
	return &GormRecurringTransactionRepository{db: db}
}

func (r *GormRecurringTransactionRepository) toDomain(model RecurringTransaction) (*finance.RecurringTransaction, error) {
	amount, err := finance.NewMoney(model.Amount, finance.NewCurrencyID(int(model.CurrencyID)))
	if err != nil {
		return nil, err
	}
	txnType := finance.TransactionType(model.Type)
	if txnType == "" {
		txnType = finance.TransactionTypeExpense
	}
	return finance.ReconstituteRecurringTransaction(
		finance.NewRecurringTransactionID(int(model.ID)),
		finance.NewUserID(int(model.UserID)),
		finance.NewCategoryID(int(model.CategoryID)),
		finance.NewCurrencyID(int(model.CurrencyID)),
		amount,
		model.Description,
		finance.Frequency(model.Frequency),
		txnType,
		model.NextDueDate,
		model.IsActive,
		model.CreatedAt,
	), nil
}

func (r *GormRecurringTransactionRepository) Save(ctx context.Context, rt *finance.RecurringTransaction) error {
	model := &RecurringTransaction{
		UserID:      uint(rt.UserID().Value()),
		CategoryID:  uint(rt.CategoryID().Value()),
		CurrencyID:  uint(rt.CurrencyID().Value()),
		Amount:      rt.Amount().Amount(),
		Description: rt.Description(),
		Frequency:   string(rt.Frequency()),
		Type:        string(rt.Type()),
		NextDueDate: rt.NextDueDate(),
		IsActive:    rt.IsActive(),
	}

	if rt.ID().Value() != 0 {
		model.ID = uint(rt.ID().Value())
		return r.db.WithContext(ctx).Model(&RecurringTransaction{}).Where("id = ?", model.ID).Updates(map[string]interface{}{
			"category_id":   model.CategoryID,
			"currency_id":   model.CurrencyID,
			"amount":        model.Amount,
			"description":   model.Description,
			"frequency":     model.Frequency,
			"type":          model.Type,
			"next_due_date": model.NextDueDate,
			"is_active":     model.IsActive,
			"updated_at":    time.Now(),
		}).Error
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	rt.AssignID(finance.NewRecurringTransactionID(int(model.ID)))
	return nil
}

func (r *GormRecurringTransactionRepository) FindByID(ctx context.Context, id finance.RecurringTransactionID) (*finance.RecurringTransaction, error) {
	var model RecurringTransaction
	err := r.db.WithContext(ctx).First(&model, id.Value()).Error
	if err != nil {
		return nil, err
	}
	return r.toDomain(model)
}

func (r *GormRecurringTransactionRepository) FindByUserID(ctx context.Context, userID finance.UserID) ([]*finance.RecurringTransaction, error) {
	var models []RecurringTransaction
	err := r.db.WithContext(ctx).Where("user_id = ?", userID.Value()).Order("next_due_date ASC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	result := make([]*finance.RecurringTransaction, 0, len(models))
	for _, m := range models {
		rt, err := r.toDomain(m)
		if err != nil {
			return nil, err
		}
		result = append(result, rt)
	}
	return result, nil
}

func (r *GormRecurringTransactionRepository) FindActiveByUserID(ctx context.Context, userID finance.UserID) ([]*finance.RecurringTransaction, error) {
	var models []RecurringTransaction
	err := r.db.WithContext(ctx).Where("user_id = ? AND is_active = ?", userID.Value(), true).Find(&models).Error
	if err != nil {
		return nil, err
	}
	result := make([]*finance.RecurringTransaction, 0, len(models))
	for _, m := range models {
		rt, err := r.toDomain(m)
		if err != nil {
			return nil, err
		}
		result = append(result, rt)
	}
	return result, nil
}

func (r *GormRecurringTransactionRepository) FindDueTransactions(ctx context.Context) ([]*finance.RecurringTransaction, error) {
	var models []RecurringTransaction
	today := time.Now().Truncate(24 * time.Hour)
	err := r.db.WithContext(ctx).Where("is_active = ? AND next_due_date <= ?", true, today).Find(&models).Error
	if err != nil {
		return nil, err
	}
	result := make([]*finance.RecurringTransaction, 0, len(models))
	for _, m := range models {
		rt, err := r.toDomain(m)
		if err != nil {
			return nil, err
		}
		result = append(result, rt)
	}
	return result, nil
}

func (r *GormRecurringTransactionRepository) Delete(ctx context.Context, id finance.RecurringTransactionID) error {
	result := r.db.WithContext(ctx).Delete(&RecurringTransaction{}, id.Value())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("recurring transaction not found")
	}
	return nil
}
