package finance

import (
	"context"
	"errors"
	"os"
	"strconv"
	
	"panda-pocket/internal/domain/finance"
	domainIdentity "panda-pocket/internal/domain/identity"
)

type CreateWalletRequest struct {
	Name   string  `json:"name" binding:"required"`
	Amount float64 `json:"amount"`
}

type CreateWalletUseCase struct {
	walletService *finance.WalletService
	userRepo      domainIdentity.UserRepository
}

func NewCreateWalletUseCase(walletService *finance.WalletService, userRepo domainIdentity.UserRepository) *CreateWalletUseCase {
	return &CreateWalletUseCase{
		walletService: walletService,
		userRepo:      userRepo,
	}
}

func (uc *CreateWalletUseCase) Execute(ctx context.Context, userID int, req CreateWalletRequest) (*WalletResponse, error) {
	// Check user wallet limit
	user, err := uc.userRepo.FindByID(ctx, domainIdentity.NewUserID(userID))
	if err != nil {
		return nil, err
	}

	if user.LimitWallet() {
		maxWalletsStr := os.Getenv("MAX_WALLETS_PER_USER")
		maxWallets := 5
		if maxWalletsStr != "" {
			if parsed, err := strconv.Atoi(maxWalletsStr); err == nil {
				maxWallets = parsed
			}
		}

		wallets, err := uc.walletService.GetWalletsByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}

		if len(wallets) >= maxWallets {
			return nil, errors.New("maximum number of wallets reached")
		}
	}

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
