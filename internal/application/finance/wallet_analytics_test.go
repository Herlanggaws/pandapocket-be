package finance

import (
	"testing"
	"time"

	domainFinance "panda-pocket/internal/domain/finance"
	"panda-pocket/internal/infrastructure/analytics"
)

func TestEmitWalletCreated(t *testing.T) {
	wallet := domainFinance.ReconstituteWallet(
		domainFinance.NewWalletID(1),
		domainFinance.NewUserID(7),
		"Cash",
		domainFinance.WalletTypeCash,
		domainFinance.NewCurrencyID(1),
		0,
		true,
		false,
		time.Now(),
	)

	t.Run("captures when created", func(t *testing.T) {
		var capturedEvent string
		var capturedUser int
		var capturedProps map[string]interface{}
		analytics.SetTestCapture(func(userID int, event string, properties map[string]interface{}) {
			capturedUser = userID
			capturedEvent = event
			capturedProps = properties
		})
		t.Cleanup(func() { analytics.SetTestCapture(nil) })

		emitWalletCreated(7, wallet, true, "ensure_default")

		if capturedUser != 7 || capturedEvent != "wallet_created" {
			t.Fatalf("got user=%d event=%q", capturedUser, capturedEvent)
		}
		if got, _ := capturedProps["source"].(string); got != "ensure_default" {
			t.Fatalf("source=%q, want ensure_default", got)
		}
		if got, _ := capturedProps["wallet_type"].(string); got != "cash" {
			t.Fatalf("wallet_type=%q, want cash", got)
		}
		if _, ok := capturedProps["amount"]; ok {
			t.Fatal("amount must not be in analytics properties")
		}
		if _, ok := capturedProps["name"]; ok {
			t.Fatal("name must not be in analytics properties")
		}
	})

	t.Run("uses onboarding_default source", func(t *testing.T) {
		var capturedProps map[string]interface{}
		analytics.SetTestCapture(func(_ int, _ string, properties map[string]interface{}) {
			capturedProps = properties
		})
		t.Cleanup(func() { analytics.SetTestCapture(nil) })

		emitWalletCreated(7, wallet, true, "onboarding_default")
		if got, _ := capturedProps["source"].(string); got != "onboarding_default" {
			t.Fatalf("source=%q, want onboarding_default", got)
		}
	})

	t.Run("uses lazy_default source", func(t *testing.T) {
		var capturedProps map[string]interface{}
		analytics.SetTestCapture(func(_ int, _ string, properties map[string]interface{}) {
			capturedProps = properties
		})
		t.Cleanup(func() { analytics.SetTestCapture(nil) })

		emitWalletCreated(7, wallet, true, "lazy_default")
		if got, _ := capturedProps["source"].(string); got != "lazy_default" {
			t.Fatalf("source=%q, want lazy_default", got)
		}
	})

	t.Run("skips when already exists", func(t *testing.T) {
		called := false
		analytics.SetTestCapture(func(userID int, event string, properties map[string]interface{}) {
			called = true
		})
		t.Cleanup(func() { analytics.SetTestCapture(nil) })

		emitWalletCreated(7, wallet, false, "ensure_default")
		if called {
			t.Fatal("expected no capture when wallet already existed")
		}
	})

	t.Run("skips nil wallet", func(t *testing.T) {
		called := false
		analytics.SetTestCapture(func(userID int, event string, properties map[string]interface{}) {
			called = true
		})
		t.Cleanup(func() { analytics.SetTestCapture(nil) })

		emitWalletCreated(7, nil, true, "lazy_default")
		if called {
			t.Fatal("expected no capture for nil wallet")
		}
	})
}
