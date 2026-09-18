package finance

import (
	"context"
	"errors"
	"testing"

	"panda-pocket/internal/domain/entitlement"
)

func TestGetAnalyticsUseCase_FreeBlocksYearlyAndCustom(t *testing.T) {
	uc := &GetAnalyticsUseCase{entitlements: entitlement.StaticChecker{Pro: false}}

	_, err := uc.Execute(context.Background(), 1, GetAnalyticsRequest{Period: "yearly"})
	if !errors.Is(err, entitlement.ErrPremiumRequired) {
		t.Fatalf("expected premium required for yearly, got %v", err)
	}

	_, err = uc.Execute(context.Background(), 1, GetAnalyticsRequest{
		Period:    "monthly",
		StartDate: "2026-01-01",
		EndDate:   "2026-01-31",
	})
	if !errors.Is(err, entitlement.ErrPremiumRequired) {
		t.Fatalf("expected premium required for custom range, got %v", err)
	}
}
