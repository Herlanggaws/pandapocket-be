package finance

import (
	"context"
	"errors"
	"fmt"
	"time"

	appNotification "panda-pocket/internal/application/notification"
	"panda-pocket/internal/domain/finance"
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
	WalletID        *int    `json:"wallet_id"`
	WalletName      *string `json:"wallet_name,omitempty"`
	ProgressSource  string  `json:"progress_source"`
	CreatedAt       string  `json:"created_at"`
}

type CreateGoalRequest struct {
	Name          string  `json:"name" binding:"required"`
	TargetAmount  float64 `json:"target_amount" binding:"required,gt=0"`
	CurrentAmount float64 `json:"current_amount"`
	TargetDate    string  `json:"target_date" binding:"required"`
	WalletID      *int    `json:"wallet_id"`
}

// UpdateGoalRequest: wallet_id omitted or null unlinks when previously linked;
// frontend always sends wallet_id on edit (number or null).
type UpdateGoalRequest struct {
	Name          string  `json:"name" binding:"required"`
	TargetAmount  float64 `json:"target_amount" binding:"required,gt=0"`
	CurrentAmount float64 `json:"current_amount"`
	TargetDate    string  `json:"target_date" binding:"required"`
	Status        string  `json:"status"`
	WalletID      *int    `json:"wallet_id"`
}

type goalWalletResolver struct {
	walletService *finance.WalletService
}

func (r *goalWalletResolver) resolveLinkedWallet(
	ctx context.Context,
	userID finance.UserID,
	walletID int,
) (*finance.Wallet, float64, error) {
	wallet, err := r.walletService.GetWalletForUser(ctx, userID, finance.NewWalletID(walletID))
	if err != nil {
		return nil, 0, errors.New("wallet not found")
	}
	if wallet.IsArchived() {
		return nil, 0, errors.New("wallet is archived")
	}
	breakdown, err := r.walletService.GetBalance(ctx, userID, wallet.ID())
	if err != nil {
		return nil, 0, err
	}
	return wallet, breakdown.Balance, nil
}

type goalResponseBuilder struct {
	walletService      *finance.WalletService
	goalService        *finance.GoalService
	notificationHelper *appNotification.CreateNotificationHelper
}

func (b *goalResponseBuilder) build(
	ctx context.Context,
	goal *finance.FinancialGoal,
	persistComplete bool,
) (GoalResponse, error) {
	effective := goal.CurrentAmount()
	source := "manual"
	var walletID *int
	var walletName *string

	if goal.WalletID() != nil {
		source = "wallet"
		id := goal.WalletID().Value()
		walletID = &id
		wallet, err := b.walletService.GetWalletForUser(ctx, goal.UserID(), *goal.WalletID())
		if err == nil {
			name := wallet.Name()
			walletName = &name
			breakdown, balErr := b.walletService.GetBalance(ctx, goal.UserID(), *goal.WalletID())
			if balErr == nil {
				effective = breakdown.Balance
				if persistComplete {
					prevAmount := goal.CurrentAmount()
					prevStatus := goal.Status()
					_ = goal.SetCurrentAmount(effective)
					becameComplete := prevStatus == finance.GoalStatusActive &&
						goal.Status() == finance.GoalStatusCompleted
					if b.goalService != nil && (prevAmount != effective || prevStatus != goal.Status()) {
						_ = b.goalService.Save(ctx, goal)
					}
					if becameComplete && b.notificationHelper != nil {
						title := fmt.Sprintf("Goal completed: %s", goal.Name())
						message := fmt.Sprintf("You've reached your target for %s.", goal.Name())
						_ = b.notificationHelper.CreateIfNotRecent(
							ctx,
							goal.UserID().Value(),
							title,
							message,
							"goal_completed",
						)
					}
				}
			}
		}
	}

	return GoalResponse{
		ID:              goal.ID().Value(),
		Name:            goal.Name(),
		TargetAmount:    goal.TargetAmount(),
		CurrencyID:      goal.CurrencyID().Value(),
		CurrentAmount:   effective,
		TargetDate:      goal.TargetDate().Format("2006-01-02"),
		Status:          string(goal.Status()),
		ProgressPercent: finance.ProgressPercent(effective, goal.TargetAmount()),
		WalletID:        walletID,
		WalletName:      walletName,
		ProgressSource:  source,
		CreatedAt:       goal.CreatedAt().Format(time.RFC3339),
	}, nil
}

