package billing

import "context"

const (
	KindPro       = "pro"
	KindAICredits = "ai_credits"
)

// PendingPayment is a checkout or top-up we created, before the webhook arrives.
type PendingPayment struct {
	PaymentID string
	UserID    int
	Kind      string
	Interval  string
	Pack      string
	Amount    int
}

// PendingPaymentRepository stores payments created by this API.
type PendingPaymentRepository interface {
	Save(ctx context.Context, payment PendingPayment) error
	FindByPaymentID(ctx context.Context, paymentID string) (PendingPayment, bool, error)
}

// PriceForInterval is the rupiah amount charged for a Pro interval.
func PriceForInterval(interval BillingInterval) (int, bool) {
	switch interval {
	case IntervalMonthly:
		return 19000, true
	case IntervalSemiannual:
		return 99000, true
	case IntervalYearly:
		return 149000, true
	default:
		return 0, false
	}
}
