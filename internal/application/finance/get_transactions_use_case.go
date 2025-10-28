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

// GetTransactionsUseCase handles getting transactions for a user
type GetTransactionsUseCase struct {
	transactionService *finance.TransactionService
	categoryService    *finance.CategoryService
}

// NewGetTransactionsUseCase creates a new get transactions use case
func NewGetTransactionsUseCase(transactionService *finance.TransactionService, categoryService *finance.CategoryService) *GetTransactionsUseCase {
	return &GetTransactionsUseCase{
		transactionService: transactionService,
		categoryService:    categoryService,
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
		// Fetch category details
		category, err := uc.categoryService.GetCategoryByID(ctx, transaction.CategoryID())
		if err != nil {
			// If category not found, create a default response
			category = &finance.Category{}
		}

		transactionResponses[i] = TransactionResponse{
			ID:     transaction.ID().Value(),
			UserID: transaction.UserID().Value(),
			Category: CategoryResponse{
				ID:        category.ID().Value(),
				Name:      category.Name(),
				Color:     category.Color(),
				Type:      string(category.Type()),
				IsDefault: category.IsDefault(),
			},
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
