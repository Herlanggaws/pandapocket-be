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

type GoalContributionRepository interface {
	Save(ctx context.Context, contribution *GoalContribution) error
	FindByGoalID(ctx context.Context, goalID GoalID) ([]*GoalContribution, error)
	FindByExpenseID(ctx context.Context, expenseID int) (*GoalContribution, error)
	FindByIncomeID(ctx context.Context, incomeID int) (*GoalContribution, error)
	FindByTransferID(ctx context.Context, transferID int) (*GoalContribution, error)
	Delete(ctx context.Context, id GoalContributionID) error
}

type GoalService struct {
	goalRepo         GoalRepository
	contributionRepo GoalContributionRepository
}

func NewGoalService(goalRepo GoalRepository, contributionRepo GoalContributionRepository) *GoalService {
	return &GoalService{goalRepo: goalRepo, contributionRepo: contributionRepo}
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

func (s *GoalService) RecordContribution(
	ctx context.Context,
	userID UserID,
	goalID GoalID,
	amount float64,
	contributedAt time.Time,
	expenseID *int,
	incomeID *int,
	transferID *int,
	note string,
	bumpCurrentAmount bool,
) (*FinancialGoal, *GoalContribution, error) {
	goal, err := s.GetGoalForUser(ctx, userID, goalID)
	if err != nil {
		return nil, nil, err
	}
	if bumpCurrentAmount {
		if err := goal.ApplyContribution(amount); err != nil {
			return nil, nil, err
		}
		if err := s.goalRepo.Save(ctx, goal); err != nil {
			return nil, nil, err
		}
	}
	contribution, err := NewGoalContribution(goalID, userID, amount, contributedAt, expenseID, incomeID, transferID, note)
	if err != nil {
		return nil, nil, err
	}
	if s.contributionRepo == nil {
		return nil, nil, errors.New("contribution repository is not available")
	}
	if err := s.contributionRepo.Save(ctx, contribution); err != nil {
		return nil, nil, err
	}
	return goal, contribution, nil
}

func (s *GoalService) ListContributions(ctx context.Context, userID UserID, goalID GoalID) ([]*GoalContribution, error) {
	if _, err := s.GetGoalForUser(ctx, userID, goalID); err != nil {
		return nil, err
	}
	if s.contributionRepo == nil {
		return []*GoalContribution{}, nil
	}
	return s.contributionRepo.FindByGoalID(ctx, goalID)
}

func (s *GoalService) ReverseContributionByExpenseID(ctx context.Context, userID UserID, expenseID int) error {
	return s.reverseByRef(ctx, userID, func() (*GoalContribution, error) {
		if s.contributionRepo == nil {
			return nil, nil
		}
		return s.contributionRepo.FindByExpenseID(ctx, expenseID)
	})
}

func (s *GoalService) ReverseContributionByIncomeID(ctx context.Context, userID UserID, incomeID int) error {
	return s.reverseByRef(ctx, userID, func() (*GoalContribution, error) {
		if s.contributionRepo == nil {
			return nil, nil
		}
		return s.contributionRepo.FindByIncomeID(ctx, incomeID)
	})
}

func (s *GoalService) ReverseContributionByTransferID(ctx context.Context, userID UserID, transferID int) error {
	return s.reverseByRef(ctx, userID, func() (*GoalContribution, error) {
		if s.contributionRepo == nil {
			return nil, nil
		}
		return s.contributionRepo.FindByTransferID(ctx, transferID)
	})
}

func (s *GoalService) reverseByRef(
	ctx context.Context,
	userID UserID,
	find func() (*GoalContribution, error),
) error {
	contribution, err := find()
	if err != nil {
		return err
	}
	if contribution == nil {
		return nil
	}
	if contribution.UserID().Value() != userID.Value() {
		return errors.New("goal contribution not found")
	}
	if contribution.BumpsCurrentAmount() {
		goal, err := s.GetGoalForUser(ctx, userID, contribution.GoalID())
		if err != nil {
			return err
		}
		if err := goal.ReverseContribution(contribution.Amount()); err != nil {
			return err
		}
		if err := s.goalRepo.Save(ctx, goal); err != nil {
			return err
		}
	}
	return s.contributionRepo.Delete(ctx, contribution.ID())
}
