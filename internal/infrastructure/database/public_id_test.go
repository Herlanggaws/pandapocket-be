package database

import (
	"context"
	"strings"
	"testing"

	"panda-pocket/internal/domain/identity"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestSaveAssignsPublicID(t *testing.T) {
	db := openPublicIDTestDB(t)
	repo := NewGormUserRepository(db)
	email, err := identity.NewEmail("new@example.com")
	if err != nil {
		t.Fatal(err)
	}
	role, err := identity.NewRole("user")
	if err != nil {
		t.Fatal(err)
	}
	user := identity.NewUser(identity.UserID{}, email, identity.NewPasswordHash("hash"), role)
	if err := repo.Save(context.Background(), user); err != nil {
		t.Fatal(err)
	}
	if _, err := uuid.Parse(user.PublicID()); err != nil {
		t.Fatalf("public id %q: %v", user.PublicID(), err)
	}

	loaded, err := repo.FindByID(context.Background(), user.ID())
	if err != nil {
		t.Fatal(err)
	}
	if loaded.PublicID() != user.PublicID() {
		t.Fatalf("loaded %q saved %q", loaded.PublicID(), user.PublicID())
	}
	lookedUp, err := repo.PublicID(context.Background(), user.ID().Value())
	if err != nil {
		t.Fatal(err)
	}
	if lookedUp != user.PublicID() {
		t.Fatalf("lookup %q saved %q", lookedUp, user.PublicID())
	}
}

func TestFillBlankPublicIDsKeepsExisting(t *testing.T) {
	db := openPublicIDTestDB(t)
	kept := "11111111-1111-4111-8111-111111111111"
	existing := User{Email: "kept@example.com", PasswordHash: "x", Role: "user", PublicID: kept}
	blank := User{Email: "blank@example.com", PasswordHash: "x", Role: "user"}
	if err := db.Create(&existing).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&blank).Error; err != nil {
		t.Fatal(err)
	}

	if err := fillBlankPublicIDs(db); err != nil {
		t.Fatal(err)
	}

	var storedExisting, storedBlank User
	if err := db.First(&storedExisting, existing.ID).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.First(&storedBlank, blank.ID).Error; err != nil {
		t.Fatal(err)
	}
	if storedExisting.PublicID != kept {
		t.Fatalf("existing public id changed to %q", storedExisting.PublicID)
	}
	if _, err := uuid.Parse(storedBlank.PublicID); err != nil {
		t.Fatalf("blank row public id %q: %v", storedBlank.PublicID, err)
	}
	if storedBlank.PublicID == kept {
		t.Fatal("blank row reused an existing public id")
	}
}

func openPublicIDTestDB(t *testing.T) *gorm.DB {
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
