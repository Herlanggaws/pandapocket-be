package database

import (
	"context"
	"errors"
	"panda-pocket/internal/domain/identity"

	"gorm.io/gorm"
)

type GormIdentityRepository struct {
	db *gorm.DB
}

func NewGormIdentityRepository(db *gorm.DB) *GormIdentityRepository {
	return &GormIdentityRepository{db: db}
}

func (r *GormIdentityRepository) FindByProviderSubject(ctx context.Context, provider, subject string) (*identity.ExternalIdentity, error) {
	var row UserIdentity
	err := r.db.WithContext(ctx).
		Where("provider = ? AND subject = ?", provider, subject).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &identity.ExternalIdentity{
		UserID:   int(row.UserID),
		Provider: row.Provider,
		Subject:  row.Subject,
		Email:    row.Email,
	}, nil
}

func (r *GormIdentityRepository) Save(ctx context.Context, external identity.ExternalIdentity) error {
	row := UserIdentity{
		UserID:   uint(external.UserID),
		Provider: external.Provider,
		Subject:  external.Subject,
		Email:    external.Email,
	}
	return r.db.WithContext(ctx).Create(&row).Error
}
