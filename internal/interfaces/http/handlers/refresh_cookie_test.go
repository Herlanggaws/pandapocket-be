package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	appIdentity "panda-pocket/internal/application/identity"
	domainIdentity "panda-pocket/internal/domain/identity"

	"github.com/gin-gonic/gin"
)

type refreshSession struct {
	userID  int
	revoked bool
}

type rotatingTokenRepo struct {
	mu            sync.Mutex
	byRefresh     map[string]*refreshSession
	revokedTokens []string
}

func (r *rotatingTokenRepo) Save(_ context.Context, userID int, _, refreshToken string, _ int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.byRefresh == nil {
		r.byRefresh = map[string]*refreshSession{}
	}
	r.byRefresh[refreshToken] = &refreshSession{userID: userID}
	return nil
}

func (r *rotatingTokenRepo) FindByRefreshToken(_ context.Context, refreshToken string) (int, bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	session, ok := r.byRefresh[refreshToken]
	if !ok {
		return 0, false, nil
	}
	return session.userID, session.revoked, nil
}

func (r *rotatingTokenRepo) FindByAccessToken(context.Context, string) (bool, bool, error) {
	return false, false, nil
}
func (r *rotatingTokenRepo) DeleteByAccessToken(context.Context, string) error { return nil }
func (r *rotatingTokenRepo) Revoke(_ context.Context, refreshToken string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if session, ok := r.byRefresh[refreshToken]; ok {
		session.revoked = true
	}
	r.revokedTokens = append(r.revokedTokens, refreshToken)
	return nil
}
func (r *rotatingTokenRepo) RevokeAllForUser(context.Context, int) error { return nil }
func (r *rotatingTokenRepo) DeleteExpired(context.Context, time.Time) (int64, error) {
	return 0, nil
}

type cookieUserRepo struct {
	users map[int]*domainIdentity.User
}

func (r *cookieUserRepo) Save(context.Context, *domainIdentity.User) error { return nil }
func (r *cookieUserRepo) Update(context.Context, *domainIdentity.User) error {
	return nil
}
func (r *cookieUserRepo) FindByID(_ context.Context, id domainIdentity.UserID) (*domainIdentity.User, error) {
	user, ok := r.users[id.Value()]
	if !ok {
		return nil, errors.New("not found")
	}
	return user, nil
}
func (r *cookieUserRepo) FindByEmail(context.Context, domainIdentity.Email) (*domainIdentity.User, error) {
	return nil, errors.New("not found")
}
func (r *cookieUserRepo) FindAll(context.Context) ([]*domainIdentity.User, error) {
	return nil, nil
}
func (r *cookieUserRepo) Delete(context.Context, domainIdentity.UserID) error { return nil }
func (r *cookieUserRepo) ExistsByEmail(context.Context, domainIdentity.Email) (bool, error) {
	return false, nil
}
func (r *cookieUserRepo) ExistsActive(context.Context, domainIdentity.UserID) (bool, error) {
	return true, nil
}
func (r *cookieUserRepo) SoftDelete(context.Context, domainIdentity.UserID, string, time.Time) error {
	return nil
}
func (r *cookieUserRepo) ListDueForPurge(context.Context, time.Time) ([]domainIdentity.UserID, error) {
	return nil, nil
}

func TestRefreshBodyRotatesAndIgnoresForeignCookie(t *testing.T) {
	handler, tokens, users := newRefreshHandler(t)
	ownerToken := issueRefresh(t, tokens, users, 4, "owner@example.com")
	otherToken := issueRefresh(t, tokens, users, 9, "other@example.com")

	body := []byte(`{"refresh_token":"` + ownerToken + `"}`)
	recorder := serveRefresh(handler, body, otherToken, "")

	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	payload := decodeSuccess(t, recorder)
	newRefresh, _ := payload["refresh_token"].(string)
	if newRefresh == "" {
		t.Fatal("missing refresh token")
	}
	if !containsToken(tokens.revokedTokens, domainIdentity.SessionTokenHash(ownerToken)) {
		t.Fatal("body refresh token was not revoked")
	}
	if containsToken(tokens.revokedTokens, domainIdentity.SessionTokenHash(otherToken)) {
		t.Fatal("cookie refresh token was revoked")
	}
	cookie := refreshCookie(t, recorder)
	if cookie.Value != newRefresh {
		t.Fatalf("cookie=%q response=%q", cookie.Value, newRefresh)
	}
	if cookie.Secure {
		t.Fatal("cookie was Secure on plain HTTP")
	}
}

