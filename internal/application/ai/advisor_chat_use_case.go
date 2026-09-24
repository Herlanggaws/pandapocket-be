package ai

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"

	domainAI "panda-pocket/internal/domain/ai"
	"panda-pocket/internal/domain/entitlement"
	"panda-pocket/internal/infrastructure/paas"
)

type ChatStreamer interface {
	Configured() bool
	StreamChat(ctx context.Context, messages []paas.Message, onDelta func(string) error) (string, int, int, error)
	CompleteChat(ctx context.Context, messages []paas.Message, maxTokens int) (string, error)
}

// ChatCompleter is the subset used by the topic gate.
type ChatCompleter interface {
	CompleteChat(ctx context.Context, messages []paas.Message, maxTokens int) (string, error)
}

type StreamEventKind string

const (
	StreamDelta StreamEventKind = "delta"
	StreamDone  StreamEventKind = "done"
	StreamError StreamEventKind = "error"
)

type StreamEvent struct {
	Kind     StreamEventKind
	Delta    string
	Credits  *CreditsView
	ErrCode  string
	ErrMsg   string
}

type chatJob struct {
	mu   sync.Mutex
	subs []chan StreamEvent
}

func (j *chatJob) subscribe() (<-chan StreamEvent, func()) {
	ch := make(chan StreamEvent, 64)
	j.mu.Lock()
	j.subs = append(j.subs, ch)
	j.mu.Unlock()
	unsub := func() {
		j.mu.Lock()
		defer j.mu.Unlock()
		for i, s := range j.subs {
			if s == ch {
				j.subs = append(j.subs[:i], j.subs[i+1:]...)
				break
			}
		}
		close(ch)
	}
	return ch, unsub
}

func (j *chatJob) publish(ev StreamEvent) {
	j.mu.Lock()
	defer j.mu.Unlock()
	for _, ch := range j.subs {
		select {
		case ch <- ev:
		default:
		}
	}
}

type AdvisorChatUseCase struct {
	credits      *CreditService
	threads      domainAI.ThreadRepository
	entitlements entitlement.Checker
	paas         ChatStreamer
	contextDeps  *AdvisorContextDeps
	prefsLang    func(ctx context.Context, userID int) string

	jobsMu sync.Mutex
	jobs   map[int]*chatJob
}

func NewAdvisorChatUseCase(
	credits *CreditService,
	threads domainAI.ThreadRepository,
	entitlements entitlement.Checker,
	paasClient ChatStreamer,
	contextDeps *AdvisorContextDeps,
	prefsLang func(ctx context.Context, userID int) string,
) *AdvisorChatUseCase {
	return &AdvisorChatUseCase{
		credits:      credits,
		threads:      threads,
		entitlements: entitlements,
		paas:         paasClient,
		contextDeps:  contextDeps,
		prefsLang:    prefsLang,
		jobs:         map[int]*chatJob{},
	}
}

func (uc *AdvisorChatUseCase) getOrCreateJob(threadID int) *chatJob {
	uc.jobsMu.Lock()
	defer uc.jobsMu.Unlock()
	if j, ok := uc.jobs[threadID]; ok {
		return j
	}
	j := &chatJob{}
	uc.jobs[threadID] = j
	return j
}

func (uc *AdvisorChatUseCase) removeJob(threadID int) {
	uc.jobsMu.Lock()
	defer uc.jobsMu.Unlock()
	delete(uc.jobs, threadID)
}

// Subscribe attaches to an in-flight job (if any). Caller must unsubscribe.
func (uc *AdvisorChatUseCase) Subscribe(threadID int) (<-chan StreamEvent, func(), bool) {
	uc.jobsMu.Lock()
	j, ok := uc.jobs[threadID]
	uc.jobsMu.Unlock()
	if !ok {
		return nil, func() {}, false
	}
	ch, unsub := j.subscribe()
	return ch, unsub, true
}

// Start validates, marks pending, appends the user message, and runs generation
// on a detached timeout context so client disconnect does not cancel PAAS.
func (uc *AdvisorChatUseCase) Start(
	ctx context.Context,
	userID, threadID int,
	message string,
) (<-chan StreamEvent, func(), error) {
	if err := RequireProAI(ctx, uc.entitlements, userID); err != nil {
		return nil, nil, err
	}
	message = strings.TrimSpace(message)
	if message == "" {
		return nil, nil, fmt.Errorf("message is required")
	}
	if utf8.RuneCountInString(message) > domainAI.MaxMessageLen {
		return nil, nil, domainAI.ErrMessageTooLong
	}
	if !uc.paas.Configured() {
		return nil, nil, domainAI.ErrNotConfigured
	}

	if _, err := uc.threads.FindByIDForUser(ctx, threadID, userID); err != nil {
		return nil, nil, err
	}

	creditsBefore, err := uc.credits.View(ctx, userID)
	if err != nil {
		return nil, nil, err
	}
	if creditsBefore.Available < 1 {
		return nil, nil, domainAI.ErrCreditsRequired
	}

	if err := uc.threads.TryBeginGeneration(ctx, threadID, userID); err != nil {
		return nil, nil, err
	}

	if err := uc.threads.AppendMessage(ctx, threadID, domainAI.RoleUser, message, 0, 0); err != nil {
		_ = uc.threads.FinishGeneration(ctx, threadID, domainAI.GenerationFailed)
		return nil, nil, err
	}
	derivedTitle := domainAI.TruncateTitle(message)
	if err := uc.threads.SetTitleIfEmpty(ctx, threadID, derivedTitle); err != nil {
		// Non-fatal — list/get will backfill from first user message.
		_ = err
	}

	lang := "id"
	if uc.prefsLang != nil {
		if l := uc.prefsLang(ctx, userID); l != "" {
			lang = l
		}
	}

	job := uc.getOrCreateJob(threadID)
	ch, unsub := job.subscribe()

	go uc.runGeneration(userID, threadID, message, lang, job, creditsBefore)

	return ch, unsub, nil
}

