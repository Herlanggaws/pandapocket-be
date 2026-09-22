package database

import (
	"context"
	"panda-pocket/internal/domain/finance"

	"gorm.io/gorm"
)

type GormTransferRepository struct {
	db *gorm.DB
}

func NewGormTransferRepository(db *gorm.DB) *GormTransferRepository {
	return &GormTransferRepository{db: db}
}

func (r *GormTransferRepository) toDomain(model Transfer) *finance.Transfer {
	return finance.ReconstituteTransfer(
		finance.NewTransferID(int(model.ID)),
		finance.NewUserID(int(model.UserID)),
		finance.NewWalletID(int(model.FromWalletID)),
		finance.NewWalletID(int(model.ToWalletID)),
		model.Amount,
		model.Description,
		model.Date,
		model.CreatedAt,
	)
}

func (r *GormTransferRepository) Save(ctx context.Context, transfer *finance.Transfer) error {
	model := &Transfer{
		UserID:       uint(transfer.UserID().Value()),
		FromWalletID: uint(transfer.FromWalletID().Value()),
		ToWalletID:   uint(transfer.ToWalletID().Value()),
		Amount:       transfer.Amount(),
		Description:  transfer.Description(),
		Date:         transfer.Date(),
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	transfer.AssignID(finance.NewTransferID(int(model.ID)))
	return nil
}

func (r *GormTransferRepository) FindByID(ctx context.Context, id finance.TransferID) (*finance.Transfer, error) {
	var model Transfer
	err := r.db.WithContext(ctx).First(&model, id.Value()).Error
	if err == gorm.ErrRecordNotFound {
		return nil, gorm.ErrRecordNotFound
	}
	if err != nil {
		return nil, err
	}
	return r.toDomain(model), nil
}

func (r *GormTransferRepository) Delete(ctx context.Context, id finance.TransferID) error {
	return r.db.WithContext(ctx).Delete(&Transfer{}, id.Value()).Error
}

func (r *GormTransferRepository) FindByUserID(
	ctx context.Context,
	userID finance.UserID,
	filters finance.TransferFilters,
) ([]*finance.Transfer, error) {
	query := r.db.WithContext(ctx).Where("user_id = ?", userID.Value())

	if filters.WalletID != nil {
		query = query.Where("from_wallet_id = ? OR to_wallet_id = ?", filters.WalletID.Value(), filters.WalletID.Value())
	}
	if filters.StartDate != nil {
		query = query.Where("date >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("date <= ?", *filters.EndDate)
	}

	var models []Transfer
	if err := query.Order("date DESC, id DESC").Find(&models).Error; err != nil {
		return nil, err
	}

	result := make([]*finance.Transfer, 0, len(models))
	for _, m := range models {
		result = append(result, r.toDomain(m))
	}
	return result, nil
}
