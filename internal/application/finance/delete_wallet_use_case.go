package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
)

type DeleteWalletUseCase struct {
	walletService *finance.WalletService
}

func NewDeleteWalletUseCase(walletService *finance.WalletService) *DeleteWalletUseCase {
	return &DeleteWalletUseCase{
		walletService: walletService,
	}
}

func (uc *DeleteWalletUseCase) Execute(ctx context.Context, userID int, walletID int) error {
	return uc.walletService.DeleteWallet(ctx, userID, walletID)
}
