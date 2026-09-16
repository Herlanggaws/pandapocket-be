package finance

import (
	"errors"
	"time"
)

// Frequency represents the frequency of a recurring transaction
type Frequency string

const (
	FrequencyDaily   Frequency = "daily"
	FrequencyWeekly  Frequency = "weekly"
	FrequencyMonthly Frequency = "monthly"
	FrequencyYearly  Frequency = "yearly"
)

// RecurringTransaction represents a recurring transaction
type RecurringTransaction struct {
	id              RecurringTransactionID
	userID          UserID
	categoryID      CategoryID
	currencyID      CurrencyID
	amount          Money
	description     string
	frequency       Frequency
	transactionType TransactionType
	nextDueDate     time.Time
	isActive        bool
	createdAt       time.Time
}

// RecurringTransactionID is a value object representing a recurring transaction identifier
type RecurringTransactionID struct {
	value int
}

func NewRecurringTransactionID(id int) RecurringTransactionID {
	return RecurringTransactionID{value: id}
}

func (r RecurringTransactionID) Value() int {
	return r.value
}

// NewRecurringTransaction creates a new recurring transaction
func NewRecurringTransaction(
	userID UserID,
	categoryID CategoryID,
	currencyID CurrencyID,
	amount Money,
	description string,
	frequency Frequency,
	transactionType TransactionType,
	nextDueDate time.Time,
) (*RecurringTransaction, error) {
	if amount.Amount() <= 0 {
		return nil, errors.New("recurring transaction amount must be positive")
	}

	switch frequency {
	case FrequencyDaily, FrequencyWeekly, FrequencyMonthly, FrequencyYearly:
	default:
		return nil, errors.New("invalid frequency")
	}

	switch transactionType {
	case TransactionTypeExpense, TransactionTypeIncome:
	default:
		return nil, errors.New("invalid transaction type")
	}

	return &RecurringTransaction{
		userID:          userID,
		categoryID:      categoryID,
		currencyID:      currencyID,
		amount:          amount,
		description:     description,
		frequency:       frequency,
		transactionType: transactionType,
		nextDueDate:     nextDueDate,
		isActive:        true,
		createdAt:       time.Now(),
	}, nil
}

func ReconstituteRecurringTransaction(
	id RecurringTransactionID,
	userID UserID,
	categoryID CategoryID,
	currencyID CurrencyID,
	amount Money,
	description string,
	frequency Frequency,
	transactionType TransactionType,
	nextDueDate time.Time,
	isActive bool,
	createdAt time.Time,
) *RecurringTransaction {
	return &RecurringTransaction{
		id:              id,
		userID:          userID,
		categoryID:      categoryID,
		currencyID:      currencyID,
		amount:          amount,
		description:     description,
		frequency:       frequency,
		transactionType: transactionType,
		nextDueDate:     nextDueDate,
		isActive:        isActive,
		createdAt:       createdAt,
	}
}

func (r *RecurringTransaction) AssignID(id RecurringTransactionID) {
	r.id = id
}

func (r *RecurringTransaction) ID() RecurringTransactionID { return r.id }
func (r *RecurringTransaction) UserID() UserID             { return r.userID }
func (r *RecurringTransaction) CategoryID() CategoryID     { return r.categoryID }
func (r *RecurringTransaction) CurrencyID() CurrencyID     { return r.currencyID }
func (r *RecurringTransaction) Amount() Money              { return r.amount }
func (r *RecurringTransaction) Description() string        { return r.description }
func (r *RecurringTransaction) Frequency() Frequency       { return r.frequency }
func (r *RecurringTransaction) Type() TransactionType      { return r.transactionType }
func (r *RecurringTransaction) NextDueDate() time.Time     { return r.nextDueDate }
func (r *RecurringTransaction) IsActive() bool             { return r.isActive }
func (r *RecurringTransaction) CreatedAt() time.Time       { return r.createdAt }

func (r *RecurringTransaction) UpdateNextDueDate(newDate time.Time) {
	r.nextDueDate = newDate
}

func (r *RecurringTransaction) CalculateNextDueDate() time.Time {
	switch r.frequency {
	case FrequencyDaily:
		return r.nextDueDate.AddDate(0, 0, 1)
	case FrequencyWeekly:
		return r.nextDueDate.AddDate(0, 0, 7)
	case FrequencyMonthly:
		return r.nextDueDate.AddDate(0, 1, 0)
	case FrequencyYearly:
		return r.nextDueDate.AddDate(1, 0, 0)
	default:
		return r.nextDueDate
	}
}

// IsDue checks if the recurring transaction is due (inclusive of today)
func (r *RecurringTransaction) IsDue() bool {
	if !r.isActive {
		return false
	}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	due := time.Date(r.nextDueDate.Year(), r.nextDueDate.Month(), r.nextDueDate.Day(), 0, 0, 0, 0, r.nextDueDate.Location())
	return !due.After(today)
}
