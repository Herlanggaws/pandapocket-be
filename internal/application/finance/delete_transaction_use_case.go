package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
	"strconv"
)

// DeleteTransactionUseCase handles transaction deletion
type DeleteTransactionUseCase struct {
	transactionService *finance.TransactionService
}

// NewDeleteTransactionUseCase creates a new delete transaction use case
func NewDeleteTransactionUseCase(transactionService *finance.TransactionService) *DeleteTransactionUseCase {
	return &DeleteTransactionUseCase{
		transactionService: transactionService,
	}
}

// Execute deletes a transaction
func (uc *DeleteTransactionUseCase) Execute(ctx context.Context, transactionIDStr string, userID int) error {
	// Parse transaction ID
	transactionIDInt, err := strconv.Atoi(transactionIDStr)
	if err != nil {
		return err
	}

	// Convert to domain types
	transactionID := finance.NewTransactionID(transactionIDInt)
	userIDDomain := finance.NewUserID(userID)

	// Delete transaction
	return uc.transactionService.DeleteTransaction(ctx, transactionID, userIDDomain)
}
