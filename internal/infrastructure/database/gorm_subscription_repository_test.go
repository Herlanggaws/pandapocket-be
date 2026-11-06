package database

import (
	"context"
	"strings"
	"testing"
	"time"

	"panda-pocket/internal/domain/billing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSubscriptionUpdatePreservesCreatedAt(t *testing.T) {
	db := openSubscriptionRepoTestDB(t)
	repo := NewGormSubscriptionRepository(db)

	sub, err := billing.NewFreeSubscription(7, "pub-test-7")
	if err != nil {
		t.Fatal(err)
	}
	if err := repo.Save(context.Background(), sub); err != nil {
		t.Fatal(err)
	}

	createdAt := time.Date(2024, 1, 10, 12, 0, 0, 0, time.UTC)
	if err := db.Model(&Subscription{}).Where("id = ?", sub.ID().Value()).Update("created_at", createdAt).Error; err != nil {
		t.Fatal(err)
	}

	loaded, err := repo.FindByUserID(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	loaded.DowngradeToFree(time.Now().UTC())
	if err := repo.Save(context.Background(), loaded); err != nil {
		t.Fatal(err)
	}

	var stored Subscription
	if err := db.First(&stored, sub.ID().Value()).Error; err != nil {
		t.Fatal(err)
	}
	if !stored.CreatedAt.Equal(createdAt) {
		t.Fatalf("created_at overwritten to %v", stored.CreatedAt)
	}
}

func openSubscriptionRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	name := strings.ReplaceAll(t.Name(), "/", "_")
	db, err := gorm.Open(sqlite.Open("file:"+name+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&Subscription{}); err != nil {
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
