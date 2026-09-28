package finance

import (
	"context"
	"strings"
	"time"
)

// FxQuote is one currency's value in the EUR pivot (units of that currency per 1 EUR).
type FxQuote struct {
	Code        string
	UnitsPerEUR float64
	AsOf        time.Time
}

// FxRateRepository stores the cached daily ECB rate book.
type FxRateRepository interface {
	LatestFetchedAt(ctx context.Context) (*time.Time, error)
	Upsert(ctx context.Context, quotes []FxQuote) error
	List(ctx context.Context) ([]FxQuote, error)
}

// FxRateBook converts amounts through the EUR pivot.
type FxRateBook struct {
	units map[string]float64
	asOf  string
}

func NewFxRateBook(quotes []FxQuote) FxRateBook {
	units := make(map[string]float64, len(quotes))
	var latest time.Time
	hasDate := false
	for _, quote := range quotes {
		code := strings.ToUpper(strings.TrimSpace(quote.Code))
		if code == "" || quote.UnitsPerEUR <= 0 {
			continue
		}
		units[code] = quote.UnitsPerEUR
		if !hasDate || quote.AsOf.After(latest) {
			latest = quote.AsOf
			hasDate = true
		}
	}
	asOf := ""
	if hasDate && !latest.IsZero() {
		asOf = latest.UTC().Format("2006-01-02")
	}
	return FxRateBook{units: units, asOf: asOf}
}

func (b FxRateBook) AsOf() string {
	return b.asOf
}

// Convert returns amount expressed in primaryCode. ok is false when a rate is missing.
func (b FxRateBook) Convert(amount float64, sourceCode, primaryCode string) (float64, bool) {
	source := b.units[strings.ToUpper(sourceCode)]
	primary := b.units[strings.ToUpper(primaryCode)]
	if source <= 0 || primary <= 0 {
		return 0, false
	}
	return amount * primary / source, true
}
