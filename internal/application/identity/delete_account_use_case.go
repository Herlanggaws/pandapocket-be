package identity

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainIdentity "panda-pocket/internal/domain/identity"
)

const AccountDeletionRetention = 14 * 24 * time.Hour

type DeleteAccountRequest struct {
	UserID   int    `json:"-"`
	Password string `json:"password" binding:"required"`
}

type DeleteAccountResponse struct {
	Message           string    `json:"message"`
	ScheduledPurgeAt  time.Time `json:"scheduled_purge_at"`
}

type tokenRevoker interface {
	RevokeAllForUser(ctx context.Context, userID int) error
}

type DeleteAccountUseCase struct {
	userRepo     domainIdentity.UserRepository
	tokenRevoker tokenRevoker
	now          func() time.Time
}

func NewDeleteAccountUseCase(
	userRepo domainIdentity.UserRepository,
	tokenRevoker tokenRevoker,
) *DeleteAccountUseCase {
	return &DeleteAccountUseCase{
		userRepo:     userRepo,
		tokenRevoker: tokenRevoker,
		now:          time.Now,
	}
}

func mangledDeletedEmail(userID int, at time.Time) string {
	return fmt.Sprintf("deleted+%d+%d@deleted.local", userID, at.UnixNano())
}

func (uc *DeleteAccountUseCase) Execute(ctx context.Context, req DeleteAccountRequest) (*DeleteAccountResponse, error) {
	if req.Password == "" {
		return nil, errors.New("password is required")
	}

	userID := domainIdentity.NewUserID(req.UserID)
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	if !user.CheckPassword(req.Password) {
		return nil, errors.New("invalid password")
	}

	deletedAt := uc.now().UTC()
	mangledEmail := mangledDeletedEmail(req.UserID, deletedAt)
	if err := uc.userRepo.SoftDelete(ctx, userID, mangledEmail, deletedAt); err != nil {
		return nil, err
	}

	if err := uc.tokenRevoker.RevokeAllForUser(ctx, req.UserID); err != nil {
		return nil, err
	}

	return &DeleteAccountResponse{
		Message:          "Account scheduled for deletion",
		ScheduledPurgeAt: deletedAt.Add(AccountDeletionRetention),
	}, nil
}
