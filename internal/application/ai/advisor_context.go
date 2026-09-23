package ai

import (
	"context"
	"encoding/json"
	"time"

	appFinance "panda-pocket/internal/application/finance"
)

const (
	maxRecentTransactions = 40
	maxRecentTransfers    = 20
	recentTxLookbackDays  = 45
)

// AdvisorContextDeps gathers read-only finance use cases for Tanya AI context.
type AdvisorContextDeps struct {
	Analytics      *appFinance.GetAnalyticsUseCase
	Liabilities    *appFinance.GetLiabilitiesUseCase
	Assets         *appFinance.GetAssetsUseCase
	Goals          *appFinance.GetGoalsUseCase
	Budgets        *appFinance.GetBudgetsUseCase
	NetWorth       *appFinance.GetNetWorthSummaryUseCase
	Health         *appFinance.GetHealthScoreUseCase
	Wallets        *appFinance.GetWalletsUseCase
	WalletSummary  *appFinance.GetWalletSummaryUseCase
	Recurring      *appFinance.GetRecurringTransactionsUseCase
	Transactions   *appFinance.GetAllTransactionsUseCase
	Transfers      *appFinance.GetTransfersUseCase
	PrimaryCurrency *appFinance.GetDefaultCurrencyUseCase
}

type compactLiability struct {
	Name                     string   `json:"name"`
	Type                     string   `json:"type"`
	CurrencyID               int      `json:"currency_id"`
	CurrentBalance           float64  `json:"current_balance"`
	OriginalPrincipal        *float64 `json:"original_principal,omitempty"`
	InterestRateAPR          *float64 `json:"interest_rate_apr,omitempty"`
	MinimumPayment           float64  `json:"minimum_payment"`
	NextDueDate              *string  `json:"next_due_date,omitempty"`
	PayoffProgressPercent    *float64 `json:"payoff_progress_percent,omitempty"`
	EstimatedMonthsRemaining *int     `json:"estimated_months_remaining,omitempty"`
	Notes                    string   `json:"notes,omitempty"`
}

type compactAsset struct {
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	CurrencyID   int     `json:"currency_id"`
	CurrentValue float64 `json:"current_value"`
	Notes        string  `json:"notes,omitempty"`
}

type compactGoal struct {
	Name            string  `json:"name"`
	TargetAmount    float64 `json:"target_amount"`
	CurrentAmount   float64 `json:"current_amount"`
	CurrencyID      int     `json:"currency_id"`
	TargetDate      string  `json:"target_date"`
	Status          string  `json:"status"`
	ProgressPercent float64 `json:"progress_percent"`
	WalletName      *string `json:"wallet_name,omitempty"`
	ProgressSource  string  `json:"progress_source,omitempty"`
}

type compactBudget struct {
	Category       string  `json:"category"`
	Period         string  `json:"period"`
	EffectiveLimit float64 `json:"effective_limit"`
	Spent          float64 `json:"spent,omitempty"`
	Remaining      float64 `json:"remaining,omitempty"`
	PercentUsed    float64 `json:"percent_used,omitempty"`
	IsOnTrack      *bool   `json:"is_on_track,omitempty"`
}

type compactWallet struct {
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	CurrencyID int     `json:"currency_id"`
	Balance    float64 `json:"balance"`
	IsDefault  bool    `json:"is_default"`
}

type compactRecurring struct {
	Type          string  `json:"type"`
	Amount        float64 `json:"amount"`
	CurrencyID    int     `json:"currency_id"`
	Description   string  `json:"description,omitempty"`
	Frequency     string  `json:"frequency"`
	ScheduleLabel string  `json:"schedule_label,omitempty"`
	NextDueDate   string  `json:"next_due_date,omitempty"`
	Category      string  `json:"category,omitempty"`
	IsActive      bool    `json:"is_active"`
}

type compactTransaction struct {
	Type        string  `json:"type"`
	Amount      float64 `json:"amount"`
	CurrencyID  int     `json:"currency_id"`
	Category    string  `json:"category,omitempty"`
	Description string  `json:"description,omitempty"`
	Date        string  `json:"date"`
	WalletID    int     `json:"wallet_id"`
}

type compactTransfer struct {
	FromWalletID int     `json:"from_wallet_id"`
	ToWalletID   int     `json:"to_wallet_id"`
	Amount       float64 `json:"amount"`
	Description  string  `json:"description,omitempty"`
	Date         string  `json:"date"`
}

type compactCashflow struct {
	Period      string  `json:"period"`
	StartDate   string  `json:"start_date,omitempty"`
	EndDate     string  `json:"end_date,omitempty"`
	TotalIncome float64 `json:"total_income"`
	TotalSpent  float64 `json:"total_spent"`
	NetAmount   float64 `json:"net_amount"`
	CurrencyID  int     `json:"currency_id"`
}

