package identity

import (
	"context"
	"errors"
	"panda-pocket/internal/domain/identity"
)

// RefreshTokenRequest represents the request to refresh an access token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse represents the response after refreshing a token
type RefreshTokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshTokenUseCase handles token refresh
type RefreshTokenUseCase struct {
	userService  *identity.UserService
	tokenService TokenService
}

// NewRefreshTokenUseCase creates a new refresh token use case
func NewRefreshTokenUseCase(userService *identity.UserService, tokenService TokenService) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{
		userService:  userService,
		tokenService: tokenService,
	}
}

// Execute executes the refresh token use case
func (uc *RefreshTokenUseCase) Execute(ctx context.Context, req RefreshTokenRequest) (*RefreshTokenResponse, error) {
	// Validate refresh token
	claims, err := uc.tokenService.ValidateRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Get user to verify existence and get current role/email
	userID := identity.NewUserID(claims.UserID)
	user, err := uc.userService.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// Generate new token pair
	accessToken, refreshToken, err := uc.tokenService.GenerateToken(ctx, user.ID().Value(), user.Email().Value(), user.Role().Value())
	if err != nil {
		return nil, errors.New("failed to generate tokens")
	}

	return &RefreshTokenResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
	}, nil
}
