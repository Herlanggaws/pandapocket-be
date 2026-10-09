package billing

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	domainAI "panda-pocket/internal/domain/ai"
	domainBilling "panda-pocket/internal/domain/billing"
	"panda-pocket/internal/infrastructure/doit"

	"github.com/google/uuid"
)

var (
	ErrWebhookSignatureInvalid = errors.New("invalid webhook signature")
	ErrWebhookSignatureExpired = errors.New("webhook signature expired")
	ErrWebhookSecretMissing    = errors.New("DOIT_WEBHOOK_SECRET is not configured")
	ErrWebhookPaymentMismatch  = errors.New("payment does not match a recorded checkout")
)

// AICreditApplier credits purchased AI packs after doit payment.paid.
type AICreditApplier interface {
	ApplyPurchase(ctx context.Context, userID int, pack, doitPaymentID string) error
	UnlockFullIncluded(ctx context.Context, userID int) error
}

type HandleDoitWebhookUseCase struct {
	secret    string
	events    domainBilling.WebhookEventRepository
	subs      domainBilling.SubscriptionRepository
	pending   domainBilling.PendingPaymentRepository
	publicIDs domainBilling.PublicIDLookup
	aiCredits AICreditApplier
	appURL    string
}

func NewHandleDoitWebhookUseCase(
	events domainBilling.WebhookEventRepository,
	subs domainBilling.SubscriptionRepository,
	aiCredits AICreditApplier,
	pending domainBilling.PendingPaymentRepository,
	publicIDs domainBilling.PublicIDLookup,
) *HandleDoitWebhookUseCase {
	return &HandleDoitWebhookUseCase{
		secret:    os.Getenv("DOIT_WEBHOOK_SECRET"),
		events:    events,
		subs:      subs,
		pending:   pending,
		publicIDs: publicIDs,
		aiCredits: aiCredits,
		appURL:    os.Getenv("APP_URL"),
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
	Amount    *float64               `json:"amount"`
	PaidAt    *string                `json:"paid_at"`
	ReturnURL string                 `json:"return_url"`
	Metadata  map[string]interface{} `json:"metadata"`
}

func (uc *HandleDoitWebhookUseCase) Execute(ctx context.Context, signatureHeader string, rawBody []byte) error {
	if uc.secret == "" {
		return ErrWebhookSecretMissing
	}
	if err := doit.VerifyPayBridgeSignature(uc.secret, signatureHeader, rawBody); err != nil {
		if errors.Is(err, doit.ErrSignatureExpired) {
			return ErrWebhookSignatureExpired
		}
		return ErrWebhookSignatureInvalid
	}

	var envelope webhookEnvelope
	if err := json.Unmarshal(rawBody, &envelope); err != nil {
		return fmt.Errorf("invalid webhook payload: %w", err)
	}
	if envelope.ID == "" {
		return fmt.Errorf("webhook event id is required")
	}

	seen, err := uc.events.Exists(ctx, envelope.ID)
	if err != nil {
		return err
	}
	if seen {
		return nil
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
	if uc.shouldSkipStagingBoundPayment(payment) {
		log.Printf("doit webhook: skipping staging-bound payment on non-stg host payment_id=%s return_url=%s app_url=%s",
			payment.ID, payment.ReturnURL, uc.appURL)
		return nil
	}
	amount, amountOK := wholeRupiah(payment.Amount)
	recorded, found, err := uc.findPending(ctx, payment.ID)
	if err != nil {
		return err
	}
	if found {
		return uc.applyRecordedPayment(ctx, payment, recorded, amount, amountOK)
	}
	return uc.applyLegacyPayment(ctx, payment, amount, amountOK)
}

func (uc *HandleDoitWebhookUseCase) findPending(ctx context.Context, paymentID string) (domainBilling.PendingPayment, bool, error) {
	if uc.pending == nil || paymentID == "" {
		return domainBilling.PendingPayment{}, false, nil
	}
	return uc.pending.FindByPaymentID(ctx, paymentID)
}

func (uc *HandleDoitWebhookUseCase) applyRecordedPayment(ctx context.Context, payment paymentData, recorded domainBilling.PendingPayment, amount int, amountOK bool) error {
	claim, ok := claimedIdentity(payment)
	if !ok || !amountOK || amount != recorded.Amount {
		return ErrWebhookPaymentMismatch
	}
	if claim.userID != 0 && claim.userID != recorded.UserID {
		return ErrWebhookPaymentMismatch
	}
	if claim.publicID != "" {
		ownerPublicID, err := lookupPublicID(uc.publicIDs, ctx, recorded.UserID)
		if err != nil {
			return err
		}
		if !strings.EqualFold(claim.publicID, ownerPublicID) {
			return ErrWebhookPaymentMismatch
		}
	}
	switch recorded.Kind {
	case domainBilling.KindPro:
		interval := domainBilling.BillingInterval(recorded.Interval)
		if _, priceOK := domainBilling.PriceForInterval(interval); !priceOK {
			return ErrWebhookPaymentMismatch
		}
		return uc.activatePro(ctx, recorded.UserID, interval, paidAt(payment))
	case domainBilling.KindAICredits:
		if _, _, err := domainAI.PackCredits(recorded.Pack); err != nil {
			return ErrWebhookPaymentMismatch
		}
		if uc.aiCredits == nil {
			return nil
		}
		return uc.aiCredits.ApplyPurchase(ctx, recorded.UserID, recorded.Pack, recorded.PaymentID)
	default:
		return ErrWebhookPaymentMismatch
	}
}

func (uc *HandleDoitWebhookUseCase) applyLegacyPayment(ctx context.Context, payment paymentData, amount int, amountOK bool) error {
	if !amountOK {
		return ErrWebhookPaymentMismatch
	}
	if userID, interval, ok := legacyProReference(payment.Reference); ok {
		price, priceOK := domainBilling.PriceForInterval(interval)
		if !priceOK || amount != price {
			return ErrWebhookPaymentMismatch
		}
		return uc.activatePro(ctx, userID, interval, paidAt(payment))
	}
	if userID, pack, ok := legacyCreditReference(payment.Reference); ok {
		_, price, err := domainAI.PackCredits(pack)
		if err != nil || amount != price {
			return ErrWebhookPaymentMismatch
		}
		if uc.aiCredits == nil {
			return nil
		}
		return uc.aiCredits.ApplyPurchase(ctx, userID, pack, payment.ID)
	}
	return ErrWebhookPaymentMismatch
}

func (uc *HandleDoitWebhookUseCase) activatePro(ctx context.Context, userID int, interval domainBilling.BillingInterval, paidAt time.Time) error {
	sub, err := uc.subs.FindByUserID(ctx, userID)
	if err != nil {
		if !errors.Is(err, domainBilling.ErrNotFound) {
			return err
		}
		publicID, lookupErr := lookupPublicID(uc.publicIDs, ctx, userID)
		if lookupErr != nil {
			return lookupErr
		}
		sub, err = domainBilling.NewFreeSubscription(userID, publicID)
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

// shouldSkipStagingBoundPayment avoids writing staging Doit payments into the prod DB
// when the test-app webhook URL is mis-pointed at api.berbudget.com.
func (uc *HandleDoitWebhookUseCase) shouldSkipStagingBoundPayment(payment paymentData) bool {
	if !isStagingBoundReturnURL(payment.ReturnURL) {
		return false
	}
	appURL := uc.appURL
	if appURL == "" {
		appURL = os.Getenv("APP_URL")
	}
	if strings.Contains(strings.ToLower(appURL), "stg.") {
		return false
	}
	return true
}

func isStagingBoundReturnURL(returnURL string) bool {
	return strings.Contains(strings.ToLower(returnURL), "stg.berbudget.com")
}

func paidAt(payment paymentData) time.Time {
	if payment.PaidAt != nil && *payment.PaidAt != "" {
		if parsed, err := time.Parse(time.RFC3339, *payment.PaidAt); err == nil {
			return parsed.UTC()
		}
	}
	return time.Now().UTC()
}

func wholeRupiah(amount *float64) (int, bool) {
	if amount == nil || *amount <= 0 || *amount != float64(int(*amount)) {
		return 0, false
	}
	return int(*amount), true
}

type paymentClaim struct {
	userID   int
	publicID string
}

func claimedIdentity(payment paymentData) (paymentClaim, bool) {
	refUserID, refPublicID, refOK := subjectFromReference(payment.Reference)
	metaUserID, metaUserOK := userIDFromMetadata(payment.Metadata)
	metaPublicID, metaPublicOK := publicIDFromMetadata(payment.Metadata)
	if refOK && refUserID != 0 && metaUserOK && refUserID != metaUserID {
		return paymentClaim{}, false
	}
	if refOK && refPublicID != "" && metaPublicOK && !strings.EqualFold(refPublicID, metaPublicID) {
		return paymentClaim{}, false
	}

	claim := paymentClaim{}
	if refOK && refUserID != 0 {
		claim.userID = refUserID
	} else if metaUserOK {
		claim.userID = metaUserID
	}
	if refOK && refPublicID != "" {
		claim.publicID = refPublicID
	} else if metaPublicOK {
		claim.publicID = metaPublicID
	}
	if claim.userID == 0 && claim.publicID == "" {
		return paymentClaim{}, false
	}
	return claim, true
}

func userIDFromMetadata(metadata map[string]interface{}) (int, bool) {
	if metadata == nil {
		return 0, false
	}
	raw, ok := metadata["user_id"]
	if !ok {
		return 0, false
	}
	userID, err := coerceInt(raw)
	if err != nil || userID <= 0 {
		return 0, false
	}
	return userID, true
}

func subjectFromReference(reference string) (int, string, bool) {
	parts := strings.Split(reference, ":")
	if len(parts) < 2 || parts[0] != "user" {
		return 0, "", false
	}
	return parseExternalSubject(parts[1])
}

func publicIDFromMetadata(metadata map[string]interface{}) (string, bool) {
	if metadata == nil {
		return "", false
	}
	raw, ok := metadata["public_id"]
	if !ok {
		return "", false
	}
	value, ok := raw.(string)
	if !ok {
		return "", false
	}
	return parsePublicID(value)
}

func parseExternalSubject(raw string) (int, string, bool) {
	if publicID, ok := parsePublicID(raw); ok {
		return 0, publicID, true
	}
	userID, err := strconv.Atoi(raw)
	if err != nil || userID <= 0 {
		return 0, "", false
	}
	return userID, "", true
}

func parsePublicID(raw string) (string, bool) {
	parsed, err := uuid.Parse(raw)
	if err != nil {
		return "", false
	}
	return parsed.String(), true
}

func legacyProReference(reference string) (int, domainBilling.BillingInterval, bool) {
	parts := strings.Split(reference, ":")
	if len(parts) < 3 || parts[0] != "user" || parts[2] == "ai" {
		return 0, "", false
	}
	userID, err := strconv.Atoi(parts[1])
	if err != nil || userID <= 0 {
		return 0, "", false
	}
	interval := domainBilling.BillingInterval(parts[2])
	if _, ok := domainBilling.PriceForInterval(interval); !ok {
		return 0, "", false
	}
	return userID, interval, true
}

func legacyCreditReference(reference string) (int, string, bool) {
	parts := strings.Split(reference, ":")
	if len(parts) < 4 || parts[0] != "user" || parts[2] != "ai" {
		return 0, "", false
	}
	userID, err := strconv.Atoi(parts[1])
	if err != nil || userID <= 0 {
		return 0, "", false
	}
	if _, _, err := domainAI.PackCredits(parts[3]); err != nil {
		return 0, "", false
	}
	return userID, parts[3], true
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
