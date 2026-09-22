package finance

import (
	"errors"
	"time"
)

// WalletType classifies how money is held.
type WalletType string

const (
	WalletTypeCash    WalletType = "cash"
	WalletTypeBank    WalletType = "bank"
	WalletTypeEWallet WalletType = "e_wallet"
)

func ParseWalletType(value string) (WalletType, error) {
	switch WalletType(value) {
	case WalletTypeCash, WalletTypeBank, WalletTypeEWallet:
		return WalletType(value), nil
	default:
		return "", errors.New("invalid wallet type")
	}
}

// WalletID identifies a wallet.
type WalletID struct {
	value int
}

func NewWalletID(id int) WalletID {
	return WalletID{value: id}
}

func (w WalletID) Value() int {
	return w.value
}

// Wallet is a user-owned money account.
type Wallet struct {
	id             WalletID
	userID         UserID
	name           string
	walletType     WalletType
	currencyID     CurrencyID
	openingBalance float64
	isDefault      bool
	isArchived     bool
	createdAt      time.Time
}

func NewWallet(
	userID UserID,
	name string,
	walletType WalletType,
	currencyID CurrencyID,
	openingBalance float64,
	isDefault bool,
) (*Wallet, error) {
	if name == "" {
		return nil, errors.New("wallet name cannot be empty")
	}
	if _, err := ParseWalletType(string(walletType)); err != nil {
		return nil, err
	}
	return &Wallet{
		userID:         userID,
		name:           name,
		walletType:     walletType,
		currencyID:     currencyID,
		openingBalance: openingBalance,
		isDefault:      isDefault,
		isArchived:     false,
		createdAt:      time.Now(),
	}, nil
}

func ReconstituteWallet(
	id WalletID,
	userID UserID,
	name string,
	walletType WalletType,
	currencyID CurrencyID,
	openingBalance float64,
	isDefault bool,
	isArchived bool,
	createdAt time.Time,
) *Wallet {
	return &Wallet{
		id:             id,
		userID:         userID,
		name:           name,
		walletType:     walletType,
		currencyID:     currencyID,
		openingBalance: openingBalance,
		isDefault:      isDefault,
		isArchived:     isArchived,
		createdAt:      createdAt,
	}
}

func (w *Wallet) AssignID(id WalletID) { w.id = id }

func (w *Wallet) ID() WalletID             { return w.id }
func (w *Wallet) UserID() UserID           { return w.userID }
func (w *Wallet) Name() string             { return w.name }
func (w *Wallet) Type() WalletType         { return w.walletType }
func (w *Wallet) CurrencyID() CurrencyID   { return w.currencyID }
func (w *Wallet) OpeningBalance() float64  { return w.openingBalance }
func (w *Wallet) IsDefault() bool          { return w.isDefault }
func (w *Wallet) IsArchived() bool         { return w.isArchived }
func (w *Wallet) CreatedAt() time.Time     { return w.createdAt }

func (w *Wallet) UpdateName(name string) error {
	if name == "" {
		return errors.New("wallet name cannot be empty")
	}
	w.name = name
	return nil
}

func (w *Wallet) UpdateType(walletType WalletType) error {
	if _, err := ParseWalletType(string(walletType)); err != nil {
		return err
	}
	w.walletType = walletType
	return nil
}

func (w *Wallet) UpdateOpeningBalance(amount float64) {
	w.openingBalance = amount
}

func (w *Wallet) UpdateCurrencyID(currencyID CurrencyID) {
	w.currencyID = currencyID
}

func (w *Wallet) MarkDefault() {
	w.isDefault = true
}

func (w *Wallet) ClearDefault() {
	w.isDefault = false
}

func (w *Wallet) Archive() error {
	if w.isDefault {
		return errors.New("cannot archive the default wallet; set another default first")
	}
	w.isArchived = true
	return nil
}

func (w *Wallet) Unarchive() {
	w.isArchived = false
}

func (w *Wallet) CanAcceptTransactions() error {
	if w.isArchived {
		return errors.New("wallet is archived")
	}
	return nil
}

// WalletBalanceBreakdown is the ledger composition for a wallet.
type WalletBalanceBreakdown struct {
	OpeningBalance float64
	TotalIncome    float64
	TotalExpense   float64
	TransfersIn    float64
	TransfersOut   float64
	Balance        float64
}

func ComputeWalletBalance(opening, income, expense, transfersIn, transfersOut float64) WalletBalanceBreakdown {
	balance := opening + income - expense + transfersIn - transfersOut
	return WalletBalanceBreakdown{
		OpeningBalance: opening,
		TotalIncome:    income,
		TotalExpense:   expense,
		TransfersIn:    transfersIn,
		TransfersOut:   transfersOut,
		Balance:        balance,
	}
}
