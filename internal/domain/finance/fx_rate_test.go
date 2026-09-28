package finance

import (
	"math"
	"testing"
	"time"
)

func TestFxRateBook_ConvertUSDToIDRViaEUR(t *testing.T) {
	asOf := time.Date(2026, 9, 26, 0, 0, 0, 0, time.UTC)
	book := NewFxRateBook([]FxQuote{
		{Code: "EUR", UnitsPerEUR: 1, AsOf: asOf},
		{Code: "USD", UnitsPerEUR: 1.17, AsOf: asOf},
		{Code: "IDR", UnitsPerEUR: 17800, AsOf: asOf},
	})

	got, ok := book.Convert(10, "USD", "IDR")
	if !ok {
		t.Fatal("expected conversion")
	}
	want := 10 * 17800 / 1.17
	if math.Abs(got-want) > 0.001 {
		t.Fatalf("converted = %v, want %v", got, want)
	}
	if book.AsOf() != "2026-09-26" {
		t.Fatalf("as_of = %s", book.AsOf())
	}
}

func TestFxRateBook_ConvertMissingRate(t *testing.T) {
	book := NewFxRateBook([]FxQuote{
		{Code: "EUR", UnitsPerEUR: 1, AsOf: time.Now()},
		{Code: "IDR", UnitsPerEUR: 17800, AsOf: time.Now()},
	})
	if _, ok := book.Convert(10, "USD", "IDR"); ok {
		t.Fatal("expected missing USD rate")
	}
}
