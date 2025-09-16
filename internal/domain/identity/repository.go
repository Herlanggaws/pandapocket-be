package identity

import (
	"context"
)

// UserRepository defines the contract for user persistence
type UserRepository interface {
	Save(ctx context.Context, user *User) error
	FindByID(ctx context.Context, id UserID) (*User, error)
	FindByEmail(ctx context.Context, email Email) (*User, error)
	Delete(ctx context.Context, id UserID) error
	ExistsByEmail(ctx context.Context, email Email) (bool, error)
}
