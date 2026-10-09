package billing

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	domainBilling "panda-pocket/internal/domain/billing"
	"panda-pocket/internal/infrastructure/doit"
)

const checkoutPublicID = "11111111-1111-4111-8111-111111111111"

type stubPublicIDs map[int]string

func (s stubPublicIDs) PublicID(_ context.Context, userID int) (string, error) {
	publicID, ok := s[userID]
	if !ok || publicID == "" {
		return "", errors.New("public id not found")
	}
	return publicID, nil
}

func checkoutIDs() stubPublicIDs {
	return stubPublicIDs{1: checkoutPublicID, 9: checkoutPublicID}
}

type stubPayments struct {
	calls   int
	amount  int
	last    doit.CreatePaymentRequest
	lastKey string
}

func (s *stubPayments) Configured() bool  { return true }
func (s *stubPayments) ReturnURL() string { return "" }
func (s *stubPayments) CreatePayment(_ context.Context, idempotencyKey string, req doit.CreatePaymentRequest) (*doit.CreatePaymentResponse, error) {
	s.calls++
	s.amount = req.Amount
	s.last = req
	s.lastKey = idempotencyKey
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
		domainBilling.CustomerRef("legacy"),
		false,
		now,
		now,
	)
}

func TestCheckoutAmountsKeepExistingPrices(t *testing.T) {
	payments := &stubPayments{}
	uc := NewCreateCheckoutUseCase(payments, &checkoutSubs{}, &memoryPending{}, checkoutIDs())

	monthly, err := uc.Execute(context.Background(), 1, CreateCheckoutRequest{Interval: "monthly"})
	if err != nil {
		t.Fatal(err)
	}
	if payments.amount != 19000 || monthly.Reference != "user:"+checkoutPublicID+":monthly" {
		t.Fatalf("monthly checkout amount=%d ref=%s", payments.amount, monthly.Reference)
	}
	if payments.last.Metadata["public_id"] != checkoutPublicID || payments.last.Metadata["user_id"] != nil {
		t.Fatalf("metadata=%v", payments.last.Metadata)
	}
	if !strings.HasPrefix(payments.lastKey, "checkout:"+checkoutPublicID+":") {
		t.Fatalf("idempotency=%s", payments.lastKey)
	}

	semiannual, err := uc.Execute(context.Background(), 1, CreateCheckoutRequest{Interval: "semiannual"})
	if err != nil {
		t.Fatal(err)
	}
	if payments.amount != 99000 || semiannual.Reference != "user:"+checkoutPublicID+":semiannual" {
		t.Fatalf("semiannual checkout amount=%d ref=%s", payments.amount, semiannual.Reference)
	}

	yearly, err := uc.Execute(context.Background(), 1, CreateCheckoutRequest{Interval: "yearly"})
	if err != nil {
		t.Fatal(err)
	}
	if payments.amount != 149000 || yearly.Reference != "user:"+checkoutPublicID+":yearly" {
		t.Fatalf("yearly checkout amount=%d ref=%s", payments.amount, yearly.Reference)
	}
}

func TestCheckoutRejectsShorterIntervalWithoutSaving(t *testing.T) {
	periodEnd := time.Now().UTC().AddDate(0, 0, 40)
	originalEnd := periodEnd
	subs := &checkoutSubs{sub: paidSubscription(9, domainBilling.IntervalYearly, periodEnd)}
	payments := &stubPayments{}
	uc := NewCreateCheckoutUseCase(payments, subs, &memoryPending{}, checkoutIDs())

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
	if payments.amount != 149000 || longer.Reference != "user:"+checkoutPublicID+":yearly" {
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
	uc := NewCreateCheckoutUseCase(payments, subs, &memoryPending{}, checkoutIDs())

	res, err := uc.Execute(context.Background(), 9, CreateCheckoutRequest{Interval: "monthly"})
	if err != nil {
		t.Fatal(err)
	}
	if payments.amount != 19000 || res.Reference != "user:"+checkoutPublicID+":monthly" {
		t.Fatalf("amount=%d ref=%s", payments.amount, res.Reference)
	}
}
