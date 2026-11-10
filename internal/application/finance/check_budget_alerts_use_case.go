package finance

import (
	"context"
	"fmt"
	"time"

	appNotification "panda-pocket/internal/application/notification"
	domainFinance "panda-pocket/internal/domain/finance"
	domainIdentity "panda-pocket/internal/domain/identity"
	"panda-pocket/internal/infrastructure/analytics"
)

const onboardingExpenseEstimateDescription = "Onboarding monthly expense estimate"

// BudgetAlertTrigger identifies what caused CheckBudgetAlerts to run.
type BudgetAlertTrigger string

const (
	BudgetAlertTriggerManualExpense  BudgetAlertTrigger = "manual_expense"
	BudgetAlertTriggerPendingConfirm BudgetAlertTrigger = "pending_confirm"
)

type budgetAlertKind string

const (
	budgetAlertNone      budgetAlertKind = ""
	budgetAlertWarning   budgetAlertKind = "warning"
	budgetAlertFullyUsed budgetAlertKind = "fully_used"
	budgetAlertOver      budgetAlertKind = "over"
)

// CheckBudgetAlertsUseCase creates in-app notifications when a category budget is overspent
type CheckBudgetAlertsUseCase struct {
	budgetService      *domainFinance.BudgetService
	categoryService    *domainFinance.CategoryService
	transactionService *domainFinance.TransactionService
	prefsRepo          domainIdentity.PreferencesRepository
	notificationHelper *appNotification.CreateNotificationHelper
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

func classifyBudgetAlert(percentageUsed float64) budgetAlertKind {
	if percentageUsed < 80 {
		return budgetAlertNone
	}
	if percentageUsed > 100 {
		return budgetAlertOver
	}
	if percentageUsed >= 100 && percentageUsed <= 100 {
		return budgetAlertFullyUsed
	}
	return budgetAlertWarning
}

func isLegacyOnboardingExpenseEstimate(description string) bool {
	return description == onboardingExpenseEstimateDescription
}

func shouldSkipNewBudgetPendingAlert(trigger BudgetAlertTrigger, budgetCreatedAt, now time.Time) bool {
	if trigger != BudgetAlertTriggerPendingConfirm {
		return false
	}
	return now.Sub(budgetCreatedAt) < 24*time.Hour
}

func budgetAlertNotificationCopy(language, categoryName string, kind budgetAlertKind, percentageUsed float64) (title, message string) {
	lang := domainIdentity.NormalizeLanguage(language)
	pct := fmt.Sprintf("%.0f", percentageUsed)

	if lang == "en" {
		title = fmt.Sprintf("Budget %s", categoryName)
		switch kind {
		case budgetAlertOver:
			message = fmt.Sprintf("Spending in %s is over budget (%s%%).", categoryName, pct)
		case budgetAlertFullyUsed:
			message = fmt.Sprintf("Budget %s is fully used this month.", categoryName)
		default:
			message = fmt.Sprintf("Budget %s is %s%% used.", categoryName, pct)
		}
		return title, message
	}

	title = fmt.Sprintf("Budget %s", categoryName)
	switch kind {
	case budgetAlertOver:
		message = fmt.Sprintf("Pengeluaran %s udah lewat budget (%s%%).", categoryName, pct)
	case budgetAlertFullyUsed:
		message = fmt.Sprintf("Budget %s bulan ini udah habis.", categoryName)
	default:
		message = fmt.Sprintf("Budget %s udah kepake %s%%.", categoryName, pct)
	}
	return title, message
}

func (uc *CheckBudgetAlertsUseCase) Execute(
	ctx context.Context,
	userID, categoryID int,
	trigger BudgetAlertTrigger,
	expenseDescription string,
) {
	if uc.notificationHelper == nil || uc.prefsRepo == nil {
		return
	}

	if trigger == BudgetAlertTriggerPendingConfirm && isLegacyOnboardingExpenseEstimate(expenseDescription) {
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

	language := prefs.Language()
	now := time.Now()

	for _, budget := range budgets {
		if budget.CategoryID().Value() != categoryID {
			continue
		}
		if shouldSkipNewBudgetPendingAlert(trigger, budget.CreatedAt(), now) {
			continue
		}

		report, err := calculateBudgetReport(ctx, uc.transactionService, budget)
		if err != nil || report == nil {
			continue
		}

		kind := classifyBudgetAlert(report.PercentageUsed)
		if kind == budgetAlertNone {
			continue
		}

		if kind == budgetAlertOver {
			analytics.Capture(userID, "budget_over", map[string]interface{}{
				"category_id": categoryID,
			})
		}

		title, message := budgetAlertNotificationCopy(language, categoryName, kind, report.PercentageUsed)
		_ = uc.notificationHelper.CreateIfNotRecent(ctx, userID, title, message, "budget_alert")
	}
}
