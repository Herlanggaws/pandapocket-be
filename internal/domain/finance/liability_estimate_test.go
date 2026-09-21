package finance

import "testing"

func TestEstimatedMonthsRemainingZeroAPR(t *testing.T) {
	l := &Liability{currentBalance: 1000, minimumPayment: 100}
	got := l.EstimatedMonthsRemaining()
	if got == nil || *got != 10 {
		t.Fatalf("want 10, got %v", got)
	}
}

func TestEstimatedMonthsRemainingWithAPR(t *testing.T) {
	apr := 12.0
	l := &Liability{currentBalance: 1000, minimumPayment: 100, interestRateAPR: &apr}
	got := l.EstimatedMonthsRemaining()
	if got == nil || *got < 10 {
		t.Fatalf("amortized months should be > zero-APR estimate, got %v", got)
	}
}

func TestEstimatedMonthsRemainingPaymentTooSmall(t *testing.T) {
	apr := 24.0
	l := &Liability{currentBalance: 10000, minimumPayment: 10, interestRateAPR: &apr}
	if l.EstimatedMonthsRemaining() != nil {
		t.Fatal("expected nil when payment does not cover interest")
	}
}
