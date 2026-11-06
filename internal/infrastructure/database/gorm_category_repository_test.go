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

func TestCategoryUpdatePreservesCreatedAt(t *testing.T) {
	db := openCategoryRepoTestDB(t)
	repo := NewGormCategoryRepository(db)

	userID := finance.NewUserID(1)
	category, err := finance.NewCategory(
		finance.NewCategoryID(0),
		&userID,
		"Food",
		"#ff0000",
		false,
		finance.CategoryTypeExpense,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(context.Background(), category); err != nil {
		t.Fatal(err)
	}

	var inserted Category
	if err := db.Where("name = ?", "Food").First(&inserted).Error; err != nil {
		t.Fatal(err)
	}
	createdAt := time.Date(2024, 2, 1, 9, 0, 0, 0, time.UTC)
	if err := db.Model(&Category{}).Where("id = ?", inserted.ID).Update("created_at", createdAt).Error; err != nil {
		t.Fatal(err)
	}

	loaded, err := repo.FindByID(context.Background(), finance.NewCategoryID(int(inserted.ID)))
	if err != nil {
		t.Fatal(err)
	}
	if err := loaded.UpdateName("Groceries"); err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(context.Background(), loaded); err != nil {
		t.Fatal(err)
	}

	var stored Category
	if err := db.First(&stored, inserted.ID).Error; err != nil {
		t.Fatal(err)
	}
	if !stored.CreatedAt.Equal(createdAt) {
		t.Fatalf("created_at overwritten to %v", stored.CreatedAt)
	}
	if stored.Name != "Groceries" {
		t.Fatalf("name=%q", stored.Name)
	}
}

func openCategoryRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	name := strings.ReplaceAll(t.Name(), "/", "_")
	db, err := gorm.Open(sqlite.Open("file:"+name+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&Category{}); err != nil {
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
