package identity

import (
	"context"
	"log"
	"time"

	domainIdentity "panda-pocket/internal/domain/identity"
)

type accountHardDeleter interface {
	HardDeleteAccount(ctx context.Context, userID uint) error
}

type PurgeDeletedAccountsUseCase struct {
	userRepo domainIdentity.UserRepository
	deleter  accountHardDeleter
	now      func() time.Time
}

func NewPurgeDeletedAccountsUseCase(
	userRepo domainIdentity.UserRepository,
	deleter accountHardDeleter,
) *PurgeDeletedAccountsUseCase {
	return &PurgeDeletedAccountsUseCase{
		userRepo: userRepo,
		deleter:  deleter,
		now:      time.Now,
	}
}

func (uc *PurgeDeletedAccountsUseCase) Execute(ctx context.Context) (int, error) {
	cutoff := uc.now().UTC().Add(-AccountDeletionRetention)
	ids, err := uc.userRepo.ListDueForPurge(ctx, cutoff)
	if err != nil {
		return 0, err
	}

	purged := 0
	for _, id := range ids {
		if err := uc.deleter.HardDeleteAccount(ctx, uint(id.Value())); err != nil {
			log.Printf("failed to purge deleted account %d: %v", id.Value(), err)
			continue
		}
		purged++
	}
	return purged, nil
}
