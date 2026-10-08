package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"panda-pocket/internal/application/identity"

	"github.com/gin-gonic/gin"
)

type stubTokenService struct {
	active    bool
	activeErr error
}

func (s *stubTokenService) GenerateToken(context.Context, int, string, string) (string, string, error) {
	return "", "", nil
}
func (s *stubTokenService) ValidateToken(string) (*identity.Claims, error) {
	return &identity.Claims{UserID: 1, Email: "user@example.com", Role: "user"}, nil
}
func (s *stubTokenService) AccessTokenActive(context.Context, string) (bool, error) {
	return s.active, s.activeErr
}
func (s *stubTokenService) RefreshTokenOwner(context.Context, string) (int, bool, error) {
	return 0, false, nil
}
func (s *stubTokenService) ValidateRefreshToken(context.Context, string) (*identity.RefreshTokenClaims, error) {
	return nil, nil
}
func (s *stubTokenService) GenerateAccessTokenResult(int, string, string) (string, error) {
	return "", nil
}
func (s *stubTokenService) RevokeToken(context.Context, string) error { return nil }
func (s *stubTokenService) RevokeAllForUser(context.Context, int) error { return nil }
func (s *stubTokenService) RefreshCookieMaxAge() int                 { return 0 }
func (s *stubTokenService) CleanupExpiredToken(context.Context, string) error { return nil }

func TestRevokedAccessTokenIsRejected(t *testing.T) {
	status := serveAuthed(&stubTokenService{active: false})
	if status != http.StatusUnauthorized {
		t.Fatalf("status=%d", status)
	}
}

func TestActiveAccessTokenIsAllowed(t *testing.T) {
	status := serveAuthed(&stubTokenService{active: true})
	if status != http.StatusNoContent {
		t.Fatalf("status=%d", status)
	}
}

func TestSessionLookupErrorIsServerError(t *testing.T) {
	status := serveAuthed(&stubTokenService{activeErr: context.DeadlineExceeded})
	if status != http.StatusInternalServerError {
		t.Fatalf("status=%d", status)
	}
}

func serveAuthed(tokens identity.TokenService) int {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/private", NewAuthMiddleware(tokens, nil).RequireAuth(), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder.Code
}
