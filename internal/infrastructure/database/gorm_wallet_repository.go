package database

import (
	"context"
	"errors"
	"panda-pocket/internal/domain/finance"

	"gorm.io/gorm"
)

type GormWalletRepository struct {
	db *gorm.DB
}

func NewGormWalletRepository(db *gorm.DB) finance.WalletRepository {
	return &GormWalletRepository{db: db}
}

func (r *GormWalletRepository) Save(ctx context.Context, wallet *finance.Wallet) error {
	db := dbFromContext(ctx, r.db)
	dbWallet := toDBWallet(wallet)
	if err := db.WithContext(ctx).Create(dbWallet).Error; err != nil {
		return err
	}
	wallet.SetID(finance.NewWalletID(int(dbWallet.ID)))
	return nil
}

func (r *GormWalletRepository) FindByID(ctx context.Context, id finance.WalletID) (*finance.Wallet, error) {
	db := dbFromContext(ctx, r.db)
	var dbWallet Wallet
	if err := db.WithContext(ctx).First(&dbWallet, id.Value()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("wallet not found")
		}
		return nil, err
	}
	return toDomainWallet(&dbWallet)
}

func (r *GormWalletRepository) FindByUserID(ctx context.Context, userID finance.UserID) ([]*finance.Wallet, error) {
	db := dbFromContext(ctx, r.db)
	var dbWallets []Wallet
	if err := db.WithContext(ctx).Where("user_id = ?", userID.Value()).Find(&dbWallets).Error; err != nil {
		return nil, err
	}

	var wallets []*finance.Wallet
	for _, dbWallet := range dbWallets {
		wallet, err := toDomainWallet(&dbWallet)
		if err != nil {
			return nil, err
		}
		wallets = append(wallets, wallet)
	}
	return wallets, nil
}

func (r *GormWalletRepository) FindByUserIDWithFilters(ctx context.Context, userID finance.UserID, search string, limit, offset int) ([]*finance.Wallet, int64, error) {
	db := dbFromContext(ctx, r.db)
	var dbWallets []Wallet
	var totalCount int64

	query := db.WithContext(ctx).Model(&Wallet{}).Where("user_id = ?", userID.Value()).Order("created_at ASC")

	if search != "" {
		query = query.Where("name LIKE ?", "%"+search+"%")
	}

	if err := query.Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset(offset).Find(&dbWallets).Error; err != nil {
		return nil, 0, err
	}

	var wallets []*finance.Wallet
	for _, dbWallet := range dbWallets {
		wallet, err := toDomainWallet(&dbWallet)
		if err != nil {
			return nil, 0, err
		}
		wallets = append(wallets, wallet)
	}

	return wallets, totalCount, nil
}

func (r *GormWalletRepository) UnsetPrimaryByUserIDExcept(ctx context.Context, userID finance.UserID, walletID finance.WalletID) error {
	db := dbFromContext(ctx, r.db)
	return db.WithContext(ctx).Model(&Wallet{}).
		Where("user_id = ? AND id <> ?", userID.Value(), walletID.Value()).
		Update("is_primary", false).Error
}

func (r *GormWalletRepository) Update(ctx context.Context, wallet *finance.Wallet) error {
	db := dbFromContext(ctx, r.db)
	dbWallet := toDBWallet(wallet)
	// We use Updates to only update non-zero fields, or Save to update everything
	if err := db.WithContext(ctx).Model(&Wallet{}).Where("id = ?", dbWallet.ID).Updates(map[string]interface{}{
		"name":       dbWallet.Name,
		"amount":     dbWallet.Amount,
		"is_primary": dbWallet.IsPrimary,
		"updated_at": dbWallet.UpdatedAt,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (r *GormWalletRepository) Delete(ctx context.Context, id finance.WalletID) error {
	db := dbFromContext(ctx, r.db)
	result := db.WithContext(ctx).Delete(&Wallet{}, id.Value())
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("wallet not found")
	}
	return nil
}

// Convert domain Wallet to DB Wallet
func toDBWallet(w *finance.Wallet) *Wallet {
	return &Wallet{
		ID:        uint(w.ID().Value()),
		UserID:    uint(w.UserID().Value()),
		Name:      w.Name(),
		Amount:    w.Amount(),
		IsPrimary: w.IsPrimary(),
		CreatedAt: w.CreatedAt(),
		UpdatedAt: w.UpdatedAt(),
	}
}

// Convert DB Wallet to domain Wallet
func toDomainWallet(w *Wallet) (*finance.Wallet, error) {
	// Reconstruct the domain wallet
	wallet, err := finance.ReconstructWallet(
		finance.NewWalletID(int(w.ID)),
		finance.NewUserID(int(w.UserID)),
		w.Name,
		w.Amount,
		w.IsPrimary,
		w.CreatedAt,
		w.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return wallet, nil
}
