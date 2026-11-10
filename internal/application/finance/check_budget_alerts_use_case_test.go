package finance

import (
	"context"
	"errors"
	"testing"
	"time"

	appNotification "panda-pocket/internal/application/notification"
	domainIdentity "panda-pocket/internal/domain/identity"
	domainNotification "panda-pocket/internal/domain/notification"
	"panda-pocket/internal/infrastructure/analytics"
)

type alertTestPrefsRepo struct {
	prefs *domainIdentity.UserPreferences
	finds int
}

func (r *alertTestPrefsRepo) FindByUserID(ctx context.Context, userID domainIdentity.UserID) (*domainIdentity.UserPreferences, error) {
	r.finds++
	if r.prefs == nil {
		return nil, errors.New("not found")
	}
	return r.prefs, nil
}

func (r *alertTestPrefsRepo) Save(ctx context.Context, prefs *domainIdentity.UserPreferences) error {
	return nil
}

type alertTestNotifRepo struct {
	saved []*domainNotification.InAppNotification
}

func (r *alertTestNotifRepo) Save(ctx context.Context, notification *domainNotification.InAppNotification) error {
	r.saved = append(r.saved, notification)
	return nil
}
func (r *alertTestNotifRepo) FindByUserID(ctx context.Context, userID int) ([]*domainNotification.InAppNotification, error) {
	return nil, nil
}
func (r *alertTestNotifRepo) FindByIDAndUserID(ctx context.Context, id domainNotification.NotificationID, userID int) (*domainNotification.InAppNotification, error) {
	return nil, errors.New("not found")
}
func (r *alertTestNotifRepo) DeleteByIDAndUserID(ctx context.Context, id domainNotification.NotificationID, userID int) error {
	return nil
}
func (r *alertTestNotifRepo) ExistsRecent(ctx context.Context, userID int, notifType, title string, since time.Time) (bool, error) {
	return false, nil
}

func TestClassifyBudgetAlert(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		pct   float64
		want  budgetAlertKind
	}{
		{name: "below warning", pct: 79.9, want: budgetAlertNone},
		{name: "warning lower bound", pct: 80, want: budgetAlertWarning},
		{name: "just under fully used", pct: 99.9, want: budgetAlertWarning},
		{name: "exactly fully used", pct: 100, want: budgetAlertFullyUsed},
		{name: "just over", pct: 100.01, want: budgetAlertOver},
		{name: "well over", pct: 150, want: budgetAlertOver},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := classifyBudgetAlert(tc.pct); got != tc.want {
				t.Fatalf("classifyBudgetAlert(%v)=%q, want %q", tc.pct, got, tc.want)
			}
		})
	}
}

func TestIsLegacyOnboardingExpenseEstimate(t *testing.T) {
	t.Parallel()

	if !isLegacyOnboardingExpenseEstimate("Onboarding monthly expense estimate") {
		t.Fatal("expected legacy description to match")
	}
	if isLegacyOnboardingExpenseEstimate("Onboarding monthly income") {
		t.Fatal("income description must not match")
	}
	if isLegacyOnboardingExpenseEstimate("Groceries") {
		t.Fatal("unrelated description must not match")
	}
}

func TestShouldSkipNewBudgetPendingAlert(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)
	fresh := now.Add(-2 * time.Hour)
	old := now.Add(-25 * time.Hour)

	if !shouldSkipNewBudgetPendingAlert(BudgetAlertTriggerPendingConfirm, fresh, now) {
		t.Fatal("pending confirm on fresh budget should skip")
	}
	if shouldSkipNewBudgetPendingAlert(BudgetAlertTriggerPendingConfirm, old, now) {
		t.Fatal("pending confirm on old budget should not skip")
	}
	if shouldSkipNewBudgetPendingAlert(BudgetAlertTriggerManualExpense, fresh, now) {
		t.Fatal("manual expense must not skip even for fresh budget")
	}
}

func TestBudgetAlertNotificationCopyLanguage(t *testing.T) {
	t.Parallel()

	t.Run("indonesian default", func(t *testing.T) {
		t.Parallel()
		title, message := budgetAlertNotificationCopy("id", "Other", budgetAlertWarning, 85)
		if title != "Budget Other" {
			t.Fatalf("title=%q", title)
		}
		if message != "Budget Other udah kepake 85%." {
			t.Fatalf("message=%q", message)
		}

		_, fully := budgetAlertNotificationCopy("", "Other", budgetAlertFullyUsed, 100)
		if fully != "Budget Other bulan ini udah habis." {
			t.Fatalf("fully used message=%q", fully)
		}

		_, over := budgetAlertNotificationCopy("unknown", "Other", budgetAlertOver, 120)
		if over != "Pengeluaran Other udah lewat budget (120%)." {
			t.Fatalf("over message=%q", over)
		}
	})

	t.Run("english", func(t *testing.T) {
		t.Parallel()
		title, message := budgetAlertNotificationCopy("en", "Food", budgetAlertWarning, 90)
		if title != "Budget Food" {
			t.Fatalf("title=%q", title)
		}
		if message != "Budget Food is 90% used." {
			t.Fatalf("message=%q", message)
		}

		_, fully := budgetAlertNotificationCopy("en", "Food", budgetAlertFullyUsed, 100)
		if fully != "Budget Food is fully used this month." {
			t.Fatalf("fully used message=%q", fully)
		}

		_, over := budgetAlertNotificationCopy("en", "Food", budgetAlertOver, 110)
		if over != "Spending in Food is over budget (110%)." {
			t.Fatalf("over message=%q", over)
		}
	})

	t.Run("normalize via prefs helper", func(t *testing.T) {
		t.Parallel()
		if domainIdentity.NormalizeLanguage("fr") != "id" {
			t.Fatal("unknown language should default to id")
		}
	})
}

func TestCheckBudgetAlertsSkipsLegacyOnboardingPending(t *testing.T) {
	prefsRepo := &alertTestPrefsRepo{
		prefs: domainIdentity.NewUserPreferences(domainIdentity.NewUserID(1), 1, true, true, true, true, nil),
	}
	notifRepo := &alertTestNotifRepo{}
	helper := appNotification.NewCreateNotificationHelper(notifRepo)

	captured := false
	analytics.SetTestCapture(func(userID int, event string, properties map[string]interface{}) {
		captured = true
	})
	t.Cleanup(func() { analytics.SetTestCapture(nil) })

	uc := NewCheckBudgetAlertsUseCase(nil, nil, nil, prefsRepo, helper)
	uc.Execute(
		context.Background(),
		1,
		8,
		BudgetAlertTriggerPendingConfirm,
		"Onboarding monthly expense estimate",
	)

	if prefsRepo.finds != 0 {
		t.Fatal("legacy skip should return before loading preferences")
	}
	if len(notifRepo.saved) != 0 {
		t.Fatal("legacy skip must not create notifications")
	}
	if captured {
		t.Fatal("legacy skip must not emit budget_over")
	}
}
