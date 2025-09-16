package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
	"time"
)

// GetTransactionsResponse represents the response for getting transactions
type GetTransactionsResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
}

// TransactionResponse represents a transaction in the response
type TransactionResponse struct {
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

// GetTransactionsUseCase handles getting transactions for a user
type GetTransactionsUseCase struct {
	transactionService *finance.TransactionService
}

// NewGetTransactionsUseCase creates a new get transactions use case
func NewGetTransactionsUseCase(transactionService *finance.TransactionService) *GetTransactionsUseCase {
	return &GetTransactionsUseCase{
		transactionService: transactionService,
	}
}

// Execute executes the get transactions use case
func (uc *GetTransactionsUseCase) Execute(ctx context.Context, userID int) (*GetTransactionsResponse, error) {
	// Get transactions
	transactions, err := uc.transactionService.GetTransactionsByUser(ctx, finance.NewUserID(userID))
	if err != nil {
		return nil, err
	}
	
	// Convert to response format
	transactionResponses := make([]TransactionResponse, len(transactions))
	for i, transaction := range transactions {
		transactionResponses[i] = TransactionResponse{
			ID:          transaction.ID().Value(),
			UserID:      transaction.UserID().Value(),
			CategoryID:  transaction.CategoryID().Value(),
			CurrencyID:  transaction.CurrencyID().Value(),
			Amount:      transaction.Amount().Amount(),
			Description: transaction.Description(),
			Date:        transaction.Date().Format("2006-01-02"),
			Type:        string(transaction.Type()),
			CreatedAt:   transaction.CreatedAt().Format(time.RFC3339),
		}
	}
	
	return &GetTransactionsResponse{
		Transactions: transactionResponses,
	}, nil
}
