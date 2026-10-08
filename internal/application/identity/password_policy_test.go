package identity

import "testing"

func TestPasswordRepeatsEmailIgnoresCase(t *testing.T) {
	if !passwordRepeatsEmail("Ada@Example.com", "ada@example.com") {
		t.Fatal("expected match")
	}
	if passwordRepeatsEmail("long-enough-password", "ada@example.com") {
		t.Fatal("expected distinct password")
	}
}
