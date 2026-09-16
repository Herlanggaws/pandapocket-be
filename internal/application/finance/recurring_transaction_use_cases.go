package finance

import (
	"context"
	"errors"
	"fmt"
	"time"

	appNotification "panda-pocket/internal/application/notification"
	domainFinance "panda-pocket/internal/domain/finance"
	domainIdentity "panda-pocket/internal/domain/identity"
)

type RecurringTransactionResponse struct {
	ID          int     `json:"id"`
	Type        string  `json:"type"`
	CategoryID  int     `json:"category_id"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Frequency   string  `json:"frequency"`
	NextDueDate string  `json:"next_due_date"`
	NextDate    string  `json:"next_date"`
	IsActive    bool    `json:"is_active"`
	Category    *CategoryResponse `json:"category,omitempty"`
}

type CreateRecurringTransactionRequest struct {
	Type        string  `json:"type" binding:"required,oneof=expense income"`
	CategoryID  int     `json:"category_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
	Frequency   string  `json:"frequency" binding:"required,oneof=daily weekly monthly yearly"`
	NextDueDate string  `json:"next_due_date"`
}

type CreateRecurringTransactionUseCase struct {
	recurringRepo   domainFinance.RecurringTransactionRepository
	currencyService *domainFinance.CurrencyService
	categoryService *domainFinance.CategoryService
}

func NewCreateRecurringTransactionUseCase(
	recurringRepo domainFinance.RecurringTransactionRepository,
	currencyService *domainFinance.CurrencyService,
	categoryService *domainFinance.CategoryService,
) *CreateRecurringTransactionUseCase {
	return &CreateRecurringTransactionUseCase{
		recurringRepo:   recurringRepo,
		currencyService: currencyService,
		categoryService: categoryService,
	}
}

func (uc *CreateRecurringTransactionUseCase) Execute(ctx context.Context, userID int, req CreateRecurringTransactionRequest) (*RecurringTransactionResponse, error) {
	currency, err := uc.currencyService.GetPrimaryCurrency(ctx, domainFinance.NewUserID(userID))
	if err != nil {
		return nil, errors.New("failed to get primary currency")
	}

	category, err := uc.categoryService.GetCategoryByID(ctx, domainFinance.NewCategoryID(req.CategoryID))
	if err != nil {
		return nil, errors.New("category not found")
	}

	money, err := domainFinance.NewMoney(req.Amount, currency.ID())
	if err != nil {
		return nil, err
	}

	nextDue := time.Now()
	if req.NextDueDate != "" {
		parsed, err := time.Parse("2006-01-02", req.NextDueDate)
		if err != nil {
			return nil, errors.New("invalid next_due_date format. Expected YYYY-MM-DD")
		}
		nextDue = parsed
	} else {
		nextDue = time.Date(nextDue.Year(), nextDue.Month(), nextDue.Day(), 0, 0, 0, 0, nextDue.Location())
	}

	rt, err := domainFinance.NewRecurringTransaction(
		domainFinance.NewUserID(userID),
		domainFinance.NewCategoryID(req.CategoryID),
		currency.ID(),
		money,
		req.Description,
		domainFinance.Frequency(req.Frequency),
		domainFinance.TransactionType(req.Type),
		nextDue,
	)
	if err != nil {
		return nil, err
	}

	if err := uc.recurringRepo.Save(ctx, rt); err != nil {
		return nil, err
	}

	resp := toRecurringResponse(rt, category)
	return &resp, nil
}

type GetRecurringTransactionsUseCase struct {
	recurringRepo      domainFinance.RecurringTransactionRepository
	transactionService *domainFinance.TransactionService
	categoryService    *domainFinance.CategoryService
	prefsRepo          domainIdentity.PreferencesRepository
	notificationHelper *appNotification.CreateNotificationHelper
}

