package finance

import (
	"errors"
	"fmt"
	"time"
)

// Frequency represents the frequency of a recurring transaction
type Frequency string

const (
	FrequencyDaily   Frequency = "daily"
	FrequencyWeekly  Frequency = "weekly"
	FrequencyMonthly Frequency = "monthly"
	FrequencyYearly  Frequency = "yearly"
)

// RecurringSchedule holds frequency-specific schedule fields
type RecurringSchedule struct {
	Weekday      *int // 0=Sunday … 6=Saturday (time.Weekday)
	DayOfMonth   *int // 1–31
	MonthOfYear  *int // 1–12
}

// RecurringTransaction represents a recurring transaction
type RecurringTransaction struct {
	id              RecurringTransactionID
	userID          UserID
	walletID        WalletID
	categoryID      CategoryID
	currencyID      CurrencyID
	amount          Money
	description     string
	frequency       Frequency
	transactionType TransactionType
	schedule        RecurringSchedule
	nextDueDate     time.Time
	isActive        bool
	createdAt       time.Time
}

// RecurringTransactionID is a value object representing a recurring transaction identifier
type RecurringTransactionID struct {
	value int
}

func NewRecurringTransactionID(id int) RecurringTransactionID {
	return RecurringTransactionID{value: id}
}

func (r RecurringTransactionID) Value() int {
	return r.value
}

func truncateDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func clampDayOfMonth(year int, month time.Month, day int) int {
	maxDay := daysInMonth(year, month)
	if day > maxDay {
		return maxDay
	}
	if day < 1 {
		return 1
	}
	return day
}

func dateInMonth(year int, month time.Month, day int, loc *time.Location) time.Time {
	return time.Date(year, month, clampDayOfMonth(year, month, day), 0, 0, 0, 0, loc)
}

// ValidateSchedule validates schedule fields for a frequency
func ValidateSchedule(frequency Frequency, schedule RecurringSchedule) error {
	switch frequency {
	case FrequencyDaily:
		return nil
	case FrequencyWeekly:
		if schedule.Weekday == nil {
			return errors.New("weekday is required for weekly frequency")
		}
		if *schedule.Weekday < 0 || *schedule.Weekday > 6 {
			return errors.New("weekday must be between 0 (Sunday) and 6 (Saturday)")
		}
		return nil
	case FrequencyMonthly:
		if schedule.DayOfMonth == nil {
			return errors.New("day_of_month is required for monthly frequency")
		}
		if *schedule.DayOfMonth < 1 || *schedule.DayOfMonth > 31 {
			return errors.New("day_of_month must be between 1 and 31")
		}
		return nil
	case FrequencyYearly:
		if schedule.MonthOfYear == nil {
			return errors.New("month_of_year is required for yearly frequency")
		}
		if schedule.DayOfMonth == nil {
			return errors.New("day_of_month is required for yearly frequency")
		}
		if *schedule.MonthOfYear < 1 || *schedule.MonthOfYear > 12 {
			return errors.New("month_of_year must be between 1 and 12")
		}
		if *schedule.DayOfMonth < 1 || *schedule.DayOfMonth > 31 {
			return errors.New("day_of_month must be between 1 and 31")
		}
		month := time.Month(*schedule.MonthOfYear)
		// Allow Feb 29; other impossible days (e.g. Apr 31) are rejected
		if *schedule.DayOfMonth > 29 && month == time.February {
			return errors.New("invalid day for February")
		}
		if *schedule.DayOfMonth == 31 && (month == time.April || month == time.June || month == time.September || month == time.November) {
			return errors.New("invalid day for selected month")
		}
		if *schedule.DayOfMonth == 30 && month == time.February {
			return errors.New("invalid day for February")
		}
		return nil
	default:
		return errors.New("invalid frequency")
	}
}

// NormalizeSchedule clears unused fields and returns a clean schedule for the frequency
func NormalizeSchedule(frequency Frequency, schedule RecurringSchedule) RecurringSchedule {
	switch frequency {
	case FrequencyWeekly:
		return RecurringSchedule{Weekday: schedule.Weekday}
	case FrequencyMonthly:
		return RecurringSchedule{DayOfMonth: schedule.DayOfMonth}
	case FrequencyYearly:
		return RecurringSchedule{DayOfMonth: schedule.DayOfMonth, MonthOfYear: schedule.MonthOfYear}
	default:
		return RecurringSchedule{}
	}
}

