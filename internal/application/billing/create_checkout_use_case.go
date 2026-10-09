package billing

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainBilling "panda-pocket/internal/domain/billing"
	"panda-pocket/internal/infrastructure/doit"
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
	payments  PaymentCreator
	subs      domainBilling.SubscriptionRepository
	pending   domainBilling.PendingPaymentRepository
	publicIDs domainBilling.PublicIDLookup
}

func NewCreateCheckoutUseCase(payments PaymentCreator, subs domainBilling.SubscriptionRepository, pending domainBilling.PendingPaymentRepository, publicIDs domainBilling.PublicIDLookup) *CreateCheckoutUseCase {
	return &CreateCheckoutUseCase{payments: payments, subs: subs, pending: pending, publicIDs: publicIDs}
}

func (uc *CreateCheckoutUseCase) ExternalID(ctx context.Context, userID int) (string, error) {
	return lookupPublicID(uc.publicIDs, ctx, userID)
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

	publicID, err := lookupPublicID(uc.publicIDs, ctx, userID)
	if err != nil {
		return nil, err
	}
	reference := fmt.Sprintf("%s:%s", domainBilling.CustomerRef(publicID), interval)
	idempotencyKey := fmt.Sprintf("checkout:%s:%s:%s", publicID, interval, time.Now().UTC().Format("2006-01-02"))

	paymentReq := doit.CreatePaymentRequest{
		Amount:    amount,
		Rail:      "any",
		Reference: reference,
		Metadata: map[string]interface{}{
			"public_id": publicID,
			"interval":  string(interval),
		},
	}
	if returnURL := uc.payments.ReturnURL(); returnURL != "" {
		paymentReq.ReturnURL = returnURL
	}

	payment, err := uc.payments.CreatePayment(ctx, idempotencyKey, paymentReq)
	if err != nil {
		return nil, err
	}
	if payment.ID == "" {
		return nil, fmt.Errorf("payment id missing")
	}
	if err := uc.pending.Save(ctx, domainBilling.PendingPayment{
		PaymentID: payment.ID,
		UserID:    userID,
		Kind:      domainBilling.KindPro,
		Interval:  string(interval),
		Amount:    recordedAmount(payment.Amount, amount),
	}); err != nil {
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
	interval := domainBilling.BillingInterval(raw)
	amount, ok := domainBilling.PriceForInterval(interval)
	if !ok {
		return "", 0, fmt.Errorf("interval must be monthly, semiannual, or yearly")
	}
	return interval, amount, nil
}

func recordedAmount(responseAmount, requestedAmount int) int {
	if responseAmount > 0 {
		return responseAmount
	}
	return requestedAmount
}
