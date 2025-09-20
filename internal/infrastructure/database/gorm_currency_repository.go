package database

import (
	"context"
	"panda-pocket/internal/domain/finance"

	"gorm.io/gorm"
)

// GormCurrencyRepository implements the CurrencyRepository interface using GORM
type GormCurrencyRepository struct {
	db *gorm.DB
}

// NewGormCurrencyRepository creates a new GORM currency repository
func NewGormCurrencyRepository(db *gorm.DB) *GormCurrencyRepository {
	return &GormCurrencyRepository{db: db}
}

// Save saves a currency to the database
func (r *GormCurrencyRepository) Save(ctx context.Context, currency *finance.Currency) error {
	// Convert domain currency to GORM model
	currencyModel := &Currency{
		Code:      currency.Code(),
		Name:      currency.Name(),
		Symbol:    currency.Symbol(),
		IsDefault: currency.IsDefault(),
	}

	if currency.ID().Value() != 0 {
		currencyModel.ID = uint(currency.ID().Value())
	}

	if currency.UserID().Value() != 0 {
		userID := uint(currency.UserID().Value())
		currencyModel.UserID = &userID
	}

	// Save using GORM
	if err := r.db.WithContext(ctx).Save(currencyModel).Error; err != nil {
		return err
	}

	return nil
}

// FindByID finds a currency by ID
func (r *GormCurrencyRepository) FindByID(ctx context.Context, id finance.CurrencyID) (*finance.Currency, error) {
	var currencyModel Currency

	err := r.db.WithContext(ctx).First(&currencyModel, id.Value()).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}

	// Convert GORM model to domain currency
	currencyID := finance.NewCurrencyID(int(currencyModel.ID))
	var userID *finance.UserID
	if currencyModel.UserID != nil {
		userIDVal := finance.NewUserID(int(*currencyModel.UserID))
		userID = &userIDVal
	}

	currency, err := finance.NewCurrency(
		currencyID,
		userID,
		currencyModel.Code,
		currencyModel.Name,
		currencyModel.Symbol,
		currencyModel.IsDefault,
	)
	if err != nil {
		return nil, err
	}

	return currency, nil
}

// FindByUserID finds all currencies for a user
func (r *GormCurrencyRepository) FindByUserID(ctx context.Context, userID finance.UserID) ([]*finance.Currency, error) {
	var currencyModels []Currency

	err := r.db.WithContext(ctx).Where("user_id = ? OR user_id IS NULL", userID.Value()).Find(&currencyModels).Error
	if err != nil {
		return nil, err
	}

	// Convert GORM models to domain currencies
	var currencies []*finance.Currency
	for _, model := range currencyModels {
		currencyID := finance.NewCurrencyID(int(model.ID))
		var userID *finance.UserID
		if model.UserID != nil {
			userIDVal := finance.NewUserID(int(*model.UserID))
			userID = &userIDVal
		}

		currency, err := finance.NewCurrency(
			currencyID,
			userID,
			model.Code,
			model.Name,
			model.Symbol,
			model.IsDefault,
		)
		if err != nil {
			return nil, err
		}
		currencies = append(currencies, currency)
	}

	return currencies, nil
}

// FindByCode finds a currency by code
func (r *GormCurrencyRepository) FindByCode(ctx context.Context, code string) (*finance.Currency, error) {
	var currencyModel Currency

	err := r.db.WithContext(ctx).Where("code = ?", code).First(&currencyModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}

	// Convert GORM model to domain currency
	currencyID := finance.NewCurrencyID(int(currencyModel.ID))
	var userID *finance.UserID
	if currencyModel.UserID != nil {
		userIDVal := finance.NewUserID(int(*currencyModel.UserID))
		userID = &userIDVal
	}

	currency, err := finance.NewCurrency(
		currencyID,
		userID,
		currencyModel.Code,
		currencyModel.Name,
		currencyModel.Symbol,
		currencyModel.IsDefault,
	)
	if err != nil {
		return nil, err
	}

	return currency, nil
}

// FindDefaultCurrencies finds all default currencies
func (r *GormCurrencyRepository) FindDefaultCurrencies(ctx context.Context) ([]*finance.Currency, error) {
	var currencyModels []Currency

	err := r.db.WithContext(ctx).Where("is_default = ?", true).Find(&currencyModels).Error
	if err != nil {
		return nil, err
	}

	// Convert GORM models to domain currencies
	var currencies []*finance.Currency
	for _, model := range currencyModels {
		currencyID := finance.NewCurrencyID(int(model.ID))
		var userID *finance.UserID
		if model.UserID != nil {
			userIDVal := finance.NewUserID(int(*model.UserID))
			userID = &userIDVal
		}

		currency, err := finance.NewCurrency(
			currencyID,
			userID,
			model.Code,
			model.Name,
			model.Symbol,
			model.IsDefault,
		)
		if err != nil {
			return nil, err
		}
		currencies = append(currencies, currency)
	}

	return currencies, nil
}

// Delete deletes a currency by ID
func (r *GormCurrencyRepository) Delete(ctx context.Context, id finance.CurrencyID) error {
	return r.db.WithContext(ctx).Delete(&Currency{}, id.Value()).Error
}

// ExistsByID checks if a currency exists with the given ID
func (r *GormCurrencyRepository) ExistsByID(ctx context.Context, id finance.CurrencyID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Currency{}).Where("id = ?", id.Value()).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsByCodeAndUserID checks if a currency exists with the given code and user ID
func (r *GormCurrencyRepository) ExistsByCodeAndUserID(ctx context.Context, code string, userID finance.UserID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Currency{}).Where("code = ? AND (user_id = ? OR user_id IS NULL)", code, userID.Value()).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
