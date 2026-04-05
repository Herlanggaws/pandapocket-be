package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
)

type UpdateWalletRequest struct {
	Name      string `json:"name" binding:"required"`
	IsPrimary *bool  `json:"is_primary"`
}

type UpdateWalletUseCase struct {
	walletService      *finance.WalletService
	transactionManager finance.TransactionManager
}

func NewUpdateWalletUseCase(walletService *finance.WalletService, transactionManager finance.TransactionManager) *UpdateWalletUseCase {
	return &UpdateWalletUseCase{
		walletService:      walletService,
		transactionManager: transactionManager,
	}
}

func (uc *UpdateWalletUseCase) Execute(ctx context.Context, userID int, walletID int, req UpdateWalletRequest) (*WalletResponse, error) {
	var wallet *finance.Wallet
	err := uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		updatedWallet, err := uc.walletService.UpdateWallet(txCtx, userID, walletID, req.Name, req.IsPrimary)
		if err != nil {
			return err
		}
		wallet = updatedWallet
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &WalletResponse{
		ID:        wallet.ID().Value(),
		UserID:    wallet.UserID().Value(),
		Name:      wallet.Name(),
		Amount:    wallet.Amount(),
		IsPrimary: wallet.IsPrimary(),
		CreatedAt: wallet.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: wallet.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
