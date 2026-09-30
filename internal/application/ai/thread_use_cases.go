package ai

import (
	"context"
	"encoding/json"
	"strings"
	"time"
	"unicode/utf8"

	domainAI "panda-pocket/internal/domain/ai"
	"panda-pocket/internal/domain/entitlement"
)

type ThreadMessageDTO struct {
	ID        int               `json:"id"`
	Role      string            `json:"role"`
	Content   string            `json:"content"`
	CreatedAt time.Time         `json:"created_at"`
	Draft     *TransactionDraft `json:"draft,omitempty"`
}

type ThreadSummaryDTO struct {
	ID               int       `json:"id"`
	Title            string    `json:"title"`
	UpdatedAt        time.Time `json:"updated_at"`
	CreatedAt        time.Time `json:"created_at"`
	GenerationStatus string    `json:"generation_status"`
}

type GetThreadResponse struct {
	ID               int                `json:"id"`
	Title            string             `json:"title"`
	GenerationStatus string             `json:"generation_status"`
	Messages         []ThreadMessageDTO `json:"messages"`
	Credits          *CreditsView       `json:"credits"`
}

type ThreadUseCases struct {
	credits      *CreditService
	threads      domainAI.ThreadRepository
	entitlements entitlement.Checker
}

func NewThreadUseCases(credits *CreditService, threads domainAI.ThreadRepository, entitlements entitlement.Checker) *ThreadUseCases {
	return &ThreadUseCases{credits: credits, threads: threads, entitlements: entitlements}
}

func (uc *ThreadUseCases) List(ctx context.Context, userID int) ([]ThreadSummaryDTO, error) {
	if err := RequireProAI(ctx, uc.entitlements, userID); err != nil {
		return nil, err
	}
	rows, err := uc.threads.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]ThreadSummaryDTO, 0, len(rows))
	for _, t := range rows {
		title := uc.resolveTitle(ctx, t)
		status := t.GenerationStatus
		if status == "" {
			status = domainAI.GenerationIdle
		}
		out = append(out, ThreadSummaryDTO{
			ID:               t.ID,
			Title:            title,
			UpdatedAt:        t.UpdatedAt,
			CreatedAt:        t.CreatedAt,
			GenerationStatus: status,
		})
	}
	return out, nil
}

func (uc *ThreadUseCases) resolveTitle(ctx context.Context, t domainAI.Thread) string {
	title := strings.TrimSpace(t.Title)
	if title != "" {
		return title
	}
	content, err := uc.threads.FirstUserMessageContent(ctx, t.ID)
	if err == nil && strings.TrimSpace(content) != "" {
		title = domainAI.TruncateTitle(content)
		_ = uc.threads.SetTitleIfEmpty(ctx, t.ID, title)
		return title
	}
	return "New chat"
}

