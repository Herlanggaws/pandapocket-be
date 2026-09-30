package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode"

	appFinance "panda-pocket/internal/application/finance"
	domainAI "panda-pocket/internal/domain/ai"
	"panda-pocket/internal/infrastructure/paas"
)

const (
	DraftStatusPending    = "pending"
	DraftStatusSaved      = "saved"
	draftNoteMaxRunes     = 160
	draftExtractMaxTokens = 4096
	draftClassifyTimeout  = 8 * time.Second
	draftHistoryMessages  = 6
	draftHistoryRuneMax   = 300
)

const draftClassifySystem = `Classify whether the user is instructing the app to RECORD one transaction (expense, income, or transfer) versus ASKING about their finances.
Reply with ONLY one token: RECORD or ASK.
RECORD examples: "catat kopi 25rb", "gaji masuk 8jt hari ini", "transfer 100rb dari cash ke bca".
ASK examples: "berapa pengeluaran kopi", "kenapa budget jebol", advice questions.`

const draftExtractSystem = `You extract one transaction the user wants to record in Berbudget.
Return JSON only, no markdown fences:
{"kind":"expense","amount":0,"date":"YYYY-MM-DD","note":"","wallet_name":"","category_name":"","to_wallet_name":""}
kind is expense (money out), income (money in), or transfer (between wallets).
amount is a number. Use 0 when the amount is missing. Do not invent an amount.
date is YYYY-MM-DD when the user stated one, otherwise "".
wallet_name, category_name, and to_wallet_name must be copied exactly from the lists, or "".
One transaction only. If the user mentions several, extract the first.`

type DraftWalletOption struct {
	ID        int
	Name      string
	IsDefault bool
}

type DraftCategoryOption struct {
	ID   int
	Name string
	Kind string
}

type TransactionDraftCatalog struct {
	Wallets    []DraftWalletOption
	Categories []DraftCategoryOption
}

type TransactionDraft struct {
	Kind         string  `json:"kind"`
	Amount       float64 `json:"amount"`
	Date         string  `json:"date"`
	Note         string  `json:"note"`
	WalletID     int     `json:"wallet_id"`
	WalletName   string  `json:"wallet_name"`
	CategoryID   *int    `json:"category_id,omitempty"`
	CategoryName string  `json:"category_name,omitempty"`
	ToWalletID   *int    `json:"to_wallet_id,omitempty"`
	ToWalletName string  `json:"to_wallet_name,omitempty"`
	Status       string  `json:"status"`
}

func CatalogFromFinance(wallets []appFinance.WalletResponse, categories []appFinance.CategoryResponse) TransactionDraftCatalog {
	catalog := TransactionDraftCatalog{}
	for _, wallet := range wallets {
		if wallet.IsArchived {
			continue
		}
		catalog.Wallets = append(catalog.Wallets, DraftWalletOption{
			ID:        wallet.ID,
			Name:      wallet.Name,
			IsDefault: wallet.IsDefault,
		})
	}
	for _, category := range categories {
		if category.Type != "expense" && category.Type != "income" {
			continue
		}
		catalog.Categories = append(catalog.Categories, DraftCategoryOption{
			ID:   category.ID,
			Name: category.Name,
			Kind: category.Type,
		})
	}
	return catalog
}

func (uc *AdvisorChatUseCase) EnableTransactionDrafts(load func(context.Context, int) (TransactionDraftCatalog, error)) {
	uc.loadDraftCatalog = load
}

func (uc *AdvisorChatUseCase) respondIfTransactionDraft(
	bgCtx, finishCtx context.Context,
	userID, threadID int,
	userMessage, lang string,
	history []domainAI.ThreadMessage,
	job *chatJob,
	creditsBefore *CreditsView,
) bool {
	record, err := classifyRecordIntent(bgCtx, uc.paas, userMessage)
	if err != nil || !record {
		return false
	}
	uc.writeTransactionDraft(bgCtx, finishCtx, userID, threadID, lang, history, job, creditsBefore)
	return true
}

