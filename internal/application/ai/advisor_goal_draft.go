package ai

import (
	"context"
	"encoding/json"
	"strings"

	domainAI "panda-pocket/internal/domain/ai"
	"panda-pocket/internal/infrastructure/paas"
)

const goalNameMaxRunes = 80

const goalExtractSystem = `You extract one savings goal the user wants to create in Berbudget.
Return JSON only, no markdown fences:
{"name":"","target_amount":0,"target_date":"YYYY-MM-DD","wallet_name":"","category_name":"","note":""}
target_amount is a number. Use 0 when the amount is missing. Do not invent an amount.
target_date is YYYY-MM-DD. When the user says next year without a month, use July 1 of the next calendar year relative to Today. When they give no date at all, use "".
wallet_name and category_name must be copied exactly from the lists, or "".
category_name is an expense category for related spending. It is not stored on the goal.
note is a short remark, such as a later monthly fee. One goal only. If the user mentions a lump sum and a monthly fee, extract the lump sum and put the monthly fee in note.`

type GoalDraft struct {
	Kind          string  `json:"kind"`
	Name          string  `json:"name"`
	TargetAmount  float64 `json:"target_amount"`
	TargetDate    string  `json:"target_date"`
	CurrentAmount float64 `json:"current_amount"`
	WalletID      int     `json:"wallet_id"`
	WalletName    string  `json:"wallet_name"`
	CategoryID    *int    `json:"category_id"`
	CategoryName  string  `json:"category_name,omitempty"`
	Note          string  `json:"note,omitempty"`
	Status        string  `json:"status"`
}

type parsedGoalDraft struct {
	Name         string  `json:"name"`
	TargetAmount float64 `json:"target_amount"`
	TargetDate   string  `json:"target_date"`
	WalletName   string  `json:"wallet_name"`
	CategoryName string  `json:"category_name"`
	Note         string  `json:"note"`
}

func (uc *AdvisorChatUseCase) writeGoalDraft(
	bgCtx, finishCtx context.Context,
	userID, threadID int,
	lang string,
	history []domainAI.ThreadMessage,
	job *chatJob,
	creditsBefore *CreditsView,
) {
	catalog, err := uc.loadDraftCatalog(bgCtx, userID)
	if err != nil || len(catalog.Wallets) == 0 {
		if err != nil {
			uc.failDraftTurn(finishCtx, threadID, job, creditsBefore)
			return
		}
		uc.finishDraftAsk(finishCtx, threadID, draftAskCopy(lang, "wallets"), lang, job, creditsBefore)
		return
	}

	reply, err := uc.paas.CompleteChat(bgCtx, goalExtractionMessages(lang, catalog, history), draftExtractMaxTokens)
	if err != nil || strings.TrimSpace(reply) == "" {
		uc.failDraftTurn(finishCtx, threadID, job, creditsBefore)
		return
	}

	draft, miss := resolveGoalDraft(reply, catalog, draftToday())
	if miss != "" {
		uc.finishDraftAsk(finishCtx, threadID, goalAskCopy(lang, miss), lang, job, creditsBefore)
		return
	}
	raw, err := json.Marshal(draft)
	if err != nil {
		uc.failDraftTurn(finishCtx, threadID, job, creditsBefore)
		return
	}
	text := goalReadyCopy(lang)
	if err := uc.threads.AppendAssistant(finishCtx, threadID, text, string(raw), 0, 0); err != nil {
		uc.failDraftTurn(finishCtx, threadID, job, creditsBefore)
		return
	}
	creditsAfter, spendErr := uc.credits.SpendOne(finishCtx, userID, domainAI.SpendSourceAdvisorDraft)
	_ = uc.threads.FinishGeneration(finishCtx, threadID, domainAI.GenerationIdle)
	_ = uc.threads.TrimOldest(finishCtx, threadID, domainAI.MaxThreadMsgs)
	if spendErr != nil {
		job.publish(StreamEvent{Kind: StreamError, ErrCode: "AI_CREDITS_REQUIRED", Credits: creditsBefore})
		return
	}
	uc.attachFallbackFollowUps(finishCtx, threadID, lang, true)
	job.publish(StreamEvent{Kind: StreamDelta, Delta: text})
	job.publish(StreamEvent{Kind: StreamDone, Credits: creditsAfter})
}

func goalExtractionMessages(lang string, catalog TransactionDraftCatalog, history []domainAI.ThreadMessage) []paas.Message {
	system := goalExtractSystem
	if lang == "en" {
		system += "\nThe name and note may stay in the user's language."
	}
	return []paas.Message{
		{Role: "system", Content: system},
		{Role: "user", Content: draftCatalogText(catalog, history)},
	}
}

