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

type GoalContributionResponse struct {
	ID            int     `json:"id"`
	GoalID        int     `json:"goal_id"`
	Amount        float64 `json:"amount"`
	ContributedAt string  `json:"contributed_at"`
	ExpenseID     *int    `json:"expense_id,omitempty"`
	IncomeID      *int    `json:"income_id,omitempty"`
	TransferID    *int    `json:"transfer_id,omitempty"`
	Note          string  `json:"note"`
	CreatedAt     string  `json:"created_at"`
}

func toGoalContributionResponse(c *finance.GoalContribution) GoalContributionResponse {
	return GoalContributionResponse{
		ID:            c.ID().Value(),
		GoalID:        c.GoalID().Value(),
		Amount:        c.Amount(),
		ContributedAt: c.ContributedAt().Format("2006-01-02"),
		ExpenseID:     c.ExpenseID(),
		IncomeID:      c.IncomeID(),
		TransferID:    c.TransferID(),
		Note:          c.Note(),
		CreatedAt:     c.CreatedAt().Format(time.RFC3339),
	}
}

type RecordGoalContributionRequest struct {
	Amount         float64 `json:"amount" binding:"required,gt=0"`
	ContributedAt  string  `json:"contributed_at" binding:"required"`
	Note           string  `json:"note"`
	FromWalletID   *int    `json:"from_wallet_id"`
}

type RecordGoalContributionResponse struct {
	Goal         *GoalResponse               `json:"goal"`
	Contribution *GoalContributionResponse   `json:"contribution"`
	Kind         string                      `json:"kind"` // expense | income | transfer
}

type ListGoalContributionsUseCase struct {
	goalService *finance.GoalService
}

func NewListGoalContributionsUseCase(goalService *finance.GoalService) *ListGoalContributionsUseCase {
	return &ListGoalContributionsUseCase{goalService: goalService}
}

func (uc *ListGoalContributionsUseCase) Execute(ctx context.Context, userID, goalID int) ([]GoalContributionResponse, error) {
	list, err := uc.goalService.ListContributions(ctx, finance.NewUserID(userID), finance.NewGoalID(goalID))
	if err != nil {
		return nil, err
	}
	result := make([]GoalContributionResponse, 0, len(list))
	for _, item := range list {
		result = append(result, toGoalContributionResponse(item))
	}
	return result, nil
}

type RecordGoalContributionUseCase struct {
	goalService              *finance.GoalService
	walletService            *finance.WalletService
	transferService          *finance.TransferService
	createTransactionUseCase *CreateTransactionUseCase
	categoryService          *finance.CategoryService
	notificationHelper       *appNotification.CreateNotificationHelper
}

func NewRecordGoalContributionUseCase(
	goalService *finance.GoalService,
	walletService *finance.WalletService,
	transferService *finance.TransferService,
	createTransactionUseCase *CreateTransactionUseCase,
	categoryService *finance.CategoryService,
	notificationHelper *appNotification.CreateNotificationHelper,
) *RecordGoalContributionUseCase {
	return &RecordGoalContributionUseCase{
		goalService:              goalService,
		walletService:            walletService,
		transferService:          transferService,
		createTransactionUseCase: createTransactionUseCase,
		categoryService:          categoryService,
		notificationHelper:       notificationHelper,
	}
}

func (uc *RecordGoalContributionUseCase) resolveCategoryID(
	ctx context.Context,
	userID int,
	categoryType finance.CategoryType,
	preferred ...string,
) (int, error) {
	categories, err := uc.categoryService.GetCategoriesByUserAndType(ctx, finance.NewUserID(userID), categoryType)
	if err != nil {
		return 0, err
	}
	for _, name := range preferred {
		for _, category := range categories {
			if category.Name() == name {
				return category.ID().Value(), nil
			}
		}
	}
	if len(categories) == 0 {
		return 0, errors.New("no category available for goal contribution")
	}
	return categories[0].ID().Value(), nil
}