func buildAdvisorContextJSON(ctx context.Context, userID int, credits *CreditsView, deps *AdvisorContextDeps) string {
	payload := map[string]interface{}{
		"generated_at": time.Now().UTC().Format(time.RFC3339),
		"credits":      credits,
		"notes": "Full Berbudget snapshot for this user. Amounts use each row's currency_id; prefer primary_currency. " +
			"liabilities = debts (mortgage/loan/credit_card/other). recent_transactions are the latest " +
			"~45 days (capped). Use only this JSON — do not invent missing accounts.",
	}
	if deps == nil {
		raw, _ := json.Marshal(payload)
		return string(raw)
	}

	if deps.PrimaryCurrency != nil {
		cur, err := deps.PrimaryCurrency.Execute(ctx, userID)
		if err == nil && cur != nil {
			payload["primary_currency"] = map[string]interface{}{
				"id":     cur.ID().Value(),
				"code":   cur.Code(),
				"name":   cur.Name(),
				"symbol": cur.Symbol(),
			}
		}
	}

	if deps.Wallets != nil {
		rows, err := deps.Wallets.Execute(ctx, userID, false)
		if err == nil {
			out := make([]compactWallet, 0, len(rows))
			for _, row := range rows {
				out = append(out, compactWallet{
					Name:       row.Name,
					Type:       row.Type,
					CurrencyID: row.CurrencyID,
					Balance:    row.Balance,
					IsDefault:  row.IsDefault,
				})
			}
			payload["wallets"] = out
		}
	}

	if deps.WalletSummary != nil {
		summary, err := deps.WalletSummary.Execute(ctx, userID)
		if err == nil && summary != nil {
			payload["wallet_summary"] = summary
		}
	}

	if deps.Analytics != nil {
		monthly, err := deps.Analytics.Execute(ctx, userID, appFinance.GetAnalyticsRequest{Period: "monthly"})
		if err == nil && monthly != nil {
			payload["cashflow_monthly"] = compactCashflow{
				Period:      monthly.Period,
				TotalIncome: monthly.TotalIncome,
				TotalSpent:  monthly.TotalSpent,
				NetAmount:   monthly.NetAmount,
				CurrencyID:  monthly.CurrencyID,
			}
			if len(monthly.SpendingByCategory) > 0 {
				top := monthly.SpendingByCategory
				if len(top) > 5 {
					top = top[:5]
				}
				payload["top_expense_categories"] = top
			}
		}

		now := time.Now()
		prevStart := time.Date(now.Year(), now.Month()-1, 1, 0, 0, 0, 0, now.Location())
		prevEnd := prevStart.AddDate(0, 1, -1)
		prev, err := deps.Analytics.Execute(ctx, userID, appFinance.GetAnalyticsRequest{
			Period:    "custom",
			StartDate: prevStart.Format("2006-01-02"),
			EndDate:   prevEnd.Format("2006-01-02"),
		})
		if err == nil && prev != nil {
			payload["cashflow_previous_month"] = compactCashflow{
				Period:      "previous_month",
				StartDate:   prevStart.Format("2006-01-02"),
				EndDate:     prevEnd.Format("2006-01-02"),
				TotalIncome: prev.TotalIncome,
				TotalSpent:  prev.TotalSpent,
				NetAmount:   prev.NetAmount,
				CurrencyID:  prev.CurrencyID,
			}
		}
	}

	if deps.Liabilities != nil {
		rows, err := deps.Liabilities.Execute(ctx, userID, false)
		if err == nil {
			out := make([]compactLiability, 0, len(rows))
			for _, row := range rows {
				out = append(out, compactLiability{
					Name:                     row.Name,
					Type:                     row.Type,
					CurrencyID:               row.CurrencyID,
					CurrentBalance:           row.CurrentBalance,
					OriginalPrincipal:        row.OriginalPrincipal,
					InterestRateAPR:          row.InterestRateAPR,
					MinimumPayment:           row.MinimumPayment,
					NextDueDate:              row.NextDueDate,
					PayoffProgressPercent:    row.PayoffProgressPercent,
					EstimatedMonthsRemaining: row.EstimatedMonthsRemaining,
					Notes:                    row.Notes,
				})
			}
			payload["liabilities"] = out
		}
	}

	if deps.Assets != nil {
		rows, err := deps.Assets.Execute(ctx, userID, false)
		if err == nil {
			out := make([]compactAsset, 0, len(rows))
			for _, row := range rows {
				out = append(out, compactAsset{
					Name:         row.Name,
					Type:         row.Type,
					CurrencyID:   row.CurrencyID,
					CurrentValue: row.CurrentValue,
					Notes:        row.Notes,
				})
			}
			payload["assets"] = out
		}
	}

	if deps.Goals != nil {
		rows, err := deps.Goals.Execute(ctx, userID, false)
		if err == nil {
			out := make([]compactGoal, 0, len(rows))
			for _, row := range rows {
				out = append(out, compactGoal{
					Name:            row.Name,
					TargetAmount:    row.TargetAmount,
					CurrentAmount:   row.CurrentAmount,
					CurrencyID:      row.CurrencyID,
					TargetDate:      row.TargetDate,
					Status:          row.Status,
					ProgressPercent: row.ProgressPercent,
					WalletName:      row.WalletName,
					ProgressSource:  row.ProgressSource,
				})
			}
			payload["goals"] = out
		}
	}

	if deps.Budgets != nil {
		resp, err := deps.Budgets.Execute(ctx, userID)
		if err == nil && resp != nil {
			out := make([]compactBudget, 0, len(resp.Budgets))
			for _, row := range resp.Budgets {
				item := compactBudget{
					Period:         row.Period,
					EffectiveLimit: row.EffectiveAmount,
				}
				if row.Category != nil {
					item.Category = row.Category.Name
				}
				if row.Report != nil {
					item.Spent = row.Report.TotalSpent
					item.Remaining = row.Report.Remaining
					item.PercentUsed = row.Report.PercentageUsed
					onTrack := row.Report.IsOnTrack
					item.IsOnTrack = &onTrack
				}
				out = append(out, item)
			}
			payload["budgets"] = out
		}
	}

	if deps.NetWorth != nil {
		nw, err := deps.NetWorth.Execute(ctx, userID)
		if err == nil && nw != nil {
			payload["net_worth"] = nw
		}
	}

	if deps.Health != nil {
		health, err := deps.Health.Execute(ctx, userID)
		if err == nil && health != nil {
			payload["health"] = health
		}
	}

	if deps.Recurring != nil {
		rows, err := deps.Recurring.Execute(ctx, userID)
		if err == nil {
			out := make([]compactRecurring, 0, len(rows))
			for _, row := range rows {
				item := compactRecurring{
					Type:          row.Type,
					Amount:        row.Amount,
					CurrencyID:    row.CurrencyID,
					Description:   row.Description,
					Frequency:     row.Frequency,
					ScheduleLabel: row.ScheduleLabel,
					NextDueDate:   row.NextDueDate,
					IsActive:      row.IsActive,
				}
				if row.Category != nil {
					item.Category = row.Category.Name
				}
				out = append(out, item)
			}
			payload["recurring"] = out
		}
	}

	if deps.Transactions != nil {
		end := time.Now()
		start := end.AddDate(0, 0, -recentTxLookbackDays)
		resp, err := deps.Transactions.Execute(ctx, userID, appFinance.GetAllTransactionsRequest{
			StartDate: start.Format("2006-01-02"),
			EndDate:   end.Format("2006-01-02"),
			Page:      1,
			Limit:     maxRecentTransactions,
		})
		if err == nil && resp != nil {
			out := make([]compactTransaction, 0, len(resp.Transactions))
			for _, row := range resp.Transactions {
				out = append(out, compactTransaction{
					Type:        row.Type,
					Amount:      row.Amount,
					CurrencyID:  row.CurrencyID,
					Category:    row.Category.Name,
					Description: row.Description,
					Date:        row.Date,
					WalletID:    row.WalletID,
				})
			}
			payload["recent_transactions"] = out
			payload["recent_transactions_meta"] = map[string]interface{}{
				"lookback_days": recentTxLookbackDays,
				"returned":      len(out),
				"total_matched": resp.Total,
			}
		}
	}

	if deps.Transfers != nil {
		end := time.Now()
		start := end.AddDate(0, 0, -recentTxLookbackDays)
		startStr := start.Format("2006-01-02")
		endStr := end.Format("2006-01-02")
		rows, err := deps.Transfers.Execute(ctx, userID, nil, &startStr, &endStr)
		if err == nil {
			if len(rows) > maxRecentTransfers {
				rows = rows[:maxRecentTransfers]
			}
			out := make([]compactTransfer, 0, len(rows))
			for _, row := range rows {
				out = append(out, compactTransfer{
					FromWalletID: row.FromWalletID,
					ToWalletID:   row.ToWalletID,
					Amount:       row.Amount,
					Description:  row.Description,
					Date:         row.Date,
				})
			}
			payload["recent_transfers"] = out
		}
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return `{}`
	}
	return string(raw)
}
