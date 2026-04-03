package identity

import (
	"context"
	"errors"
	domainIdentity "panda-pocket/internal/domain/identity"
)

type ChangePasswordRequest struct {
	UserID             int    `json:"-"` // From token
	OldPassword        string `json:"old_password" binding:"required"`
	NewPassword        string `json:"new_password" binding:"required,min=8"`
	ConfirmNewPassword string `json:"confirm_new_password" binding:"required"`
}

type ChangePasswordUseCase struct {
	userRepo domainIdentity.UserRepository
}

func NewChangePasswordUseCase(userRepo domainIdentity.UserRepository) *ChangePasswordUseCase {
	return &ChangePasswordUseCase{
		userRepo: userRepo,
	}
}

func (uc *ChangePasswordUseCase) Execute(ctx context.Context, req ChangePasswordRequest) error {
	// 1. Validate new password confirmation
	if req.NewPassword != req.ConfirmNewPassword {
		return errors.New("new password and confirm new password do not match")
	}

	// 2. Get user by ID
	userID := domainIdentity.NewUserID(req.UserID)
	user, err := uc.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	// 3. Verify old password
	if !user.CheckPassword(req.OldPassword) {
		return errors.New("invalid old password")
	}

	// 4. Hash new password
	newPasswordHash, err := domainIdentity.NewPasswordHashFromPlain(req.NewPassword)
	if err != nil {
		return err
	}

	// 5. Update user password
	if err := user.ChangePassword(newPasswordHash); err != nil {
		return err
	}

	// 6. Save user
	return uc.userRepo.Update(ctx, user)
}
