package identity

import (
	"context"
	"time"

	domainIdentity "panda-pocket/internal/domain/identity"
)

type CleanupExpiredTokensResult struct {
	SessionTokensDeleted       int64
	PasswordResetTokensDeleted int64
}

type CleanupExpiredTokensUseCase struct {
	authTokens     domainIdentity.TokenRepository
	passwordResets domainIdentity.PasswordResetTokenRepository
	now            func() time.Time
}

func NewCleanupExpiredTokensUseCase(
	authTokens domainIdentity.TokenRepository,
	passwordResets domainIdentity.PasswordResetTokenRepository,
) *CleanupExpiredTokensUseCase {
	return &CleanupExpiredTokensUseCase{
		authTokens:     authTokens,
		passwordResets: passwordResets,
		now:            time.Now,
	}
}

func (uc *CleanupExpiredTokensUseCase) Execute(ctx context.Context) (CleanupExpiredTokensResult, error) {
	before := uc.now().UTC()
	sessionDeleted, err := uc.authTokens.DeleteExpired(ctx, before)
	if err != nil {
		return CleanupExpiredTokensResult{}, err
	}
	resetDeleted, err := uc.passwordResets.DeleteExpired(ctx, before)
	if err != nil {
		return CleanupExpiredTokensResult{}, err
	}
	return CleanupExpiredTokensResult{
		SessionTokensDeleted:       sessionDeleted,
		PasswordResetTokensDeleted: resetDeleted,
	}, nil
}
