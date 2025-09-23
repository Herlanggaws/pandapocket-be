package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
)

// UpdateCurrencyUseCase handles currency updates
type UpdateCurrencyUseCase struct {
	currencyService *finance.CurrencyService
}

// NewUpdateCurrencyUseCase creates a new update currency use case
func NewUpdateCurrencyUseCase(currencyService *finance.CurrencyService) *UpdateCurrencyUseCase {
	return &UpdateCurrencyUseCase{
		currencyService: currencyService,
	}
}

// UpdateCurrencyRequest represents the request to update a currency
type UpdateCurrencyRequest struct {
	Code   string `json:"code" binding:"required"`
	Name   string `json:"name" binding:"required"`
	Symbol string `json:"symbol" binding:"required"`
}

// UpdateCurrencyResponse represents the response after updating a currency
type UpdateCurrencyResponse struct {
	Message  string            `json:"message"`
	Currency *finance.Currency `json:"currency"`
}

// Execute executes the update currency use case
func (uc *UpdateCurrencyUseCase) Execute(ctx context.Context, userID finance.UserID, currencyID finance.CurrencyID, req UpdateCurrencyRequest) (*UpdateCurrencyResponse, error) {
	// Update currency using domain service
	err := uc.currencyService.UpdateCurrency(
		ctx,
		currencyID,
		userID,
		req.Code,
		req.Name,
		req.Symbol,
	)
	if err != nil {
		return nil, err
	}

	// Get updated currency
	currency, err := uc.currencyService.GetCurrenciesByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Find the updated currency
	var updatedCurrency *finance.Currency
	for _, c := range currency {
		if c.ID().Value() == currencyID.Value() {
			updatedCurrency = c
			break
		}
	}

	return &UpdateCurrencyResponse{
		Message:  "Currency updated successfully",
		Currency: updatedCurrency,
	}, nil
}

