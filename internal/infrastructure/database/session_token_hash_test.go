package database

import (
	"testing"
	"time"

	domainIdentity "panda-pocket/internal/domain/identity"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestHashStoredSessionTokensIsIdempotent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:session_hash?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TABLE tokens (
		id text PRIMARY KEY,
		user_id integer NOT NULL,
		access_token text NOT NULL,
		refresh_token text NOT NULL,
		expires_at datetime NOT NULL,
		revoked numeric DEFAULT 0,
		created_at datetime,
		updated_at datetime
	)`).Error; err != nil {
		t.Fatal(err)
	}
	rawAccess := "eyJhbGciOiJIUzI1NiJ9.access.sig"
	rawRefresh := "eyJhbGciOiJIUzI1NiJ9.refresh.sig"
	alreadyHashed := domainIdentity.SessionTokenHash("already-hashed-token")
	row := Token{
		ID:           uuid.New(),
		UserID:       4,
		AccessToken:  rawAccess,
		RefreshToken: rawRefresh,
		ExpiresAt:    time.Now().Add(time.Hour),
	}
	kept := Token{
		ID:           uuid.New(),
		UserID:       9,
		AccessToken:  alreadyHashed,
		RefreshToken: alreadyHashed,
		ExpiresAt:    time.Now().Add(time.Hour),
	}
	if err := db.Create(&row).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&kept).Error; err != nil {
		t.Fatal(err)
	}

	if err := hashStoredSessionTokens(db); err != nil {
		t.Fatal(err)
	}
	if err := hashStoredSessionTokens(db); err != nil {
		t.Fatal(err)
	}

	var stored Token
	if err := db.First(&stored, "id = ?", row.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stored.AccessToken != domainIdentity.SessionTokenHash(rawAccess) {
		t.Fatalf("access hash = %s", stored.AccessToken)
	}
	if stored.RefreshToken != domainIdentity.SessionTokenHash(rawRefresh) {
		t.Fatalf("refresh hash = %s", stored.RefreshToken)
	}
	var unchanged Token
	if err := db.First(&unchanged, "id = ?", kept.ID).Error; err != nil {
		t.Fatal(err)
	}
	if unchanged.AccessToken != alreadyHashed || unchanged.RefreshToken != alreadyHashed {
		t.Fatal("hashed row was rewritten")
	}

	repo := NewGormTokenRepository(db)
	found, revoked, err := repo.FindByAccessToken(t.Context(), domainIdentity.SessionTokenHash(rawAccess))
	if err != nil || !found || revoked {
		t.Fatalf("access lookup found=%v revoked=%v err=%v", found, revoked, err)
	}
	userID, refreshRevoked, err := repo.FindByRefreshToken(t.Context(), domainIdentity.SessionTokenHash(rawRefresh))
	if err != nil || userID != 4 || refreshRevoked {
		t.Fatalf("refresh lookup user=%d revoked=%v err=%v", userID, refreshRevoked, err)
	}
}
