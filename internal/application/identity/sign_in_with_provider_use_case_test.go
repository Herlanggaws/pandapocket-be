package identity

import (
	"context"
	"testing"

	"panda-pocket/internal/domain/billing"
	domainIdentity "panda-pocket/internal/domain/identity"
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

func TestSignInLinksExistingEmail(t *testing.T) {
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
}

func TestSignInRegistersWithoutPassword(t *testing.T) {
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
}
