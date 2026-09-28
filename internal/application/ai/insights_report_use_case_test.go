package ai

import (
	"context"
	"errors"
	"strings"
	"testing"

	appFinance "panda-pocket/internal/application/finance"
	domainAI "panda-pocket/internal/domain/ai"
	"panda-pocket/internal/domain/entitlement"
	domainFinance "panda-pocket/internal/domain/finance"
	"panda-pocket/internal/infrastructure/paas"
)

type neverPro struct{}

func (neverPro) IsPro(context.Context, int) (bool, error) { return false, nil }

type captureCompleter struct {
	reply string
	err   error
	calls int
	last  []paas.Message
}

func (c *captureCompleter) CompleteChat(_ context.Context, messages []paas.Message, _ int) (string, error) {
	c.calls++
	c.last = messages
	if c.err != nil {
		return "", c.err
	}
	return c.reply, nil
}

type stubAnalytics struct {
	resp  *appFinance.GetAnalyticsResponse
	err   error
	calls int
}

func (s *stubAnalytics) Execute(context.Context, int, appFinance.GetAnalyticsRequest) (*appFinance.GetAnalyticsResponse, error) {
	s.calls++
	return s.resp, s.err
}

type stubHealth struct {
	resp  *appFinance.HealthScoreResponse
	calls int
}

func (s *stubHealth) Execute(context.Context, int) (*appFinance.HealthScoreResponse, error) {
	s.calls++
	return s.resp, nil
}

func activeAnalytics() *appFinance.GetAnalyticsResponse {
	return &appFinance.GetAnalyticsResponse{
		TotalIncome: 1000,
		TotalSpent:  400,
		NetAmount:   600,
		Period:      "weekly",
		CurrencyID:  1,
		SpendingByCategory: []appFinance.SpendingByCategoryItem{{
			CategoryName: "Food",
			Amount:       400,
			Percentage:   100,
		}},
	}
}

func reportUseCase(bal *domainAI.CreditBalance, analytics *stubAnalytics, health *stubHealth, completer *captureCompleter, checker entitlement.Checker) (*InsightsReportUseCase, *memCredits) {
	repo := &memCredits{balance: bal}
	credits := NewCreditService(repo, emptySubs{})
	return NewInsightsReportUseCase(credits, checker, analytics, health, nil, completer, func(context.Context, int) string {
		return "en"
	}), repo
}

