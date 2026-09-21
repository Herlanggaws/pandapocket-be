package finance

import (
	"context"
	"fmt"
	"time"

	appNotification "panda-pocket/internal/application/notification"
	domainFinance "panda-pocket/internal/domain/finance"
	domainIdentity "panda-pocket/internal/domain/identity"
)

// CheckGoalDeadlineAlertsUseCase creates in-app notifications for goals due in 7 or 1 day(s).
type CheckGoalDeadlineAlertsUseCase struct {
	goalService        *domainFinance.GoalService
	walletService      *domainFinance.WalletService
	prefsRepo          domainIdentity.PreferencesRepository
	notificationHelper *appNotification.CreateNotificationHelper
}

func NewCheckGoalDeadlineAlertsUseCase(
	goalService *domainFinance.GoalService,
	walletService *domainFinance.WalletService,
	prefsRepo domainIdentity.PreferencesRepository,
	notificationHelper *appNotification.CreateNotificationHelper,
) *CheckGoalDeadlineAlertsUseCase {
	return &CheckGoalDeadlineAlertsUseCase{
		goalService:        goalService,
		walletService:      walletService,
		prefsRepo:          prefsRepo,
		notificationHelper: notificationHelper,
	}
}

func (uc *CheckGoalDeadlineAlertsUseCase) Execute(ctx context.Context) {
	if uc.notificationHelper == nil || uc.goalService == nil {
		return
	}

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	d7 := today.AddDate(0, 0, 7)
	d1 := today.AddDate(0, 0, 1)

	goals, err := uc.goalService.FindActiveForDeadlineDates(ctx, []time.Time{d7, d1})
	if err != nil || len(goals) == 0 {
		return
	}

	prefCache := map[int]bool{}
	for _, goal := range goals {
		userID := goal.UserID().Value()
		enabled, ok := prefCache[userID]
		if !ok {
			enabled = true
			if uc.prefsRepo != nil {
				prefs, err := uc.prefsRepo.FindByUserID(ctx, domainIdentity.NewUserID(userID))
				if err == nil && prefs != nil {
					enabled = prefs.GoalDeadlineAlerts()
				}
			}
			prefCache[userID] = enabled
		}
		if !enabled {
			continue
		}

		effective := goal.CurrentAmount()
		if goal.WalletID() != nil && uc.walletService != nil {
			if breakdown, balErr := uc.walletService.GetBalance(ctx, goal.UserID(), *goal.WalletID()); balErr == nil {
				effective = breakdown.Balance
			}
		}
		if effective >= goal.TargetAmount() {
			continue
		}

		daysLeft := int(goal.TargetDate().UTC().Sub(today).Hours() / 24)
		milestone := "7d"
		if daysLeft <= 1 {
			milestone = "1d"
		}
		title := fmt.Sprintf("Goal deadline (%s): %s", milestone, goal.Name())
		message := fmt.Sprintf(
			"%s is due in %d day(s). Progress: %.0f%%.",
			goal.Name(),
			daysLeft,
			domainFinance.ProgressPercent(effective, goal.TargetAmount()),
		)
		_ = uc.notificationHelper.CreateIfNotRecent(ctx, userID, title, message, "goal_deadline")
	}
}
