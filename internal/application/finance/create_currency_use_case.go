package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
)

// CreateCurrencyUseCase handles currency creation
type CreateCurrencyUseCase struct {
	currencyService *finance.CurrencyService
}

// NewCreateCurrencyUseCase creates a new create currency use case
func NewCreateCurrencyUseCase(currencyService *finance.CurrencyService) *CreateCurrencyUseCase {
	return &CreateCurrencyUseCase{
		currencyService: currencyService,
	}
}

// CreateCurrencyRequest represents the request to create a currency
type CreateCurrencyRequest struct {
	Code   string `json:"code" binding:"required"`
	Name   string `json:"name" binding:"required"`
	Symbol string `json:"symbol" binding:"required"`
}

// CreateCurrencyResponse represents the response after creating a currency
type CreateCurrencyResponse struct {
	Message  string            `json:"message"`
	Currency *finance.Currency `json:"currency"`
}

// Execute executes the create currency use case
func (uc *CreateCurrencyUseCase) Execute(ctx context.Context, userID finance.UserID, req CreateCurrencyRequest) (*CreateCurrencyResponse, error) {
	// Create currency using domain service
	currency, err := uc.currencyService.CreateCurrency(
		ctx,
		userID,
		req.Code,
		req.Name,
		req.Symbol,
	)
	if err != nil {
		return nil, err
	}

	return &CreateCurrencyResponse{
		Message:  "Currency created successfully",
		Currency: currency,
	}, nil
}

