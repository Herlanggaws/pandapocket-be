package identity

import (
	"context"
	"testing"
	"time"

	domainIdentity "panda-pocket/internal/domain/identity"
)

type storedSession struct {
	userID  int
	revoked bool
}

type recordingTokenRepo struct {
	byAccess  map[string]*storedSession
	byRefresh map[string]*storedSession
}

func (r *recordingTokenRepo) Save(_ context.Context, userID int, accessToken, refreshToken string, _ int64) error {
	if r.byAccess == nil {
		r.byAccess = map[string]*storedSession{}
		r.byRefresh = map[string]*storedSession{}
	}
	session := &storedSession{userID: userID}
	r.byAccess[accessToken] = session
	r.byRefresh[refreshToken] = session
	return nil
}

func (r *recordingTokenRepo) FindByRefreshToken(_ context.Context, refreshToken string) (int, bool, error) {
	session, ok := r.byRefresh[refreshToken]
	if !ok {
		return 0, false, nil
	}
	return session.userID, session.revoked, nil
}

func (r *recordingTokenRepo) FindByAccessToken(_ context.Context, accessToken string) (bool, bool, error) {
	session, ok := r.byAccess[accessToken]
	if !ok {
		return false, false, nil
	}
	return true, session.revoked, nil
}

func (r *recordingTokenRepo) DeleteByAccessToken(_ context.Context, accessToken string) error {
	delete(r.byAccess, accessToken)
	return nil
}

func (r *recordingTokenRepo) Revoke(_ context.Context, refreshToken string) error {
	if session, ok := r.byRefresh[refreshToken]; ok {
		session.revoked = true
	}
	return nil
}

func (r *recordingTokenRepo) RevokeAllForUser(context.Context, int) error { return nil }
func (r *recordingTokenRepo) DeleteExpired(context.Context, time.Time) (int64, error) {
	return 0, nil
}

func TestSessionTokensAreStoredAsHashes(t *testing.T) {
	repo := &recordingTokenRepo{}
	service := NewTokenService(repo)
	accessToken, refreshToken, err := service.GenerateToken(context.Background(), 7, "ada@example.com", "user")
	if err != nil {
		t.Fatal(err)
	}
	accessHash := domainIdentity.SessionTokenHash(accessToken)
	refreshHash := domainIdentity.SessionTokenHash(refreshToken)
	if _, ok := repo.byAccess[accessToken]; ok {
		t.Fatal("raw access token was stored")
	}
	if _, ok := repo.byRefresh[refreshHash]; !ok || !domainIdentity.IsSessionTokenHash(accessHash) {
		t.Fatal("stored session is not a hash")
	}

	active, err := service.AccessTokenActive(context.Background(), accessToken)
	if err != nil || !active {
		t.Fatalf("active=%v err=%v", active, err)
	}
	ownerID, found, err := service.RefreshTokenOwner(context.Background(), refreshToken)
	if err != nil || !found || ownerID != 7 {
		t.Fatalf("owner=%d found=%v err=%v", ownerID, found, err)
	}
	if _, err := service.ValidateRefreshToken(context.Background(), refreshToken); err != nil {
		t.Fatal(err)
	}
	if err := service.RevokeToken(context.Background(), refreshToken); err != nil {
		t.Fatal(err)
	}
	if _, err := service.ValidateRefreshToken(context.Background(), refreshToken); err == nil {
		t.Fatal("revoked refresh token was accepted")
	}
	if err := service.CleanupExpiredToken(context.Background(), accessToken); err != nil {
		t.Fatal(err)
	}
	if _, ok := repo.byAccess[accessHash]; ok {
		t.Fatal("hashed access token was not deleted")
	}
}
