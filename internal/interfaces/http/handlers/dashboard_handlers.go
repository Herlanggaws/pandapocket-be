package handlers

import (
	"net/http"
	"panda-pocket/internal/application/identity"

	"github.com/gin-gonic/gin"
)

// DashboardHandlers handles dashboard-related HTTP requests
type DashboardHandlers struct {
	getDashboardStatsUseCase *identity.GetDashboardStatsUseCase
}

// NewDashboardHandlers creates a new dashboard handlers instance
func NewDashboardHandlers(getDashboardStatsUseCase *identity.GetDashboardStatsUseCase) *DashboardHandlers {
	return &DashboardHandlers{
		getDashboardStatsUseCase: getDashboardStatsUseCase,
	}
}

// GetDashboardStats handles getting dashboard statistics
func (h *DashboardHandlers) GetDashboardStats(c *gin.Context) {
	response, err := h.getDashboardStatsUseCase.Execute(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "Failed to fetch dashboard statistics",
		})
		return
	}

	c.JSON(http.StatusOK, response)
}
