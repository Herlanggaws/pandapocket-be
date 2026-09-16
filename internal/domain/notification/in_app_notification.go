package notification

import (
	"context"
	"errors"
	"time"
)

// InAppNotification is a user-facing in-app notification
type InAppNotification struct {
	id        NotificationID
	userID    int
	title     string
	message   string
	notifType string
	isRead    bool
	createdAt time.Time
	updatedAt time.Time
}

type NotificationID struct {
	value int
}

func NewNotificationID(id int) NotificationID {
	return NotificationID{value: id}
}

func (n NotificationID) Value() int {
	return n.value
}

func NewInAppNotification(userID int, title, message, notifType string) (*InAppNotification, error) {
	if title == "" {
		return nil, errors.New("title is required")
	}
	if message == "" {
		return nil, errors.New("message is required")
	}
	if notifType == "" {
		notifType = "info"
	}
	now := time.Now()
	return &InAppNotification{
		userID:    userID,
		title:     title,
		message:   message,
		notifType: notifType,
		isRead:    false,
		createdAt: now,
		updatedAt: now,
	}, nil
}

func ReconstituteInAppNotification(
	id NotificationID,
	userID int,
	title, message, notifType string,
	isRead bool,
	createdAt, updatedAt time.Time,
) *InAppNotification {
	return &InAppNotification{
		id:        id,
		userID:    userID,
		title:     title,
		message:   message,
		notifType: notifType,
		isRead:    isRead,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (n *InAppNotification) ID() NotificationID { return n.id }
func (n *InAppNotification) UserID() int         { return n.userID }
func (n *InAppNotification) Title() string       { return n.title }
func (n *InAppNotification) Message() string     { return n.message }
func (n *InAppNotification) Type() string        { return n.notifType }
func (n *InAppNotification) IsRead() bool        { return n.isRead }
func (n *InAppNotification) CreatedAt() time.Time { return n.createdAt }
func (n *InAppNotification) UpdatedAt() time.Time { return n.updatedAt }

func (n *InAppNotification) AssignID(id NotificationID) {
	n.id = id
}

func (n *InAppNotification) MarkRead() {
	n.isRead = true
	n.updatedAt = time.Now()
}

// NotificationRepository persists in-app notifications
type NotificationRepository interface {
	Save(ctx context.Context, notification *InAppNotification) error
	FindByUserID(ctx context.Context, userID int) ([]*InAppNotification, error)
	FindByIDAndUserID(ctx context.Context, id NotificationID, userID int) (*InAppNotification, error)
	DeleteByIDAndUserID(ctx context.Context, id NotificationID, userID int) error
	ExistsRecent(ctx context.Context, userID int, notifType, title string, since time.Time) (bool, error)
}
