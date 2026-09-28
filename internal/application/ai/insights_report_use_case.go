package ai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	appFinance "panda-pocket/internal/application/finance"
	domainAI "panda-pocket/internal/domain/ai"
	"panda-pocket/internal/domain/entitlement"
	domainFinance "panda-pocket/internal/domain/finance"
	"panda-pocket/internal/infrastructure/paas"
)

// Same floor as advisor chat: a 512 cap is consumed by model reasoning
// before any visible report text.
const insightsReportMaxTokens = 4096

var (
	ErrInsightsReportEmpty = errors.New("insights period has no income or spending")
	ErrInsightsReportBlank = errors.New("insights report blank")
)

type periodAnalytics interface {
	Execute(ctx context.Context, userID int, req appFinance.GetAnalyticsRequest) (*appFinance.GetAnalyticsResponse, error)
}

type monthlyHealth interface {
	Execute(ctx context.Context, userID int) (*appFinance.HealthScoreResponse, error)
}

type primaryCurrency interface {
	Execute(ctx context.Context, userID int) (*domainFinance.Currency, error)
}

type InsightsReportRequest struct {
	Period    string `json:"period"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
}

type InsightsReportResponse struct {
	Summary string       `json:"summary"`
	Points  []string     `json:"points"`
	Period  string       `json:"period"`
	Credits *CreditsView `json:"credits"`
}

type insightsReportPayload struct {
	Period                   string                              `json:"period"`
	CurrencyCode             string                              `json:"currency_code,omitempty"`
	CurrencySymbol           string                              `json:"currency_symbol,omitempty"`
	TotalIncome              float64                             `json:"total_income"`
	TotalSpent               float64                             `json:"total_spent"`
	NetAmount                float64                             `json:"net_amount"`
	TransactionCount         int                                 `json:"transaction_count"`
	ExcludedTransactionCount int                                 `json:"excluded_transaction_count"`
	SpendingByCategory       []appFinance.SpendingByCategoryItem `json:"spending_by_category"`
	SpendingByPeriod         []appFinance.SpendingByPeriodItem   `json:"spending_by_period"`
	HealthScore              *insightsHealthScore                `json:"health_score,omitempty"`
}

type insightsHealthScore struct {
	Score           int     `json:"score"`
	YearMonth       string  `json:"year_month"`
	BudgetAdherence float64 `json:"budget_adherence"`
	Cashflow        float64 `json:"cashflow"`
	Coverage        float64 `json:"coverage"`
}

type InsightsReportUseCase struct {
	credits      *CreditService
	entitlements entitlement.Checker
	analytics    periodAnalytics
	health       monthlyHealth
	currency     primaryCurrency
	completer    ChatCompleter
	prefsLang    func(ctx context.Context, userID int) string
}

func NewInsightsReportUseCase(
	credits *CreditService,
	entitlements entitlement.Checker,
	analytics periodAnalytics,
	health monthlyHealth,
	currency primaryCurrency,
	completer ChatCompleter,
	prefsLang func(ctx context.Context, userID int) string,
) *InsightsReportUseCase {
	return &InsightsReportUseCase{
		credits:      credits,
		entitlements: entitlements,
		analytics:    analytics,
		health:       health,
		currency:     currency,
		completer:    completer,
		prefsLang:    prefsLang,
	}
}

func (uc *InsightsReportUseCase) ViewCredits(ctx context.Context, userID int) (*CreditsView, error) {
	if err := RequireProAI(ctx, uc.entitlements, userID); err != nil {
		return nil, err
	}
	return uc.credits.View(ctx, userID)
}

func (uc *InsightsReportUseCase) Execute(ctx context.Context, userID int, req InsightsReportRequest) (*InsightsReportResponse, error) {
	if err := RequireProAI(ctx, uc.entitlements, userID); err != nil {
		return nil, err
	}

	creditsBefore, err := uc.credits.View(ctx, userID)
	if err != nil {
		return nil, err
	}
	if creditsBefore.Available < 1 {
		return nil, domainAI.ErrCreditsRequired
	}

	analytics, err := uc.analytics.Execute(ctx, userID, appFinance.GetAnalyticsRequest{
		Period:    req.Period,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	})
	if err != nil {
		return nil, err
	}
	if analytics == nil || (analytics.TotalIncome == 0 && analytics.TotalSpent == 0) {
		return nil, ErrInsightsReportEmpty
	}

	payload := uc.buildPayload(ctx, userID, analytics)
	rawPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	reply, err := uc.completer.CompleteChat(ctx, []paas.Message{
		{Role: "system", Content: insightsReportPrompt(uc.language(ctx, userID))},
		{Role: "user", Content: string(rawPayload)},
	}, insightsReportMaxTokens)
	if err != nil {
		if strings.Contains(err.Error(), "not configured") {
			return nil, domainAI.ErrNotConfigured
		}
		return nil, fmt.Errorf("%w", domainAI.ErrUpstream)
	}

	summary, points, err := parseInsightsReport(reply)
	if err != nil {
		return nil, ErrInsightsReportBlank
	}

	creditsAfter, err := uc.credits.SpendOne(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &InsightsReportResponse{
		Summary: summary,
		Points:  points,
		Period:  analytics.Period,
		Credits: creditsAfter,
	}, nil
}

func (uc *InsightsReportUseCase) language(ctx context.Context, userID int) string {
	if uc.prefsLang == nil {
		return "id"
	}
	if lang := uc.prefsLang(ctx, userID); lang != "" {
		return lang
	}
	return "id"
}

func (uc *InsightsReportUseCase) buildPayload(ctx context.Context, userID int, analytics *appFinance.GetAnalyticsResponse) insightsReportPayload {
	payload := insightsReportPayload{
		Period:                   analytics.Period,
		TotalIncome:              analytics.TotalIncome,
		TotalSpent:               analytics.TotalSpent,
		NetAmount:                analytics.NetAmount,
		TransactionCount:         analytics.TransactionCount,
		ExcludedTransactionCount: analytics.ExcludedTransactionCount,
		SpendingByCategory:       analytics.SpendingByCategory,
		SpendingByPeriod:         analytics.SpendingByPeriod,
	}
	if uc.currency != nil {
		cur, curErr := uc.currency.Execute(ctx, userID)
		if curErr == nil && cur != nil {
			payload.CurrencyCode = cur.Code()
			payload.CurrencySymbol = cur.Symbol()
		}
	}
	if analytics.Period == "monthly" && uc.health != nil {
		health, healthErr := uc.health.Execute(ctx, userID)
		if healthErr == nil && health != nil {
			payload.HealthScore = &insightsHealthScore{
				Score:           health.Score,
				YearMonth:       health.YearMonth,
				BudgetAdherence: health.Components.BudgetAdherence,
				Cashflow:        health.Components.Cashflow,
				Coverage:        health.Components.Coverage,
			}
		}
	}
	return payload
}

func insightsReportPrompt(lang string) string {
	base := `You write a short Insights report for Berbudget.
Use ONLY the numbers in the user JSON. Do not invent amounts, categories, or transactions.
Do not recalculate totals. Quote the given totals.
Do not claim you created or changed any record.
This is not licensed financial advice.
Return JSON only, no markdown fences:
{"summary":"one short paragraph","points":["sentence","sentence"]}
points length must be 2 or 3. Each point is one sentence grounded in the JSON.
health_score, when present, is the current calendar month and matches a monthly period only. Do not mention health score when it is absent.`
	if lang == "en" {
		return base + "\nWrite summary and points in English."
	}
	return base + "\nWrite summary and points in Bahasa Indonesia."
}

func parseInsightsReport(raw string) (string, []string, error) {
	text := strings.TrimSpace(raw)
	if strings.HasPrefix(text, "```") {
		text = strings.TrimPrefix(text, "```json")
		text = strings.TrimPrefix(text, "```")
		if idx := strings.LastIndex(text, "```"); idx >= 0 {
			text = text[:idx]
		}
		text = strings.TrimSpace(text)
	}
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start < 0 || end <= start {
		return "", nil, ErrInsightsReportBlank
	}

	var parsed struct {
		Summary string   `json:"summary"`
		Points  []string `json:"points"`
	}
	if err := json.Unmarshal([]byte(text[start:end+1]), &parsed); err != nil {
		return "", nil, ErrInsightsReportBlank
	}

	summary := strings.TrimSpace(parsed.Summary)
	points := make([]string, 0, 3)
	for _, point := range parsed.Points {
		point = strings.TrimSpace(point)
		if point == "" {
			continue
		}
		points = append(points, point)
		if len(points) == 3 {
			break
		}
	}
	if summary == "" || len(points) == 0 {
		return "", nil, ErrInsightsReportBlank
	}
	return summary, points, nil
}
