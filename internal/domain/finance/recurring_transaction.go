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
	id          RecurringTransactionID
	userID      UserID
	categoryID  CategoryID
	currencyID  CurrencyID
	amount      Money
	description string
	frequency   Frequency
	nextDueDate time.Time
	isActive    bool
	createdAt   time.Time
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
	id RecurringTransactionID,
	userID UserID,
	categoryID CategoryID,
	currencyID CurrencyID,
	amount Money,
	description string,
	frequency Frequency,
	nextDueDate time.Time,
) (*RecurringTransaction, error) {
	if amount.Amount() <= 0 {
		return nil, errors.New("recurring transaction amount must be positive")
	}
	
	// Validate frequency
	switch frequency {
	case FrequencyDaily, FrequencyWeekly, FrequencyMonthly, FrequencyYearly:
		// Valid frequencies
	default:
		return nil, errors.New("invalid frequency")
	}
	
	return &RecurringTransaction{
		id:          id,
		userID:      userID,
		categoryID:  categoryID,
		currencyID:  currencyID,
		amount:      amount,
		description: description,
		frequency:   frequency,
		nextDueDate: nextDueDate,
		isActive:    true,
		createdAt:   time.Now(),
	}, nil
}

// Getters
func (r *RecurringTransaction) ID() RecurringTransactionID {
	return r.id
}

func (r *RecurringTransaction) UserID() UserID {
	return r.userID
}

func (r *RecurringTransaction) CategoryID() CategoryID {
	return r.categoryID
}

func (r *RecurringTransaction) CurrencyID() CurrencyID {
	return r.currencyID
}

func (r *RecurringTransaction) Amount() Money {
	return r.amount
}

func (r *RecurringTransaction) Description() string {
	return r.description
}

func (r *RecurringTransaction) Frequency() Frequency {
	return r.frequency
}

func (r *RecurringTransaction) NextDueDate() time.Time {
	return r.nextDueDate
}

func (r *RecurringTransaction) IsActive() bool {
	return r.isActive
}

func (r *RecurringTransaction) CreatedAt() time.Time {
	return r.createdAt
}

// UpdateAmount updates the recurring transaction amount
func (r *RecurringTransaction) UpdateAmount(newAmount Money) error {
	if newAmount.Amount() <= 0 {
		return errors.New("recurring transaction amount must be positive")
	}
	if newAmount.Currency() != r.amount.Currency() {
		return errors.New("cannot change currency of existing recurring transaction")
	}
	r.amount = newAmount
	return nil
}

// UpdateDescription updates the recurring transaction description
func (r *RecurringTransaction) UpdateDescription(description string) {
	r.description = description
}

// UpdateFrequency updates the recurring transaction frequency
func (r *RecurringTransaction) UpdateFrequency(newFrequency Frequency) error {
	switch newFrequency {
	case FrequencyDaily, FrequencyWeekly, FrequencyMonthly, FrequencyYearly:
		r.frequency = newFrequency
		return nil
	default:
		return errors.New("invalid frequency")
	}
}

// UpdateNextDueDate updates the next due date
func (r *RecurringTransaction) UpdateNextDueDate(newDate time.Time) {
	r.nextDueDate = newDate
}

// Activate activates the recurring transaction
func (r *RecurringTransaction) Activate() {
	r.isActive = true
}

// Deactivate deactivates the recurring transaction
func (r *RecurringTransaction) Deactivate() {
	r.isActive = false
}

// CalculateNextDueDate calculates the next due date based on frequency
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

// IsDue checks if the recurring transaction is due
func (r *RecurringTransaction) IsDue() bool {
	return r.isActive && time.Now().After(r.nextDueDate)
}
