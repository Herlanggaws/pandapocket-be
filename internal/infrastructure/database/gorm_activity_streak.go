package database

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// created_at is timestamptz. Shifting to Asia/Jakarta means one sitting counts as one day,
// even when the saved transaction dates span a month.
const activityDatesSQL = `
SELECT activity_date FROM (
	SELECT (created_at AT TIME ZONE 'Asia/Jakarta')::date AS activity_date
	FROM expenses WHERE user_id = ?
	UNION
	SELECT (created_at AT TIME ZONE 'Asia/Jakarta')::date
	FROM incomes WHERE user_id = ?
	UNION
	SELECT (created_at AT TIME ZONE 'Asia/Jakarta')::date
	FROM transfers WHERE user_id = ?
) days
`

type GormActivityDayRepository struct {
	db *gorm.DB
}

func NewGormActivityDayRepository(db *gorm.DB) *GormActivityDayRepository {
	return &GormActivityDayRepository{db: db}
}

func (r *GormActivityDayRepository) ListActivityDates(ctx context.Context, userID int) ([]time.Time, error) {
	var rows []struct {
		ActivityDate time.Time `gorm:"column:activity_date"`
	}
	err := r.db.WithContext(ctx).Raw(activityDatesSQL, userID, userID, userID).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	dates := make([]time.Time, 0, len(rows))
	for _, row := range rows {
		dates = append(dates, row.ActivityDate)
	}
	return dates, nil
}

type GormStreakFreezeRepository struct {
	db *gorm.DB
}

func NewGormStreakFreezeRepository(db *gorm.DB) *GormStreakFreezeRepository {
	return &GormStreakFreezeRepository{db: db}
}

func (r *GormStreakFreezeRepository) FindGap(ctx context.Context, userID int, yearMonth string) (*time.Time, error) {
	var model LoggingStreakFreeze
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND year_month = ?", userID, yearMonth).
		First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	year, month, day := model.GapDate.Date()
	gap := time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
	return &gap, nil
}

func (r *GormStreakFreezeRepository) SaveGap(
	ctx context.Context,
	userID int,
	yearMonth string,
	gapDate, consumedAt time.Time,
) (bool, error) {
	model := LoggingStreakFreeze{
		UserID:     uint(userID),
		YearMonth:  yearMonth,
		GapDate:    gapDate,
		ConsumedAt: consumedAt,
	}
	result := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "year_month"}},
		DoNothing: true,
	}).Create(&model)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}
