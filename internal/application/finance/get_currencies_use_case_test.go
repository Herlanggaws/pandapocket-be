package finance

import (
	"context"
	"testing"

	domainFinance "panda-pocket/internal/domain/finance"
)

type catalogCurrencyRepo struct {
	system []*domainFinance.Currency
	byUser map[int][]*domainFinance.Currency
}

func (r *catalogCurrencyRepo) Save(ctx context.Context, currency *domainFinance.Currency) error {
	return nil
}
func (r *catalogCurrencyRepo) FindByID(ctx context.Context, id domainFinance.CurrencyID) (*domainFinance.Currency, error) {
	return nil, nil
}
func (r *catalogCurrencyRepo) FindByUserID(ctx context.Context, userID domainFinance.UserID) ([]*domainFinance.Currency, error) {
	return r.byUser[userID.Value()], nil
}
func (r *catalogCurrencyRepo) FindDefaultCurrencies(ctx context.Context) ([]*domainFinance.Currency, error) {
	return r.system, nil
}
func (r *catalogCurrencyRepo) Delete(ctx context.Context, id domainFinance.CurrencyID) error {
	return nil
}
func (r *catalogCurrencyRepo) ExistsByID(ctx context.Context, id domainFinance.CurrencyID) (bool, error) {
	return false, nil
}
func (r *catalogCurrencyRepo) ExistsByCodeAndUserID(ctx context.Context, code string, userID domainFinance.UserID) (bool, error) {
	return false, nil
}
func (r *catalogCurrencyRepo) SetUserDefaultCurrency(ctx context.Context, userID domainFinance.UserID, currencyID domainFinance.CurrencyID) error {
	return nil
}
func (r *catalogCurrencyRepo) GetUserDefaultCurrency(ctx context.Context, userID domainFinance.UserID) (*domainFinance.Currency, error) {
	return nil, nil
}

func mustCurrency(t *testing.T, id int, code, name, symbol string, isDefault bool, userID *int) *domainFinance.Currency {
	t.Helper()
	var uid *domainFinance.UserID
	if userID != nil {
		v := domainFinance.NewUserID(*userID)
		uid = &v
	}
	c, err := domainFinance.NewCurrency(domainFinance.NewCurrencyID(id), uid, code, name, symbol, isDefault)
	if err != nil {
		t.Fatalf("NewCurrency: %v", err)
	}
	return c
}

func TestGetCurrenciesCatalogExcludesUserScopedRows(t *testing.T) {
	owner := 42
	repo := &catalogCurrencyRepo{
		system: []*domainFinance.Currency{
			mustCurrency(t, 21, "IDR", "Indonesian Rupiah", "Rp", true, nil),
			mustCurrency(t, 1, "USD", "US Dollar", "$", true, nil),
		},
		byUser: map[int][]*domainFinance.Currency{
			owner: {
				mustCurrency(t, 21, "IDR", "Indonesian Rupiah", "Rp", true, nil),
				mustCurrency(t, 99, "XYZ", "Custom", "X", false, &owner),
			},
		},
	}
	uc := NewGetCurrenciesUseCase(domainFinance.NewCurrencyService(repo))

	catalog, err := uc.ExecuteCatalog(context.Background())
	if err != nil {
		t.Fatalf("ExecuteCatalog: %v", err)
	}
	if len(catalog.Currencies) != 2 {
		t.Fatalf("catalog len=%d want 2", len(catalog.Currencies))
	}
	for _, c := range catalog.Currencies {
		if c.UserID() != nil {
			t.Fatalf("catalog leaked user-owned currency id=%d", c.ID().Value())
		}
	}

	authed, err := uc.Execute(context.Background(), domainFinance.NewUserID(owner))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if len(authed.Currencies) != 2 {
		t.Fatalf("authed len=%d want 2 (system+custom)", len(authed.Currencies))
	}
	foundCustom := false
	for _, c := range authed.Currencies {
		if c.Code() == "XYZ" {
			foundCustom = true
			if c.UserID() == nil || c.UserID().Value() != owner {
				t.Fatalf("custom currency not scoped to owner")
			}
		}
	}
	if !foundCustom {
		t.Fatal("expected user custom currency in authenticated list")
	}
}
