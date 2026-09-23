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

	domainAI "panda-pocket/internal/domain/ai"
	domainBilling "panda-pocket/internal/domain/billing"
	"panda-pocket/internal/infrastructure/doit"
)

var (
	ErrWebhookSignatureInvalid = errors.New("invalid webhook signature")
	ErrWebhookSecretMissing    = errors.New("DOIT_WEBHOOK_SECRET is not configured")
)

// AICreditApplier credits purchased AI packs after doit payment.paid.
type AICreditApplier interface {
	ApplyPurchase(ctx context.Context, userID int, pack, doitPaymentID string) error
	UnlockFullIncluded(ctx context.Context, userID int) error
}

type HandleDoitWebhookUseCase struct {
	secret   string
	events   domainBilling.WebhookEventRepository
	subs     domainBilling.SubscriptionRepository
	aiCredits AICreditApplier
}

func NewHandleDoitWebhookUseCase(
	events domainBilling.WebhookEventRepository,
	subs domainBilling.SubscriptionRepository,
	aiCredits AICreditApplier,
) *HandleDoitWebhookUseCase {
	return &HandleDoitWebhookUseCase{
		secret:    os.Getenv("DOIT_WEBHOOK_SECRET"),
		events:    events,
		subs:      subs,
		aiCredits: aiCredits,
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

	if isAICreditPayment(payment) {
		return uc.handleAICreditPurchase(ctx, payment)
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
	if err := uc.subs.Save(ctx, sub); err != nil {
		return err
	}
	if uc.aiCredits != nil {
		_ = uc.aiCredits.UnlockFullIncluded(ctx, userID)
	}
	return nil
}

func (uc *HandleDoitWebhookUseCase) handleAICreditPurchase(ctx context.Context, payment paymentData) error {
	if uc.aiCredits == nil {
		return nil
	}
	userID, pack, err := resolveAICreditPurchase(payment)
	if err != nil {
		return nil
	}
	paymentID := payment.ID
	if paymentID == "" {
		return nil
	}
	return uc.aiCredits.ApplyPurchase(ctx, userID, pack, paymentID)
}

func isAICreditPayment(payment paymentData) bool {
	if payment.Metadata != nil {
		if raw, ok := payment.Metadata["product"]; ok {
			if s, ok := raw.(string); ok && s == domainAI.ProductAICredits {
				return true
			}
		}
		if raw, ok := payment.Metadata["pack"]; ok {
			if s, ok := raw.(string); ok && (s == domainAI.PackS || s == domainAI.PackM) {
				return true
			}
		}
	}
	if strings.Contains(payment.Reference, ":ai:") {
		return true
	}
	return false
}

func resolveAICreditPurchase(payment paymentData) (int, string, error) {
	var userID int
	pack := ""

	if payment.Metadata != nil {
		if raw, ok := payment.Metadata["user_id"]; ok {
			if parsed, err := coerceInt(raw); err == nil {
				userID = parsed
			}
		}
		if raw, ok := payment.Metadata["pack"]; ok {
			if s, ok := raw.(string); ok {
				pack = s
			}
		}
	}

	if userID == 0 || pack == "" {
		parts := strings.Split(payment.Reference, ":")
		// user:{id}:ai:{pack}:...
		if len(parts) >= 4 && parts[0] == "user" && parts[2] == "ai" {
			if parsed, err := strconv.Atoi(parts[1]); err == nil {
				userID = parsed
			}
			if pack == "" {
				pack = parts[3]
			}
		}
	}

	if userID <= 0 {
		return 0, "", fmt.Errorf("user_id not found")
	}
	if _, _, err := domainAI.PackCredits(pack); err != nil {
		return 0, "", err
	}
	return userID, pack, nil
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