func (uc *AdvisorChatUseCase) runGeneration(
	userID, threadID int,
	userMessage, lang string,
	job *chatJob,
	creditsBefore *CreditsView,
) {
	defer uc.removeJob(threadID)

	bgCtx, cancel := context.WithTimeout(context.Background(), domainAI.GenerationTimeout)
	defer cancel()
	finishCtx := context.Background()

	inScope, classifyErr := classifyTopic(bgCtx, uc.paas, userMessage)
	if classifyErr == nil && !inScope {
		refusal := offTopicRefusal(lang)
		_ = uc.threads.AppendMessage(finishCtx, threadID, domainAI.RoleAssistant, refusal, 0, 0)
		_ = uc.threads.FinishGeneration(finishCtx, threadID, domainAI.GenerationIdle)
		_ = uc.threads.TrimOldest(finishCtx, threadID, domainAI.MaxThreadMsgs)
		job.publish(StreamEvent{Kind: StreamDelta, Delta: refusal})
		job.publish(StreamEvent{Kind: StreamDone, Credits: creditsBefore})
		return
	}

	contextJSON := buildAdvisorContextJSON(bgCtx, userID, creditsBefore, uc.contextDeps)
	history, err := uc.threads.ListMessages(bgCtx, threadID)
	if err != nil {
		_ = uc.threads.FinishGeneration(finishCtx, threadID, domainAI.GenerationFailed)
		job.publish(StreamEvent{Kind: StreamError, ErrCode: "AI_CHAT_ERROR", Credits: creditsBefore})
		return
	}

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

	full, promptTokens, completionTokens, err := uc.paas.StreamChat(bgCtx, messages, func(delta string) error {
		job.publish(StreamEvent{Kind: StreamDelta, Delta: delta})
		return nil
	})

	if err != nil || strings.TrimSpace(full) == "" {
		_ = uc.threads.AppendMessage(finishCtx, threadID, domainAI.RoleAssistant, "(gagal menghasilkan jawaban — coba lagi)", 0, 0)
		_ = uc.threads.FinishGeneration(finishCtx, threadID, domainAI.GenerationFailed)
		code := "AI_UPSTREAM_ERROR"
		job.publish(StreamEvent{Kind: StreamError, ErrCode: code, Credits: creditsBefore})
		return
	}

	if appendErr := uc.threads.AppendMessage(finishCtx, threadID, domainAI.RoleAssistant, full, promptTokens, completionTokens); appendErr != nil {
		creditsAfter, spendErr := uc.credits.SpendOne(finishCtx, userID)
		_ = uc.threads.FinishGeneration(finishCtx, threadID, domainAI.GenerationIdle)
		if spendErr != nil {
			job.publish(StreamEvent{Kind: StreamError, ErrCode: "AI_CHAT_ERROR", Credits: creditsBefore})
			return
		}
		_ = uc.threads.TrimOldest(finishCtx, threadID, domainAI.MaxThreadMsgs)
		job.publish(StreamEvent{Kind: StreamDone, Credits: creditsAfter})
		return
	}

	creditsAfter, spendErr := uc.credits.SpendOne(finishCtx, userID)
	_ = uc.threads.FinishGeneration(finishCtx, threadID, domainAI.GenerationIdle)
	if spendErr != nil {
		job.publish(StreamEvent{Kind: StreamError, ErrCode: "AI_CREDITS_REQUIRED", Credits: creditsBefore})
		return
	}
	_ = uc.threads.TrimOldest(finishCtx, threadID, domainAI.MaxThreadMsgs)
	job.publish(StreamEvent{Kind: StreamDone, Credits: creditsAfter})
}

func systemPrompt(lang string) string {
	base := `You are Tanya AI / Ask AI for Berbudget, a personal finance app.
Give practical, non-judgmental advice using ONLY the provided financial context JSON.
The JSON is a full read-only snapshot: primary currency, wallets + balances, cashflow (this month + previous month), top expense categories, budgets, goals, assets, liabilities/debts (including mortgage), net worth, health score, recurring rules, recent transactions, and recent transfers.
If a section is empty, say what is missing and suggest recording it in Berbudget (e.g. [Debts](/debts) for hutang, [Goals](/goals) for target).
You are NOT a licensed financial advisor — include that caveat briefly when giving material advice.
Read-only: never claim you created or changed transactions, budgets, liabilities, goals, or transfers.
Prefer concise answers with clear next steps.
When useful, include markdown links to in-app paths only, e.g. [Budgets](/budgets), [Goals](/goals), [Debts](/debts), [Insights](/insights), [Net worth](/net-worth), [Health](/health), [Transactions](/transactions), [Wallets](/wallets), [Settings billing](/settings/billing).
Do not use external http(s) links.

SCOPE — only the user's personal finances in Berbudget and advice grounded in the provided context JSON.
OUT OF SCOPE — cooking/recipes, coding, weather, entertainment, general trivia, or any topic not about this user's money/data.
If out of scope: refuse in 1–2 short sentences; do not answer the off-topic content; invite a finance question about their Berbudget data.
Never provide recipes, code, or step-by-step for unrelated topics.`
	if lang == "en" {
		return base + "\nRespond in English."
	}
	return base + "\nRespond in Bahasa Indonesia."
}
