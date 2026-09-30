package finance

import "testing"

func TestResolveOnboardingGoals(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		goals      []string
		legacyGoal string
		want       []string
		wantErr    bool
	}{
		{
			name:  "multiple goals",
			goals: []string{"budget", "debt"},
			want:  []string{"budget", "debt"},
		},
		{
			name:       "legacy single goal when list is empty",
			legacyGoal: "save",
			want:       []string{"save"},
		},
		{
			name:       "list wins over legacy",
			goals:      []string{"track"},
			legacyGoal: "debt",
			want:       []string{"track"},
		},
		{
			name:  "dedupes",
			goals: []string{"save", "save", "track"},
			want:  []string{"save", "track"},
		},
		{
			name:    "empty",
			wantErr: true,
		},
		{
			name:    "unknown id",
			goals:   []string{"invest"},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got, err := resolveOnboardingGoals(test.goals, test.legacyGoal)
			if test.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(got) != len(test.want) {
				t.Fatalf("got %v, want %v", got, test.want)
			}
			for i := range test.want {
				if got[i] != test.want[i] {
					t.Fatalf("got %v, want %v", got, test.want)
				}
			}
		})
	}
}

func TestOnboardingGoalsIncludeDebt(t *testing.T) {
	t.Parallel()
	if !onboardingGoalsIncludeDebt([]string{"save", "debt"}) {
		t.Fatal("expected debt")
	}
	if onboardingGoalsIncludeDebt([]string{"save", "budget"}) {
		t.Fatal("did not expect debt")
	}
}
