package finance

import (
	"context"
	"math"
	"testing"
	"time"

	domainFinance "panda-pocket/internal/domain/finance"
)

type netWorthCurrencyRepo struct {
	currencies []*domainFinance.Currency
}

func (r *netWorthCurrencyRepo) Save(ctx context.Context, currency *domainFinance.Currency) error {
	return nil
}
func (r *netWorthCurrencyRepo) FindByID(ctx context.Context, id domainFinance.CurrencyID) (*domainFinance.Currency, error) {
	for _, currency := range r.currencies {
		if currency.ID().Value() == id.Value() {
			return currency, nil
		}
	}
	return nil, context.Canceled
}
func (r *netWorthCurrencyRepo) FindByUserID(ctx context.Context, userID domainFinance.UserID) ([]*domainFinance.Currency, error) {
	return r.currencies, nil
}
func (r *netWorthCurrencyRepo) FindSystemCurrencies(ctx context.Context) ([]*domainFinance.Currency, error) {
	return r.currencies, nil
}
func (r *netWorthCurrencyRepo) Delete(ctx context.Context, id domainFinance.CurrencyID) error {
	return nil
}
func (r *netWorthCurrencyRepo) ExistsByID(ctx context.Context, id domainFinance.CurrencyID) (bool, error) {
	return true, nil
}
func (r *netWorthCurrencyRepo) ExistsByCodeAndUserID(ctx context.Context, code string, userID domainFinance.UserID) (bool, error) {
	return false, nil
}
func (r *netWorthCurrencyRepo) SetUserDefaultCurrency(ctx context.Context, userID domainFinance.UserID, currencyID domainFinance.CurrencyID) error {
	return nil
}
func (r *netWorthCurrencyRepo) GetUserDefaultCurrency(ctx context.Context, userID domainFinance.UserID) (*domainFinance.Currency, error) {
	return r.currencies[0], nil
}

type netWorthAssetRepo struct {
	assets []*domainFinance.Asset
}

func (r *netWorthAssetRepo) Save(ctx context.Context, asset *domainFinance.Asset) error { return nil }
func (r *netWorthAssetRepo) FindByID(ctx context.Context, id domainFinance.AssetID) (*domainFinance.Asset, error) {
	return nil, context.Canceled
}
func (r *netWorthAssetRepo) FindByUserID(ctx context.Context, userID domainFinance.UserID, includeArchived bool) ([]*domainFinance.Asset, error) {
	return r.assets, nil
}

type netWorthLiabilityRepo struct {
	liabilities []*domainFinance.Liability
}

func (r *netWorthLiabilityRepo) Save(ctx context.Context, liability *domainFinance.Liability) error {
	return nil
}
func (r *netWorthLiabilityRepo) FindByID(ctx context.Context, id domainFinance.LiabilityID) (*domainFinance.Liability, error) {
	return nil, context.Canceled
}
func (r *netWorthLiabilityRepo) FindByUserID(ctx context.Context, userID domainFinance.UserID, includeArchived bool) ([]*domainFinance.Liability, error) {
	return r.liabilities, nil
}

type netWorthPaymentRepo struct{}

func (r *netWorthPaymentRepo) Save(ctx context.Context, payment *domainFinance.LiabilityPayment) error {
	return nil
}
func (r *netWorthPaymentRepo) FindByLiabilityID(ctx context.Context, liabilityID domainFinance.LiabilityID) ([]*domainFinance.LiabilityPayment, error) {
	return nil, nil
}
func (r *netWorthPaymentRepo) FindByExpenseID(ctx context.Context, expenseID int) (*domainFinance.LiabilityPayment, error) {
	return nil, context.Canceled
}
func (r *netWorthPaymentRepo) Delete(ctx context.Context, id domainFinance.LiabilityPaymentID) error {
	return nil
}

