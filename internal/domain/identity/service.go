package identity

import (
	"context"
	"errors"
)

// UserService handles user-related domain operations
type UserService struct {
	userRepo UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// RegisterUser registers a new user
func (s *UserService) RegisterUser(ctx context.Context, email Email, password PasswordHash) (*User, error) {
	// Check if user already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, errors.New("user already exists")
	}

	// Create new user
	user := NewUser(UserID{}, email, password)

	// Save user
	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// AuthenticateUser authenticates a user with email and password
func (s *UserService) AuthenticateUser(ctx context.Context, email Email, password PasswordHash) (*User, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// In a real implementation, you would verify the password hash here
	// For now, we'll assume the password is already hashed and matches
	if user.PasswordHash() != password {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(ctx context.Context, id UserID) (*User, error) {
	return s.userRepo.FindByID(ctx, id)
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(ctx context.Context, email Email) (*User, error) {
	return s.userRepo.FindByEmail(ctx, email)
}

// GetAllUsers retrieves all users
func (s *UserService) GetAllUsers(ctx context.Context) ([]*User, error) {
	return s.userRepo.FindAll(ctx)
}

// UpdateUserEmail updates a user's email
func (s *UserService) UpdateUserEmail(ctx context.Context, id UserID, newEmail Email) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Check if new email already exists
	exists, err := s.userRepo.ExistsByEmail(ctx, newEmail)
	if err != nil {
		return err
	}

	if exists {
		return errors.New("email already exists")
	}

	user.ChangeEmail(newEmail)
	return s.userRepo.Save(ctx, user)
}
