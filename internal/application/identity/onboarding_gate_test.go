package identity

import (
	"context"
	"encoding/json"
	"testing"
)

type stubWalletCounter struct {
	count int64
	err   error
}

func (s stubWalletCounter) CountActiveByUserID(context.Context, int) (int64, error) {
	return s.count, s.err
}

func TestApplyOnboardingGate(t *testing.T) {
	cases := []struct {
		name       string
		completed  interface{}
		walletN    int64
		wantNeeds  bool
		wantWallet bool
	}{
		{name: "new user", completed: false, walletN: 0, wantNeeds: true, wantWallet: false},
		{name: "completed without wallet", completed: true, walletN: 0, wantNeeds: true, wantWallet: false},
		{name: "wallet without complete", completed: false, walletN: 1, wantNeeds: true, wantWallet: true},
		{name: "done", completed: true, walletN: 1, wantNeeds: false, wantWallet: true},
		{name: "nil completed", completed: nil, walletN: 0, wantNeeds: true, wantWallet: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := &PreferencesResponse{
				OnboardingCompleted: tc.completed,
				Onboarding:          json.RawMessage(`{}`),
			}
			applyOnboardingGate(context.Background(), stubWalletCounter{count: tc.walletN}, 7, resp)
			if resp.NeedsOnboarding != tc.wantNeeds {
				t.Fatalf("needs_onboarding=%v want %v", resp.NeedsOnboarding, tc.wantNeeds)
			}
			if resp.HasWallet != tc.wantWallet {
				t.Fatalf("has_wallet=%v want %v", resp.HasWallet, tc.wantWallet)
			}
		})
	}
}
