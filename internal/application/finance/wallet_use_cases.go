package finance

import (
	"context"
	"errors"
	"panda-pocket/internal/domain/entitlement"
	"panda-pocket/internal/domain/finance"
	"time"
)

type WalletResponse struct {
	ID             int     `json:"id"`
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	CurrencyID     int     `json:"currency_id"`
	OpeningBalance float64 `json:"opening_balance"`
	IsDefault      bool    `json:"is_default"`
	IsArchived     bool    `json:"is_archived"`
	Balance        float64 `json:"balance,omitempty"`
	CreatedAt      string  `json:"created_at"`
}

type WalletBalanceResponse struct {
	WalletID       int     `json:"wallet_id"`
	OpeningBalance float64 `json:"opening_balance"`
	TotalIncome    float64 `json:"total_income"`
	TotalExpense   float64 `json:"total_expense"`
	TransfersIn    float64 `json:"transfers_in"`
	TransfersOut   float64 `json:"transfers_out"`
	Balance        float64 `json:"balance"`
}

type CreateWalletRequest struct {
	Name           string  `json:"name" binding:"required"`
	Type           string  `json:"type" binding:"required,oneof=cash bank e_wallet"`
	CurrencyID     int     `json:"currency_id" binding:"required"`
	OpeningBalance float64 `json:"opening_balance"`
	IsDefault      bool    `json:"is_default"`
}

type UpdateWalletRequest struct {
	Name           *string  `json:"name"`
	Type           *string  `json:"type" binding:"omitempty,oneof=cash bank e_wallet"`
	OpeningBalance *float64 `json:"opening_balance"`
}

type CreateTransferRequest struct {
	FromWalletID int     `json:"from_wallet_id" binding:"required"`
	ToWalletID   int     `json:"to_wallet_id" binding:"required"`
	Amount       float64 `json:"amount" binding:"required,gt=0"`
	Description  string  `json:"description"`
	Date         string  `json:"date" binding:"required"`
}

type TransferResponse struct {
	ID           int     `json:"id"`
	FromWalletID int     `json:"from_wallet_id"`
	ToWalletID   int     `json:"to_wallet_id"`
	Amount       float64 `json:"amount"`
	Description  string  `json:"description"`
	Date         string  `json:"date"`
	CreatedAt    string  `json:"created_at"`
}

func toWalletResponse(wallet *finance.Wallet, balance *float64) WalletResponse {
	resp := WalletResponse{
		ID:             wallet.ID().Value(),
		Name:           wallet.Name(),
		Type:           string(wallet.Type()),
		CurrencyID:     wallet.CurrencyID().Value(),
		OpeningBalance: wallet.OpeningBalance(),
		IsDefault:      wallet.IsDefault(),
		IsArchived:     wallet.IsArchived(),
		CreatedAt:      wallet.CreatedAt().Format(time.RFC3339),
	}
	if balance != nil {
		resp.Balance = *balance
	}
	return resp
}

func toTransferResponse(transfer *finance.Transfer) TransferResponse {
	return TransferResponse{
		ID:           transfer.ID().Value(),
		FromWalletID: transfer.FromWalletID().Value(),
		ToWalletID:   transfer.ToWalletID().Value(),
		Amount:       transfer.Amount(),
		Description:  transfer.Description(),
		Date:         transfer.Date().Format("2006-01-02"),
		CreatedAt:    transfer.CreatedAt().Format(time.RFC3339),
	}
}

type CreateWalletUseCase struct {
	walletService *finance.WalletService
	entitlements  entitlement.Checker
}

func NewCreateWalletUseCase(
	walletService *finance.WalletService,
	entitlements entitlement.Checker,
) *CreateWalletUseCase {
	return &CreateWalletUseCase{
		walletService: walletService,
		entitlements:  entitlements,
	}
}

func (uc *CreateWalletUseCase) Execute(ctx context.Context, userID int, req CreateWalletRequest) (*WalletResponse, error) {
	walletType, err := finance.ParseWalletType(req.Type)
	if err != nil {
		return nil, err
	}

	existing, err := uc.walletService.GetWallets(ctx, finance.NewUserID(userID), false)
	if err != nil {
		return nil, err
	}
	if err := entitlement.EnforceCreateLimit(
		ctx, uc.entitlements, userID,
		entitlement.FeatureWallets, len(existing), entitlement.FreeWallets,
	); err != nil {
		return nil, err
	}

	wallet, err := uc.walletService.CreateWallet(
		ctx,
		finance.NewUserID(userID),
		req.Name,
		walletType,
		finance.NewCurrencyID(req.CurrencyID),
		req.OpeningBalance,
		req.IsDefault,
	)
	if err != nil {
		return nil, err
	}
	balance := wallet.OpeningBalance()
	resp := toWalletResponse(wallet, &balance)
	return &resp, nil
}

type GetWalletsUseCase struct {
	walletService *finance.WalletService
}

func NewGetWalletsUseCase(walletService *finance.WalletService) *GetWalletsUseCase {
	return &GetWalletsUseCase{walletService: walletService}
}

