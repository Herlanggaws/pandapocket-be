package handlers

import (
	"net/http"
	"strconv"

	"panda-pocket/internal/application/finance"

	"github.com/gin-gonic/gin"
)

func (h *FinanceHandlers) GetHealthScoreHistory(c *gin.Context) {
	userID := c.GetInt("user_id")
	limit, _ := strconv.Atoi(c.Query("limit"))
	response, err := h.getHealthScoreHistoryUseCase.Execute(c.Request.Context(), userID, limit)
	if err != nil {
		InternalServerErrorResponse(c, "HEALTH_SCORE_HISTORY_ERROR", "Failed to fetch health score history")
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"history": response})
}

func (h *FinanceHandlers) CreateGoal(c *gin.Context) {
	userID := c.GetInt("user_id")
	var req finance.CreateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, err.Error())
		return
	}
	response, err := h.createGoalUseCase.Execute(c.Request.Context(), userID, req)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusCreated, gin.H{"goal": response})
}

func (h *FinanceHandlers) GetGoals(c *gin.Context) {
	userID := c.GetInt("user_id")
	includeArchived := c.Query("include_archived") == "true" || c.Query("include_archived") == "1"
	response, err := h.getGoalsUseCase.Execute(c.Request.Context(), userID, includeArchived)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"goals": response})
}

func (h *FinanceHandlers) GetGoal(c *gin.Context) {
	userID := c.GetInt("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ValidationErrorResponse(c, "invalid goal id")
		return
	}
	response, err := h.getGoalUseCase.Execute(c.Request.Context(), userID, id)
	if err != nil {
		HandleError(c, err, http.StatusNotFound)
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"goal": response})
}

func (h *FinanceHandlers) UpdateGoal(c *gin.Context) {
	userID := c.GetInt("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ValidationErrorResponse(c, "invalid goal id")
		return
	}
	var req finance.UpdateGoalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, err.Error())
		return
	}
	response, err := h.updateGoalUseCase.Execute(c.Request.Context(), userID, id, req)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"goal": response})
}

func (h *FinanceHandlers) DeleteGoal(c *gin.Context) {
	userID := c.GetInt("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ValidationErrorResponse(c, "invalid goal id")
		return
	}
	if err := h.deleteGoalUseCase.Execute(c.Request.Context(), userID, id); err != nil {
		HandleError(c, err, http.StatusNotFound)
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"message": "Goal deleted successfully"})
}

func (h *FinanceHandlers) CreateAsset(c *gin.Context) {
	userID := c.GetInt("user_id")
	var req finance.CreateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, err.Error())
		return
	}
	response, err := h.createAssetUseCase.Execute(c.Request.Context(), userID, req)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusCreated, gin.H{"asset": response})
}

func (h *FinanceHandlers) GetAssets(c *gin.Context) {
	userID := c.GetInt("user_id")
	includeArchived := c.Query("include_archived") == "true" || c.Query("include_archived") == "1"
	response, err := h.getAssetsUseCase.Execute(c.Request.Context(), userID, includeArchived)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"assets": response})
}

func (h *FinanceHandlers) UpdateAsset(c *gin.Context) {
	userID := c.GetInt("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ValidationErrorResponse(c, "invalid asset id")
		return
	}
	var req finance.UpdateAssetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, err.Error())
		return
	}
	response, err := h.updateAssetUseCase.Execute(c.Request.Context(), userID, id, req)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"asset": response})
}

func (h *FinanceHandlers) ArchiveAsset(c *gin.Context) {
	userID := c.GetInt("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ValidationErrorResponse(c, "invalid asset id")
		return
	}
	response, err := h.archiveAssetUseCase.Execute(c.Request.Context(), userID, id)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"asset": response})
}

func (h *FinanceHandlers) UnarchiveAsset(c *gin.Context) {
	userID := c.GetInt("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ValidationErrorResponse(c, "invalid asset id")
		return
	}
	response, err := h.unarchiveAssetUseCase.Execute(c.Request.Context(), userID, id)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"asset": response})
}

func (h *FinanceHandlers) CreateLiability(c *gin.Context) {
	userID := c.GetInt("user_id")
	var req finance.CreateLiabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, err.Error())
		return
	}
	response, err := h.createLiabilityUseCase.Execute(c.Request.Context(), userID, req)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusCreated, gin.H{"liability": response})
}

func (h *FinanceHandlers) GetLiabilities(c *gin.Context) {
	userID := c.GetInt("user_id")
	includeArchived := c.Query("include_archived") == "true" || c.Query("include_archived") == "1"
	response, err := h.getLiabilitiesUseCase.Execute(c.Request.Context(), userID, includeArchived)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"liabilities": response})
}

func (h *FinanceHandlers) UpdateLiability(c *gin.Context) {
	userID := c.GetInt("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ValidationErrorResponse(c, "invalid liability id")
		return
	}
	var req finance.UpdateLiabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, err.Error())
		return
	}
	response, err := h.updateLiabilityUseCase.Execute(c.Request.Context(), userID, id, req)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"liability": response})
}

func (h *FinanceHandlers) ArchiveLiability(c *gin.Context) {
	userID := c.GetInt("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ValidationErrorResponse(c, "invalid liability id")
		return
	}
	response, err := h.archiveLiabilityUseCase.Execute(c.Request.Context(), userID, id)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"liability": response})
}

func (h *FinanceHandlers) UnarchiveLiability(c *gin.Context) {
	userID := c.GetInt("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ValidationErrorResponse(c, "invalid liability id")
		return
	}
	response, err := h.unarchiveLiabilityUseCase.Execute(c.Request.Context(), userID, id)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"liability": response})
}

func (h *FinanceHandlers) GetNetWorthSummary(c *gin.Context) {
	userID := c.GetInt("user_id")
	response, err := h.getNetWorthSummaryUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusOK, response)
}
