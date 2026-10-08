package identity

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestAccessTokenExpiresInFifteenMinutes(t *testing.T) {
	t.Setenv("JWT_EXPIRATION_HOURS", "24")
	service := NewTokenService(&memAuthTokenRepo{})

	raw, err := service.GenerateAccessTokenResult(1, "ada@example.com", "user")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := service.ValidateToken(raw)
	if err != nil {
		t.Fatal(err)
	}
	lifetime := claims.ExpiresAt.Time.Sub(claims.IssuedAt.Time)
	if lifetime < 14*time.Minute || lifetime > 16*time.Minute {
		t.Fatalf("access token lifetime = %s", lifetime)
	}
	if service.RefreshCookieMaxAge() != DefaultRefreshTokenExpirationHours*3600 {
		t.Fatalf("cookie max-age = %d", service.RefreshCookieMaxAge())
	}
}

func TestValidateTokenRejectsNonHS256(t *testing.T) {
	service := NewTokenService(&memAuthTokenRepo{}).(*tokenService)
	claims := Claims{UserID: 1, Email: "ada@example.com", Role: "admin"}
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
	raw, err := token.SignedString([]byte(service.jwtSecret))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.ValidateToken(raw); err == nil {
		t.Fatal("expected non-HS256 token to be rejected")
	}
}

func TestRefuseDefaultTokenSecrets(t *testing.T) {
	t.Setenv("JWT_SECRET", "")
	t.Setenv("REFRESH_TOKEN_SECRET", "")
	if err := RefuseDefaultTokenSecrets(); err == nil {
		t.Fatal("expected empty secrets to fail")
	}
	t.Setenv("JWT_SECRET", DefaultJWTSecret)
	t.Setenv("REFRESH_TOKEN_SECRET", "refresh-not-default")
	if err := RefuseDefaultTokenSecrets(); err == nil {
		t.Fatal("expected default access secret to fail")
	}
	t.Setenv("JWT_SECRET", "access-not-default")
	t.Setenv("REFRESH_TOKEN_SECRET", "refresh-not-default")
	if err := RefuseDefaultTokenSecrets(); err != nil {
		t.Fatal(err)
	}
}
