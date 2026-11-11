package finance

import "panda-pocket/internal/infrastructure/analytics"

// emitBudgetCreated records budget_created after a successful create.
// Properties must never include amounts, names, or descriptions.
func emitBudgetCreated(userID int, categoryID int, period string) {
	analytics.Capture(userID, "budget_created", map[string]interface{}{
		"category_id": categoryID,
		"period":      period,
	})
}