// FirstDueOnOrAfter returns the first occurrence of the schedule on or after from (date-only)
func FirstDueOnOrAfter(frequency Frequency, schedule RecurringSchedule, from time.Time) (time.Time, error) {
	if err := ValidateSchedule(frequency, schedule); err != nil {
		return time.Time{}, err
	}
	from = truncateDay(from)
	loc := from.Location()

	switch frequency {
	case FrequencyDaily:
		return from, nil
	case FrequencyWeekly:
		target := time.Weekday(*schedule.Weekday)
		delta := int(target - from.Weekday())
		if delta < 0 {
			delta += 7
		}
		return from.AddDate(0, 0, delta), nil
	case FrequencyMonthly:
		day := *schedule.DayOfMonth
		candidate := dateInMonth(from.Year(), from.Month(), day, loc)
		if candidate.Before(from) {
			firstOfMonth := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, loc)
			next := firstOfMonth.AddDate(0, 1, 0)
			candidate = dateInMonth(next.Year(), next.Month(), day, loc)
		}
		return candidate, nil
	case FrequencyYearly:
		month := time.Month(*schedule.MonthOfYear)
		day := *schedule.DayOfMonth
		candidate := dateInMonth(from.Year(), month, day, loc)
		if candidate.Before(from) {
			candidate = dateInMonth(from.Year()+1, month, day, loc)
		}
		return candidate, nil
	default:
		return time.Time{}, errors.New("invalid frequency")
	}
}

// AdvanceFrom returns the next due date strictly after the given due date
func AdvanceFrom(frequency Frequency, schedule RecurringSchedule, dueDate time.Time) (time.Time, error) {
	if err := ValidateSchedule(frequency, schedule); err != nil {
		return time.Time{}, err
	}
	dueDate = truncateDay(dueDate)
	loc := dueDate.Location()

	switch frequency {
	case FrequencyDaily:
		return dueDate.AddDate(0, 0, 1), nil
	case FrequencyWeekly:
		return dueDate.AddDate(0, 0, 7), nil
	case FrequencyMonthly:
		day := *schedule.DayOfMonth
		firstOfMonth := time.Date(dueDate.Year(), dueDate.Month(), 1, 0, 0, 0, 0, loc)
		nextMonth := firstOfMonth.AddDate(0, 1, 0)
		return dateInMonth(nextMonth.Year(), nextMonth.Month(), day, loc), nil
	case FrequencyYearly:
		month := time.Month(*schedule.MonthOfYear)
		day := *schedule.DayOfMonth
		return dateInMonth(dueDate.Year()+1, month, day, loc), nil
	default:
		return time.Time{}, errors.New("invalid frequency")
	}
}

// DeriveScheduleFromDueDate fills missing schedule fields from next_due_date (legacy rows)
func DeriveScheduleFromDueDate(frequency Frequency, schedule RecurringSchedule, dueDate time.Time) RecurringSchedule {
	dueDate = truncateDay(dueDate)
	switch frequency {
	case FrequencyWeekly:
		if schedule.Weekday == nil {
			w := int(dueDate.Weekday())
			schedule.Weekday = &w
		}
	case FrequencyMonthly:
		if schedule.DayOfMonth == nil {
			d := dueDate.Day()
			schedule.DayOfMonth = &d
		}
	case FrequencyYearly:
		if schedule.MonthOfYear == nil {
			m := int(dueDate.Month())
			schedule.MonthOfYear = &m
		}
		if schedule.DayOfMonth == nil {
			d := dueDate.Day()
			schedule.DayOfMonth = &d
		}
	}
	return NormalizeSchedule(frequency, schedule)
}

// NewRecurringTransaction creates a new recurring transaction with schedule
func NewRecurringTransaction(
	userID UserID,
	walletID WalletID,
	categoryID CategoryID,
	currencyID CurrencyID,
	amount Money,
	description string,
	frequency Frequency,
	transactionType TransactionType,
	schedule RecurringSchedule,
	from time.Time,
) (*RecurringTransaction, error) {
	if amount.Amount() <= 0 {
		return nil, errors.New("recurring transaction amount must be positive")
	}

	switch transactionType {
	case TransactionTypeExpense, TransactionTypeIncome:
	default:
		return nil, errors.New("invalid transaction type")
	}

	schedule = NormalizeSchedule(frequency, schedule)
	if err := ValidateSchedule(frequency, schedule); err != nil {
		return nil, err
	}

	nextDue, err := FirstDueOnOrAfter(frequency, schedule, from)
	if err != nil {
		return nil, err
	}

	return &RecurringTransaction{
		userID:          userID,
		walletID:        walletID,
		categoryID:      categoryID,
		currencyID:      currencyID,
		amount:          amount,
		description:     description,
		frequency:       frequency,
		transactionType: transactionType,
		schedule:        schedule,
		nextDueDate:     nextDue,
		isActive:        true,
		createdAt:       time.Now(),
	}, nil
}

