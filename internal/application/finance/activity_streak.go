package finance

import (
	"context"
	"time"

	"panda-pocket/internal/domain/entitlement"
)

// Activity days use Asia/Jakarta so a backfill saved in one evening stays one day.
// created_at is the logging habit; the transaction date field can be backdated.
var activityClockLocation = loadJakarta()

func loadJakarta() *time.Location {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return time.FixedZone("WIB", 7*60*60)
	}
	return loc
}

func ActivityCalendarDate(now time.Time) time.Time {
	local := now.In(activityClockLocation)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
}

type streakComputeInput struct {
	Today       time.Time
	Logged      map[string]struct{}
	IsPro       bool
	ExistingGap *time.Time
}

type streakComputeResult struct {
	CurrentStreak       int
	LoggedToday         bool
	FreezeAvailable     bool
	FreezeUsedThisMonth bool
	WouldHaveSaved      bool
	NewGap              *time.Time
}

func ComputeActivityStreak(in streakComputeInput) streakComputeResult {
	actual, newGap := walkStreak(in.Today, in.Logged, in.IsPro, in.ExistingGap)
	hypothetical, _ := walkStreak(in.Today, in.Logged, true, nil)
	used := in.ExistingGap != nil || newGap != nil
	return streakComputeResult{
		CurrentStreak:       actual,
		LoggedToday:         hasActivity(in.Logged, in.Today),
		FreezeAvailable:     in.IsPro && !used,
		FreezeUsedThisMonth: used,
		WouldHaveSaved:      !in.IsPro && hypothetical > actual,
		NewGap:              newGap,
	}
}

func walkStreak(
	today time.Time,
	logged map[string]struct{},
	canFreeze bool,
	existingGap *time.Time,
) (int, *time.Time) {
	cursor := today
	if !hasActivity(logged, today) {
		cursor = today.AddDate(0, 0, -1)
	}

	consumed := existingGap != nil
	var newGap *time.Time
	streak := 0
	for guard := 0; guard < 4000; guard++ {
		if hasActivity(logged, cursor) {
			streak++
			cursor = cursor.AddDate(0, 0, -1)
			continue
		}
		if !coverGap(cursor, logged, canFreeze, existingGap, &consumed, &newGap) {
			break
		}
		cursor = cursor.AddDate(0, 0, -1)
	}
	return streak, newGap
}

func coverGap(
	cursor time.Time,
	logged map[string]struct{},
	canFreeze bool,
	existingGap *time.Time,
	consumed *bool,
	newGap **time.Time,
) bool {
	if existingGap != nil && existingGap.Format("2006-01-02") == cursor.Format("2006-01-02") {
		return true
	}
	if !canFreeze || *consumed || !hasActivity(logged, cursor.AddDate(0, 0, -1)) {
		return false
	}
	*consumed = true
	gap := cursor
	*newGap = &gap
	return true
}

func hasActivity(logged map[string]struct{}, day time.Time) bool {
	_, ok := logged[day.Format("2006-01-02")]
	return ok
}

type ActivityDayReader interface {
	ListActivityDates(ctx context.Context, userID int) ([]time.Time, error)
}

type StreakFreezeStore interface {
	FindGap(ctx context.Context, userID int, yearMonth string) (*time.Time, error)
	SaveGap(ctx context.Context, userID int, yearMonth string, gapDate, consumedAt time.Time) (bool, error)
}

type ActivityStreakResponse struct {
	CurrentStreak       int  `json:"current_streak"`
	LoggedToday         bool `json:"logged_today"`
	FreezeAvailable     bool `json:"freeze_available"`
	FreezeUsedThisMonth bool `json:"freeze_used_this_month"`
	WouldHaveSaved      bool `json:"would_have_saved"`
}

type GetActivityStreakUseCase struct {
	days         ActivityDayReader
	freezes      StreakFreezeStore
	entitlements entitlement.Checker
	now          func() time.Time
}

func NewGetActivityStreakUseCase(
	days ActivityDayReader,
	freezes StreakFreezeStore,
	entitlements entitlement.Checker,
) *GetActivityStreakUseCase {
	return &GetActivityStreakUseCase{
		days:         days,
		freezes:      freezes,
		entitlements: entitlements,
		now:          time.Now,
	}
}

func (uc *GetActivityStreakUseCase) Execute(ctx context.Context, userID int) (*ActivityStreakResponse, error) {
	today := ActivityCalendarDate(uc.now())
	dates, err := uc.days.ListActivityDates(ctx, userID)
	if err != nil {
		return nil, err
	}
	isPro, err := uc.entitlements.IsPro(ctx, userID)
	if err != nil {
		return nil, err
	}
	yearMonth := today.Format("2006-01")
	existing, err := uc.freezes.FindGap(ctx, userID, yearMonth)
	if err != nil {
		return nil, err
	}
	result := ComputeActivityStreak(streakComputeInput{
		Today:       today,
		Logged:      activityDateSet(dates),
		IsPro:       isPro,
		ExistingGap: existing,
	})
	if result.NewGap != nil {
		saved, saveErr := uc.freezes.SaveGap(ctx, userID, yearMonth, *result.NewGap, uc.now())
		if saveErr != nil {
			return nil, saveErr
		}
		if !saved {
			existing, err = uc.freezes.FindGap(ctx, userID, yearMonth)
			if err != nil {
				return nil, err
			}
			result = ComputeActivityStreak(streakComputeInput{
				Today:       today,
				Logged:      activityDateSet(dates),
				IsPro:       isPro,
				ExistingGap: existing,
			})
		}
	}
	return &ActivityStreakResponse{
		CurrentStreak:       result.CurrentStreak,
		LoggedToday:         result.LoggedToday,
		FreezeAvailable:     result.FreezeAvailable,
		FreezeUsedThisMonth: result.FreezeUsedThisMonth,
		WouldHaveSaved:      result.WouldHaveSaved,
	}, nil
}

func activityDateSet(dates []time.Time) map[string]struct{} {
	logged := make(map[string]struct{}, len(dates))
	for _, date := range dates {
		year, month, day := date.Date()
		key := time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
		logged[key] = struct{}{}
	}
	return logged
}
