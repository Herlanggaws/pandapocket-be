package finance

import (
	"context"
	"log"
	"time"

	domainFinance "panda-pocket/internal/domain/finance"
)

const fxRateMaxAge = 24 * time.Hour

// FxRateSource loads the latest EUR-pivot quotes.
type FxRateSource interface {
	LatestEUR(ctx context.Context) ([]domainFinance.FxQuote, error)
}

type RefreshFxRatesUseCase struct {
	repo   domainFinance.FxRateRepository
	source FxRateSource
	now    func() time.Time
}

func NewRefreshFxRatesUseCase(repo domainFinance.FxRateRepository, source FxRateSource) *RefreshFxRatesUseCase {
	return &RefreshFxRatesUseCase{
		repo:   repo,
		source: source,
		now:    time.Now,
	}
}

func (uc *RefreshFxRatesUseCase) Execute(ctx context.Context) {
	if uc.repo == nil || uc.source == nil {
		return
	}
	latest, err := uc.repo.LatestFetchedAt(ctx)
	if err != nil {
		log.Printf("fx rate refresh failed to read cache: %v", err)
		return
	}
	if latest != nil && uc.now().Sub(*latest) < fxRateMaxAge {
		return
	}
	quotes, err := uc.source.LatestEUR(ctx)
	if err != nil {
		log.Printf("fx rate refresh failed, keeping last rates: %v", err)
		return
	}
	if err := uc.repo.Upsert(ctx, quotes); err != nil {
		log.Printf("fx rate refresh failed to store rates: %v", err)
	}
}
