package finance

import (
	"context"
	"errors"
	"fmt"
	"time"

	"panda-pocket/internal/domain/entitlement"
	"panda-pocket/internal/domain/finance"
)

type ReceivableResponse struct {
	ID                        int      `json:"id"`
	Name                      string   `json:"name"`
	Type                      string   `json:"type"`
	CurrencyID                int      `json:"currency_id"`
	CurrentBalance            float64  `json:"current_balance"`
	OriginalPrincipal         *float64 `json:"original_principal,omitempty"`
	NextDueDate               *string  `json:"next_due_date,omitempty"`
	Notes                     string   `json:"notes"`
	IsArchived                bool     `json:"is_archived"`
	AsOfDate                  *string  `json:"as_of_date,omitempty"`
	CreateExpenseID           *int     `json:"create_expense_id,omitempty"`
	CreatedAt                 string   `json:"created_at"`
	CollectionProgressPercent *float64 `json:"collection_progress_percent,omitempty"`
}

type CreateReceivableRequest struct {
	Name              string   `json:"name" binding:"required"`
	Type              string   `json:"type" binding:"required,oneof=personal_loan invoice other"`
	CurrencyID        int      `json:"currency_id" binding:"required"`
	CurrentBalance    float64  `json:"current_balance"`
	OriginalPrincipal *float64 `json:"original_principal"`
	NextDueDate       *string  `json:"next_due_date"`
	Notes             string   `json:"notes"`
	AsOfDate          *string  `json:"as_of_date"`
	CreateExpense     bool     `json:"create_expense"`
	CategoryID        *int     `json:"category_id"`
	WalletID          *int     `json:"wallet_id"`
}

type UpdateReceivableRequest struct {
	Name              string   `json:"name" binding:"required"`
	Type              string   `json:"type" binding:"required,oneof=personal_loan invoice other"`
	CurrentBalance    float64  `json:"current_balance"`
	OriginalPrincipal *float64 `json:"original_principal"`
	NextDueDate       *string  `json:"next_due_date"`
	Notes             string   `json:"notes"`
	AsOfDate          *string  `json:"as_of_date"`
}

type ReceivableCollectionResponse struct {
	ID           int     `json:"id"`
	ReceivableID int     `json:"receivable_id"`
	Amount       float64 `json:"amount"`
	CollectedAt  string  `json:"collected_at"`
	IncomeID     *int    `json:"income_id,omitempty"`
	Note         string  `json:"note"`
	CreatedAt    string  `json:"created_at"`
}

type RecordReceivableCollectionRequest struct {
	Amount       float64 `json:"amount" binding:"required,gt=0"`
	CollectedAt  string  `json:"collected_at" binding:"required"`
	Note         string  `json:"note"`
	CreateIncome bool    `json:"create_income"`
	CategoryID   *int    `json:"category_id"`
	WalletID     *int    `json:"wallet_id"`
}

func toReceivableResponse(receivable *finance.Receivable) ReceivableResponse {
	return ReceivableResponse{
		ID:                        receivable.ID().Value(),
		Name:                      receivable.Name(),
		Type:                      string(receivable.Type()),
		CurrencyID:                receivable.CurrencyID().Value(),
		CurrentBalance:            receivable.CurrentBalance(),
		OriginalPrincipal:         receivable.OriginalPrincipal(),
		NextDueDate:               formatOptionalDate(receivable.NextDueDate()),
		Notes:                     receivable.Notes(),
		IsArchived:                receivable.IsArchived(),
		AsOfDate:                  formatOptionalDate(receivable.AsOfDate()),
		CreateExpenseID:           receivable.CreateExpenseID(),
		CreatedAt:                 receivable.CreatedAt().Format(time.RFC3339),
		CollectionProgressPercent: receivable.CollectionProgressPercent(),
	}
}

func toReceivableCollectionResponse(collection *finance.ReceivableCollection) ReceivableCollectionResponse {
	return ReceivableCollectionResponse{
		ID:           collection.ID().Value(),
		ReceivableID: collection.ReceivableID().Value(),
		Amount:       collection.Amount(),
		CollectedAt:  collection.CollectedAt().Format("2006-01-02"),
		IncomeID:     collection.IncomeID(),
		Note:         collection.Note(),
		CreatedAt:    collection.CreatedAt().Format(time.RFC3339),
	}
}

