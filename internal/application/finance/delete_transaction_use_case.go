package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
	"strconv"
)

// DeleteTransactionUseCase handles transaction deletion
type DeleteTransactionUseCase struct {
	transactionService *finance.TransactionService
	liabilityService   *finance.LiabilityService
}

// NewDeleteTransactionUseCase creates a new delete transaction use case
func NewDeleteTransactionUseCase(
	transactionService *finance.TransactionService,
	liabilityService *finance.LiabilityService,
) *DeleteTransactionUseCase {
	return &DeleteTransactionUseCase{
		transactionService: transactionService,
		liabilityService:   liabilityService,
	}
}

// Execute deletes a transaction of the expected type
func (uc *DeleteTransactionUseCase) Execute(ctx context.Context, transactionIDStr string, userID int, expectedType finance.TransactionType) error {
	transactionIDInt, err := strconv.Atoi(transactionIDStr)
	if err != nil {
		return err
	}

	transactionID := finance.NewTransactionID(transactionIDInt)
	userIDDomain := finance.NewUserID(userID)

	// Lookup payment by expense_id before deleting the expense (FK SET NULL would orphan the link).
	if expectedType == finance.TransactionTypeExpense && uc.liabilityService != nil {
		if err := uc.liabilityService.ReversePaymentByExpenseID(ctx, userIDDomain, transactionIDInt); err != nil {
			return err
		}
	}

	return uc.transactionService.DeleteTransaction(ctx, transactionID, userIDDomain, expectedType)
}
