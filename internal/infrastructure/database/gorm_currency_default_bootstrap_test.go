package database

import (
	"context"
	"testing"

	"panda-pocket/internal/domain/finance"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupCurrencyBootstrapDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&Currency{}, &UserPreferences{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	idr := Currency{Code: "IDR", Name: "Rupiah", Symbol: "Rp", IsSystem: true}
	if err := db.Create(&idr).Error; err != nil {
		t.Fatalf("seed IDR: %v", err)
	}
	return db
}

func TestGetUserDefaultCurrency_BootstrapsPreferences(t *testing.T) {
	db := setupCurrencyBootstrapDB(t)
	repo := NewGormCurrencyRepository(db)
	ctx := context.Background()
	userID := finance.NewUserID(42)

	currency, err := repo.GetUserDefaultCurrency(ctx, userID)
	if err != nil {
		t.Fatalf("GetUserDefaultCurrency: %v", err)
	}
	if currency.Code() != "IDR" {
		t.Fatalf("want IDR, got %s", currency.Code())
	}

	var prefs UserPreferences
	if err := db.Where("user_id = ?", userID.Value()).First(&prefs).Error; err != nil {
		t.Fatalf("prefs not created: %v", err)
	}
	if int(prefs.PrimaryCurrencyID) != currency.ID().Value() {
		t.Fatalf("prefs currency mismatch: %d vs %d", prefs.PrimaryCurrencyID, currency.ID().Value())
	}
}
