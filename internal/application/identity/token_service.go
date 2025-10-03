package identity

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	JWTSecret = "panda-pocket-secret-key-change-in-production"
)

// Claims represents JWT claims
type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// TokenService interface defines token operations
type TokenService interface {
	GenerateToken(userID int, email string, role string) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
}

// tokenService implements TokenService interface
type tokenService struct{}

// NewTokenService creates a new token service
func NewTokenService() TokenService {
	return &tokenService{}
}

// GenerateToken generates a JWT token for a user
func (s *tokenService) GenerateToken(userID int, email string, role string) (string, error) {
	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JWTSecret))
}

// ValidateToken validates a JWT token and returns the claims
func (s *tokenService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}
