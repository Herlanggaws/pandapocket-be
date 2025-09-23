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
	categoryIDStr string,
	currencyIDStr string,
	amount float64,
	description string,
	dateStr string,
) (*finance.Transaction, error) {
	// Parse transaction ID
	transactionIDInt, err := strconv.Atoi(transactionIDStr)
	if err != nil {
		return nil, err
	}

	// Parse category ID
	categoryIDInt, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		return nil, err
	}

	// Parse currency ID
	currencyIDInt, err := strconv.Atoi(currencyIDStr)
	if err != nil {
		return nil, err
	}

	// Parse date
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, err
	}

	// Convert to domain types
	transactionID := finance.NewTransactionID(transactionIDInt)
	userIDDomain := finance.NewUserID(userID)
	categoryID := finance.NewCategoryID(categoryIDInt)
	currencyID := finance.NewCurrencyID(currencyIDInt)
	amountDomain, err := finance.NewMoney(amount, currencyID)
	if err != nil {
		return nil, err
	}

	// Update transaction
	return uc.transactionService.UpdateTransaction(
		ctx,
		transactionID,
		userIDDomain,
		categoryID,
		currencyID,
		amountDomain,
		description,
		date,
	)
}
