package finance

import (
	"testing"

	"panda-pocket/internal/infrastructure/analytics"
)

func TestEmitBudgetCreated(t *testing.T) {
	t.Run("captures category_id and period only", func(t *testing.T) {
		var capturedUser int
		var capturedEvent string
		var capturedProps map[string]interface{}
		analytics.SetTestCapture(func(userID int, event string, properties map[string]interface{}) {
			capturedUser = userID
			capturedEvent = event
			capturedProps = properties
		})
		t.Cleanup(func() { analytics.SetTestCapture(nil) })

		emitBudgetCreated(9, 42, "monthly")

		if capturedUser != 9 || capturedEvent != "budget_created" {
			t.Fatalf("got user=%d event=%q", capturedUser, capturedEvent)
		}
		if got, _ := capturedProps["category_id"].(int); got != 42 {
			t.Fatalf("category_id=%v, want 42", capturedProps["category_id"])
		}
		if got, _ := capturedProps["period"].(string); got != "monthly" {
			t.Fatalf("period=%q, want monthly", got)
		}
		for _, banned := range []string{"amount", "name", "description", "merchant"} {
			if _, ok := capturedProps[banned]; ok {
				t.Fatalf("%s must not be in analytics properties", banned)
			}
		}
	})
}
