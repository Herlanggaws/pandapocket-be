package finance

import (
	"context"
	"testing"
)

func TestEnsureDefaultWallet_CreatesWhenMissing(t *testing.T) {
	currencyRepo := &memCurrencyRepo{byID: map[int]*Currency{
		1: mustSystemCurrency(t, 1, "IDR"),
	}}
	walletRepo := newMemWalletRepo()
	svc := NewWalletService(walletRepo, currencyRepo)

	wallet, created, err := svc.EnsureDefaultWallet(context.Background(), NewUserID(1), NewCurrencyID(1))
	if err != nil {
		t.Fatalf("EnsureDefaultWallet: %v", err)
	}
	if !created {
		t.Fatal("expected created=true when no default wallet exists")
	}
	if wallet == nil || wallet.Type() != WalletTypeCash {
		t.Fatalf("want cash wallet, got %#v", wallet)
	}
	if wallet.ID().Value() == 0 {
		t.Fatal("expected assigned wallet id")
	}
}

func TestEnsureDefaultWallet_AlreadyExists(t *testing.T) {
	existing := emptyWallet(1, 10, 1, 0, true)
	currencyRepo := &memCurrencyRepo{byID: map[int]*Currency{
		1: mustSystemCurrency(t, 1, "IDR"),
	}}
	walletRepo := newMemWalletRepo(existing)
	svc := NewWalletService(walletRepo, currencyRepo)

	wallet, created, err := svc.EnsureDefaultWallet(context.Background(), NewUserID(1), NewCurrencyID(1))
	if err != nil {
		t.Fatalf("EnsureDefaultWallet: %v", err)
	}
	if created {
		t.Fatal("expected created=false when default wallet already exists")
	}
	if wallet == nil || wallet.ID().Value() != 10 {
		t.Fatalf("want existing wallet id 10, got %#v", wallet)
	}
}