func receivableDetailsFromRequest(original *float64, nextDue *string) (finance.ReceivableDetails, error) {
	due, err := parseOptionalDate(nextDue)
	if err != nil {
		return finance.ReceivableDetails{}, err
	}
	return finance.ReceivableDetails{
		OriginalPrincipal: original,
		NextDueDate:       due,
	}, nil
}

type CreateReceivableUseCase struct {
	receivableService        *finance.ReceivableService
	createTransactionUseCase *CreateTransactionUseCase
	categoryService          *finance.CategoryService
	entitlement              entitlement.Checker
}

func NewCreateReceivableUseCase(
	s *finance.ReceivableService,
	createTransactionUseCase *CreateTransactionUseCase,
	categoryService *finance.CategoryService,
	checker entitlement.Checker,
) *CreateReceivableUseCase {
	return &CreateReceivableUseCase{
		receivableService:        s,
		createTransactionUseCase: createTransactionUseCase,
		categoryService:          categoryService,
		entitlement:              checker,
	}
}

func (uc *CreateReceivableUseCase) resolveReceivableExpenseCategoryID(ctx context.Context, userID int, override *int) (int, error) {
	if override != nil && *override > 0 {
		return *override, nil
	}
	categories, err := uc.categoryService.GetCategoriesByUserAndType(ctx, finance.NewUserID(userID), finance.CategoryTypeExpense)
	if err != nil {
		return 0, err
	}
	for _, category := range categories {
		if category.Name() == "Receivable" {
			return category.ID().Value(), nil
		}
	}
	for _, category := range categories {
		if category.Name() == "Other" || category.Name() == "Debt" {
			return category.ID().Value(), nil
		}
	}
	if len(categories) == 0 {
		return 0, errors.New("no expense category available for receivable")
	}
	return categories[0].ID().Value(), nil
}

func (uc *CreateReceivableUseCase) Execute(ctx context.Context, userID int, req CreateReceivableRequest) (*ReceivableResponse, error) {
	existing, err := uc.receivableService.List(ctx, finance.NewUserID(userID), false)
	if err != nil {
		return nil, err
	}
	if err := entitlement.EnforceCreateLimit(
		ctx, uc.entitlement, userID,
		entitlement.FeatureReceivables, len(existing), entitlement.FreeReceivables,
	); err != nil {
		return nil, err
	}

	receivableType, err := finance.ParseReceivableType(req.Type)
	if err != nil {
		return nil, err
	}
	asOf, err := parseOptionalDate(req.AsOfDate)
	if err != nil {
		return nil, err
	}
	details, err := receivableDetailsFromRequest(req.OriginalPrincipal, req.NextDueDate)
	if err != nil {
		return nil, err
	}
	if details.OriginalPrincipal == nil && req.CurrentBalance > 0 {
		principal := req.CurrentBalance
		details.OriginalPrincipal = &principal
	}

	receivable, err := uc.receivableService.Create(
		ctx,
		finance.NewUserID(userID),
		req.Name,
		receivableType,
		finance.NewCurrencyID(req.CurrencyID),
		req.CurrentBalance,
		req.Notes,
		asOf,
		details,
	)
	if err != nil {
		return nil, err
	}

	if req.CreateExpense && req.CurrentBalance > 0 {
		if uc.createTransactionUseCase == nil || uc.categoryService == nil {
			return nil, errors.New("expense creation is not available")
		}
		categoryID, catErr := uc.resolveReceivableExpenseCategoryID(ctx, userID, req.CategoryID)
		if catErr != nil {
			return nil, catErr
		}
		description := req.Notes
		if description == "" {
			description = fmt.Sprintf("Receivable lend (%s)", receivable.Name())
		}
		date := time.Now().Format("2006-01-02")
		if asOf != nil {
			date = asOf.Format("2006-01-02")
		}
		tx, txErr := uc.createTransactionUseCase.Execute(ctx, userID, CreateTransactionRequest{
			CategoryID:  categoryID,
			WalletID:    req.WalletID,
			Amount:      req.CurrentBalance,
			Description: description,
			Date:        date,
			Type:        "expense",
		})
		if txErr != nil {
			return nil, txErr
		}
		if tx.ID == 0 {
			return nil, errors.New("created expense is missing id")
		}
		receivable, err = uc.receivableService.AttachCreateExpenseID(
			ctx,
			finance.NewUserID(userID),
			receivable.ID(),
			tx.ID,
		)
		if err != nil {
			return nil, err
		}
	}

	resp := toReceivableResponse(receivable)
	return &resp, nil
}

