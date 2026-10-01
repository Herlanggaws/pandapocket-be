package finance

import (
	"testing"
	"time"
)

func streakDay(value string) time.Time {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		panic(err)
	}
	return parsed
}

func loggedDays(days ...string) map[string]struct{} {
	logged := make(map[string]struct{}, len(days))
	for _, day := range days {
		logged[day] = struct{}{}
	}
	return logged
}

func TestStreakTodayEmptyKeepsYesterday(t *testing.T) {
	result := ComputeActivityStreak(streakComputeInput{
		Today:  streakDay("2026-10-01"),
		Logged: loggedDays("2026-09-30", "2026-09-29"),
		IsPro:  false,
	})
	if result.CurrentStreak != 2 || result.LoggedToday || result.NewGap != nil || result.WouldHaveSaved {
		t.Fatalf("unexpected %+v", result)
	}
}

func TestStreakProFreezeCoversOneGap(t *testing.T) {
	result := ComputeActivityStreak(streakComputeInput{
		Today:  streakDay("2026-10-01"),
		Logged: loggedDays("2026-10-01", "2026-09-29", "2026-09-28"),
		IsPro:  true,
	})
	if result.CurrentStreak != 3 || result.NewGap == nil || result.NewGap.Format("2006-01-02") != "2026-09-30" {
		t.Fatalf("unexpected %+v gap=%v", result, result.NewGap)
	}
	if result.FreezeAvailable || !result.FreezeUsedThisMonth || result.WouldHaveSaved {
		t.Fatalf("freeze flags %+v", result)
	}
}

func TestStreakTwoDayBreakDoesNotFreeze(t *testing.T) {
	result := ComputeActivityStreak(streakComputeInput{
		Today:  streakDay("2026-10-01"),
		Logged: loggedDays("2026-10-01", "2026-09-28"),
		IsPro:  true,
	})
	if result.CurrentStreak != 1 || result.NewGap != nil || result.WouldHaveSaved {
		t.Fatalf("unexpected %+v", result)
	}
	if !result.FreezeAvailable || result.FreezeUsedThisMonth {
		t.Fatalf("freeze should stay unused: %+v", result)
	}
}

func TestStreakFreeDoesNotConsumeFreeze(t *testing.T) {
	result := ComputeActivityStreak(streakComputeInput{
		Today:  streakDay("2026-10-01"),
		Logged: loggedDays("2026-10-01", "2026-09-29"),
		IsPro:  false,
	})
	if result.CurrentStreak != 1 || result.NewGap != nil || !result.WouldHaveSaved {
		t.Fatalf("unexpected %+v", result)
	}
	if result.FreezeAvailable || result.FreezeUsedThisMonth {
		t.Fatalf("free must not hold a freeze: %+v", result)
	}
}