func TestRefreshCookieIssuesNewPair(t *testing.T) {
	handler, tokens, users := newRefreshHandler(t)
	cookieToken := issueRefresh(t, tokens, users, 4, "owner@example.com")

	recorder := serveRefresh(handler, nil, cookieToken, "https")
	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", recorder.Code, recorder.Body.String())
	}
	payload := decodeSuccess(t, recorder)
	newRefresh, _ := payload["refresh_token"].(string)
	access, _ := payload["token"].(string)
	if newRefresh == "" || access == "" {
		t.Fatalf("pair was not issued: access=%q refresh=%q", access, newRefresh)
	}
	if !containsToken(tokens.revokedTokens, domainIdentity.SessionTokenHash(cookieToken)) {
		t.Fatal("cookie refresh token was not revoked")
	}
	cookie := refreshCookie(t, recorder)
	if cookie.Value != newRefresh || !cookie.HttpOnly || !cookie.Secure {
		t.Fatalf("cookie=%+v", cookie)
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("samesite=%v", cookie.SameSite)
	}
}

func TestRefreshCredentialPrefersBody(t *testing.T) {
	if got := refreshCredential("body", "cookie"); got != "body" {
		t.Fatalf("got %q", got)
	}
	if got := refreshCredential("", "cookie"); got != "cookie" {
		t.Fatalf("got %q", got)
	}
}

func newRefreshHandler(t *testing.T) (*IdentityHandlers, *rotatingTokenRepo, map[int]*domainIdentity.User) {
	t.Helper()
	repo := &rotatingTokenRepo{}
	tokenService := appIdentity.NewTokenService(repo)
	users := map[int]*domainIdentity.User{}
	handler := NewIdentityHandlers(
		nil, nil, nil, nil, nil,
		appIdentity.NewRefreshTokenUseCase(domainIdentity.NewUserService(&cookieUserRepo{users: users}), tokenService),
		tokenService,
		nil, nil, nil, nil, nil,
	)
	return handler, repo, users
}

func issueRefresh(t *testing.T, tokens *rotatingTokenRepo, users map[int]*domainIdentity.User, userID int, email string) string {
	t.Helper()
	if _, ok := users[userID]; !ok {
		users[userID] = mustUser(t, userID, email)
	}
	service := appIdentity.NewTokenService(tokens)
	_, refresh, err := service.GenerateToken(context.Background(), userID, email, "user")
	if err != nil {
		t.Fatal(err)
	}
	return refresh
}

func mustUser(t *testing.T, id int, email string) *domainIdentity.User {
	t.Helper()
	emailVO, err := domainIdentity.NewEmail(email)
	if err != nil {
		t.Fatal(err)
	}
	hash, err := domainIdentity.NewPasswordHashFromPlain("password123")
	if err != nil {
		t.Fatal(err)
	}
	role, err := domainIdentity.NewRole("user")
	if err != nil {
		t.Fatal(err)
	}
	return domainIdentity.NewUser(domainIdentity.NewUserID(id), emailVO, hash, role)
}

func serveRefresh(handler *IdentityHandlers, body []byte, cookieToken, forwardedProto string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/refresh", handler.RefreshToken)
	reader := bytes.NewReader(body)
	request := httptest.NewRequest(http.MethodPost, "/refresh", reader)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if forwardedProto != "" {
		request.Header.Set("X-Forwarded-Proto", forwardedProto)
	}
	if cookieToken != "" {
		request.AddCookie(&http.Cookie{Name: refreshCookieName, Value: cookieToken})
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func decodeSuccess(t *testing.T, recorder *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var payload APIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	data, ok := payload.Data.(map[string]any)
	if !ok {
		t.Fatalf("data=%T", payload.Data)
	}
	return data
}

func containsToken(tokens []string, want string) bool {
	for _, token := range tokens {
		if token == want {
			return true
		}
	}
	return false
}

func refreshCookie(t *testing.T, recorder *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, cookie := range recorder.Result().Cookies() {
		if cookie.Name == refreshCookieName {
			return cookie
		}
	}
	t.Fatal("refresh cookie missing")
	return nil
}
