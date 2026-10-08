package database

import (
	"context"
	"errors"
	"time"

	domainBilling "panda-pocket/internal/domain/billing"

	"gorm.io/gorm"
)

type GormPendingPaymentRepository struct {
	db *gorm.DB
}

func NewGormPendingPaymentRepository(db *gorm.DB) *GormPendingPaymentRepository {
	return &GormPendingPaymentRepository{db: db}
}

func (r *GormPendingPaymentRepository) Save(ctx context.Context, payment domainBilling.PendingPayment) error {
	row := PendingPayment{
		PaymentID: payment.PaymentID,
		UserID:    uint(payment.UserID),
		Kind:      payment.Kind,
		Interval:  payment.Interval,
		Pack:      payment.Pack,
		Amount:    payment.Amount,
		CreatedAt: time.Now().UTC(),
	}
	return r.db.WithContext(ctx).Create(&row).Error
}

func (r *GormPendingPaymentRepository) FindByPaymentID(ctx context.Context, paymentID string) (domainBilling.PendingPayment, bool, error) {
	var row PendingPayment
	err := r.db.WithContext(ctx).Where("payment_id = ?", paymentID).First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domainBilling.PendingPayment{}, false, nil
	}
	if err != nil {
		return domainBilling.PendingPayment{}, false, err
	}
	return domainBilling.PendingPayment{
		PaymentID: row.PaymentID,
		UserID:    int(row.UserID),
		Kind:      row.Kind,
		Interval:  row.Interval,
		Pack:      row.Pack,
		Amount:    row.Amount,
	}, true, nil
}
