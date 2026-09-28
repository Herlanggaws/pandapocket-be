package finance

import (
	"context"
	"errors"
	"testing"
	"time"

	domainFinance "panda-pocket/internal/domain/finance"
)

type memoryFxRateRepo struct {
	quotes  []domainFinance.FxQuote
	fetched *time.Time
	upserts int
}

func (r *memoryFxRateRepo) LatestFetchedAt(ctx context.Context) (*time.Time, error) {
	return r.fetched, nil
}

func (r *memoryFxRateRepo) Upsert(ctx context.Context, quotes []domainFinance.FxQuote) error {
	r.upserts++
	r.quotes = append([]domainFinance.FxQuote(nil), quotes...)
	now := time.Now().UTC()
	r.fetched = &now
	return nil
}

func (r *memoryFxRateRepo) List(ctx context.Context) ([]domainFinance.FxQuote, error) {
	return append([]domainFinance.FxQuote(nil), r.quotes...), nil
}

type stubFxSource struct {
	quotes []domainFinance.FxQuote
	err    error
	calls  int
}

func (s *stubFxSource) LatestEUR(ctx context.Context) ([]domainFinance.FxQuote, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return s.quotes, nil
}

func TestRefreshFxRates_FetchFailureKeepsStoredRates(t *testing.T) {
	asOf := time.Date(2026, 9, 25, 0, 0, 0, 0, time.UTC)
	fetched := time.Now().UTC().Add(-48 * time.Hour)
	repo := &memoryFxRateRepo{
		quotes:  []domainFinance.FxQuote{{Code: "USD", UnitsPerEUR: 1.1, AsOf: asOf}},
		fetched: &fetched,
	}
	source := &stubFxSource{err: errors.New("down")}
	uc := NewRefreshFxRatesUseCase(repo, source)
	uc.now = func() time.Time { return time.Now().UTC() }

	uc.Execute(context.Background())

	if source.calls != 1 {
		t.Fatalf("source calls = %d, want 1", source.calls)
	}
	if repo.upserts != 0 {
		t.Fatalf("upserts = %d, want 0", repo.upserts)
	}
	if len(repo.quotes) != 1 || repo.quotes[0].UnitsPerEUR != 1.1 {
		t.Fatalf("stored rates changed: %#v", repo.quotes)
	}
}

func TestRefreshFxRates_SkipsFreshCache(t *testing.T) {
	fetched := time.Now().UTC().Add(-time.Hour)
	repo := &memoryFxRateRepo{fetched: &fetched}
	source := &stubFxSource{}
	uc := NewRefreshFxRatesUseCase(repo, source)

	uc.Execute(context.Background())

	if source.calls != 0 {
		t.Fatalf("source calls = %d, want 0", source.calls)
	}
}