func (uc *ThreadUseCases) Create(ctx context.Context, userID int) (*ThreadSummaryDTO, error) {
	if err := RequireProAI(ctx, uc.entitlements, userID); err != nil {
		return nil, err
	}
	t, err := uc.threads.Create(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &ThreadSummaryDTO{
		ID:               t.ID,
		Title:            "New chat",
		UpdatedAt:        t.UpdatedAt,
		CreatedAt:        t.CreatedAt,
		GenerationStatus: domainAI.GenerationIdle,
	}, nil
}

func (uc *ThreadUseCases) Get(ctx context.Context, userID, threadID int) (*GetThreadResponse, error) {
	if err := RequireProAI(ctx, uc.entitlements, userID); err != nil {
		return nil, err
	}
	thread, err := uc.threads.FindByIDForUser(ctx, threadID, userID)
	if err != nil {
		return nil, err
	}
	thread, err = uc.settleIfAbandoned(ctx, thread, domainAI.GenerationTimeout)
	if err != nil {
		return nil, err
	}
	credits, err := uc.credits.View(ctx, userID)
	if err != nil {
		return nil, err
	}
	msgs, err := uc.threads.ListMessages(ctx, threadID)
	if err != nil {
		return nil, err
	}
	out := make([]ThreadMessageDTO, 0, len(msgs))
	for _, m := range msgs {
		if m.Role != domainAI.RoleUser && m.Role != domainAI.RoleAssistant {
			continue
		}
		out = append(out, threadMessageDTO(m))
	}
	title := uc.resolveTitle(ctx, *thread)
	status := thread.GenerationStatus
	if status == "" {
		status = domainAI.GenerationIdle
	}
	return &GetThreadResponse{
		ID:               thread.ID,
		Title:            title,
		GenerationStatus: status,
		Messages:         out,
		Credits:          credits,
	}, nil
}

// RecoverAbandonedGenerations clears every pending turn. Call it on process
// start: in-memory chat jobs died with the previous process, so a pending row
// would lock the composer until the client gave up.
func (uc *ThreadUseCases) RecoverAbandonedGenerations(ctx context.Context) (int, error) {
	pending, err := uc.threads.ListPendingGenerations(ctx)
	if err != nil {
		return 0, err
	}
	released := 0
	for i := range pending {
		if _, err := uc.abandonGeneration(ctx, &pending[i]); err != nil {
			return released, err
		}
		released++
	}
	return released, nil
}

func (uc *ThreadUseCases) settleIfAbandoned(ctx context.Context, thread *domainAI.Thread, maxAge time.Duration) (*domainAI.Thread, error) {
	if !generationAbandoned(thread, maxAge) {
		return thread, nil
	}
	return uc.abandonGeneration(ctx, thread)
}

func generationAbandoned(thread *domainAI.Thread, maxAge time.Duration) bool {
	if thread.GenerationStatus != domainAI.GenerationPending {
		return false
	}
	if thread.GenerationStartedAt == nil {
		return true
	}
	return time.Since(*thread.GenerationStartedAt) >= maxAge
}

func (uc *ThreadUseCases) abandonGeneration(ctx context.Context, thread *domainAI.Thread) (*domainAI.Thread, error) {
	if err := uc.appendFailureIfUnanswered(ctx, thread.ID); err != nil {
		return nil, err
	}
	if err := uc.threads.FinishGeneration(ctx, thread.ID, domainAI.GenerationFailed); err != nil {
		return nil, err
	}
	return uc.threads.FindByIDForUser(ctx, thread.ID, thread.UserID)
}

func (uc *ThreadUseCases) appendFailureIfUnanswered(ctx context.Context, threadID int) error {
	messages, err := uc.threads.ListMessages(ctx, threadID)
	if err != nil {
		return err
	}
	if len(messages) == 0 || messages[len(messages)-1].Role != domainAI.RoleUser {
		return nil
	}
	return uc.threads.AppendMessage(ctx, threadID, domainAI.RoleAssistant, generationFailedReply, 0, 0)
}

func threadMessageDTO(message domainAI.ThreadMessage) ThreadMessageDTO {
	dto := ThreadMessageDTO{
		ID:        message.ID,
		Role:      message.Role,
		Content:   message.Content,
		CreatedAt: message.CreatedAt,
	}
	if strings.TrimSpace(message.DraftJSON) == "" {
		return dto
	}
	var draft TransactionDraft
	if err := json.Unmarshal([]byte(message.DraftJSON), &draft); err != nil {
		return dto
	}
	dto.Draft = &draft
	return dto
}

func (uc *ThreadUseCases) MarkDraftSaved(ctx context.Context, userID, threadID, messageID int) error {
	if err := RequireProAI(ctx, uc.entitlements, userID); err != nil {
		return err
	}
	if _, err := uc.threads.FindByIDForUser(ctx, threadID, userID); err != nil {
		return err
	}
	message, err := uc.threads.FindMessage(ctx, threadID, messageID)
	if err != nil {
		return err
	}
	if strings.TrimSpace(message.DraftJSON) == "" {
		return domainAI.ErrDraftNotFound
	}
	var draft TransactionDraft
	if err := json.Unmarshal([]byte(message.DraftJSON), &draft); err != nil {
		return domainAI.ErrDraftNotFound
	}
	if draft.Status == DraftStatusSaved {
		return nil
	}
	draft.Status = DraftStatusSaved
	raw, err := json.Marshal(draft)
	if err != nil {
		return err
	}
	return uc.threads.UpdateMessageDraft(ctx, threadID, messageID, string(raw))
}

func (uc *ThreadUseCases) Delete(ctx context.Context, userID, threadID int) error {
	if err := RequireProAI(ctx, uc.entitlements, userID); err != nil {
		return err
	}
	return uc.threads.Delete(ctx, threadID, userID)
}

type RenameThreadRequest struct {
	Title string `json:"title"`
}

func (uc *ThreadUseCases) Rename(ctx context.Context, userID, threadID int, title string) (*ThreadSummaryDTO, error) {
	if err := RequireProAI(ctx, uc.entitlements, userID); err != nil {
		return nil, err
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, domainAI.ErrThreadNotFound
	}
	if utf8.RuneCountInString(title) > domainAI.MaxThreadTitle {
		runes := []rune(title)
		title = string(runes[:domainAI.MaxThreadTitle])
	}
	if err := uc.threads.UpdateTitle(ctx, threadID, userID, title); err != nil {
		return nil, err
	}
	thread, err := uc.threads.FindByIDForUser(ctx, threadID, userID)
	if err != nil {
		return nil, err
	}
	return &ThreadSummaryDTO{
		ID:               thread.ID,
		Title:            thread.Title,
		UpdatedAt:        thread.UpdatedAt,
		CreatedAt:        thread.CreatedAt,
		GenerationStatus: thread.GenerationStatus,
	}, nil
}

// LegacyGetOrCreate returns the newest thread or creates one (for old /thread aliases).
func (uc *ThreadUseCases) LegacyGetOrCreate(ctx context.Context, userID int) (*GetThreadResponse, error) {
	if err := RequireProAI(ctx, uc.entitlements, userID); err != nil {
		return nil, err
	}
	list, err := uc.threads.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		created, err := uc.threads.Create(ctx, userID)
		if err != nil {
			return nil, err
		}
		return uc.Get(ctx, userID, created.ID)
	}
	return uc.Get(ctx, userID, list[0].ID)
}

func (uc *ThreadUseCases) LegacyClear(ctx context.Context, userID int) error {
	resp, err := uc.LegacyGetOrCreate(ctx, userID)
	if err != nil {
		return err
	}
	return uc.threads.ClearMessages(ctx, resp.ID)
}