func (uc *AdvisorChatUseCase) writeTransactionDraft(
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
		uc.finishDraftAsk(finishCtx, threadID, draftAskCopy(lang, "wallets"), job, creditsBefore)
		return
	}

	reply, err := uc.paas.CompleteChat(bgCtx, draftExtractionMessages(lang, catalog, history), draftExtractMaxTokens)
	if err != nil || strings.TrimSpace(reply) == "" {
		uc.failDraftTurn(finishCtx, threadID, job, creditsBefore)
		return
	}

	draft, miss := resolveTransactionDraft(reply, catalog, draftToday())
	if miss != "" {
		uc.finishDraftAsk(finishCtx, threadID, draftAskCopy(lang, miss), job, creditsBefore)
		return
	}
	raw, err := json.Marshal(draft)
	if err != nil {
		uc.failDraftTurn(finishCtx, threadID, job, creditsBefore)
		return
	}
	text := draftReadyCopy(lang)
	if err := uc.threads.AppendAssistant(finishCtx, threadID, text, string(raw), 0, 0); err != nil {
		uc.failDraftTurn(finishCtx, threadID, job, creditsBefore)
		return
	}
	creditsAfter, spendErr := uc.credits.SpendOne(finishCtx, userID)
	_ = uc.threads.FinishGeneration(finishCtx, threadID, domainAI.GenerationIdle)
	_ = uc.threads.TrimOldest(finishCtx, threadID, domainAI.MaxThreadMsgs)
	if spendErr != nil {
		job.publish(StreamEvent{Kind: StreamError, ErrCode: "AI_CREDITS_REQUIRED", Credits: creditsBefore})
		return
	}
	job.publish(StreamEvent{Kind: StreamDelta, Delta: text})
	job.publish(StreamEvent{Kind: StreamDone, Credits: creditsAfter})
}

func (uc *AdvisorChatUseCase) finishDraftAsk(ctx context.Context, threadID int, text string, job *chatJob, credits *CreditsView) {
	_ = uc.threads.AppendMessage(ctx, threadID, domainAI.RoleAssistant, text, 0, 0)
	_ = uc.threads.FinishGeneration(ctx, threadID, domainAI.GenerationIdle)
	_ = uc.threads.TrimOldest(ctx, threadID, domainAI.MaxThreadMsgs)
	job.publish(StreamEvent{Kind: StreamDelta, Delta: text})
	job.publish(StreamEvent{Kind: StreamDone, Credits: credits})
}

func (uc *AdvisorChatUseCase) failDraftTurn(ctx context.Context, threadID int, job *chatJob, credits *CreditsView) {
	_ = uc.threads.AppendMessage(ctx, threadID, domainAI.RoleAssistant, "(gagal menghasilkan jawaban — coba lagi)", 0, 0)
	_ = uc.threads.FinishGeneration(ctx, threadID, domainAI.GenerationFailed)
	job.publish(StreamEvent{Kind: StreamError, ErrCode: "AI_UPSTREAM_ERROR", Credits: credits})
}

func classifyRecordIntent(ctx context.Context, completer ChatCompleter, userMessage string) (bool, error) {
	if record, confident := quickRecordIntent(userMessage); confident {
		return record, nil
	}
	if completer == nil {
		return false, fmt.Errorf("completer is nil")
	}
	classifyCtx, cancel := context.WithTimeout(ctx, draftClassifyTimeout)
	defer cancel()
	raw, err := completer.CompleteChat(classifyCtx, []paas.Message{
		{Role: "system", Content: draftClassifySystem},
		{Role: "user", Content: userMessage},
	}, topicClassifyMaxTokens)
	if err != nil {
		return false, err
	}
	return parseRecordIntent(raw)
}

func quickRecordIntent(userMessage string) (record bool, confident bool) {
	asked := containsAnyWord(userMessage, recordQuestionWords)
	verb := hasRecordVerb(userMessage)
	transferCommand := strings.Contains(strings.ToLower(userMessage), "transfer") && messageHasAmount(userMessage)
	if verb && !asked {
		return true, true
	}
	if transferCommand && !asked {
		return true, true
	}
	if asked && !verb {
		return false, true
	}
	if messageHasAmount(userMessage) {
		return false, false
	}
	return false, true
}

var recordQuestionWords = []string{
	"kenapa", "mengapa", "bagaimana", "gimana", "berapa", "apakah", "apa",
	"why", "how", "what", "should", "which", "saran", "advice",
}

