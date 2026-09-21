package billing

import (
	"context"
	"fmt"
	"time"

	domainBilling "panda-pocket/internal/domain/billing"
	"panda-pocket/internal/infrastructure/doit"
)

const (
	amountMonthly = 19000
	amountYearly  = 149000
)

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
}

func NewCreateCheckoutUseCase(payments PaymentCreator) *CreateCheckoutUseCase {
	return &CreateCheckoutUseCase{payments: payments}
}

func (uc *CreateCheckoutUseCase) Execute(ctx context.Context, userID int, req CreateCheckoutRequest) (*CreateCheckoutResponse, error) {
	if !uc.payments.Configured() {
		return nil, fmt.Errorf("billing checkout is not configured")
	}

	interval, amount, err := resolveCheckoutInterval(req.Interval)
	if err != nil {
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

func resolveCheckoutInterval(raw string) (domainBilling.BillingInterval, int, error) {
	switch domainBilling.BillingInterval(raw) {
	case domainBilling.IntervalMonthly:
		return domainBilling.IntervalMonthly, amountMonthly, nil
	case domainBilling.IntervalYearly:
		return domainBilling.IntervalYearly, amountYearly, nil
	default:
		return "", 0, fmt.Errorf("interval must be monthly or yearly")
	}
}
