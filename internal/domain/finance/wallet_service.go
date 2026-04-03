package finance

import (
	"context"
	"errors"
)

type WalletService struct {
	walletRepo WalletRepository
}

func NewWalletService(walletRepo WalletRepository) *WalletService {
	return &WalletService{
		walletRepo: walletRepo,
	}
}

func (s *WalletService) CreateWallet(ctx context.Context, userID int, name string, amount float64) (*Wallet, error) {
	uID := NewUserID(userID)
	wallet, err := NewWallet(NewWalletID(0), uID, name, amount)
	if err != nil {
		return nil, err
	}

	if err := s.walletRepo.Save(ctx, wallet); err != nil {
		return nil, err
	}

	return wallet, nil
}

func (s *WalletService) UpdateWalletName(ctx context.Context, userID int, walletID int, name string) (*Wallet, error) {
	wID := NewWalletID(walletID)
	wallet, err := s.walletRepo.FindByID(ctx, wID)
	if err != nil {
		return nil, err
	}

	// Ensure wallet belongs to user
	if wallet.UserID().Value() != userID {
		return nil, errors.New("forbidden: wallet does not belong to the user")
	}

	if err := wallet.UpdateName(name); err != nil {
		return nil, err
	}

	if err := s.walletRepo.Update(ctx, wallet); err != nil {
		return nil, err
	}

	return wallet, nil
}

func (s *WalletService) GetWalletsByUserID(ctx context.Context, userID int) ([]*Wallet, error) {
	return s.walletRepo.FindByUserID(ctx, NewUserID(userID))
}

func (s *WalletService) GetWalletsByUserIDWithFilters(ctx context.Context, userID int, search string, limit, offset int) ([]*Wallet, int64, error) {
	return s.walletRepo.FindByUserIDWithFilters(ctx, NewUserID(userID), search, limit, offset)
}

func (s *WalletService) DeleteWallet(ctx context.Context, userID int, walletID int) error {
	wID := NewWalletID(walletID)
	wallet, err := s.walletRepo.FindByID(ctx, wID)
	if err != nil {
		return err
	}

	if wallet.UserID().Value() != userID {
		return errors.New("forbidden: wallet does not belong to the user")
	}

	return s.walletRepo.Delete(ctx, wID)
}
