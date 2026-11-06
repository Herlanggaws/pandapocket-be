package identity

import "testing"

func TestIsInternalEmailPlusTest(t *testing.T) {
	cases := []struct {
		email string
		want  bool
	}{
		{"user+test@example.com", true},
		{"user+testfoo@example.com", true},
		{"+test@example.com", true},
		{"User+TEST@Example.com", true},
		{"user+foo@example.com", false},
		{"test@example.com", false},
		{"usertest@example.com", false},
		{"", false},
		{"not-an-email", false},
		{"@example.com", false},
	}
	for _, tc := range cases {
		if got := IsInternalEmail(tc.email, nil); got != tc.want {
			t.Fatalf("IsInternalEmail(%q, nil)=%v want %v", tc.email, got, tc.want)
		}
	}
}

func TestIsInternalEmailDomains(t *testing.T) {
	domains := []string{"xerv.io", "bearbudget.internal"}
	if !IsInternalEmail("dev@xerv.io", domains) {
		t.Fatal("expected domain match")
	}
	if !IsInternalEmail("Dev@Xerv.IO", domains) {
		t.Fatal("expected case-insensitive domain match")
	}
	if IsInternalEmail("dev@other.com", domains) {
		t.Fatal("unexpected domain match")
	}
	if !IsInternalEmail("dev+test@other.com", domains) {
		t.Fatal("+test should match even when domain is not listed")
	}
}

func TestParseInternalEmailDomains(t *testing.T) {
	if got := ParseInternalEmailDomains(""); got != nil {
		t.Fatalf("empty raw want nil, got %#v", got)
	}
	got := ParseInternalEmailDomains(" xerv.io , , BearBudget.Internal ")
	if len(got) != 2 || got[0] != "xerv.io" || got[1] != "bearbudget.internal" {
		t.Fatalf("unexpected parse: %#v", got)
	}
}
