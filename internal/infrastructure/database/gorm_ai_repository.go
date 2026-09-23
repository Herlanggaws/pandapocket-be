package database

import (
	"context"
	"errors"
	"time"

	domainAI "panda-pocket/internal/domain/ai"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type AICreditBalance struct {
	UserID             uint      `gorm:"primaryKey" json:"user_id"`
	IncludedGranted    int       `gorm:"not null;default:75" json:"included_granted"`
	IncludedUnlocked   int       `gorm:"not null;default:15" json:"included_unlocked"`
	IncludedUsed       int       `gorm:"not null;default:0" json:"included_used"`
	PurchasedRemaining int       `gorm:"not null;default:0" json:"purchased_remaining"`
	UpdatedAt          time.Time `json:"updated_at"`
}

func (AICreditBalance) TableName() string { return "ai_credit_balances" }

type AICreditLedgerEntry struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	UserID        uint      `gorm:"not null;index" json:"user_id"`
	Kind          string    `gorm:"type:varchar(32);not null" json:"kind"`
	DeltaPurchased int      `gorm:"not null;default:0" json:"delta_purchased"`
	DeltaIncludedUsed int   `gorm:"not null;default:0" json:"delta_included_used"`
	Pack          *string   `gorm:"type:varchar(32)" json:"pack,omitempty"`
	DoitPaymentID *string   `gorm:"type:varchar(128);uniqueIndex" json:"doit_payment_id,omitempty"`
	Note          string    `gorm:"type:text" json:"note"`
	CreatedAt     time.Time `json:"created_at"`
}

func (AICreditLedgerEntry) TableName() string { return "ai_credit_ledger" }

type AIAdvisorThread struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex;not null" json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (AIAdvisorThread) TableName() string { return "ai_advisor_threads" }

type AIAdvisorMessage struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	ThreadID         uint      `gorm:"not null;index" json:"thread_id"`
	Role             string    `gorm:"type:varchar(16);not null" json:"role"`
	Content          string    `gorm:"type:text;not null" json:"content"`
	PromptTokens     int       `gorm:"not null;default:0" json:"prompt_tokens"`
	CompletionTokens int       `gorm:"not null;default:0" json:"completion_tokens"`
	CreatedAt        time.Time `json:"created_at"`
}

func (AIAdvisorMessage) TableName() string { return "ai_advisor_messages" }

type GormAICreditRepository struct {
	db *gorm.DB
}

func NewGormAICreditRepository(db *gorm.DB) *GormAICreditRepository {
	return &GormAICreditRepository{db: db}
}

func (r *GormAICreditRepository) FindByUserID(ctx context.Context, userID int) (*domainAI.CreditBalance, error) {
	var model AICreditBalance
	err := r.db.WithContext(ctx).First(&model, "user_id = ?", userID).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domainAI.ErrBalanceNotFound
	}
	if err != nil {
		return nil, err
	}
	return &domainAI.CreditBalance{
		UserID:             int(model.UserID),
		IncludedGranted:    model.IncludedGranted,
		IncludedUnlocked:   model.IncludedUnlocked,
		IncludedUsed:       model.IncludedUsed,
		PurchasedRemaining: model.PurchasedRemaining,
		UpdatedAt:          model.UpdatedAt,
	}, nil
}

func (r *GormAICreditRepository) Save(ctx context.Context, balance *domainAI.CreditBalance) error {
	model := AICreditBalance{
		UserID:             uint(balance.UserID),
		IncludedGranted:    balance.IncludedGranted,
		IncludedUnlocked:   balance.IncludedUnlocked,
		IncludedUsed:       balance.IncludedUsed,
		PurchasedRemaining: balance.PurchasedRemaining,
		UpdatedAt:          balance.UpdatedAt,
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"included_granted", "included_unlocked", "included_used", "purchased_remaining", "updated_at"}),
	}).Create(&model).Error
}

func (r *GormAICreditRepository) PurchaseExists(ctx context.Context, doitPaymentID string) (bool, error) {
	if doitPaymentID == "" {
		return false, nil
	}
	var count int64
	err := r.db.WithContext(ctx).Model(&AICreditLedgerEntry{}).Where("doit_payment_id = ?", doitPaymentID).Count(&count).Error
	return count > 0, err
}

