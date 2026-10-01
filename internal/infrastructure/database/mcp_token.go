package database

import (
	"context"
	"errors"
	"time"

	appMCP "panda-pocket/internal/application/mcp"

	"gorm.io/gorm"
)

// McpToken is the single remote MCP credential for a user. Only the hash is stored.
type McpToken struct {
	ID        uint       `gorm:"primaryKey"`
	UserID    uint       `gorm:"not null;uniqueIndex"`
	TokenHash string     `gorm:"not null;uniqueIndex;size:64"`
	Prefix    string     `gorm:"not null;size:32"`
	CreatedAt time.Time  `gorm:"not null"`
	RevokedAt *time.Time `gorm:"index"`
}

func (McpToken) TableName() string {
	return "mcp_tokens"
}

type GormMcpTokenRepository struct {
	db *gorm.DB
}

func NewGormMcpTokenRepository(db *gorm.DB) *GormMcpTokenRepository {
	return &GormMcpTokenRepository{db: db}
}

func (r *GormMcpTokenRepository) FindByUserID(ctx context.Context, userID int) (*appMCP.StoredToken, error) {
	var row McpToken
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appMCP.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toStoredToken(row), nil
}

func (r *GormMcpTokenRepository) FindActiveByHash(ctx context.Context, hash string) (*appMCP.StoredToken, error) {
	var row McpToken
	err := r.db.WithContext(ctx).
		Where("token_hash = ? AND revoked_at IS NULL", hash).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, appMCP.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toStoredToken(row), nil
}

func (r *GormMcpTokenRepository) Save(ctx context.Context, token appMCP.StoredToken) error {
	var existing McpToken
	err := r.db.WithContext(ctx).Where("user_id = ?", token.UserID).First(&existing).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return r.db.WithContext(ctx).Create(&McpToken{
			UserID:    uint(token.UserID),
			TokenHash: token.TokenHash,
			Prefix:    token.Prefix,
			CreatedAt: token.CreatedAt,
			RevokedAt: nil,
		}).Error
	}
	if err != nil {
		return err
	}
	existing.TokenHash = token.TokenHash
	existing.Prefix = token.Prefix
	existing.CreatedAt = token.CreatedAt
	existing.RevokedAt = nil
	return r.db.WithContext(ctx).Save(&existing).Error
}

func (r *GormMcpTokenRepository) Revoke(ctx context.Context, userID int, at time.Time) error {
	result := r.db.WithContext(ctx).
		Model(&McpToken{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", at)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return appMCP.ErrNotFound
	}
	return nil
}

func toStoredToken(row McpToken) *appMCP.StoredToken {
	return &appMCP.StoredToken{
		UserID:    int(row.UserID),
		TokenHash: row.TokenHash,
		Prefix:    row.Prefix,
		CreatedAt: row.CreatedAt,
		RevokedAt: row.RevokedAt,
	}
}
