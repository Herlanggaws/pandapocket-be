package ai

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	domainAI "panda-pocket/internal/domain/ai"
	domainBilling "panda-pocket/internal/domain/billing"
	"panda-pocket/internal/domain/entitlement"
	"panda-pocket/internal/infrastructure/doit"
)

type AIPaymentCreator interface {
	Configured() bool
	CreatePayment(ctx context.Context, idempotencyKey string, req doit.CreatePaymentRequest) (*doit.CreatePaymentResponse, error)
}

type CreateTopupRequest struct {
	Pack string `json:"pack"`
}

type CreateTopupResponse struct {
	HostedURL string `json:"hosted_url"`
	PaymentID string `json:"payment_id"`
	Reference string `json:"reference"`
	Pack      string `json:"pack"`
	Credits   int    `json:"credits"`
	Amount    int    `json:"amount"`
}

type CreateTopupUseCase struct {
	payments     AIPaymentCreator
	entitlements entitlement.Checker
	pending      domainBilling.PendingPaymentRepository
	publicIDs    domainBilling.PublicIDLookup
	now          func() time.Time
}

func NewCreateTopupUseCase(payments AIPaymentCreator, entitlements entitlement.Checker, pending domainBilling.PendingPaymentRepository, publicIDs domainBilling.PublicIDLookup) *CreateTopupUseCase {
	return &CreateTopupUseCase{
		payments:     payments,
		entitlements: entitlements,
		pending:      pending,
		publicIDs:    publicIDs,
		now:          time.Now,
	}
}

func (uc *CreateTopupUseCase) ExternalID(ctx context.Context, userID int) (string, error) {
	return lookupTopupPublicID(uc.publicIDs, ctx, userID)
}

func (uc *CreateTopupUseCase) Execute(ctx context.Context, userID int, req CreateTopupRequest) (*CreateTopupResponse, error) {
	if err := RequireProAI(ctx, uc.entitlements, userID); err != nil {
		return nil, err
	}
	if !uc.payments.Configured() {
		return nil, fmt.Errorf("billing checkout is not configured")
	}
	credits, amount, err := domainAI.PackCredits(req.Pack)
	if err != nil {
		return nil, err
	}

	publicID, err := lookupTopupPublicID(uc.publicIDs, ctx, userID)
	if err != nil {
		return nil, err
	}

	paymentReq := doit.CreatePaymentRequest{
		Amount: amount,
		Rail:   "any",
		Metadata: map[string]interface{}{
			"public_id": publicID,
			"product":   domainAI.ProductAICredits,
			"pack":      req.Pack,
		},
	}
	if returnURL := aiReturnURL(); returnURL != "" {
		paymentReq.ReturnURL = returnURL
	}

	payment, err := uc.createUnpaidPayment(ctx, userID, publicID, req.Pack, amount, paymentReq)
	if err != nil {
		return nil, err
	}
	return &CreateTopupResponse{
		HostedURL: payment.HostedURL,
		PaymentID: payment.ID,
		Reference: payment.Reference,
		Pack:      req.Pack,
		Credits:   credits,
		Amount:    amount,
	}, nil
}

func (uc *CreateTopupUseCase) createUnpaidPayment(
	ctx context.Context,
	userID int,
	publicID string,
	pack string,
	amount int,
	paymentReq doit.CreatePaymentRequest,
) (*doit.CreatePaymentResponse, error) {
	nowFn := uc.now
	if nowFn == nil {
		nowFn = time.Now
	}

	for attempt := 0; attempt < 2; attempt++ {
		stamp := nowFn().UTC().UnixNano() + int64(attempt)
		paymentReq.Reference = fmt.Sprintf("user:%s:ai:%s:%d", publicID, pack, stamp)
		idempotencyKey := fmt.Sprintf("ai-topup:%s:%s:%d", publicID, pack, stamp)

		payment, err := uc.payments.CreatePayment(ctx, idempotencyKey, paymentReq)
		if err != nil {
			return nil, err
		}
		if !isPaidPayment(payment) {
			if err := uc.recordPending(ctx, userID, pack, amount, payment); err != nil {
				return nil, err
			}
			return payment, nil
		}
	}
	return nil, fmt.Errorf("doit returned paid payment for new top-up attempt")
}

func (uc *CreateTopupUseCase) recordPending(ctx context.Context, userID int, pack string, amount int, payment *doit.CreatePaymentResponse) error {
	if payment == nil || payment.ID == "" {
		return fmt.Errorf("payment id missing")
	}
	recorded := payment.Amount
	if recorded <= 0 {
		recorded = amount
	}
	return uc.pending.Save(ctx, domainBilling.PendingPayment{
		PaymentID: payment.ID,
		UserID:    userID,
		Kind:      domainBilling.KindAICredits,
		Pack:      pack,
		Amount:    recorded,
	})
}

func lookupTopupPublicID(lookup domainBilling.PublicIDLookup, ctx context.Context, userID int) (string, error) {
	if lookup == nil {
		return "", fmt.Errorf("public id lookup is not configured")
	}
	publicID, err := lookup.PublicID(ctx, userID)
	if err != nil {
		return "", err
	}
	if publicID == "" {
		return "", fmt.Errorf("public id is required")
	}
	return publicID, nil
}

func isPaidPayment(payment *doit.CreatePaymentResponse) bool {
	if payment == nil {
		return false
	}
	return strings.EqualFold(payment.Status, "paid")
}

func aiReturnURL() string {
	if v := os.Getenv("DOIT_AI_RETURN_URL"); v != "" {
		return v
	}
	return os.Getenv("DOIT_RETURN_URL")
}
