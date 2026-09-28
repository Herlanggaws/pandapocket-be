package database

import (
	"context"
	"time"

	domainFinance "panda-pocket/internal/domain/finance"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// FxRate is the cached ECB quote for one ISO currency.
type FxRate struct {
	ID           uint      `gorm:"primaryKey"`
	CurrencyCode string    `gorm:"uniqueIndex;not null;size:8"`
	UnitsPerEUR  float64   `gorm:"type:decimal(24,8);not null"`
	AsOf         time.Time `gorm:"type:date;not null"`
	FetchedAt    time.Time `gorm:"not null"`
}

type GormFxRateRepository struct {
	db *gorm.DB
}

func NewGormFxRateRepository(db *gorm.DB) *GormFxRateRepository {
	return &GormFxRateRepository{db: db}
}

func (r *GormFxRateRepository) LatestFetchedAt(ctx context.Context) (*time.Time, error) {
	var row FxRate
	err := r.db.WithContext(ctx).Order("fetched_at DESC").Limit(1).Find(&row).Error
	if err != nil {
		return nil, err
	}
	if row.ID == 0 {
		return nil, nil
	}
	fetchedAt := row.FetchedAt
	return &fetchedAt, nil
}

func (r *GormFxRateRepository) Upsert(ctx context.Context, quotes []domainFinance.FxQuote) error {
	if len(quotes) == 0 {
		return nil
	}
	fetchedAt := time.Now().UTC()
	rows := make([]FxRate, 0, len(quotes))
	for _, quote := range quotes {
		if quote.Code == "" || quote.UnitsPerEUR <= 0 {
			continue
		}
		rows = append(rows, FxRate{
			CurrencyCode: quote.Code,
			UnitsPerEUR:  quote.UnitsPerEUR,
			AsOf:         quote.AsOf,
			FetchedAt:    fetchedAt,
		})
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "currency_code"}},
		DoUpdates: clause.AssignmentColumns([]string{"units_per_eur", "as_of", "fetched_at"}),
	}).Create(&rows).Error
}

func (r *GormFxRateRepository) List(ctx context.Context) ([]domainFinance.FxQuote, error) {
	var rows []FxRate
	if err := r.db.WithContext(ctx).Find(&rows).Error; err != nil {
		return nil, err
	}
	quotes := make([]domainFinance.FxQuote, 0, len(rows))
	for _, row := range rows {
		quotes = append(quotes, domainFinance.FxQuote{
			Code:        row.CurrencyCode,
			UnitsPerEUR: row.UnitsPerEUR,
			AsOf:        row.AsOf,
		})
	}
	return quotes, nil
}
