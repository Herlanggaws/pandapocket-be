package ai

import (
	"context"
	"errors"
	"testing"

	domainAI "panda-pocket/internal/domain/ai"
	domainBilling "panda-pocket/internal/domain/billing"
	"panda-pocket/internal/infrastructure/paas"
)

type stubStreamer struct {
	full string
	err  error
}

func (s *stubStreamer) Configured() bool { return true }

func (s *stubStreamer) StreamChat(_ context.Context, _ []paas.Message, onDelta func(string) error) (string, int, int, error) {
	if s.err != nil {
		return "", 0, 0, s.err
	}
	if onDelta != nil && s.full != "" {
		_ = onDelta(s.full)
	}
	return s.full, 1, 1, nil
}

type memCredits struct {
	balance *domainAI.CreditBalance
}

func (m *memCredits) FindByUserID(_ context.Context, _ int) (*domainAI.CreditBalance, error) {
	if m.balance == nil {
		return nil, domainAI.ErrBalanceNotFound
	}
	cp := *m.balance
	return &cp, nil
}

func (m *memCredits) Save(_ context.Context, balance *domainAI.CreditBalance) error {
	cp := *balance
	m.balance = &cp
	return nil
}

func (m *memCredits) PurchaseExists(_ context.Context, _ string) (bool, error) {
	return false, nil
}

func (m *memCredits) RecordPurchase(_ context.Context, _ int, _, _ string, _ int) error {
	return nil
}

type memThreads struct {
	id   int
	msgs []domainAI.ThreadMessage
}

func (m *memThreads) GetOrCreateThreadID(_ context.Context, _ int) (int, error) {
	if m.id == 0 {
		m.id = 1
	}
	return m.id, nil
}

func (m *memThreads) ListMessages(_ context.Context, _ int) ([]domainAI.ThreadMessage, error) {
	out := make([]domainAI.ThreadMessage, len(m.msgs))
	copy(out, m.msgs)
	return out, nil
}

func (m *memThreads) AppendMessage(_ context.Context, _ int, role, content string, _, _ int) error {
	m.msgs = append(m.msgs, domainAI.ThreadMessage{Role: role, Content: content})
	return nil
}

func (m *memThreads) ClearMessages(_ context.Context, _ int) error {
	m.msgs = nil
	return nil
}

func (m *memThreads) TrimOldest(_ context.Context, _, _ int) error { return nil }

type alwaysPro struct{}

func (alwaysPro) IsPro(_ context.Context, _ int) (bool, error) { return true, nil }

type emptySubs struct{}

func (emptySubs) Save(_ context.Context, _ *domainBilling.Subscription) error { return nil }
func (emptySubs) FindByUserID(_ context.Context, _ int) (*domainBilling.Subscription, error) {
	return nil, domainBilling.ErrNotFound
}
func (emptySubs) ListAll(_ context.Context) ([]*domainBilling.Subscription, error) {
	return nil, nil
}

func newTestChat(streamer *stubStreamer) (*AdvisorChatUseCase, *memCredits) {
	bal := domainAI.NewCreditBalance(1, false)
	creditsRepo := &memCredits{balance: bal}
	svc := NewCreditService(creditsRepo, emptySubs{})
	uc := NewAdvisorChatUseCase(svc, &memThreads{}, alwaysPro{}, streamer, nil, nil)
	return uc, creditsRepo
}

func TestAdvisorChatSpendsOnlyAfterSuccess(t *testing.T) {
	uc, repo := newTestChat(&stubStreamer{full: "Halo, ini jawaban."})
	view, err := uc.Execute(context.Background(), 1, "Apa kabar?", nil)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if view.Available != domainAI.IncludedGrant()-1 {
		t.Fatalf("available=%d want %d", view.Available, domainAI.IncludedGrant()-1)
	}
	if repo.balance.Available() != domainAI.IncludedGrant()-1 {
		t.Fatalf("persisted available=%d", repo.balance.Available())
	}
}

func TestAdvisorChatNoSpendOnUpstreamFailure(t *testing.T) {
	uc, repo := newTestChat(&stubStreamer{err: errors.New("provider down")})
	view, err := uc.Execute(context.Background(), 1, "Apa kabar?", nil)
	if err == nil || !errors.Is(err, domainAI.ErrUpstream) {
		t.Fatalf("want upstream err, got %v", err)
	}
	if view.Available != domainAI.IncludedGrant() {
		t.Fatalf("available=%d — must not spend on failure", view.Available)
	}
	if repo.balance.Available() != domainAI.IncludedGrant() {
		t.Fatalf("persisted available=%d", repo.balance.Available())
	}
}

func TestAdvisorChatNoSpendOnEmptyReply(t *testing.T) {
	uc, repo := newTestChat(&stubStreamer{full: "   "})
	view, err := uc.Execute(context.Background(), 1, "Apa kabar?", nil)
	if err == nil || !errors.Is(err, domainAI.ErrUpstream) {
		t.Fatalf("want upstream err, got %v", err)
	}
	if view.Available != domainAI.IncludedGrant() {
		t.Fatalf("available=%d — must not spend on empty", view.Available)
	}
	if repo.balance.Available() != domainAI.IncludedGrant() {
		t.Fatalf("persisted available=%d", repo.balance.Available())
	}
}
