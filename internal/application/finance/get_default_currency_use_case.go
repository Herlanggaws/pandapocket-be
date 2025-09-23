package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
)

// GetDefaultCurrencyUseCase handles getting the default currency for a user
type GetDefaultCurrencyUseCase struct {
	currencyService *finance.CurrencyService
}

// NewGetDefaultCurrencyUseCase creates a new GetDefaultCurrencyUseCase
func NewGetDefaultCurrencyUseCase(currencyService *finance.CurrencyService) *GetDefaultCurrencyUseCase {
	return &GetDefaultCurrencyUseCase{
		currencyService: currencyService,
	}
}

// Execute gets the default currency for a user
func (uc *GetDefaultCurrencyUseCase) Execute(
	ctx context.Context,
	userID int,
) (*finance.Currency, error) {
	// Convert to domain types
	userIDDomain := finance.NewUserID(userID)

	// Get default currency
	return uc.currencyService.GetDefaultCurrency(ctx, userIDDomain)
}
