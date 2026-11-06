package identity

import (
	"context"
	"testing"
	"time"

	domainIdentity "panda-pocket/internal/domain/identity"
)

func TestLoginRecordsLastLoginAt(t *testing.T) {
	users := newMemUserRepo()
	user := seedUser(t, users, "login@example.com", "password-ok")
	uc := NewLoginUserUseCase(domainIdentity.NewUserService(users), users, stubTokenService{})

	before := time.Now().UTC().Add(-time.Second)
	response, err := uc.Execute(context.Background(), LoginUserRequest{
		Email:    "login@example.com",
		Password: "password-ok",
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.UserID != user.ID().Value() {
		t.Fatalf("user_id=%d", response.UserID)
	}

	at, ok := users.lastLoginAt[user.ID().Value()]
	if !ok {
		t.Fatal("expected last_login_at to be recorded")
	}
	if at.Before(before) {
		t.Fatalf("last_login_at too early: %v", at)
	}
}

func TestLoginDoesNotFailWhenLastLoginUpdateFails(t *testing.T) {
	users := newMemUserRepo()
	_ = seedUser(t, users, "login-fail@example.com", "password-ok")
	users.failLastLogin = true
	uc := NewLoginUserUseCase(domainIdentity.NewUserService(users), users, stubTokenService{})

	response, err := uc.Execute(context.Background(), LoginUserRequest{
		Email:    "login-fail@example.com",
		Password: "password-ok",
	})
	if err != nil {
		t.Fatalf("login must succeed even if last_login_at update fails: %v", err)
	}
	if response.Token == "" {
		t.Fatal("expected token")
	}
}
