package database

import (
	"context"
	"errors"
	"panda-pocket/internal/domain/identity"
	"time"

	"gorm.io/gorm"
)

// GormTokenRepository implements identity.TokenRepository using GORM
type GormTokenRepository struct {
	db *gorm.DB
}

// NewGormTokenRepository creates a new GORM token repository
func NewGormTokenRepository(db *gorm.DB) identity.TokenRepository {
	return &GormTokenRepository{db: db}
}

// Save saves a new token to the database
func (r *GormTokenRepository) Save(ctx context.Context, userID int, accessToken, refreshToken string, expiresAt int64) error {
	token := &Token{
		UserID:       uint(userID),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Unix(expiresAt, 0),
		Revoked:      false,
	}
	return r.db.WithContext(ctx).Create(token).Error
}

// FindByRefreshToken finds a token by refresh token string
func (r *GormTokenRepository) FindByRefreshToken(ctx context.Context, refreshToken string) (int, bool, error) {
	var token Token
	err := r.db.WithContext(ctx).Where("refresh_token = ?", refreshToken).First(&token).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, false, nil // Not found, treated as invalid
		}
		return 0, false, err
	}
	return int(token.UserID), token.Revoked, nil
}

// Revoke revokes a token by refresh token string
func (r *GormTokenRepository) Revoke(ctx context.Context, refreshToken string) error {
	return r.db.WithContext(ctx).Model(&Token{}).Where("refresh_token = ?", refreshToken).Update("revoked", true).Error
}

// DeleteByAccessToken deletes a token by access token string
func (r *GormTokenRepository) DeleteByAccessToken(ctx context.Context, accessToken string) error {
	return r.db.WithContext(ctx).Where("access_token = ?", accessToken).Delete(&Token{}).Error
}

// RevokeAllForUser revokes all tokens for a user
func (r *GormTokenRepository) RevokeAllForUser(ctx context.Context, userID int) error {
	return r.db.WithContext(ctx).Model(&Token{}).Where("user_id = ?", userID).Update("revoked", true).Error
}
