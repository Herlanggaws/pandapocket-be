package handlers

import (
	"errors"
	"net/http"

	appTicket "panda-pocket/internal/application/ticket"
	domainTicket "panda-pocket/internal/domain/ticket"

	"github.com/gin-gonic/gin"
)

type TicketHandlers struct {
	createTicketUseCase       *appTicket.CreateTicketUseCase
	listTicketsUseCase        *appTicket.ListTicketsUseCase
	getTicketUseCase          *appTicket.GetTicketUseCase
	reopenTicketUseCase       *appTicket.ReopenTicketUseCase
	listAdminTicketsUseCase   *appTicket.ListAdminTicketsUseCase
	getAdminTicketUseCase     *appTicket.GetAdminTicketUseCase
	updateTicketStatusUseCase *appTicket.UpdateTicketStatusUseCase
}

func NewTicketHandlers(
	createTicketUseCase *appTicket.CreateTicketUseCase,
	listTicketsUseCase *appTicket.ListTicketsUseCase,
	getTicketUseCase *appTicket.GetTicketUseCase,
	reopenTicketUseCase *appTicket.ReopenTicketUseCase,
	listAdminTicketsUseCase *appTicket.ListAdminTicketsUseCase,
	getAdminTicketUseCase *appTicket.GetAdminTicketUseCase,
	updateTicketStatusUseCase *appTicket.UpdateTicketStatusUseCase,
) *TicketHandlers {
	return &TicketHandlers{
		createTicketUseCase:       createTicketUseCase,
		listTicketsUseCase:        listTicketsUseCase,
		getTicketUseCase:          getTicketUseCase,
		reopenTicketUseCase:       reopenTicketUseCase,
		listAdminTicketsUseCase:   listAdminTicketsUseCase,
		getAdminTicketUseCase:     getAdminTicketUseCase,
		updateTicketStatusUseCase: updateTicketStatusUseCase,
	}
}

func (h *TicketHandlers) CreateTicket(c *gin.Context) {
	userID := c.GetInt("user_id")
	var req appTicket.CreateTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, err.Error())
		return
	}

	response, err := h.createTicketUseCase.Execute(c.Request.Context(), userID, req)
	if err != nil {
		handleTicketError(c, err)
		return
	}

	SuccessResponse(c, http.StatusCreated, gin.H{"ticket": response})
}

func (h *TicketHandlers) ListTickets(c *gin.Context) {
	userID := c.GetInt("user_id")
	response, err := h.listTicketsUseCase.Execute(
		c.Request.Context(),
		userID,
		c.Query("status"),
		c.Query("category"),
	)
	if err != nil {
		handleTicketError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{"tickets": response})
}

func (h *TicketHandlers) GetTicket(c *gin.Context) {
	userID := c.GetInt("user_id")
	response, err := h.getTicketUseCase.Execute(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		handleTicketError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{"ticket": response})
}

func (h *TicketHandlers) ReopenTicket(c *gin.Context) {
	userID := c.GetInt("user_id")
	response, err := h.reopenTicketUseCase.Execute(c.Request.Context(), userID, c.Param("id"))
	if err != nil {
		handleTicketError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{"ticket": response})
}

func (h *TicketHandlers) ListAdminTickets(c *gin.Context) {
	response, err := h.listAdminTicketsUseCase.Execute(
		c.Request.Context(),
		c.Query("status"),
		c.Query("category"),
		c.Query("user_id"),
	)
	if err != nil {
		handleTicketError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{"tickets": response})
}

func (h *TicketHandlers) GetAdminTicket(c *gin.Context) {
	response, err := h.getAdminTicketUseCase.Execute(c.Request.Context(), c.Param("id"))
	if err != nil {
		handleTicketError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{"ticket": response})
}

func (h *TicketHandlers) UpdateTicketStatus(c *gin.Context) {
	var req appTicket.UpdateTicketStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, err.Error())
		return
	}

	response, err := h.updateTicketStatusUseCase.Execute(c.Request.Context(), c.Param("id"), req)
	if err != nil {
		handleTicketError(c, err)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{"ticket": response})
}

func handleTicketError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domainTicket.ErrPremiumRequired):
		ForbiddenResponse(c, "PREMIUM_REQUIRED", err.Error())
	case errors.Is(err, domainTicket.ErrNotFound):
		NotFoundResponse(c, "TICKET_NOT_FOUND", err.Error())
	case errors.Is(err, domainTicket.ErrAccessDenied):
		ForbiddenResponse(c, "ACCESS_DENIED", err.Error())
	case errors.Is(err, domainTicket.ErrCannotReopen):
		BadRequestResponse(c, "INVALID_REQUEST", err.Error())
	default:
		HandleError(c, err, http.StatusBadRequest)
	}
}