func hasRecordVerb(userMessage string) bool {
	if containsAnyWord(userMessage, []string{"catat", "dicatat", "catetin", "catatkan", "record"}) {
		return true
	}
	lower := strings.ToLower(userMessage)
	phrases := []string{
		"tambah pengeluaran", "tambah pemasukan", "tambah transaksi",
		"masukkan transaksi", "add expense", "add income", "add transfer",
	}
	for _, phrase := range phrases {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}

func messageHasAmount(userMessage string) bool {
	for _, r := range userMessage {
		if unicode.IsDigit(r) {
			return true
		}
	}
	return containsAnyWord(userMessage, []string{"rb", "ribu", "juta", "jt"})
}

func parseRecordIntent(raw string) (bool, error) {
	fields := strings.Fields(strings.ToUpper(raw))
	if len(fields) == 0 {
		return false, fmt.Errorf("empty record classify")
	}
	switch fields[len(fields)-1] {
	case "RECORD":
		return true, nil
	case "ASK":
		return false, nil
	default:
		return false, fmt.Errorf("unparsed record classify")
	}
}

func draftExtractionMessages(lang string, catalog TransactionDraftCatalog, history []domainAI.ThreadMessage) []paas.Message {
	system := draftExtractSystem
	if lang == "en" {
		system += "\nThe note may stay in the user's language."
	}
	return []paas.Message{
		{Role: "system", Content: system},
		{Role: "user", Content: draftCatalogText(catalog, history)},
	}
}

func draftCatalogText(catalog TransactionDraftCatalog, history []domainAI.ThreadMessage) string {
	var b strings.Builder
	b.WriteString("Today: ")
	b.WriteString(draftToday())
	b.WriteString("\nWallets:\n")
	for _, wallet := range catalog.Wallets {
		b.WriteString("- ")
		b.WriteString(wallet.Name)
		if wallet.IsDefault {
			b.WriteString(" (default)")
		}
		b.WriteByte('\n')
	}
	b.WriteString("Expense categories: ")
	b.WriteString(categoryNames(catalog.Categories, "expense"))
	b.WriteString("\nIncome categories: ")
	b.WriteString(categoryNames(catalog.Categories, "income"))
	b.WriteString("\nConversation:\n")
	b.WriteString(recentDraftConversation(history))
	return b.String()
}

func categoryNames(categories []DraftCategoryOption, kind string) string {
	names := make([]string, 0)
	for _, category := range categories {
		if category.Kind == kind && strings.TrimSpace(category.Name) != "" {
			names = append(names, category.Name)
		}
	}
	if len(names) == 0 {
		return "none"
	}
	return strings.Join(names, ", ")
}

func recentDraftConversation(history []domainAI.ThreadMessage) string {
	start := 0
	if len(history) > draftHistoryMessages {
		start = len(history) - draftHistoryMessages
	}
	var b strings.Builder
	for _, msg := range history[start:] {
		text := strings.TrimSpace(msg.Content)
		if text == "" {
			continue
		}
		runes := []rune(text)
		if len(runes) > draftHistoryRuneMax {
			text = string(runes[:draftHistoryRuneMax])
		}
		b.WriteString(msg.Role)
		b.WriteString(": ")
		b.WriteString(text)
		b.WriteByte('\n')
	}
	return b.String()
}

func resolveTransactionDraft(raw string, catalog TransactionDraftCatalog, today string) (*TransactionDraft, string) {
	parsed, ok := parseDraftJSON(raw)
	if !ok {
		return nil, "parse"
	}
	kind := normalizeDraftKind(parsed.Kind)
	if kind == "" {
		return nil, "parse"
	}
	if parsed.Amount <= 0 {
		return nil, "amount"
	}
	from, to, miss := resolveDraftWallets(kind, parsed.WalletName, parsed.ToWalletName, catalog.Wallets)
	if miss != "" {
		return nil, miss
	}
	draft := &TransactionDraft{
		Kind:       kind,
		Amount:     parsed.Amount,
		Date:       draftDateOrToday(parsed.Date, today),
		Note:       truncateDraftNote(parsed.Note),
		WalletID:   from.ID,
		WalletName: from.Name,
		Status:     DraftStatusPending,
	}
	if kind == "transfer" {
		id := to.ID
		draft.ToWalletID = &id
		draft.ToWalletName = to.Name
		return draft, ""
	}
	if category, matched := matchDraftCategory(parsed.CategoryName, kind, catalog.Categories); matched {
		id := category.ID
		draft.CategoryID = &id
		draft.CategoryName = category.Name
	}
	return draft, ""
}

type parsedDraft struct {
	Kind         string  `json:"kind"`
	Amount       float64 `json:"amount"`
	Date         string  `json:"date"`
	Note         string  `json:"note"`
	WalletName   string  `json:"wallet_name"`
	CategoryName string  `json:"category_name"`
	ToWalletName string  `json:"to_wallet_name"`
}

func parseDraftJSON(raw string) (parsedDraft, bool) {
	text := stripJSONFence(raw)
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start < 0 || end <= start {
		return parsedDraft{}, false
	}
	var parsed parsedDraft
	if err := json.Unmarshal([]byte(text[start:end+1]), &parsed); err != nil {
		return parsedDraft{}, false
	}
	return parsed, true
}

func normalizeDraftKind(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "expense", "pengeluaran":
		return "expense"
	case "income", "pemasukan":
		return "income"
	case "transfer":
		return "transfer"
	default:
		return ""
	}
}

