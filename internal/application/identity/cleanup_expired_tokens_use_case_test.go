package identity

import (
	"context"
	"testing"
	"time"

	domainIdentity "panda-pocket/internal/domain/identity"

	"github.com/google/uuid"
)

type memAuthTokenRepo struct {
	deleted int64
	before  time.Time
}

func (r *memAuthTokenRepo) Save(context.Context, int, string, string, int64) error { return nil }
func (r *memAuthTokenRepo) FindByRefreshToken(context.Context, string) (int, bool, error) {
	return 0, false, nil
}
func (r *memAuthTokenRepo) DeleteByAccessToken(context.Context, string) error { return nil }
func (r *memAuthTokenRepo) Revoke(context.Context, string) error              { return nil }
func (r *memAuthTokenRepo) RevokeAllForUser(context.Context, int) error       { return nil }
func (r *memAuthTokenRepo) DeleteExpired(_ context.Context, before time.Time) (int64, error) {
	r.before = before
	return r.deleted, nil
}

type memPasswordResetRepo struct {
	deleted int64
}

func (r *memPasswordResetRepo) Save(context.Context, *domainIdentity.PasswordResetToken) error {
	return nil
}
func (r *memPasswordResetRepo) FindByToken(context.Context, string) (*domainIdentity.PasswordResetToken, error) {
	return nil, nil
}
func (r *memPasswordResetRepo) Delete(context.Context, uuid.UUID) error { return nil }
func (r *memPasswordResetRepo) DeleteExpired(context.Context, time.Time) (int64, error) {
	return r.deleted, nil
}

func TestCleanupExpiredTokensUseCase(t *testing.T) {
	auth := &memAuthTokenRepo{deleted: 3}
	resets := &memPasswordResetRepo{deleted: 2}
	uc := NewCleanupExpiredTokensUseCase(auth, resets)
	fixed := time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC)
	uc.now = func() time.Time { return fixed }

	result, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SessionTokensDeleted != 3 || result.PasswordResetTokensDeleted != 2 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if !auth.before.Equal(fixed) {
		t.Fatalf("expected cutoff %v, got %v", fixed, auth.before)
	}
}
