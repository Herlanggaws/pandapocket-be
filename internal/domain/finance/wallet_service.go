package finance

import (
	"context"
	"errors"
	"time"
)

// WalletService handles wallet domain operations.
type WalletService struct {
	walletRepo   WalletRepository
	currencyRepo CurrencyRepository
	linkedGoals  WalletLinkedGoalsChecker
}

// WalletLinkedGoalsChecker reports whether a wallet still has goals linked.
type WalletLinkedGoalsChecker interface {
	HasLinkedGoals(ctx context.Context, walletID WalletID) (bool, error)
}

func NewWalletService(walletRepo WalletRepository, currencyRepo CurrencyRepository) *WalletService {
	return &WalletService{
		walletRepo:   walletRepo,
		currencyRepo: currencyRepo,
	}
}

func (s *WalletService) SetLinkedGoalsChecker(checker WalletLinkedGoalsChecker) {
	s.linkedGoals = checker
}

func (s *WalletService) CreateWallet(
	ctx context.Context,
	userID UserID,
	name string,
	walletType WalletType,
	currencyID CurrencyID,
	openingBalance float64,
	makeDefault bool,
) (*Wallet, error) {
	currency, err := s.currencyRepo.FindByID(ctx, currencyID)
	if err != nil {
		return nil, errors.New("currency not found")
	}
	if !currency.IsSystem() && (currency.UserID() == nil || currency.UserID().Value() != userID.Value()) {
		return nil, errors.New("access denied to currency")
	}

	activeCount, err := s.walletRepo.CountActiveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	isDefault := makeDefault || activeCount == 0

	wallet, err := NewWallet(userID, name, walletType, currencyID, openingBalance, isDefault)
	if err != nil {
		return nil, err
	}

	if isDefault {
		if err := s.walletRepo.ClearDefaultForUser(ctx, userID); err != nil {
			return nil, err
		}
		wallet.MarkDefault()
	}

	if err := s.walletRepo.Save(ctx, wallet); err != nil {
		return nil, err
	}
	return wallet, nil
}

func (s *WalletService) GetWallets(ctx context.Context, userID UserID, includeArchived bool) ([]*Wallet, error) {
	return s.walletRepo.FindByUserID(ctx, userID, includeArchived)
}

func (s *WalletService) GetWalletForUser(ctx context.Context, userID UserID, id WalletID) (*Wallet, error) {
	wallet, err := s.walletRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if wallet.UserID().Value() != userID.Value() {
		return nil, errors.New("wallet not found")
	}
	return wallet, nil
}

func (s *WalletService) GetDefaultWallet(ctx context.Context, userID UserID) (*Wallet, error) {
	return s.walletRepo.FindDefaultByUserID(ctx, userID)
}

func (s *WalletService) EnsureDefaultWallet(ctx context.Context, userID UserID, currencyID CurrencyID) (*Wallet, error) {
	_, err := s.walletRepo.FindDefaultByUserID(ctx, userID)
	if err == nil {
		if syncErr := s.SyncDefaultWalletCurrency(ctx, userID, currencyID); syncErr != nil {
			return nil, syncErr
		}
		return s.GetDefaultWallet(ctx, userID)
	}
	return s.CreateWallet(ctx, userID, "Cash", WalletTypeCash, currencyID, 0, true)
}

func (s *WalletService) ResolveUsableWallet(ctx context.Context, userID UserID, walletID *int) (*Wallet, error) {
	var wallet *Wallet
	var err error
	if walletID != nil && *walletID > 0 {
		wallet, err = s.GetWalletForUser(ctx, userID, NewWalletID(*walletID))
	} else {
		wallet, err = s.GetDefaultWallet(ctx, userID)
		if err != nil {
			primary, cerr := s.currencyRepo.GetUserDefaultCurrency(ctx, userID)
			if cerr != nil {
				return nil, errors.New("default wallet not found")
			}
			wallet, err = s.EnsureDefaultWallet(ctx, userID, primary.ID())
		}
	}
	if err != nil {
		return nil, err
	}
	if err := wallet.CanAcceptTransactions(); err != nil {
		return nil, err
	}
	return wallet, nil
}

