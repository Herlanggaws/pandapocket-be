package ai

import (
	"context"
	"testing"
	"time"

	"panda-pocket/internal/domain/entitlement"
	"panda-pocket/internal/infrastructure/doit"
)

type stubPayments struct {
	configured bool
	calls      []string
	responses  []*doit.CreatePaymentResponse
}

func (s *stubPayments) Configured() bool { return s.configured }

func (s *stubPayments) CreatePayment(_ context.Context, idempotencyKey string, _ doit.CreatePaymentRequest) (*doit.CreatePaymentResponse, error) {
	s.calls = append(s.calls, idempotencyKey)
	idx := len(s.calls) - 1
	if idx < len(s.responses) {
		return s.responses[idx], nil
	}
	return &doit.CreatePaymentResponse{
		ID:        "pay_new",
		Status:    "pending",
		Reference: "ref",
		HostedURL: "https://pay.doit.id/p/pay_new",
	}, nil
}

func TestCreateTopupDistinctIdempotencyKeys(t *testing.T) {
	payments := &stubPayments{configured: true}
	uc := NewCreateTopupUseCase(payments, entitlement.StaticChecker{Pro: true})
	base := time.Date(2026, 9, 24, 8, 0, 0, 0, time.UTC)
	n := 0
	uc.now = func() time.Time {
		n++
		return base.Add(time.Duration(n) * time.Nanosecond)
	}

	ctx := context.Background()
	first, err := uc.Execute(ctx, 32, CreateTopupRequest{Pack: "ai_credits_s"})
	if err != nil {
		t.Fatalf("first: %v", err)
	}
	second, err := uc.Execute(ctx, 32, CreateTopupRequest{Pack: "ai_credits_s"})
	if err != nil {
		t.Fatalf("second: %v", err)
	}
	if len(payments.calls) != 2 {
		t.Fatalf("expected 2 CreatePayment calls, got %d", len(payments.calls))
	}
	if payments.calls[0] == payments.calls[1] {
		t.Fatalf("expected distinct idempotency keys, got %q twice", payments.calls[0])
	}
	if first.PaymentID == "" || second.HostedURL == "" {
		t.Fatalf("expected payment responses")
	}
}

func TestCreateTopupRetriesWhenDoitReturnsPaid(t *testing.T) {
	payments := &stubPayments{
		configured: true,
		responses: []*doit.CreatePaymentResponse{
			{ID: "pay_old", Status: "paid", Reference: "old", HostedURL: "https://pay.doit.id/p/pay_old"},
			{ID: "pay_new", Status: "pending", Reference: "new", HostedURL: "https://pay.doit.id/p/pay_new"},
		},
	}
	uc := NewCreateTopupUseCase(payments, entitlement.StaticChecker{Pro: true})
	uc.now = func() time.Time { return time.Date(2026, 9, 24, 8, 30, 0, 0, time.UTC) }

	resp, err := uc.Execute(context.Background(), 7, CreateTopupRequest{Pack: "ai_credits_m"})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if len(payments.calls) != 2 {
		t.Fatalf("expected retry after paid, got %d calls", len(payments.calls))
	}
	if resp.PaymentID != "pay_new" {
		t.Fatalf("expected unpaid payment, got %s", resp.PaymentID)
	}
	if resp.Credits != 150 {
		t.Fatalf("expected 150 credits, got %d", resp.Credits)
	}
}
