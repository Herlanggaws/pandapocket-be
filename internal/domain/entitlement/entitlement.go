package entitlement

import (
	"context"
	"errors"
	"fmt"
	"time"

	"panda-pocket/internal/domain/billing"
)

var ErrPremiumRequired = errors.New("premium required")

const (
	FreeTransactionsPerMonth = 50
	FreeCustomCategories     = 10
	FreeActiveBudgets        = 3
	FreeWallets              = 1
	FreeAssets               = 1
	FreeDebts                = 1
)

const (
	FeatureTransactions = "transactions"
	FeatureCategories   = "categories"
	FeatureBudgets      = "budgets"
	FeatureRecurring    = "recurring"
	FeatureTickets      = "tickets"
	FeatureExport       = "export"
	FeatureWallets      = "wallets"
	FeatureInsights     = "insights"
	FeatureAssets       = "assets"
	FeatureDebts        = "debts"
	FeatureAIAdvisor    = "ai_advisor"
)

// Checker decides whether a user may use Pro-only features.
type Checker interface {
	IsPro(ctx context.Context, userID int) (bool, error)
}

// LimitExceeded is returned when a Free user hits a plan limit.
// It unwraps to ErrPremiumRequired so errors.Is keeps working.
type LimitExceeded struct {
	Feature string
	Limit   int
	Used    int
}

func (e *LimitExceeded) Error() string {
	if e.Limit == 0 {
		return fmt.Sprintf("%s requires Pro", e.Feature)
	}
	return fmt.Sprintf("%s limit reached on Free plan (%d/%d). Upgrade to Pro.", e.Feature, e.Used, e.Limit)
}

func (e *LimitExceeded) Unwrap() error {
	return ErrPremiumRequired
}

// CheckFreeLimit returns LimitExceeded when used >= limit.
func CheckFreeLimit(feature string, used, limit int) error {
	if used >= limit {
		return &LimitExceeded{Feature: feature, Limit: limit, Used: used}
	}
	return nil
}

// RequirePro returns LimitExceeded for Pro-only features (limit 0).
func RequirePro(feature string) error {
	return &LimitExceeded{Feature: feature, Limit: 0, Used: 0}
}

// EnforceCreateLimit skips Free limits when the user is Pro.
// Callers may set WithEntitlementBypass for trusted bootstrap paths (e.g. onboarding seed).
func EnforceCreateLimit(ctx context.Context, checker Checker, userID int, feature string, used, limit int) error {
	if HasEntitlementBypass(ctx) {
		return nil
	}
	isPro, err := checker.IsPro(ctx, userID)
	if err != nil {
		return err
	}
	if isPro {
		return nil
	}
	return CheckFreeLimit(feature, used, limit)
}

type entitlementBypassKey struct{}

func WithEntitlementBypass(ctx context.Context) context.Context {
	return context.WithValue(ctx, entitlementBypassKey{}, true)
}

func HasEntitlementBypass(ctx context.Context) bool {
	v, _ := ctx.Value(entitlementBypassKey{}).(bool)
	return v
}

// StaticChecker is a fixed Pro/Free stub for tests.
type StaticChecker struct {
	Pro bool
}

func (c StaticChecker) IsPro(_ context.Context, _ int) (bool, error) {
	return c.Pro, nil
}

// SubscriptionChecker uses billing subscription state for IsPro.
type SubscriptionChecker struct {
	repo billing.SubscriptionRepository
	now  func() time.Time
}

func NewSubscriptionChecker(repo billing.SubscriptionRepository) *SubscriptionChecker {
	return &SubscriptionChecker{
		repo: repo,
		now:  func() time.Time { return time.Now().UTC() },
	}
}

func (c *SubscriptionChecker) IsPro(ctx context.Context, userID int) (bool, error) {
	sub, err := c.repo.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, billing.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return sub.IsPro(c.now()), nil
}

// InterimChecker is retained for legacy tests; production uses SubscriptionChecker.
type InterimChecker struct{}

func NewInterimChecker() *InterimChecker {
	return &InterimChecker{}
}

func (c *InterimChecker) IsPro(_ context.Context, _ int) (bool, error) {
	return true, nil
}