func (s *WalletService) UpdateWallet(
	ctx context.Context,
	userID UserID,
	id WalletID,
	name *string,
	walletType *string,
	openingBalance *float64,
	currencyID *int,
) (*Wallet, error) {
	wallet, err := s.GetWalletForUser(ctx, userID, id)
	if err != nil {
		return nil, err
	}

	if name != nil {
		if err := wallet.UpdateName(*name); err != nil {
			return nil, err
		}
	}
	if walletType != nil {
		parsed, err := ParseWalletType(*walletType)
		if err != nil {
			return nil, err
		}
		if err := wallet.UpdateType(parsed); err != nil {
			return nil, err
		}
	}
	if openingBalance != nil {
		wallet.UpdateOpeningBalance(*openingBalance)
	}
	if currencyID != nil {
		if err := s.applyCurrencyChange(ctx, userID, wallet, NewCurrencyID(*currencyID), true); err != nil {
			return nil, err
		}
	}

	if err := s.walletRepo.Save(ctx, wallet); err != nil {
		return nil, err
	}
	if currencyID != nil && wallet.CurrencyID().Value() == *currencyID {
		if err := s.alignDerivedCurrencies(ctx, wallet.ID(), wallet.CurrencyID()); err != nil {
			return nil, err
		}
	}
	return wallet, nil
}

