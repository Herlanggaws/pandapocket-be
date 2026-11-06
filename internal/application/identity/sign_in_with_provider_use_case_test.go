package identity

import (
	"context"
	"testing"
	"time"

	"panda-pocket/internal/domain/billing"
	domainIdentity "panda-pocket/internal/domain/identity"
	"panda-pocket/internal/infrastructure/analytics"
)

type fakeVerifier struct {
	identity VerifiedSignIn
	err      error
}

func (f fakeVerifier) Verify(context.Context, string, string, string) (VerifiedSignIn, error) {
	return f.identity, f.err
}

type fakeSessions struct{}

func (fakeSessions) GenerateToken(context.Context, int, string, string) (string, string, error) {
	return "access", "refresh", nil
}

type fakeTrials struct {
	saved int
}

func (f *fakeTrials) Save(context.Context, *billing.Subscription) error {
	f.saved++
	return nil
}

type memIdentities struct {
	rows []domainIdentity.ExternalIdentity
}

func (m *memIdentities) FindByProviderSubject(_ context.Context, provider, subject string) (*domainIdentity.ExternalIdentity, error) {
	for _, row := range m.rows {
		if row.Provider == provider && row.Subject == subject {
			copy := row
			return &copy, nil
		}
	}
	return nil, nil
}

func (m *memIdentities) Save(_ context.Context, identity domainIdentity.ExternalIdentity) error {
	m.rows = append(m.rows, identity)
	return nil
}

func (m *memIdentities) DeleteByProviderSubject(_ context.Context, provider, subject string) error {
	kept := m.rows[:0]
	for _, row := range m.rows {
		if row.Provider == provider && row.Subject == subject {
			continue
		}
		kept = append(kept, row)
	}
	m.rows = kept
	return nil
}

func (m *memIdentities) DeleteByUserID(_ context.Context, userID int) error {
	kept := m.rows[:0]
	for _, row := range m.rows {
		if row.UserID == userID {
			continue
		}
		kept = append(kept, row)
	}
	m.rows = kept
	return nil
}

func TestSignInLinksExistingEmail(t *testing.T) {
	var captures []string
	analytics.SetTestCapture(func(_ int, event string, _ map[string]interface{}) {
		captures = append(captures, event)
	})
	t.Cleanup(func() { analytics.SetTestCapture(nil) })

	users := newMemUserRepo()
	existing := seedUser(t, users, "ada@example.com", "password-ok")
	identities := &memIdentities{}
	trials := &fakeTrials{}
	uc := NewSignInWithProviderUseCase(
		fakeVerifier{identity: VerifiedSignIn{Provider: "google", Subject: "sub-1", Email: "Ada@example.com", EmailVerified: true}},
		users,
		identities,
		domainIdentity.NewUserService(users),
		fakeSessions{},
		trials,
	)

	response, err := uc.Execute(context.Background(), SignInWithProviderRequest{Provider: "google", IDToken: "token", Nonce: "nonce"})
	if err != nil {
		t.Fatal(err)
	}
	if response.UserID != existing.ID().Value() || !response.HasPassword {
		t.Fatalf("linked user = %+v", response)
	}
	if trials.saved != 0 || len(identities.rows) != 1 {
		t.Fatalf("trials=%d identities=%d", trials.saved, len(identities.rows))
	}
	if len(captures) != 0 {
		t.Fatalf("existing-user login must not fire signup/trial events, got %v", captures)
	}
	if _, ok := users.lastLoginAt[existing.ID().Value()]; !ok {
		t.Fatal("expected last_login_at on SSO login for existing user")
	}
}

func TestSignInRegistersWithoutPassword(t *testing.T) {
	var captures []struct {
		userID int
		event  string
		props  map[string]interface{}
	}
	analytics.SetTestCapture(func(userID int, event string, properties map[string]interface{}) {
		captures = append(captures, struct {
			userID int
			event  string
			props  map[string]interface{}
		}{userID, event, properties})
	})
	t.Cleanup(func() { analytics.SetTestCapture(nil) })

	users := newMemUserRepo()
	identities := &memIdentities{}
	trials := &fakeTrials{}
	uc := NewSignInWithProviderUseCase(
		fakeVerifier{identity: VerifiedSignIn{Provider: "google", Subject: "sub-2", Email: "new@example.com", EmailVerified: true}},
		users,
		identities,
		domainIdentity.NewUserService(users),
		fakeSessions{},
		trials,
	)

	response, err := uc.Execute(context.Background(), SignInWithProviderRequest{Provider: "google", IDToken: "token", Nonce: "nonce"})
	if err != nil {
		t.Fatal(err)
	}
	if response.HasPassword || response.Email != "new@example.com" || trials.saved != 1 {
		t.Fatalf("response=%+v trials=%d", response, trials.saved)
	}
	if len(captures) != 2 {
		t.Fatalf("expected signup_completed and trial_started, got %+v", captures)
	}
	if captures[0].event != "signup_completed" || captures[0].userID != response.UserID {
		t.Fatalf("expected signup_completed, got %+v", captures[0])
	}
	if captures[0].props["method"] != "google" {
		t.Fatalf("unexpected signup method: %+v", captures[0].props)
	}
	if captures[1].event != "trial_started" || captures[1].props["trial_days"] != billing.TrialDurationDays {
		t.Fatalf("unexpected trial_started: %+v", captures[1])
	}
	if _, ok := users.lastLoginAt[response.UserID]; !ok {
		t.Fatal("expected last_login_at on SSO signup")
	}
}

func TestSignInRemovesStaleIdentityAndLinksLiveUser(t *testing.T) {
	users := newMemUserRepo()
	stale := seedUser(t, users, "ada@example.com", "password-ok")
	deletedAt := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	if err := users.SoftDelete(context.Background(), stale.ID(), mangledDeletedEmail(stale.ID().Value(), deletedAt), deletedAt); err != nil {
		t.Fatal(err)
	}
	live := seedUser(t, users, "ada@example.com", "password-ok")

	identities := &memIdentities{rows: []domainIdentity.ExternalIdentity{{
		UserID:   stale.ID().Value(),
		Provider: "google",
		Subject:  "sub-stale",
		Email:    "ada@example.com",
	}}}
	trials := &fakeTrials{}
	uc := NewSignInWithProviderUseCase(
		fakeVerifier{identity: VerifiedSignIn{Provider: "google", Subject: "sub-stale", Email: "ada@example.com", EmailVerified: true}},
		users,
		identities,
		domainIdentity.NewUserService(users),
		fakeSessions{},
		trials,
	)

	response, err := uc.Execute(context.Background(), SignInWithProviderRequest{Provider: "google", IDToken: "token", Nonce: "nonce"})
	if err != nil {
		t.Fatal(err)
	}
	if response.UserID != live.ID().Value() {
		t.Fatalf("expected live user %d, got %d", live.ID().Value(), response.UserID)
	}
	if trials.saved != 0 {
		t.Fatalf("expected no new trial, got %d", trials.saved)
	}
	if len(identities.rows) != 1 {
		t.Fatalf("expected one identity, got %d", len(identities.rows))
	}
	got := identities.rows[0]
	if got.UserID != live.ID().Value() || got.Provider != "google" || got.Subject != "sub-stale" {
		t.Fatalf("identity = %+v", got)
	}
}
