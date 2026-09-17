package finance

import (
	"context"
	"errors"
	"time"

	domainFinance "panda-pocket/internal/domain/finance"
)

type RecurringTransactionResponse struct {
	ID            int               `json:"id"`
	Type          string            `json:"type"`
	CategoryID    int               `json:"category_id"`
	Amount        float64           `json:"amount"`
	Description   string            `json:"description"`
	Frequency     string            `json:"frequency"`
	Weekday       *int              `json:"weekday,omitempty"`
	DayOfMonth    *int              `json:"day_of_month,omitempty"`
	MonthOfYear   *int              `json:"month_of_year,omitempty"`
	ScheduleLabel string            `json:"schedule_label"`
	NextDueDate   string            `json:"next_due_date"`
	NextDate      string            `json:"next_date"`
	IsActive      bool              `json:"is_active"`
	Category      *CategoryResponse `json:"category,omitempty"`
}

type CreateRecurringTransactionRequest struct {
	Type        string  `json:"type" binding:"required,oneof=expense income"`
	CategoryID  int     `json:"category_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
	Frequency   string  `json:"frequency" binding:"required,oneof=daily weekly monthly yearly"`
	Weekday     *int    `json:"weekday"`
	DayOfMonth  *int    `json:"day_of_month"`
	MonthOfYear *int    `json:"month_of_year"`
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
	if string(category.Type()) != req.Type {
		return nil, errors.New("category type does not match transaction type")
	}

	money, err := domainFinance.NewMoney(req.Amount, currency.ID())
	if err != nil {
		return nil, err
	}

	from := time.Now()
	if req.NextDueDate != "" {
		parsed, err := time.Parse("2006-01-02", req.NextDueDate)
		if err != nil {
			return nil, errors.New("invalid next_due_date format. Expected YYYY-MM-DD")
		}
		from = parsed
	}

	schedule := domainFinance.RecurringSchedule{
		Weekday:     req.Weekday,
		DayOfMonth:  req.DayOfMonth,
		MonthOfYear: req.MonthOfYear,
	}

	rt, err := domainFinance.NewRecurringTransaction(
		domainFinance.NewUserID(userID),
		domainFinance.NewCategoryID(req.CategoryID),
		currency.ID(),
		money,
		req.Description,
		domainFinance.Frequency(req.Frequency),
		domainFinance.TransactionType(req.Type),
		schedule,
		from,
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
	recurringRepo   domainFinance.RecurringTransactionRepository
	categoryService *domainFinance.CategoryService
	enqueueUseCase  *EnqueueDueRecurringUseCase
}

func NewGetRecurringTransactionsUseCase(
	recurringRepo domainFinance.RecurringTransactionRepository,
	categoryService *domainFinance.CategoryService,
	enqueueUseCase *EnqueueDueRecurringUseCase,
) *GetRecurringTransactionsUseCase {
	return &GetRecurringTransactionsUseCase{
		recurringRepo:   recurringRepo,
		categoryService: categoryService,
		enqueueUseCase:  enqueueUseCase,
	}
}

func (uc *GetRecurringTransactionsUseCase) Execute(ctx context.Context, userID int) ([]RecurringTransactionResponse, error) {
	if uc.enqueueUseCase != nil {
		if err := uc.enqueueUseCase.Execute(ctx, userID); err != nil {
			return nil, err
		}
	}

	items, err := uc.recurringRepo.FindByUserID(ctx, domainFinance.NewUserID(userID))
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
		ID:            rt.ID().Value(),
		Type:          string(rt.Type()),
		CategoryID:    rt.CategoryID().Value(),
		Amount:        rt.Amount().Amount(),
		Description:   rt.Description(),
		Frequency:     string(rt.Frequency()),
		Weekday:       rt.Weekday(),
		DayOfMonth:    rt.DayOfMonth(),
		MonthOfYear:   rt.MonthOfYear(),
		ScheduleLabel: rt.ScheduleLabel(),
		NextDueDate:   due,
		NextDate:      due,
		IsActive:      rt.IsActive(),
	}
	if category != nil {
		resp.Category = buildCategoryResponse(category)
	}
	return resp
}
