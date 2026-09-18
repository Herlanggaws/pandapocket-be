package identity

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
)

type memChallengeRepo struct {
	mu         sync.Mutex
	byUser     map[uint]*AccountResetChallenge
	deletedIDs map[uuid.UUID]bool
}

func newMemChallengeRepo() *memChallengeRepo {
	return &memChallengeRepo{
		byUser:     map[uint]*AccountResetChallenge{},
		deletedIDs: map[uuid.UUID]bool{},
	}
}

func (r *memChallengeRepo) DeleteByUserID(ctx context.Context, userID uint) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.byUser, userID)
	return nil
}

func (r *memChallengeRepo) Save(ctx context.Context, challenge *AccountResetChallenge) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	copy := *challenge
	r.byUser[challenge.UserID] = &copy
	return nil
}

func (r *memChallengeRepo) FindLatestByUserID(ctx context.Context, userID uint) (*AccountResetChallenge, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	challenge, ok := r.byUser[userID]
	if !ok {
		return nil, nil
	}
	if r.deletedIDs[challenge.ID] {
		return nil, nil
	}
	copy := *challenge
	return &copy, nil
}

func (r *memChallengeRepo) IncrementAttempts(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, challenge := range r.byUser {
		if challenge.ID == id {
			challenge.AttemptCount++
			return nil
		}
	}
	return nil
}

func (r *memChallengeRepo) DeleteByID(ctx context.Context, id uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.deletedIDs[id] = true
	for userID, challenge := range r.byUser {
		if challenge.ID == id {
			delete(r.byUser, userID)
		}
	}
	return nil
}

type memWiper struct {
	mu          sync.Mutex
	wipedUserID uint
	wipeCount   int
	err         error
}

func (w *memWiper) WipeUserData(ctx context.Context, userID uint) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.err != nil {
		return w.err
	}
	w.wipedUserID = userID
	w.wipeCount++
	return nil
}

func TestCreateChallengeReturnsConfirmationText(t *testing.T) {
	repo := newMemChallengeRepo()
	wiper := &memWiper{}
	uc := NewResetAccountDataUseCase(repo, wiper)
	fixedNow := time.Date(2026, 9, 18, 10, 0, 0, 0, time.UTC)
	uc.now = func() time.Time { return fixedNow }
	uc.generateCode = func() (string, error) { return "AFkLJdl9879x", nil }

	resp, err := uc.CreateChallenge(context.Background(), 42)
	if err != nil {
		t.Fatalf("CreateChallenge: %v", err)
	}
	if resp.ConfirmationText != "AFkLJdl9879x" {
		t.Fatalf("confirmation_text = %q", resp.ConfirmationText)
	}
	if !resp.ExpiresAt.Equal(fixedNow.Add(accountResetChallengeTTL)) {
		t.Fatalf("expires_at = %v", resp.ExpiresAt)
	}

	stored, err := repo.FindLatestByUserID(context.Background(), 42)
	if err != nil || stored == nil {
		t.Fatalf("expected stored challenge, err=%v", err)
	}
	if stored.CodeHash != HashAccountResetConfirmation("AFkLJdl9879x") {
		t.Fatalf("stored hash mismatch")
	}
}

func TestResetMismatchReturnsErrorAndDoesNotWipe(t *testing.T) {
	repo := newMemChallengeRepo()
	wiper := &memWiper{}
	uc := NewResetAccountDataUseCase(repo, wiper)
	fixedNow := time.Date(2026, 9, 18, 10, 0, 0, 0, time.UTC)
	uc.now = func() time.Time { return fixedNow }
	uc.generateCode = func() (string, error) { return "CorrectCode12", nil }

	if _, err := uc.CreateChallenge(context.Background(), 7); err != nil {
		t.Fatalf("CreateChallenge: %v", err)
	}

	_, err := uc.Execute(context.Background(), ResetAccountDataRequest{
		UserID:           7,
		ConfirmationText: "WrongCode!!!!",
	})
	if err != ErrAccountResetConfirmationMismatch {
		t.Fatalf("expected mismatch, got %v", err)
	}
	if wiper.wipeCount != 0 {
		t.Fatalf("wipe should not run on mismatch")
	}
}

