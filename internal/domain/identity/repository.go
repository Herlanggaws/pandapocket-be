package identity

import (
	"context"
	"time"
)

// UserRepository defines the contract for user persistence
type UserRepository interface {
	Save(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id UserID) (*User, error)
	FindByEmail(ctx context.Context, email Email) (*User, error)
	FindAll(ctx context.Context) ([]*User, error)
	Delete(ctx context.Context, id UserID) error
	ExistsByEmail(ctx context.Context, email Email) (bool, error)
	ExistsActive(ctx context.Context, id UserID) (bool, error)
	SoftDelete(ctx context.Context, id UserID, mangledEmail string, deletedAt time.Time) error
	ListDueForPurge(ctx context.Context, before time.Time) ([]UserID, error)
}
