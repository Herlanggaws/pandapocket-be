package identity

import (
	"context"
	"testing"
	"time"

	domainIdentity "panda-pocket/internal/domain/identity"

	"github.com/google/uuid"
)

func TestChangePasswordRevokesSessions(t *testing.T) {
	users := newMemUserRepo()
	user := seedUser(t, users, "user@example.com", "old-password")
	revoker := &memTokenRevoker{}
	uc := NewChangePasswordUseCase(users, revoker)

	err := uc.Execute(context.Background(), ChangePasswordRequest{
		UserID:             user.ID().Value(),
		OldPassword:        "old-password",
		NewPassword:        "new-password",
		ConfirmNewPassword: "new-password",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(revoker.revokedUsers) != 1 || revoker.revokedUsers[0] != user.ID().Value() {
		t.Fatalf("revoked=%v", revoker.revokedUsers)
	}
}

func TestResetPasswordRevokesSessions(t *testing.T) {
	users := newMemUserRepo()
	user := seedUser(t, users, "user@example.com", "old-password")
	resetToken := domainIdentity.NewPasswordResetToken(user.ID(), "reset-token", time.Now().Add(time.Hour))
	revoker := &memTokenRevoker{}
	uc := NewResetPasswordUseCase(users, &oneResetRepo{token: resetToken}, revoker)

	_, err := uc.Execute(context.Background(), &ResetPasswordRequest{
		Token:              "reset-token",
		NewPassword:        "new-password",
		ConfirmNewPassword: "new-password",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(revoker.revokedUsers) != 1 || revoker.revokedUsers[0] != user.ID().Value() {
		t.Fatalf("revoked=%v", revoker.revokedUsers)
	}
}

type oneResetRepo struct {
	token *domainIdentity.PasswordResetToken
}

func (r *oneResetRepo) Save(context.Context, *domainIdentity.PasswordResetToken) error { return nil }
func (r *oneResetRepo) FindByToken(context.Context, string) (*domainIdentity.PasswordResetToken, error) {
	return r.token, nil
}
func (r *oneResetRepo) Delete(context.Context, uuid.UUID) error { return nil }
func (r *oneResetRepo) DeleteExpired(context.Context, time.Time) (int64, error) {
	return 0, nil
}