func resolveGoalDraft(raw string, catalog TransactionDraftCatalog, today string) (*GoalDraft, string) {
	parsed, ok := parseGoalJSON(raw)
	if !ok {
		return nil, "parse"
	}
	if parsed.TargetAmount <= 0 {
		return nil, "amount"
	}
	targetDate, ok := futureGoalDate(parsed.TargetDate, today)
	if !ok {
		return nil, "date"
	}
	wallet, miss := resolveGoalWallet(parsed.WalletName, catalog.Wallets)
	if miss != "" {
		return nil, miss
	}
	name := goalDisplayName(parsed.Name, parsed.Note)
	if name == "" {
		return nil, "name"
	}
	draft := &GoalDraft{
		Kind:         "goal",
		Name:         name,
		TargetAmount: parsed.TargetAmount,
		TargetDate:   targetDate,
		WalletID:     wallet.ID,
		WalletName:   wallet.Name,
		Note:         truncateDraftNote(parsed.Note),
		Status:       DraftStatusPending,
	}
	applyGoalCategory(draft, parsed.CategoryName, catalog.Categories)
	return draft, ""
}

func parseGoalJSON(raw string) (parsedGoalDraft, bool) {
	text := stripJSONFence(raw)
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start < 0 || end <= start {
		return parsedGoalDraft{}, false
	}
	var parsed parsedGoalDraft
	if err := json.Unmarshal([]byte(text[start:end+1]), &parsed); err != nil {
		return parsedGoalDraft{}, false
	}
	return parsed, true
}

func futureGoalDate(value, today string) (string, bool) {
	if validReceiptDate(value) == "" || validReceiptDate(today) == "" {
		return "", false
	}
	if value <= today {
		return "", false
	}
	return value, true
}

func resolveGoalWallet(name string, wallets []DraftWalletOption) (DraftWalletOption, string) {
	if len(wallets) == 0 {
		return DraftWalletOption{}, "wallets"
	}
	wallet, ok := pickDraftWallet(name, wallets)
	if strings.TrimSpace(name) != "" && !ok {
		return DraftWalletOption{}, "wallet"
	}
	if !ok {
		return DraftWalletOption{}, "wallet"
	}
	return wallet, ""
}

func goalDisplayName(name, note string) string {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		trimmed = strings.TrimSpace(note)
	}
	runes := []rune(trimmed)
	if len(runes) > goalNameMaxRunes {
		return string(runes[:goalNameMaxRunes])
	}
	return trimmed
}

func applyGoalCategory(draft *GoalDraft, name string, categories []DraftCategoryOption) {
	label := strings.TrimSpace(name)
	if label == "" {
		return
	}
	if category, matched := matchDraftCategory(label, "expense", categories); matched {
		id := category.ID
		draft.CategoryID = &id
		draft.CategoryName = category.Name
		return
	}
	draft.CategoryName = label
}

func goalReadyCopy(lang string) string {
	if lang == "en" {
		return "Goal suggestion is ready. Check it and save — it is not in your goals yet."
	}
	return "Usulan target siap. Cek lalu simpan — belum masuk ke daftar target."
}

func goalAskCopy(lang, miss string) string {
	if lang == "en" {
		switch miss {
		case "amount":
			return "The goal amount is missing. Send the number, for example 30000000."
		case "date":
			return "The goal date is missing or already past. Send a future date."
		case "name":
			return "The goal name is missing. Say what you are saving for."
		case "wallet":
			return "The wallet is unclear. Name one of your wallets."
		case "wallets":
			return "There is no wallet yet. Create one, then send this again."
		default:
			return "Could not prepare one goal. Include the name, amount, and a future date."
		}
	}
	switch miss {
	case "amount":
		return "Nominal targetnya belum jelas. Kirim angkanya, misalnya 30000000."
	case "date":
		return "Tanggal targetnya belum jelas atau sudah lewat. Kirim tanggal di masa depan."
	case "name":
		return "Nama targetnya belum jelas. Sebutkan untuk apa tabungannya."
	case "wallet":
		return "Dompetnya belum jelas. Sebutkan nama dompet yang ada."
	case "wallets":
		return "Belum ada dompet. Buat dompet dulu, lalu kirim lagi."
	default:
		return "Belum bisa menyiapkan satu target. Sebutkan nama, nominal, dan tanggal di masa depan."
	}
}
