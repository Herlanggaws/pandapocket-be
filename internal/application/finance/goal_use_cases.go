package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
	"time"
)

type GoalResponse struct {
	ID              int     `json:"id"`
	Name            string  `json:"name"`
	TargetAmount    float64 `json:"target_amount"`
	CurrencyID      int     `json:"currency_id"`
	CurrentAmount   float64 `json:"current_amount"`
	TargetDate      string  `json:"target_date"`
	Status          string  `json:"status"`
	ProgressPercent float64 `json:"progress_percent"`
	CreatedAt       string  `json:"created_at"`
}

type CreateGoalRequest struct {
	Name          string  `json:"name" binding:"required"`
	TargetAmount  float64 `json:"target_amount" binding:"required,gt=0"`
	CurrentAmount float64 `json:"current_amount"`
	TargetDate    string  `json:"target_date" binding:"required"`
}

type UpdateGoalRequest struct {
	Name          string  `json:"name" binding:"required"`
	TargetAmount  float64 `json:"target_amount" binding:"required,gt=0"`
	CurrentAmount float64 `json:"current_amount"`
	TargetDate    string  `json:"target_date" binding:"required"`
	Status        string  `json:"status"`
}

func toGoalResponse(goal *finance.FinancialGoal) GoalResponse {
	return GoalResponse{
		ID:              goal.ID().Value(),
		Name:            goal.Name(),
		TargetAmount:    goal.TargetAmount(),
		CurrencyID:      goal.CurrencyID().Value(),
		CurrentAmount:   goal.CurrentAmount(),
		TargetDate:      goal.TargetDate().Format("2006-01-02"),
		Status:          string(goal.Status()),
		ProgressPercent: goal.ProgressPercent(),
		CreatedAt:       goal.CreatedAt().Format(time.RFC3339),
	}
}

type CreateGoalUseCase struct {
	goalService     *finance.GoalService
	currencyService *finance.CurrencyService
}

func NewCreateGoalUseCase(goalService *finance.GoalService, currencyService *finance.CurrencyService) *CreateGoalUseCase {
	return &CreateGoalUseCase{goalService: goalService, currencyService: currencyService}
}

func (uc *CreateGoalUseCase) Execute(ctx context.Context, userID int, req CreateGoalRequest) (*GoalResponse, error) {
	targetDate, err := time.Parse("2006-01-02", req.TargetDate)
	if err != nil {
		return nil, err
	}
	currency, err := uc.currencyService.GetPrimaryCurrency(ctx, finance.NewUserID(userID))
	if err != nil {
		return nil, err
	}
	goal, err := uc.goalService.CreateGoal(
		ctx,
		finance.NewUserID(userID),
		req.Name,
		req.TargetAmount,
		currency.ID(),
		req.CurrentAmount,
		targetDate,
	)
	if err != nil {
		return nil, err
	}
	resp := toGoalResponse(goal)
	return &resp, nil
}

type GetGoalsUseCase struct {
	goalService *finance.GoalService
}

func NewGetGoalsUseCase(goalService *finance.GoalService) *GetGoalsUseCase {
	return &GetGoalsUseCase{goalService: goalService}
}

func (uc *GetGoalsUseCase) Execute(ctx context.Context, userID int, includeArchived bool) ([]GoalResponse, error) {
	goals, err := uc.goalService.GetGoals(ctx, finance.NewUserID(userID), includeArchived)
	if err != nil {
		return nil, err
	}
	result := make([]GoalResponse, 0, len(goals))
	for _, g := range goals {
		result = append(result, toGoalResponse(g))
	}
	return result, nil
}

type GetGoalUseCase struct {
	goalService *finance.GoalService
}

func NewGetGoalUseCase(goalService *finance.GoalService) *GetGoalUseCase {
	return &GetGoalUseCase{goalService: goalService}
}

func (uc *GetGoalUseCase) Execute(ctx context.Context, userID, id int) (*GoalResponse, error) {
	goal, err := uc.goalService.GetGoalForUser(ctx, finance.NewUserID(userID), finance.NewGoalID(id))
	if err != nil {
		return nil, err
	}
	resp := toGoalResponse(goal)
	return &resp, nil
}

type UpdateGoalUseCase struct {
	goalService *finance.GoalService
}

func NewUpdateGoalUseCase(goalService *finance.GoalService) *UpdateGoalUseCase {
	return &UpdateGoalUseCase{goalService: goalService}
}

func (uc *UpdateGoalUseCase) Execute(ctx context.Context, userID, id int, req UpdateGoalRequest) (*GoalResponse, error) {
	targetDate, err := time.Parse("2006-01-02", req.TargetDate)
	if err != nil {
		return nil, err
	}
	status, err := finance.ParseGoalStatus(req.Status)
	if err != nil {
		return nil, err
	}
	goal, err := uc.goalService.UpdateGoal(
		ctx,
		finance.NewUserID(userID),
		finance.NewGoalID(id),
		req.Name,
		req.TargetAmount,
		req.CurrentAmount,
		targetDate,
		status,
	)
	if err != nil {
		return nil, err
	}
	resp := toGoalResponse(goal)
	return &resp, nil
}

type DeleteGoalUseCase struct {
	goalService *finance.GoalService
}

func NewDeleteGoalUseCase(goalService *finance.GoalService) *DeleteGoalUseCase {
	return &DeleteGoalUseCase{goalService: goalService}
}

func (uc *DeleteGoalUseCase) Execute(ctx context.Context, userID, id int) error {
	return uc.goalService.DeleteGoal(ctx, finance.NewUserID(userID), finance.NewGoalID(id))
}
