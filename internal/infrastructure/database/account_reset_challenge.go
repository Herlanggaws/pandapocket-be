package database

import (
	"context"
	"errors"
	"time"

	appIdentity "panda-pocket/internal/application/identity"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AccountResetChallenge stores a one-time confirmation code for account data reset.
type AccountResetChallenge struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID       uint      `gorm:"not null;index"`
	CodeHash     string    `gorm:"not null"`
	AttemptCount int       `gorm:"not null;default:0"`
	ExpiresAt    time.Time `gorm:"not null;index"`
	CreatedAt    time.Time `gorm:"not null"`
}

func (AccountResetChallenge) TableName() string {
	return "account_reset_challenges"
}

// GormAccountResetChallengeRepository persists account reset challenges.
type GormAccountResetChallengeRepository struct {
	db *gorm.DB
}

func NewGormAccountResetChallengeRepository(db *gorm.DB) *GormAccountResetChallengeRepository {
	return &GormAccountResetChallengeRepository{db: db}
}

func (r *GormAccountResetChallengeRepository) DeleteByUserID(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&AccountResetChallenge{}).Error
}

func (r *GormAccountResetChallengeRepository) Save(ctx context.Context, challenge *appIdentity.AccountResetChallenge) error {
	model := &AccountResetChallenge{
		ID:           challenge.ID,
		UserID:       challenge.UserID,
		CodeHash:     challenge.CodeHash,
		AttemptCount: challenge.AttemptCount,
		ExpiresAt:    challenge.ExpiresAt,
		CreatedAt:    challenge.CreatedAt,
	}
	return r.db.WithContext(ctx).Create(model).Error
}

func (r *GormAccountResetChallengeRepository) FindLatestByUserID(ctx context.Context, userID uint) (*appIdentity.AccountResetChallenge, error) {
	var model AccountResetChallenge
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &appIdentity.AccountResetChallenge{
		ID:           model.ID,
		UserID:       model.UserID,
		CodeHash:     model.CodeHash,
		AttemptCount: model.AttemptCount,
		ExpiresAt:    model.ExpiresAt,
		CreatedAt:    model.CreatedAt,
	}, nil
}

func (r *GormAccountResetChallengeRepository) IncrementAttempts(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&AccountResetChallenge{}).
		Where("id = ?", id).
		UpdateColumn("attempt_count", gorm.Expr("attempt_count + 1")).Error
}

func (r *GormAccountResetChallengeRepository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&AccountResetChallenge{}, "id = ?", id).Error
}

// GormUserDataWiper hard-deletes all user-owned financial data in one transaction.
type GormUserDataWiper struct {
	db *gorm.DB
}

func NewGormUserDataWiper(db *gorm.DB) *GormUserDataWiper {
	return &GormUserDataWiper{db: db}
}

func wipeUserOwnedRows(tx *gorm.DB, userID uint) error {
	deletes := []func() error{
		func() error { return tx.Where("user_id = ?", userID).Delete(&PendingTransaction{}).Error },
		func() error { return tx.Where("user_id = ?", userID).Delete(&GoalContribution{}).Error },
		func() error { return tx.Where("user_id = ?", userID).Delete(&LiabilityPayment{}).Error },
		func() error { return tx.Where("user_id = ?", userID).Delete(&RecurringTransaction{}).Error },
		func() error { return tx.Where("user_id = ?", userID).Delete(&Transfer{}).Error },
		func() error { return tx.Where("user_id = ?", userID).Delete(&Expense{}).Error },
		func() error { return tx.Where("user_id = ?", userID).Delete(&Income{}).Error },
		func() error { return tx.Where("user_id = ?", userID).Delete(&Budget{}).Error },
		func() error { return tx.Where("user_id = ?", userID).Delete(&FinancialGoal{}).Error },
		func() error { return tx.Where("user_id = ?", userID).Delete(&Asset{}).Error },
		func() error { return tx.Where("user_id = ?", userID).Delete(&Liability{}).Error },
		func() error { return tx.Where("user_id = ?", userID).Delete(&Wallet{}).Error },
		func() error { return tx.Where("user_id = ?", userID).Delete(&HealthScoreSnapshot{}).Error },
		func() error { return tx.Where("user_id = ?", userID).Delete(&Notification{}).Error },
		func() error { return tx.Where("user_id = ?", userID).Delete(&UserFeedback{}).Error },
		func() error { return tx.Where("user_id = ?", userID).Delete(&SupportTicket{}).Error },
		func() error { return tx.Where("user_id = ?", userID).Delete(&Category{}).Error },
		func() error { return tx.Where("user_id = ?", userID).Delete(&Currency{}).Error },
		func() error { return tx.Where("user_id = ?", userID).Delete(&UserPreferences{}).Error },
		func() error { return tx.Where("user_id = ?", userID).Delete(&AccountResetChallenge{}).Error },
		func() error { return tx.Where("user_id = ?", userID).Delete(&PasswordResetToken{}).Error },
	}

	for _, deleteFn := range deletes {
		if err := deleteFn(); err != nil {
			return err
		}
	}
	return nil
}

func (w *GormUserDataWiper) WipeUserData(ctx context.Context, userID uint) error {
	return w.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return wipeUserOwnedRows(tx, userID)
	})
}

// HardDeleteAccount wipes owned data, auth tokens, and the user row.
func (w *GormUserDataWiper) HardDeleteAccount(ctx context.Context, userID uint) error {
	return w.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := wipeUserOwnedRows(tx, userID); err != nil {
			return err
		}
		if err := tx.Where("user_id = ?", userID).Delete(&Token{}).Error; err != nil {
			return err
		}
		return tx.Delete(&User{}, userID).Error
	})
}
