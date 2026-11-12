package finance

import "time"

// CalendarPeriodWindow returns the inclusive calendar window for period that contains now.
// Weeks are Monday–Sunday.
func CalendarPeriodWindow(period BudgetPeriod, now time.Time) (time.Time, time.Time) {
	day := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	switch period {
	case BudgetPeriodWeekly:
		weekday := int(day.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start := day.AddDate(0, 0, -(weekday - 1))
		end := start.AddDate(0, 0, 6)
		return start, end
	case BudgetPeriodYearly:
		start := time.Date(day.Year(), time.January, 1, 0, 0, 0, 0, day.Location())
		end := time.Date(day.Year(), time.December, 31, 0, 0, 0, 0, day.Location())
		return start, end
	default:
		start := time.Date(day.Year(), day.Month(), 1, 0, 0, 0, 0, day.Location())
		end := start.AddDate(0, 1, 0).AddDate(0, 0, -1)
		return start, end
	}
}

// IntersectInclusive returns the inclusive overlap of two date ranges.
func IntersectInclusive(aStart, aEnd, bStart, bEnd time.Time) (time.Time, time.Time, bool) {
	start := aStart
	if bStart.After(start) {
		start = bStart
	}
	end := aEnd
	if bEnd.Before(end) {
		end = bEnd
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, false
	}
	return start, end, true
}

// EffectiveBudgetWindow is the date range used for spending reports.
// Prefer intersection of the stored range with the current calendar period.
// If the budget has expired (now after stored end), use the current calendar
// period so recurring budgets keep rolling. If not yet started, keep the stored range.
func EffectiveBudgetWindow(budget *Budget, now time.Time) (time.Time, time.Time) {
	storedStart := dateOnly(budget.StartDate())
	storedEnd := dateOnly(budget.EndDate())
	periodStart, periodEnd := CalendarPeriodWindow(budget.Period(), now)
	today := dateOnly(now)

	if start, end, ok := IntersectInclusive(storedStart, storedEnd, periodStart, periodEnd); ok {
		return start, end
	}
	if today.After(storedEnd) {
		return periodStart, periodEnd
	}
	return storedStart, storedEnd
}

func dateOnly(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}
