package finance

import (
	"errors"
	"time"
)

// WalletID is a value object representing a wallet identifier
type WalletID struct {
	value int
}

func NewWalletID(id int) WalletID {
	return WalletID{value: id}
}

func (w WalletID) Value() int {
	return w.value
}

// Wallet represents a user's wallet
type Wallet struct {
	id        WalletID
	userID    UserID
	name      string
	amount    float64
	createdAt time.Time
	updatedAt time.Time
}

// NewWallet creates a new wallet
func NewWallet(
	id WalletID,
	userID UserID,
	name string,
	amount float64,
) (*Wallet, error) {
	if name == "" {
		return nil, errors.New("wallet name cannot be empty")
	}

	return &Wallet{
		id:        id,
		userID:    userID,
		name:      name,
		amount:    amount,
		createdAt: time.Now(),
		updatedAt: time.Now(),
	}, nil
}

// ReconstructWallet reconstructs an existing wallet from persistence
func ReconstructWallet(
	id WalletID,
	userID UserID,
	name string,
	amount float64,
	createdAt time.Time,
	updatedAt time.Time,
) (*Wallet, error) {
	if name == "" {
		return nil, errors.New("wallet name cannot be empty")
	}

	return &Wallet{
		id:        id,
		userID:    userID,
		name:      name,
		amount:    amount,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}, nil
}

// Getters
func (w *Wallet) ID() WalletID {
	return w.id
}

func (w *Wallet) UserID() UserID {
	return w.userID
}

func (w *Wallet) Name() string {
	return w.name
}

func (w *Wallet) Amount() float64 {
	return w.amount
}

func (w *Wallet) CreatedAt() time.Time {
	return w.createdAt
}

func (w *Wallet) UpdatedAt() time.Time {
	return w.updatedAt
}

// UpdateName updates the wallet name
func (w *Wallet) UpdateName(name string) error {
	if name == "" {
		return errors.New("wallet name cannot be empty")
	}
	w.name = name
	w.updatedAt = time.Now()
	return nil
}

// UpdateAmount updates the wallet amount
func (w *Wallet) UpdateAmount(amount float64) {
	w.amount = amount
	w.updatedAt = time.Now()
}
