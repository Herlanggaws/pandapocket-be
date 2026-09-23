package ai

import (
	"context"
	"fmt"
	"os"
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
}

func NewCreateTopupUseCase(payments AIPaymentCreator, entitlements entitlement.Checker) *CreateTopupUseCase {
	return &CreateTopupUseCase{payments: payments, entitlements: entitlements}
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

	reference := fmt.Sprintf("user:%d:ai:%s:%s", userID, req.Pack, time.Now().UTC().Format("2006-01-02"))
	idempotencyKey := fmt.Sprintf("ai-topup:%d:%s:%s", userID, req.Pack, time.Now().UTC().Format("2006-01-02T15"))

	paymentReq := doit.CreatePaymentRequest{
		Amount:    amount,
		Rail:      "any",
		Reference: reference,
		Metadata: map[string]interface{}{
			"user_id": userID,
			"product": domainAI.ProductAICredits,
			"pack":    req.Pack,
		},
	}
	if returnURL := aiReturnURL(); returnURL != "" {
		paymentReq.ReturnURL = returnURL
	}

	payment, err := uc.payments.CreatePayment(ctx, idempotencyKey, paymentReq)
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

func aiReturnURL() string {
	if v := os.Getenv("DOIT_AI_RETURN_URL"); v != "" {
		return v
	}
	return os.Getenv("DOIT_RETURN_URL")
}