func TestResetExpiredChallenge(t *testing.T) {
	repo := newMemChallengeRepo()
	wiper := &memWiper{}
	uc := NewResetAccountDataUseCase(repo, wiper)
	createdAt := time.Date(2026, 9, 18, 10, 0, 0, 0, time.UTC)
	uc.now = func() time.Time { return createdAt }
	uc.generateCode = func() (string, error) { return "ExpiredCode1", nil }

	if _, err := uc.CreateChallenge(context.Background(), 9); err != nil {
		t.Fatalf("CreateChallenge: %v", err)
	}

	uc.now = func() time.Time { return createdAt.Add(accountResetChallengeTTL + time.Second) }
	_, err := uc.Execute(context.Background(), ResetAccountDataRequest{
		UserID:           9,
		ConfirmationText: "ExpiredCode1",
	})
	if err != ErrAccountResetChallengeExpired {
		t.Fatalf("expected expired, got %v", err)
	}
	if wiper.wipeCount != 0 {
		t.Fatalf("wipe should not run when expired")
	}
}

func TestResetSuccessWipesOnlyAfterMatch(t *testing.T) {
	repo := newMemChallengeRepo()
	wiper := &memWiper{}
	uc := NewResetAccountDataUseCase(repo, wiper)
	fixedNow := time.Date(2026, 9, 18, 10, 0, 0, 0, time.UTC)
	uc.now = func() time.Time { return fixedNow }
	uc.generateCode = func() (string, error) { return "MatchCode1234", nil }

	if _, err := uc.CreateChallenge(context.Background(), 11); err != nil {
		t.Fatalf("CreateChallenge: %v", err)
	}

	resp, err := uc.Execute(context.Background(), ResetAccountDataRequest{
		UserID:           11,
		ConfirmationText: "MatchCode1234",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if resp.Message == "" {
		t.Fatalf("expected success message")
	}
	if wiper.wipeCount != 1 || wiper.wipedUserID != 11 {
		t.Fatalf("expected wipe for user 11, got count=%d user=%d", wiper.wipeCount, wiper.wipedUserID)
	}
}

func TestResetTooManyAttemptsInvalidatesChallenge(t *testing.T) {
	repo := newMemChallengeRepo()
	wiper := &memWiper{}
	uc := NewResetAccountDataUseCase(repo, wiper)
	fixedNow := time.Date(2026, 9, 18, 10, 0, 0, 0, time.UTC)
	uc.now = func() time.Time { return fixedNow }
	uc.generateCode = func() (string, error) { return "LimitCode1234", nil }

	if _, err := uc.CreateChallenge(context.Background(), 15); err != nil {
		t.Fatalf("CreateChallenge: %v", err)
	}

	for i := 0; i < accountResetMaxFailedAttempts-1; i++ {
		_, err := uc.Execute(context.Background(), ResetAccountDataRequest{
			UserID:           15,
			ConfirmationText: "bad",
		})
		if err != ErrAccountResetConfirmationMismatch {
			t.Fatalf("attempt %d: expected mismatch, got %v", i+1, err)
		}
	}

	_, err := uc.Execute(context.Background(), ResetAccountDataRequest{
		UserID:           15,
		ConfirmationText: "bad",
	})
	if err != ErrAccountResetTooManyAttempts {
		t.Fatalf("expected too many attempts, got %v", err)
	}

	_, err = uc.Execute(context.Background(), ResetAccountDataRequest{
		UserID:           15,
		ConfirmationText: "LimitCode1234",
	})
	if err != ErrAccountResetChallengeNotFound {
		t.Fatalf("expected challenge cleared, got %v", err)
	}
	if wiper.wipeCount != 0 {
		t.Fatalf("wipe should not run after lockout")
	}
}

func TestHashAccountResetConfirmationIsDeterministic(t *testing.T) {
	a := HashAccountResetConfirmation("AFkLJdl9879x")
	b := HashAccountResetConfirmation("AFkLJdl9879x")
	c := HashAccountResetConfirmation("different")
	if a != b {
		t.Fatalf("hash should be deterministic")
	}
	if a == c {
		t.Fatalf("different inputs should hash differently")
	}
}
