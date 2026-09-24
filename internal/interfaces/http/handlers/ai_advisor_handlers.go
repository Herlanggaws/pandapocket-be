package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	appAI "panda-pocket/internal/application/ai"
	domainAI "panda-pocket/internal/domain/ai"
	"panda-pocket/internal/domain/entitlement"

	"github.com/gin-gonic/gin"
)

type AIAdvisorHandlers struct {
	threads *appAI.ThreadUseCases
	chat    *appAI.AdvisorChatUseCase
	topup   *appAI.CreateTopupUseCase
}

func NewAIAdvisorHandlers(
	threads *appAI.ThreadUseCases,
	chat *appAI.AdvisorChatUseCase,
	topup *appAI.CreateTopupUseCase,
) *AIAdvisorHandlers {
	return &AIAdvisorHandlers{
		threads: threads,
		chat:    chat,
		topup:   topup,
	}
}

func (h *AIAdvisorHandlers) ListThreads(c *gin.Context) {
	userID := c.GetInt("user_id")
	list, err := h.threads.List(c.Request.Context(), userID)
	if err != nil {
		h.mapThreadErr(c, err, "AI_THREAD_LIST_ERROR", "Failed to list advisor threads")
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"threads": list})
}

func (h *AIAdvisorHandlers) CreateThread(c *gin.Context) {
	userID := c.GetInt("user_id")
	thread, err := h.threads.Create(c.Request.Context(), userID)
	if err != nil {
		h.mapThreadErr(c, err, "AI_THREAD_CREATE_ERROR", "Failed to create advisor thread")
		return
	}
	SuccessResponse(c, http.StatusCreated, thread)
}

func (h *AIAdvisorHandlers) GetThreadByID(c *gin.Context) {
	userID := c.GetInt("user_id")
	threadID, ok := parseThreadID(c)
	if !ok {
		return
	}
	resp, err := h.threads.Get(c.Request.Context(), userID, threadID)
	if err != nil {
		h.mapThreadErr(c, err, "AI_THREAD_ERROR", "Failed to load advisor thread")
		return
	}
	SuccessResponse(c, http.StatusOK, resp)
}

func (h *AIAdvisorHandlers) DeleteThread(c *gin.Context) {
	userID := c.GetInt("user_id")
	threadID, ok := parseThreadID(c)
	if !ok {
		return
	}
	if err := h.threads.Delete(c.Request.Context(), userID, threadID); err != nil {
		h.mapThreadErr(c, err, "AI_THREAD_DELETE_ERROR", "Failed to delete advisor thread")
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"deleted": true})
}

func (h *AIAdvisorHandlers) RenameThread(c *gin.Context) {
	userID := c.GetInt("user_id")
	threadID, ok := parseThreadID(c)
	if !ok {
		return
	}
	var req appAI.RenameThreadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, "Invalid request body")
		return
	}
	thread, err := h.threads.Rename(c.Request.Context(), userID, threadID, req.Title)
	if err != nil {
		h.mapThreadErr(c, err, "AI_THREAD_RENAME_ERROR", "Failed to rename advisor thread")
		return
	}
	SuccessResponse(c, http.StatusOK, thread)
}

// LegacyGetThread keeps GET /ai/advisor/thread working (newest or create).
func (h *AIAdvisorHandlers) LegacyGetThread(c *gin.Context) {
	userID := c.GetInt("user_id")
	resp, err := h.threads.LegacyGetOrCreate(c.Request.Context(), userID)
	if err != nil {
		h.mapThreadErr(c, err, "AI_THREAD_ERROR", "Failed to load advisor thread")
		return
	}
	SuccessResponse(c, http.StatusOK, resp)
}

func (h *AIAdvisorHandlers) LegacyClearThread(c *gin.Context) {
	userID := c.GetInt("user_id")
	if err := h.threads.LegacyClear(c.Request.Context(), userID); err != nil {
		h.mapThreadErr(c, err, "AI_THREAD_CLEAR_ERROR", "Failed to clear advisor thread")
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
		log.Printf("AI topup error user=%d: %v", userID, err)
		InternalServerErrorResponse(c, "AI_TOPUP_ERROR", "Failed to create AI credit top-up")
		return
	}
	SuccessResponse(c, http.StatusOK, resp)
}

type chatRequestBody struct {
	Message string `json:"message"`
}

func (h *AIAdvisorHandlers) ChatOnThread(c *gin.Context) {
	userID := c.GetInt("user_id")
	threadID, ok := parseThreadID(c)
	if !ok {
		return
	}
	var req chatRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, "Invalid request body")
		return
	}

	events, unsub, err := h.chat.Start(c.Request.Context(), userID, threadID, req.Message)
	if err != nil {
		h.mapChatStartErr(c, err)
		return
	}
	defer unsub()

	h.streamEvents(c, events)
}