func resolveDraftWallets(kind, fromName, toName string, wallets []DraftWalletOption) (DraftWalletOption, DraftWalletOption, string) {
	if len(wallets) == 0 {
		return DraftWalletOption{}, DraftWalletOption{}, "wallets"
	}
	from, fromOK := pickDraftWallet(fromName, wallets)
	if strings.TrimSpace(fromName) != "" && !fromOK {
		return DraftWalletOption{}, DraftWalletOption{}, "wallet"
	}
	if kind != "transfer" {
		if !fromOK {
			return DraftWalletOption{}, DraftWalletOption{}, "wallet"
		}
		return from, DraftWalletOption{}, ""
	}
	to, toOK := pickNamedWallet(toName, wallets)
	if !toOK {
		return DraftWalletOption{}, DraftWalletOption{}, "transfer"
	}
	if !fromOK {
		from, fromOK = pickDraftWallet("", wallets)
	}
	if !fromOK || from.ID == to.ID {
		return DraftWalletOption{}, DraftWalletOption{}, "transfer"
	}
	return from, to, ""
}

func pickDraftWallet(name string, wallets []DraftWalletOption) (DraftWalletOption, bool) {
	if strings.TrimSpace(name) != "" {
		return pickNamedWallet(name, wallets)
	}
	for _, wallet := range wallets {
		if wallet.IsDefault {
			return wallet, true
		}
	}
	if len(wallets) == 1 {
		return wallets[0], true
	}
	return DraftWalletOption{}, false
}

func pickNamedWallet(name string, wallets []DraftWalletOption) (DraftWalletOption, bool) {
	key := normalizeLabel(name)
	if key == "" {
		return DraftWalletOption{}, false
	}
	for _, wallet := range wallets {
		if normalizeLabel(wallet.Name) == key {
			return wallet, true
		}
	}
	return DraftWalletOption{}, false
}

func matchDraftCategory(name, kind string, categories []DraftCategoryOption) (DraftCategoryOption, bool) {
	key := normalizeLabel(name)
	if key == "" {
		return DraftCategoryOption{}, false
	}
	for _, category := range categories {
		if category.Kind == kind && normalizeLabel(category.Name) == key {
			return category, true
		}
	}
	return DraftCategoryOption{}, false
}

func draftDateOrToday(value, today string) string {
	if validReceiptDate(value) != "" {
		return value
	}
	if validReceiptDate(today) != "" {
		return today
	}
	return time.Now().UTC().Format("2006-01-02")
}

func truncateDraftNote(value string) string {
	note := strings.TrimSpace(value)
	runes := []rune(note)
	if len(runes) <= draftNoteMaxRunes {
		return note
	}
	return string(runes[:draftNoteMaxRunes])
}

func draftToday() string {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*60*60)
	}
	return time.Now().In(loc).Format("2006-01-02")
}

func draftReadyCopy(lang string) string {
	if lang == "en" {
		return "Draft is ready. Check it and save — it is not in your ledger yet."
	}
	return "Draf siap. Cek lalu simpan — belum masuk ke buku transaksi."
}

func draftAskCopy(lang, miss string) string {
	if lang == "en" {
		switch miss {
		case "amount":
			return "The amount is missing. Send the number, for example 25000."
		case "wallet":
			return "The wallet is unclear. Name one of your wallets."
		case "transfer":
			return "A transfer needs two different wallets. Say which wallet it leaves and which it enters."
		case "wallets":
			return "There is no wallet yet. Create one, then send this again."
		default:
			return "Could not prepare one draft. Include the type, amount, and wallet."
		}
	}
	switch miss {
	case "amount":
		return "Jumlahnya belum jelas. Kirim nominalnya, misalnya 25000."
	case "wallet":
		return "Dompetnya belum jelas. Sebutkan nama dompet yang ada."
	case "transfer":
		return "Transfer butuh dua dompet yang berbeda. Sebutkan dari dompet mana ke mana."
	case "wallets":
		return "Belum ada dompet. Buat dompet dulu, lalu kirim lagi."
	default:
		return "Belum bisa menyiapkan satu draf. Sebutkan jenis, nominal, dan dompet."
	}
}
