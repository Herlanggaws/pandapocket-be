package database

import (
	"context"
	"panda-pocket/internal/domain/identity"
	"time"

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

func (r *GormUserRepository) activeScope(db *gorm.DB) *gorm.DB {
	return db.Where("deleted_at IS NULL")
}

func toDomainUser(userModel User) (*identity.User, error) {
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
	if userModel.DeletedAt != nil {
		user.MarkDeleted(*userModel.DeletedAt)
	}
	return user, nil
}

// Save saves a user to the database
func (r *GormUserRepository) Save(ctx context.Context, user *identity.User) error {
	userModel := &User{
		Email:        user.Email().Value(),
		PasswordHash: user.PasswordHash().Value(),
		Role:         user.Role().Value(),
		DeletedAt:    user.DeletedAt(),
	}

	if user.ID().Value() != 0 {
		userModel.ID = uint(user.ID().Value())
	}

	if err := r.db.WithContext(ctx).Save(userModel).Error; err != nil {
		return err
	}

	user.AssignID(identity.NewUserID(int(userModel.ID)))
	return nil
}

// Update updates a user in the database
func (r *GormUserRepository) Update(ctx context.Context, user *identity.User) error {
	return r.Save(ctx, user)
}

// FindByID finds an active user by ID
func (r *GormUserRepository) FindByID(ctx context.Context, id identity.UserID) (*identity.User, error) {
	var userModel User

	err := r.activeScope(r.db.WithContext(ctx)).First(&userModel, id.Value()).Error
	if err != nil {
		return nil, err
	}

	return toDomainUser(userModel)
}

// FindByEmail finds an active user by email
func (r *GormUserRepository) FindByEmail(ctx context.Context, email identity.Email) (*identity.User, error) {
	var userModel User

	err := r.activeScope(r.db.WithContext(ctx)).Where("email = ?", email.Value()).First(&userModel).Error
	if err != nil {
		return nil, err
	}

	return toDomainUser(userModel)
}

// Delete deletes a user by ID
func (r *GormUserRepository) Delete(ctx context.Context, id identity.UserID) error {
	return r.db.WithContext(ctx).Delete(&User{}, id.Value()).Error
}

// FindAll finds all active users
func (r *GormUserRepository) FindAll(ctx context.Context) ([]*identity.User, error) {
	var userModels []User

	err := r.activeScope(r.db.WithContext(ctx)).Find(&userModels).Error
	if err != nil {
		return nil, err
	}

	users := make([]*identity.User, len(userModels))
	for i, userModel := range userModels {
		user, err := toDomainUser(userModel)
		if err != nil {
			return nil, err
		}
		users[i] = user
	}

	return users, nil
}

// ExistsByEmail checks if an active user exists with the given email
func (r *GormUserRepository) ExistsByEmail(ctx context.Context, email identity.Email) (bool, error) {
	var count int64
	err := r.activeScope(r.db.WithContext(ctx).Model(&User{})).Where("email = ?", email.Value()).Count(&count).Error
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsActive reports whether the user exists and is not soft-deleted.
func (r *GormUserRepository) ExistsActive(ctx context.Context, id identity.UserID) (bool, error) {
	var count int64
	err := r.activeScope(r.db.WithContext(ctx).Model(&User{})).Where("id = ?", id.Value()).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// SoftDelete marks the user deleted and frees the email unique index.
func (r *GormUserRepository) SoftDelete(ctx context.Context, id identity.UserID, mangledEmail string, deletedAt time.Time) error {
	result := r.db.WithContext(ctx).Model(&User{}).
		Where("id = ? AND deleted_at IS NULL", id.Value()).
		Updates(map[string]interface{}{
			"email":      mangledEmail,
			"deleted_at": deletedAt,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// ListDueForPurge returns soft-deleted users whose deleted_at is on or before the cutoff.
func (r *GormUserRepository) ListDueForPurge(ctx context.Context, before time.Time) ([]identity.UserID, error) {
	var userModels []User
	err := r.db.WithContext(ctx).
		Select("id").
		Where("deleted_at IS NOT NULL AND deleted_at <= ?", before).
		Find(&userModels).Error
	if err != nil {
		return nil, err
	}

	ids := make([]identity.UserID, len(userModels))
	for i, userModel := range userModels {
		ids[i] = identity.NewUserID(int(userModel.ID))
	}
	return ids, nil
}
