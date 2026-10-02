package billing

import (
	"context"
	"errors"
	"testing"
	"time"

	domainBilling "panda-pocket/internal/domain/billing"
	"panda-pocket/internal/infrastructure/doit"
)

type stubPayments struct {
	calls  int
	amount int
}

func (s *stubPayments) Configured() bool  { return true }
func (s *stubPayments) ReturnURL() string { return "" }
func (s *stubPayments) CreatePayment(_ context.Context, _ string, req doit.CreatePaymentRequest) (*doit.CreatePaymentResponse, error) {
	s.calls++
	s.amount = req.Amount
	return &doit.CreatePaymentResponse{
		ID:        "pay_test",
		HostedURL: "https://pay.example/hosted",
		Reference: req.Reference,
		Amount:    req.Amount,
	}, nil
}

type checkoutSubs struct {
	sub   *domainBilling.Subscription
	saves int
}

func (c *checkoutSubs) Save(context.Context, *domainBilling.Subscription) error {
	c.saves++
	return nil
}

func (c *checkoutSubs) FindByUserID(context.Context, int) (*domainBilling.Subscription, error) {
	if c.sub == nil {
		return nil, domainBilling.ErrNotFound
	}
	return c.sub, nil
}

func (c *checkoutSubs) ListAll(context.Context) ([]*domainBilling.Subscription, error) {
	return nil, nil
}

func paidSubscription(userID int, interval domainBilling.BillingInterval, periodEnd time.Time) *domainBilling.Subscription {
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	return domainBilling.ReconstituteSubscription(
		domainBilling.NewSubscriptionID(1),
		userID,
		domainBilling.PlanPro,
		&interval,
		domainBilling.StatusActive,
		nil,
		&periodEnd,
		nil,
		nil,
		domainBilling.CustomerRef(userID),
		false,
		now,
		now,
	)
}

func TestCheckoutAmountsKeepExistingPrices(t *testing.T) {
	payments := &stubPayments{}
	uc := NewCreateCheckoutUseCase(payments, &checkoutSubs{})

	monthly, err := uc.Execute(context.Background(), 1, CreateCheckoutRequest{Interval: "monthly"})
	if err != nil {
		t.Fatal(err)
	}
	if payments.amount != 19000 || monthly.Reference != "user:1:monthly" {
		t.Fatalf("monthly checkout amount=%d ref=%s", payments.amount, monthly.Reference)
	}

	semiannual, err := uc.Execute(context.Background(), 1, CreateCheckoutRequest{Interval: "semiannual"})
	if err != nil {
		t.Fatal(err)
	}
	if payments.amount != 99000 || semiannual.Reference != "user:1:semiannual" {
		t.Fatalf("semiannual checkout amount=%d ref=%s", payments.amount, semiannual.Reference)
	}

	yearly, err := uc.Execute(context.Background(), 1, CreateCheckoutRequest{Interval: "yearly"})
	if err != nil {
		t.Fatal(err)
	}
	if payments.amount != 149000 || yearly.Reference != "user:1:yearly" {
		t.Fatalf("yearly checkout amount=%d ref=%s", payments.amount, yearly.Reference)
	}
}

func TestCheckoutRejectsShorterIntervalWithoutSaving(t *testing.T) {
	periodEnd := time.Now().UTC().AddDate(0, 0, 40)
	originalEnd := periodEnd
	subs := &checkoutSubs{sub: paidSubscription(9, domainBilling.IntervalYearly, periodEnd)}
	payments := &stubPayments{}
	uc := NewCreateCheckoutUseCase(payments, subs)

	_, err := uc.Execute(context.Background(), 9, CreateCheckoutRequest{Interval: "monthly"})
	if !errors.Is(err, ErrShorterIntervalBlocked) {
		t.Fatalf("monthly from yearly: %v", err)
	}
	_, err = uc.Execute(context.Background(), 9, CreateCheckoutRequest{Interval: "semiannual"})
	if !errors.Is(err, ErrShorterIntervalBlocked) {
		t.Fatalf("semiannual from yearly: %v", err)
	}
	if payments.calls != 0 || subs.saves != 0 {
		t.Fatalf("blocked checkout must not charge or save, calls=%d saves=%d", payments.calls, subs.saves)
	}
	if subs.sub.CurrentPeriodEnd() == nil || !subs.sub.CurrentPeriodEnd().Equal(originalEnd) {
		t.Fatal("existing period end changed")
	}
	if subs.sub.BillingInterval() == nil || *subs.sub.BillingInterval() != domainBilling.IntervalYearly {
		t.Fatal("existing interval changed")
	}

	longer, err := uc.Execute(context.Background(), 9, CreateCheckoutRequest{Interval: "yearly"})
	if err != nil {
		t.Fatal(err)
	}
	if payments.amount != 149000 || longer.Reference != "user:9:yearly" {
		t.Fatalf("same-interval repurchase amount=%d ref=%s", payments.amount, longer.Reference)
	}
	if subs.saves != 0 {
		t.Fatal("checkout must not write the subscription")
	}
}

func TestCheckoutAllowsShorterIntervalAfterPeriodEnds(t *testing.T) {
	periodEnd := time.Now().UTC().AddDate(0, 0, -1)
	subs := &checkoutSubs{sub: paidSubscription(9, domainBilling.IntervalYearly, periodEnd)}
	payments := &stubPayments{}
	uc := NewCreateCheckoutUseCase(payments, subs)

	res, err := uc.Execute(context.Background(), 9, CreateCheckoutRequest{Interval: "monthly"})
	if err != nil {
		t.Fatal(err)
	}
	if payments.amount != 19000 || res.Reference != "user:9:monthly" {
		t.Fatalf("amount=%d ref=%s", payments.amount, res.Reference)
	}
}
