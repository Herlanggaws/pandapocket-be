package billing

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	domainBilling "panda-pocket/internal/domain/billing"
)

type fixedPublicIDs struct{}

func (fixedPublicIDs) PublicID(_ context.Context, userID int) (string, error) {
	return fmt.Sprintf("pub-%d", userID), nil
}

type mapPublicIDs map[int]string

func (m mapPublicIDs) PublicID(_ context.Context, userID int) (string, error) {
	publicID, ok := m[userID]
	if !ok || publicID == "" {
		return "", errors.New("public id not found")
	}
	return publicID, nil
}

type memoryWebhookEvents struct {
	seen map[string]string
}

func (m *memoryWebhookEvents) Exists(_ context.Context, eventID string) (bool, error) {
	if m.seen == nil {
		return false, nil
	}
	_, ok := m.seen[eventID]
	return ok, nil
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
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + "."))
	mac.Write(body)
	return "t=" + ts + ",v1=" + hex.EncodeToString(mac.Sum(nil))
}

func TestHandleDoitWebhookPaymentPaid(t *testing.T) {
	secret := "whsec_test"
	t.Setenv("DOIT_WEBHOOK_SECRET", secret)
	t.Setenv("APP_URL", "https://berbudget.com")

	events := &memoryWebhookEvents{}
	subs := &memorySubs{}
	uc := NewHandleDoitWebhookUseCase(events, subs, nil, &memoryPending{}, fixedPublicIDs{})

	payload := map[string]interface{}{
		"id":   "evt_paid_1",
		"type": "payment.paid",
		"data": map[string]interface{}{
			"id":        "pay_1",
			"reference": "user:42:monthly",
			"amount":    19000,
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
	uc := NewHandleDoitWebhookUseCase(&memoryWebhookEvents{}, &memorySubs{}, nil, &memoryPending{}, fixedPublicIDs{})
	body := []byte(`{"id":"evt_x","type":"webhook.test"}`)
	err := uc.Execute(context.Background(), "t=1,v1=deadbeef", body)
	if err != ErrWebhookSignatureInvalid {
		t.Fatalf("expected signature error, got %v", err)
	}
}

type recordingAICredits struct {
	purchases []string
	unlocks   []int
}

func (r *recordingAICredits) ApplyPurchase(_ context.Context, userID int, pack, doitPaymentID string) error {
	r.purchases = append(r.purchases, fmt.Sprintf("%d:%s:%s", userID, pack, doitPaymentID))
	return nil
}

func (r *recordingAICredits) UnlockFullIncluded(_ context.Context, userID int) error {
	r.unlocks = append(r.unlocks, userID)
	return nil
}

func TestHandleDoitWebhookSkipsStagingBoundOnProd(t *testing.T) {
	secret := "whsec_test"
	t.Setenv("DOIT_WEBHOOK_SECRET", secret)

	ai := &recordingAICredits{}
	events := &memoryWebhookEvents{}
	uc := NewHandleDoitWebhookUseCase(events, &memorySubs{}, ai, &memoryPending{}, fixedPublicIDs{})
	uc.appURL = "https://berbudget.com"

	payload := map[string]interface{}{
		"id":   "evt_stg_ai",
		"type": "payment.paid",
		"data": map[string]interface{}{
			"id":         "pay_stg_m",
			"reference":  "user:32:ai:ai_credits_m:1",
			"return_url": "https://stg.berbudget.com/advisor",
			"metadata": map[string]interface{}{
				"user_id": 32,
				"product": "ai_credits",
				"pack":    "ai_credits_m",
			},
		},
	}
	body, _ := json.Marshal(payload)
	header := signBody(secret, body)

	if err := uc.Execute(context.Background(), header, body); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ai.purchases) != 0 {
		t.Fatalf("prod must not apply staging-bound AI credits, got %v", ai.purchases)
	}
	if _, ok := events.seen["evt_stg_ai"]; !ok {
		t.Fatal("event should still be deduped on prod")
	}
}

func TestHandleDoitWebhookAppliesStagingBoundOnStaging(t *testing.T) {
	secret := "whsec_test"
	t.Setenv("DOIT_WEBHOOK_SECRET", secret)

	ai := &recordingAICredits{}
	uc := NewHandleDoitWebhookUseCase(&memoryWebhookEvents{}, &memorySubs{}, ai, &memoryPending{}, fixedPublicIDs{})
	uc.appURL = "https://stg.berbudget.com"

	payload := map[string]interface{}{
		"id":   "evt_stg_ai_ok",
		"type": "payment.paid",
		"data": map[string]interface{}{
			"id":         "pay_stg_s",
			"reference":  "user:32:ai:ai_credits_s:2",
			"amount":     9900,
			"return_url": "https://stg.berbudget.com/advisor",
			"metadata": map[string]interface{}{
				"user_id": 32,
				"product": "ai_credits",
				"pack":    "ai_credits_s",
			},
		},
	}
	body, _ := json.Marshal(payload)
	header := signBody(secret, body)

	if err := uc.Execute(context.Background(), header, body); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ai.purchases) != 1 || ai.purchases[0] != "32:ai_credits_s:pay_stg_s" {
		t.Fatalf("staging must apply AI credits, got %v", ai.purchases)
	}
	if len(ai.unlocks) != 0 {
		t.Fatalf("credit pack must not unlock included grant, got %v", ai.unlocks)
	}
}

