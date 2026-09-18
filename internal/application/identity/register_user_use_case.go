package identity

import (
	"context"
	"errors"
	"panda-pocket/internal/domain/billing"
	"panda-pocket/internal/domain/identity"

	"golang.org/x/crypto/bcrypt"
)

// RegisterUserRequest represents the request to register a new user
type RegisterUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterUserResponse represents the response after registering a user
type RegisterUserResponse struct {
	UserID       int    `json:"user_id"`
	Email        string `json:"email"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// RegisterUserUseCase handles user registration
type RegisterUserUseCase struct {
	userService      *identity.UserService
	tokenService     TokenService
	subscriptionRepo billing.SubscriptionRepository
}

// NewRegisterUserUseCase creates a new register user use case
func NewRegisterUserUseCase(
	userService *identity.UserService,
	tokenService TokenService,
	subscriptionRepo billing.SubscriptionRepository,
) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		userService:      userService,
		tokenService:     tokenService,
		subscriptionRepo: subscriptionRepo,
	}
}

// Execute executes the register user use case
func (uc *RegisterUserUseCase) Execute(ctx context.Context, req RegisterUserRequest) (*RegisterUserResponse, error) {
	email, err := identity.NewEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email format")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	passwordHash := identity.NewPasswordHash(string(hashedPassword))

	user, err := uc.userService.RegisterUser(ctx, email, passwordHash)
	if err != nil {
		return nil, err
	}

	sub, err := billing.NewTrialSubscription(user.ID().Value())
	if err != nil {
		return nil, err
	}
	if err := uc.subscriptionRepo.Save(ctx, sub); err != nil {
		return nil, errors.New("failed to create subscription")
	}

	token, refreshToken, err := uc.tokenService.GenerateToken(ctx, user.ID().Value(), user.Email().Value(), user.Role().Value())
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	return &RegisterUserResponse{
		UserID:       user.ID().Value(),
		Email:        user.Email().Value(),
		Token:        token,
		RefreshToken: refreshToken,
	}, nil
}
