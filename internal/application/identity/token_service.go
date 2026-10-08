package identity

import (
	"context"
	"errors"
	"os"
	"panda-pocket/internal/domain/identity"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	DefaultJWTSecret                   = "panda-pocket-secret-key-change-in-production"
	DefaultRefreshTokenSecret          = "panda-pocket-refresh-token-secret-change-in-production"
	AccessTokenTTL                     = 15 * time.Minute
	DefaultRefreshTokenExpirationHours = 168
)

// Claims represents JWT claims
type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// RefreshTokenClaims represents claims for refresh token
type RefreshTokenClaims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// TokenService interface defines token operations
type TokenService interface {
	GenerateToken(ctx context.Context, userID int, email string, role string) (string, string, error)
	ValidateToken(tokenString string) (*Claims, error)
	AccessTokenActive(ctx context.Context, tokenString string) (bool, error)
	RefreshTokenOwner(ctx context.Context, refreshToken string) (userID int, found bool, err error)
	ValidateRefreshToken(ctx context.Context, tokenString string) (*RefreshTokenClaims, error)
	GenerateAccessTokenResult(userID int, email string, role string) (string, error)
	RevokeToken(ctx context.Context, tokenString string) error
	RevokeAllForUser(ctx context.Context, userID int) error
	RefreshCookieMaxAge() int
	CleanupExpiredToken(ctx context.Context, tokenString string) error
}

// tokenService implements TokenService interface
type tokenService struct {
	repo                        identity.TokenRepository
	jwtSecret                   string
	refreshTokenSecret          string
	refreshTokenExpirationHours int
}

// NewTokenService creates a new token service
func NewTokenService(repo identity.TokenRepository) TokenService {
	jwtSecret := getEnv("JWT_SECRET", DefaultJWTSecret)
	refreshTokenSecret := getEnv("REFRESH_TOKEN_SECRET", DefaultRefreshTokenSecret)

	refreshTokenExpirationHoursStr := getEnv("REFRESH_TOKEN_EXPIRATION_HOURS", strconv.Itoa(DefaultRefreshTokenExpirationHours))
	refreshTokenExpirationHours, err := strconv.Atoi(refreshTokenExpirationHoursStr)
	if err != nil {
		refreshTokenExpirationHours = DefaultRefreshTokenExpirationHours
	}

	return &tokenService{
		repo:                        repo,
		jwtSecret:                   jwtSecret,
		refreshTokenSecret:          refreshTokenSecret,
		refreshTokenExpirationHours: refreshTokenExpirationHours,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GenerateToken generates a JWT token pair for a user
func (s *tokenService) GenerateToken(ctx context.Context, userID int, email string, role string) (string, string, error) {
	if userID <= 0 {
		return "", "", errors.New("invalid token user")
	}

	// Generate Access Token
	accessToken, err := s.GenerateAccessTokenResult(userID, email, role)
	if err != nil {
		return "", "", err
	}

	// Calculate expiration time
	expiresIn := time.Duration(s.refreshTokenExpirationHours) * time.Hour
	expiresAt := time.Now().Add(expiresIn)

	// Generate Refresh Token
	refreshClaims := RefreshTokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	refreshTokenObj := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshToken, err := refreshTokenObj.SignedString([]byte(s.refreshTokenSecret))
	if err != nil {
		return "", "", err
	}

	// Save token to database
	err = s.repo.Save(ctx, userID, accessToken, refreshToken, expiresAt.Unix())
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

// GenerateAccessTokenResult generates a single access token (internal helper exposed for refresh flow)
func (s *tokenService) GenerateAccessTokenResult(userID int, email string, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(AccessTokenTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// ValidateToken validates a JWT token and returns the claims
func (s *tokenService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		if claims.UserID <= 0 {
			return nil, errors.New("invalid token user")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// AccessTokenActive reports whether this access token still has a live session row.
// A missing or revoked row is inactive. Database failures are returned as errors.
func (s *tokenService) AccessTokenActive(ctx context.Context, tokenString string) (bool, error) {
	found, revoked, err := s.repo.FindByAccessToken(ctx, tokenString)
	if err != nil {
		return false, err
	}
	return found && !revoked, nil
}

// RefreshTokenOwner returns the user that owns the refresh token row.
func (s *tokenService) RefreshTokenOwner(ctx context.Context, refreshToken string) (int, bool, error) {
	userID, _, err := s.repo.FindByRefreshToken(ctx, refreshToken)
	if err != nil {
		return 0, false, err
	}
	if userID <= 0 {
		return 0, false, nil
	}
	return userID, true, nil
}

// ValidateRefreshToken validates a refresh token and returns the claims
func (s *tokenService) ValidateRefreshToken(ctx context.Context, tokenString string) (*RefreshTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.refreshTokenSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*RefreshTokenClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}
	if claims.UserID <= 0 {
		return nil, errors.New("invalid refresh token user")
	}

	// Check against database
	_, revoked, err := s.repo.FindByRefreshToken(ctx, tokenString)
	if err != nil {
		return nil, err // Database error
	}
	if revoked {
		return nil, errors.New("token revoked")
	}
	// Note: We could verify userID matches claims.UserID here too for extra safety

	return claims, nil
}

// RevokeToken revokes a refresh token
func (s *tokenService) RevokeToken(ctx context.Context, tokenString string) error {
	return s.repo.Revoke(ctx, tokenString)
}

// RevokeAllForUser revokes all tokens for a user
func (s *tokenService) RevokeAllForUser(ctx context.Context, userID int) error {
	return s.repo.RevokeAllForUser(ctx, userID)
}

func (s *tokenService) RefreshCookieMaxAge() int {
	return s.refreshTokenExpirationHours * 3600
}

// CleanupExpiredToken removes an expired access token from the database
func (s *tokenService) CleanupExpiredToken(ctx context.Context, tokenString string) error {
	// We swallow the error here as this is a best-effort cleanup
	// and we don't want to disrupt the flow if the DB is down or record missing
	_ = s.repo.DeleteByAccessToken(ctx, tokenString)
	return nil
}
