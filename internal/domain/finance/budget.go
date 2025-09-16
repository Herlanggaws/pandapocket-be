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

// NewBudget creates a new budget
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
	
	// Calculate end date based on period
	var endDate time.Time
	switch period {
	case BudgetPeriodWeekly:
		endDate = startDate.AddDate(0, 0, 7)
	case BudgetPeriodMonthly:
		endDate = startDate.AddDate(0, 1, 0)
	case BudgetPeriodYearly:
		endDate = startDate.AddDate(1, 0, 0)
	default:
		return nil, errors.New("invalid budget period")
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
	var endDate time.Time
	switch newPeriod {
	case BudgetPeriodWeekly:
		endDate = b.startDate.AddDate(0, 0, 7)
	case BudgetPeriodMonthly:
		endDate = b.startDate.AddDate(0, 1, 0)
	case BudgetPeriodYearly:
		endDate = b.startDate.AddDate(1, 0, 0)
	default:
		return errors.New("invalid budget period")
	}
	
	b.period = newPeriod
	b.endDate = endDate
	return nil
}

// UpdateStartDate updates the start date and recalculates end date
func (b *Budget) UpdateStartDate(newStartDate time.Time) {
	b.startDate = newStartDate
	
	// Recalculate end date
	switch b.period {
	case BudgetPeriodWeekly:
		b.endDate = newStartDate.AddDate(0, 0, 7)
	case BudgetPeriodMonthly:
		b.endDate = newStartDate.AddDate(0, 1, 0)
	case BudgetPeriodYearly:
		b.endDate = newStartDate.AddDate(1, 0, 0)
	}
}

// IsActive checks if the budget is currently active
func (b *Budget) IsActive() bool {
	now := time.Now()
	return now.After(b.startDate) && now.Before(b.endDate)
}

// IsExpired checks if the budget has expired
func (b *Budget) IsExpired() bool {
	return time.Now().After(b.endDate)
}
