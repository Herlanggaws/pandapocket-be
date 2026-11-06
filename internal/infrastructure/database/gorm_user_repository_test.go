package database

import (
	"context"
	"strings"
	"testing"
	"time"

	"panda-pocket/internal/domain/identity"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSaveSetsCreatedAtOnInsert(t *testing.T) {
	db := openUserRepoTestDB(t)
	repo := NewGormUserRepository(db)
	user := newTestUser(t, "create@example.com")

	before := time.Now().UTC().Add(-time.Second)
	if err := repo.Save(context.Background(), user); err != nil {
		t.Fatal(err)
	}
	after := time.Now().UTC().Add(time.Second)

	var stored User
	if err := db.First(&stored, user.ID().Value()).Error; err != nil {
		t.Fatal(err)
	}
	if stored.CreatedAt.Before(before) || stored.CreatedAt.After(after) {
		t.Fatalf("created_at=%v not in expected window", stored.CreatedAt)
	}
	if stored.CreatedAt.Year() < 1971 {
		t.Fatalf("created_at looks like Go zero time: %v", stored.CreatedAt)
	}
}

func TestUpdatePreservesCreatedAtAndLastLoginAt(t *testing.T) {
	db := openUserRepoTestDB(t)
	repo := NewGormUserRepository(db)
	user := newTestUser(t, "update@example.com")
	if err := repo.Save(context.Background(), user); err != nil {
		t.Fatal(err)
	}

	createdAt := time.Date(2024, 3, 15, 10, 0, 0, 0, time.UTC)
	lastLogin := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	if err := db.Model(&User{}).Where("id = ?", user.ID().Value()).Updates(map[string]interface{}{
		"created_at":    createdAt,
		"last_login_at": lastLogin,
	}).Error; err != nil {
		t.Fatal(err)
	}

	loaded, err := repo.FindByID(context.Background(), user.ID())
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.CreatedAt().Equal(createdAt) {
		t.Fatalf("domain created_at=%v want %v", loaded.CreatedAt(), createdAt)
	}

	newHash, err := identity.NewPasswordHashFromPlain("new-password-ok")
	if err != nil {
		t.Fatal(err)
	}
	loaded.UpdatePassword(newHash)
	if err := repo.Update(context.Background(), loaded); err != nil {
		t.Fatal(err)
	}

	var stored User
	if err := db.First(&stored, user.ID().Value()).Error; err != nil {
		t.Fatal(err)
	}
	if !stored.CreatedAt.Equal(createdAt) {
		t.Fatalf("created_at overwritten to %v", stored.CreatedAt)
	}
	if stored.LastLoginAt == nil || !stored.LastLoginAt.Equal(lastLogin) {
		t.Fatalf("last_login_at overwritten to %v", stored.LastLoginAt)
	}
}

func TestUpdateLastLoginAt(t *testing.T) {
	db := openUserRepoTestDB(t)
	repo := NewGormUserRepository(db)
	user := newTestUser(t, "login@example.com")
	if err := repo.Save(context.Background(), user); err != nil {
		t.Fatal(err)
	}

	at := time.Date(2026, 11, 6, 9, 0, 0, 0, time.UTC)
	if err := repo.UpdateLastLoginAt(context.Background(), user.ID(), at); err != nil {
		t.Fatal(err)
	}

	var stored User
	if err := db.First(&stored, user.ID().Value()).Error; err != nil {
		t.Fatal(err)
	}
	if stored.LastLoginAt == nil || !stored.LastLoginAt.Equal(at) {
		t.Fatalf("last_login_at=%v want %v", stored.LastLoginAt, at)
	}
}

func newTestUser(t *testing.T, email string) *identity.User {
	t.Helper()
	emailVO, err := identity.NewEmail(email)
	if err != nil {
		t.Fatal(err)
	}
	role, err := identity.NewRole("user")
	if err != nil {
		t.Fatal(err)
	}
	return identity.NewUser(identity.UserID{}, emailVO, identity.NewPasswordHash("hash"), role)
}

func openUserRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	name := strings.ReplaceAll(t.Name(), "/", "_")
	db, err := gorm.Open(sqlite.Open("file:"+name+"?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&User{}); err != nil {
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
