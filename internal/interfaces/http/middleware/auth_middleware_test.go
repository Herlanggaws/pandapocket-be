package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"panda-pocket/internal/application/identity"
	domainIdentity "panda-pocket/internal/domain/identity"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
func (s *stubTokenService) RevokeToken(context.Context, string) error         { return nil }
func (s *stubTokenService) RevokeAllForUser(context.Context, int) error       { return nil }
func (s *stubTokenService) RefreshCookieMaxAge() int                          { return 0 }
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

func TestRequireRoleUsesDatabaseRole(t *testing.T) {
	users := &roleLookup{user: testUser(t, "user")}
	status := serveRole(users, "admin")
	if status != http.StatusForbidden {
		t.Fatalf("status=%d", status)
	}
}

func TestRequireRoleAllowsDatabaseAdmin(t *testing.T) {
	users := &roleLookup{user: testUser(t, "admin")}
	status := serveRole(users, "admin")
	if status != http.StatusNoContent {
		t.Fatalf("status=%d", status)
	}
}

func TestRequireRoleMissingUserIsUnauthorized(t *testing.T) {
	status := serveRole(&roleLookup{err: gorm.ErrRecordNotFound}, "admin")
	if status != http.StatusUnauthorized {
		t.Fatalf("status=%d", status)
	}
}

func TestRequireRoleDatabaseErrorIsServerError(t *testing.T) {
	status := serveRole(&roleLookup{err: context.DeadlineExceeded}, "admin")
	if status != http.StatusInternalServerError {
		t.Fatalf("status=%d", status)
	}
}

type roleLookup struct {
	user *domainIdentity.User
	err  error
}

func (r *roleLookup) ExistsActive(context.Context, domainIdentity.UserID) (bool, error) {
	return true, nil
}

func (r *roleLookup) FindByID(context.Context, domainIdentity.UserID) (*domainIdentity.User, error) {
	return r.user, r.err
}

func testUser(t *testing.T, role string) *domainIdentity.User {
	t.Helper()
	email, err := domainIdentity.NewEmail("ada@example.com")
	if err != nil {
		t.Fatal(err)
	}
	parsedRole, err := domainIdentity.NewRole(role)
	if err != nil {
		t.Fatal(err)
	}
	return domainIdentity.NewUser(domainIdentity.NewUserID(1), email, domainIdentity.NewPasswordHash("hash"), parsedRole)
}

func serveRole(users *roleLookup, required string) int {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	auth := NewAuthMiddleware(&stubTokenService{active: true}, users)
	router.GET("/admin", auth.RequireAuth(), auth.RequireRole(required), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})
	request := httptest.NewRequest(http.MethodGet, "/admin", nil)
	request.Header.Set("Authorization", "Bearer access-token")
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder.Code
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
