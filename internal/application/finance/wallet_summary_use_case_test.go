package finance

import (
	"context"
	"testing"
	"time"

	domainFinance "panda-pocket/internal/domain/finance"
)

type walletSummaryCurrencyRepo struct {
	primaryID int
}

func (r *walletSummaryCurrencyRepo) Save(ctx context.Context, currency *domainFinance.Currency) error {
	return nil
}
func (r *walletSummaryCurrencyRepo) FindByID(ctx context.Context, id domainFinance.CurrencyID) (*domainFinance.Currency, error) {
	code := "USD"
	symbol := "$"
	if id.Value() == 1 {
		code = "IDR"
		symbol = "Rp"
	}
	return domainFinance.NewCurrency(id, nil, code, code, symbol, true)
}
func (r *walletSummaryCurrencyRepo) FindByUserID(ctx context.Context, userID domainFinance.UserID) ([]*domainFinance.Currency, error) {
	return nil, nil
}
func (r *walletSummaryCurrencyRepo) FindDefaultCurrencies(ctx context.Context) ([]*domainFinance.Currency, error) {
	c, _ := r.FindByID(ctx, domainFinance.NewCurrencyID(r.primaryID))
	return []*domainFinance.Currency{c}, nil
}
func (r *walletSummaryCurrencyRepo) Delete(ctx context.Context, id domainFinance.CurrencyID) error {
	return nil
}
func (r *walletSummaryCurrencyRepo) ExistsByID(ctx context.Context, id domainFinance.CurrencyID) (bool, error) {
	return true, nil
}
func (r *walletSummaryCurrencyRepo) ExistsByCodeAndUserID(ctx context.Context, code string, userID domainFinance.UserID) (bool, error) {
	return false, nil
}
func (r *walletSummaryCurrencyRepo) SetUserDefaultCurrency(ctx context.Context, userID domainFinance.UserID, currencyID domainFinance.CurrencyID) error {
	return nil
}
func (r *walletSummaryCurrencyRepo) GetUserDefaultCurrency(ctx context.Context, userID domainFinance.UserID) (*domainFinance.Currency, error) {
	return r.FindByID(ctx, domainFinance.NewCurrencyID(r.primaryID))
}

type walletSummaryWalletRepo struct {
	wallets   []*domainFinance.Wallet
	balances  map[int]float64
}

func (r *walletSummaryWalletRepo) Save(ctx context.Context, wallet *domainFinance.Wallet) error {
	return nil
}
func (r *walletSummaryWalletRepo) FindByID(ctx context.Context, id domainFinance.WalletID) (*domainFinance.Wallet, error) {
	for _, w := range r.wallets {
		if w.ID().Value() == id.Value() {
			return w, nil
		}
	}
	return nil, context.Canceled
}
func (r *walletSummaryWalletRepo) FindByUserID(ctx context.Context, userID domainFinance.UserID, includeArchived bool) ([]*domainFinance.Wallet, error) {
	return r.wallets, nil
}
func (r *walletSummaryWalletRepo) FindDefaultByUserID(ctx context.Context, userID domainFinance.UserID) (*domainFinance.Wallet, error) {
	if len(r.wallets) == 0 {
		return nil, context.Canceled
	}
	return r.wallets[0], nil
}
func (r *walletSummaryWalletRepo) CountActiveByUserID(ctx context.Context, userID domainFinance.UserID) (int64, error) {
	return int64(len(r.wallets)), nil
}
func (r *walletSummaryWalletRepo) ClearDefaultForUser(ctx context.Context, userID domainFinance.UserID) error {
	return nil
}
func (r *walletSummaryWalletRepo) HasTransactions(ctx context.Context, id domainFinance.WalletID) (bool, error) {
	return false, nil
}
func (r *walletSummaryWalletRepo) GetBalanceBreakdown(ctx context.Context, id domainFinance.WalletID) (domainFinance.WalletBalanceBreakdown, error) {
	balance := r.balances[id.Value()]
	return domainFinance.WalletBalanceBreakdown{Balance: balance}, nil
}
func (r *walletSummaryWalletRepo) AlignPendingCurrency(ctx context.Context, id domainFinance.WalletID, currencyID domainFinance.CurrencyID) error {
	return nil
}

func (r *walletSummaryWalletRepo) AlignRecurringCurrency(ctx context.Context, id domainFinance.WalletID, currencyID domainFinance.CurrencyID) error {
	return nil
}

func mustReconstituteWallet(
	t *testing.T,
	id int,
	currencyID int,
) *domainFinance.Wallet {
	t.Helper()
	return domainFinance.ReconstituteWallet(
		domainFinance.NewWalletID(id),
		domainFinance.NewUserID(1),
		"Wallet",
		domainFinance.WalletTypeCash,
		domainFinance.NewCurrencyID(currencyID),
		0,
		id == 1,
		false,
		time.Now(),
	)
}

func TestGetWalletSummaryUseCase_PrimaryCurrencyOnly(t *testing.T) {
	const primaryCurrencyID = 2

	currencyRepo := &walletSummaryCurrencyRepo{primaryID: primaryCurrencyID}
	walletRepo := &walletSummaryWalletRepo{
		wallets: []*domainFinance.Wallet{
			mustReconstituteWallet(t, 1, primaryCurrencyID),
			mustReconstituteWallet(t, 2, 1),
		},
		balances: map[int]float64{
			1: 500,
			2: 9000,
		},
	}

	walletService := domainFinance.NewWalletService(walletRepo, currencyRepo)
	currencyService := domainFinance.NewCurrencyService(currencyRepo)
	uc := NewGetWalletSummaryUseCase(walletService, currencyService)

	resp, err := uc.Execute(context.Background(), 1)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if resp.CurrencyID != primaryCurrencyID {
		t.Fatalf("currency_id = %d, want %d", resp.CurrencyID, primaryCurrencyID)
	}
	if resp.LiquidNetWorth != 500 {
		t.Fatalf("liquid_net_worth = %v, want 500", resp.LiquidNetWorth)
	}
	if resp.WalletCount != 1 {
		t.Fatalf("wallet_count = %d, want 1", resp.WalletCount)
	}
	if resp.ExcludedWalletCount != 1 {
		t.Fatalf("excluded_wallet_count = %d, want 1", resp.ExcludedWalletCount)
	}
}
