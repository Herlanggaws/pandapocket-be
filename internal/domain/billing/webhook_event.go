package billing

import (
	"context"
)

// WebhookEventRepository stores webhook receipts for dedup.
type WebhookEventRepository interface {
	// Exists reports whether eventID was already stored.
	Exists(ctx context.Context, eventID string) (bool, error)
	// TryInsert returns inserted=false when eventID already exists.
	TryInsert(ctx context.Context, eventID, payload string) (inserted bool, err error)
}
