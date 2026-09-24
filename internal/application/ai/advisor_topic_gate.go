package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"panda-pocket/internal/infrastructure/paas"
)

const topicClassifyMaxTokens = 16
const topicClassifyTimeout = 15 * time.Second

const topicClassifySystem = `Classify whether the user message is about THIS user's personal finances in a budgeting app (wallets, cashflow, budgets, categories, goals, debts/assets, net worth, health score, recurring, transactions, transfers) or practical advice grounded in their money data.
Reply with ONLY one token: IN_SCOPE or OUT_OF_SCOPE.
OUT_OF_SCOPE examples: cooking/recipes, coding, weather, entertainment, sports, general trivia, or any topic not about this user's money/data.`

func classifyTopic(ctx context.Context, completer ChatCompleter, userMessage string) (inScope bool, err error) {
	if completer == nil {
		return true, fmt.Errorf("completer is nil")
	}
	classifyCtx, cancel := context.WithTimeout(ctx, topicClassifyTimeout)
	defer cancel()

	raw, err := completer.CompleteChat(classifyCtx, []paas.Message{
		{Role: "system", Content: topicClassifySystem},
		{Role: "user", Content: userMessage},
	}, topicClassifyMaxTokens)
	if err != nil {
		return true, err
	}
	return parseTopicScope(raw)
}

func parseTopicScope(raw string) (inScope bool, err error) {
	normalized := strings.ToUpper(strings.TrimSpace(raw))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, " ", "_")

	hasIn := strings.Contains(normalized, "IN_SCOPE")
	hasOut := strings.Contains(normalized, "OUT_OF_SCOPE")
	switch {
	case hasOut && !hasIn:
		return false, nil
	case hasIn && !hasOut:
		return true, nil
	case hasOut && hasIn:
		// Prefer the more specific OUT token if both appear (e.g. "not IN_SCOPE → OUT_OF_SCOPE").
		if strings.LastIndex(normalized, "OUT_OF_SCOPE") > strings.LastIndex(normalized, "IN_SCOPE") {
			return false, nil
		}
		return true, nil
	default:
		return true, fmt.Errorf("unrecognized topic label: %q", raw)
	}
}

func offTopicRefusal(lang string) string {
	if lang == "en" {
		return "I can only help with your personal finances in Berbudget — budgets, spending, goals, debts, wallets, and similar. Ask something about your money data, for example your [Budgets](/budgets) or [Goals](/goals)."
	}
	return "Saya hanya bisa membantu soal keuanganmu di Berbudget — anggaran, pengeluaran, target, utang, dompet, dan sejenisnya. Tanya tentang data keuanganmu, misalnya [Anggaran](/budgets) atau [Goals](/goals)."
}
