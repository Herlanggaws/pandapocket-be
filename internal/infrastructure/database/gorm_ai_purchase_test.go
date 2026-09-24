package database_test

import (
	"context"
	"testing"
	"time"

	"panda-pocket/internal/infrastructure/database"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRecordPurchaseAccumulatesDistinctPayments(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:ai_credits_accum?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	if err := db.AutoMigrate(&database.AICreditBalance{}, &database.AICreditLedgerEntry{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	repo := database.NewGormAICreditRepository(db)
	ctx := context.Background()

	if err := repo.RecordPurchase(ctx, 7, "ai_credits_s", "pay_s1", 50); err != nil {
		t.Fatal(err)
	}
	if err := repo.RecordPurchase(ctx, 7, "ai_credits_m", "pay_m1", 150); err != nil {
		t.Fatal(err)
	}
	// replay same payment must not double-credit
	if err := repo.RecordPurchase(ctx, 7, "ai_credits_m", "pay_m1", 150); err != nil {
		t.Fatal(err)
	}

	bal, err := repo.FindByUserID(ctx, 7)
	if err != nil {
		t.Fatal(err)
	}
	if bal.PurchasedRemaining != 200 {
		t.Fatalf("expected 200 purchased, got %d", bal.PurchasedRemaining)
	}
	_ = time.Now()
}
