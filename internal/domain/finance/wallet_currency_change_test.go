package finance

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

type memCurrencyRepo struct {
	byID map[int]*Currency
}

func (r *memCurrencyRepo) Save(ctx context.Context, currency *Currency) error { return nil }
func (r *memCurrencyRepo) FindByID(ctx context.Context, id CurrencyID) (*Currency, error) {
	c, ok := r.byID[id.Value()]
	if !ok {
		return nil, errors.New("currency not found")
	}
	return c, nil
}
func (r *memCurrencyRepo) FindByUserID(ctx context.Context, userID UserID) ([]*Currency, error) {
	return nil, nil
}
func (r *memCurrencyRepo) FindDefaultCurrencies(ctx context.Context) ([]*Currency, error) {
	return nil, nil
}
func (r *memCurrencyRepo) Delete(ctx context.Context, id CurrencyID) error { return nil }
func (r *memCurrencyRepo) ExistsByID(ctx context.Context, id CurrencyID) (bool, error) {
	_, ok := r.byID[id.Value()]
	return ok, nil
}
func (r *memCurrencyRepo) ExistsByCodeAndUserID(ctx context.Context, code string, userID UserID) (bool, error) {
	return false, nil
}
func (r *memCurrencyRepo) SetUserDefaultCurrency(ctx context.Context, userID UserID, currencyID CurrencyID) error {
	return nil
}
func (r *memCurrencyRepo) GetUserDefaultCurrency(ctx context.Context, userID UserID) (*Currency, error) {
	return r.FindByID(ctx, NewCurrencyID(1))
}

type memWalletRepo struct {
	wallets           map[int]*Wallet
	hasTx             map[int]bool
	savedCurrency     map[int]int
	alignedPending    map[int]int
	alignedRecurring  map[int]int
}

func newMemWalletRepo(wallets ...*Wallet) *memWalletRepo {
	r := &memWalletRepo{
		wallets:          map[int]*Wallet{},
		hasTx:            map[int]bool{},
		savedCurrency:    map[int]int{},
		alignedPending:   map[int]int{},
		alignedRecurring: map[int]int{},
	}
	for _, w := range wallets {
		r.wallets[w.ID().Value()] = w
	}
	return r
}

func (r *memWalletRepo) Save(ctx context.Context, wallet *Wallet) error {
	r.wallets[wallet.ID().Value()] = wallet
	r.savedCurrency[wallet.ID().Value()] = wallet.CurrencyID().Value()
	return nil
}
func (r *memWalletRepo) FindByID(ctx context.Context, id WalletID) (*Wallet, error) {
	w, ok := r.wallets[id.Value()]
	if !ok {
		return nil, errors.New("wallet not found")
	}
	return w, nil
}
func (r *memWalletRepo) FindByUserID(ctx context.Context, userID UserID, includeArchived bool) ([]*Wallet, error) {
	out := make([]*Wallet, 0)
	for _, w := range r.wallets {
		if w.UserID().Value() == userID.Value() {
			out = append(out, w)
		}
	}
	return out, nil
}
func (r *memWalletRepo) FindDefaultByUserID(ctx context.Context, userID UserID) (*Wallet, error) {
	for _, w := range r.wallets {
		if w.UserID().Value() == userID.Value() && w.IsDefault() {
			return w, nil
		}
	}
	return nil, errors.New("default wallet not found")
}
func (r *memWalletRepo) CountActiveByUserID(ctx context.Context, userID UserID) (int64, error) {
	return int64(len(r.wallets)), nil
}
func (r *memWalletRepo) ClearDefaultForUser(ctx context.Context, userID UserID) error { return nil }
func (r *memWalletRepo) HasTransactions(ctx context.Context, id WalletID) (bool, error) {
	return r.hasTx[id.Value()], nil
}
func (r *memWalletRepo) GetBalanceBreakdown(ctx context.Context, id WalletID) (WalletBalanceBreakdown, error) {
	return WalletBalanceBreakdown{}, nil
}
func (r *memWalletRepo) AlignPendingCurrency(ctx context.Context, id WalletID, currencyID CurrencyID) error {
	r.alignedPending[id.Value()] = currencyID.Value()
	return nil
}
func (r *memWalletRepo) AlignRecurringCurrency(ctx context.Context, id WalletID, currencyID CurrencyID) error {
	r.alignedRecurring[id.Value()] = currencyID.Value()
	return nil
}

type memLinkedGoals struct {
	linked map[int]bool
}

func (m *memLinkedGoals) HasLinkedGoals(ctx context.Context, walletID WalletID) (bool, error) {
	return m.linked[walletID.Value()], nil
}

func mustSystemCurrency(t *testing.T, id int, code string) *Currency {
	t.Helper()
	c, err := NewCurrency(NewCurrencyID(id), nil, code, code, code, true)
	if err != nil {
		t.Fatalf("currency: %v", err)
	}
	return c
}

func emptyWallet(userID, walletID, currencyID int, opening float64, isDefault bool) *Wallet {
	return ReconstituteWallet(
		NewWalletID(walletID),
		NewUserID(userID),
		"Cash",
		WalletTypeCash,
		NewCurrencyID(currencyID),
		opening,
		isDefault,
		false,
		time.Now(),
	)
}