func TestGetNetWorthSummary_ConvertsSystemCurrencies(t *testing.T) {
	owner := domainFinance.NewUserID(1)
	idr, _ := domainFinance.NewCurrency(domainFinance.NewCurrencyID(1), nil, "IDR", "Rupiah", "Rp", true)
	usd, _ := domainFinance.NewCurrency(domainFinance.NewCurrencyID(2), nil, "USD", "Dollar", "$", true)
	custom, _ := domainFinance.NewCurrency(domainFinance.NewCurrencyID(3), &owner, "GOLD", "Gold", "G", false)

	currencyRepo := &netWorthCurrencyRepo{currencies: []*domainFinance.Currency{idr, usd, custom}}
	walletRepo := &walletSummaryWalletRepo{
		wallets: []*domainFinance.Wallet{
			mustReconstituteWallet(t, 1, 1),
			mustReconstituteWallet(t, 2, 2),
			mustReconstituteWallet(t, 3, 3),
		},
		balances: map[int]float64{1: 1000, 2: 10, 3: 50},
	}
	assetRepo := &netWorthAssetRepo{assets: []*domainFinance.Asset{
		domainFinance.ReconstituteAsset(domainFinance.NewAssetID(1), owner, "House", domainFinance.AssetTypeProperty, domainFinance.NewCurrencyID(1), 500, "", false, nil, time.Now()),
		domainFinance.ReconstituteAsset(domainFinance.NewAssetID(2), owner, "Broker", domainFinance.AssetTypeInvestment, domainFinance.NewCurrencyID(2), 2, "", false, nil, time.Now()),
		domainFinance.ReconstituteAsset(domainFinance.NewAssetID(3), owner, "Coin", domainFinance.AssetTypeOther, domainFinance.NewCurrencyID(3), 9, "", false, nil, time.Now()),
	}}
	liabilityRepo := &netWorthLiabilityRepo{liabilities: []*domainFinance.Liability{
		domainFinance.ReconstituteLiability(domainFinance.NewLiabilityID(1), owner, "Card", domainFinance.LiabilityTypeLoan, domainFinance.NewCurrencyID(2), 1, "", false, nil, time.Now(), domainFinance.LiabilityDebtDetails{}),
		domainFinance.ReconstituteLiability(domainFinance.NewLiabilityID(2), owner, "Unknown", domainFinance.LiabilityTypeOther, domainFinance.NewCurrencyID(99), 4, "", false, nil, time.Now(), domainFinance.LiabilityDebtDetails{}),
	}}

	asOf := time.Date(2026, 9, 26, 0, 0, 0, 0, time.UTC)
	fxRepo := &memoryFxRateRepo{quotes: []domainFinance.FxQuote{
		{Code: "EUR", UnitsPerEUR: 1, AsOf: asOf},
		{Code: "USD", UnitsPerEUR: 1.17, AsOf: asOf},
		{Code: "IDR", UnitsPerEUR: 17800, AsOf: asOf},
	}}

	currencyService := domainFinance.NewCurrencyService(currencyRepo)
	walletService := domainFinance.NewWalletService(walletRepo, currencyRepo)
	uc := NewGetNetWorthSummaryUseCase(
		NewGetWalletSummaryUseCase(walletService, currencyService),
		walletService,
		domainFinance.NewAssetService(assetRepo),
		nil,
		domainFinance.NewLiabilityService(liabilityRepo, &netWorthPaymentRepo{}),
		currencyService,
		fxRepo,
	)

	resp, err := uc.Execute(context.Background(), 1)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	usdToIDR := 17800.0 / 1.17
	wantLiquid := 1000 + 10*usdToIDR
	wantAssets := 500 + 2*usdToIDR
	wantLiabilities := usdToIDR
	if math.Abs(resp.LiquidNetWorth-wantLiquid) > 0.01 {
		t.Fatalf("liquid = %v, want %v", resp.LiquidNetWorth, wantLiquid)
	}
	if math.Abs(resp.AssetsTotal-wantAssets) > 0.01 {
		t.Fatalf("assets = %v, want %v", resp.AssetsTotal, wantAssets)
	}
	if math.Abs(resp.LiabilitiesTotal-wantLiabilities) > 0.01 {
		t.Fatalf("liabilities = %v, want %v", resp.LiabilitiesTotal, wantLiabilities)
	}
	if math.Abs(resp.NetWorth-(wantLiquid+wantAssets-wantLiabilities)) > 0.01 {
		t.Fatalf("net worth = %v", resp.NetWorth)
	}
	if resp.ConvertedWalletCount != 1 || resp.ConvertedAssetCount != 1 || resp.ConvertedLiabilityCount != 1 {
		t.Fatalf("converted counts = %d/%d/%d", resp.ConvertedWalletCount, resp.ConvertedAssetCount, resp.ConvertedLiabilityCount)
	}
	if resp.ExcludedWalletCount != 1 || resp.ExcludedAssetCount != 1 || resp.ExcludedLiabilityCount != 1 {
		t.Fatalf("excluded counts = %d/%d/%d", resp.ExcludedWalletCount, resp.ExcludedAssetCount, resp.ExcludedLiabilityCount)
	}
	if resp.FxAsOf == nil || *resp.FxAsOf != "2026-09-26" {
		t.Fatalf("fx_as_of = %v", resp.FxAsOf)
	}
}

func TestGetNetWorthSummary_MissingRateStaysExcluded(t *testing.T) {
	idr, _ := domainFinance.NewCurrency(domainFinance.NewCurrencyID(1), nil, "IDR", "Rupiah", "Rp", true)
	usd, _ := domainFinance.NewCurrency(domainFinance.NewCurrencyID(2), nil, "USD", "Dollar", "$", true)
	currencyRepo := &netWorthCurrencyRepo{currencies: []*domainFinance.Currency{idr, usd}}
	walletRepo := &walletSummaryWalletRepo{
		wallets: []*domainFinance.Wallet{
			mustReconstituteWallet(t, 1, 1),
			mustReconstituteWallet(t, 2, 2),
		},
		balances: map[int]float64{1: 1000, 2: 10},
	}
	currencyService := domainFinance.NewCurrencyService(currencyRepo)
	walletService := domainFinance.NewWalletService(walletRepo, currencyRepo)
	asOf := time.Date(2026, 9, 26, 0, 0, 0, 0, time.UTC)
	fxRepo := &memoryFxRateRepo{quotes: []domainFinance.FxQuote{
		{Code: "EUR", UnitsPerEUR: 1, AsOf: asOf},
		{Code: "IDR", UnitsPerEUR: 17800, AsOf: asOf},
	}}
	uc := NewGetNetWorthSummaryUseCase(
		NewGetWalletSummaryUseCase(walletService, currencyService),
		walletService,
		domainFinance.NewAssetService(&netWorthAssetRepo{}),
		nil,
		domainFinance.NewLiabilityService(&netWorthLiabilityRepo{}, &netWorthPaymentRepo{}),
		currencyService,
		fxRepo,
	)

	resp, err := uc.Execute(context.Background(), 1)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if resp.LiquidNetWorth != 1000 {
		t.Fatalf("liquid = %v, want 1000", resp.LiquidNetWorth)
	}
	if resp.ConvertedWalletCount != 0 || resp.ExcludedWalletCount != 1 {
		t.Fatalf("wallets converted=%d excluded=%d", resp.ConvertedWalletCount, resp.ExcludedWalletCount)
	}
}