func TestLegacyPaidPayloadsKeepOriginalPeriods(t *testing.T) {
	secret := "whsec_test"
	t.Setenv("DOIT_WEBHOOK_SECRET", secret)
	t.Setenv("APP_URL", "https://berbudget.com")
	paidAt := time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)

	cases := []struct {
		name      string
		eventID   string
		reference string
		interval  string
		userID    int
		days      int
		amount    int
	}{
		{name: "monthly reference", eventID: "evt_legacy_m", reference: "user:8:monthly", interval: "monthly", userID: 8, days: 30, amount: 19000},
		{name: "yearly metadata", eventID: "evt_legacy_y", reference: "user:7:yearly", interval: "yearly", userID: 7, days: 365, amount: 149000},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			subs := &memorySubs{}
			uc := NewHandleDoitWebhookUseCase(&memoryWebhookEvents{}, subs, nil, &memoryPending{}, fixedPublicIDs{})
			payload := map[string]interface{}{
				"id":   tc.eventID,
				"type": "payment.paid",
				"data": map[string]interface{}{
					"id":        "pay_" + tc.eventID,
					"reference": tc.reference,
					"amount":    tc.amount,
					"paid_at":   paidAt.Format(time.RFC3339),
					"metadata": map[string]interface{}{
						"user_id":  tc.userID,
						"interval": tc.interval,
					},
				},
			}
			body, _ := json.Marshal(payload)
			if err := uc.Execute(context.Background(), signBody(secret, body), body); err != nil {
				t.Fatal(err)
			}
			sub, err := subs.FindByUserID(context.Background(), tc.userID)
			if err != nil {
				t.Fatal(err)
			}
			want := paidAt.AddDate(0, 0, tc.days)
			if sub.CurrentPeriodEnd() == nil || !sub.CurrentPeriodEnd().Equal(want) {
				t.Fatalf("period end want %v got %v", want, sub.CurrentPeriodEnd())
			}
			if sub.BillingInterval() == nil || string(*sub.BillingInterval()) != tc.interval {
				t.Fatalf("interval=%v", sub.BillingInterval())
			}
		})
	}
}

func TestSemiannualPaymentActivates183Days(t *testing.T) {
	secret := "whsec_test"
	t.Setenv("DOIT_WEBHOOK_SECRET", secret)
	t.Setenv("APP_URL", "https://berbudget.com")
	paidAt := time.Date(2026, 10, 2, 8, 0, 0, 0, time.UTC)

	subs := &memorySubs{}
	uc := NewHandleDoitWebhookUseCase(&memoryWebhookEvents{}, subs, nil, &memoryPending{}, fixedPublicIDs{})
	payload := map[string]interface{}{
		"id":   "evt_semi",
		"type": "payment.paid",
		"data": map[string]interface{}{
			"id":        "pay_semi",
			"reference": "user:11:semiannual",
			"amount":    99000,
			"paid_at":   paidAt.Format(time.RFC3339),
			"metadata": map[string]interface{}{
				"user_id":  11,
				"interval": "semiannual",
			},
		},
	}
	body, _ := json.Marshal(payload)
	if err := uc.Execute(context.Background(), signBody(secret, body), body); err != nil {
		t.Fatal(err)
	}
	sub, err := subs.FindByUserID(context.Background(), 11)
	if err != nil {
		t.Fatal(err)
	}
	want := paidAt.AddDate(0, 0, 183)
	if sub.CurrentPeriodEnd() == nil || !sub.CurrentPeriodEnd().Equal(want) {
		t.Fatalf("period end want %v got %v", want, sub.CurrentPeriodEnd())
	}
}

