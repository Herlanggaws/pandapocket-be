package database

import (
	"context"
	"errors"
	"panda-pocket/internal/domain/finance"
	"time"

	"gorm.io/gorm"
)

type GormPendingTransactionRepository struct {
	db *gorm.DB
}

func NewGormPendingTransactionRepository(db *gorm.DB) *GormPendingTransactionRepository {
	return &GormPendingTransactionRepository{db: db}
}

func (r *GormPendingTransactionRepository) toDomain(model PendingTransaction) (*finance.PendingTransaction, error) {
	amount, err := finance.NewMoney(model.Amount, finance.NewCurrencyID(int(model.CurrencyID)))
	if err != nil {
		return nil, err
	}
	return finance.ReconstitutePendingTransaction(
		finance.NewPendingTransactionID(int(model.ID)),
		finance.NewUserID(int(model.UserID)),
		finance.NewRecurringTransactionID(int(model.RecurringTransactionID)),
		model.DueDate,
		amount,
		model.Description,
		finance.TransactionType(model.Type),
		finance.NewCategoryID(int(model.CategoryID)),
		finance.NewCurrencyID(int(model.CurrencyID)),
		finance.PendingTransactionStatus(model.Status),
		model.CreatedAt,
		model.ResolvedAt,
	), nil
}

func (r *GormPendingTransactionRepository) Save(ctx context.Context, pending *finance.PendingTransaction) error {
	model := &PendingTransaction{
		UserID:                 uint(pending.UserID().Value()),
		RecurringTransactionID: uint(pending.RecurringTransactionID().Value()),
		DueDate:                pending.DueDate(),
		Amount:                 pending.Amount().Amount(),
		Description:            pending.Description(),
		Type:                   string(pending.Type()),
		CategoryID:             uint(pending.CategoryID().Value()),
		CurrencyID:             uint(pending.CurrencyID().Value()),
		Status:                 string(pending.Status()),
		ResolvedAt:             pending.ResolvedAt(),
	}

	if pending.ID().Value() != 0 {
		model.ID = uint(pending.ID().Value())
		return r.db.WithContext(ctx).Model(&PendingTransaction{}).Where("id = ?", model.ID).Updates(map[string]interface{}{
			"status":      model.Status,
			"resolved_at": model.ResolvedAt,
			"updated_at":  time.Now(),
		}).Error
	}

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	pending.AssignID(finance.NewPendingTransactionID(int(model.ID)))
	return nil
}

func (r *GormPendingTransactionRepository) FindByID(ctx context.Context, id finance.PendingTransactionID) (*finance.PendingTransaction, error) {
	var model PendingTransaction
	err := r.db.WithContext(ctx).First(&model, id.Value()).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("pending transaction not found")
		}
		return nil, err
	}
	return r.toDomain(model)
}

func (r *GormPendingTransactionRepository) FindOpenByUserID(ctx context.Context, userID finance.UserID) ([]*finance.PendingTransaction, error) {
	var models []PendingTransaction
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND status = ?", userID.Value(), string(finance.PendingStatusPending)).
		Order("due_date ASC, id ASC").
		Find(&models).Error
	if err != nil {
		return nil, err
	}
	result := make([]*finance.PendingTransaction, 0, len(models))
	for _, m := range models {
		pt, err := r.toDomain(m)
		if err != nil {
			return nil, err
		}
		result = append(result, pt)
	}
	return result, nil
}

func (r *GormPendingTransactionRepository) ExistsByRecurringAndDueDate(
	ctx context.Context,
	recurringID finance.RecurringTransactionID,
	dueDate time.Time,
) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&PendingTransaction{}).
		Where("recurring_transaction_id = ? AND due_date = ?", recurringID.Value(), dueDate.Format("2006-01-02")).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
