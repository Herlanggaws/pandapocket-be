package finance

import (
	"context"
	"errors"
	"fmt"
	appNotification "panda-pocket/internal/application/notification"
	domainFinance "panda-pocket/internal/domain/finance"
	domainIdentity "panda-pocket/internal/domain/identity"
)

type PendingTransactionResponse struct {
	ID                     int               `json:"id"`
	UserID                 int               `json:"user_id"`
	WalletID               int               `json:"wallet_id"`
	CurrencyID             int               `json:"currency_id"`
	RecurringTransactionID int               `json:"recurring_transaction_id"`
	DueDate                string            `json:"due_date"`
	Amount                 float64           `json:"amount"`
	Description            string            `json:"description"`
	Type                   string            `json:"type"`
	CategoryID             int               `json:"category_id"`
	Status                 string            `json:"status"`
	Category               *CategoryResponse `json:"category,omitempty"`
	CreatedAt              string            `json:"created_at"`
}

type EnqueueDueRecurringUseCase struct {
	recurringRepo      domainFinance.RecurringTransactionRepository
	pendingRepo        domainFinance.PendingTransactionRepository
	prefsRepo          domainIdentity.PreferencesRepository
	notificationHelper *appNotification.CreateNotificationHelper
}

func NewEnqueueDueRecurringUseCase(
	recurringRepo domainFinance.RecurringTransactionRepository,
	pendingRepo domainFinance.PendingTransactionRepository,
	prefsRepo domainIdentity.PreferencesRepository,
	notificationHelper *appNotification.CreateNotificationHelper,
) *EnqueueDueRecurringUseCase {
	return &EnqueueDueRecurringUseCase{
		recurringRepo:      recurringRepo,
		pendingRepo:        pendingRepo,
		prefsRepo:          prefsRepo,
		notificationHelper: notificationHelper,
	}
}

func (uc *EnqueueDueRecurringUseCase) Execute(ctx context.Context, userID int) error {
	items, err := uc.recurringRepo.FindByUserID(ctx, domainFinance.NewUserID(userID))
	if err != nil {
		return err
	}

	remindersEnabled := true
	if prefs, err := uc.prefsRepo.FindByUserID(ctx, domainIdentity.NewUserID(userID)); err == nil && prefs != nil {
		remindersEnabled = prefs.RecurringReminders()
	}

	for _, rt := range items {
		for rt.IsDue() {
			dueDate := rt.NextDueDate()

			exists, err := uc.pendingRepo.ExistsByRecurringAndDueDate(ctx, rt.ID(), dueDate)
			if err != nil {
				return err
			}
			if !exists {
				pending, err := domainFinance.NewPendingTransaction(
					rt.UserID(),
					rt.WalletID(),
					rt.ID(),
					dueDate,
					rt.Amount(),
					rt.Description(),
					rt.Type(),
					rt.CategoryID(),
					rt.CurrencyID(),
				)
				if err != nil {
					break
				}
				if err := uc.pendingRepo.Save(ctx, pending); err != nil {
					// Unique race: still advance so we do not loop forever
					existsAfter, existsErr := uc.pendingRepo.ExistsByRecurringAndDueDate(ctx, rt.ID(), dueDate)
					if existsErr != nil || !existsAfter {
						break
					}
				} else if remindersEnabled && uc.notificationHelper != nil {
					title := "Recurring transaction awaiting confirmation"
					message := fmt.Sprintf("%s (%s) needs confirmation.", rt.Description(), rt.Type())
					if rt.Description() == "" {
						message = fmt.Sprintf("A recurring %s needs confirmation.", rt.Type())
					}
					_ = uc.notificationHelper.CreateIfNotRecent(ctx, userID, title, message, "recurring_reminder")
				}
			}

			if err := rt.AdvanceNextDue(); err != nil {
				break
			}
			if err := uc.recurringRepo.Save(ctx, rt); err != nil {
				break
			}
			if !rt.NextDueDate().After(dueDate) {
				break
			}
		}
	}
	return nil
}

type ListPendingTransactionsUseCase struct {
	enqueueUseCase  *EnqueueDueRecurringUseCase
	pendingRepo     domainFinance.PendingTransactionRepository
	categoryService *domainFinance.CategoryService
}

func NewListPendingTransactionsUseCase(
	enqueueUseCase *EnqueueDueRecurringUseCase,
	pendingRepo domainFinance.PendingTransactionRepository,
	categoryService *domainFinance.CategoryService,
) *ListPendingTransactionsUseCase {
	return &ListPendingTransactionsUseCase{
		enqueueUseCase:  enqueueUseCase,
		pendingRepo:     pendingRepo,
		categoryService: categoryService,
	}
}

