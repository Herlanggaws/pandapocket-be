package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"panda-pocket/internal/application/identity"

	"github.com/gin-gonic/gin"
)

type sessionTokenService struct {
	refreshOwners map[string]int
	revokedUsers  []int
}

func (s *sessionTokenService) GenerateToken(context.Context, int, string, string) (string, string, error) {
	return "", "", nil
}
func (s *sessionTokenService) ValidateToken(string) (*identity.Claims, error) { return nil, nil }
func (s *sessionTokenService) AccessTokenActive(context.Context, string) (bool, error) {
	return false, nil
}
func (s *sessionTokenService) RefreshTokenOwner(_ context.Context, refreshToken string) (int, bool, error) {
	ownerID, found := s.refreshOwners[refreshToken]
	return ownerID, found, nil
}
func (s *sessionTokenService) ValidateRefreshToken(context.Context, string) (*identity.RefreshTokenClaims, error) {
	return nil, nil
}
func (s *sessionTokenService) GenerateAccessTokenResult(int, string, string) (string, error) {
	return "", nil
}
func (s *sessionTokenService) RevokeToken(context.Context, string) error { return nil }
func (s *sessionTokenService) RevokeAllForUser(_ context.Context, userID int) error {
	s.revokedUsers = append(s.revokedUsers, userID)
	return nil
}
func (s *sessionTokenService) RefreshCookieMaxAge() int                 { return 3600 }
func (s *sessionTokenService) CleanupExpiredToken(context.Context, string) error { return nil }

func TestLogoutWithoutRefreshRevokesCaller(t *testing.T) {
	tokens := &sessionTokenService{}
	status := serveLogout(tokens, 4, nil)
	if status != http.StatusOK {
		t.Fatalf("status=%d", status)
	}
	if len(tokens.revokedUsers) != 1 || tokens.revokedUsers[0] != 4 {
		t.Fatalf("revoked=%v", tokens.revokedUsers)
	}
}

func TestLogoutRejectsForeignRefreshWithoutRevoking(t *testing.T) {
	tokens := &sessionTokenService{refreshOwners: map[string]int{"other-refresh": 9}}
	body := []byte(`{"refresh_token":"other-refresh"}`)
	status := serveLogout(tokens, 4, body)
	if status != http.StatusForbidden {
		t.Fatalf("status=%d", status)
	}
	if len(tokens.revokedUsers) != 0 {
		t.Fatalf("revoked=%v", tokens.revokedUsers)
	}
}

func serveLogout(tokens *sessionTokenService, callerID int, body []byte) int {
	gin.SetMode(gin.TestMode)
	handler := NewIdentityHandlers(nil, nil, nil, nil, nil, nil, tokens, nil, nil, nil, nil, nil)
	router := gin.New()
	router.POST("/logout", func(c *gin.Context) {
		c.Set("user_id", callerID)
		handler.Logout(c)
	})
	var reader *bytes.Reader
	if body == nil {
		reader = bytes.NewReader(nil)
	} else {
		reader = bytes.NewReader(body)
	}
	request := httptest.NewRequest(http.MethodPost, "/logout", reader)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK && recorder.Body.Len() > 0 {
		var payload APIResponse
		_ = json.Unmarshal(recorder.Body.Bytes(), &payload)
	}
	return recorder.Code
}
