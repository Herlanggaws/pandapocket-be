package ai

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	domainAI "panda-pocket/internal/domain/ai"
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
	now          func() time.Time
}

func NewCreateTopupUseCase(payments AIPaymentCreator, entitlements entitlement.Checker) *CreateTopupUseCase {
	return &CreateTopupUseCase{
		payments:     payments,
		entitlements: entitlements,
		now:          time.Now,
	}
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

	paymentReq := doit.CreatePaymentRequest{
		Amount: amount,
		Rail:   "any",
		Metadata: map[string]interface{}{
			"user_id": userID,
			"product": domainAI.ProductAICredits,
			"pack":    req.Pack,
		},
	}
	if returnURL := aiReturnURL(); returnURL != "" {
		paymentReq.ReturnURL = returnURL
	}

	payment, err := uc.createUnpaidPayment(ctx, userID, req.Pack, paymentReq)
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
	pack string,
	paymentReq doit.CreatePaymentRequest,
) (*doit.CreatePaymentResponse, error) {
	nowFn := uc.now
	if nowFn == nil {
		nowFn = time.Now
	}

	for attempt := 0; attempt < 2; attempt++ {
		stamp := nowFn().UTC().UnixNano() + int64(attempt)
		paymentReq.Reference = fmt.Sprintf("user:%d:ai:%s:%d", userID, pack, stamp)
		idempotencyKey := fmt.Sprintf("ai-topup:%d:%s:%d", userID, pack, stamp)

		payment, err := uc.payments.CreatePayment(ctx, idempotencyKey, paymentReq)
		if err != nil {
			return nil, err
		}
		if !isPaidPayment(payment) {
			return payment, nil
		}
	}
	return nil, fmt.Errorf("doit returned paid payment for new top-up attempt")
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
