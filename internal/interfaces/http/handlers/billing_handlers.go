package handlers

import (
	"errors"
	"io"
	"net/http"
	"strings"

	appBilling "panda-pocket/internal/application/billing"

	"github.com/gin-gonic/gin"
)

type BillingHandlers struct {
	getSubscriptionUseCase    *appBilling.GetSubscriptionUseCase
	createCheckoutUseCase     *appBilling.CreateCheckoutUseCase
	handleDoitWebhookUseCase  *appBilling.HandleDoitWebhookUseCase
	cancelSubscriptionUseCase *appBilling.CancelSubscriptionUseCase
}

func NewBillingHandlers(
	getSubscriptionUseCase *appBilling.GetSubscriptionUseCase,
	createCheckoutUseCase *appBilling.CreateCheckoutUseCase,
	handleDoitWebhookUseCase *appBilling.HandleDoitWebhookUseCase,
	cancelSubscriptionUseCase *appBilling.CancelSubscriptionUseCase,
) *BillingHandlers {
	return &BillingHandlers{
		getSubscriptionUseCase:    getSubscriptionUseCase,
		createCheckoutUseCase:     createCheckoutUseCase,
		handleDoitWebhookUseCase:  handleDoitWebhookUseCase,
		cancelSubscriptionUseCase: cancelSubscriptionUseCase,
	}
}

func (h *BillingHandlers) GetSubscription(c *gin.Context) {
	userID := c.GetInt("user_id")
	response, err := h.getSubscriptionUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		InternalServerErrorResponse(c, "GET_SUBSCRIPTION_ERROR", "Failed to get subscription")
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"subscription": response})
}

func (h *BillingHandlers) Checkout(c *gin.Context) {
	userID := c.GetInt("user_id")
	var req appBilling.CreateCheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, "Invalid request body")
		return
	}

	response, err := h.createCheckoutUseCase.Execute(c.Request.Context(), userID, req)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "interval must be") {
			ValidationErrorResponse(c, msg)
			return
		}
		if strings.Contains(msg, "not configured") {
			InternalServerErrorResponse(c, "BILLING_NOT_CONFIGURED", "Billing checkout is not configured")
			return
		}
		InternalServerErrorResponse(c, "CHECKOUT_ERROR", "Failed to create checkout")
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"hosted_url": response.HostedURL,
		"payment_id": response.PaymentID,
		"reference":  response.Reference,
	})
}

func (h *BillingHandlers) CancelSubscription(c *gin.Context) {
	userID := c.GetInt("user_id")
	response, err := h.cancelSubscriptionUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "no active") {
			ValidationErrorResponse(c, msg)
			return
		}
		InternalServerErrorResponse(c, "CANCEL_SUBSCRIPTION_ERROR", "Failed to cancel subscription")
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"subscription": response})
}

func (h *BillingHandlers) HandleDoitWebhook(c *gin.Context) {
	rawBody, err := io.ReadAll(c.Request.Body)
	if err != nil {
		BadRequestResponse(c, "INVALID_BODY", "Failed to read request body")
		return
	}

	signature := c.GetHeader("PayBridge-Signature")
	err = h.handleDoitWebhookUseCase.Execute(c.Request.Context(), signature, rawBody)
	if err != nil {
		if errors.Is(err, appBilling.ErrWebhookSignatureInvalid) {
			UnauthorizedResponse(c, "INVALID_SIGNATURE", "Invalid webhook signature")
			return
		}
		if errors.Is(err, appBilling.ErrWebhookSecretMissing) {
			InternalServerErrorResponse(c, "WEBHOOK_NOT_CONFIGURED", "Webhook secret is not configured")
			return
		}
		InternalServerErrorResponse(c, "WEBHOOK_ERROR", "Failed to process webhook")
		return
	}

	c.Status(http.StatusOK)
}