type CreateGoalUseCase struct {
	goalService        *finance.GoalService
	currencyService    *finance.CurrencyService
	walletService      *finance.WalletService
	notificationHelper *appNotification.CreateNotificationHelper
}

func NewCreateGoalUseCase(
	goalService *finance.GoalService,
	currencyService *finance.CurrencyService,
	walletService *finance.WalletService,
	notificationHelper *appNotification.CreateNotificationHelper,
) *CreateGoalUseCase {
	return &CreateGoalUseCase{
		goalService:        goalService,
		currencyService:    currencyService,
		walletService:      walletService,
		notificationHelper: notificationHelper,
	}
}

func (uc *CreateGoalUseCase) Execute(ctx context.Context, userID int, req CreateGoalRequest) (*GoalResponse, error) {
	targetDate, err := time.Parse("2006-01-02", req.TargetDate)
	if err != nil {
		return nil, err
	}

	user := finance.NewUserID(userID)
	var currencyID finance.CurrencyID
	currentAmount := req.CurrentAmount
	var walletID *finance.WalletID

	if req.WalletID != nil {
		resolver := &goalWalletResolver{walletService: uc.walletService}
		wallet, balance, err := resolver.resolveLinkedWallet(ctx, user, *req.WalletID)
		if err != nil {
			return nil, err
		}
		currencyID = wallet.CurrencyID()
		currentAmount = balance
		id := wallet.ID()
		walletID = &id
	} else {
		currency, err := uc.currencyService.GetPrimaryCurrency(ctx, user)
		if err != nil {
			return nil, err
		}
		currencyID = currency.ID()
	}

	goal, err := uc.goalService.CreateGoal(
		ctx,
		user,
		req.Name,
		req.TargetAmount,
		currencyID,
		currentAmount,
		targetDate,
		walletID,
	)
	if err != nil {
		return nil, err
	}

	builder := &goalResponseBuilder{
		walletService:      uc.walletService,
		goalService:        uc.goalService,
		notificationHelper: uc.notificationHelper,
	}
	resp, err := builder.build(ctx, goal, true)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type GetGoalsUseCase struct {
	goalService        *finance.GoalService
	walletService      *finance.WalletService
	notificationHelper *appNotification.CreateNotificationHelper
}

func NewGetGoalsUseCase(
	goalService *finance.GoalService,
	walletService *finance.WalletService,
	notificationHelper *appNotification.CreateNotificationHelper,
) *GetGoalsUseCase {
	return &GetGoalsUseCase{
		goalService:        goalService,
		walletService:      walletService,
		notificationHelper: notificationHelper,
	}
}

func (uc *GetGoalsUseCase) Execute(ctx context.Context, userID int, includeArchived bool) ([]GoalResponse, error) {
	goals, err := uc.goalService.GetGoals(ctx, finance.NewUserID(userID), includeArchived)
	if err != nil {
		return nil, err
	}
	builder := &goalResponseBuilder{
		walletService:      uc.walletService,
		goalService:        uc.goalService,
		notificationHelper: uc.notificationHelper,
	}
	result := make([]GoalResponse, 0, len(goals))
	for _, g := range goals {
		resp, err := builder.build(ctx, g, true)
		if err != nil {
			return nil, err
		}
		result = append(result, resp)
	}
	return result, nil
}

type GetGoalUseCase struct {
	goalService        *finance.GoalService
	walletService      *finance.WalletService
	notificationHelper *appNotification.CreateNotificationHelper
}

func NewGetGoalUseCase(
	goalService *finance.GoalService,
	walletService *finance.WalletService,
	notificationHelper *appNotification.CreateNotificationHelper,
) *GetGoalUseCase {
	return &GetGoalUseCase{
		goalService:        goalService,
		walletService:      walletService,
		notificationHelper: notificationHelper,
	}
}

func (uc *GetGoalUseCase) Execute(ctx context.Context, userID, id int) (*GoalResponse, error) {
	goal, err := uc.goalService.GetGoalForUser(ctx, finance.NewUserID(userID), finance.NewGoalID(id))
	if err != nil {
		return nil, err
	}
	builder := &goalResponseBuilder{
		walletService:      uc.walletService,
		goalService:        uc.goalService,
		notificationHelper: uc.notificationHelper,
	}
	resp, err := builder.build(ctx, goal, true)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type UpdateGoalUseCase struct {
	goalService        *finance.GoalService
	walletService      *finance.WalletService
	notificationHelper *appNotification.CreateNotificationHelper
}

func NewUpdateGoalUseCase(
	goalService *finance.GoalService,
	walletService *finance.WalletService,
	notificationHelper *appNotification.CreateNotificationHelper,
) *UpdateGoalUseCase {
	return &UpdateGoalUseCase{
		goalService:        goalService,
		walletService:      walletService,
		notificationHelper: notificationHelper,
	}
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

	user := finance.NewUserID(userID)
	goal, err := uc.goalService.GetGoalForUser(ctx, user, finance.NewGoalID(id))
	if err != nil {
		return nil, err
	}

	currentAmount := req.CurrentAmount
	resolver := &goalWalletResolver{walletService: uc.walletService}

	if req.WalletID != nil {
		wallet, balance, err := resolver.resolveLinkedWallet(ctx, user, *req.WalletID)
		if err != nil {
			return nil, err
		}
		if wallet.CurrencyID().Value() != goal.CurrencyID().Value() {
			return nil, errors.New("wallet currency must match goal currency")
		}
		if err := goal.LinkWallet(wallet.ID(), wallet.CurrencyID(), balance); err != nil {
			return nil, err
		}
		currentAmount = balance
	} else if goal.IsLinked() {
		freeze := goal.CurrentAmount()
		if goal.WalletID() != nil {
			if breakdown, balErr := uc.walletService.GetBalance(ctx, user, *goal.WalletID()); balErr == nil {
				freeze = breakdown.Balance
			}
		}
		if err := goal.UnlinkWallet(freeze); err != nil {
			return nil, err
		}
		currentAmount = freeze
	}

	if err := goal.Update(req.Name, req.TargetAmount, currentAmount, targetDate, status); err != nil {
		return nil, err
	}

	// Update() does not clear wallet; re-assert link state after field update
	if req.WalletID != nil {
		wallet, balance, err := resolver.resolveLinkedWallet(ctx, user, *req.WalletID)
		if err != nil {
			return nil, err
		}
		if err := goal.LinkWallet(wallet.ID(), wallet.CurrencyID(), balance); err != nil {
			return nil, err
		}
	}

	if err := uc.goalService.Save(ctx, goal); err != nil {
		return nil, err
	}

	builder := &goalResponseBuilder{
		walletService:      uc.walletService,
		goalService:        uc.goalService,
		notificationHelper: uc.notificationHelper,
	}
	resp, err := builder.build(ctx, goal, true)
	if err != nil {
		return nil, err
	}
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

// UnlinkGoalsForWallet freezes progress and clears wallet_id for all goals linked to a wallet.
func UnlinkGoalsForWallet(
	ctx context.Context,
	goalService *finance.GoalService,
	walletService *finance.WalletService,
	userID finance.UserID,
	walletID finance.WalletID,
) error {
	goals, err := goalService.FindByWalletID(ctx, walletID)
	if err != nil {
		return err
	}
	freeze := 0.0
	if breakdown, balErr := walletService.GetBalance(ctx, userID, walletID); balErr == nil {
		freeze = breakdown.Balance
	}
	for _, goal := range goals {
		if goal.UserID().Value() != userID.Value() {
			continue
		}
		if err := goal.UnlinkWallet(freeze); err != nil {
			return err
		}
		if err := goalService.Save(ctx, goal); err != nil {
			return err
		}
	}
	return nil
}
