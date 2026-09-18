package ticket

import (
	"context"
	"os"
	"strings"
)

// EntitlementChecker decides whether a user may use Pro-only features.
type EntitlementChecker interface {
	IsPro(ctx context.Context, userID int) (bool, error)
}

// InterimEntitlementChecker bypasses Pro checks until billing entitlements ship.
// When BILLING_ENTITLEMENTS_ENABLED=true, all users are treated as Free
// (real subscription IsPro will replace this in billing PR2).
type InterimEntitlementChecker struct{}

func NewInterimEntitlementChecker() *InterimEntitlementChecker {
	return &InterimEntitlementChecker{}
}

func (c *InterimEntitlementChecker) IsPro(_ context.Context, _ int) (bool, error) {
	enabled := strings.EqualFold(strings.TrimSpace(os.Getenv("BILLING_ENTITLEMENTS_ENABLED")), "true")
	if !enabled {
		return true, nil
	}
	return false, nil
}
