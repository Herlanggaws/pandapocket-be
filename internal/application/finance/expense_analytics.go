package finance

import (
	"context"
	"log"

	"panda-pocket/internal/domain/finance"
	"panda-pocket/internal/infrastructure/analytics"
)

// wasFirstExpense reports whether the user currently has zero expenses.
// Call before inserting the next expense so first_expense_added does not depend
// on a post-insert recount (which can miss the new row or error silently).
func wasFirstExpense(ctx context.Context, svc *finance.TransactionService, userID int) bool {
	expenseType := finance.TransactionTypeExpense
	_, expenseCount, countErr := svc.GetTransactionsByUserWithFilters(
		ctx,
		finance.NewUserID(userID),
		finance.TransactionFilters{TransactionType: &expenseType, Limit: 1},
	)
	if countErr != nil {
		log.Printf("expense analytics: count expenses for user %d: %v", userID, countErr)
		return false
	}
	return expenseCount == 0
}

// trackExpenseCreated emits expense_added for every expense insert, and
// first_expense_added when isFirst is true (caller decided via wasFirstExpense).
// Properties must never include amounts or descriptions.
func trackExpenseCreated(userID int, source string, isFirst bool) {
	analytics.Capture(userID, "expense_added", map[string]interface{}{
		"source": source,
	})
	if isFirst {
		analytics.Capture(userID, "first_expense_added", map[string]interface{}{
			"source": source,
		})
	}
}

// emitFirstExpenseOnly fires first_expense_added when isFirst is true.
// Used when the caller emits a custom expense_added (debt / receivable).
func emitFirstExpenseOnly(userID int, source string, isFirst bool) {
	if !isFirst {
		return
	}
	analytics.Capture(userID, "first_expense_added", map[string]interface{}{
		"source": source,
	})
}
