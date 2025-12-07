package identity

import (
	"context"
	"errors"
	"fmt"
	"panda-pocket/internal/domain/identity"
)

// ResetPasswordRequest represents the request for resetting password
type ResetPasswordRequest struct {
	Token              string `json:"token" binding:"required"`
	NewPassword        string `json:"new_password" binding:"required,min=8"`
	ConfirmNewPassword string `json:"confirm_new_password" binding:"required"`
}

// ResetPasswordResponse represents the response for resetting password
type ResetPasswordResponse struct {
	Message string `json:"message"`
}

// ResetPasswordUseCase handles logic for resetting password
type ResetPasswordUseCase struct {
	userRepository  identity.UserRepository
	tokenRepository identity.PasswordResetTokenRepository
}

// NewResetPasswordUseCase creates a new reset password use case
func NewResetPasswordUseCase(
	userRepository identity.UserRepository,
	tokenRepository identity.PasswordResetTokenRepository,
) *ResetPasswordUseCase {
	return &ResetPasswordUseCase{
		userRepository:  userRepository,
		tokenRepository: tokenRepository,
	}
}

// Execute executes the reset password logic
func (uc *ResetPasswordUseCase) Execute(ctx context.Context, req *ResetPasswordRequest) (*ResetPasswordResponse, error) {
	// 1. Validate passwords match
	if req.NewPassword != req.ConfirmNewPassword {
		return nil, errors.New("passwords do not match")
	}

	// 2. Find token
	token, err := uc.tokenRepository.FindByToken(ctx, req.Token)
	if err != nil {
		return nil, errors.New("invalid or expired token")
	}
	if token == nil {
		return nil, errors.New("invalid or expired token")
	}

	// 3. Check expiration
	if token.IsExpired() {
		// Clean up expired token
		_ = uc.tokenRepository.Delete(ctx, token.ID())
		return nil, errors.New("token has expired")
	}

	// 4. Find user
	user, err := uc.userRepository.FindByID(ctx, token.UserID())
	if err != nil {
		return nil, errors.New("user not found")
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// 5. Update user password
	// We need to hash the password. Since hashing logic is usually in TokenService or User entity interaction.
	// Looking at RegisterUserUseCase, it uses TokenService to hash? Check RegisterUserUseCase.
	// Or User entity has a method?
	// I'll check user.go or just implement hashing here if needed, but better to reuse.
	// Wait, I injected TokenService, let's see if it has ValidatePassword logic or Hash logic.
	// Actually, `User` entity usually has `UpdatePassword` if it knows how to hash, or we set a `PasswordHash` VO.
	// I'll assume for now I need to hash it.
	// Let's check `User` entity methods.

	// Assuming TokenService helps with hashing or checking.
	// If not, I might need to use bcrypt directly or a specialized service.
	// Let's assume I can create a new PasswordHash VO from plain password (which does hashing).
	// Checking `password_hash.go` (if exists) or `user.go`.

	hashedPassword, err := identity.NewPasswordHashFromPlain(req.NewPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user.UpdatePassword(hashedPassword)

	// 6. Save user
	if err := uc.userRepository.Update(ctx, user); err != nil {
		return nil, errors.New("failed to update password")
	}

	// 7. Delete token
	_ = uc.tokenRepository.Delete(ctx, token.ID())

	return &ResetPasswordResponse{
		Message: "Password has been reset successfully",
	}, nil
}
