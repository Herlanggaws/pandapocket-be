package database

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupWipeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	err = db.AutoMigrate(
		&User{},
		&Currency{},
		&Category{},
		&Wallet{},
		&Expense{},
		&Income{},
		&Budget{},
		&FinancialGoal{},
		&Asset{},
		&Liability{},
		&LiabilityPayment{},
		&GoalContribution{},
		&HealthScoreSnapshot{},
		&RecurringTransaction{},
		&PendingTransaction{},
		&Transfer{},
		&UserPreferences{},
		&Notification{},
		&UserFeedback{},
		&SupportTicket{},
		&PasswordResetToken{},
		&AccountResetChallenge{},
	)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestWipeUserDataScopedByUserID(t *testing.T) {
	db := setupWipeTestDB(t)
	wiper := NewGormUserDataWiper(db)
	ctx := context.Background()

	userA := User{Email: "a@example.com", PasswordHash: "x", Role: "user"}
	userB := User{Email: "b@example.com", PasswordHash: "x", Role: "user"}
	if err := db.Create(&userA).Error; err != nil {
		t.Fatalf("create userA: %v", err)
	}
	if err := db.Create(&userB).Error; err != nil {
		t.Fatalf("create userB: %v", err)
	}

	systemCategory := Category{Name: "Food", Color: "#000", IsDefault: true, CategoryType: "expense"}
	userACategory := Category{UserID: &userA.ID, Name: "Custom A", Color: "#111", CategoryType: "expense"}
	userBCategory := Category{UserID: &userB.ID, Name: "Custom B", Color: "#222", CategoryType: "expense"}
	if err := db.Create(&systemCategory).Error; err != nil {
		t.Fatalf("system category: %v", err)
	}
	if err := db.Create(&userACategory).Error; err != nil {
		t.Fatalf("userA category: %v", err)
	}
	if err := db.Create(&userBCategory).Error; err != nil {
		t.Fatalf("userB category: %v", err)
	}

	systemCurrency := Currency{Code: "IDR", Name: "Rupiah", Symbol: "Rp", IsSystem: true}
	userACurrencyID := userA.ID
	userACurrency := Currency{UserID: &userACurrencyID, Code: "AAA", Name: "Aaa", Symbol: "A"}
	userBCurrencyID := userB.ID
	userBCurrency := Currency{UserID: &userBCurrencyID, Code: "BBB", Name: "Bbb", Symbol: "B"}
	if err := db.Create(&systemCurrency).Error; err != nil {
		t.Fatalf("system currency: %v", err)
	}
	if err := db.Create(&userACurrency).Error; err != nil {
		t.Fatalf("userA currency: %v", err)
	}
	if err := db.Create(&userBCurrency).Error; err != nil {
		t.Fatalf("userB currency: %v", err)
	}

	walletA := Wallet{
		UserID:         userA.ID,
		Name:           "Cash A",
		Type:           "cash",
		CurrencyID:     systemCurrency.ID,
		OpeningBalance: 100,
		IsDefault:      true,
	}
	walletB := Wallet{
		UserID:         userB.ID,
		Name:           "Cash B",
		Type:           "cash",
		CurrencyID:     systemCurrency.ID,
		OpeningBalance: 200,
		IsDefault:      true,
	}
	if err := db.Create(&walletA).Error; err != nil {
		t.Fatalf("walletA: %v", err)
	}
	if err := db.Create(&walletB).Error; err != nil {
		t.Fatalf("walletB: %v", err)
	}

	expenseA := Expense{
		UserID:      userA.ID,
		WalletID:    &walletA.ID,
		CategoryID:  userACategory.ID,
		CurrencyID:  systemCurrency.ID,
		Amount:      50,
		Description: "A expense",
		Date:        time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
	}
	expenseB := Expense{
		UserID:      userB.ID,
		WalletID:    &walletB.ID,
		CategoryID:  userBCategory.ID,
		CurrencyID:  systemCurrency.ID,
		Amount:      75,
		Description: "B expense",
		Date:        time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
	}
	if err := db.Create(&expenseA).Error; err != nil {
		t.Fatalf("expenseA: %v", err)
	}
	if err := db.Create(&expenseB).Error; err != nil {
		t.Fatalf("expenseB: %v", err)
	}

	prefsA := UserPreferences{UserID: userA.ID, PrimaryCurrencyID: systemCurrency.ID}
	prefsB := UserPreferences{UserID: userB.ID, PrimaryCurrencyID: systemCurrency.ID}
	if err := db.Create(&prefsA).Error; err != nil {
		t.Fatalf("prefsA: %v", err)
	}
	if err := db.Create(&prefsB).Error; err != nil {
		t.Fatalf("prefsB: %v", err)
	}

	if err := wiper.WipeUserData(ctx, userA.ID); err != nil {
		t.Fatalf("WipeUserData: %v", err)
	}

	var expenseCountA, expenseCountB int64
	db.Model(&Expense{}).Where("user_id = ?", userA.ID).Count(&expenseCountA)
	db.Model(&Expense{}).Where("user_id = ?", userB.ID).Count(&expenseCountB)
	if expenseCountA != 0 {
		t.Fatalf("userA expenses should be wiped, got %d", expenseCountA)
	}
	if expenseCountB != 1 {
		t.Fatalf("userB expenses should remain, got %d", expenseCountB)
	}

	var walletCountA, walletCountB int64
	db.Model(&Wallet{}).Where("user_id = ?", userA.ID).Count(&walletCountA)
	db.Model(&Wallet{}).Where("user_id = ?", userB.ID).Count(&walletCountB)
	if walletCountA != 0 || walletCountB != 1 {
		t.Fatalf("wallets A=%d B=%d", walletCountA, walletCountB)
	}

	var categoryUserA, categoryUserB, categorySystem int64
	db.Model(&Category{}).Where("user_id = ?", userA.ID).Count(&categoryUserA)
	db.Model(&Category{}).Where("user_id = ?", userB.ID).Count(&categoryUserB)
	db.Model(&Category{}).Where("user_id IS NULL").Count(&categorySystem)
	if categoryUserA != 0 {
		t.Fatalf("userA categories should be wiped")
	}
	if categoryUserB != 1 {
		t.Fatalf("userB categories should remain")
	}
	if categorySystem != 1 {
		t.Fatalf("system categories must not be deleted, got %d", categorySystem)
	}

	var currencyUserA, currencySystem int64
	db.Model(&Currency{}).Where("user_id = ?", userA.ID).Count(&currencyUserA)
	db.Model(&Currency{}).Where("user_id IS NULL").Count(&currencySystem)
	if currencyUserA != 0 {
		t.Fatalf("userA currencies should be wiped")
	}
	if currencySystem != 1 {
		t.Fatalf("system currencies must not be deleted")
	}

	var prefsCountA, prefsCountB int64
	db.Model(&UserPreferences{}).Where("user_id = ?", userA.ID).Count(&prefsCountA)
	db.Model(&UserPreferences{}).Where("user_id = ?", userB.ID).Count(&prefsCountB)
	if prefsCountA != 0 || prefsCountB != 1 {
		t.Fatalf("prefs A=%d B=%d", prefsCountA, prefsCountB)
	}

	var userCount int64
	db.Model(&User{}).Where("id = ?", userA.ID).Count(&userCount)
	if userCount != 1 {
		t.Fatalf("user row must remain after wipe")
	}
}
