package ai

import (
	"context"
	"encoding/json"
	"log"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	domainAI "panda-pocket/internal/domain/ai"
	"panda-pocket/internal/infrastructure/paas"
)

const (
	replyMetaMarker    = "reply_meta_v1"
	replyMetaMaxTokens = 512
	replyMetaTimeout   = 12 * time.Second
	maxFollowUps       = 3
	maxReplyActions    = 2
	maxFollowUpRunes   = 160
	maxBudgetCards     = 3
)

var replyActionHrefs = map[string]string{
	"open_health":       "/health",
	"open_budgets":      "/budgets",
	"open_transactions": "/transactions",
	"open_wallets":      "/settings/wallets",
	"open_goals":        "/goals",
	"open_insights":     "/insights",
	"open_debts":        "/debts",
}

var allowedReplySources = map[string]struct{}{
	"cashflow": {}, "budgets": {}, "health": {},
	"transactions": {}, "wallets": {}, "goals": {}, "debts": {},
}

const replyMetaSystem = replyMetaMarker + `
You choose which slices of the user's Berbudget data the answer relied on.
Reply with JSON only, no markdown:
{"sources":["cashflow"],"follow_ups":["..."],"actions":["open_budgets"]}
sources: zero or more of cashflow, budgets, health, transactions, wallets, goals, debts.
follow_ups: exactly 3 short next questions in the same language as the answer, specific to the advice just given.
actions: 0 to 2 of open_health, open_budgets, open_transactions, open_wallets, open_goals, open_insights, open_debts.
Do not invent numbers.`

type replyMetaModel struct {
	Sources   []string `json:"sources"`
	FollowUps []string `json:"follow_ups"`
	Actions   []string `json:"actions"`
}

type storedReplyMeta struct {
	Sources     []string          `json:"sources"`
	Cards       []json.RawMessage `json:"cards"`
	FollowUps   []string          `json:"follow_ups"`
	Actions     []storedAction    `json:"actions"`
	CreditSpent bool              `json:"credit_spent"`
	Feedback    string            `json:"feedback,omitempty"`
}

type storedAction struct {
	ID   string `json:"id"`
	Href string `json:"href"`
}

type snapshotCurrency struct {
	Symbol string `json:"symbol"`
}

type snapshotHealth struct {
	Score      int `json:"score"`
	Components struct {
		Cashflow float64 `json:"cashflow"`
	} `json:"components"`
}

type snapshotBudget struct {
	Category    string  `json:"category"`
	Period      string  `json:"period"`
	PercentUsed float64 `json:"percent_used"`
}

type advisorSnapshot struct {
	PrimaryCurrency *snapshotCurrency `json:"primary_currency"`
	CashflowMonthly *compactCashflow  `json:"cashflow_monthly"`
	CashflowPrev    *compactCashflow  `json:"cashflow_previous_month"`
	Health          *snapshotHealth   `json:"health"`
	Budgets         []snapshotBudget  `json:"budgets"`
}

func (uc *AdvisorChatUseCase) attachReplyMeta(
	ctx context.Context,
	threadID int,
	userMessage, answer, lang, contextJSON string,
) {
	if uc.paas == nil {
		return
	}
	metaCtx, cancel := context.WithTimeout(ctx, replyMetaTimeout)
	defer cancel()
	raw, err := uc.paas.CompleteChat(metaCtx, replyMetaMessages(lang, userMessage, answer), replyMetaMaxTokens)
	if err != nil {
		log.Printf("AI advisor reply meta skipped thread=%d: %v", threadID, err)
		return
	}
	encoded, ok := buildReplyMeta(raw, contextJSON, true)
	if !ok {
		log.Printf("AI advisor reply meta unusable thread=%d", threadID)
		return
	}
	messageID, err := latestAssistantID(ctx, uc.threads, threadID)
	if err != nil {
		log.Printf("AI advisor reply meta message thread=%d: %v", threadID, err)
		return
	}
	if err := uc.threads.UpdateMessageReplyMeta(ctx, threadID, messageID, encoded); err != nil {
		log.Printf("AI advisor reply meta save thread=%d: %v", threadID, err)
	}
}

