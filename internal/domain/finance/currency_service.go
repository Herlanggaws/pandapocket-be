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

// GetPrimaryCurrency gets the primary currency for a user
func (s *CurrencyService) GetPrimaryCurrency(ctx context.Context, userID UserID) (*Currency, error) {
	// For now, we'll return the first default currency
	// In a real implementation, you'd get this from user preferences
	defaultCurrencies, err := s.currencyRepo.FindDefaultCurrencies(ctx)
	if err != nil {
		return nil, err
	}

	if len(defaultCurrencies) == 0 {
		return nil, errors.New("no default currency found")
	}

	return defaultCurrencies[0], nil
}

// GetCurrenciesByUser retrieves all currencies accessible to a user
func (s *CurrencyService) GetCurrenciesByUser(ctx context.Context, userID UserID) ([]*Currency, error) {
	// Get user's currencies
	userCurrencies, err := s.currencyRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Get default currencies
	defaultCurrencies, err := s.currencyRepo.FindDefaultCurrencies(ctx)
	if err != nil {
		return nil, err
	}

	// Combine and return
	allCurrencies := append(defaultCurrencies, userCurrencies...)
	return allCurrencies, nil
}

// CreateCurrency creates a new currency
func (s *CurrencyService) CreateCurrency(
	ctx context.Context,
	userID UserID,
	code string,
	name string,
	symbol string,
) (*Currency, error) {
	// Check if currency code already exists for this user
	exists, err := s.currencyRepo.ExistsByCodeAndUserID(ctx, code, userID)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, errors.New("currency code already exists")
	}

	// Create currency
	currency, err := NewCurrency(
		CurrencyID{}, // Will be set by repository
		&userID,
		code,
		name,
		symbol,
		false, // User currencies are not default
	)
	if err != nil {
		return nil, err
	}

	// Save currency
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
	// Get currency
	currency, err := s.currencyRepo.FindByID(ctx, currencyID)
	if err != nil {
		return errors.New("currency not found")
	}

	// Check if user can update this currency
	if currency.IsDefault() {
		return errors.New("cannot update default currency")
	}

	if currency.UserID() == nil || currency.UserID().Value() != userID.Value() {
		return errors.New("access denied")
	}

	// Update currency
	if err := currency.UpdateCode(code); err != nil {
		return err
	}

	if err := currency.UpdateName(name); err != nil {
		return err
	}

	if err := currency.UpdateSymbol(symbol); err != nil {
		return err
	}

	// Save updated currency
	return s.currencyRepo.Save(ctx, currency)
}

// DeleteCurrency deletes a currency
func (s *CurrencyService) DeleteCurrency(ctx context.Context, currencyID CurrencyID, userID UserID) error {
	// Get currency
	currency, err := s.currencyRepo.FindByID(ctx, currencyID)
	if err != nil {
		return errors.New("currency not found")
	}

	// Check if user can delete this currency
	if !currency.CanBeDeleted() {
		return errors.New("cannot delete default currency")
	}

	if currency.UserID() == nil || currency.UserID().Value() != userID.Value() {
		return errors.New("access denied")
	}

	return s.currencyRepo.Delete(ctx, currencyID)
}

// SetDefaultCurrency sets the default currency for a user
func (s *CurrencyService) SetDefaultCurrency(ctx context.Context, userID UserID, currencyID CurrencyID) error {
	// Get currency to verify it exists and user has access
	currency, err := s.currencyRepo.FindByID(ctx, currencyID)
	if err != nil {
		return errors.New("currency not found")
	}

	// Check if user has access to this currency (default or user's own)
	if !currency.IsDefault() && (currency.UserID() == nil || currency.UserID().Value() != userID.Value()) {
		return errors.New("access denied to currency")
	}

	// Set as user's default currency
	return s.currencyRepo.SetUserDefaultCurrency(ctx, userID, currencyID)
}

// GetDefaultCurrency gets the default currency for a user
func (s *CurrencyService) GetDefaultCurrency(ctx context.Context, userID UserID) (*Currency, error) {
	// Try to get user's default currency
	defaultCurrency, err := s.currencyRepo.GetUserDefaultCurrency(ctx, userID)
	if err == nil && defaultCurrency != nil {
		return defaultCurrency, nil
	}

	// If no user default currency is set, return the first default currency
	defaultCurrencies, err := s.currencyRepo.FindDefaultCurrencies(ctx)
	if err != nil {
		return nil, err
	}

	if len(defaultCurrencies) == 0 {
		return nil, errors.New("no default currency found")
	}

	return defaultCurrencies[0], nil
}
