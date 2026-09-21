package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	domainBilling "panda-pocket/internal/domain/billing"
	"panda-pocket/internal/infrastructure/doit"
)

var (
	ErrWebhookSignatureInvalid = errors.New("invalid webhook signature")
	ErrWebhookSecretMissing    = errors.New("DOIT_WEBHOOK_SECRET is not configured")
)

type HandleDoitWebhookUseCase struct {
	secret string
	events domainBilling.WebhookEventRepository
	subs   domainBilling.SubscriptionRepository
}

func NewHandleDoitWebhookUseCase(
	events domainBilling.WebhookEventRepository,
	subs domainBilling.SubscriptionRepository,
) *HandleDoitWebhookUseCase {
	return &HandleDoitWebhookUseCase{
		secret: os.Getenv("DOIT_WEBHOOK_SECRET"),
		events: events,
		subs:   subs,
	}
}

type webhookEnvelope struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	CreatedAt string          `json:"created_at"`
	Data      json.RawMessage `json:"data"`
}

type paymentData struct {
	ID        string                 `json:"id"`
	Reference string                 `json:"reference"`
	PaidAt    *string                `json:"paid_at"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// Execute verifies signature, dedups by event id, and applies entitlement when needed.
// Returns ErrWebhookSignatureInvalid for bad signatures (caller must respond non-2xx).
func (uc *HandleDoitWebhookUseCase) Execute(ctx context.Context, signatureHeader string, rawBody []byte) error {
	if uc.secret == "" {
		return ErrWebhookSecretMissing
	}
	if err := doit.VerifyPayBridgeSignature(uc.secret, signatureHeader, rawBody); err != nil {
		return ErrWebhookSignatureInvalid
	}

	var envelope webhookEnvelope
	if err := json.Unmarshal(rawBody, &envelope); err != nil {
		return fmt.Errorf("invalid webhook payload: %w", err)
	}
	if envelope.ID == "" {
		return fmt.Errorf("webhook event id is required")
	}

	switch envelope.Type {
	case "payment.paid":
		if err := uc.handlePaymentPaid(ctx, envelope.Data); err != nil {
			return err
		}
	case "webhook.test", "payment.expired":
		// no entitlement change
	default:
		// ignore unknown types
	}

	// Record after handling so a failed activate can be retried by Doit.
	if _, err := uc.events.TryInsert(ctx, envelope.ID, string(rawBody)); err != nil {
		return err
	}
	return nil
}

func (uc *HandleDoitWebhookUseCase) handlePaymentPaid(ctx context.Context, data json.RawMessage) error {
	var payment paymentData
	if err := json.Unmarshal(data, &payment); err != nil {
		return fmt.Errorf("invalid payment.paid data: %w", err)
	}

	userID, interval, err := resolvePaidUser(payment)
	if err != nil {
		return nil
	}

	paidAt := time.Now().UTC()
	if payment.PaidAt != nil && *payment.PaidAt != "" {
		if parsed, parseErr := time.Parse(time.RFC3339, *payment.PaidAt); parseErr == nil {
			paidAt = parsed.UTC()
		}
	}

	sub, err := uc.subs.FindByUserID(ctx, userID)
	if err != nil {
		if !errors.Is(err, domainBilling.ErrNotFound) {
			return err
		}
		sub, err = domainBilling.NewFreeSubscription(userID)
		if err != nil {
			return err
		}
	}

	if err := sub.ActivatePro(interval, paidAt); err != nil {
		return err
	}
	return uc.subs.Save(ctx, sub)
}

func resolvePaidUser(payment paymentData) (int, domainBilling.BillingInterval, error) {
	var userID int
	var interval domainBilling.BillingInterval

	if payment.Metadata != nil {
		if raw, ok := payment.Metadata["user_id"]; ok {
			parsed, err := coerceInt(raw)
			if err == nil {
				userID = parsed
			}
		}
		if raw, ok := payment.Metadata["interval"]; ok {
			if s, ok := raw.(string); ok {
				interval = domainBilling.BillingInterval(s)
			}
		}
	}

	if userID == 0 && payment.Reference != "" {
		parts := strings.Split(payment.Reference, ":")
		if len(parts) >= 2 && parts[0] == "user" {
			parsed, err := strconv.Atoi(parts[1])
			if err == nil {
				userID = parsed
			}
		}
		if interval == "" && len(parts) >= 3 {
			interval = domainBilling.BillingInterval(parts[2])
		}
	}

	if userID <= 0 {
		return 0, "", fmt.Errorf("user_id not found in payment")
	}
	if interval != domainBilling.IntervalMonthly && interval != domainBilling.IntervalYearly {
		return 0, "", fmt.Errorf("billing interval not found in payment")
	}
	return userID, interval, nil
}

func coerceInt(raw interface{}) (int, error) {
	switch v := raw.(type) {
	case float64:
		return int(v), nil
	case json.Number:
		i, err := v.Int64()
		return int(i), err
	case string:
		return strconv.Atoi(v)
	case int:
		return v, nil
	case int64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("unsupported user_id type")
	}
}