func TestInsightsReportDebitsAfterSuccess(t *testing.T) {
	analytics := &stubAnalytics{resp: activeAnalytics()}
	completer := &captureCompleter{reply: `{"summary":"Net is positive.","points":["Food is the only expense.","Income covers spending."]}`}
	uc, repo := reportUseCase(domainAI.NewCreditBalance(1, false), analytics, &stubHealth{}, completer, alwaysPro{})

	resp, err := uc.Execute(context.Background(), 1, InsightsReportRequest{Period: "weekly"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Summary == "" || len(resp.Points) != 2 || resp.Period != "weekly" {
		t.Fatalf("resp=%+v", resp)
	}
	if completer.calls != 1 {
		t.Fatalf("calls=%d", completer.calls)
	}
	if repo.balance.IncludedUsed != 1 {
		t.Fatalf("used=%d", repo.balance.IncludedUsed)
	}
	if resp.Credits == nil || resp.Credits.Available != domainAI.IncludedGrant()-1 {
		t.Fatalf("credits=%+v", resp.Credits)
	}
}

func TestInsightsReportUpstreamDoesNotDebit(t *testing.T) {
	analytics := &stubAnalytics{resp: activeAnalytics()}
	completer := &captureCompleter{err: errors.New("upstream down")}
	uc, repo := reportUseCase(domainAI.NewCreditBalance(1, false), analytics, &stubHealth{}, completer, alwaysPro{})

	_, err := uc.Execute(context.Background(), 1, InsightsReportRequest{Period: "weekly"})
	if !errors.Is(err, domainAI.ErrUpstream) {
		t.Fatalf("err=%v", err)
	}
	if repo.balance.IncludedUsed != 0 {
		t.Fatalf("used=%d", repo.balance.IncludedUsed)
	}
}

func TestInsightsReportZeroCreditsSkipsModel(t *testing.T) {
	bal := domainAI.NewCreditBalance(1, false)
	bal.IncludedUsed = domainAI.IncludedGrant()
	analytics := &stubAnalytics{resp: activeAnalytics()}
	completer := &captureCompleter{reply: `{"summary":"x","points":["a","b"]}`}
	uc, _ := reportUseCase(bal, analytics, &stubHealth{}, completer, alwaysPro{})

	_, err := uc.Execute(context.Background(), 1, InsightsReportRequest{Period: "weekly"})
	if !errors.Is(err, domainAI.ErrCreditsRequired) {
		t.Fatalf("err=%v", err)
	}
	if completer.calls != 0 || analytics.calls != 0 {
		t.Fatalf("model=%d analytics=%d", completer.calls, analytics.calls)
	}
}

func TestInsightsReportEmptyPeriodSkipsModel(t *testing.T) {
	analytics := &stubAnalytics{resp: &appFinance.GetAnalyticsResponse{Period: "monthly"}}
	completer := &captureCompleter{reply: `{"summary":"x","points":["a"]}`}
	uc, repo := reportUseCase(domainAI.NewCreditBalance(1, false), analytics, &stubHealth{}, completer, alwaysPro{})

	_, err := uc.Execute(context.Background(), 1, InsightsReportRequest{Period: "monthly"})
	if !errors.Is(err, ErrInsightsReportEmpty) {
		t.Fatalf("err=%v", err)
	}
	if completer.calls != 0 {
		t.Fatal("model was called")
	}
	if repo.balance.IncludedUsed != 0 {
		t.Fatalf("used=%d", repo.balance.IncludedUsed)
	}
}

func TestInsightsReportFreeIsPremiumRequired(t *testing.T) {
	analytics := &stubAnalytics{resp: activeAnalytics()}
	completer := &captureCompleter{}
	uc, _ := reportUseCase(domainAI.NewCreditBalance(1, false), analytics, &stubHealth{}, completer, neverPro{})

	_, err := uc.Execute(context.Background(), 1, InsightsReportRequest{Period: "weekly"})
	if !errors.Is(err, entitlement.ErrPremiumRequired) {
		t.Fatalf("err=%v", err)
	}
	if analytics.calls != 0 || completer.calls != 0 {
		t.Fatal("free user reached analytics or model")
	}
}

func TestInsightsReportMonthlyIncludesHealthWeeklyDoesNot(t *testing.T) {
	monthly := activeAnalytics()
	monthly.Period = "monthly"
	health := &stubHealth{resp: &appFinance.HealthScoreResponse{
		Score:     72,
		YearMonth: "2026-09",
		Components: appFinance.HealthScoreComponents{
			BudgetAdherence: 80,
			Cashflow:        60,
			Coverage:        100,
		},
	}}
	completer := &captureCompleter{reply: `{"summary":"Month looks steady.","points":["Spending is under income."]}`}
	uc, _ := reportUseCase(domainAI.NewCreditBalance(1, false), &stubAnalytics{resp: monthly}, health, completer, alwaysPro{})

	if _, err := uc.Execute(context.Background(), 1, InsightsReportRequest{Period: "monthly"}); err != nil {
		t.Fatal(err)
	}
	if health.calls != 1 {
		t.Fatalf("health calls=%d", health.calls)
	}
	joined := completer.last[1].Content
	if !strings.Contains(joined, `"health_score"`) || !strings.Contains(joined, "72") {
		t.Fatalf("payload=%s", joined)
	}

	weeklyHealth := &stubHealth{resp: health.resp}
	weeklyCompleter := &captureCompleter{reply: completer.reply}
	weeklyUC, _ := reportUseCase(domainAI.NewCreditBalance(1, false), &stubAnalytics{resp: activeAnalytics()}, weeklyHealth, weeklyCompleter, alwaysPro{})
	if _, err := weeklyUC.Execute(context.Background(), 1, InsightsReportRequest{Period: "weekly"}); err != nil {
		t.Fatal(err)
	}
	if weeklyHealth.calls != 0 {
		t.Fatal("weekly report asked for health score")
	}
	if strings.Contains(weeklyCompleter.last[1].Content, "health_score") {
		t.Fatal("weekly payload included health")
	}
}

func TestInsightsReportBlankReplyDoesNotDebit(t *testing.T) {
	analytics := &stubAnalytics{resp: activeAnalytics()}
	completer := &captureCompleter{reply: "```json\n{\"summary\":\"\",\"points\":[]}\n```"}
	uc, repo := reportUseCase(domainAI.NewCreditBalance(1, false), analytics, &stubHealth{}, completer, alwaysPro{})

	_, err := uc.Execute(context.Background(), 1, InsightsReportRequest{Period: "weekly"})
	if !errors.Is(err, ErrInsightsReportBlank) {
		t.Fatalf("err=%v", err)
	}
	if repo.balance.IncludedUsed != 0 {
		t.Fatalf("used=%d", repo.balance.IncludedUsed)
	}
}

func TestInsightsReportStripsMarkdownFence(t *testing.T) {
	summary, points, err := parseInsightsReport("```json\n{\"summary\":\"Ok.\",\"points\":[\"One\",\"Two\",\"Three\",\"Four\"]}\n```")
	if err != nil {
		t.Fatal(err)
	}
	if summary != "Ok." || len(points) != 3 || points[2] != "Three" {
		t.Fatalf("summary=%q points=%v", summary, points)
	}
}

func TestInsightsReportViewCreditsFreeDenied(t *testing.T) {
	uc, _ := reportUseCase(domainAI.NewCreditBalance(1, false), &stubAnalytics{}, &stubHealth{}, &captureCompleter{}, neverPro{})
	_, err := uc.ViewCredits(context.Background(), 1)
	if !errors.Is(err, entitlement.ErrPremiumRequired) {
		t.Fatalf("err=%v", err)
	}
}

func TestPrimaryCurrencySatisfiesInterface(t *testing.T) {
	var _ primaryCurrency = (*appFinance.GetDefaultCurrencyUseCase)(nil)
	var _ periodAnalytics = (*appFinance.GetAnalyticsUseCase)(nil)
	var _ monthlyHealth = (*appFinance.GetHealthScoreUseCase)(nil)
	_ = domainFinance.Currency{}
}
