package finance

import (
	"errors"
	"time"
)

type GoalStatus string

const (
	GoalStatusActive    GoalStatus = "active"
	GoalStatusCompleted GoalStatus = "completed"
	GoalStatusArchived  GoalStatus = "archived"
)

func ParseGoalStatus(value string) (GoalStatus, error) {
	switch GoalStatus(value) {
	case GoalStatusActive, GoalStatusCompleted, GoalStatusArchived:
		return GoalStatus(value), nil
	case "":
		return GoalStatusActive, nil
	default:
		return "", errors.New("invalid goal status")
	}
}

type GoalID struct {
	value int
}

func NewGoalID(id int) GoalID {
	return GoalID{value: id}
}

func (g GoalID) Value() int {
	return g.value
}

type FinancialGoal struct {
	id            GoalID
	userID        UserID
	name          string
	targetAmount  float64
	currencyID    CurrencyID
	currentAmount float64
	targetDate    time.Time
	status        GoalStatus
	walletID      *WalletID
	createdAt     time.Time
}

func NewFinancialGoal(
	userID UserID,
	name string,
	targetAmount float64,
	currencyID CurrencyID,
	currentAmount float64,
	targetDate time.Time,
) (*FinancialGoal, error) {
	if name == "" {
		return nil, errors.New("goal name cannot be empty")
	}
	if targetAmount <= 0 {
		return nil, errors.New("target amount must be positive")
	}
	if currentAmount < 0 {
		return nil, errors.New("current amount cannot be negative")
	}
	if targetDate.IsZero() {
		return nil, errors.New("target date is required")
	}

	status := GoalStatusActive
	if currentAmount >= targetAmount {
		status = GoalStatusCompleted
	}

	return &FinancialGoal{
		userID:        userID,
		name:          name,
		targetAmount:  targetAmount,
		currencyID:    currencyID,
		currentAmount: currentAmount,
		targetDate:    targetDate,
		status:        status,
		createdAt:     time.Now(),
	}, nil
}

func ReconstituteFinancialGoal(
	id GoalID,
	userID UserID,
	name string,
	targetAmount float64,
	currencyID CurrencyID,
	currentAmount float64,
	targetDate time.Time,
	status GoalStatus,
	walletID *WalletID,
	createdAt time.Time,
) *FinancialGoal {
	return &FinancialGoal{
		id:            id,
		userID:        userID,
		name:          name,
		targetAmount:  targetAmount,
		currencyID:    currencyID,
		currentAmount: currentAmount,
		targetDate:    targetDate,
		status:        status,
		walletID:      walletID,
		createdAt:     createdAt,
	}
}

func (g *FinancialGoal) AssignID(id GoalID) { g.id = id }
func (g *FinancialGoal) ID() GoalID         { return g.id }
func (g *FinancialGoal) UserID() UserID     { return g.userID }
func (g *FinancialGoal) Name() string       { return g.name }
func (g *FinancialGoal) TargetAmount() float64 {
	return g.targetAmount
}
func (g *FinancialGoal) CurrencyID() CurrencyID { return g.currencyID }
func (g *FinancialGoal) CurrentAmount() float64 {
	return g.currentAmount
}
func (g *FinancialGoal) TargetDate() time.Time { return g.targetDate }
func (g *FinancialGoal) Status() GoalStatus {
	if g.status == "" {
		return GoalStatusActive
	}
	return g.status
}
func (g *FinancialGoal) WalletID() *WalletID { return g.walletID }
func (g *FinancialGoal) CreatedAt() time.Time { return g.createdAt }

func (g *FinancialGoal) IsLinked() bool {
	return g.walletID != nil
}

func (g *FinancialGoal) ProgressPercent() float64 {
	return ProgressPercent(g.currentAmount, g.targetAmount)
}

func ProgressPercent(currentAmount, targetAmount float64) float64 {
	if targetAmount <= 0 {
		return 0
	}
	pct := (currentAmount / targetAmount) * 100
	if pct > 100 {
		return 100
	}
	if pct < 0 {
		return 0
	}
	return pct
}

func (g *FinancialGoal) Update(
	name string,
	targetAmount float64,
	currentAmount float64,
	targetDate time.Time,
	status GoalStatus,
) error {
	if name == "" {
		return errors.New("goal name cannot be empty")
	}
	if targetAmount <= 0 {
		return errors.New("target amount must be positive")
	}
	if currentAmount < 0 {
		return errors.New("current amount cannot be negative")
	}
	if targetDate.IsZero() {
		return errors.New("target date is required")
	}
	parsed, err := ParseGoalStatus(string(status))
	if err != nil {
		return err
	}
	g.name = name
	g.targetAmount = targetAmount
	g.currentAmount = currentAmount
	g.targetDate = targetDate
	if parsed != GoalStatusArchived && currentAmount >= targetAmount {
		g.status = GoalStatusCompleted
	} else {
		g.status = parsed
	}
	return nil
}

func (g *FinancialGoal) LinkWallet(walletID WalletID, currencyID CurrencyID, currentAmount float64) error {
	if currentAmount < 0 {
		return errors.New("current amount cannot be negative")
	}
	id := walletID
	g.walletID = &id
	g.currencyID = currencyID
	g.currentAmount = currentAmount
	if g.status != GoalStatusArchived && currentAmount >= g.targetAmount {
		g.status = GoalStatusCompleted
	}
	return nil
}

func (g *FinancialGoal) UnlinkWallet(freezeAmount float64) error {
	if freezeAmount < 0 {
		return errors.New("current amount cannot be negative")
	}
	g.walletID = nil
	g.currentAmount = freezeAmount
	if g.status != GoalStatusArchived && freezeAmount >= g.targetAmount {
		g.status = GoalStatusCompleted
	} else if g.status == GoalStatusCompleted && freezeAmount < g.targetAmount {
		g.status = GoalStatusActive
	}
	return nil
}

func (g *FinancialGoal) SetCurrentAmount(amount float64) error {
	if amount < 0 {
		return errors.New("current amount cannot be negative")
	}
	g.currentAmount = amount
	if g.status != GoalStatusArchived && amount >= g.targetAmount {
		g.status = GoalStatusCompleted
	}
	return nil
}

func (g *FinancialGoal) MarkCompleted() {
	if g.status != GoalStatusArchived {
		g.status = GoalStatusCompleted
	}
}

func (g *FinancialGoal) Archive() {
	g.status = GoalStatusArchived
}
