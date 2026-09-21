package database

import (
	"context"
	"encoding/json"
	"errors"
	"panda-pocket/internal/domain/identity"

	"gorm.io/gorm"
)

// GormPreferencesRepository implements PreferencesRepository
type GormPreferencesRepository struct {
	db *gorm.DB
}

func NewGormPreferencesRepository(db *gorm.DB) *GormPreferencesRepository {
	return &GormPreferencesRepository{db: db}
}

func (r *GormPreferencesRepository) toDomain(model UserPreferences) *identity.UserPreferences {
	onboarding := json.RawMessage(model.Onboarding)
	if len(onboarding) == 0 {
		onboarding = json.RawMessage("{}")
	}
	return identity.ReconstituteUserPreferences(
		identity.NewPreferencesID(int(model.ID)),
		identity.NewUserID(int(model.UserID)),
		int(model.PrimaryCurrencyID),
		model.EmailNotifications,
		model.BudgetAlerts,
		model.RecurringReminders,
		model.GoalDeadlineAlerts,
		model.Language,
		onboarding,
		model.CreatedAt,
		model.UpdatedAt,
	)
}

func (r *GormPreferencesRepository) Save(ctx context.Context, prefs *identity.UserPreferences) error {
	onboarding := prefs.Onboarding()
	if len(onboarding) == 0 {
		onboarding = json.RawMessage("{}")
	}

	model := &UserPreferences{
		UserID:             uint(prefs.UserID().Value()),
		PrimaryCurrencyID:  uint(prefs.PrimaryCurrencyID()),
		EmailNotifications: prefs.EmailNotifications(),
		BudgetAlerts:       prefs.BudgetAlerts(),
		RecurringReminders: prefs.RecurringReminders(),
		GoalDeadlineAlerts: prefs.GoalDeadlineAlerts(),
		Language:           prefs.Language(),
		Onboarding:         JSONRaw(onboarding),
	}

	if prefs.ID().Value() != 0 {
		model.ID = uint(prefs.ID().Value())
		return r.db.WithContext(ctx).Model(&UserPreferences{}).Where("id = ?", model.ID).Updates(map[string]interface{}{
			"primary_currency_id":  model.PrimaryCurrencyID,
			"email_notifications":  model.EmailNotifications,
			"budget_alerts":        model.BudgetAlerts,
			"recurring_reminders":  model.RecurringReminders,
			"goal_deadline_alerts": model.GoalDeadlineAlerts,
			"language":             model.Language,
			"onboarding":           string(onboarding),
		}).Error
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	prefs.AssignID(identity.NewPreferencesID(int(model.ID)))
	return nil
}

func (r *GormPreferencesRepository) FindByUserID(ctx context.Context, userID identity.UserID) (*identity.UserPreferences, error) {
	var model UserPreferences
	err := r.db.WithContext(ctx).Where("user_id = ?", userID.Value()).First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return r.toDomain(model), nil
}

// FindDefaultCurrencyID returns a seeded system currency ID for bootstrapping prefs (prefers IDR).
func (r *GormPreferencesRepository) FindDefaultCurrencyID(ctx context.Context) (int, error) {
	var currency Currency
	err := r.db.WithContext(ctx).
		Where("code = ? AND user_id IS NULL", "IDR").
		First(&currency).Error
	if err == nil {
		return int(currency.ID), nil
	}

	err = r.db.WithContext(ctx).Where("is_default = ? AND user_id IS NULL", true).First(&currency).Error
	if err != nil {
		err = r.db.WithContext(ctx).First(&currency).Error
		if err != nil {
			return 0, err
		}
	}
	return int(currency.ID), nil
}
