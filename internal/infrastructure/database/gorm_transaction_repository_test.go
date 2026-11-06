package database

import (
	"context"
	"strings"
	"testing"
	"time"

	"panda-pocket/internal/domain/finance"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestTransactionSaveSetsCreatedAtOnInsert(t *testing.T) {
	db := openTransactionRepoTestDB(t)
	repo := NewGormTransactionRepository(db)

	tx := newTestExpense(t, 0)
	before := time.Now().UTC().Add(-time.Second)
	if err := repo.Save(context.Background(), tx); err != nil {
		t.Fatal(err)
	}
	after := time.Now().UTC().Add(time.Second)

	var stored Expense
	if err := db.First(&stored, tx.ID().Value()).Error; err != nil {
		t.Fatal(err)
	}
	if stored.CreatedAt.Before(before) || stored.CreatedAt.After(after) {
		t.Fatalf("created_at=%v not in expected window", stored.CreatedAt)
	}
	if stored.CreatedAt.Year() < 1971 {
		t.Fatalf("created_at looks like Go zero time: %v", stored.CreatedAt)
	}
}

func TestTransactionUpdatePreservesCreatedAt(t *testing.T) {
	db := openTransactionRepoTestDB(t)
	repo := NewGormTransactionRepository(db)

	tx := newTestExpense(t, 0)
	if err := repo.Save(context.Background(), tx); err != nil {
		t.Fatal(err)
	}

	createdAt := time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)
	if err := db.Model(&Expense{}).Where("id = ?", tx.ID().Value()).Update("created_at", createdAt).Error; err != nil {
		t.Fatal(err)
	}

	loaded, err := repo.FindByIDAndType(context.Background(), tx.ID(), finance.TransactionTypeExpense)
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.CreatedAt().Equal(createdAt) {
		t.Fatalf("domain created_at=%v want %v", loaded.CreatedAt(), createdAt)
	}

	loaded.UpdateDescription("updated note")
	if err := repo.Save(context.Background(), loaded); err != nil {
		t.Fatal(err)
	}

	var stored Expense
	if err := db.First(&stored, tx.ID().Value()).Error; err != nil {
		t.Fatal(err)
	}
	if !stored.CreatedAt.Equal(createdAt) {
		t.Fatalf("created_at overwritten to %v", stored.CreatedAt)
	}
	if stored.Description != "updated note" {
		t.Fatalf("description=%q", stored.Description)
	}
}

func TestIncomeUpdatePreservesCreatedAt(t *testing.T) {
	db := openTransactionRepoTestDB(t)
	repo := NewGormTransactionRepository(db)

	tx := newTestIncome(t, 0)
	if err := repo.Save(context.Background(), tx); err != nil {
		t.Fatal(err)
	}

	createdAt := time.Date(2024, 5, 1, 8, 30, 0, 0, time.UTC)
	if err := db.Model(&Income{}).Where("id = ?", tx.ID().Value()).Update("created_at", createdAt).Error; err != nil {
		t.Fatal(err)
	}

	loaded, err := repo.FindByIDAndType(context.Background(), tx.ID(), finance.TransactionTypeIncome)
	if err != nil {
		t.Fatal(err)
	}
	loaded.UpdateDescription("salary tweak")
	if err := repo.Save(context.Background(), loaded); err != nil {
		t.Fatal(err)
	}

	var stored Income
	if err := db.First(&stored, tx.ID().Value()).Error; err != nil {
		t.Fatal(err)
	}
	if !stored.CreatedAt.Equal(createdAt) {
		t.Fatalf("created_at overwritten to %v", stored.CreatedAt)
	}
}

func newTestExpense(t *testing.T, id int) *finance.Transaction {
	t.Helper()
	currencyID := finance.NewCurrencyID(1)
	amount, err := finance.NewMoney(12.5, currencyID)
	if err != nil {
		t.Fatal(err)
	}
	return finance.NewTransaction(
		finance.NewTransactionID(id),
		finance.NewUserID(1),
		finance.NewWalletID(1),
		finance.NewCategoryID(1),
		currencyID,
		amount,
		"coffee",
		time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC),
		finance.TransactionTypeExpense,
	)
}

func newTestIncome(t *testing.T, id int) *finance.Transaction {
	t.Helper()
	currencyID := finance.NewCurrencyID(1)
	amount, err := finance.NewMoney(100, currencyID)
	if err != nil {
		t.Fatal(err)
	}
	return finance.NewTransaction(
		finance.NewTransactionID(id),
		finance.NewUserID(1),
		finance.NewWalletID(1),
		finance.NewCategoryID(2),
		currencyID,
		amount,
		"salary",
		time.Date(2024, 4, 2, 0, 0, 0, 0, time.UTC),
		finance.TransactionTypeIncome,
	)
}

func openTransactionRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	name := strings.ReplaceAll(t.Name(), "/", "_")
	db, err := gorm.Open(sqlite.Open("file:"+name+"?mode=memory&cache=shared"), &gorm.Config{
		Logger:                                   logger.Default.LogMode(logger.Silent),
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&Expense{}, &Income{}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
	return db
}
