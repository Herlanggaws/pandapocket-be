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
}

func NewWalletService(walletRepo WalletRepository, currencyRepo CurrencyRepository) *WalletService {
	return &WalletService{
		walletRepo:   walletRepo,
		currencyRepo: currencyRepo,
	}
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
	if !currency.IsDefault() && (currency.UserID() == nil || currency.UserID().Value() != userID.Value()) {
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
	wallet, err := s.walletRepo.FindDefaultByUserID(ctx, userID)
	if err == nil {
		return wallet, nil
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

	if err := s.walletRepo.Save(ctx, wallet); err != nil {
		return nil, err
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

// SyncDefaultWalletCurrency updates the default wallet currency when it has no ledger history.
func (s *WalletService) SyncDefaultWalletCurrency(ctx context.Context, userID UserID, currencyID CurrencyID) error {
	wallet, err := s.GetDefaultWallet(ctx, userID)
	if err != nil {
		return nil
	}
	if wallet.CurrencyID().Value() == currencyID.Value() {
		return nil
	}

	currency, err := s.currencyRepo.FindByID(ctx, currencyID)
	if err != nil {
		return errors.New("currency not found")
	}
	if !currency.IsDefault() && (currency.UserID() == nil || currency.UserID().Value() != userID.Value()) {
		return errors.New("access denied to currency")
	}

	hasTx, err := s.walletRepo.HasTransactions(ctx, wallet.ID())
	if err != nil {
		return err
	}
	if hasTx {
		return nil
	}

	wallet.UpdateCurrencyID(currencyID)
	return s.walletRepo.Save(ctx, wallet)
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
