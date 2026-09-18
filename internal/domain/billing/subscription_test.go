package billing

import (
	"testing"
	"time"
)

func TestSubscriptionIsPro(t *testing.T) {
	now := time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC)
	future := now.Add(24 * time.Hour)
	past := now.Add(-24 * time.Hour)

	t.Run("free expired is not pro", func(t *testing.T) {
		sub, err := NewFreeSubscription(1)
		if err != nil {
			t.Fatal(err)
		}
		if sub.IsPro(now) {
			t.Fatal("expected free subscription not to be Pro")
		}
	})

	t.Run("active with future period end is pro", func(t *testing.T) {
		sub := ReconstituteSubscription(
			NewSubscriptionID(1), 1, PlanPro, nil, StatusActive,
			nil, &future, nil, nil, CustomerRef(1), false, now, now,
		)
		if !sub.IsPro(now) {
			t.Fatal("expected active subscription to be Pro")
		}
	})

	t.Run("active with expired period is not pro", func(t *testing.T) {
		sub := ReconstituteSubscription(
			NewSubscriptionID(1), 1, PlanPro, nil, StatusActive,
			nil, &past, nil, nil, CustomerRef(1), false, now, now,
		)
		if sub.IsPro(now) {
			t.Fatal("expected expired period not to be Pro")
		}
	})

	t.Run("past_due within grace is pro", func(t *testing.T) {
		sub := ReconstituteSubscription(
			NewSubscriptionID(1), 1, PlanPro, nil, StatusPastDue,
			nil, &past, &future, nil, CustomerRef(1), false, now, now,
		)
		if !sub.IsPro(now) {
			t.Fatal("expected past_due within grace to be Pro")
		}
	})

	t.Run("past_due after grace is not pro", func(t *testing.T) {
		sub := ReconstituteSubscription(
			NewSubscriptionID(1), 1, PlanPro, nil, StatusPastDue,
			nil, &past, &past, nil, CustomerRef(1), false, now, now,
		)
		if sub.IsPro(now) {
			t.Fatal("expected past_due after grace not to be Pro")
		}
	})

	t.Run("trial active is pro regardless of status", func(t *testing.T) {
		sub := ReconstituteSubscription(
			NewSubscriptionID(1), 1, PlanFree, nil, StatusTrialing,
			&future, nil, nil, nil, CustomerRef(1), false, now, now,
		)
		if !sub.IsPro(now) {
			t.Fatal("expected active trial to be Pro")
		}
	})

	t.Run("expired trial is not pro", func(t *testing.T) {
		sub := ReconstituteSubscription(
			NewSubscriptionID(1), 1, PlanFree, nil, StatusExpired,
			&past, nil, nil, nil, CustomerRef(1), false, now, now,
		)
		if sub.IsPro(now) {
			t.Fatal("expected expired trial not to be Pro")
		}
	})
}

func TestNewTrialSubscription(t *testing.T) {
	before := time.Now().UTC()
	sub, err := NewTrialSubscription(5)
	if err != nil {
		t.Fatal(err)
	}
	after := time.Now().UTC()

	if sub.Plan() != PlanFree || sub.Status() != StatusTrialing {
		t.Fatalf("expected free/trialing, got %s/%s", sub.Plan(), sub.Status())
	}
	if !sub.HasUsedTrial() || sub.TrialEndsAt() == nil {
		t.Fatal("expected trial_ends_at set")
	}
	minEnd := before.AddDate(0, 0, TrialDurationDays)
	maxEnd := after.AddDate(0, 0, TrialDurationDays)
	if sub.TrialEndsAt().Before(minEnd.Add(-time.Second)) || sub.TrialEndsAt().After(maxEnd.Add(time.Second)) {
		t.Fatalf("trial end out of range: %v", sub.TrialEndsAt())
	}
	if !sub.IsPro(time.Now().UTC()) {
		t.Fatal("active trial should be Pro")
	}
	if sub.DoitCustomerRef() != CustomerRef(5) {
		t.Fatalf("unexpected customer ref %s", sub.DoitCustomerRef())
	}
}

func TestExpireTrialIfNeeded(t *testing.T) {
	now := time.Date(2026, 9, 18, 12, 0, 0, 0, time.UTC)
	past := now.Add(-time.Hour)
	future := now.Add(24 * time.Hour)

	t.Run("expires ended trial", func(t *testing.T) {
		sub := ReconstituteSubscription(
			NewSubscriptionID(1), 1, PlanFree, nil, StatusTrialing,
			&past, nil, nil, nil, CustomerRef(1), false, now, now,
		)
		if !sub.ExpireTrialIfNeeded(now) {
			t.Fatal("expected expire")
		}
		if sub.Status() != StatusExpired {
			t.Fatalf("expected expired, got %s", sub.Status())
		}
		if sub.TrialEndsAt() == nil || !sub.HasUsedTrial() {
			t.Fatal("trial_ends_at must remain")
		}
		if sub.IsPro(now) {
			t.Fatal("expired trial must not be Pro")
		}
	})

	t.Run("does not expire active trial", func(t *testing.T) {
		sub := ReconstituteSubscription(
			NewSubscriptionID(1), 1, PlanFree, nil, StatusTrialing,
			&future, nil, nil, nil, CustomerRef(1), false, now, now,
		)
		if sub.ExpireTrialIfNeeded(now) {
			t.Fatal("should not expire")
		}
		if sub.Status() != StatusTrialing {
			t.Fatal("status should stay trialing")
		}
	})

	t.Run("does not clear already expired", func(t *testing.T) {
		sub := ReconstituteSubscription(
			NewSubscriptionID(1), 1, PlanFree, nil, StatusExpired,
			&past, nil, nil, nil, CustomerRef(1), false, now, now,
		)
		if sub.ExpireTrialIfNeeded(now) {
			t.Fatal("already expired should be no-op")
		}
	})
}
