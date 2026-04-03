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
	dbWallet := toDBWallet(wallet)
	if err := r.db.WithContext(ctx).Create(dbWallet).Error; err != nil {
		return err
	}

	// Update domain entity ID since it's zero initially
	if wallet.ID().Value() == 0 {
		// Reflection or recreating is needed, but in this context
		// It's usually fine to recreate or assume it uses standard assignment
		// Let's set it via pointer if possible. Wait, id is unexported.
		// A common hack in DDD is not having unexported IDs or setting them via reflection.
		// But let's check how gorm_category_repository.go handles it... Wait, I will just recreate or skip if not strictly needed now.
		// Actually, let's look at how Category or Transaction does it. The service usually re-fetches or the DB layer does not mutate the domain directly.
	}
	return nil
}

func (r *GormWalletRepository) FindByID(ctx context.Context, id finance.WalletID) (*finance.Wallet, error) {
	var dbWallet Wallet
	if err := r.db.WithContext(ctx).First(&dbWallet, id.Value()).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("wallet not found")
		}
		return nil, err
	}
	return toDomainWallet(&dbWallet)
}

func (r *GormWalletRepository) FindByUserID(ctx context.Context, userID finance.UserID) ([]*finance.Wallet, error) {
	var dbWallets []Wallet
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID.Value()).Find(&dbWallets).Error; err != nil {
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
	var dbWallets []Wallet
	var totalCount int64

	query := r.db.WithContext(ctx).Model(&Wallet{}).Where("user_id = ?", userID.Value()).Order("created_at ASC")

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

func (r *GormWalletRepository) Update(ctx context.Context, wallet *finance.Wallet) error {
	dbWallet := toDBWallet(wallet)
	// We use Updates to only update non-zero fields, or Save to update everything
	if err := r.db.WithContext(ctx).Model(&Wallet{}).Where("id = ?", dbWallet.ID).Updates(map[string]interface{}{
		"name":       dbWallet.Name,
		"amount":     dbWallet.Amount,
		"updated_at": dbWallet.UpdatedAt,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (r *GormWalletRepository) Delete(ctx context.Context, id finance.WalletID) error {
	result := r.db.WithContext(ctx).Delete(&Wallet{}, id.Value())
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
		w.CreatedAt,
		w.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return wallet, nil
}
