package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
)

// GetCurrenciesUseCase handles retrieving currencies
type GetCurrenciesUseCase struct {
	currencyService *finance.CurrencyService
}

// NewGetCurrenciesUseCase creates a new get currencies use case
func NewGetCurrenciesUseCase(currencyService *finance.CurrencyService) *GetCurrenciesUseCase {
	return &GetCurrenciesUseCase{
		currencyService: currencyService,
	}
}

// GetCurrenciesResponse represents the response after getting currencies
type GetCurrenciesResponse struct {
	Currencies []*finance.Currency `json:"currencies"`
}

// Execute executes the get currencies use case for an authenticated user.
func (uc *GetCurrenciesUseCase) Execute(ctx context.Context, userID finance.UserID) (*GetCurrenciesResponse, error) {
	currencies, err := uc.currencyService.GetCurrenciesByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &GetCurrenciesResponse{
		Currencies: currencies,
	}, nil
}

// ExecuteCatalog returns the system currency catalog (no auth / onboarding).
func (uc *GetCurrenciesUseCase) ExecuteCatalog(ctx context.Context) (*GetCurrenciesResponse, error) {
	currencies, err := uc.currencyService.GetSystemCurrencies(ctx)
	if err != nil {
		return nil, err
	}

	return &GetCurrenciesResponse{
		Currencies: currencies,
	}, nil
}

