package ai

import (
	"context"
	"time"

	domainAI "panda-pocket/internal/domain/ai"
	"panda-pocket/internal/domain/entitlement"
)

type ThreadMessageDTO struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type GetThreadResponse struct {
	Messages []ThreadMessageDTO `json:"messages"`
	Credits  *CreditsView       `json:"credits"`
}

type GetThreadUseCase struct {
	credits      *CreditService
	threads      domainAI.ThreadRepository
	entitlements entitlement.Checker
}

func NewGetThreadUseCase(credits *CreditService, threads domainAI.ThreadRepository, entitlements entitlement.Checker) *GetThreadUseCase {
	return &GetThreadUseCase{credits: credits, threads: threads, entitlements: entitlements}
}

func (uc *GetThreadUseCase) Execute(ctx context.Context, userID int) (*GetThreadResponse, error) {
	if err := RequireProAI(ctx, uc.entitlements, userID); err != nil {
		return nil, err
	}
	credits, err := uc.credits.View(ctx, userID)
	if err != nil {
		return nil, err
	}
	threadID, err := uc.threads.GetOrCreateThreadID(ctx, userID)
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
		out = append(out, ThreadMessageDTO{Role: m.Role, Content: m.Content, CreatedAt: m.CreatedAt})
	}
	return &GetThreadResponse{Messages: out, Credits: credits}, nil
}

type ClearThreadUseCase struct {
	threads      domainAI.ThreadRepository
	entitlements entitlement.Checker
}

func NewClearThreadUseCase(threads domainAI.ThreadRepository, entitlements entitlement.Checker) *ClearThreadUseCase {
	return &ClearThreadUseCase{threads: threads, entitlements: entitlements}
}

func (uc *ClearThreadUseCase) Execute(ctx context.Context, userID int) error {
	if err := RequireProAI(ctx, uc.entitlements, userID); err != nil {
		return err
	}
	threadID, err := uc.threads.GetOrCreateThreadID(ctx, userID)
	if err != nil {
		return err
	}
	return uc.threads.ClearMessages(ctx, threadID)
}
