package entitlement

import (
	"context"
	"errors"
	"os"
	"strings"
)

var ErrPremiumRequired = errors.New("premium required")

// Checker decides whether a user may use Pro-only features.
type Checker interface {
	IsPro(ctx context.Context, userID int) (bool, error)
}

// InterimChecker bypasses Pro checks until billing entitlements ship.
// When BILLING_ENTITLEMENTS_ENABLED=true, all users are treated as Free
// (real subscription IsPro will replace this in billing PR2).
type InterimChecker struct{}

func NewInterimChecker() *InterimChecker {
	return &InterimChecker{}
}

func (c *InterimChecker) IsPro(_ context.Context, _ int) (bool, error) {
	enabled := strings.EqualFold(strings.TrimSpace(os.Getenv("BILLING_ENTITLEMENTS_ENABLED")), "true")
	if !enabled {
		return true, nil
	}
	return false, nil
}
