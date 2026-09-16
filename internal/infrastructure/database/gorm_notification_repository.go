package database

import (
	"context"
	"errors"
	"panda-pocket/internal/domain/notification"
	"time"

	"gorm.io/gorm"
)

type GormNotificationRepository struct {
	db *gorm.DB
}

func NewGormNotificationRepository(db *gorm.DB) *GormNotificationRepository {
	return &GormNotificationRepository{db: db}
}

func (r *GormNotificationRepository) toDomain(model Notification) *notification.InAppNotification {
	return notification.ReconstituteInAppNotification(
		notification.NewNotificationID(int(model.ID)),
		int(model.UserID),
		model.Title,
		model.Message,
		model.Type,
		model.IsRead,
		model.CreatedAt,
		model.UpdatedAt,
	)
}

func (r *GormNotificationRepository) Save(ctx context.Context, n *notification.InAppNotification) error {
	model := &Notification{
		UserID:  uint(n.UserID()),
		Title:   n.Title(),
		Message: n.Message(),
		Type:    n.Type(),
		IsRead:  n.IsRead(),
	}

	if n.ID().Value() != 0 {
		model.ID = uint(n.ID().Value())
		return r.db.WithContext(ctx).Model(&Notification{}).Where("id = ?", model.ID).Updates(map[string]interface{}{
			"is_read":    model.IsRead,
			"updated_at": time.Now(),
		}).Error
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	n.AssignID(notification.NewNotificationID(int(model.ID)))
	return nil
}

func (r *GormNotificationRepository) FindByUserID(ctx context.Context, userID int) ([]*notification.InAppNotification, error) {
	var models []Notification
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&models).Error
	if err != nil {
		return nil, err
	}
	result := make([]*notification.InAppNotification, 0, len(models))
	for _, m := range models {
		result = append(result, r.toDomain(m))
	}
	return result, nil
}

func (r *GormNotificationRepository) FindByIDAndUserID(ctx context.Context, id notification.NotificationID, userID int) (*notification.InAppNotification, error) {
	var model Notification
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id.Value(), userID).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return r.toDomain(model), nil
}

func (r *GormNotificationRepository) DeleteByIDAndUserID(ctx context.Context, id notification.NotificationID, userID int) error {
	result := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", id.Value(), userID).Delete(&Notification{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *GormNotificationRepository) ExistsRecent(ctx context.Context, userID int, notifType, title string, since time.Time) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Notification{}).
		Where("user_id = ? AND type = ? AND title = ? AND created_at >= ?", userID, notifType, title, since).
		Count(&count).Error
	return count > 0, err
}
