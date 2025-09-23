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

// Execute executes the get currencies use case
func (uc *GetCurrenciesUseCase) Execute(ctx context.Context, userID finance.UserID) (*GetCurrenciesResponse, error) {
	// Get currencies using domain service
	currencies, err := uc.currencyService.GetCurrenciesByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &GetCurrenciesResponse{
		Currencies: currencies,
	}, nil
}

