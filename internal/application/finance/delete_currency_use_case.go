package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
)

// DeleteCurrencyUseCase handles currency deletion
type DeleteCurrencyUseCase struct {
	currencyService *finance.CurrencyService
}

// NewDeleteCurrencyUseCase creates a new delete currency use case
func NewDeleteCurrencyUseCase(currencyService *finance.CurrencyService) *DeleteCurrencyUseCase {
	return &DeleteCurrencyUseCase{
		currencyService: currencyService,
	}
}

// DeleteCurrencyResponse represents the response after deleting a currency
type DeleteCurrencyResponse struct {
	Message string `json:"message"`
}

// Execute executes the delete currency use case
func (uc *DeleteCurrencyUseCase) Execute(ctx context.Context, userID finance.UserID, currencyID finance.CurrencyID) (*DeleteCurrencyResponse, error) {
	// Delete currency using domain service
	err := uc.currencyService.DeleteCurrency(ctx, currencyID, userID)
	if err != nil {
		return nil, err
	}

	return &DeleteCurrencyResponse{
		Message: "Currency deleted successfully",
	}, nil
}

