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
	goalService        *finance.GoalService
}

// NewDeleteTransactionUseCase creates a new delete transaction use case
func NewDeleteTransactionUseCase(
	transactionService *finance.TransactionService,
	liabilityService *finance.LiabilityService,
	goalService *finance.GoalService,
) *DeleteTransactionUseCase {
	return &DeleteTransactionUseCase{
		transactionService: transactionService,
		liabilityService:   liabilityService,
		goalService:        goalService,
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

	// Reverse linked side effects before delete (FK SET NULL would orphan refs).
	if expectedType == finance.TransactionTypeExpense {
		if uc.liabilityService != nil {
			if err := uc.liabilityService.ReversePaymentByExpenseID(ctx, userIDDomain, transactionIDInt); err != nil {
				return err
			}
		}
		if uc.goalService != nil {
			if err := uc.goalService.ReverseContributionByExpenseID(ctx, userIDDomain, transactionIDInt); err != nil {
				return err
			}
		}
	}
	if expectedType == finance.TransactionTypeIncome && uc.goalService != nil {
		if err := uc.goalService.ReverseContributionByIncomeID(ctx, userIDDomain, transactionIDInt); err != nil {
			return err
		}
	}

	return uc.transactionService.DeleteTransaction(ctx, transactionID, userIDDomain, expectedType)
}
