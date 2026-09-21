package billing

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"
	"time"

	domainBilling "panda-pocket/internal/domain/billing"
)

type memoryWebhookEvents struct {
	seen map[string]string
}

func (m *memoryWebhookEvents) TryInsert(_ context.Context, eventID, payload string) (bool, error) {
	if m.seen == nil {
		m.seen = map[string]string{}
	}
	if _, ok := m.seen[eventID]; ok {
		return false, nil
	}
	m.seen[eventID] = payload
	return true, nil
}

type memorySubs struct {
	byUser map[int]*domainBilling.Subscription
}

func (m *memorySubs) Save(_ context.Context, sub *domainBilling.Subscription) error {
	if m.byUser == nil {
		m.byUser = map[int]*domainBilling.Subscription{}
	}
	m.byUser[sub.UserID()] = sub
	return nil
}

func (m *memorySubs) FindByUserID(_ context.Context, userID int) (*domainBilling.Subscription, error) {
	if sub, ok := m.byUser[userID]; ok {
		return sub, nil
	}
	return nil, domainBilling.ErrNotFound
}

func (m *memorySubs) ListAll(_ context.Context) ([]*domainBilling.Subscription, error) {
	out := make([]*domainBilling.Subscription, 0, len(m.byUser))
	for _, sub := range m.byUser {
		out = append(out, sub)
	}
	return out, nil
}

func signBody(secret string, body []byte) string {
	ts := "1755612912"
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + "."))
	mac.Write(body)
	return "t=" + ts + ",v1=" + hex.EncodeToString(mac.Sum(nil))
}

func TestHandleDoitWebhookPaymentPaid(t *testing.T) {
	secret := "whsec_test"
	t.Setenv("DOIT_WEBHOOK_SECRET", secret)

	events := &memoryWebhookEvents{}
	subs := &memorySubs{}
	uc := NewHandleDoitWebhookUseCase(events, subs)

	payload := map[string]interface{}{
		"id":   "evt_paid_1",
		"type": "payment.paid",
		"data": map[string]interface{}{
			"id":        "pay_1",
			"reference": "user:42:monthly",
			"paid_at":   "2026-09-21T12:00:00Z",
			"metadata": map[string]interface{}{
				"user_id":  42,
				"interval": "monthly",
			},
		},
	}
	body, _ := json.Marshal(payload)
	header := signBody(secret, body)

	if err := uc.Execute(context.Background(), header, body); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sub, err := subs.FindByUserID(context.Background(), 42)
	if err != nil {
		t.Fatal(err)
	}
	if !sub.IsPro(time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)) {
		t.Fatal("expected Pro after payment.paid")
	}

	// duplicate must succeed without re-extending incorrectly beyond first activate
	firstEnd := *sub.CurrentPeriodEnd()
	if err := uc.Execute(context.Background(), header, body); err != nil {
		t.Fatalf("duplicate should be ok: %v", err)
	}
	sub2, _ := subs.FindByUserID(context.Background(), 42)
	if !sub2.CurrentPeriodEnd().Equal(firstEnd) {
		t.Fatal("duplicate event must not change period end")
	}
}

func TestHandleDoitWebhookRejectsBadSignature(t *testing.T) {
	t.Setenv("DOIT_WEBHOOK_SECRET", "whsec_test")
	uc := NewHandleDoitWebhookUseCase(&memoryWebhookEvents{}, &memorySubs{})
	body := []byte(`{"id":"evt_x","type":"webhook.test"}`)
	err := uc.Execute(context.Background(), "t=1,v1=deadbeef", body)
	if err != ErrWebhookSignatureInvalid {
		t.Fatalf("expected signature error, got %v", err)
	}
}