func TestHandleDoitWebhookSubscriptionUnlocksIncludedOnce(t *testing.T) {
	secret := "whsec_test"
	t.Setenv("DOIT_WEBHOOK_SECRET", secret)
	t.Setenv("APP_URL", "https://berbudget.com")

	ai := &recordingAICredits{}
	events := &memoryWebhookEvents{}
	uc := NewHandleDoitWebhookUseCase(events, &memorySubs{}, ai, &memoryPending{}, fixedPublicIDs{})

	payload := map[string]interface{}{
		"id":   "evt_pro_unlock",
		"type": "payment.paid",
		"data": map[string]interface{}{
			"id":        "pay_pro_1",
			"reference": "user:42:monthly",
			"amount":    19000,
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
	if err := uc.Execute(context.Background(), header, body); err != nil {
		t.Fatalf("replay should be ok: %v", err)
	}
	if len(ai.unlocks) != 1 || ai.unlocks[0] != 42 {
		t.Fatalf("unlocks=%v", ai.unlocks)
	}
	if len(ai.purchases) != 0 {
		t.Fatalf("subscription payment must not purchase credits, got %v", ai.purchases)
	}
}

type memoryPending struct {
	byID map[string]domainBilling.PendingPayment
}

func (m *memoryPending) Save(_ context.Context, payment domainBilling.PendingPayment) error {
	if m.byID == nil {
		m.byID = map[string]domainBilling.PendingPayment{}
	}
	m.byID[payment.PaymentID] = payment
	return nil
}

func (m *memoryPending) FindByPaymentID(_ context.Context, paymentID string) (domainBilling.PendingPayment, bool, error) {
	payment, ok := m.byID[paymentID]
	return payment, ok, nil
}

func TestRecordedPaymentActivatesProAndIgnoresMetadataInterval(t *testing.T) {
	secret := "whsec_test"
	t.Setenv("DOIT_WEBHOOK_SECRET", secret)
	pending := &memoryPending{}
	if err := pending.Save(context.Background(), domainBilling.PendingPayment{
		PaymentID: "pay_recorded",
		UserID:    42,
		Kind:      domainBilling.KindPro,
		Interval:  "yearly",
		Amount:    149000,
	}); err != nil {
		t.Fatal(err)
	}
	subs := &memorySubs{}
	uc := NewHandleDoitWebhookUseCase(&memoryWebhookEvents{}, subs, nil, pending, fixedPublicIDs{})
	payload := map[string]interface{}{
		"id":   "evt_recorded",
		"type": "payment.paid",
		"data": map[string]interface{}{
			"id":        "pay_recorded",
			"amount":    149000,
			"reference": "user:42:monthly",
			"metadata":  map[string]interface{}{"user_id": 42, "interval": "monthly"},
		},
	}
	body, _ := json.Marshal(payload)
	if err := uc.Execute(context.Background(), signBody(secret, body), body); err != nil {
		t.Fatal(err)
	}
	sub, err := subs.FindByUserID(context.Background(), 42)
	if err != nil {
		t.Fatal(err)
	}
	if sub.BillingInterval() == nil || *sub.BillingInterval() != domainBilling.IntervalYearly {
		t.Fatalf("interval=%v", sub.BillingInterval())
	}
}

func TestPaymentMismatchDoesNotActivate(t *testing.T) {
	secret := "whsec_test"
	t.Setenv("DOIT_WEBHOOK_SECRET", secret)
	pending := &memoryPending{}
	_ = pending.Save(context.Background(), domainBilling.PendingPayment{
		PaymentID: "pay_recorded",
		UserID:    42,
		Kind:      domainBilling.KindPro,
		Interval:  "monthly",
		Amount:    19000,
	})
	subs := &memorySubs{}
	uc := NewHandleDoitWebhookUseCase(&memoryWebhookEvents{}, subs, nil, pending, fixedPublicIDs{})

	cases := []map[string]interface{}{
		{"id": "pay_recorded", "amount": 1, "reference": "user:42:monthly", "metadata": map[string]interface{}{"user_id": 42}},
		{"id": "pay_recorded", "amount": 19000, "reference": "user:9:monthly", "metadata": map[string]interface{}{"user_id": 9}},
		{"id": "pay_unknown", "amount": 1, "reference": "user:42:monthly"},
	}
	for i, data := range cases {
		payload := map[string]interface{}{"id": fmt.Sprintf("evt_bad_%d", i), "type": "payment.paid", "data": data}
		body, _ := json.Marshal(payload)
		err := uc.Execute(context.Background(), signBody(secret, body), body)
		if !errors.Is(err, ErrWebhookPaymentMismatch) {
			t.Fatalf("case %d err=%v", i, err)
		}
	}
	if _, err := subs.FindByUserID(context.Background(), 42); !errors.Is(err, domainBilling.ErrNotFound) {
		t.Fatal("mismatched payment activated Pro")
	}
}

func TestRecordedPublicIDActivatesPro(t *testing.T) {
	secret := "whsec_test"
	t.Setenv("DOIT_WEBHOOK_SECRET", secret)
	ownerPublicID := "11111111-1111-4111-8111-111111111111"
	pending := &memoryPending{}
	if err := pending.Save(context.Background(), domainBilling.PendingPayment{
		PaymentID: "pay_public",
		UserID:    42,
		Kind:      domainBilling.KindPro,
		Interval:  "monthly",
		Amount:    19000,
	}); err != nil {
		t.Fatal(err)
	}
	subs := &memorySubs{}
	uc := NewHandleDoitWebhookUseCase(&memoryWebhookEvents{}, subs, nil, pending, mapPublicIDs{42: ownerPublicID})
	payload := map[string]interface{}{
		"id":   "evt_public",
		"type": "payment.paid",
		"data": map[string]interface{}{
			"id":        "pay_public",
			"amount":    19000,
			"reference": "user:" + ownerPublicID + ":monthly",
			"metadata":  map[string]interface{}{"public_id": ownerPublicID, "interval": "monthly"},
		},
	}
	body, _ := json.Marshal(payload)
	if err := uc.Execute(context.Background(), signBody(secret, body), body); err != nil {
		t.Fatal(err)
	}
	sub, err := subs.FindByUserID(context.Background(), 42)
	if err != nil {
		t.Fatal(err)
	}
	if sub.BillingInterval() == nil || *sub.BillingInterval() != domainBilling.IntervalMonthly {
		t.Fatalf("interval=%v", sub.BillingInterval())
	}
}

func TestRecordedPublicIDMismatchDoesNotActivate(t *testing.T) {
	secret := "whsec_test"
	t.Setenv("DOIT_WEBHOOK_SECRET", secret)
	ownerPublicID := "11111111-1111-4111-8111-111111111111"
	otherPublicID := "22222222-2222-4222-8222-222222222222"
	pending := &memoryPending{}
	_ = pending.Save(context.Background(), domainBilling.PendingPayment{
		PaymentID: "pay_public",
		UserID:    42,
		Kind:      domainBilling.KindPro,
		Interval:  "monthly",
		Amount:    19000,
	})
	subs := &memorySubs{}
	uc := NewHandleDoitWebhookUseCase(&memoryWebhookEvents{}, subs, nil, pending, mapPublicIDs{42: ownerPublicID})
	payload := map[string]interface{}{
		"id":   "evt_public_other",
		"type": "payment.paid",
		"data": map[string]interface{}{
			"id":        "pay_public",
			"amount":    19000,
			"reference": "user:" + otherPublicID + ":monthly",
			"metadata":  map[string]interface{}{"public_id": otherPublicID},
		},
	}
	body, _ := json.Marshal(payload)
	err := uc.Execute(context.Background(), signBody(secret, body), body)
	if !errors.Is(err, ErrWebhookPaymentMismatch) {
		t.Fatalf("err=%v", err)
	}
	if _, findErr := subs.FindByUserID(context.Background(), 42); !errors.Is(findErr, domainBilling.ErrNotFound) {
		t.Fatal("mismatched public id activated Pro")
	}
}

func TestPublicIDWithoutPendingPaymentIsRejected(t *testing.T) {
	secret := "whsec_test"
	t.Setenv("DOIT_WEBHOOK_SECRET", secret)
	publicID := "11111111-1111-4111-8111-111111111111"
	subs := &memorySubs{}
	uc := NewHandleDoitWebhookUseCase(&memoryWebhookEvents{}, subs, nil, &memoryPending{}, fixedPublicIDs{})
	payload := map[string]interface{}{
		"id":   "evt_public_legacy",
		"type": "payment.paid",
		"data": map[string]interface{}{
			"id":        "pay_unknown_public",
			"amount":    19000,
			"reference": "user:" + publicID + ":monthly",
			"metadata":  map[string]interface{}{"public_id": publicID},
		},
	}
	body, _ := json.Marshal(payload)
	err := uc.Execute(context.Background(), signBody(secret, body), body)
	if !errors.Is(err, ErrWebhookPaymentMismatch) {
		t.Fatalf("err=%v", err)
	}
}

func TestStaleSignatureIsRejected(t *testing.T) {
	secret := "whsec_test"
	t.Setenv("DOIT_WEBHOOK_SECRET", secret)
	body := []byte(`{"id":"evt_old","type":"webhook.test"}`)
	header := signBodyAt(secret, body, time.Now().Add(-6*time.Minute).Unix())
	uc := NewHandleDoitWebhookUseCase(&memoryWebhookEvents{}, &memorySubs{}, nil, &memoryPending{}, fixedPublicIDs{})
	err := uc.Execute(context.Background(), header, body)
	if !errors.Is(err, ErrWebhookSignatureExpired) {
		t.Fatalf("err=%v", err)
	}
}

func signBodyAt(secret string, body []byte, unixSeconds int64) string {
	ts := strconv.FormatInt(unixSeconds, 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(ts + "."))
	mac.Write(body)
	return "t=" + ts + ",v1=" + hex.EncodeToString(mac.Sum(nil))
}
