package identity

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"math/big"
	"time"

	"github.com/google/uuid"
)

const (
	accountResetConfirmationLength = 12
	accountResetChallengeTTL       = 5 * time.Minute
	accountResetMaxFailedAttempts  = 5
)

var (
	ErrAccountResetChallengeNotFound    = errors.New("reset challenge not found; request a new confirmation code")
	ErrAccountResetChallengeExpired     = errors.New("reset challenge expired; request a new confirmation code")
	ErrAccountResetConfirmationMismatch = errors.New("confirmation text does not match")
	ErrAccountResetTooManyAttempts      = errors.New("too many failed attempts; request a new confirmation code")
)

// AccountResetChallenge is the application-level challenge record.
type AccountResetChallenge struct {
	ID           uuid.UUID
	UserID       uint
	CodeHash     string
	AttemptCount int
	ExpiresAt    time.Time
	CreatedAt    time.Time
}

type accountResetChallengeRepository interface {
	DeleteByUserID(ctx context.Context, userID uint) error
	Save(ctx context.Context, challenge *AccountResetChallenge) error
	FindLatestByUserID(ctx context.Context, userID uint) (*AccountResetChallenge, error)
	IncrementAttempts(ctx context.Context, id uuid.UUID) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type userDataWiper interface {
	WipeUserData(ctx context.Context, userID uint) error
}

type CreateAccountResetChallengeResponse struct {
	ConfirmationText string    `json:"confirmation_text"`
	ExpiresAt        time.Time `json:"expires_at"`
}

type ResetAccountDataRequest struct {
	UserID           int    `json:"-"`
	ConfirmationText string `json:"confirmation_text" binding:"required"`
}

type ResetAccountDataResponse struct {
	Message string `json:"message"`
}

type ResetAccountDataUseCase struct {
	challengeRepo accountResetChallengeRepository
	wiper         userDataWiper
	now           func() time.Time
	generateCode  func() (string, error)
}

func NewResetAccountDataUseCase(
	challengeRepo accountResetChallengeRepository,
	wiper userDataWiper,
) *ResetAccountDataUseCase {
	return &ResetAccountDataUseCase{
		challengeRepo: challengeRepo,
		wiper:         wiper,
		now:           time.Now,
		generateCode:  generateAccountResetConfirmationText,
	}
}

func HashAccountResetConfirmation(confirmationText string) string {
	sum := sha256.Sum256([]byte(confirmationText))
	return hex.EncodeToString(sum[:])
}

func (uc *ResetAccountDataUseCase) CreateChallenge(ctx context.Context, userID int) (*CreateAccountResetChallengeResponse, error) {
	if err := uc.challengeRepo.DeleteByUserID(ctx, uint(userID)); err != nil {
		return nil, err
	}

	confirmationText, err := uc.generateCode()
	if err != nil {
		return nil, err
	}

	expiresAt := uc.now().UTC().Add(accountResetChallengeTTL)
	challenge := &AccountResetChallenge{
		ID:           uuid.New(),
		UserID:       uint(userID),
		CodeHash:     HashAccountResetConfirmation(confirmationText),
		AttemptCount: 0,
		ExpiresAt:    expiresAt,
		CreatedAt:    uc.now().UTC(),
	}

	if err := uc.challengeRepo.Save(ctx, challenge); err != nil {
		return nil, err
	}

	return &CreateAccountResetChallengeResponse{
		ConfirmationText: confirmationText,
		ExpiresAt:        expiresAt,
	}, nil
}

func (uc *ResetAccountDataUseCase) Execute(ctx context.Context, req ResetAccountDataRequest) (*ResetAccountDataResponse, error) {
	challenge, err := uc.challengeRepo.FindLatestByUserID(ctx, uint(req.UserID))
	if err != nil {
		return nil, err
	}
	if challenge == nil {
		return nil, ErrAccountResetChallengeNotFound
	}

	now := uc.now().UTC()
	if !challenge.ExpiresAt.After(now) {
		_ = uc.challengeRepo.DeleteByID(ctx, challenge.ID)
		return nil, ErrAccountResetChallengeExpired
	}

	if challenge.AttemptCount >= accountResetMaxFailedAttempts {
		_ = uc.challengeRepo.DeleteByID(ctx, challenge.ID)
		return nil, ErrAccountResetTooManyAttempts
	}

	expectedHash := HashAccountResetConfirmation(req.ConfirmationText)
	if expectedHash != challenge.CodeHash {
		_ = uc.challengeRepo.IncrementAttempts(ctx, challenge.ID)
		if challenge.AttemptCount+1 >= accountResetMaxFailedAttempts {
			_ = uc.challengeRepo.DeleteByID(ctx, challenge.ID)
			return nil, ErrAccountResetTooManyAttempts
		}
		return nil, ErrAccountResetConfirmationMismatch
	}

	if err := uc.wiper.WipeUserData(ctx, uint(req.UserID)); err != nil {
		return nil, err
	}

	return &ResetAccountDataResponse{
		Message: "Account data reset successfully",
	}, nil
}

func generateAccountResetConfirmationText() (string, error) {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"
	result := make([]byte, accountResetConfirmationLength)
	max := big.NewInt(int64(len(alphabet)))
	for i := 0; i < accountResetConfirmationLength; i++ {
		n, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		result[i] = alphabet[n.Int64()]
	}
	return string(result), nil
}
