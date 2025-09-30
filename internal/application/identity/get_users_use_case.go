package identity

import (
	"context"
	"panda-pocket/internal/domain/identity"
)

// GetUsersUseCase handles getting all users
type GetUsersUseCase struct {
	userService *identity.UserService
}

// NewGetUsersUseCase creates a new get users use case
func NewGetUsersUseCase(userService *identity.UserService) *GetUsersUseCase {
	return &GetUsersUseCase{
		userService: userService,
	}
}

// GetUsersResponse represents the response for getting all users
type GetUsersResponse struct {
	Users []UserResponse `json:"users"`
}

// UserResponse represents a user in the response
type UserResponse struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
}

// Execute executes the get users use case
func (uc *GetUsersUseCase) Execute(ctx context.Context) (*GetUsersResponse, error) {
	users, err := uc.userService.GetAllUsers(ctx)
	if err != nil {
		return nil, err
	}

	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = UserResponse{
			ID:        user.ID().Value(),
			Email:     user.Email().Value(),
			CreatedAt: user.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	return &GetUsersResponse{
		Users: userResponses,
	}, nil
}
