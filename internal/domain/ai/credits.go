package ai

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	ErrCreditsRequired = errors.New("ai credits required")
	ErrInvalidPack     = errors.New("invalid ai credit pack")
	ErrMessageTooLong  = errors.New("message exceeds max length")
	ErrUpstream        = errors.New("ai upstream error")
	ErrNotConfigured   = errors.New("ai provider is not configured")
	ErrBalanceNotFound = errors.New("ai credit balance not found")
	ErrThreadNotFound  = errors.New("ai advisor thread not found")
	ErrTurnInProgress  = errors.New("ai turn already in progress")
)

const (
	ProductAICredits = "ai_credits"
	PackS            = "ai_credits_s"
	PackM            = "ai_credits_m"
	PackSCredits     = 50
	PackMCredits     = 150
	PackSAmount      = 9900
	PackMAmount      = 24900
	MaxMessageLen    = 2000
	MaxThreadMsgs    = 50
	MaxThreadTitle   = 60
	RoleUser         = "user"
	RoleAssistant    = "assistant"

	GenerationIdle    = "idle"
	GenerationPending = "pending"
	GenerationFailed  = "failed"

	GenerationTimeout = 90 * time.Second
)

func IncludedGrant() int {
	return envInt("AI_INCLUDED_CREDITS", 75)
}

func TrialIncludedCap() int {
	return envInt("AI_TRIAL_INCLUDED_CAP", 15)
}

func envInt(key string, fallback int) int {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return fallback
	}
	return n
}

func PackCredits(pack string) (int, int, error) {
	switch pack {
	case PackS:
		return PackSCredits, PackSAmount, nil
	case PackM:
		return PackMCredits, PackMAmount, nil
	default:
		return 0, 0, ErrInvalidPack
	}
}

// CreditBalance is the per-user AI message credit ledger snapshot.
type CreditBalance struct {
	UserID              int
	IncludedGranted     int
	IncludedUnlocked    int
	IncludedUsed        int
	PurchasedRemaining  int
	UpdatedAt           time.Time
}

func (b *CreditBalance) IncludedRemaining() int {
	remaining := b.IncludedUnlocked - b.IncludedUsed
	if remaining < 0 {
		return 0
	}
	return remaining
}

func (b *CreditBalance) Available() int {
	return b.IncludedRemaining() + b.PurchasedRemaining
}

func (b *CreditBalance) Spend() error {
	if b.Available() < 1 {
		return ErrCreditsRequired
	}
	if b.IncludedRemaining() > 0 {
		b.IncludedUsed++
	} else {
		b.PurchasedRemaining--
	}
	b.UpdatedAt = time.Now().UTC()
	return nil
}

func (b *CreditBalance) AddPurchase(credits int) {
	if credits <= 0 {
		return
	}
	b.PurchasedRemaining += credits
	b.UpdatedAt = time.Now().UTC()
}

func (b *CreditBalance) UnlockFullIncluded() {
	grant := IncludedGrant()
	if b.IncludedGranted < grant {
		b.IncludedGranted = grant
	}
	b.IncludedUnlocked = grant
	b.UpdatedAt = time.Now().UTC()
}

func NewCreditBalance(userID int, isTrialing bool) *CreditBalance {
	grant := IncludedGrant()
	unlocked := grant
	if isTrialing {
		unlocked = TrialIncludedCap()
		if unlocked > grant {
			unlocked = grant
		}
	}
	now := time.Now().UTC()
	return &CreditBalance{
		UserID:           userID,
		IncludedGranted:  grant,
		IncludedUnlocked: unlocked,
		IncludedUsed:     0,
		PurchasedRemaining: 0,
		UpdatedAt:        now,
	}
}

func SyncUnlockForSubscription(b *CreditBalance, isTrialing bool) {
	grant := IncludedGrant()
	if b.IncludedGranted < grant {
		b.IncludedGranted = grant
	}
	desired := grant
	if isTrialing {
		desired = TrialIncludedCap()
		if desired > grant {
			desired = grant
		}
	}
	// Never lower unlock below what was already granted for paid users mid-cycle.
	if !isTrialing {
		b.IncludedUnlocked = grant
	} else if b.IncludedUnlocked < desired {
		b.IncludedUnlocked = desired
	}
	b.UpdatedAt = time.Now().UTC()
}

type CreditBalanceRepository interface {
	FindByUserID(ctx context.Context, userID int) (*CreditBalance, error)
	Save(ctx context.Context, balance *CreditBalance) error
	PurchaseExists(ctx context.Context, doitPaymentID string) (bool, error)
	RecordPurchase(ctx context.Context, userID int, pack, doitPaymentID string, credits int) error
}

type ThreadMessage struct {
	ID               int
	Role             string
	Content          string
	PromptTokens     int
	CompletionTokens int
	CreatedAt        time.Time
}

type Thread struct {
	ID                  int
	UserID              int
	Title               string
	GenerationStatus    string
	GenerationStartedAt *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type ThreadRepository interface {
	ListByUserID(ctx context.Context, userID int) ([]Thread, error)
	Create(ctx context.Context, userID int) (*Thread, error)
	FindByIDForUser(ctx context.Context, threadID, userID int) (*Thread, error)
	Delete(ctx context.Context, threadID, userID int) error
	UpdateTitle(ctx context.Context, threadID, userID int, title string) error
	SetTitleIfEmpty(ctx context.Context, threadID int, title string) error
	TryBeginGeneration(ctx context.Context, threadID, userID int) error
	FinishGeneration(ctx context.Context, threadID int, status string) error
	ListMessages(ctx context.Context, threadID int) ([]ThreadMessage, error)
	AppendMessage(ctx context.Context, threadID int, role, content string, promptTokens, completionTokens int) error
	ClearMessages(ctx context.Context, threadID int) error
	TrimOldest(ctx context.Context, threadID int, keep int) error
}

func TruncateTitle(message string) string {
	runes := []rune(strings.TrimSpace(message))
	if len(runes) == 0 {
		return "New chat"
	}
	if len(runes) <= MaxThreadTitle {
		return string(runes)
	}
	return string(runes[:MaxThreadTitle-1]) + "…"
}

func FormatCreditsRequired() string {
	return fmt.Sprintf("AI credits exhausted. Top up to continue.")
}
