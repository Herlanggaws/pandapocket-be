package finance

import (
	"errors"
	"time"
)

// BudgetPeriod represents the period for a budget
type BudgetPeriod string

const (
	BudgetPeriodWeekly  BudgetPeriod = "weekly"
	BudgetPeriodMonthly BudgetPeriod = "monthly"
	BudgetPeriodYearly  BudgetPeriod = "yearly"
)

// Budget represents a budget
type Budget struct {
	id         BudgetID
	userID     UserID
	categoryID CategoryID
	amount     Money
	period     BudgetPeriod
	startDate  time.Time
	endDate    time.Time
	createdAt  time.Time
}

// BudgetID is a value object representing a budget identifier
type BudgetID struct {
	value int
}

func NewBudgetID(id int) BudgetID {
	return BudgetID{value: id}
}

func (b BudgetID) Value() int {
	return b.value
}

func calculateBudgetEndDate(period BudgetPeriod, startDate time.Time) (time.Time, error) {
	switch period {
	case BudgetPeriodWeekly:
		return startDate.AddDate(0, 0, 6), nil
	case BudgetPeriodMonthly:
		return startDate.AddDate(0, 1, 0).AddDate(0, 0, -1), nil
	case BudgetPeriodYearly:
		return startDate.AddDate(1, 0, 0).AddDate(0, 0, -1), nil
	default:
		return time.Time{}, errors.New("invalid budget period")
	}
}

// NewBudget creates a new budget with an auto-calculated inclusive end date
func NewBudget(
	id BudgetID,
	userID UserID,
	categoryID CategoryID,
	amount Money,
	period BudgetPeriod,
	startDate time.Time,
) (*Budget, error) {
	if amount.Amount() <= 0 {
		return nil, errors.New("budget amount must be positive")
	}

	endDate, err := calculateBudgetEndDate(period, startDate)
	if err != nil {
		return nil, err
	}

	return &Budget{
		id:         id,
		userID:     userID,
		categoryID: categoryID,
		amount:     amount,
		period:     period,
		startDate:  startDate,
		endDate:    endDate,
		createdAt:  time.Now(),
	}, nil
}

// ReconstituteBudget rebuilds a budget from persisted state without recalculating dates
func ReconstituteBudget(
	id BudgetID,
	userID UserID,
	categoryID CategoryID,
	amount Money,
	period BudgetPeriod,
	startDate time.Time,
	endDate time.Time,
	createdAt time.Time,
) (*Budget, error) {
	if amount.Amount() <= 0 {
		return nil, errors.New("budget amount must be positive")
	}

	switch period {
	case BudgetPeriodWeekly, BudgetPeriodMonthly, BudgetPeriodYearly:
	default:
		return nil, errors.New("invalid budget period")
	}

	if endDate.Before(startDate) {
		return nil, errors.New("end date must be on or after start date")
	}

	return &Budget{
		id:         id,
		userID:     userID,
		categoryID: categoryID,
		amount:     amount,
		period:     period,
		startDate:  startDate,
		endDate:    endDate,
		createdAt:  createdAt,
	}, nil
}

// AssignID sets the budget ID after persistence
func (b *Budget) AssignID(id BudgetID) {
	b.id = id
}

// Getters
func (b *Budget) ID() BudgetID {
	return b.id
}

func (b *Budget) UserID() UserID {
	return b.userID
}

func (b *Budget) CategoryID() CategoryID {
	return b.categoryID
}

func (b *Budget) Amount() Money {
	return b.amount
}

func (b *Budget) Period() BudgetPeriod {
	return b.period
}

func (b *Budget) StartDate() time.Time {
	return b.startDate
}

func (b *Budget) EndDate() time.Time {
	return b.endDate
}

func (b *Budget) CreatedAt() time.Time {
	return b.createdAt
}

// UpdateAmount updates the budget amount
func (b *Budget) UpdateAmount(newAmount Money) error {
	if newAmount.Amount() <= 0 {
		return errors.New("budget amount must be positive")
	}
	if newAmount.Currency() != b.amount.Currency() {
		return errors.New("cannot change currency of existing budget")
	}
	b.amount = newAmount
	return nil
}

// UpdatePeriod updates the budget period and recalculates end date
func (b *Budget) UpdatePeriod(newPeriod BudgetPeriod) error {
	endDate, err := calculateBudgetEndDate(newPeriod, b.startDate)
	if err != nil {
		return err
	}

	b.period = newPeriod
	b.endDate = endDate
	return nil
}

// UpdateStartDate updates the start date and recalculates end date
func (b *Budget) UpdateStartDate(newStartDate time.Time) error {
	endDate, err := calculateBudgetEndDate(b.period, newStartDate)
	if err != nil {
		return err
	}

	b.startDate = newStartDate
	b.endDate = endDate
	return nil
}

// UpdateEndDate updates the end date
func (b *Budget) UpdateEndDate(newEndDate time.Time) error {
	if newEndDate.Before(b.startDate) {
		return errors.New("end date must be on or after start date")
	}
	b.endDate = newEndDate
	return nil
}

// OverlapsWith reports whether this budget's date range overlaps another inclusive range
func (b *Budget) OverlapsWith(startDate, endDate time.Time) bool {
	return !b.endDate.Before(startDate) && !endDate.Before(b.startDate)
}

// IsActive checks if the budget is currently active (inclusive date window)
func (b *Budget) IsActive() bool {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	start := time.Date(b.startDate.Year(), b.startDate.Month(), b.startDate.Day(), 0, 0, 0, 0, b.startDate.Location())
	end := time.Date(b.endDate.Year(), b.endDate.Month(), b.endDate.Day(), 0, 0, 0, 0, b.endDate.Location())
	return !today.Before(start) && !today.After(end)
}

// IsExpired checks if the budget has expired
func (b *Budget) IsExpired() bool {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := time.Date(b.endDate.Year(), b.endDate.Month(), b.endDate.Day(), 0, 0, 0, 0, b.endDate.Location())
	return today.After(end)
}
