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
	Name       string  `json:"name" binding:"required"`
	Amount     float64 `json:"amount"`
	IsApproved *bool   `json:"is_approved"`
}

type CreateWalletUseCase struct {
	walletService      *finance.WalletService
	transactionService *finance.TransactionService
	currencyService    *finance.CurrencyService
	categoryService    *finance.CategoryService
	transactionManager finance.TransactionManager
	userRepo           domainIdentity.UserRepository
}

func NewCreateWalletUseCase(
	walletService *finance.WalletService,
	transactionService *finance.TransactionService,
	currencyService *finance.CurrencyService,
	categoryService *finance.CategoryService,
	transactionManager finance.TransactionManager,
	userRepo domainIdentity.UserRepository,
) *CreateWalletUseCase {
	return &CreateWalletUseCase{
		walletService:      walletService,
		transactionService: transactionService,
		currencyService:    currencyService,
		categoryService:    categoryService,
		transactionManager: transactionManager,
		userRepo:           userRepo,
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

	isApproved := true
	if req.IsApproved != nil {
		isApproved = *req.IsApproved
	}

	var primaryCurrency finance.CurrencyID
	var incomeCategoryID finance.CategoryID
	if req.Amount > 0 {
		primaryCurrencyModel, err := uc.currencyService.GetPrimaryCurrency(ctx, finance.NewUserID(userID))
		if err != nil {
			return nil, errors.New("failed to get primary currency")
		}
		primaryCurrency = primaryCurrencyModel.ID()

		incomeCategories, err := uc.categoryService.GetCategoriesByUserAndType(ctx, finance.NewUserID(userID), finance.CategoryTypeIncome)
		if err != nil || len(incomeCategories) == 0 {
			return nil, errors.New("no income category available")
		}

		for _, cat := range incomeCategories {
			if cat.IsDefault() {
				incomeCategoryID = cat.ID()
				break
			}
		}
		if incomeCategoryID.Value() == 0 {
			incomeCategoryID = incomeCategories[0].ID()
		}
	}

	var wallet *finance.Wallet
	err = uc.transactionManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		createdWallet, err := uc.walletService.CreateWallet(txCtx, userID, req.Name, req.Amount)
		if err != nil {
			return err
		}
		wallet = createdWallet

		if req.Amount > 0 {
			money, err := finance.NewMoney(req.Amount, primaryCurrency)
			if err != nil {
				return err
			}

			walletID := createdWallet.ID()
			_, err = uc.transactionService.CreateTransaction(
				txCtx,
				finance.NewUserID(userID),
				incomeCategoryID,
				primaryCurrency,
				money,
				isApproved,
				"Income from wallet initialization: "+req.Name,
				createdWallet.CreatedAt(),
				finance.TransactionTypeIncome,
				&walletID,
			)
			if err != nil {
				return err
			}
		}

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
		CreatedAt: wallet.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt: wallet.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
