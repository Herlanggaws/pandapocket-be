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

func TestRecordSpendWritesLedgerWithBalance(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:ai_credits_spend?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("sqlite: %v", err)
	}
	if err := db.AutoMigrate(&database.AICreditBalance{}, &database.AICreditLedgerEntry{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	repo := database.NewGormAICreditRepository(db)
	ctx := context.Background()
	balance := &database.AICreditBalance{
		UserID:           7,
		IncludedGranted:  75,
		IncludedUnlocked: 75,
		UpdatedAt:        time.Now().UTC(),
	}
	if err := db.Create(balance).Error; err != nil {
		t.Fatal(err)
	}

	after, err := repo.RecordSpend(ctx, 7, "advisor_chat")
	if err != nil {
		t.Fatal(err)
	}
	if after.IncludedUsed != 1 || after.Available() != 74 {
		t.Fatalf("balance after spend: used=%d available=%d", after.IncludedUsed, after.Available())
	}

	rows, err := repo.ListLedger(ctx, 7, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Kind != "spend" || rows[0].Source != "advisor_chat" || rows[0].Delta() != -1 {
		t.Fatalf("ledger=%+v", rows)
	}
	if rows[0].DeltaIncludedUsed != 1 {
		t.Fatalf("included delta=%d", rows[0].DeltaIncludedUsed)
	}
}
