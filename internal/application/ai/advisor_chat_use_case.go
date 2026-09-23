package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	domainAI "panda-pocket/internal/domain/ai"
	"panda-pocket/internal/domain/entitlement"
	"panda-pocket/internal/infrastructure/paas"

	appFinance "panda-pocket/internal/application/finance"
)

type ChatStreamer interface {
	Configured() bool
	StreamChat(ctx context.Context, messages []paas.Message, onDelta func(string) error) (string, int, int, error)
}

type AdvisorChatUseCase struct {
	credits      *CreditService
	threads      domainAI.ThreadRepository
	entitlements entitlement.Checker
	paas         ChatStreamer
	analytics    *appFinance.GetAnalyticsUseCase
	prefsLang    func(ctx context.Context, userID int) string
}

func NewAdvisorChatUseCase(
	credits *CreditService,
	threads domainAI.ThreadRepository,
	entitlements entitlement.Checker,
	paasClient ChatStreamer,
	analytics *appFinance.GetAnalyticsUseCase,
	prefsLang func(ctx context.Context, userID int) string,
) *AdvisorChatUseCase {
	return &AdvisorChatUseCase{
		credits:      credits,
		threads:      threads,
		entitlements: entitlements,
		paas:         paasClient,
		analytics:    analytics,
		prefsLang:    prefsLang,
	}
}

func (uc *AdvisorChatUseCase) Execute(
	ctx context.Context,
	userID int,
	message string,
	onDelta func(string) error,
) (*CreditsView, error) {
	if err := RequireProAI(ctx, uc.entitlements, userID); err != nil {
		return nil, err
	}
	message = strings.TrimSpace(message)
	if message == "" {
		return nil, fmt.Errorf("message is required")
	}
	if utf8.RuneCountInString(message) > domainAI.MaxMessageLen {
		return nil, domainAI.ErrMessageTooLong
	}
	if !uc.paas.Configured() {
		return nil, domainAI.ErrNotConfigured
	}

	credits, err := uc.credits.SpendOne(ctx, userID)
	if err != nil {
		return nil, err
	}

	threadID, err := uc.threads.GetOrCreateThreadID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if err := uc.threads.AppendMessage(ctx, threadID, domainAI.RoleUser, message, 0, 0); err != nil {
		return nil, err
	}

	history, err := uc.threads.ListMessages(ctx, threadID)
	if err != nil {
		return nil, err
	}

	lang := "id"
	if uc.prefsLang != nil {
		if l := uc.prefsLang(ctx, userID); l != "" {
			lang = l
		}
	}
	contextJSON := uc.buildContextJSON(ctx, userID, credits)

	messages := []paas.Message{
		{Role: "system", Content: systemPrompt(lang)},
		{Role: "system", Content: "User financial context (JSON):\n" + contextJSON},
	}
	for _, msg := range history {
		if msg.Role != domainAI.RoleUser && msg.Role != domainAI.RoleAssistant {
			continue
		}
		messages = append(messages, paas.Message{Role: msg.Role, Content: msg.Content})
	}

	full, promptTokens, completionTokens, err := uc.paas.StreamChat(ctx, messages, onDelta)
	if err != nil {
		_ = uc.threads.AppendMessage(ctx, threadID, domainAI.RoleAssistant, "(gagal menghasilkan jawaban — coba lagi)", 0, 0)
		return credits, fmt.Errorf("%w: %v", domainAI.ErrUpstream, err)
	}
	if strings.TrimSpace(full) == "" {
		full = "Maaf, saya tidak bisa menjawab sekarang. Coba lagi sebentar."
	}
	if err := uc.threads.AppendMessage(ctx, threadID, domainAI.RoleAssistant, full, promptTokens, completionTokens); err != nil {
		return credits, err
	}
	_ = uc.threads.TrimOldest(ctx, threadID, domainAI.MaxThreadMsgs)
	return credits, nil
}

func systemPrompt(lang string) string {
	base := `You are Tanya AI / Ask AI for Berbudget, a personal finance app. 
Give practical, non-judgmental advice using ONLY the provided financial context JSON.
If data is missing, say what is missing and suggest recording it in Berbudget.
You are NOT a licensed financial advisor — include that caveat briefly when giving material advice.
Read-only: never claim you created or changed transactions, budgets, or transfers.
Prefer concise answers with clear next steps.
When useful, include markdown links to in-app paths only, e.g. [Budgets](/budgets), [Goals](/goals), [Debts](/debts), [Insights](/insights), [Net worth](/net-worth), [Health](/health), [Transactions](/transactions), [Settings billing](/settings/billing).
Do not use external http(s) links.`
	if lang == "en" {
		return base + "\nRespond in English."
	}
	return base + "\nRespond in Bahasa Indonesia."
}

func (uc *AdvisorChatUseCase) buildContextJSON(ctx context.Context, userID int, credits *CreditsView) string {
	payload := map[string]interface{}{
		"generated_at": time.Now().UTC().Format(time.RFC3339),
		"credits":      credits,
	}
	if uc.analytics != nil {
		monthly, err := uc.analytics.Execute(ctx, userID, appFinance.GetAnalyticsRequest{Period: "monthly"})
		if err == nil && monthly != nil {
			payload["cashflow_monthly"] = monthly
		}
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return `{}`
	}
	return string(raw)
}
