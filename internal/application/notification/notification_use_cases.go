package notification

import (
	"context"
	"errors"
	"time"

	domainNotification "panda-pocket/internal/domain/notification"

	"gorm.io/gorm"
)

type NotificationResponse struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	Type      string `json:"type"`
	IsRead    bool   `json:"is_read"`
	CreatedAt string `json:"created_at"`
}

func toResponse(n *domainNotification.InAppNotification) NotificationResponse {
	return NotificationResponse{
		ID:        n.ID().Value(),
		Title:     n.Title(),
		Message:   n.Message(),
		Type:      n.Type(),
		IsRead:    n.IsRead(),
		CreatedAt: n.CreatedAt().Format(time.RFC3339),
	}
}

type GetNotificationsUseCase struct {
	repo domainNotification.NotificationRepository
}

func NewGetNotificationsUseCase(repo domainNotification.NotificationRepository) *GetNotificationsUseCase {
	return &GetNotificationsUseCase{repo: repo}
}

func (uc *GetNotificationsUseCase) Execute(ctx context.Context, userID int) ([]NotificationResponse, error) {
	items, err := uc.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]NotificationResponse, 0, len(items))
	for _, item := range items {
		result = append(result, toResponse(item))
	}
	return result, nil
}

type MarkNotificationReadUseCase struct {
	repo domainNotification.NotificationRepository
}

func NewMarkNotificationReadUseCase(repo domainNotification.NotificationRepository) *MarkNotificationReadUseCase {
	return &MarkNotificationReadUseCase{repo: repo}
}

func (uc *MarkNotificationReadUseCase) Execute(ctx context.Context, userID, notificationID int) error {
	n, err := uc.repo.FindByIDAndUserID(ctx, domainNotification.NewNotificationID(notificationID), userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("notification not found")
		}
		return err
	}
	n.MarkRead()
	return uc.repo.Save(ctx, n)
}

type DeleteNotificationUseCase struct {
	repo domainNotification.NotificationRepository
}

func NewDeleteNotificationUseCase(repo domainNotification.NotificationRepository) *DeleteNotificationUseCase {
	return &DeleteNotificationUseCase{repo: repo}
}

func (uc *DeleteNotificationUseCase) Execute(ctx context.Context, userID, notificationID int) error {
	err := uc.repo.DeleteByIDAndUserID(ctx, domainNotification.NewNotificationID(notificationID), userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.New("notification not found")
	}
	return err
}

// CreateNotificationHelper creates an in-app notification with 24h dedupe by type+title
type CreateNotificationHelper struct {
	repo domainNotification.NotificationRepository
}

func NewCreateNotificationHelper(repo domainNotification.NotificationRepository) *CreateNotificationHelper {
	return &CreateNotificationHelper{repo: repo}
}

func (h *CreateNotificationHelper) CreateIfNotRecent(ctx context.Context, userID int, title, message, notifType string) error {
	exists, err := h.repo.ExistsRecent(ctx, userID, notifType, title, time.Now().Add(-24*time.Hour))
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	n, err := domainNotification.NewInAppNotification(userID, title, message, notifType)
	if err != nil {
		return err
	}
	return h.repo.Save(ctx, n)
}
