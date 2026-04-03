package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
)

type CreateWalletRequest struct {
	Name   string  `json:"name" binding:"required"`
	Amount float64 `json:"amount"`
}

type CreateWalletUseCase struct {
	walletService *finance.WalletService
}

func NewCreateWalletUseCase(walletService *finance.WalletService) *CreateWalletUseCase {
	return &CreateWalletUseCase{
		walletService: walletService,
	}
}

func (uc *CreateWalletUseCase) Execute(ctx context.Context, userID int, req CreateWalletRequest) (*WalletResponse, error) {
	wallet, err := uc.walletService.CreateWallet(ctx, userID, req.Name, req.Amount)
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
