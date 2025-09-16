package identity

import (
	"context"
	"errors"
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
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Token  string `json:"token"`
}

// RegisterUserUseCase handles user registration
type RegisterUserUseCase struct {
	userService *identity.UserService
	tokenService TokenService
}

// NewRegisterUserUseCase creates a new register user use case
func NewRegisterUserUseCase(userService *identity.UserService, tokenService TokenService) *RegisterUserUseCase {
	return &RegisterUserUseCase{
		userService:  userService,
		tokenService: tokenService,
	}
}

// Execute executes the register user use case
func (uc *RegisterUserUseCase) Execute(ctx context.Context, req RegisterUserRequest) (*RegisterUserResponse, error) {
	// Create email value object
	email, err := identity.NewEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email format")
	}
	
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}
	
	passwordHash := identity.NewPasswordHash(string(hashedPassword))
	
	// Register user
	user, err := uc.userService.RegisterUser(ctx, email, passwordHash)
	if err != nil {
		return nil, err
	}
	
	// Generate token
	token, err := uc.tokenService.GenerateToken(user.ID().Value(), user.Email().Value())
	if err != nil {
		return nil, errors.New("failed to generate token")
	}
	
	return &RegisterUserResponse{
		UserID: user.ID().Value(),
		Email:  user.Email().Value(),
		Token:  token,
	}, nil
}
