package identity

import (
	"context"
	"time"
)

// TokenRepository defines the contract for token persistence
type TokenRepository interface {
	Save(ctx context.Context, userID int, accessToken, refreshToken string, expiresAt int64) error
	FindByRefreshToken(ctx context.Context, refreshToken string) (int, bool, error) // Returns userID, revoked status, error
	DeleteByAccessToken(ctx context.Context, accessToken string) error
	Revoke(ctx context.Context, refreshToken string) error
	RevokeAllForUser(ctx context.Context, userID int) error
	DeleteExpired(ctx context.Context, before time.Time) (int64, error)
}