func NewGetRecurringTransactionsUseCase(
	recurringRepo domainFinance.RecurringTransactionRepository,
	transactionService *domainFinance.TransactionService,
	categoryService *domainFinance.CategoryService,
	prefsRepo domainIdentity.PreferencesRepository,
	notificationHelper *appNotification.CreateNotificationHelper,
) *GetRecurringTransactionsUseCase {
	return &GetRecurringTransactionsUseCase{
		recurringRepo:      recurringRepo,
		transactionService: transactionService,
		categoryService:    categoryService,
		prefsRepo:          prefsRepo,
		notificationHelper: notificationHelper,
	}
}

func (uc *GetRecurringTransactionsUseCase) Execute(ctx context.Context, userID int) ([]RecurringTransactionResponse, error) {
	items, err := uc.recurringRepo.FindByUserID(ctx, domainFinance.NewUserID(userID))
	if err != nil {
		return nil, err
	}

	remindersEnabled := true
	if prefs, err := uc.prefsRepo.FindByUserID(ctx, domainIdentity.NewUserID(userID)); err == nil && prefs != nil {
		remindersEnabled = prefs.RecurringReminders()
	}

	for _, rt := range items {
		for rt.IsDue() {
			dueDate := rt.NextDueDate()
			_, err := uc.transactionService.CreateTransaction(
				ctx,
				rt.UserID(),
				rt.CategoryID(),
				rt.CurrencyID(),
				rt.Amount(),
				rt.Description(),
				dueDate,
				rt.Type(),
			)
			if err != nil {
				break
			}

			if remindersEnabled && uc.notificationHelper != nil {
				title := "Recurring transaction posted"
				message := fmt.Sprintf("%s (%s) was posted automatically.", rt.Description(), rt.Type())
				if rt.Description() == "" {
					message = fmt.Sprintf("A recurring %s was posted automatically.", rt.Type())
				}
				_ = uc.notificationHelper.CreateIfNotRecent(ctx, userID, title, message, "recurring_reminder")
			}

			rt.UpdateNextDueDate(rt.CalculateNextDueDate())
			if err := uc.recurringRepo.Save(ctx, rt); err != nil {
				break
			}
			// safety: avoid infinite loop if frequency somehow stuck
			if rt.NextDueDate().Equal(dueDate) {
				break
			}
		}
	}

	items, err = uc.recurringRepo.FindByUserID(ctx, domainFinance.NewUserID(userID))
	if err != nil {
		return nil, err
	}

	result := make([]RecurringTransactionResponse, 0, len(items))
	for _, rt := range items {
		var category *domainFinance.Category
		if cat, err := uc.categoryService.GetCategoryByID(ctx, rt.CategoryID()); err == nil {
			category = cat
		}
		result = append(result, toRecurringResponse(rt, category))
	}
	return result, nil
}

type DeleteRecurringTransactionUseCase struct {
	recurringRepo domainFinance.RecurringTransactionRepository
}

func NewDeleteRecurringTransactionUseCase(recurringRepo domainFinance.RecurringTransactionRepository) *DeleteRecurringTransactionUseCase {
	return &DeleteRecurringTransactionUseCase{recurringRepo: recurringRepo}
}

func (uc *DeleteRecurringTransactionUseCase) Execute(ctx context.Context, userID, id int) error {
	rt, err := uc.recurringRepo.FindByID(ctx, domainFinance.NewRecurringTransactionID(id))
	if err != nil {
		return errors.New("recurring transaction not found")
	}
	if rt.UserID().Value() != userID {
		return errors.New("recurring transaction not found")
	}
	return uc.recurringRepo.Delete(ctx, domainFinance.NewRecurringTransactionID(id))
}

func toRecurringResponse(rt *domainFinance.RecurringTransaction, category *domainFinance.Category) RecurringTransactionResponse {
	due := rt.NextDueDate().Format("2006-01-02")
	resp := RecurringTransactionResponse{
		ID:          rt.ID().Value(),
		Type:        string(rt.Type()),
		CategoryID:  rt.CategoryID().Value(),
		Amount:      rt.Amount().Amount(),
		Description: rt.Description(),
		Frequency:   string(rt.Frequency()),
		NextDueDate: due,
		NextDate:    due,
		IsActive:    rt.IsActive(),
	}
	if category != nil {
		resp.Category = buildCategoryResponse(category)
	}
	return resp
}
