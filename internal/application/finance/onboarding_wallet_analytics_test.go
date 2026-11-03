package finance

import (
	"testing"
	"time"

	domainFinance "panda-pocket/internal/domain/finance"
	"panda-pocket/internal/infrastructure/analytics"
)

func TestEmitOnboardingDefaultWalletCreated(t *testing.T) {
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

		emitOnboardingDefaultWalletCreated(7, wallet, true)

		if capturedUser != 7 || capturedEvent != "wallet_created" {
			t.Fatalf("got user=%d event=%q", capturedUser, capturedEvent)
		}
		if got, _ := capturedProps["source"].(string); got != "onboarding_default" {
			t.Fatalf("source=%q, want onboarding_default", got)
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

	t.Run("skips when already exists", func(t *testing.T) {
		called := false
		analytics.SetTestCapture(func(userID int, event string, properties map[string]interface{}) {
			called = true
		})
		t.Cleanup(func() { analytics.SetTestCapture(nil) })

		emitOnboardingDefaultWalletCreated(7, wallet, false)
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

		emitOnboardingDefaultWalletCreated(7, nil, true)
		if called {
			t.Fatal("expected no capture for nil wallet")
		}
	})
}