func (uc *RecordGoalContributionUseCase) Execute(
	ctx context.Context,
	userID int,
	goalID int,
	req RecordGoalContributionRequest,
) (*RecordGoalContributionResponse, error) {
	contributedAt, err := time.Parse("2006-01-02", req.ContributedAt)
	if err != nil {
		return nil, errors.New("invalid contributed_at format. Expected YYYY-MM-DD")
	}

	user := finance.NewUserID(userID)
	goal, err := uc.goalService.GetGoalForUser(ctx, user, finance.NewGoalID(goalID))
	if err != nil {
		return nil, err
	}

	description := fmt.Sprintf("Goal savings (%s)", goal.Name())
	if req.Note != "" {
		description = fmt.Sprintf("%s: %s", description, req.Note)
	}

	var expenseID, incomeID, transferID *int
	var kind string
	bumpCurrent := false

	if goal.IsLinked() {
		linkedID := goal.WalletID().Value()
		wallets, wErr := uc.walletService.GetWallets(ctx, user, false)
		if wErr != nil {
			return nil, wErr
		}
		otherWallets := make([]*finance.Wallet, 0)
		for _, w := range wallets {
			if w.ID().Value() != linkedID && w.CurrencyID().Value() == goal.CurrencyID().Value() {
				otherWallets = append(otherWallets, w)
			}
		}

		if len(otherWallets) > 0 {
			if req.FromWalletID == nil || *req.FromWalletID <= 0 {
				return nil, errors.New("from_wallet_id is required when transferring to a linked goal wallet")
			}
			if *req.FromWalletID == linkedID {
				return nil, errors.New("from_wallet_id must differ from the linked goal wallet")
			}
			transfer, tErr := uc.transferService.CreateTransfer(
				ctx,
				user,
				finance.NewWalletID(*req.FromWalletID),
				finance.NewWalletID(linkedID),
				req.Amount,
				description,
				contributedAt,
			)
			if tErr != nil {
				return nil, tErr
			}
			id := transfer.ID().Value()
			transferID = &id
			kind = "transfer"
		} else {
			categoryID, cErr := uc.resolveCategoryID(ctx, userID, finance.CategoryTypeIncome, "Salary", "Other")
			if cErr != nil {
				return nil, cErr
			}
			wid := linkedID
			tx, txErr := uc.createTransactionUseCase.Execute(ctx, userID, CreateTransactionRequest{
				CategoryID:  categoryID,
				WalletID:    &wid,
				Amount:      req.Amount,
				Description: description,
				Date:        req.ContributedAt,
				Type:        "income",
			})
			if txErr != nil {
				return nil, txErr
			}
			if tx.ID == 0 {
				return nil, errors.New("created income is missing id")
			}
			id := tx.ID
			incomeID = &id
			kind = "income"
		}
	} else {
		if req.FromWalletID == nil || *req.FromWalletID <= 0 {
			return nil, errors.New("from_wallet_id is required for manual goal contributions")
		}
		categoryID, cErr := uc.resolveCategoryID(ctx, userID, finance.CategoryTypeExpense, "Savings", "Goals", "Other")
		if cErr != nil {
			return nil, cErr
		}
		tx, txErr := uc.createTransactionUseCase.Execute(ctx, userID, CreateTransactionRequest{
			CategoryID:  categoryID,
			WalletID:    req.FromWalletID,
			Amount:      req.Amount,
			Description: description,
			Date:        req.ContributedAt,
			Type:        "expense",
		})
		if txErr != nil {
			return nil, txErr
		}
		if tx.ID == 0 {
			return nil, errors.New("created expense is missing id")
		}
		id := tx.ID
		expenseID = &id
		kind = "expense"
		bumpCurrent = true
	}

	updatedGoal, contribution, err := uc.goalService.RecordContribution(
		ctx,
		user,
		finance.NewGoalID(goalID),
		req.Amount,
		contributedAt,
		expenseID,
		incomeID,
		transferID,
		req.Note,
		bumpCurrent,
	)
	if err != nil {
		return nil, err
	}

	builder := &goalResponseBuilder{
		walletService:      uc.walletService,
		goalService:        uc.goalService,
		notificationHelper: uc.notificationHelper,
	}
	goalResp, err := builder.build(ctx, updatedGoal, true)
	if err != nil {
		return nil, err
	}
	contribResp := toGoalContributionResponse(contribution)
	return &RecordGoalContributionResponse{
		Goal:         &goalResp,
		Contribution: &contribResp,
		Kind:         kind,
	}, nil
}
