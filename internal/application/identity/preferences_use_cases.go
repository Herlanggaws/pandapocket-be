package identity

import (
	"context"
	"encoding/json"
	"errors"

	domainIdentity "panda-pocket/internal/domain/identity"

	"gorm.io/gorm"
)

// PreferencesResponse is the API shape for user preferences
type PreferencesResponse struct {
	ID                 int             `json:"id"`
	UserID             int             `json:"user_id"`
	PrimaryCurrencyID  int             `json:"primary_currency_id"`
	EmailNotifications bool            `json:"email_notifications"`
	BudgetAlerts       bool            `json:"budget_alerts"`
	RecurringReminders bool            `json:"recurring_reminders"`
	Onboarding         json.RawMessage `json:"onboarding"`
	Goal               interface{}     `json:"goal,omitempty"`
	Topics             interface{}     `json:"topics,omitempty"`
	Cadence            interface{}     `json:"cadence,omitempty"`
	StartPath          interface{}     `json:"start_path,omitempty"`
	OnboardingCompleted interface{}    `json:"onboarding_completed,omitempty"`
}

type preferencesDefaultCurrencyFinder interface {
	FindDefaultCurrencyID(ctx context.Context) (int, error)
}

// GetPreferencesUseCase loads preferences, creating defaults if missing
type GetPreferencesUseCase struct {
	prefsRepo      domainIdentity.PreferencesRepository
	currencyFinder preferencesDefaultCurrencyFinder
}

func NewGetPreferencesUseCase(
	prefsRepo domainIdentity.PreferencesRepository,
	currencyFinder preferencesDefaultCurrencyFinder,
) *GetPreferencesUseCase {
	return &GetPreferencesUseCase{prefsRepo: prefsRepo, currencyFinder: currencyFinder}
}

func (uc *GetPreferencesUseCase) Execute(ctx context.Context, userID int) (*PreferencesResponse, error) {
	prefs, err := uc.prefsRepo.FindByUserID(ctx, domainIdentity.NewUserID(userID))
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		currencyID, cerr := uc.currencyFinder.FindDefaultCurrencyID(ctx)
		if cerr != nil {
			return nil, errors.New("no default currency available to create preferences")
		}
		prefs = domainIdentity.NewUserPreferences(
			domainIdentity.NewUserID(userID),
			currencyID,
			true, true, true,
			json.RawMessage("{}"),
		)
		if err := uc.prefsRepo.Save(ctx, prefs); err != nil {
			return nil, err
		}
	}
	return toPreferencesResponse(prefs), nil
}

// UpdatePreferencesRequest accepts flat onboarding fields and preference flags
type UpdatePreferencesRequest struct {
	PrimaryCurrencyID   *int        `json:"primary_currency_id"`
	EmailNotifications  *bool       `json:"email_notifications"`
	BudgetAlerts        *bool       `json:"budget_alerts"`
	RecurringReminders  *bool       `json:"recurring_reminders"`
	OnboardingCompleted *bool       `json:"onboarding_completed"`
	Goal                interface{} `json:"goal"`
	Topics              interface{} `json:"topics"`
	Cadence             interface{} `json:"cadence"`
	StartPath           interface{} `json:"start_path"`
	Onboarding          json.RawMessage `json:"onboarding"`
}

// UpdatePreferencesUseCase updates preferences
type UpdatePreferencesUseCase struct {
	prefsRepo      domainIdentity.PreferencesRepository
	currencyFinder preferencesDefaultCurrencyFinder
}

func NewUpdatePreferencesUseCase(
	prefsRepo domainIdentity.PreferencesRepository,
	currencyFinder preferencesDefaultCurrencyFinder,
) *UpdatePreferencesUseCase {
	return &UpdatePreferencesUseCase{prefsRepo: prefsRepo, currencyFinder: currencyFinder}
}

type UpdatePreferencesResponse struct {
	Message     string               `json:"message"`
	Preferences *PreferencesResponse `json:"preferences"`
}

func (uc *UpdatePreferencesUseCase) Execute(ctx context.Context, userID int, req UpdatePreferencesRequest) (*UpdatePreferencesResponse, error) {
	getUC := &GetPreferencesUseCase{prefsRepo: uc.prefsRepo, currencyFinder: uc.currencyFinder}
	if _, err := getUC.Execute(ctx, userID); err != nil {
		return nil, err
	}

	prefs, err := uc.prefsRepo.FindByUserID(ctx, domainIdentity.NewUserID(userID))
	if err != nil {
		return nil, err
	}

	if req.PrimaryCurrencyID != nil {
		prefs.SetPrimaryCurrencyID(*req.PrimaryCurrencyID)
	}
	if req.EmailNotifications != nil {
		prefs.SetEmailNotifications(*req.EmailNotifications)
	}
	if req.BudgetAlerts != nil {
		prefs.SetBudgetAlerts(*req.BudgetAlerts)
	}
	if req.RecurringReminders != nil {
		prefs.SetRecurringReminders(*req.RecurringReminders)
	}

	onboardingMap := map[string]interface{}{}
	if len(prefs.Onboarding()) > 0 {
		_ = json.Unmarshal(prefs.Onboarding(), &onboardingMap)
	}
	if req.Onboarding != nil {
		_ = json.Unmarshal(req.Onboarding, &onboardingMap)
	}
	if req.OnboardingCompleted != nil {
		onboardingMap["onboarding_completed"] = *req.OnboardingCompleted
	}
	if req.Goal != nil {
		onboardingMap["goal"] = req.Goal
	}
	if req.Topics != nil {
		onboardingMap["topics"] = req.Topics
	}
	if req.Cadence != nil {
		onboardingMap["cadence"] = req.Cadence
	}
	if req.StartPath != nil {
		onboardingMap["start_path"] = req.StartPath
	}
	merged, err := json.Marshal(onboardingMap)
	if err != nil {
		return nil, err
	}
	prefs.SetOnboarding(merged)

	if err := uc.prefsRepo.Save(ctx, prefs); err != nil {
		return nil, err
	}

	return &UpdatePreferencesResponse{
		Message:     "Preferences updated successfully",
		Preferences: toPreferencesResponse(prefs),
	}, nil
}

func toPreferencesResponse(prefs *domainIdentity.UserPreferences) *PreferencesResponse {
	resp := &PreferencesResponse{
		ID:                 prefs.ID().Value(),
		UserID:             prefs.UserID().Value(),
		PrimaryCurrencyID:  prefs.PrimaryCurrencyID(),
		EmailNotifications: prefs.EmailNotifications(),
		BudgetAlerts:       prefs.BudgetAlerts(),
		RecurringReminders: prefs.RecurringReminders(),
		Onboarding:         prefs.Onboarding(),
	}

	var onboardingMap map[string]interface{}
	if err := json.Unmarshal(prefs.Onboarding(), &onboardingMap); err == nil {
		if v, ok := onboardingMap["goal"]; ok {
			resp.Goal = v
		}
		if v, ok := onboardingMap["topics"]; ok {
			resp.Topics = v
		}
		if v, ok := onboardingMap["cadence"]; ok {
			resp.Cadence = v
		}
		if v, ok := onboardingMap["start_path"]; ok {
			resp.StartPath = v
		}
		if v, ok := onboardingMap["onboarding_completed"]; ok {
			resp.OnboardingCompleted = v
		}
	}
	return resp
}