func ReconstituteRecurringTransaction(
	id RecurringTransactionID,
	userID UserID,
	walletID WalletID,
	categoryID CategoryID,
	currencyID CurrencyID,
	amount Money,
	description string,
	frequency Frequency,
	transactionType TransactionType,
	schedule RecurringSchedule,
	nextDueDate time.Time,
	isActive bool,
	createdAt time.Time,
) *RecurringTransaction {
	schedule = DeriveScheduleFromDueDate(frequency, schedule, nextDueDate)
	return &RecurringTransaction{
		id:              id,
		userID:          userID,
		walletID:        walletID,
		categoryID:      categoryID,
		currencyID:      currencyID,
		amount:          amount,
		description:     description,
		frequency:       frequency,
		transactionType: transactionType,
		schedule:        schedule,
		nextDueDate:     nextDueDate,
		isActive:        isActive,
		createdAt:       createdAt,
	}
}

func (r *RecurringTransaction) AssignID(id RecurringTransactionID) {
	r.id = id
}

func (r *RecurringTransaction) ID() RecurringTransactionID { return r.id }
func (r *RecurringTransaction) UserID() UserID             { return r.userID }
func (r *RecurringTransaction) WalletID() WalletID         { return r.walletID }
func (r *RecurringTransaction) CategoryID() CategoryID     { return r.categoryID }
func (r *RecurringTransaction) CurrencyID() CurrencyID     { return r.currencyID }
func (r *RecurringTransaction) Amount() Money              { return r.amount }
func (r *RecurringTransaction) Description() string        { return r.description }
func (r *RecurringTransaction) Frequency() Frequency       { return r.frequency }
func (r *RecurringTransaction) Type() TransactionType      { return r.transactionType }
func (r *RecurringTransaction) Schedule() RecurringSchedule { return r.schedule }
func (r *RecurringTransaction) Weekday() *int              { return r.schedule.Weekday }
func (r *RecurringTransaction) DayOfMonth() *int           { return r.schedule.DayOfMonth }
func (r *RecurringTransaction) MonthOfYear() *int          { return r.schedule.MonthOfYear }
func (r *RecurringTransaction) NextDueDate() time.Time     { return r.nextDueDate }
func (r *RecurringTransaction) IsActive() bool             { return r.isActive }
func (r *RecurringTransaction) CreatedAt() time.Time       { return r.createdAt }

func (r *RecurringTransaction) UpdateNextDueDate(newDate time.Time) {
	r.nextDueDate = truncateDay(newDate)
}

// AdvanceNextDue advances next_due_date after a posting
func (r *RecurringTransaction) AdvanceNextDue() error {
	next, err := AdvanceFrom(r.frequency, r.schedule, r.nextDueDate)
	if err != nil {
		return err
	}
	r.nextDueDate = next
	return nil
}

// CalculateNextDueDate is kept for compatibility; prefers schedule-aware advance
func (r *RecurringTransaction) CalculateNextDueDate() time.Time {
	next, err := AdvanceFrom(r.frequency, r.schedule, r.nextDueDate)
	if err != nil {
		return r.nextDueDate.AddDate(0, 0, 1)
	}
	return next
}

// IsDue checks if the recurring transaction is due (inclusive of today)
func (r *RecurringTransaction) IsDue() bool {
	if !r.isActive {
		return false
	}
	now := time.Now()
	today := truncateDay(now)
	due := truncateDay(r.nextDueDate)
	return !due.After(today)
}

// ScheduleLabel returns a short human-readable schedule description
func (r *RecurringTransaction) ScheduleLabel() string {
	switch r.frequency {
	case FrequencyDaily:
		return "Every day"
	case FrequencyWeekly:
		if r.schedule.Weekday != nil {
			return fmt.Sprintf("Every %s", time.Weekday(*r.schedule.Weekday).String())
		}
		return "Weekly"
	case FrequencyMonthly:
		if r.schedule.DayOfMonth != nil {
			return fmt.Sprintf("Monthly on day %d", *r.schedule.DayOfMonth)
		}
		return "Monthly"
	case FrequencyYearly:
		if r.schedule.MonthOfYear != nil && r.schedule.DayOfMonth != nil {
			month := time.Month(*r.schedule.MonthOfYear).String()[:3]
			return fmt.Sprintf("Every %s %d", month, *r.schedule.DayOfMonth)
		}
		return "Yearly"
	default:
		return string(r.frequency)
	}
}