type GetReceivablesUseCase struct{ receivableService *finance.ReceivableService }

func NewGetReceivablesUseCase(s *finance.ReceivableService) *GetReceivablesUseCase {
	return &GetReceivablesUseCase{receivableService: s}
}

func (uc *GetReceivablesUseCase) Execute(ctx context.Context, userID int, includeArchived bool) ([]ReceivableResponse, error) {
	items, err := uc.receivableService.List(ctx, finance.NewUserID(userID), includeArchived)
	if err != nil {
		return nil, err
	}
	result := make([]ReceivableResponse, 0, len(items))
	for _, item := range items {
		result = append(result, toReceivableResponse(item))
	}
	return result, nil
}

type UpdateReceivableUseCase struct{ receivableService *finance.ReceivableService }

func NewUpdateReceivableUseCase(s *finance.ReceivableService) *UpdateReceivableUseCase {
	return &UpdateReceivableUseCase{receivableService: s}
}

func (uc *UpdateReceivableUseCase) Execute(ctx context.Context, userID, id int, req UpdateReceivableRequest) (*ReceivableResponse, error) {
	receivableType, err := finance.ParseReceivableType(req.Type)
	if err != nil {
		return nil, err
	}
	asOf, err := parseOptionalDate(req.AsOfDate)
	if err != nil {
		return nil, err
	}
	details, err := receivableDetailsFromRequest(req.OriginalPrincipal, req.NextDueDate)
	if err != nil {
		return nil, err
	}
	receivable, err := uc.receivableService.Update(
		ctx,
		finance.NewUserID(userID),
		finance.NewReceivableID(id),
		req.Name,
		receivableType,
		req.CurrentBalance,
		req.Notes,
		asOf,
		details,
	)
	if err != nil {
		return nil, err
	}
	resp := toReceivableResponse(receivable)
	return &resp, nil
}

type ArchiveReceivableUseCase struct{ receivableService *finance.ReceivableService }

func NewArchiveReceivableUseCase(s *finance.ReceivableService) *ArchiveReceivableUseCase {
	return &ArchiveReceivableUseCase{receivableService: s}
}

func (uc *ArchiveReceivableUseCase) Execute(ctx context.Context, userID, id int) (*ReceivableResponse, error) {
	receivable, err := uc.receivableService.Archive(ctx, finance.NewUserID(userID), finance.NewReceivableID(id))
	if err != nil {
		return nil, err
	}
	resp := toReceivableResponse(receivable)
	return &resp, nil
}

type UnarchiveReceivableUseCase struct{ receivableService *finance.ReceivableService }

func NewUnarchiveReceivableUseCase(s *finance.ReceivableService) *UnarchiveReceivableUseCase {
	return &UnarchiveReceivableUseCase{receivableService: s}
}

func (uc *UnarchiveReceivableUseCase) Execute(ctx context.Context, userID, id int) (*ReceivableResponse, error) {
	receivable, err := uc.receivableService.Unarchive(ctx, finance.NewUserID(userID), finance.NewReceivableID(id))
	if err != nil {
		return nil, err
	}
	resp := toReceivableResponse(receivable)
	return &resp, nil
}

type ListReceivableCollectionsUseCase struct {
	receivableService *finance.ReceivableService
}

func NewListReceivableCollectionsUseCase(s *finance.ReceivableService) *ListReceivableCollectionsUseCase {
	return &ListReceivableCollectionsUseCase{receivableService: s}
}

