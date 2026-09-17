package finance

import (
	"testing"
	"time"
)

func intPtr(v int) *int { return &v }

func TestFirstDueOnOrAfter_Weekly(t *testing.T) {
	// Sunday 2024-01-07 → next Monday 2024-01-08
	from := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)
	schedule := RecurringSchedule{Weekday: intPtr(int(time.Monday))}
	got, err := FirstDueOnOrAfter(FrequencyWeekly, schedule, from)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("got %v want %v", got, want)
	}

	// Already Monday → same day
	from = time.Date(2024, 1, 8, 0, 0, 0, 0, time.UTC)
	got, err = FirstDueOnOrAfter(FrequencyWeekly, schedule, from)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(from) {
		t.Fatalf("got %v want %v", got, from)
	}
}

func TestAdvanceFrom_Monthly31Clamp(t *testing.T) {
	schedule := RecurringSchedule{DayOfMonth: intPtr(31)}
	jan31 := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	feb, err := AdvanceFrom(FrequencyMonthly, schedule, jan31)
	if err != nil {
		t.Fatal(err)
	}
	// 2024 is leap year
	wantFeb := time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)
	if !feb.Equal(wantFeb) {
		t.Fatalf("Feb clamp: got %v want %v", feb, wantFeb)
	}

	apr, err := AdvanceFrom(FrequencyMonthly, schedule, time.Date(2024, 3, 31, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	wantApr := time.Date(2024, 4, 30, 0, 0, 0, 0, time.UTC)
	if !apr.Equal(wantApr) {
		t.Fatalf("Apr clamp: got %v want %v", apr, wantApr)
	}
}

func TestAdvanceFrom_YearlyFeb29(t *testing.T) {
	schedule := RecurringSchedule{MonthOfYear: intPtr(2), DayOfMonth: intPtr(29)}
	leap := time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)
	next, err := AdvanceFrom(FrequencyYearly, schedule, leap)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2025, 2, 28, 0, 0, 0, 0, time.UTC)
	if !next.Equal(want) {
		t.Fatalf("got %v want %v", next, want)
	}
}

func TestFirstDueOnOrAfter_Monthly31FromMidFeb(t *testing.T) {
	schedule := RecurringSchedule{DayOfMonth: intPtr(31)}
	from := time.Date(2024, 2, 15, 0, 0, 0, 0, time.UTC)
	got, err := FirstDueOnOrAfter(FrequencyMonthly, schedule, from)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2024, 2, 29, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("got %v want %v", got, want)
	}
}