func (s *WalletService) SetDefault(ctx context.Context, userID UserID, id WalletID) (*Wallet, error) {
	wallet, err := s.GetWalletForUser(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if wallet.IsArchived() {
		return nil, errors.New("cannot set archived wallet as default")
	}
	if err := s.walletRepo.ClearDefaultForUser(ctx, userID); err != nil {
		return nil, err
	}
	wallet.MarkDefault()
	if err := s.walletRepo.Save(ctx, wallet); err != nil {
		return nil, err
	}
	return wallet, nil
}

// SyncDefaultWalletCurrency updates the default wallet currency to match primary when the wallet is empty.
func (s *WalletService) SyncDefaultWalletCurrency(ctx context.Context, userID UserID, currencyID CurrencyID) error {
	wallet, err := s.GetDefaultWallet(ctx, userID)
	if err != nil {
		return nil
	}
	if err := s.applyCurrencyChange(ctx, userID, wallet, currencyID, false); err != nil {
		return err
	}
	if wallet.CurrencyID().Value() != currencyID.Value() {
		return nil
	}
	if err := s.walletRepo.Save(ctx, wallet); err != nil {
		return err
	}
	return s.alignDerivedCurrencies(ctx, wallet.ID(), currencyID)
}

// applyCurrencyChange mutates wallet currency when allowed.
// When strict is true, violations return an error; when false (primary sync), skip silently.
func (s *WalletService) applyCurrencyChange(
	ctx context.Context,
	userID UserID,
	wallet *Wallet,
	currencyID CurrencyID,
	strict bool,
) error {
	if wallet.CurrencyID().Value() == currencyID.Value() {
		return nil
	}

	currency, err := s.currencyRepo.FindByID(ctx, currencyID)
	if err != nil {
		return errors.New("currency not found")
	}
	if !currency.IsSystem() && (currency.UserID() == nil || currency.UserID().Value() != userID.Value()) {
		return errors.New("access denied to currency")
	}

	hasTx, err := s.walletRepo.HasTransactions(ctx, wallet.ID())
	if err != nil {
		return err
	}
	if hasTx {
		if strict {
			return errors.New("cannot change currency: wallet has transactions")
		}
		return nil
	}
	if wallet.OpeningBalance() != 0 {
		if strict {
			return errors.New("cannot change currency: wallet has opening balance")
		}
		return nil
	}
	if s.linkedGoals != nil {
		hasGoals, err := s.linkedGoals.HasLinkedGoals(ctx, wallet.ID())
		if err != nil {
			return err
		}
		if hasGoals {
			if strict {
				return errors.New("cannot change currency: unlink goals first")
			}
			return nil
		}
	}

	wallet.UpdateCurrencyID(currencyID)
	return nil
}

func (s *WalletService) alignDerivedCurrencies(ctx context.Context, walletID WalletID, currencyID CurrencyID) error {
	if err := s.walletRepo.AlignPendingCurrency(ctx, walletID, currencyID); err != nil {
		return err
	}
	return s.walletRepo.AlignRecurringCurrency(ctx, walletID, currencyID)
}

func (s *WalletService) Archive(ctx context.Context, userID UserID, id WalletID) (*Wallet, error) {
	wallet, err := s.GetWalletForUser(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	activeCount, err := s.walletRepo.CountActiveByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if activeCount <= 1 {
		return nil, errors.New("cannot archive the last active wallet")
	}
	if err := wallet.Archive(); err != nil {
		return nil, err
	}
	if err := s.walletRepo.Save(ctx, wallet); err != nil {
		return nil, err
	}
	return wallet, nil
}

func (s *WalletService) Unarchive(ctx context.Context, userID UserID, id WalletID) (*Wallet, error) {
	wallet, err := s.GetWalletForUser(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	wallet.Unarchive()
	if err := s.walletRepo.Save(ctx, wallet); err != nil {
		return nil, err
	}
	return wallet, nil
}

func (s *WalletService) GetBalance(ctx context.Context, userID UserID, id WalletID) (WalletBalanceBreakdown, error) {
	wallet, err := s.GetWalletForUser(ctx, userID, id)
	if err != nil {
		return WalletBalanceBreakdown{}, err
	}
	return s.walletRepo.GetBalanceBreakdown(ctx, wallet.ID())
}

// TransferService handles transfers between wallets.
type TransferService struct {
	transferRepo TransferRepository
	walletRepo   WalletRepository
}

func NewTransferService(transferRepo TransferRepository, walletRepo WalletRepository) *TransferService {
	return &TransferService{
		transferRepo: transferRepo,
		walletRepo:   walletRepo,
	}
}

func (s *TransferService) CreateTransfer(
	ctx context.Context,
	userID UserID,
	fromWalletID WalletID,
	toWalletID WalletID,
	amount float64,
	description string,
	date time.Time,
) (*Transfer, error) {
	fromWallet, err := s.walletRepo.FindByID(ctx, fromWalletID)
	if err != nil {
		return nil, errors.New("from wallet not found")
	}
	if fromWallet.UserID().Value() != userID.Value() {
		return nil, errors.New("from wallet not found")
	}
	if err := fromWallet.CanAcceptTransactions(); err != nil {
		return nil, errors.New("from wallet is archived")
	}

	toWallet, err := s.walletRepo.FindByID(ctx, toWalletID)
	if err != nil {
		return nil, errors.New("to wallet not found")
	}
	if toWallet.UserID().Value() != userID.Value() {
		return nil, errors.New("to wallet not found")
	}
	if err := toWallet.CanAcceptTransactions(); err != nil {
		return nil, errors.New("to wallet is archived")
	}

	if fromWallet.CurrencyID().Value() != toWallet.CurrencyID().Value() {
		return nil, errors.New("wallets must share the same currency")
	}

	transfer, err := NewTransfer(userID, fromWalletID, toWalletID, amount, description, date)
	if err != nil {
		return nil, err
	}
	if err := s.transferRepo.Save(ctx, transfer); err != nil {
		return nil, err
	}
	return transfer, nil
}

func (s *TransferService) ListTransfers(ctx context.Context, userID UserID, filters TransferFilters) ([]*Transfer, error) {
	return s.transferRepo.FindByUserID(ctx, userID, filters)
}

func (s *TransferService) GetTransferForUser(ctx context.Context, userID UserID, id TransferID) (*Transfer, error) {
	transfer, err := s.transferRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if transfer.UserID().Value() != userID.Value() {
		return nil, errors.New("transfer not found")
	}
	return transfer, nil
}

func (s *TransferService) DeleteTransfer(ctx context.Context, userID UserID, id TransferID) error {
	if _, err := s.GetTransferForUser(ctx, userID, id); err != nil {
		return err
	}
	return s.transferRepo.Delete(ctx, id)
}