func (r *GormAICreditRepository) RecordPurchase(ctx context.Context, userID int, pack, doitPaymentID string, credits int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var existing int64
		if err := tx.Model(&AICreditLedgerEntry{}).Where("doit_payment_id = ?", doitPaymentID).Count(&existing).Error; err != nil {
			return err
		}
		if existing > 0 {
			return nil
		}

		var model AICreditBalance
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&model, "user_id = ?", userID).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			balance := domainAI.NewCreditBalance(userID, false)
			balance.AddPurchase(credits)
			model = AICreditBalance{
				UserID:             uint(userID),
				IncludedGranted:    balance.IncludedGranted,
				IncludedUnlocked:   balance.IncludedUnlocked,
				IncludedUsed:       balance.IncludedUsed,
				PurchasedRemaining: balance.PurchasedRemaining,
				UpdatedAt:          balance.UpdatedAt,
			}
			if err := tx.Create(&model).Error; err != nil {
				return err
			}
		} else if err != nil {
			return err
		} else {
			model.PurchasedRemaining += credits
			model.UpdatedAt = time.Now().UTC()
			if err := tx.Save(&model).Error; err != nil {
				return err
			}
		}

		packCopy := pack
		paymentCopy := doitPaymentID
		entry := AICreditLedgerEntry{
			UserID:         uint(userID),
			Kind:           "purchase",
			DeltaPurchased: credits,
			Pack:           &packCopy,
			DoitPaymentID:  &paymentCopy,
			CreatedAt:      time.Now().UTC(),
		}
		return tx.Create(&entry).Error
	})
}

type GormAIThreadRepository struct {
	db *gorm.DB
}

func NewGormAIThreadRepository(db *gorm.DB) *GormAIThreadRepository {
	return &GormAIThreadRepository{db: db}
}

func (r *GormAIThreadRepository) GetOrCreateThreadID(ctx context.Context, userID int) (int, error) {
	var thread AIAdvisorThread
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&thread).Error
	if err == nil {
		return int(thread.ID), nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	thread = AIAdvisorThread{UserID: uint(userID), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	if err := r.db.WithContext(ctx).Create(&thread).Error; err != nil {
		return 0, err
	}
	return int(thread.ID), nil
}

func (r *GormAIThreadRepository) ListMessages(ctx context.Context, threadID int) ([]domainAI.ThreadMessage, error) {
	var rows []AIAdvisorMessage
	if err := r.db.WithContext(ctx).Where("thread_id = ?", threadID).Order("created_at asc, id asc").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domainAI.ThreadMessage, 0, len(rows))
	for _, row := range rows {
		out = append(out, domainAI.ThreadMessage{
			ID:               int(row.ID),
			Role:             row.Role,
			Content:          row.Content,
			PromptTokens:     row.PromptTokens,
			CompletionTokens: row.CompletionTokens,
			CreatedAt:        row.CreatedAt,
		})
	}
	return out, nil
}

func (r *GormAIThreadRepository) AppendMessage(ctx context.Context, threadID int, role, content string, promptTokens, completionTokens int) error {
	msg := AIAdvisorMessage{
		ThreadID:         uint(threadID),
		Role:             role,
		Content:          content,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		CreatedAt:        time.Now().UTC(),
	}
	if err := r.db.WithContext(ctx).Create(&msg).Error; err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&AIAdvisorThread{}).Where("id = ?", threadID).Update("updated_at", time.Now().UTC()).Error
}

func (r *GormAIThreadRepository) ClearMessages(ctx context.Context, threadID int) error {
	return r.db.WithContext(ctx).Where("thread_id = ?", threadID).Delete(&AIAdvisorMessage{}).Error
}

func (r *GormAIThreadRepository) TrimOldest(ctx context.Context, threadID int, keep int) error {
	var count int64
	if err := r.db.WithContext(ctx).Model(&AIAdvisorMessage{}).Where("thread_id = ?", threadID).Count(&count).Error; err != nil {
		return err
	}
	if int(count) <= keep {
		return nil
	}
	overflow := int(count) - keep
	var ids []uint
	if err := r.db.WithContext(ctx).Model(&AIAdvisorMessage{}).
		Where("thread_id = ?", threadID).
		Order("created_at asc, id asc").
		Limit(overflow).
		Pluck("id", &ids).Error; err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&AIAdvisorMessage{}).Error
}