func (uc *ListReceivableCollectionsUseCase) Execute(ctx context.Context, userID, receivableID int) ([]ReceivableCollectionResponse, error) {
	collections, err := uc.receivableService.ListCollections(
		ctx,
		finance.NewUserID(userID),
		finance.NewReceivableID(receivableID),
	)
	if err != nil {
		return nil, err
	}
	result := make([]ReceivableCollectionResponse, 0, len(collections))
	for _, collection := range collections {
		result = append(result, toReceivableCollectionResponse(collection))
	}
	return result, nil
}

type RecordReceivableCollectionUseCase struct {
	receivableService        *finance.ReceivableService
	createTransactionUseCase *CreateTransactionUseCase
	categoryService          *finance.CategoryService
}

func NewRecordReceivableCollectionUseCase(
	receivableService *finance.ReceivableService,
	createTransactionUseCase *CreateTransactionUseCase,
	categoryService *finance.CategoryService,
) *RecordReceivableCollectionUseCase {
	return &RecordReceivableCollectionUseCase{
		receivableService:        receivableService,
		createTransactionUseCase: createTransactionUseCase,
		categoryService:          categoryService,
	}
}

type RecordReceivableCollectionResponse struct {
	Receivable *ReceivableResponse           `json:"receivable"`
	Collection *ReceivableCollectionResponse `json:"collection"`
}

func (uc *RecordReceivableCollectionUseCase) resolveReceivableIncomeCategoryID(ctx context.Context, userID int, override *int) (int, error) {
	if override != nil && *override > 0 {
		return *override, nil
	}
	categories, err := uc.categoryService.GetCategoriesByUserAndType(ctx, finance.NewUserID(userID), finance.CategoryTypeIncome)
	if err != nil {
		return 0, err
	}
	for _, category := range categories {
		if category.Name() == "Receivable" {
			return category.ID().Value(), nil
		}
	}
	for _, category := range categories {
		if category.Name() == "Other" {
			return category.ID().Value(), nil
		}
	}
	if len(categories) == 0 {
		return 0, errors.New("no income category available for receivable collection")
	}
	return categories[0].ID().Value(), nil
}

func (uc *RecordReceivableCollectionUseCase) Execute(
	ctx context.Context,
	userID int,
	receivableID int,
	req RecordReceivableCollectionRequest,
) (*RecordReceivableCollectionResponse, error) {
	collectedAt, err := time.Parse("2006-01-02", req.CollectedAt)
	if err != nil {
		return nil, errors.New("invalid collected_at format. Expected YYYY-MM-DD")
	}

	var incomeID *int
	if req.CreateIncome {
		if uc.createTransactionUseCase == nil || uc.categoryService == nil {
			return nil, errors.New("income creation is not available")
		}
		categoryID, catErr := uc.resolveReceivableIncomeCategoryID(ctx, userID, req.CategoryID)
		if catErr != nil {
			return nil, catErr
		}
		receivableForIncome, getErr := uc.receivableService.GetForUser(
			ctx,
			finance.NewUserID(userID),
			finance.NewReceivableID(receivableID),
		)
		if getErr != nil {
			return nil, getErr
		}
		description := req.Note
		if description == "" {
			description = fmt.Sprintf("Receivable collection (%s)", receivableForIncome.Name())
		}
		tx, txErr := uc.createTransactionUseCase.Execute(ctx, userID, CreateTransactionRequest{
			CategoryID:  categoryID,
			WalletID:    req.WalletID,
			Amount:      req.Amount,
			Description: description,
			Date:        req.CollectedAt,
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
	}

	receivable, collection, err := uc.receivableService.RecordCollection(
		ctx,
		finance.NewUserID(userID),
		finance.NewReceivableID(receivableID),
		req.Amount,
		collectedAt,
		incomeID,
		req.Note,
	)
	if err != nil {
		return nil, err
	}
	receivableResp := toReceivableResponse(receivable)
	collectionResp := toReceivableCollectionResponse(collection)
	return &RecordReceivableCollectionResponse{
		Receivable: &receivableResp,
		Collection: &collectionResp,
	}, nil
}
