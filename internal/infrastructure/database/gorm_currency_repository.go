package database

import (
	"context"
	"errors"
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
	currencyModel := &Currency{
		Code:     currency.Code(),
		Name:     currency.Name(),
		Symbol:   currency.Symbol(),
		IsSystem: currency.IsSystem(),
	}

	if currency.ID().Value() != 0 {
		currencyModel.ID = uint(currency.ID().Value())
	}

	if currency.UserID() != nil && currency.UserID().Value() != 0 {
		userID := uint(currency.UserID().Value())
		currencyModel.UserID = &userID
	}

	if err := r.db.WithContext(ctx).Save(currencyModel).Error; err != nil {
		return err
	}

	return nil
}

func toDomainCurrency(currencyModel Currency) (*finance.Currency, error) {
	currencyID := finance.NewCurrencyID(int(currencyModel.ID))
	var userID *finance.UserID
	if currencyModel.UserID != nil {
		userIDVal := finance.NewUserID(int(*currencyModel.UserID))
		userID = &userIDVal
	}

	return finance.NewCurrency(
		currencyID,
		userID,
		currencyModel.Code,
		currencyModel.Name,
		currencyModel.Symbol,
		currencyModel.IsSystem,
	)
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

	return toDomainCurrency(currencyModel)
}

// FindByUserID finds all currencies for a user
func (r *GormCurrencyRepository) FindByUserID(ctx context.Context, userID finance.UserID) ([]*finance.Currency, error) {
	var currencyModels []Currency

	err := r.db.WithContext(ctx).Where("user_id = ? OR user_id IS NULL", userID.Value()).Find(&currencyModels).Error
	if err != nil {
		return nil, err
	}

	var currencies []*finance.Currency
	for _, model := range currencyModels {
		currency, err := toDomainCurrency(model)
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

	return toDomainCurrency(currencyModel)
}

// FindSystemCurrencies finds system-seeded currencies only (never user-owned rows).
func (r *GormCurrencyRepository) FindSystemCurrencies(ctx context.Context) ([]*finance.Currency, error) {
	var currencyModels []Currency

	err := r.db.WithContext(ctx).
		Where("is_system = ? AND user_id IS NULL", true).
		Find(&currencyModels).Error
	if err != nil {
		return nil, err
	}

	var currencies []*finance.Currency
	for _, model := range currencyModels {
		currency, err := toDomainCurrency(model)
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

// SetUserDefaultCurrency sets the default currency for a user
func (r *GormCurrencyRepository) SetUserDefaultCurrency(ctx context.Context, userID finance.UserID, currencyID finance.CurrencyID) error {
	var preferences UserPreferences

	err := r.db.WithContext(ctx).Where("user_id = ?", userID.Value()).First(&preferences).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			preferences = UserPreferences{
				UserID:             uint(userID.Value()),
				PrimaryCurrencyID:  uint(currencyID.Value()),
				EmailNotifications: true,
				BudgetAlerts:       true,
				RecurringReminders: true,
				Onboarding:         JSONRaw("{}"),
			}
			return r.db.WithContext(ctx).Create(&preferences).Error
		}
		return err
	}

	preferences.PrimaryCurrencyID = uint(currencyID.Value())
	return r.db.WithContext(ctx).Save(&preferences).Error
}

// GetUserDefaultCurrency gets the default currency for a user
func (r *GormCurrencyRepository) GetUserDefaultCurrency(ctx context.Context, userID finance.UserID) (*finance.Currency, error) {
	var preferences UserPreferences

	err := r.db.WithContext(ctx).Where("user_id = ?", userID.Value()).First(&preferences).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("no default currency set")
		}
		return nil, err
	}

	return r.FindByID(ctx, finance.NewCurrencyID(int(preferences.PrimaryCurrencyID)))
}