func replyMetaMessages(lang, userMessage, answer string) []paas.Message {
	system := replyMetaSystem
	if lang == "en" {
		system += "\nWrite follow_ups in English."
	} else {
		system += "\nWrite follow_ups in Bahasa Indonesia."
	}
	return []paas.Message{
		{Role: "system", Content: system},
		{Role: "user", Content: "Question:\n" + clipRunes(userMessage, 500) + "\n\nAnswer:\n" + clipRunes(answer, 2000)},
	}
}

func latestAssistantID(ctx context.Context, threads domainAI.ThreadRepository, threadID int) (int, error) {
	messages, err := threads.ListMessages(ctx, threadID)
	if err != nil {
		return 0, err
	}
	for i := len(messages) - 1; i >= 0; i-- {
		if messages[i].Role == domainAI.RoleAssistant && messages[i].ID > 0 {
			return messages[i].ID, nil
		}
	}
	return 0, domainAI.ErrMessageNotFound
}

func buildReplyMeta(modelRaw, contextJSON string, creditSpent bool) (string, bool) {
	parsed, ok := parseReplyMetaModel(modelRaw)
	if !ok {
		return "", false
	}
	meta := storedReplyMeta{
		Sources:     filterSources(parsed.Sources),
		Cards:       cardsForSources(parsed.Sources, contextJSON),
		FollowUps:   clipFollowUps(parsed.FollowUps),
		Actions:     filterActions(parsed.Actions),
		CreditSpent: creditSpent,
	}
	raw, err := json.Marshal(meta)
	if err != nil {
		return "", false
	}
	return string(raw), true
}

func parseReplyMetaModel(raw string) (replyMetaModel, bool) {
	body := extractJSONObject(raw)
	if body == "" {
		return replyMetaModel{}, false
	}
	var parsed replyMetaModel
	if err := json.Unmarshal([]byte(body), &parsed); err != nil {
		return replyMetaModel{}, false
	}
	return parsed, true
}

func extractJSONObject(raw string) string {
	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start < 0 || end <= start {
		return ""
	}
	return raw[start : end+1]
}

func filterSources(sources []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(sources))
	for _, source := range sources {
		key := strings.ToLower(strings.TrimSpace(source))
		if _, allowed := allowedReplySources[key]; !allowed {
			continue
		}
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, key)
	}
	return out
}

func clipFollowUps(lines []string) []string {
	out := make([]string, 0, maxFollowUps)
	for _, line := range lines {
		text := strings.TrimSpace(line)
		if text == "" {
			continue
		}
		text = clipRunes(text, maxFollowUpRunes)
		out = append(out, text)
		if len(out) == maxFollowUps {
			break
		}
	}
	if len(out) < maxFollowUps {
		return []string{}
	}
	return out
}

func filterActions(ids []string) []storedAction {
	seen := map[string]struct{}{}
	out := make([]storedAction, 0, maxReplyActions)
	for _, id := range ids {
		key := strings.ToLower(strings.TrimSpace(id))
		href, ok := replyActionHrefs[key]
		if !ok {
			continue
		}
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, storedAction{ID: key, Href: href})
		if len(out) == maxReplyActions {
			break
		}
	}
	return out
}

func cardsForSources(sources []string, contextJSON string) []json.RawMessage {
	allowed := map[string]struct{}{}
	for _, source := range filterSources(sources) {
		allowed[source] = struct{}{}
	}
	var snap advisorSnapshot
	if err := json.Unmarshal([]byte(contextJSON), &snap); err != nil {
		return []json.RawMessage{}
	}
	symbol := "Rp"
	if snap.PrimaryCurrency != nil && strings.TrimSpace(snap.PrimaryCurrency.Symbol) != "" {
		symbol = snap.PrimaryCurrency.Symbol
	}
	month := advisorMonthKey(time.Now())
	cards := make([]json.RawMessage, 0, 3)
	if _, ok := allowed["health"]; ok {
		if card, ok := healthCard(snap, symbol, month); ok {
			cards = append(cards, card)
		}
	}
	if _, ok := allowed["cashflow"]; ok {
		if card, ok := cashflowCard(snap, symbol, month); ok {
			cards = append(cards, card)
		}
	}
	if _, ok := allowed["budgets"]; ok {
		if card, ok := budgetCard(snap); ok {
			cards = append(cards, card)
		}
	}
	return cards
}

