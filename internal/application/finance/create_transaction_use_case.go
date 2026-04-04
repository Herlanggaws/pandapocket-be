package finance

import (
	"context"
	"errors"
	"panda-pocket/internal/domain/finance"
	"time"
)

// CreateTransactionRequest represents the request to create a transaction
type CreateTransactionRequest struct {
	CategoryID  int     `json:"category_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
	Date        string  `json:"date" binding:"required"`
	IsApproved  *bool   `json:"is_approved"`
	WalletID    *int    `json:"wallet_id"`
	Type        string  `json:"type"`
}

// CreateTransactionResponse represents the response after creating a transaction
type CreateTransactionResponse struct {
	ID          int     `json:"id"`
	UserID      int     `json:"user_id"`
	CategoryID  int     `json:"category_id"`
	CurrencyID  int     `json:"currency_id"`
	WalletID    *int    `json:"wallet_id,omitempty"`
	Amount      float64 `json:"amount"`
	IsApproved  bool    `json:"is_approved"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
	Type        string  `json:"type"`
	CreatedAt   string  `json:"created_at"`
}

// CreateTransactionUseCase handles transaction creation
type CreateTransactionUseCase struct {
	transactionService *finance.TransactionService
	currencyService    *finance.CurrencyService
}

// NewCreateTransactionUseCase creates a new create transaction use case
func NewCreateTransactionUseCase(
	transactionService *finance.TransactionService,
	currencyService *finance.CurrencyService,
) *CreateTransactionUseCase {
	return &CreateTransactionUseCase{
		transactionService: transactionService,
		currencyService:    currencyService,
	}
}

// Execute executes the create transaction use case
func (uc *CreateTransactionUseCase) Execute(ctx context.Context, userID int, req CreateTransactionRequest) (*CreateTransactionResponse, error) {
	// Parse date
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, errors.New("invalid date format. Expected YYYY-MM-DD")
	}

	// Get user's primary currency
	primaryCurrency, err := uc.currencyService.GetPrimaryCurrency(ctx, finance.NewUserID(userID))
	if err != nil {
		return nil, errors.New("failed to get primary currency")
	}

	// Create money value object
	money, err := finance.NewMoney(req.Amount, primaryCurrency.ID())
	if err != nil {
		return nil, err
	}

	isApproved := true
	if req.IsApproved != nil {
		isApproved = *req.IsApproved
	}

	var walletID *finance.WalletID
	if req.WalletID != nil {
		id := finance.NewWalletID(*req.WalletID)
		walletID = &id
	}

	// Create transaction
	transaction, err := uc.transactionService.CreateTransaction(
		ctx,
		finance.NewUserID(userID),
		finance.NewCategoryID(req.CategoryID),
		primaryCurrency.ID(),
		money,
		isApproved,
		req.Description,
		date,
		finance.TransactionType(req.Type),
		walletID,
	)
	if err != nil {
		return nil, err
	}

	walletIDResp := (*int)(nil)
	if transaction.WalletID() != nil {
		wID := transaction.WalletID().Value()
		walletIDResp = &wID
	}

	return &CreateTransactionResponse{
		ID:          transaction.ID().Value(),
		UserID:      transaction.UserID().Value(),
		CategoryID:  transaction.CategoryID().Value(),
		CurrencyID:  transaction.CurrencyID().Value(),
		WalletID:    walletIDResp,
		Amount:      transaction.Amount().Amount(),
		IsApproved:  transaction.IsApproved(),
		Description: transaction.Description(),
		Date:        transaction.Date().Format("2006-01-02"),
		Type:        string(transaction.Type()),
		CreatedAt:   transaction.CreatedAt().Format(time.RFC3339),
	}, nil
}
