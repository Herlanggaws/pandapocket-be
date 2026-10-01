package database

import (
	"context"
	"errors"
	"testing"
	"time"

	appMCP "panda-pocket/internal/application/mcp"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestMcpTokenRotateClearsRevoke(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:mcp_rotate?mode=memory&cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&McpToken{}); err != nil {
		t.Fatal(err)
	}
	repo := NewGormMcpTokenRepository(db)
	ctx := context.Background()
	first := appMCP.StoredToken{UserID: 1, TokenHash: "hash-one", Prefix: "bb_mcp_aaaaaaaa", CreatedAt: time.Now().UTC()}
	if err := repo.Save(ctx, first); err != nil {
		t.Fatal(err)
	}
	if err := repo.Revoke(ctx, 1, time.Now().UTC()); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.FindActiveByHash(ctx, "hash-one"); !errors.Is(err, appMCP.ErrNotFound) {
		t.Fatalf("revoked lookup err=%v", err)
	}
	second := appMCP.StoredToken{UserID: 1, TokenHash: "hash-two", Prefix: "bb_mcp_bbbbbbbb", CreatedAt: time.Now().UTC()}
	if err := repo.Save(ctx, second); err != nil {
		t.Fatal(err)
	}
	active, err := repo.FindActiveByHash(ctx, "hash-two")
	if err != nil || active.UserID != 1 {
		t.Fatalf("active=%+v err=%v", active, err)
	}
}
