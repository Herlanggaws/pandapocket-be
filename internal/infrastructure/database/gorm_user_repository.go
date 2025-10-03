package database

import (
	"context"
	"panda-pocket/internal/domain/identity"

	"gorm.io/gorm"
)

// GormUserRepository implements the UserRepository interface using GORM
type GormUserRepository struct {
	db *gorm.DB
}

// NewGormUserRepository creates a new GORM user repository
func NewGormUserRepository(db *gorm.DB) *GormUserRepository {
	return &GormUserRepository{db: db}
}

// Save saves a user to the database
func (r *GormUserRepository) Save(ctx context.Context, user *identity.User) error {
	// Convert domain user to GORM model
	userModel := &User{
		Email:        user.Email().Value(),
		PasswordHash: user.PasswordHash().Value(),
		Role:         user.Role().Value(),
	}

	if user.ID().Value() != 0 {
		userModel.ID = uint(user.ID().Value())
	}

	// Save using GORM
	if err := r.db.WithContext(ctx).Save(userModel).Error; err != nil {
		return err
	}

	// Note: In a real implementation, you'd want to update the domain user's ID
	// This would require modifying the domain user to be mutable or using a different approach
	// For now, we rely on the database to handle ID generation

	return nil
}

// FindByID finds a user by ID
func (r *GormUserRepository) FindByID(ctx context.Context, id identity.UserID) (*identity.User, error) {
	var userModel User

	err := r.db.WithContext(ctx).First(&userModel, id.Value()).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}

	// Convert GORM model to domain user
	emailVO, err := identity.NewEmail(userModel.Email)
	if err != nil {
		return nil, err
	}

	passwordHashVO := identity.NewPasswordHash(userModel.PasswordHash)
	roleVO, err := identity.NewRole(userModel.Role)
	if err != nil {
		return nil, err
	}
	userID := identity.NewUserID(int(userModel.ID))

	user := identity.NewUser(userID, emailVO, passwordHashVO, roleVO)

	return user, nil
}

// FindByEmail finds a user by email
func (r *GormUserRepository) FindByEmail(ctx context.Context, email identity.Email) (*identity.User, error) {
	var userModel User

	err := r.db.WithContext(ctx).Where("email = ?", email.Value()).First(&userModel).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, err
		}
		return nil, err
	}

	// Convert GORM model to domain user
	passwordHashVO := identity.NewPasswordHash(userModel.PasswordHash)
	roleVO, err := identity.NewRole(userModel.Role)
	if err != nil {
		return nil, err
	}
	userID := identity.NewUserID(int(userModel.ID))

	user := identity.NewUser(userID, email, passwordHashVO, roleVO)

	return user, nil
}

// Delete deletes a user by ID
func (r *GormUserRepository) Delete(ctx context.Context, id identity.UserID) error {
	return r.db.WithContext(ctx).Delete(&User{}, id.Value()).Error
}

// FindAll finds all users
func (r *GormUserRepository) FindAll(ctx context.Context) ([]*identity.User, error) {
	var userModels []User

	err := r.db.WithContext(ctx).Find(&userModels).Error
	if err != nil {
		return nil, err
	}

	// Convert GORM models to domain users
	users := make([]*identity.User, len(userModels))
	for i, userModel := range userModels {
		emailVO, err := identity.NewEmail(userModel.Email)
		if err != nil {
			return nil, err
		}

		passwordHashVO := identity.NewPasswordHash(userModel.PasswordHash)
		roleVO, err := identity.NewRole(userModel.Role)
		if err != nil {
			return nil, err
		}
		userID := identity.NewUserID(int(userModel.ID))

		users[i] = identity.NewUser(userID, emailVO, passwordHashVO, roleVO)
	}

	return users, nil
}

// ExistsByEmail checks if a user exists with the given email
func (r *GormUserRepository) ExistsByEmail(ctx context.Context, email identity.Email) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&User{}).Where("email = ?", email.Value()).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
