package billing

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainBilling "panda-pocket/internal/domain/billing"
	"panda-pocket/internal/infrastructure/doit"
)

const (
	amountMonthly    = 19000
	amountSemiannual = 99000
	amountYearly     = 149000
)

// ErrShorterIntervalBlocked means a shorter prepaid interval was requested while a longer paid period is still open.
var ErrShorterIntervalBlocked = errors.New("shorter billing interval applies at the end of the current period")

// PaymentCreator creates a Doit one-shot payment.
type PaymentCreator interface {
	Configured() bool
	ReturnURL() string
	CreatePayment(ctx context.Context, idempotencyKey string, req doit.CreatePaymentRequest) (*doit.CreatePaymentResponse, error)
}

type CreateCheckoutRequest struct {
	Interval string `json:"interval"`
}

type CreateCheckoutResponse struct {
	HostedURL string `json:"hosted_url"`
	PaymentID string `json:"payment_id"`
	Reference string `json:"reference"`
}

type CreateCheckoutUseCase struct {
	payments PaymentCreator
	subs     domainBilling.SubscriptionRepository
}

func NewCreateCheckoutUseCase(payments PaymentCreator, subs domainBilling.SubscriptionRepository) *CreateCheckoutUseCase {
	return &CreateCheckoutUseCase{payments: payments, subs: subs}
}

func (uc *CreateCheckoutUseCase) Execute(ctx context.Context, userID int, req CreateCheckoutRequest) (*CreateCheckoutResponse, error) {
	if !uc.payments.Configured() {
		return nil, fmt.Errorf("billing checkout is not configured")
	}

	interval, amount, err := resolveCheckoutInterval(req.Interval)
	if err != nil {
		return nil, err
	}

	if err := uc.rejectShorterInterval(ctx, userID, interval); err != nil {
		return nil, err
	}

	reference := fmt.Sprintf("%s:%s", domainBilling.CustomerRef(userID), interval)
	idempotencyKey := fmt.Sprintf("checkout:%d:%s:%s", userID, interval, time.Now().UTC().Format("2006-01-02"))

	paymentReq := doit.CreatePaymentRequest{
		Amount:    amount,
		Rail:      "any",
		Reference: reference,
		Metadata: map[string]interface{}{
			"user_id":  userID,
			"interval": string(interval),
		},
	}
	if returnURL := uc.payments.ReturnURL(); returnURL != "" {
		paymentReq.ReturnURL = returnURL
	}

	payment, err := uc.payments.CreatePayment(ctx, idempotencyKey, paymentReq)
	if err != nil {
		return nil, err
	}

	return &CreateCheckoutResponse{
		HostedURL: payment.HostedURL,
		PaymentID: payment.ID,
		Reference: payment.Reference,
	}, nil
}

func (uc *CreateCheckoutUseCase) rejectShorterInterval(ctx context.Context, userID int, requested domainBilling.BillingInterval) error {
	if uc.subs == nil {
		return nil
	}
	sub, err := uc.subs.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, domainBilling.ErrNotFound) {
			return nil
		}
		return err
	}
	if blocksShorterInterval(sub, requested, time.Now().UTC()) {
		return ErrShorterIntervalBlocked
	}
	return nil
}

func blocksShorterInterval(sub *domainBilling.Subscription, requested domainBilling.BillingInterval, now time.Time) bool {
	current := sub.BillingInterval()
	periodEnd := sub.CurrentPeriodEnd()
	if current == nil || periodEnd == nil || !periodEnd.After(now) {
		return false
	}
	return requested.ShorterThan(*current)
}

func resolveCheckoutInterval(raw string) (domainBilling.BillingInterval, int, error) {
	switch domainBilling.BillingInterval(raw) {
	case domainBilling.IntervalMonthly:
		return domainBilling.IntervalMonthly, amountMonthly, nil
	case domainBilling.IntervalSemiannual:
		return domainBilling.IntervalSemiannual, amountSemiannual, nil
	case domainBilling.IntervalYearly:
		return domainBilling.IntervalYearly, amountYearly, nil
	default:
		return "", 0, fmt.Errorf("interval must be monthly, semiannual, or yearly")
	}
}