// LegacyChat starts chat on the newest (or newly created) thread.
func (h *AIAdvisorHandlers) LegacyChat(c *gin.Context) {
	userID := c.GetInt("user_id")
	var req chatRequestBody
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, "Invalid request body")
		return
	}
	resp, err := h.threads.LegacyGetOrCreate(c.Request.Context(), userID)
	if err != nil {
		h.mapThreadErr(c, err, "AI_THREAD_ERROR", "Failed to prepare chat")
		return
	}
	events, unsub, err := h.chat.Start(c.Request.Context(), userID, resp.ID, req.Message)
	if err != nil {
		h.mapChatStartErr(c, err)
		return
	}
	defer unsub()
	h.streamEvents(c, events)
}

func (h *AIAdvisorHandlers) streamEvents(c *gin.Context, events <-chan appAI.StreamEvent) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	// Disable nginx/proxy response buffering so deltas reach the client promptly.
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		return
	}

	writeEvent := func(event, data string) {
		_, _ = fmt.Fprintf(c.Writer, "event: %s\ndata: %s\n\n", event, data)
		flusher.Flush()
	}

	// Flush an SSE comment immediately so proxies open the stream before TTFT.
	_, _ = fmt.Fprint(c.Writer, ": connected\n\n")
	flusher.Flush()

	notify := c.Request.Context().Done()
	for {
		select {
		case <-notify:
			// Client disconnected — generation continues server-side.
			return
		case ev, open := <-events:
			if !open {
				return
			}
			switch ev.Kind {
			case appAI.StreamDelta:
				payload, marshalErr := json.Marshal(ev.Delta)
				if marshalErr != nil {
					continue
				}
				writeEvent("delta", string(payload))
			case appAI.StreamDone:
				if ev.Credits != nil {
					writeEvent("done", fmt.Sprintf(
						`{"available":%d,"included_unlocked":%d,"included_used":%d,"purchased_remaining":%d,"is_trialing":%t}`,
						ev.Credits.Available, ev.Credits.IncludedUnlocked, ev.Credits.IncludedUsed, ev.Credits.PurchasedRemaining, ev.Credits.IsTrialing,
					))
				} else {
					writeEvent("done", `{}`)
				}
				return
			case appAI.StreamError:
				code := ev.ErrCode
				if code == "" {
					code = "AI_CHAT_ERROR"
				}
				writeEvent("error", fmt.Sprintf(`{"error_code":%q}`, code))
				return
			}
		}
	}
}

func parseThreadID(c *gin.Context) (int, bool) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id < 1 {
		ValidationErrorResponse(c, "invalid thread id")
		return 0, false
	}
	return id, true
}

func (h *AIAdvisorHandlers) mapThreadErr(c *gin.Context, err error, code, msg string) {
	if errors.Is(err, entitlement.ErrPremiumRequired) {
		PremiumRequiredResponse(c, err)
		return
	}
	if errors.Is(err, domainAI.ErrThreadNotFound) {
		c.JSON(http.StatusNotFound, APIResponse{
			Status: "error",
			Error: &ErrorResponse{
				ErrorCode:    "AI_THREAD_NOT_FOUND",
				ErrorMessage: "Advisor thread not found",
			},
		})
		return
	}
	InternalServerErrorResponse(c, code, msg)
}

func (h *AIAdvisorHandlers) mapChatStartErr(c *gin.Context, err error) {
	if errors.Is(err, entitlement.ErrPremiumRequired) {
		PremiumRequiredResponse(c, err)
		return
	}
	if errors.Is(err, domainAI.ErrCreditsRequired) {
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
	if errors.Is(err, domainAI.ErrTurnInProgress) {
		c.JSON(http.StatusConflict, APIResponse{
			Status: "error",
			Error: &ErrorResponse{
				ErrorCode:    "AI_TURN_IN_PROGRESS",
				ErrorMessage: "AI is still generating a reply for this thread",
			},
		})
		return
	}
	if errors.Is(err, domainAI.ErrThreadNotFound) {
		c.JSON(http.StatusNotFound, APIResponse{
			Status: "error",
			Error: &ErrorResponse{
				ErrorCode:    "AI_THREAD_NOT_FOUND",
				ErrorMessage: "Advisor thread not found",
			},
		})
		return
	}
	if errors.Is(err, domainAI.ErrMessageTooLong) {
		ValidationErrorResponse(c, "message too long")
		return
	}
	if errors.Is(err, domainAI.ErrNotConfigured) {
		InternalServerErrorResponse(c, "AI_NOT_CONFIGURED", "AI provider is not configured")
		return
	}
	if strings.Contains(err.Error(), "message is required") {
		ValidationErrorResponse(c, "message is required")
		return
	}
	InternalServerErrorResponse(c, "AI_CHAT_ERROR", "Failed to start chat")
}
