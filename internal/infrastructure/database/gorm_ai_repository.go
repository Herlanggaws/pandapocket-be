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
	ID                  uint       `gorm:"primaryKey" json:"id"`
	UserID              uint       `gorm:"index;not null" json:"user_id"`
	Title               string     `gorm:"type:varchar(80);not null;default:''" json:"title"`
	GenerationStatus    string     `gorm:"type:varchar(16);not null;default:'idle'" json:"generation_status"`
	GenerationStartedAt *time.Time `json:"generation_started_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
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

func mapThread(row AIAdvisorThread) domainAI.Thread {
	return domainAI.Thread{
		ID:                  int(row.ID),
		UserID:              int(row.UserID),
		Title:               row.Title,
		GenerationStatus:    row.GenerationStatus,
		GenerationStartedAt: row.GenerationStartedAt,
		CreatedAt:           row.CreatedAt,
		UpdatedAt:           row.UpdatedAt,
	}
}

func (r *GormAIThreadRepository) ListByUserID(ctx context.Context, userID int) ([]domainAI.Thread, error) {
	var rows []AIAdvisorThread
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("updated_at desc, id desc").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domainAI.Thread, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapThread(row))
	}
	return out, nil
}

func (r *GormAIThreadRepository) Create(ctx context.Context, userID int) (*domainAI.Thread, error) {
	now := time.Now().UTC()
	thread := AIAdvisorThread{
		UserID:           uint(userID),
		Title:            "",
		GenerationStatus: domainAI.GenerationIdle,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := r.db.WithContext(ctx).Create(&thread).Error; err != nil {
		return nil, err
	}
	t := mapThread(thread)
	return &t, nil
}

func (r *GormAIThreadRepository) FindByIDForUser(ctx context.Context, threadID, userID int) (*domainAI.Thread, error) {
	var row AIAdvisorThread
	err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", threadID, userID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domainAI.ErrThreadNotFound
	}
	if err != nil {
		return nil, err
	}
	t := mapThread(row)
	return &t, nil
}

func (r *GormAIThreadRepository) Delete(ctx context.Context, threadID, userID int) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Where("id = ? AND user_id = ?", threadID, userID).Delete(&AIAdvisorThread{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return domainAI.ErrThreadNotFound
		}
		return tx.Where("thread_id = ?", threadID).Delete(&AIAdvisorMessage{}).Error
	})
}

func (r *GormAIThreadRepository) UpdateTitle(ctx context.Context, threadID, userID int, title string) error {
	res := r.db.WithContext(ctx).Model(&AIAdvisorThread{}).
		Where("id = ? AND user_id = ?", threadID, userID).
		Updates(map[string]interface{}{"title": title, "updated_at": time.Now().UTC()})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return domainAI.ErrThreadNotFound
	}
	return nil
}

func (r *GormAIThreadRepository) SetTitleIfEmpty(ctx context.Context, threadID int, title string) error {
	return r.db.WithContext(ctx).Model(&AIAdvisorThread{}).
		Where("id = ? AND (title = '' OR title IS NULL)", threadID).
		Updates(map[string]interface{}{"title": title, "updated_at": time.Now().UTC()}).Error
}

func (r *GormAIThreadRepository) TryBeginGeneration(ctx context.Context, threadID, userID int) error {
	now := time.Now().UTC()
	res := r.db.WithContext(ctx).Model(&AIAdvisorThread{}).
		Where("id = ? AND user_id = ? AND generation_status <> ?", threadID, userID, domainAI.GenerationPending).
		Updates(map[string]interface{}{
			"generation_status":     domainAI.GenerationPending,
			"generation_started_at": now,
			"updated_at":            now,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		var row AIAdvisorThread
		err := r.db.WithContext(ctx).Where("id = ? AND user_id = ?", threadID, userID).First(&row).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domainAI.ErrThreadNotFound
		}
		if err != nil {
			return err
		}
		return domainAI.ErrTurnInProgress
	}
	return nil
}

func (r *GormAIThreadRepository) FinishGeneration(ctx context.Context, threadID int, status string) error {
	if status != domainAI.GenerationIdle && status != domainAI.GenerationFailed {
		status = domainAI.GenerationIdle
	}
	// Surface failed briefly then idle — plan: clear to idle on success/fail for poll simplicity.
	_ = status
	return r.db.WithContext(ctx).Model(&AIAdvisorThread{}).
		Where("id = ?", threadID).
		Updates(map[string]interface{}{
			"generation_status":     domainAI.GenerationIdle,
			"generation_started_at": nil,
			"updated_at":            time.Now().UTC(),
		}).Error
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