func TestUpdateWallet_ChangeCurrencyWhenEmpty(t *testing.T) {
	wallet := emptyWallet(1, 10, 1, 0, true)
	walletRepo := newMemWalletRepo(wallet)
	currencyRepo := &memCurrencyRepo{byID: map[int]*Currency{
		1: mustSystemCurrency(t, 1, "USD"),
		2: mustSystemCurrency(t, 2, "IDR"),
	}}
	svc := NewWalletService(walletRepo, currencyRepo)
	svc.SetLinkedGoalsChecker(&memLinkedGoals{linked: map[int]bool{}})

	currencyID := 2
	updated, err := svc.UpdateWallet(context.Background(), NewUserID(1), NewWalletID(10), nil, nil, nil, &currencyID)
	if err != nil {
		t.Fatalf("UpdateWallet: %v", err)
	}
	if updated.CurrencyID().Value() != 2 {
		t.Fatalf("want currency 2, got %d", updated.CurrencyID().Value())
	}
	if walletRepo.savedCurrency[10] != 2 {
		t.Fatalf("Save did not persist currency_id")
	}
	if walletRepo.alignedPending[10] != 2 || walletRepo.alignedRecurring[10] != 2 {
		t.Fatalf("expected pending+recurring align to 2")
	}
}

func TestUpdateWallet_RejectWhenHasTransactions(t *testing.T) {
	wallet := emptyWallet(1, 10, 1, 0, true)
	walletRepo := newMemWalletRepo(wallet)
	walletRepo.hasTx[10] = true
	currencyRepo := &memCurrencyRepo{byID: map[int]*Currency{
		1: mustSystemCurrency(t, 1, "USD"),
		2: mustSystemCurrency(t, 2, "IDR"),
	}}
	svc := NewWalletService(walletRepo, currencyRepo)

	currencyID := 2
	_, err := svc.UpdateWallet(context.Background(), NewUserID(1), NewWalletID(10), nil, nil, nil, &currencyID)
	if err == nil || !strings.Contains(err.Error(), "wallet has transactions") {
		t.Fatalf("want transactions error, got %v", err)
	}
	if wallet.CurrencyID().Value() != 1 {
		t.Fatalf("currency should stay 1")
	}
}

func TestUpdateWallet_RejectWhenOpeningBalance(t *testing.T) {
	wallet := emptyWallet(1, 10, 1, 1000, true)
	walletRepo := newMemWalletRepo(wallet)
	currencyRepo := &memCurrencyRepo{byID: map[int]*Currency{
		1: mustSystemCurrency(t, 1, "USD"),
		2: mustSystemCurrency(t, 2, "IDR"),
	}}
	svc := NewWalletService(walletRepo, currencyRepo)

	currencyID := 2
	_, err := svc.UpdateWallet(context.Background(), NewUserID(1), NewWalletID(10), nil, nil, nil, &currencyID)
	if err == nil || !strings.Contains(err.Error(), "opening balance") {
		t.Fatalf("want opening balance error, got %v", err)
	}
}

func TestUpdateWallet_RejectWhenLinkedGoals(t *testing.T) {
	wallet := emptyWallet(1, 10, 1, 0, true)
	walletRepo := newMemWalletRepo(wallet)
	currencyRepo := &memCurrencyRepo{byID: map[int]*Currency{
		1: mustSystemCurrency(t, 1, "USD"),
		2: mustSystemCurrency(t, 2, "IDR"),
	}}
	svc := NewWalletService(walletRepo, currencyRepo)
	svc.SetLinkedGoalsChecker(&memLinkedGoals{linked: map[int]bool{10: true}})

	currencyID := 2
	_, err := svc.UpdateWallet(context.Background(), NewUserID(1), NewWalletID(10), nil, nil, nil, &currencyID)
	if err == nil || !strings.Contains(err.Error(), "unlink goals") {
		t.Fatalf("want unlink goals error, got %v", err)
	}
}

func TestSyncDefaultWalletCurrency_SkipsWhenNotEmpty(t *testing.T) {
	wallet := emptyWallet(1, 10, 1, 0, true)
	walletRepo := newMemWalletRepo(wallet)
	walletRepo.hasTx[10] = true
	currencyRepo := &memCurrencyRepo{byID: map[int]*Currency{
		1: mustSystemCurrency(t, 1, "USD"),
		2: mustSystemCurrency(t, 2, "IDR"),
	}}
	svc := NewWalletService(walletRepo, currencyRepo)

	if err := svc.SyncDefaultWalletCurrency(context.Background(), NewUserID(1), NewCurrencyID(2)); err != nil {
		t.Fatalf("SyncDefault should no-op, got %v", err)
	}
	if wallet.CurrencyID().Value() != 1 {
		t.Fatalf("currency should stay 1 on skip")
	}
	if _, ok := walletRepo.savedCurrency[10]; ok {
		t.Fatalf("Save should not run when sync skips")
	}
}

func TestSyncDefaultWalletCurrency_UpdatesWhenEmpty(t *testing.T) {
	wallet := emptyWallet(1, 10, 1, 0, true)
	walletRepo := newMemWalletRepo(wallet)
	currencyRepo := &memCurrencyRepo{byID: map[int]*Currency{
		1: mustSystemCurrency(t, 1, "USD"),
		2: mustSystemCurrency(t, 2, "IDR"),
	}}
	svc := NewWalletService(walletRepo, currencyRepo)
	svc.SetLinkedGoalsChecker(&memLinkedGoals{linked: map[int]bool{}})

	if err := svc.SyncDefaultWalletCurrency(context.Background(), NewUserID(1), NewCurrencyID(2)); err != nil {
		t.Fatalf("SyncDefault: %v", err)
	}
	if wallet.CurrencyID().Value() != 2 {
		t.Fatalf("want currency 2, got %d", wallet.CurrencyID().Value())
	}
	if walletRepo.savedCurrency[10] != 2 {
		t.Fatalf("Save did not persist currency_id")
	}
}