func (uc *GetWalletsUseCase) Execute(ctx context.Context, userID int, includeArchived bool) ([]WalletResponse, error) {
	user := finance.NewUserID(userID)
	wallets, err := uc.walletService.GetWallets(ctx, user, includeArchived)
	if err != nil {
		return nil, err
	}
	if len(wallets) == 0 {
		if _, err := uc.walletService.ResolveUsableWallet(ctx, user, nil); err != nil {
			return nil, err
		}
		wallets, err = uc.walletService.GetWallets(ctx, user, includeArchived)
		if err != nil {
			return nil, err
		}
	}
	result := make([]WalletResponse, 0, len(wallets))
	for _, wallet := range wallets {
		breakdown, err := uc.walletService.GetBalance(ctx, user, wallet.ID())
		var balance *float64
		if err == nil {
			b := breakdown.Balance
			balance = &b
		}
		result = append(result, toWalletResponse(wallet, balance))
	}
	return result, nil
}

type GetWalletUseCase struct {
	walletService *finance.WalletService
}

func NewGetWalletUseCase(walletService *finance.WalletService) *GetWalletUseCase {
	return &GetWalletUseCase{walletService: walletService}
}

func (uc *GetWalletUseCase) Execute(ctx context.Context, userID, id int) (*WalletResponse, error) {
	wallet, err := uc.walletService.GetWalletForUser(ctx, finance.NewUserID(userID), finance.NewWalletID(id))
	if err != nil {
		return nil, err
	}
	breakdown, err := uc.walletService.GetBalance(ctx, finance.NewUserID(userID), wallet.ID())
	var balance *float64
	if err == nil {
		b := breakdown.Balance
		balance = &b
	}
	resp := toWalletResponse(wallet, balance)
	return &resp, nil
}

type UpdateWalletUseCase struct {
	walletService *finance.WalletService
}

func NewUpdateWalletUseCase(walletService *finance.WalletService) *UpdateWalletUseCase {
	return &UpdateWalletUseCase{walletService: walletService}
}

func (uc *UpdateWalletUseCase) Execute(ctx context.Context, userID, id int, req UpdateWalletRequest) (*WalletResponse, error) {
	wallet, err := uc.walletService.UpdateWallet(
		ctx,
		finance.NewUserID(userID),
		finance.NewWalletID(id),
		req.Name,
		req.Type,
		req.OpeningBalance,
	)
	if err != nil {
		return nil, err
	}
	breakdown, _ := uc.walletService.GetBalance(ctx, finance.NewUserID(userID), wallet.ID())
	b := breakdown.Balance
	resp := toWalletResponse(wallet, &b)
	return &resp, nil
}

type SetDefaultWalletUseCase struct {
	walletService *finance.WalletService
}

func NewSetDefaultWalletUseCase(walletService *finance.WalletService) *SetDefaultWalletUseCase {
	return &SetDefaultWalletUseCase{walletService: walletService}
}

func (uc *SetDefaultWalletUseCase) Execute(ctx context.Context, userID, id int) (*WalletResponse, error) {
	wallet, err := uc.walletService.SetDefault(ctx, finance.NewUserID(userID), finance.NewWalletID(id))
	if err != nil {
		return nil, err
	}
	breakdown, _ := uc.walletService.GetBalance(ctx, finance.NewUserID(userID), wallet.ID())
	b := breakdown.Balance
	resp := toWalletResponse(wallet, &b)
	return &resp, nil
}

type ArchiveWalletUseCase struct {
	walletService *finance.WalletService
	goalService   *finance.GoalService
}

func NewArchiveWalletUseCase(walletService *finance.WalletService, goalService *finance.GoalService) *ArchiveWalletUseCase {
	return &ArchiveWalletUseCase{walletService: walletService, goalService: goalService}
}

func (uc *ArchiveWalletUseCase) Execute(ctx context.Context, userID, id int) (*WalletResponse, error) {
	user := finance.NewUserID(userID)
	walletID := finance.NewWalletID(id)
	if uc.goalService != nil {
		if err := UnlinkGoalsForWallet(ctx, uc.goalService, uc.walletService, user, walletID); err != nil {
			return nil, err
		}
	}
	wallet, err := uc.walletService.Archive(ctx, user, walletID)
	if err != nil {
		return nil, err
	}
	resp := toWalletResponse(wallet, nil)
	return &resp, nil
}

type UnarchiveWalletUseCase struct {
	walletService *finance.WalletService
}

func NewUnarchiveWalletUseCase(walletService *finance.WalletService) *UnarchiveWalletUseCase {
	return &UnarchiveWalletUseCase{walletService: walletService}
}

func (uc *UnarchiveWalletUseCase) Execute(ctx context.Context, userID, id int) (*WalletResponse, error) {
	wallet, err := uc.walletService.Unarchive(ctx, finance.NewUserID(userID), finance.NewWalletID(id))
	if err != nil {
		return nil, err
	}
	breakdown, _ := uc.walletService.GetBalance(ctx, finance.NewUserID(userID), wallet.ID())
	b := breakdown.Balance
	resp := toWalletResponse(wallet, &b)
	return &resp, nil
}