func healthCard(snap advisorSnapshot, symbol, month string) (json.RawMessage, bool) {
	if snap.Health == nil {
		return nil, false
	}
	card := map[string]any{
		"kind":            "health",
		"score":           snap.Health.Score,
		"cashflow_score":  snap.Health.Components.Cashflow,
		"year_month":      month,
		"currency_symbol": symbol,
	}
	if snap.CashflowMonthly != nil {
		card["tone"] = cashflowTone(snap.CashflowMonthly.NetAmount)
	}
	if snap.CashflowPrev != nil {
		card["previous_net"] = snap.CashflowPrev.NetAmount
	}
	return marshalCard(card)
}

func cashflowCard(snap advisorSnapshot, symbol, month string) (json.RawMessage, bool) {
	if snap.CashflowMonthly == nil {
		return nil, false
	}
	flow := snap.CashflowMonthly
	return marshalCard(map[string]any{
		"kind":            "cashflow",
		"income":          flow.TotalIncome,
		"expense":         flow.TotalSpent,
		"net":             flow.NetAmount,
		"income_missing":  flow.TotalIncome == 0,
		"year_month":      month,
		"currency_symbol": symbol,
	})
}

func budgetCard(snap advisorSnapshot) (json.RawMessage, bool) {
	if len(snap.Budgets) == 0 {
		return nil, false
	}
	rows := append([]snapshotBudget(nil), snap.Budgets...)
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i].PercentUsed > rows[j].PercentUsed
	})
	if len(rows) > maxBudgetCards {
		rows = rows[:maxBudgetCards]
	}
	items := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		items = append(items, map[string]any{
			"category":     row.Category,
			"percent_used": row.PercentUsed,
			"period":       row.Period,
		})
	}
	return marshalCard(map[string]any{
		"kind":  "budgets",
		"items": items,
	})
}

func cashflowTone(net float64) string {
	if net < 0 {
		return "deficit"
	}
	return "surplus"
}

func marshalCard(card map[string]any) (json.RawMessage, bool) {
	raw, err := json.Marshal(card)
	if err != nil {
		return nil, false
	}
	return raw, true
}

func advisorMonthKey(now time.Time) string {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return now.UTC().Format("2006-01")
	}
	return now.In(loc).Format("2006-01")
}

func clipRunes(text string, limit int) string {
	if utf8.RuneCountInString(text) <= limit {
		return text
	}
	runes := []rune(text)
	return string(runes[:limit])
}

func mergeFeedback(existing, value string) (string, error) {
	switch value {
	case "up", "down", "clear":
	default:
		return "", domainAI.ErrInvalidFeedback
	}
	meta := storedReplyMeta{
		Sources:   []string{},
		Cards:     []json.RawMessage{},
		FollowUps: []string{},
		Actions:   []storedAction{},
	}
	if strings.TrimSpace(existing) != "" {
		if err := json.Unmarshal([]byte(existing), &meta); err != nil {
			meta = storedReplyMeta{}
		}
	}
	if value == "clear" {
		meta.Feedback = ""
	} else {
		meta.Feedback = value
	}
	if meta.Sources == nil {
		meta.Sources = []string{}
	}
	if meta.Cards == nil {
		meta.Cards = []json.RawMessage{}
	}
	if meta.FollowUps == nil {
		meta.FollowUps = []string{}
	}
	if meta.Actions == nil {
		meta.Actions = []storedAction{}
	}
	raw, err := json.Marshal(meta)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}