func (uc *ListPendingTransactionsUseCase) Execute(ctx context.Context, userID int) ([]PendingTransactionResponse, error) {
	if err := uc.enqueueUseCase.Execute(ctx, userID); err != nil {
		return nil, err
	}

	items, err := uc.pendingRepo.FindOpenByUserID(ctx, domainFinance.NewUserID(userID))
	if err != nil {
		return nil, err
	}

	result := make([]PendingTransactionResponse, 0, len(items))
	for _, pt := range items {
		var category *domainFinance.Category
		if cat, err := uc.categoryService.GetCategoryByID(ctx, pt.CategoryID()); err == nil {
			category = cat
		}
		result = append(result, toPendingResponse(pt, category))
	}
	return result, nil
}

type ConfirmPendingTransactionUseCase struct {
	pendingRepo        domainFinance.PendingTransactionRepository
	transactionService *domainFinance.TransactionService
}

func NewConfirmPendingTransactionUseCase(
	pendingRepo domainFinance.PendingTransactionRepository,
	transactionService *domainFinance.TransactionService,
) *ConfirmPendingTransactionUseCase {
	return &ConfirmPendingTransactionUseCase{
		pendingRepo:        pendingRepo,
		transactionService: transactionService,
	}
}

func (uc *ConfirmPendingTransactionUseCase) Execute(ctx context.Context, userID, id int) (*PendingTransactionResponse, error) {
	pending, err := uc.pendingRepo.FindByID(ctx, domainFinance.NewPendingTransactionID(id))
	if err != nil {
		return nil, errors.New("pending transaction not found")
	}
	if pending.UserID().Value() != userID {
		return nil, errors.New("pending transaction not found")
	}
	if !pending.IsOpen() {
		return nil, errors.New("pending transaction is already resolved")
	}

	_, err = uc.transactionService.CreateTransaction(
		ctx,
		pending.UserID(),
		pending.WalletID(),
		pending.CategoryID(),
		pending.CurrencyID(),
		pending.Amount(),
		pending.Description(),
		pending.DueDate(),
		pending.Type(),
	)
	if err != nil {
		return nil, err
	}

	if err := pending.Confirm(); err != nil {
		return nil, err
	}
	if err := uc.pendingRepo.Save(ctx, pending); err != nil {
		return nil, err
	}

	resp := toPendingResponse(pending, nil)
	return &resp, nil
}

type RejectPendingTransactionUseCase struct {
	pendingRepo domainFinance.PendingTransactionRepository
}

func NewRejectPendingTransactionUseCase(
	pendingRepo domainFinance.PendingTransactionRepository,
) *RejectPendingTransactionUseCase {
	return &RejectPendingTransactionUseCase{pendingRepo: pendingRepo}
}

func (uc *RejectPendingTransactionUseCase) Execute(ctx context.Context, userID, id int) (*PendingTransactionResponse, error) {
	pending, err := uc.pendingRepo.FindByID(ctx, domainFinance.NewPendingTransactionID(id))
	if err != nil {
		return nil, errors.New("pending transaction not found")
	}
	if pending.UserID().Value() != userID {
		return nil, errors.New("pending transaction not found")
	}
	if !pending.IsOpen() {
		return nil, errors.New("pending transaction is already resolved")
	}

	if err := pending.Reject(); err != nil {
		return nil, err
	}
	if err := uc.pendingRepo.Save(ctx, pending); err != nil {
		return nil, err
	}

	resp := toPendingResponse(pending, nil)
	return &resp, nil
}

func toPendingResponse(pt *domainFinance.PendingTransaction, category *domainFinance.Category) PendingTransactionResponse {
	return PendingTransactionResponse{
		ID:                     pt.ID().Value(),
		UserID:                 pt.UserID().Value(),
		WalletID:               pt.WalletID().Value(),
		CurrencyID:             pt.CurrencyID().Value(),
		RecurringTransactionID: pt.RecurringTransactionID().Value(),
		DueDate:                pt.DueDate().Format("2006-01-02"),
		Amount:                 pt.Amount().Amount(),
		Description:            pt.Description(),
		Type:                   string(pt.Type()),
		CategoryID:             pt.CategoryID().Value(),
		Status:                 string(pt.Status()),
		Category:               buildCategoryResponse(category),
		CreatedAt:              pt.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
}
