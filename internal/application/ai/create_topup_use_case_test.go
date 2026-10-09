package ai

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	domainBilling "panda-pocket/internal/domain/billing"
	"panda-pocket/internal/domain/entitlement"
	"panda-pocket/internal/infrastructure/doit"
)

const topupPublicID = "33333333-3333-4333-8333-333333333333"

type stubPublicIDs map[int]string

func (s stubPublicIDs) PublicID(_ context.Context, userID int) (string, error) {
	publicID, ok := s[userID]
	if !ok || publicID == "" {
		return "", errors.New("public id not found")
	}
	return publicID, nil
}

func topupIDs() stubPublicIDs {
	return stubPublicIDs{32: topupPublicID, 7: topupPublicID}
}

type stubPayments struct {
	configured bool
	calls      []string
	last       doit.CreatePaymentRequest
	responses  []*doit.CreatePaymentResponse
}

func (s *stubPayments) Configured() bool { return s.configured }

func (s *stubPayments) CreatePayment(_ context.Context, idempotencyKey string, req doit.CreatePaymentRequest) (*doit.CreatePaymentResponse, error) {
	s.calls = append(s.calls, idempotencyKey)
	s.last = req
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
	uc := NewCreateTopupUseCase(payments, entitlement.StaticChecker{Pro: true}, &memoryPending{}, topupIDs())
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
	if payments.last.Metadata["public_id"] != topupPublicID || payments.last.Metadata["user_id"] != nil {
		t.Fatalf("metadata=%v", payments.last.Metadata)
	}
	if !strings.HasPrefix(payments.last.Reference, "user:"+topupPublicID+":ai:") {
		t.Fatalf("reference=%s", payments.last.Reference)
	}
	if strings.Contains(payments.calls[0], ":32:") {
		t.Fatalf("idempotency still contains sequential user id: %s", payments.calls[0])
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
	pending := &memoryPending{}
	uc := NewCreateTopupUseCase(payments, entitlement.StaticChecker{Pro: true}, pending, topupIDs())
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
	if len(pending.saved) != 1 || pending.saved[0].PaymentID != "pay_new" || pending.saved[0].Amount != 24900 {
		t.Fatalf("recorded=%+v", pending.saved)
	}
}

type memoryPending struct {
	saved []domainBilling.PendingPayment
}

func (m *memoryPending) Save(_ context.Context, payment domainBilling.PendingPayment) error {
	m.saved = append(m.saved, payment)
	return nil
}

func (m *memoryPending) FindByPaymentID(context.Context, string) (domainBilling.PendingPayment, bool, error) {
	return domainBilling.PendingPayment{}, false, nil
}
