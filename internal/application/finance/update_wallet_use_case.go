package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
)

type UpdateWalletRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateWalletUseCase struct {
	walletService *finance.WalletService
}

func NewUpdateWalletUseCase(walletService *finance.WalletService) *UpdateWalletUseCase {
	return &UpdateWalletUseCase{
		walletService: walletService,
	}
}

func (uc *UpdateWalletUseCase) Execute(ctx context.Context, userID int, walletID int, req UpdateWalletRequest) (*WalletResponse, error) {
	wallet, err := uc.walletService.UpdateWalletName(ctx, userID, walletID, req.Name)
	if err != nil {
		return nil, err
	}

	return &WalletResponse{
		ID:        wallet.ID().Value(),
		UserID:    wallet.UserID().Value(),
		Name:      wallet.Name(),
		Amount:    wallet.Amount(),
		CreatedAt: wallet.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: wallet.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
