package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
	"strconv"
)

// SetDefaultCurrencyUseCase handles setting the default currency for a user
type SetDefaultCurrencyUseCase struct {
	currencyService *finance.CurrencyService
}

// NewSetDefaultCurrencyUseCase creates a new SetDefaultCurrencyUseCase
func NewSetDefaultCurrencyUseCase(currencyService *finance.CurrencyService) *SetDefaultCurrencyUseCase {
	return &SetDefaultCurrencyUseCase{
		currencyService: currencyService,
	}
}

// Execute sets the default currency for a user
func (uc *SetDefaultCurrencyUseCase) Execute(
	ctx context.Context,
	userID int,
	currencyIDStr string,
) error {
	// Parse currency ID
	currencyIDInt, err := strconv.Atoi(currencyIDStr)
	if err != nil {
		return err
	}

	// Convert to domain types
	userIDDomain := finance.NewUserID(userID)
	currencyID := finance.NewCurrencyID(currencyIDInt)

	// Set default currency
	return uc.currencyService.SetDefaultCurrency(ctx, userIDDomain, currencyID)
}
