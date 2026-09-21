package finance

import (
	"context"
	"errors"
	"time"
)

type GoalRepository interface {
	Save(ctx context.Context, goal *FinancialGoal) error
	FindByID(ctx context.Context, id GoalID) (*FinancialGoal, error)
	FindByUserID(ctx context.Context, userID UserID, includeArchived bool) ([]*FinancialGoal, error)
	FindByWalletID(ctx context.Context, walletID WalletID) ([]*FinancialGoal, error)
	FindActiveForDeadlineDates(ctx context.Context, dates []time.Time) ([]*FinancialGoal, error)
	Delete(ctx context.Context, id GoalID, userID UserID) error
}

type GoalService struct {
	goalRepo GoalRepository
}

func NewGoalService(goalRepo GoalRepository) *GoalService {
	return &GoalService{goalRepo: goalRepo}
}

func (s *GoalService) CreateGoal(
	ctx context.Context,
	userID UserID,
	name string,
	targetAmount float64,
	currencyID CurrencyID,
	currentAmount float64,
	targetDate time.Time,
	walletID *WalletID,
) (*FinancialGoal, error) {
	goal, err := NewFinancialGoal(userID, name, targetAmount, currencyID, currentAmount, targetDate)
	if err != nil {
		return nil, err
	}
	if walletID != nil {
		if err := goal.LinkWallet(*walletID, currencyID, currentAmount); err != nil {
			return nil, err
		}
	}
	if err := s.goalRepo.Save(ctx, goal); err != nil {
		return nil, err
	}
	return goal, nil
}

func (s *GoalService) GetGoals(ctx context.Context, userID UserID, includeArchived bool) ([]*FinancialGoal, error) {
	return s.goalRepo.FindByUserID(ctx, userID, includeArchived)
}

func (s *GoalService) GetGoalForUser(ctx context.Context, userID UserID, id GoalID) (*FinancialGoal, error) {
	goal, err := s.goalRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if goal.UserID().Value() != userID.Value() {
		return nil, errors.New("goal not found")
	}
	return goal, nil
}

func (s *GoalService) UpdateGoal(
	ctx context.Context,
	userID UserID,
	id GoalID,
	name string,
	targetAmount float64,
	currentAmount float64,
	targetDate time.Time,
	status GoalStatus,
) (*FinancialGoal, error) {
	goal, err := s.GetGoalForUser(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if err := goal.Update(name, targetAmount, currentAmount, targetDate, status); err != nil {
		return nil, err
	}
	if err := s.goalRepo.Save(ctx, goal); err != nil {
		return nil, err
	}
	return goal, nil
}

func (s *GoalService) Save(ctx context.Context, goal *FinancialGoal) error {
	return s.goalRepo.Save(ctx, goal)
}

func (s *GoalService) FindByWalletID(ctx context.Context, walletID WalletID) ([]*FinancialGoal, error) {
	return s.goalRepo.FindByWalletID(ctx, walletID)
}

func (s *GoalService) FindActiveForDeadlineDates(ctx context.Context, dates []time.Time) ([]*FinancialGoal, error) {
	return s.goalRepo.FindActiveForDeadlineDates(ctx, dates)
}

func (s *GoalService) DeleteGoal(ctx context.Context, userID UserID, id GoalID) error {
	if _, err := s.GetGoalForUser(ctx, userID, id); err != nil {
		return err
	}
	return s.goalRepo.Delete(ctx, id, userID)
}
