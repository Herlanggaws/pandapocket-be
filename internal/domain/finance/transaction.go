package finance

import (
	"encoding/json"
	"errors"
	"time"
)

// TransactionType represents the type of transaction
type TransactionType string

const (
	TransactionTypeExpense TransactionType = "expense"
	TransactionTypeIncome  TransactionType = "income"
)

// TransactionFilters represents filters for querying transactions
type TransactionFilters struct {
	TransactionType *TransactionType
	CategoryIDs     []CategoryID
	StartDate       *time.Time
	EndDate         *time.Time
	Limit           int
	Offset          int
}

// Transaction represents a financial transaction
type Transaction struct {
	id              TransactionID
	userID          UserID
	categoryID      CategoryID
	currencyID      CurrencyID
	amount          Money
	description     string
	date            time.Time
	transactionType TransactionType
	createdAt       time.Time
}

// TransactionID is a value object representing a transaction identifier
type TransactionID struct {
	value int
}

func NewTransactionID(id int) TransactionID {
	return TransactionID{value: id}
}

func (t TransactionID) Value() int {
	return t.value
}

// UserID is a value object representing a user identifier
type UserID struct {
	value int
}

func NewUserID(id int) UserID {
	return UserID{value: id}
}

func (u UserID) Value() int {
	return u.value
}

// MarshalJSON implements json.Marshaler interface
func (u UserID) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.value)
}

// CategoryID is a value object representing a category identifier
type CategoryID struct {
	value int
}

func NewCategoryID(id int) CategoryID {
	return CategoryID{value: id}
}

func (c CategoryID) Value() int {
	return c.value
}

// CurrencyID is a value object representing a currency identifier
type CurrencyID struct {
	value int
}

func NewCurrencyID(id int) CurrencyID {
	return CurrencyID{value: id}
}

func (c CurrencyID) Value() int {
	return c.value
}

// MarshalJSON implements json.Marshaler interface
func (c CurrencyID) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.value)
}

// Money represents a monetary amount
type Money struct {
	amount   float64
	currency CurrencyID
}

func NewMoney(amount float64, currency CurrencyID) (Money, error) {
	if amount < 0 {
		return Money{}, errors.New("amount cannot be negative")
	}
	return Money{amount: amount, currency: currency}, nil
}

func (m Money) Amount() float64 {
	return m.amount
}

func (m Money) Currency() CurrencyID {
	return m.currency
}

// NewTransaction creates a new transaction
func NewTransaction(
	id TransactionID,
	userID UserID,
	categoryID CategoryID,
	currencyID CurrencyID,
	amount Money,
	description string,
	date time.Time,
	transactionType TransactionType,
) *Transaction {
	return &Transaction{
		id:              id,
		userID:          userID,
		categoryID:      categoryID,
		currencyID:      currencyID,
		amount:          amount,
		description:     description,
		date:            date,
		transactionType: transactionType,
		createdAt:       time.Now(),
	}
}

// Getters
func (t *Transaction) ID() TransactionID {
	return t.id
}

func (t *Transaction) UserID() UserID {
	return t.userID
}

func (t *Transaction) CategoryID() CategoryID {
	return t.categoryID
}

func (t *Transaction) CurrencyID() CurrencyID {
	return t.currencyID
}

func (t *Transaction) Amount() Money {
	return t.amount
}

func (t *Transaction) Description() string {
	return t.description
}

func (t *Transaction) Date() time.Time {
	return t.date
}

func (t *Transaction) Type() TransactionType {
	return t.transactionType
}

func (t *Transaction) CreatedAt() time.Time {
	return t.createdAt
}

// UpdateAmount updates the transaction amount
func (t *Transaction) UpdateAmount(newAmount Money) error {
	if newAmount.Currency() != t.currencyID {
		return errors.New("cannot change currency of existing transaction")
	}
	t.amount = newAmount
	return nil
}

// UpdateDescription updates the transaction description
func (t *Transaction) UpdateDescription(description string) {
	t.description = description
}

// UpdateDate updates the transaction date
func (t *Transaction) UpdateDate(date time.Time) {
	t.date = date
}
