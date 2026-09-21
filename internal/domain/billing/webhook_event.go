package billing

import (
	"context"
)

// WebhookEventRepository stores webhook receipts for dedup.
type WebhookEventRepository interface {
	// TryInsert returns inserted=false when eventID already exists.
	TryInsert(ctx context.Context, eventID, payload string) (inserted bool, err error)
}
