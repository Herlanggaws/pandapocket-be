package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
	"strconv"
	"time"
)

// UpdateTransactionUseCase handles transaction updates
type UpdateTransactionUseCase struct {
	transactionService *finance.TransactionService
}

// NewUpdateTransactionUseCase creates a new update transaction use case
func NewUpdateTransactionUseCase(transactionService *finance.TransactionService) *UpdateTransactionUseCase {
	return &UpdateTransactionUseCase{
		transactionService: transactionService,
	}
}

// Execute updates a transaction
func (uc *UpdateTransactionUseCase) Execute(
	ctx context.Context,
	transactionIDStr string,
	userID int,
	walletID int,
	categoryIDStr string,
	amount float64,
	description string,
	dateStr string,
	expectedType finance.TransactionType,
) (*finance.Transaction, error) {
	transactionIDInt, err := strconv.Atoi(transactionIDStr)
	if err != nil {
		return nil, err
	}

	categoryIDInt, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		return nil, err
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, err
	}

	transactionID := finance.NewTransactionID(transactionIDInt)
	userIDDomain := finance.NewUserID(userID)
	walletIDDomain := finance.NewWalletID(walletID)
	categoryID := finance.NewCategoryID(categoryIDInt)

	return uc.transactionService.UpdateTransaction(
		ctx,
		transactionID,
		userIDDomain,
		walletIDDomain,
		categoryID,
		amount,
		description,
		date,
		expectedType,
	)
}
