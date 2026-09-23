package finance

import (
	"context"
	"errors"
)

// CurrencyService handles currency-related domain operations
type CurrencyService struct {
	currencyRepo CurrencyRepository
}

// NewCurrencyService creates a new currency service
func NewCurrencyService(currencyRepo CurrencyRepository) *CurrencyService {
	return &CurrencyService{
		currencyRepo: currencyRepo,
	}
}

// GetPrimaryCurrency gets the primary currency for a user (user default, then system default)
func (s *CurrencyService) GetPrimaryCurrency(ctx context.Context, userID UserID) (*Currency, error) {
	return s.GetDefaultCurrency(ctx, userID)
}

// GetCurrenciesByUser retrieves all currencies accessible to a user
// (system catalog with user_id IS NULL plus the user's custom currencies).
func (s *CurrencyService) GetCurrenciesByUser(ctx context.Context, userID UserID) ([]*Currency, error) {
	return s.currencyRepo.FindByUserID(ctx, userID)
}

// GetSystemCurrencies retrieves the shared catalog (is_system / user_id IS NULL).
// Used by onboarding before the user has an account.
func (s *CurrencyService) GetSystemCurrencies(ctx context.Context) ([]*Currency, error) {
	return s.currencyRepo.FindSystemCurrencies(ctx)
}

// CreateCurrency creates a new currency
func (s *CurrencyService) CreateCurrency(
	ctx context.Context,
	userID UserID,
	code string,
	name string,
	symbol string,
) (*Currency, error) {
	exists, err := s.currencyRepo.ExistsByCodeAndUserID(ctx, code, userID)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, errors.New("currency code already exists")
	}

	currency, err := NewCurrency(
		CurrencyID{},
		&userID,
		code,
		name,
		symbol,
		false,
	)
	if err != nil {
		return nil, err
	}

	if err := s.currencyRepo.Save(ctx, currency); err != nil {
		return nil, err
	}

	return currency, nil
}

// UpdateCurrency updates a currency
func (s *CurrencyService) UpdateCurrency(
	ctx context.Context,
	currencyID CurrencyID,
	userID UserID,
	code string,
	name string,
	symbol string,
) error {
	currency, err := s.currencyRepo.FindByID(ctx, currencyID)
	if err != nil {
		return errors.New("currency not found")
	}

	if currency.IsSystem() {
		return errors.New("cannot update system currency")
	}

	if currency.UserID() == nil || currency.UserID().Value() != userID.Value() {
		return errors.New("access denied")
	}

	if err := currency.UpdateCode(code); err != nil {
		return err
	}

	if err := currency.UpdateName(name); err != nil {
		return err
	}

	if err := currency.UpdateSymbol(symbol); err != nil {
		return err
	}

	return s.currencyRepo.Save(ctx, currency)
}

// DeleteCurrency deletes a currency
func (s *CurrencyService) DeleteCurrency(ctx context.Context, currencyID CurrencyID, userID UserID) error {
	currency, err := s.currencyRepo.FindByID(ctx, currencyID)
	if err != nil {
		return errors.New("currency not found")
	}

	if !currency.CanBeDeleted() {
		return errors.New("cannot delete system currency")
	}

	if currency.UserID() == nil || currency.UserID().Value() != userID.Value() {
		return errors.New("access denied")
	}

	return s.currencyRepo.Delete(ctx, currencyID)
}

// SetDefaultCurrency sets the primary currency for a user
func (s *CurrencyService) SetDefaultCurrency(ctx context.Context, userID UserID, currencyID CurrencyID) error {
	currency, err := s.currencyRepo.FindByID(ctx, currencyID)
	if err != nil {
		return errors.New("currency not found")
	}

	if !currency.IsSystem() && (currency.UserID() == nil || currency.UserID().Value() != userID.Value()) {
		return errors.New("access denied to currency")
	}

	return s.currencyRepo.SetUserDefaultCurrency(ctx, userID, currencyID)
}

// GetDefaultCurrency gets the primary currency for a user
func (s *CurrencyService) GetDefaultCurrency(ctx context.Context, userID UserID) (*Currency, error) {
	defaultCurrency, err := s.currencyRepo.GetUserDefaultCurrency(ctx, userID)
	if err == nil && defaultCurrency != nil {
		return defaultCurrency, nil
	}

	systemCurrencies, err := s.currencyRepo.FindSystemCurrencies(ctx)
	if err != nil {
		return nil, err
	}

	if len(systemCurrencies) == 0 {
		return nil, errors.New("no system currency found")
	}

	for _, currency := range systemCurrencies {
		if currency.Code() == "IDR" {
			return currency, nil
		}
	}

	return systemCurrencies[0], nil
}