type GetWalletBalanceUseCase struct {
	walletService *finance.WalletService
}

func NewGetWalletBalanceUseCase(walletService *finance.WalletService) *GetWalletBalanceUseCase {
	return &GetWalletBalanceUseCase{walletService: walletService}
}

func (uc *GetWalletBalanceUseCase) Execute(ctx context.Context, userID, id int) (*WalletBalanceResponse, error) {
	breakdown, err := uc.walletService.GetBalance(ctx, finance.NewUserID(userID), finance.NewWalletID(id))
	if err != nil {
		return nil, err
	}
	return &WalletBalanceResponse{
		WalletID:       id,
		OpeningBalance: breakdown.OpeningBalance,
		TotalIncome:    breakdown.TotalIncome,
		TotalExpense:   breakdown.TotalExpense,
		TransfersIn:    breakdown.TransfersIn,
		TransfersOut:   breakdown.TransfersOut,
		Balance:        breakdown.Balance,
	}, nil
}

type CreateTransferUseCase struct {
	transferService *finance.TransferService
}

func NewCreateTransferUseCase(transferService *finance.TransferService) *CreateTransferUseCase {
	return &CreateTransferUseCase{transferService: transferService}
}

func (uc *CreateTransferUseCase) Execute(ctx context.Context, userID int, req CreateTransferRequest) (*TransferResponse, error) {
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("invalid date format. Expected YYYY-MM-DD")
	}
	transfer, err := uc.transferService.CreateTransfer(
		ctx,
		finance.NewUserID(userID),
		finance.NewWalletID(req.FromWalletID),
		finance.NewWalletID(req.ToWalletID),
		req.Amount,
		req.Description,
		date,
	)
	if err != nil {
		return nil, err
	}
	resp := toTransferResponse(transfer)
	return &resp, nil
}

type GetTransfersUseCase struct {
	transferService *finance.TransferService
}

func NewGetTransfersUseCase(transferService *finance.TransferService) *GetTransfersUseCase {
	return &GetTransfersUseCase{transferService: transferService}
}

func (uc *GetTransfersUseCase) Execute(ctx context.Context, userID int, walletID *int, startDate, endDate *string) ([]TransferResponse, error) {
	filters := finance.TransferFilters{}
	if walletID != nil && *walletID > 0 {
		id := finance.NewWalletID(*walletID)
		filters.WalletID = &id
	}
	if startDate != nil && *startDate != "" {
		parsed, err := time.Parse("2006-01-02", *startDate)
		if err != nil {
			return nil, err
		}
		filters.StartDate = &parsed
	}
	if endDate != nil && *endDate != "" {
		parsed, err := time.Parse("2006-01-02", *endDate)
		if err != nil {
			return nil, err
		}
		filters.EndDate = &parsed
	}

	transfers, err := uc.transferService.ListTransfers(ctx, finance.NewUserID(userID), filters)
	if err != nil {
		return nil, err
	}
	result := make([]TransferResponse, 0, len(transfers))
	for _, transfer := range transfers {
		result = append(result, toTransferResponse(transfer))
	}
	return result, nil
}

type WalletSummaryResponse struct {
	CurrencyID          int     `json:"currency_id"`
	LiquidNetWorth      float64 `json:"liquid_net_worth"`
	WalletCount         int     `json:"wallet_count"`
	ExcludedWalletCount int     `json:"excluded_wallet_count"`
}

type GetWalletSummaryUseCase struct {
	walletService   *finance.WalletService
	currencyService *finance.CurrencyService
}

func NewGetWalletSummaryUseCase(
	walletService *finance.WalletService,
	currencyService *finance.CurrencyService,
) *GetWalletSummaryUseCase {
	return &GetWalletSummaryUseCase{
		walletService:   walletService,
		currencyService: currencyService,
	}
}

func (uc *GetWalletSummaryUseCase) Execute(ctx context.Context, userID int) (*WalletSummaryResponse, error) {
	user := finance.NewUserID(userID)
	primary, err := uc.currencyService.GetPrimaryCurrency(ctx, user)
	if err != nil {
		return nil, err
	}
	primaryCurrencyID := primary.ID().Value()

	wallets, err := uc.walletService.GetWallets(ctx, user, false)
	if err != nil {
		return nil, err
	}

	var liquidNetWorth float64
	walletCount := 0
	excludedCount := 0

	for _, wallet := range wallets {
		if wallet.CurrencyID().Value() != primaryCurrencyID {
			excludedCount++
			continue
		}
		breakdown, balanceErr := uc.walletService.GetBalance(ctx, user, wallet.ID())
		if balanceErr != nil {
			continue
		}
		liquidNetWorth += breakdown.Balance
		walletCount++
	}

	return &WalletSummaryResponse{
		CurrencyID:          primaryCurrencyID,
		LiquidNetWorth:      liquidNetWorth,
		WalletCount:         walletCount,
		ExcludedWalletCount: excludedCount,
	}, nil
}
