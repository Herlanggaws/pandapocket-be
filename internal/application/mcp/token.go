package mcp

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"panda-pocket/internal/domain/entitlement"
)

const (
	tokenPrefix        = "bb_mcp_"
	tokenPrefixVisible = 8
	tokenSecretBytes   = 32
)

var ErrNotFound = errors.New("mcp token not found")

type StoredToken struct {
	UserID    int
	TokenHash string
	Prefix    string
	CreatedAt time.Time
	RevokedAt *time.Time
}

type Repository interface {
	FindByUserID(ctx context.Context, userID int) (*StoredToken, error)
	FindActiveByHash(ctx context.Context, hash string) (*StoredToken, error)
	Save(ctx context.Context, token StoredToken) error
	Revoke(ctx context.Context, userID int, at time.Time) error
}

type IssuedToken struct {
	Token     string    `json:"token"`
	Prefix    string    `json:"prefix"`
	CreatedAt time.Time `json:"created_at"`
}

type TokenStatus struct {
	Active    bool       `json:"active"`
	Prefix    string     `json:"prefix,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

type TokenService struct {
	repo         Repository
	entitlements entitlement.Checker
	now          func() time.Time
}

func NewTokenService(repo Repository, entitlements entitlement.Checker) *TokenService {
	return &TokenService{
		repo:         repo,
		entitlements: entitlements,
		now:          func() time.Time { return time.Now().UTC() },
	}
}

func (s *TokenService) Status(ctx context.Context, userID int) (*TokenStatus, error) {
	if err := s.requirePro(ctx, userID); err != nil {
		return nil, err
	}
	stored, err := s.repo.FindByUserID(ctx, userID)
	if errors.Is(err, ErrNotFound) || (err == nil && stored.RevokedAt != nil) {
		return &TokenStatus{Active: false}, nil
	}
	if err != nil {
		return nil, err
	}
	created := stored.CreatedAt
	return &TokenStatus{Active: true, Prefix: stored.Prefix, CreatedAt: &created}, nil
}

func (s *TokenService) Issue(ctx context.Context, userID int) (*IssuedToken, error) {
	if err := s.requirePro(ctx, userID); err != nil {
		return nil, err
	}
	plain, prefix, hash, err := newPlainToken()
	if err != nil {
		return nil, err
	}
	created := s.now()
	if err := s.repo.Save(ctx, StoredToken{
		UserID:    userID,
		TokenHash: hash,
		Prefix:    prefix,
		CreatedAt: created,
	}); err != nil {
		return nil, err
	}
	return &IssuedToken{Token: plain, Prefix: prefix, CreatedAt: created}, nil
}

func (s *TokenService) Revoke(ctx context.Context, userID int) error {
	if err := s.requirePro(ctx, userID); err != nil {
		return err
	}
	err := s.repo.Revoke(ctx, userID, s.now())
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	return err
}

func (s *TokenService) Authenticate(ctx context.Context, plain string) (int, error) {
	if plain == "" {
		return 0, ErrNotFound
	}
	stored, err := s.repo.FindActiveByHash(ctx, hashToken(plain))
	if err != nil {
		return 0, err
	}
	return stored.UserID, nil
}

func (s *TokenService) requirePro(ctx context.Context, userID int) error {
	isPro, err := s.entitlements.IsPro(ctx, userID)
	if err != nil {
		return err
	}
	if !isPro {
		return entitlement.RequirePro(entitlement.FeatureMCP)
	}
	return nil
}

func newPlainToken() (plain string, prefix string, hash string, err error) {
	buf := make([]byte, tokenSecretBytes)
	if _, err = rand.Read(buf); err != nil {
		return "", "", "", err
	}
	plain = tokenPrefix + hex.EncodeToString(buf)
	prefix = plain[:len(tokenPrefix)+tokenPrefixVisible]
	return plain, prefix, hashToken(plain), nil
}

func hashToken(plain string) string {
	sum := sha256.Sum256([]byte(plain))
	return hex.EncodeToString(sum[:])
}
