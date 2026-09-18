package finance

import (
	"context"
	"errors"
	"panda-pocket/internal/domain/entitlement"
	"panda-pocket/internal/domain/finance"
	"time"
)

// CreateTransactionRequest represents the request to create a transaction
type CreateTransactionRequest struct {
	CategoryID  int     `json:"category_id" binding:"required"`
	WalletID    *int    `json:"wallet_id"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
	Date        string  `json:"date" binding:"required"`
	Type        string  `json:"type"`
}

// CreateTransactionResponse represents the response after creating a transaction
type CreateTransactionResponse struct {
	ID          int     `json:"id"`
	UserID      int     `json:"user_id"`
	WalletID    int     `json:"wallet_id"`
	CategoryID  int     `json:"category_id"`
	CurrencyID  int     `json:"currency_id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
	Type        string  `json:"type"`
	CreatedAt   string  `json:"created_at"`
}

// CreateTransactionUseCase handles transaction creation
type CreateTransactionUseCase struct {
	transactionService *finance.TransactionService
	walletService      *finance.WalletService
	entitlements       entitlement.Checker
}

// NewCreateTransactionUseCase creates a new create transaction use case
func NewCreateTransactionUseCase(
	transactionService *finance.TransactionService,
	walletService *finance.WalletService,
	entitlements entitlement.Checker,
) *CreateTransactionUseCase {
	return &CreateTransactionUseCase{
		transactionService: transactionService,
		walletService:      walletService,
		entitlements:       entitlements,
	}
}

// Execute executes the create transaction use case
func (uc *CreateTransactionUseCase) Execute(ctx context.Context, userID int, req CreateTransactionRequest) (*CreateTransactionResponse, error) {
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("invalid date format. Expected YYYY-MM-DD")
	}

	now := time.Now().UTC()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	monthEnd := monthStart.AddDate(0, 1, 0).Add(-time.Nanosecond)
	_, used, err := uc.transactionService.GetTransactionsByUserWithFilters(
		ctx,
		finance.NewUserID(userID),
		finance.TransactionFilters{StartDate: &monthStart, EndDate: &monthEnd, Limit: 1},
	)
	if err != nil {
		return nil, err
	}
	if err := entitlement.EnforceCreateLimit(
		ctx, uc.entitlements, userID,
		entitlement.FeatureTransactions, int(used), entitlement.FreeTransactionsPerMonth,
	); err != nil {
		return nil, err
	}

	wallet, err := uc.walletService.ResolveUsableWallet(ctx, finance.NewUserID(userID), req.WalletID)
	if err != nil {
		return nil, err
	}

	money, err := finance.NewMoney(req.Amount, wallet.CurrencyID())
	if err != nil {
		return nil, err
	}

	transaction, err := uc.transactionService.CreateTransaction(
		ctx,
		finance.NewUserID(userID),
		wallet.ID(),
		finance.NewCategoryID(req.CategoryID),
		wallet.CurrencyID(),
		money,
		req.Description,
		date,
		finance.TransactionType(req.Type),
	)
	if err != nil {
		return nil, err
	}

	return &CreateTransactionResponse{
		ID:          transaction.ID().Value(),
		UserID:      transaction.UserID().Value(),
		WalletID:    transaction.WalletID().Value(),
		CategoryID:  transaction.CategoryID().Value(),
		CurrencyID:  transaction.CurrencyID().Value(),
		Amount:      transaction.Amount().Amount(),
		Description: transaction.Description(),
		Date:        transaction.Date().Format("2006-01-02"),
		Type:        string(transaction.Type()),
		CreatedAt:   transaction.CreatedAt().Format(time.RFC3339),
	}, nil
}
