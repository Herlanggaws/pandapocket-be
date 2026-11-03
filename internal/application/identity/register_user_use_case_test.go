package identity

import (
	"context"
	"testing"

	"panda-pocket/internal/domain/billing"
	domainIdentity "panda-pocket/internal/domain/identity"
	"panda-pocket/internal/infrastructure/analytics"
)

type stubTokenService struct{}

func (stubTokenService) GenerateToken(context.Context, int, string, string) (string, string, error) {
	return "access", "refresh", nil
}
func (stubTokenService) ValidateToken(string) (*Claims, error) { return nil, nil }
func (stubTokenService) AccessTokenActive(context.Context, string) (bool, error) {
	return true, nil
}
func (stubTokenService) RefreshTokenOwner(context.Context, string) (int, bool, error) {
	return 0, false, nil
}
func (stubTokenService) ValidateRefreshToken(context.Context, string) (*RefreshTokenClaims, error) {
	return nil, nil
}
func (stubTokenService) GenerateAccessTokenResult(int, string, string) (string, error) {
	return "access", nil
}
func (stubTokenService) RevokeToken(context.Context, string) error              { return nil }
func (stubTokenService) RevokeAllForUser(context.Context, int) error            { return nil }
func (stubTokenService) RefreshCookieMaxAge() int                               { return 0 }
func (stubTokenService) CleanupExpiredToken(context.Context, string) error      { return nil }

type memoryTrialSubs struct {
	saved int
}

func (m *memoryTrialSubs) Save(context.Context, *billing.Subscription) error {
	m.saved++
	return nil
}
func (m *memoryTrialSubs) FindByUserID(context.Context, int) (*billing.Subscription, error) {
	return nil, billing.ErrNotFound
}
func (m *memoryTrialSubs) ListAll(context.Context) ([]*billing.Subscription, error) {
	return nil, nil
}

func TestRegisterUserCapturesSignupAndTrial(t *testing.T) {
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
	subs := &memoryTrialSubs{}
	uc := NewRegisterUserUseCase(
		domainIdentity.NewUserService(users),
		stubTokenService{},
		subs,
	)

	response, err := uc.Execute(context.Background(), RegisterUserRequest{
		Email:    "new@example.com",
		Password: "password-ok",
	})
	if err != nil {
		t.Fatal(err)
	}
	if response.UserID == 0 || subs.saved != 1 {
		t.Fatalf("response=%+v trials=%d", response, subs.saved)
	}
	if len(captures) != 2 {
		t.Fatalf("expected signup_completed and trial_started, got %+v", captures)
	}
	if captures[0].event != "signup_completed" || captures[0].userID != response.UserID {
		t.Fatalf("expected signup_completed, got %+v", captures[0])
	}
	if captures[0].props["method"] != "password" {
		t.Fatalf("unexpected signup method: %+v", captures[0].props)
	}
	if _, hasEmail := captures[0].props["email"]; hasEmail {
		t.Fatal("signup_completed must not include email")
	}
	if captures[1].event != "trial_started" || captures[1].userID != response.UserID {
		t.Fatalf("expected trial_started, got %+v", captures[1])
	}
	if captures[1].props["plan"] != "pro" || captures[1].props["trial_days"] != billing.TrialDurationDays {
		t.Fatalf("unexpected trial_started properties: %+v", captures[1].props)
	}
}
