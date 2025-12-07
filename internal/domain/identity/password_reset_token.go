package identity

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PasswordResetToken represents a token used for password reset
type PasswordResetToken struct {
	id        uuid.UUID
	userID    UserID
	token     string
	expiresAt time.Time
	createdAt time.Time
}

// NewPasswordResetToken creates a new password reset token
func NewPasswordResetToken(userID UserID, token string, expiresAt time.Time) *PasswordResetToken {
	return &PasswordResetToken{
		id:        uuid.New(),
		userID:    userID,
		token:     token,
		expiresAt: expiresAt,
		createdAt: time.Now(),
	}
}

// Getters

func (t *PasswordResetToken) ID() uuid.UUID {
	return t.id
}

func (t *PasswordResetToken) UserID() UserID {
	return t.userID
}

func (t *PasswordResetToken) Token() string {
	return t.token
}

func (t *PasswordResetToken) ExpiresAt() time.Time {
	return t.expiresAt
}

func (t *PasswordResetToken) CreatedAt() time.Time {
	return t.createdAt
}

// IsExpired checks if the token is expired
func (t *PasswordResetToken) IsExpired() bool {
	return time.Now().After(t.expiresAt)
}

// PasswordResetTokenRepository defines the contract for password reset token persistence
type PasswordResetTokenRepository interface {
	Save(ctx context.Context, token *PasswordResetToken) error
	FindByToken(ctx context.Context, token string) (*PasswordResetToken, error)
}
