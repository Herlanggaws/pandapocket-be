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
	Type        string  `json:"type"`
}

// CreateTransactionResponse represents the response after creating a transaction
type CreateTransactionResponse struct {
	ID          int     `json:"id"`
	UserID      int     `json:"user_id"`
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

	// Create transaction
	transaction, err := uc.transactionService.CreateTransaction(
		ctx,
		finance.NewUserID(userID),
		finance.NewCategoryID(req.CategoryID),
		primaryCurrency.ID(),
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
		CategoryID:  transaction.CategoryID().Value(),
		CurrencyID:  transaction.CurrencyID().Value(),
		Amount:      transaction.Amount().Amount(),
		Description: transaction.Description(),
		Date:        transaction.Date().Format("2006-01-02"),
		Type:        string(transaction.Type()),
		CreatedAt:   transaction.CreatedAt().Format(time.RFC3339),
	}, nil
}
