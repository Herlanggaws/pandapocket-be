package ai

import (
	"context"
	"fmt"
	"strings"
	"time"
	"unicode"

	"panda-pocket/internal/infrastructure/paas"
)

const topicClassifyMaxTokens = 16
const topicClassifyTimeout = 8 * time.Second

const topicClassifySystem = `Classify whether the user message is about THIS user's personal finances in a budgeting app (wallets, cashflow, budgets, categories, goals, debts/assets, net worth, health score, recurring, transactions, transfers) or practical advice grounded in their money data.
Reply with ONLY one token: IN_SCOPE or OUT_OF_SCOPE.
OUT_OF_SCOPE examples: cooking/recipes, coding, weather, entertainment, sports, general trivia, or any topic not about this user's money/data.`

var financeTopicNeedles = []string{
	"budget", "anggaran", "wallet", "dompet", "cashflow", "cash flow",
	"pengeluaran", "pemasukan", "income", "expense", "spending", "spent",
	"transaksi", "transaction", "transfer", "saldo", "balance",
	"goal", "target", "tabungan", "saving", "utang", "hutang", "debt",
	"liabilit", "aset", "asset", "net worth", "kekayaan", "health score",
	"kesehatan keuangan", "recurring", "langganan", "kategori", "category",
	"tagihan", "billing", "cicilan", "mortgage", "kredit", "investasi",
	"berbudget", "keuangan", "uang", "duit", "rp", "idr",
}

var offTopicNeedles = []string{
	"resep", "recipe", "masak", "cooking", "cuaca", "weather",
	"coding", "code ", "programming", "javascript", "python",
	"olahraga", "football", "sepak bola", "nba", "film", "netflix",
	"lagu", "lyrics", "joke", "lelucon", "trivia",
}

// classifyTopic uses a local heuristic when confident; otherwise one short PAAS call.
func classifyTopic(ctx context.Context, completer ChatCompleter, userMessage string) (inScope bool, err error) {
	if inScope, ok := quickTopicScope(userMessage); ok {
		return inScope, nil
	}
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

// quickTopicScope returns (inScope, confident). When confident is false, caller should use PAAS.
func quickTopicScope(userMessage string) (inScope bool, confident bool) {
	normalized := strings.ToLower(strings.TrimSpace(userMessage))
	if normalized == "" {
		return true, false
	}
	compact := stripNonLetters(normalized)

	financeHit := false
	for _, needle := range financeTopicNeedles {
		if strings.Contains(normalized, needle) || strings.Contains(compact, stripNonLetters(needle)) {
			financeHit = true
			break
		}
	}
	offHit := false
	for _, needle := range offTopicNeedles {
		if strings.Contains(normalized, needle) {
			offHit = true
			break
		}
	}

	switch {
	case financeHit && !offHit:
		return true, true
	case offHit && !financeHit:
		return false, true
	default:
		return true, false
	}
}

func stripNonLetters(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
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
