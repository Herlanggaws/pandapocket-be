package identity

import (
	"context"
	"encoding/json"
	"time"
)

// UserPreferences holds per-user settings
type UserPreferences struct {
	id                 PreferencesID
	userID             UserID
	primaryCurrencyID  int
	emailNotifications bool
	budgetAlerts       bool
	recurringReminders bool
	goalDeadlineAlerts bool
	language           string
	onboarding         json.RawMessage
	createdAt          time.Time
	updatedAt          time.Time
}

// PreferencesID is a value object for preferences identifier
type PreferencesID struct {
	value int
}

func NewPreferencesID(id int) PreferencesID {
	return PreferencesID{value: id}
}

func (p PreferencesID) Value() int {
	return p.value
}

// UserID for preferences reuses identity UserID

func NormalizeLanguage(language string) string {
	if language == "en" {
		return "en"
	}
	return "id"
}

func NewUserPreferences(
	userID UserID,
	primaryCurrencyID int,
	emailNotifications, budgetAlerts, recurringReminders, goalDeadlineAlerts bool,
	onboarding json.RawMessage,
) *UserPreferences {
	now := time.Now()
	if onboarding == nil {
		onboarding = json.RawMessage("{}")
	}
	return &UserPreferences{
		userID:             userID,
		primaryCurrencyID:  primaryCurrencyID,
		emailNotifications: emailNotifications,
		budgetAlerts:       budgetAlerts,
		recurringReminders: recurringReminders,
		goalDeadlineAlerts: goalDeadlineAlerts,
		language:           "id",
		onboarding:         onboarding,
		createdAt:          now,
		updatedAt:          now,
	}
}

func ReconstituteUserPreferences(
	id PreferencesID,
	userID UserID,
	primaryCurrencyID int,
	emailNotifications, budgetAlerts, recurringReminders, goalDeadlineAlerts bool,
	language string,
	onboarding json.RawMessage,
	createdAt, updatedAt time.Time,
) *UserPreferences {
	if onboarding == nil {
		onboarding = json.RawMessage("{}")
	}
	return &UserPreferences{
		id:                 id,
		userID:             userID,
		primaryCurrencyID:  primaryCurrencyID,
		emailNotifications: emailNotifications,
		budgetAlerts:       budgetAlerts,
		recurringReminders: recurringReminders,
		goalDeadlineAlerts: goalDeadlineAlerts,
		language:           NormalizeLanguage(language),
		onboarding:         onboarding,
		createdAt:          createdAt,
		updatedAt:          updatedAt,
	}
}

func (p *UserPreferences) ID() PreferencesID          { return p.id }
func (p *UserPreferences) UserID() UserID              { return p.userID }
func (p *UserPreferences) PrimaryCurrencyID() int      { return p.primaryCurrencyID }
func (p *UserPreferences) EmailNotifications() bool    { return p.emailNotifications }
func (p *UserPreferences) BudgetAlerts() bool          { return p.budgetAlerts }
func (p *UserPreferences) RecurringReminders() bool    { return p.recurringReminders }
func (p *UserPreferences) GoalDeadlineAlerts() bool    { return p.goalDeadlineAlerts }
func (p *UserPreferences) Language() string            { return NormalizeLanguage(p.language) }
func (p *UserPreferences) Onboarding() json.RawMessage { return p.onboarding }
func (p *UserPreferences) CreatedAt() time.Time        { return p.createdAt }
func (p *UserPreferences) UpdatedAt() time.Time        { return p.updatedAt }

func (p *UserPreferences) AssignID(id PreferencesID) {
	p.id = id
}

func (p *UserPreferences) SetPrimaryCurrencyID(currencyID int) {
	p.primaryCurrencyID = currencyID
	p.updatedAt = time.Now()
}

func (p *UserPreferences) SetEmailNotifications(v bool) {
	p.emailNotifications = v
	p.updatedAt = time.Now()
}

func (p *UserPreferences) SetBudgetAlerts(v bool) {
	p.budgetAlerts = v
	p.updatedAt = time.Now()
}

func (p *UserPreferences) SetRecurringReminders(v bool) {
	p.recurringReminders = v
	p.updatedAt = time.Now()
}

func (p *UserPreferences) SetGoalDeadlineAlerts(v bool) {
	p.goalDeadlineAlerts = v
	p.updatedAt = time.Now()
}

func (p *UserPreferences) SetLanguage(language string) {
	p.language = NormalizeLanguage(language)
	p.updatedAt = time.Now()
}

func (p *UserPreferences) SetOnboarding(data json.RawMessage) {
	if data == nil {
		data = json.RawMessage("{}")
	}
	p.onboarding = data
	p.updatedAt = time.Now()
}

// PreferencesRepository persists user preferences
type PreferencesRepository interface {
	Save(ctx context.Context, prefs *UserPreferences) error
	FindByUserID(ctx context.Context, userID UserID) (*UserPreferences, error)
}
