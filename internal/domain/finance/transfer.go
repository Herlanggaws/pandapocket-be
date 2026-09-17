package finance

import (
	"errors"
	"time"
)

// TransferID identifies a wallet transfer.
type TransferID struct {
	value int
}

func NewTransferID(id int) TransferID {
	return TransferID{value: id}
}

func (t TransferID) Value() int {
	return t.value
}

// Transfer moves money between two wallets of the same currency.
type Transfer struct {
	id           TransferID
	userID       UserID
	fromWalletID WalletID
	toWalletID   WalletID
	amount       float64
	description  string
	date         time.Time
	createdAt    time.Time
}

func NewTransfer(
	userID UserID,
	fromWalletID WalletID,
	toWalletID WalletID,
	amount float64,
	description string,
	date time.Time,
) (*Transfer, error) {
	if fromWalletID.Value() == toWalletID.Value() {
		return nil, errors.New("from and to wallets must be different")
	}
	if amount <= 0 {
		return nil, errors.New("transfer amount must be positive")
	}
	return &Transfer{
		userID:       userID,
		fromWalletID: fromWalletID,
		toWalletID:   toWalletID,
		amount:       amount,
		description:  description,
		date:         date.Truncate(24 * time.Hour),
		createdAt:    time.Now(),
	}, nil
}

func ReconstituteTransfer(
	id TransferID,
	userID UserID,
	fromWalletID WalletID,
	toWalletID WalletID,
	amount float64,
	description string,
	date time.Time,
	createdAt time.Time,
) *Transfer {
	return &Transfer{
		id:           id,
		userID:       userID,
		fromWalletID: fromWalletID,
		toWalletID:   toWalletID,
		amount:       amount,
		description:  description,
		date:         date,
		createdAt:    createdAt,
	}
}

func (t *Transfer) AssignID(id TransferID) { t.id = id }

func (t *Transfer) ID() TransferID           { return t.id }
func (t *Transfer) UserID() UserID           { return t.userID }
func (t *Transfer) FromWalletID() WalletID   { return t.fromWalletID }
func (t *Transfer) ToWalletID() WalletID     { return t.toWalletID }
func (t *Transfer) Amount() float64          { return t.amount }
func (t *Transfer) Description() string      { return t.description }
func (t *Transfer) Date() time.Time          { return t.date }
func (t *Transfer) CreatedAt() time.Time     { return t.createdAt }

// TransferFilters for listing transfers.
type TransferFilters struct {
	WalletID  *WalletID
	StartDate *time.Time
	EndDate   *time.Time
}
