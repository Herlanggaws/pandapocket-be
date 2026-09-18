package handlers

import (
	"net/http"

	appBilling "panda-pocket/internal/application/billing"

	"github.com/gin-gonic/gin"
)

type BillingHandlers struct {
	getSubscriptionUseCase *appBilling.GetSubscriptionUseCase
}

func NewBillingHandlers(getSubscriptionUseCase *appBilling.GetSubscriptionUseCase) *BillingHandlers {
	return &BillingHandlers{getSubscriptionUseCase: getSubscriptionUseCase}
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
