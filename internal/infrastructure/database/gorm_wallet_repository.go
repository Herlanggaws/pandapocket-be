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

func NewGormWalletRepository(db *gorm.DB) *GormWalletRepository {
	return &GormWalletRepository{db: db}
}

func (r *GormWalletRepository) toDomain(model Wallet) *finance.Wallet {
	return finance.ReconstituteWallet(
		finance.NewWalletID(int(model.ID)),
		finance.NewUserID(int(model.UserID)),
		model.Name,
		finance.WalletType(model.Type),
		finance.NewCurrencyID(int(model.CurrencyID)),
		model.OpeningBalance,
		model.IsDefault,
		model.IsArchived,
		model.CreatedAt,
	)
}

func (r *GormWalletRepository) Save(ctx context.Context, wallet *finance.Wallet) error {
	model := &Wallet{
		UserID:         uint(wallet.UserID().Value()),
		Name:           wallet.Name(),
		Type:           string(wallet.Type()),
		CurrencyID:     uint(wallet.CurrencyID().Value()),
		OpeningBalance: wallet.OpeningBalance(),
		IsDefault:      wallet.IsDefault(),
		IsArchived:     wallet.IsArchived(),
	}

	if wallet.ID().Value() != 0 {
		model.ID = uint(wallet.ID().Value())
		return r.db.WithContext(ctx).Model(&Wallet{}).Where("id = ?", model.ID).Updates(map[string]interface{}{
			"name":            model.Name,
			"type":            model.Type,
			"currency_id":     model.CurrencyID,
			"opening_balance": model.OpeningBalance,
			"is_default":      model.IsDefault,
			"is_archived":     model.IsArchived,
		}).Error
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	wallet.AssignID(finance.NewWalletID(int(model.ID)))
	return nil
}

func (r *GormWalletRepository) FindByID(ctx context.Context, id finance.WalletID) (*finance.Wallet, error) {
	var model Wallet
	err := r.db.WithContext(ctx).First(&model, id.Value()).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("wallet not found")
		}
		return nil, err
	}
	return r.toDomain(model), nil
}

func (r *GormWalletRepository) FindByUserID(ctx context.Context, userID finance.UserID, includeArchived bool) ([]*finance.Wallet, error) {
	query := r.db.WithContext(ctx).Where("user_id = ?", userID.Value())
	if !includeArchived {
		query = query.Where("is_archived = ?", false)
	}
	var models []Wallet
	if err := query.Order("is_default DESC, name ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	result := make([]*finance.Wallet, 0, len(models))
	for _, m := range models {
		result = append(result, r.toDomain(m))
	}
	return result, nil
}

func (r *GormWalletRepository) FindDefaultByUserID(ctx context.Context, userID finance.UserID) (*finance.Wallet, error) {
	var model Wallet
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_default = ? AND is_archived = ?", userID.Value(), true, false).
		First(&model).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = r.db.WithContext(ctx).
				Where("user_id = ? AND is_archived = ?", userID.Value(), false).
				Order("id ASC").
				First(&model).Error
			if err != nil {
				return nil, errors.New("default wallet not found")
			}
		} else {
			return nil, err
		}
	}
	return r.toDomain(model), nil
}

func (r *GormWalletRepository) CountActiveByUserID(ctx context.Context, userID finance.UserID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Wallet{}).
		Where("user_id = ? AND is_archived = ?", userID.Value(), false).
		Count(&count).Error
	return count, err
}

func (r *GormWalletRepository) ClearDefaultForUser(ctx context.Context, userID finance.UserID) error {
	return r.db.WithContext(ctx).Model(&Wallet{}).
		Where("user_id = ? AND is_default = ?", userID.Value(), true).
		Update("is_default", false).Error
}

func (r *GormWalletRepository) HasTransactions(ctx context.Context, id finance.WalletID) (bool, error) {
	var expenseCount, incomeCount, transferCount int64
	if err := r.db.WithContext(ctx).Model(&Expense{}).Where("wallet_id = ?", id.Value()).Count(&expenseCount).Error; err != nil {
		return false, err
	}
	if expenseCount > 0 {
		return true, nil
	}
	if err := r.db.WithContext(ctx).Model(&Income{}).Where("wallet_id = ?", id.Value()).Count(&incomeCount).Error; err != nil {
		return false, err
	}
	if incomeCount > 0 {
		return true, nil
	}
	if err := r.db.WithContext(ctx).Model(&Transfer{}).
		Where("from_wallet_id = ? OR to_wallet_id = ?", id.Value(), id.Value()).
		Count(&transferCount).Error; err != nil {
		return false, err
	}
	return transferCount > 0, nil
}

func (r *GormWalletRepository) GetBalanceBreakdown(ctx context.Context, id finance.WalletID) (finance.WalletBalanceBreakdown, error) {
	wallet, err := r.FindByID(ctx, id)
	if err != nil {
		return finance.WalletBalanceBreakdown{}, err
	}

	var incomeTotal, expenseTotal, transfersIn, transfersOut float64
	if err := r.db.WithContext(ctx).Model(&Income{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("wallet_id = ?", id.Value()).
		Scan(&incomeTotal).Error; err != nil {
		return finance.WalletBalanceBreakdown{}, err
	}
	if err := r.db.WithContext(ctx).Model(&Expense{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("wallet_id = ?", id.Value()).
		Scan(&expenseTotal).Error; err != nil {
		return finance.WalletBalanceBreakdown{}, err
	}
	if err := r.db.WithContext(ctx).Model(&Transfer{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("to_wallet_id = ?", id.Value()).
		Scan(&transfersIn).Error; err != nil {
		return finance.WalletBalanceBreakdown{}, err
	}
	if err := r.db.WithContext(ctx).Model(&Transfer{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("from_wallet_id = ?", id.Value()).
		Scan(&transfersOut).Error; err != nil {
		return finance.WalletBalanceBreakdown{}, err
	}

	return finance.ComputeWalletBalance(
		wallet.OpeningBalance(),
		incomeTotal,
		expenseTotal,
		transfersIn,
		transfersOut,
	), nil
}

func (r *GormWalletRepository) AlignPendingCurrency(ctx context.Context, id finance.WalletID, currencyID finance.CurrencyID) error {
	return r.db.WithContext(ctx).Model(&PendingTransaction{}).
		Where("wallet_id = ? AND status = ?", id.Value(), "pending").
		Update("currency_id", currencyID.Value()).Error
}

func (r *GormWalletRepository) AlignRecurringCurrency(ctx context.Context, id finance.WalletID, currencyID finance.CurrencyID) error {
	return r.db.WithContext(ctx).Model(&RecurringTransaction{}).
		Where("wallet_id = ?", id.Value()).
		Update("currency_id", currencyID.Value()).Error
}
