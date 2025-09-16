package finance

import (
	"errors"
	"time"
)

// Currency represents a currency
type Currency struct {
	id        CurrencyID
	userID    *UserID // nil for default currencies
	code      string
	name      string
	symbol    string
	isDefault bool
	createdAt time.Time
}

// NewCurrency creates a new currency
func NewCurrency(
	id CurrencyID,
	userID *UserID,
	code string,
	name string,
	symbol string,
	isDefault bool,
) (*Currency, error) {
	if code == "" {
		return nil, errors.New("currency code cannot be empty")
	}
	
	if name == "" {
		return nil, errors.New("currency name cannot be empty")
	}
	
	if symbol == "" {
		return nil, errors.New("currency symbol cannot be empty")
	}
	
	return &Currency{
		id:        id,
		userID:    userID,
		code:      code,
		name:      name,
		symbol:    symbol,
		isDefault: isDefault,
		createdAt: time.Now(),
	}, nil
}

// Getters
func (c *Currency) ID() CurrencyID {
	return c.id
}

func (c *Currency) UserID() *UserID {
	return c.userID
}

func (c *Currency) Code() string {
	return c.code
}

func (c *Currency) Name() string {
	return c.name
}

func (c *Currency) Symbol() string {
	return c.symbol
}

func (c *Currency) IsDefault() bool {
	return c.isDefault
}

func (c *Currency) CreatedAt() time.Time {
	return c.createdAt
}

// UpdateCode updates the currency code
func (c *Currency) UpdateCode(code string) error {
	if code == "" {
		return errors.New("currency code cannot be empty")
	}
	c.code = code
	return nil
}

// UpdateName updates the currency name
func (c *Currency) UpdateName(name string) error {
	if name == "" {
		return errors.New("currency name cannot be empty")
	}
	c.name = name
	return nil
}

// UpdateSymbol updates the currency symbol
func (c *Currency) UpdateSymbol(symbol string) error {
	if symbol == "" {
		return errors.New("currency symbol cannot be empty")
	}
	c.symbol = symbol
	return nil
}

// CanBeDeleted checks if the currency can be deleted
func (c *Currency) CanBeDeleted() bool {
	return !c.isDefault
}
