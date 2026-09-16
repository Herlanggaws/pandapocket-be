package handlers

import (
	"net/http"
	appNotification "panda-pocket/internal/application/notification"
	"strconv"

	"github.com/gin-gonic/gin"
)

// NotificationHandlers handles in-app notification HTTP requests
type NotificationHandlers struct {
	getNotificationsUseCase     *appNotification.GetNotificationsUseCase
	markNotificationReadUseCase *appNotification.MarkNotificationReadUseCase
	deleteNotificationUseCase   *appNotification.DeleteNotificationUseCase
}

func NewNotificationHandlers(
	getNotificationsUseCase *appNotification.GetNotificationsUseCase,
	markNotificationReadUseCase *appNotification.MarkNotificationReadUseCase,
	deleteNotificationUseCase *appNotification.DeleteNotificationUseCase,
) *NotificationHandlers {
	return &NotificationHandlers{
		getNotificationsUseCase:     getNotificationsUseCase,
		markNotificationReadUseCase: markNotificationReadUseCase,
		deleteNotificationUseCase:   deleteNotificationUseCase,
	}
}

func (h *NotificationHandlers) GetNotifications(c *gin.Context) {
	userID := c.GetInt("user_id")
	response, err := h.getNotificationsUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusOK, response)
}

func (h *NotificationHandlers) MarkNotificationRead(c *gin.Context) {
	userID := c.GetInt("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ValidationErrorResponse(c, "invalid notification id")
		return
	}
	if err := h.markNotificationReadUseCase.Execute(c.Request.Context(), userID, id); err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"message": "Notification marked as read"})
}

func (h *NotificationHandlers) DeleteNotification(c *gin.Context) {
	userID := c.GetInt("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ValidationErrorResponse(c, "invalid notification id")
		return
	}
	if err := h.deleteNotificationUseCase.Execute(c.Request.Context(), userID, id); err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"message": "Notification deleted"})
}
