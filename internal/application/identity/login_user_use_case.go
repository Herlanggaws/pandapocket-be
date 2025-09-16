package identity

import (
	"context"
	"errors"
	"panda-pocket/internal/domain/identity"
	"golang.org/x/crypto/bcrypt"
)

// LoginUserRequest represents the request to login a user
type LoginUserRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// LoginUserResponse represents the response after logging in a user
type LoginUserResponse struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Token  string `json:"token"`
}

// LoginUserUseCase handles user login
type LoginUserUseCase struct {
	userService  *identity.UserService
	tokenService TokenService
}

// NewLoginUserUseCase creates a new login user use case
func NewLoginUserUseCase(userService *identity.UserService, tokenService TokenService) *LoginUserUseCase {
	return &LoginUserUseCase{
		userService:  userService,
		tokenService: tokenService,
	}
}

// Execute executes the login user use case
func (uc *LoginUserUseCase) Execute(ctx context.Context, req LoginUserRequest) (*LoginUserResponse, error) {
	// Create email value object
	email, err := identity.NewEmail(req.Email)
	if err != nil {
		return nil, errors.New("invalid email format")
	}
	
	// Get user by email
	user, err := uc.userService.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	
	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash().Value()), []byte(req.Password))
	if err != nil {
		return nil, errors.New("invalid credentials")
	}
	
	// Generate token
	token, err := uc.tokenService.GenerateToken(user.ID().Value(), user.Email().Value())
	if err != nil {
		return nil, errors.New("failed to generate token")
	}
	
	return &LoginUserResponse{
		UserID: user.ID().Value(),
		Email:  user.Email().Value(),
		Token:  token,
	}, nil
}
