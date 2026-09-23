package database

import (
	"context"
	"testing"
	"time"

	"panda-pocket/internal/domain/finance"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupWalletCurrencyTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&Wallet{}, &PendingTransaction{}, &RecurringTransaction{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestGormWalletRepository_SavePersistsCurrencyID(t *testing.T) {
	db := setupWalletCurrencyTestDB(t)
	repo := NewGormWalletRepository(db)
	ctx := context.Background()

	wallet := finance.ReconstituteWallet(
		finance.NewWalletID(0),
		finance.NewUserID(1),
		"Cash",
		finance.WalletTypeCash,
		finance.NewCurrencyID(1),
		0,
		true,
		false,
		time.Now(),
	)
	if err := repo.Save(ctx, wallet); err != nil {
		t.Fatalf("create: %v", err)
	}

	wallet.UpdateCurrencyID(finance.NewCurrencyID(21))
	if err := repo.Save(ctx, wallet); err != nil {
		t.Fatalf("update: %v", err)
	}

	loaded, err := repo.FindByID(ctx, wallet.ID())
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if loaded.CurrencyID().Value() != 21 {
		t.Fatalf("want currency_id 21, got %d", loaded.CurrencyID().Value())
	}
}

func TestGormWalletRepository_AlignRecurringCurrency(t *testing.T) {
	db := setupWalletCurrencyTestDB(t)
	repo := NewGormWalletRepository(db)
	ctx := context.Background()

	walletModel := Wallet{
		UserID:         1,
		Name:           "Cash",
		Type:           "cash",
		CurrencyID:     1,
		OpeningBalance: 0,
		IsDefault:      true,
	}
	if err := db.Create(&walletModel).Error; err != nil {
		t.Fatalf("create wallet: %v", err)
	}

	walletID := walletModel.ID
	recurring := RecurringTransaction{
		UserID:      1,
		WalletID:    &walletID,
		CurrencyID:  1,
		CategoryID:  1,
		Amount:      100,
		Description: "rent",
		Frequency:   "monthly",
		Type:        "expense",
		DayOfMonth:  intPtr(1),
		NextDueDate: time.Now().Add(24 * time.Hour),
		IsActive:    true,
	}
	if err := db.Create(&recurring).Error; err != nil {
		t.Fatalf("create recurring: %v", err)
	}

	if err := repo.AlignRecurringCurrency(ctx, finance.NewWalletID(int(walletModel.ID)), finance.NewCurrencyID(21)); err != nil {
		t.Fatalf("align: %v", err)
	}

	var updated RecurringTransaction
	if err := db.First(&updated, recurring.ID).Error; err != nil {
		t.Fatalf("reload: %v", err)
	}
	if updated.CurrencyID != 21 {
		t.Fatalf("want recurring currency 21, got %d", updated.CurrencyID)
	}
}

func intPtr(v int) *int { return &v }
