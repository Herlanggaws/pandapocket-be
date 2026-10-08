package identity

import (
	"testing"
	"time"
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
