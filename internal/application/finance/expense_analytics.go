package finance

import (
	"context"
	"log"

	"panda-pocket/internal/domain/finance"
	"panda-pocket/internal/infrastructure/analytics"
)

// trackExpenseCreated emits expense_added for every expense insert, and
// first_expense_added when the user's total expense count is exactly 1.
// Properties must never include amounts or descriptions.
func trackExpenseCreated(ctx context.Context, svc *finance.TransactionService, userID int, source string) {
	analytics.Capture(userID, "expense_added", map[string]interface{}{
		"source": source,
	})

	expenseType := finance.TransactionTypeExpense
	_, expenseCount, countErr := svc.GetTransactionsByUserWithFilters(
		ctx,
		finance.NewUserID(userID),
		finance.TransactionFilters{TransactionType: &expenseType, Limit: 1},
	)
	if countErr != nil {
		log.Printf("expense analytics: count expenses for user %d: %v", userID, countErr)
		return
	}
	if expenseCount == 1 {
		analytics.Capture(userID, "first_expense_added", map[string]interface{}{
			"source": source,
		})
	}
}
