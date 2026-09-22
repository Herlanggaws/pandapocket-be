package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
	"strconv"
)

// SetDefaultCurrencyUseCase handles setting the default currency for a user
type SetDefaultCurrencyUseCase struct {
	currencyService *finance.CurrencyService
	walletService   *finance.WalletService
}

// NewSetDefaultCurrencyUseCase creates a new SetDefaultCurrencyUseCase
func NewSetDefaultCurrencyUseCase(
	currencyService *finance.CurrencyService,
	walletService *finance.WalletService,
) *SetDefaultCurrencyUseCase {
	return &SetDefaultCurrencyUseCase{
		currencyService: currencyService,
		walletService:   walletService,
	}
}

// Execute sets the default currency for a user
func (uc *SetDefaultCurrencyUseCase) Execute(
	ctx context.Context,
	userID int,
	currencyIDStr string,
) error {
	currencyIDInt, err := strconv.Atoi(currencyIDStr)
	if err != nil {
		return err
	}

	userIDDomain := finance.NewUserID(userID)
	currencyID := finance.NewCurrencyID(currencyIDInt)

	if err := uc.currencyService.SetDefaultCurrency(ctx, userIDDomain, currencyID); err != nil {
		return err
	}
	return uc.walletService.SyncDefaultWalletCurrency(ctx, userIDDomain, currencyID)
}
