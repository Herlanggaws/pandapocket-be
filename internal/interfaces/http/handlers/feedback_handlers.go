package handlers

import (
	"net/http"
	appFeedback "panda-pocket/internal/application/feedback"

	"github.com/gin-gonic/gin"
)

type FeedbackHandlers struct {
	submitFeedbackUseCase *appFeedback.SubmitFeedbackUseCase
}

func NewFeedbackHandlers(submitFeedbackUseCase *appFeedback.SubmitFeedbackUseCase) *FeedbackHandlers {
	return &FeedbackHandlers{submitFeedbackUseCase: submitFeedbackUseCase}
}

func (h *FeedbackHandlers) SubmitFeedback(c *gin.Context) {
	userID := c.GetInt("user_id")
	var req appFeedback.SubmitFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, err.Error())
		return
	}

	response, err := h.submitFeedbackUseCase.Execute(c.Request.Context(), userID, req)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}

	SuccessResponse(c, http.StatusCreated, gin.H{"feedback": response})
}
