package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	appAI "panda-pocket/internal/application/ai"
	domainAI "panda-pocket/internal/domain/ai"
	"panda-pocket/internal/domain/entitlement"

	"github.com/gin-gonic/gin"
)

type AIAdvisorHandlers struct {
	getThread   *appAI.GetThreadUseCase
	clearThread *appAI.ClearThreadUseCase
	chat        *appAI.AdvisorChatUseCase
	topup       *appAI.CreateTopupUseCase
}

func NewAIAdvisorHandlers(
	getThread *appAI.GetThreadUseCase,
	clearThread *appAI.ClearThreadUseCase,
	chat *appAI.AdvisorChatUseCase,
	topup *appAI.CreateTopupUseCase,
) *AIAdvisorHandlers {
	return &AIAdvisorHandlers{
		getThread:   getThread,
		clearThread: clearThread,
		chat:        chat,
		topup:       topup,
	}
}

func (h *AIAdvisorHandlers) GetThread(c *gin.Context) {
	userID := c.GetInt("user_id")
	resp, err := h.getThread.Execute(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, entitlement.ErrPremiumRequired) {
			PremiumRequiredResponse(c, err)
			return
		}
		InternalServerErrorResponse(c, "AI_THREAD_ERROR", "Failed to load advisor thread")
		return
	}
	SuccessResponse(c, http.StatusOK, resp)
}

func (h *AIAdvisorHandlers) ClearThread(c *gin.Context) {
	userID := c.GetInt("user_id")
	if err := h.clearThread.Execute(c.Request.Context(), userID); err != nil {
		if errors.Is(err, entitlement.ErrPremiumRequired) {
			PremiumRequiredResponse(c, err)
			return
		}
		InternalServerErrorResponse(c, "AI_THREAD_CLEAR_ERROR", "Failed to clear advisor thread")
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"cleared": true})
}

func (h *AIAdvisorHandlers) Topup(c *gin.Context) {
	userID := c.GetInt("user_id")
	var req appAI.CreateTopupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, "Invalid request body")
		return
	}
	resp, err := h.topup.Execute(c.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, entitlement.ErrPremiumRequired) {
			PremiumRequiredResponse(c, err)
			return
		}
		if errors.Is(err, domainAI.ErrInvalidPack) {
			ValidationErrorResponse(c, "pack must be ai_credits_s or ai_credits_m")
			return
		}
		if strings.Contains(err.Error(), "not configured") {
			InternalServerErrorResponse(c, "BILLING_NOT_CONFIGURED", "Billing checkout is not configured")
			return
		}
		InternalServerErrorResponse(c, "AI_TOPUP_ERROR", "Failed to create AI credit top-up")
		return
	}
	SuccessResponse(c, http.StatusOK, resp)
}

type chatRequestBody struct {
	Message string `json:"message"`
}

func (h *AIAdvisorHandlers) Chat(c *gin.Context) {
	userID := c.GetInt("user_id")
	var req chatRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, "Invalid request body")
		return
	}

	// Preflight without streaming headers so clients get normal JSON errors.
	thread, err := h.getThread.Execute(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, entitlement.ErrPremiumRequired) {
			PremiumRequiredResponse(c, err)
			return
		}
		InternalServerErrorResponse(c, "AI_THREAD_ERROR", "Failed to prepare chat")
		return
	}
	if thread.Credits != nil && thread.Credits.Available < 1 {
		c.JSON(http.StatusPaymentRequired, APIResponse{
			Status: "error",
			Error: &ErrorResponse{
				ErrorCode:    "AI_CREDITS_REQUIRED",
				ErrorMessage: domainAI.FormatCreditsRequired(),
				Feature:      entitlement.FeatureAIAdvisor,
			},
		})
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.WriteHeader(http.StatusOK)
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		return
	}

	writeEvent := func(event, data string) {
		_, _ = fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, data)
		flusher.Flush()
	}

	credits, err := h.chat.Execute(c.Request.Context(), userID, req.Message, func(delta string) error {
		payload, marshalErr := json.Marshal(delta)
		if marshalErr != nil {
			return marshalErr
		}
		writeEvent("delta", string(payload))
		return nil
	})
	if err != nil {
		if errors.Is(err, entitlement.ErrPremiumRequired) {
			writeEvent("error", `{"error_code":"PREMIUM_REQUIRED","feature":"ai_advisor"}`)
			return
		}
		if errors.Is(err, domainAI.ErrCreditsRequired) {
			writeEvent("error", `{"error_code":"AI_CREDITS_REQUIRED"}`)
			return
		}
		if errors.Is(err, domainAI.ErrMessageTooLong) {
			writeEvent("error", `{"error_code":"VALIDATION_ERROR","error_message":"message too long"}`)
			return
		}
		if errors.Is(err, domainAI.ErrNotConfigured) {
			writeEvent("error", `{"error_code":"AI_NOT_CONFIGURED"}`)
			return
		}
		if errors.Is(err, domainAI.ErrUpstream) || strings.Contains(err.Error(), "ai upstream") {
			writeEvent("error", `{"error_code":"AI_UPSTREAM_ERROR"}`)
			return
		}
		if strings.Contains(err.Error(), "message is required") {
			writeEvent("error", `{"error_code":"VALIDATION_ERROR","error_message":"message is required"}`)
			return
		}
		writeEvent("error", `{"error_code":"AI_CHAT_ERROR"}`)
		return
	}

	if credits != nil {
		writeEvent("done", fmt.Sprintf(
			`{"available":%d,"included_unlocked":%d,"included_used":%d,"purchased_remaining":%d,"is_trialing":%t}`,
			credits.Available, credits.IncludedUnlocked, credits.IncludedUsed, credits.PurchasedRemaining, credits.IsTrialing,
		))
	} else {
		writeEvent("done", `{}`)
	}
}
