package mcp

import (
	"context"

	appFinance "panda-pocket/internal/application/finance"
)

// FinanceReader maps MCP tools onto the same read use cases as the REST handlers.
type FinanceReader struct {
	Wallets      *appFinance.GetWalletsUseCase
	Transactions *appFinance.GetAllTransactionsUseCase
	Budgets      *appFinance.GetBudgetsUseCase
	Goals        *appFinance.GetGoalsUseCase
	NetWorthSum  *appFinance.GetNetWorthSummaryUseCase
}

func (r FinanceReader) ListWallets(ctx context.Context, userID int) (any, error) {
	wallets, err := r.Wallets.Execute(ctx, userID, false)
	if err != nil {
		return nil, err
	}
	return map[string]any{"wallets": wallets}, nil
}

func (r FinanceReader) ListTransactions(ctx context.Context, userID int, startDate, endDate string, limit int) (any, error) {
	return r.Transactions.Execute(ctx, userID, appFinance.GetAllTransactionsRequest{
		StartDate: startDate,
		EndDate:   endDate,
		Page:      1,
		Limit:     limit,
	})
}

func (r FinanceReader) ListBudgets(ctx context.Context, userID int) (any, error) {
	return r.Budgets.Execute(ctx, userID)
}

func (r FinanceReader) ListGoals(ctx context.Context, userID int) (any, error) {
	goals, err := r.Goals.Execute(ctx, userID, false)
	if err != nil {
		return nil, err
	}
	return map[string]any{"goals": goals}, nil
}

func (r FinanceReader) NetWorth(ctx context.Context, userID int) (any, error) {
	return r.NetWorthSum.Execute(ctx, userID)
}
