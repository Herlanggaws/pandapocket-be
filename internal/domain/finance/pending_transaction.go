package finance

import (
	"errors"
	"time"
)

// PendingTransactionStatus represents the lifecycle of a pending recurring occurrence.
type PendingTransactionStatus string

const (
	PendingStatusPending   PendingTransactionStatus = "pending"
	PendingStatusConfirmed PendingTransactionStatus = "confirmed"
	PendingStatusRejected  PendingTransactionStatus = "rejected"
)

// PendingTransactionID is a value object for pending transaction identifiers.
type PendingTransactionID struct {
	value int
}

func NewPendingTransactionID(id int) PendingTransactionID {
	return PendingTransactionID{value: id}
}

func (p PendingTransactionID) Value() int {
	return p.value
}

// PendingTransaction is a due recurring occurrence awaiting user confirmation.
type PendingTransaction struct {
	id                     PendingTransactionID
	userID                 UserID
	walletID               WalletID
	recurringTransactionID RecurringTransactionID
	dueDate                time.Time
	amount                 Money
	description            string
	transactionType        TransactionType
	categoryID             CategoryID
	currencyID             CurrencyID
	status                 PendingTransactionStatus
	createdAt              time.Time
	resolvedAt             *time.Time
}

func NewPendingTransaction(
	userID UserID,
	walletID WalletID,
	recurringTransactionID RecurringTransactionID,
	dueDate time.Time,
	amount Money,
	description string,
	transactionType TransactionType,
	categoryID CategoryID,
	currencyID CurrencyID,
) (*PendingTransaction, error) {
	if transactionType != TransactionTypeExpense && transactionType != TransactionTypeIncome {
		return nil, errors.New("invalid transaction type")
	}
	return &PendingTransaction{
		userID:                 userID,
		walletID:               walletID,
		recurringTransactionID: recurringTransactionID,
		dueDate:                dueDate.Truncate(24 * time.Hour),
		amount:                 amount,
		description:            description,
		transactionType:        transactionType,
		categoryID:             categoryID,
		currencyID:             currencyID,
		status:                 PendingStatusPending,
		createdAt:              time.Now(),
	}, nil
}

func ReconstitutePendingTransaction(
	id PendingTransactionID,
	userID UserID,
	walletID WalletID,
	recurringTransactionID RecurringTransactionID,
	dueDate time.Time,
	amount Money,
	description string,
	transactionType TransactionType,
	categoryID CategoryID,
	currencyID CurrencyID,
	status PendingTransactionStatus,
	createdAt time.Time,
	resolvedAt *time.Time,
) *PendingTransaction {
	return &PendingTransaction{
		id:                     id,
		userID:                 userID,
		walletID:               walletID,
		recurringTransactionID: recurringTransactionID,
		dueDate:                dueDate,
		amount:                 amount,
		description:            description,
		transactionType:        transactionType,
		categoryID:             categoryID,
		currencyID:             currencyID,
		status:                 status,
		createdAt:              createdAt,
		resolvedAt:             resolvedAt,
	}
}

func (p *PendingTransaction) AssignID(id PendingTransactionID) {
	p.id = id
}

func (p *PendingTransaction) ID() PendingTransactionID { return p.id }
func (p *PendingTransaction) UserID() UserID           { return p.userID }
func (p *PendingTransaction) WalletID() WalletID       { return p.walletID }
func (p *PendingTransaction) RecurringTransactionID() RecurringTransactionID {
	return p.recurringTransactionID
}
func (p *PendingTransaction) DueDate() time.Time               { return p.dueDate }
func (p *PendingTransaction) Amount() Money                    { return p.amount }
func (p *PendingTransaction) Description() string              { return p.description }
func (p *PendingTransaction) Type() TransactionType            { return p.transactionType }
func (p *PendingTransaction) CategoryID() CategoryID           { return p.categoryID }
func (p *PendingTransaction) CurrencyID() CurrencyID           { return p.currencyID }
func (p *PendingTransaction) Status() PendingTransactionStatus { return p.status }
func (p *PendingTransaction) CreatedAt() time.Time             { return p.createdAt }
func (p *PendingTransaction) ResolvedAt() *time.Time           { return p.resolvedAt }

func (p *PendingTransaction) IsOpen() bool {
	return p.status == PendingStatusPending
}

func (p *PendingTransaction) Confirm() error {
	if !p.IsOpen() {
		return errors.New("pending transaction is already resolved")
	}
	now := time.Now()
	p.status = PendingStatusConfirmed
	p.resolvedAt = &now
	return nil
}

func (p *PendingTransaction) Reject() error {
	if !p.IsOpen() {
		return errors.New("pending transaction is already resolved")
	}
	now := time.Now()
	p.status = PendingStatusRejected
	p.resolvedAt = &now
	return nil
}
