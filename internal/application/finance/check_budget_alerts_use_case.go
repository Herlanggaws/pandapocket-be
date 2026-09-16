package finance

import (
	"context"
	"fmt"

	appNotification "panda-pocket/internal/application/notification"
	domainFinance "panda-pocket/internal/domain/finance"
	domainIdentity "panda-pocket/internal/domain/identity"
)

// CheckBudgetAlertsUseCase creates in-app notifications when a category budget is overspent
type CheckBudgetAlertsUseCase struct {
	budgetService        *domainFinance.BudgetService
	categoryService      *domainFinance.CategoryService
	transactionService   *domainFinance.TransactionService
	prefsRepo            domainIdentity.PreferencesRepository
	notificationHelper   *appNotification.CreateNotificationHelper
}

func NewCheckBudgetAlertsUseCase(
	budgetService *domainFinance.BudgetService,
	categoryService *domainFinance.CategoryService,
	transactionService *domainFinance.TransactionService,
	prefsRepo domainIdentity.PreferencesRepository,
	notificationHelper *appNotification.CreateNotificationHelper,
) *CheckBudgetAlertsUseCase {
	return &CheckBudgetAlertsUseCase{
		budgetService:      budgetService,
		categoryService:    categoryService,
		transactionService: transactionService,
		prefsRepo:          prefsRepo,
		notificationHelper: notificationHelper,
	}
}

func (uc *CheckBudgetAlertsUseCase) Execute(ctx context.Context, userID, categoryID int) {
	if uc.notificationHelper == nil || uc.prefsRepo == nil {
		return
	}

	prefs, err := uc.prefsRepo.FindByUserID(ctx, domainIdentity.NewUserID(userID))
	if err != nil || prefs == nil || !prefs.BudgetAlerts() {
		return
	}

	budgets, err := uc.budgetService.GetBudgetsByUser(ctx, domainFinance.NewUserID(userID))
	if err != nil {
		return
	}

	var categoryName string
	if category, err := uc.categoryService.GetCategoryByID(ctx, domainFinance.NewCategoryID(categoryID)); err == nil {
		categoryName = category.Name()
	} else {
		categoryName = "category"
	}

	for _, budget := range budgets {
		if budget.CategoryID().Value() != categoryID {
			continue
		}
		report, err := calculateBudgetReport(ctx, uc.transactionService, budget)
		if err != nil || report == nil {
			continue
		}
		if report.PercentageUsed < 80 {
			continue
		}

		title := fmt.Sprintf("Budget alert: %s", categoryName)
		var message string
		if report.PercentageUsed >= 100 {
			message = fmt.Sprintf(
				"You've exceeded your %s budget (%.0f%% used).",
				categoryName,
				report.PercentageUsed,
			)
		} else {
			message = fmt.Sprintf(
				"You've used %.0f%% of your %s budget.",
				report.PercentageUsed,
				categoryName,
			)
		}
		_ = uc.notificationHelper.CreateIfNotRecent(ctx, userID, title, message, "budget_alert")
	}
}
