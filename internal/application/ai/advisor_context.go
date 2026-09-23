package ai

import (
	"context"
	"encoding/json"
	"time"

	appFinance "panda-pocket/internal/application/finance"
)

// AdvisorContextDeps gathers read-only finance use cases for Tanya AI context.
type AdvisorContextDeps struct {
	Analytics   *appFinance.GetAnalyticsUseCase
	Liabilities *appFinance.GetLiabilitiesUseCase
	Assets      *appFinance.GetAssetsUseCase
	Goals       *appFinance.GetGoalsUseCase
	Budgets     *appFinance.GetBudgetsUseCase
	NetWorth    *appFinance.GetNetWorthSummaryUseCase
	Health      *appFinance.GetHealthScoreUseCase
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
}

type compactGoal struct {
	Name            string  `json:"name"`
	TargetAmount    float64 `json:"target_amount"`
	CurrentAmount   float64 `json:"current_amount"`
	CurrencyID      int     `json:"currency_id"`
	TargetDate      string  `json:"target_date"`
	Status          string  `json:"status"`
	ProgressPercent float64 `json:"progress_percent"`
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

func buildAdvisorContextJSON(ctx context.Context, userID int, credits *CreditsView, deps *AdvisorContextDeps) string {
	payload := map[string]interface{}{
		"generated_at": time.Now().UTC().Format(time.RFC3339),
		"credits":      credits,
		"notes":        "Amounts are in each item's currency_id; prefer primary-currency rows. Debts/liabilities include mortgage, loan, credit_card, other.",
	}
	if deps == nil {
		raw, _ := json.Marshal(payload)
		return string(raw)
	}

	if deps.Analytics != nil {
		monthly, err := deps.Analytics.Execute(ctx, userID, appFinance.GetAnalyticsRequest{Period: "monthly"})
		if err == nil && monthly != nil {
			payload["cashflow_monthly"] = monthly
			if len(monthly.SpendingByCategory) > 0 {
				top := monthly.SpendingByCategory
				if len(top) > 5 {
					top = top[:5]
				}
				payload["top_expense_categories"] = top
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
				if row.Status != "active" && row.Status != "" {
					continue
				}
				out = append(out, compactGoal{
					Name:            row.Name,
					TargetAmount:    row.TargetAmount,
					CurrentAmount:   row.CurrentAmount,
					CurrencyID:      row.CurrencyID,
					TargetDate:      row.TargetDate,
					Status:          row.Status,
					ProgressPercent: row.ProgressPercent,
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

	raw, err := json.Marshal(payload)
	if err != nil {
		return `{}`
	}
	return string(raw)
}
