package finance

import (
	"context"
	"errors"
	"time"

	"panda-pocket/internal/domain/entitlement"
	"panda-pocket/internal/domain/finance"
)

type AssetResponse struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Type         string  `json:"type"`
	CurrencyID   int     `json:"currency_id"`
	CurrentValue float64 `json:"current_value"`
	Notes        string  `json:"notes"`
	IsArchived   bool    `json:"is_archived"`
	AsOfDate     *string `json:"as_of_date,omitempty"`
	CreatedAt    string  `json:"created_at"`
}

type LiabilityResponse struct {
	ID                       int      `json:"id"`
	Name                     string   `json:"name"`
	Type                     string   `json:"type"`
	CurrencyID               int      `json:"currency_id"`
	CurrentBalance           float64  `json:"current_balance"`
	OriginalPrincipal        *float64 `json:"original_principal,omitempty"`
	InterestRateAPR          *float64 `json:"interest_rate_apr,omitempty"`
	MinimumPayment           float64  `json:"minimum_payment"`
	NextDueDate              *string  `json:"next_due_date,omitempty"`
	Notes                    string   `json:"notes"`
	IsArchived               bool     `json:"is_archived"`
	AsOfDate                 *string  `json:"as_of_date,omitempty"`
	CreatedAt                string   `json:"created_at"`
	PayoffProgressPercent    *float64 `json:"payoff_progress_percent,omitempty"`
	EstimatedMonthsRemaining *int     `json:"estimated_months_remaining,omitempty"`
}

type CreateAssetRequest struct {
	Name         string  `json:"name" binding:"required"`
	Type         string  `json:"type" binding:"required,oneof=property vehicle investment other"`
	CurrencyID   int     `json:"currency_id" binding:"required"`
	CurrentValue float64 `json:"current_value"`
	Notes        string  `json:"notes"`
	AsOfDate     *string `json:"as_of_date"`
}

type UpdateAssetRequest struct {
	Name         string  `json:"name" binding:"required"`
	Type         string  `json:"type" binding:"required,oneof=property vehicle investment other"`
	CurrentValue float64 `json:"current_value"`
	Notes        string  `json:"notes"`
	AsOfDate     *string `json:"as_of_date"`
}

type CreateLiabilityRequest struct {
	Name              string   `json:"name" binding:"required"`
	Type              string   `json:"type" binding:"required,oneof=loan credit_card mortgage other"`
	CurrencyID        int      `json:"currency_id" binding:"required"`
	CurrentBalance    float64  `json:"current_balance"`
	OriginalPrincipal *float64 `json:"original_principal"`
	InterestRateAPR   *float64 `json:"interest_rate_apr"`
	MinimumPayment    float64  `json:"minimum_payment"`
	NextDueDate       *string  `json:"next_due_date"`
	Notes             string   `json:"notes"`
	AsOfDate          *string  `json:"as_of_date"`
}

type UpdateLiabilityRequest struct {
	Name              string   `json:"name" binding:"required"`
	Type              string   `json:"type" binding:"required,oneof=loan credit_card mortgage other"`
	CurrentBalance    float64  `json:"current_balance"`
	OriginalPrincipal *float64 `json:"original_principal"`
	InterestRateAPR   *float64 `json:"interest_rate_apr"`
	MinimumPayment    float64  `json:"minimum_payment"`
	NextDueDate       *string  `json:"next_due_date"`
	Notes             string   `json:"notes"`
	AsOfDate          *string  `json:"as_of_date"`
}

type LiabilityPaymentResponse struct {
	ID          int     `json:"id"`
	LiabilityID int     `json:"liability_id"`
	Amount      float64 `json:"amount"`
	PaidAt      string  `json:"paid_at"`
	ExpenseID   *int    `json:"expense_id,omitempty"`
	Note        string  `json:"note"`
	CreatedAt   string  `json:"created_at"`
}

type RecordLiabilityPaymentRequest struct {
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	PaidAt        string  `json:"paid_at" binding:"required"`
	Note          string  `json:"note"`
	CreateExpense bool    `json:"create_expense"`
	CategoryID    *int    `json:"category_id"`
	WalletID      *int    `json:"wallet_id"`
}

type NetWorthSummaryResponse struct {
	CurrencyID             int     `json:"currency_id"`
	LiquidNetWorth         float64 `json:"liquid_net_worth"`
	AssetsTotal            float64 `json:"assets_total"`
	LiabilitiesTotal       float64 `json:"liabilities_total"`
	NetWorth               float64 `json:"net_worth"`
	ExcludedAssetCount     int     `json:"excluded_asset_count"`
	ExcludedLiabilityCount int     `json:"excluded_liability_count"`
	ExcludedWalletCount    int     `json:"excluded_wallet_count"`
}

func parseOptionalDate(value *string) (*time.Time, error) {
	if value == nil || *value == "" {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02", *value)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func formatOptionalDate(value *time.Time) *string {
	if value == nil {
		return nil
	}
	s := value.Format("2006-01-02")
	return &s
}

func toAssetResponse(asset *finance.Asset) AssetResponse {
	return AssetResponse{
		ID:           asset.ID().Value(),
		Name:         asset.Name(),
		Type:         string(asset.Type()),
		CurrencyID:   asset.CurrencyID().Value(),
		CurrentValue: asset.CurrentValue(),
		Notes:        asset.Notes(),
		IsArchived:   asset.IsArchived(),
		AsOfDate:     formatOptionalDate(asset.AsOfDate()),
		CreatedAt:    asset.CreatedAt().Format(time.RFC3339),
	}
}

func toLiabilityResponse(liability *finance.Liability) LiabilityResponse {
	return LiabilityResponse{
		ID:                       liability.ID().Value(),
		Name:                     liability.Name(),
		Type:                     string(liability.Type()),
		CurrencyID:               liability.CurrencyID().Value(),
		CurrentBalance:           liability.CurrentBalance(),
		OriginalPrincipal:        liability.OriginalPrincipal(),
		InterestRateAPR:          liability.InterestRateAPR(),
		MinimumPayment:           liability.MinimumPayment(),
		NextDueDate:              formatOptionalDate(liability.NextDueDate()),
		Notes:                    liability.Notes(),
		IsArchived:               liability.IsArchived(),
		AsOfDate:                 formatOptionalDate(liability.AsOfDate()),
		CreatedAt:                liability.CreatedAt().Format(time.RFC3339),
		PayoffProgressPercent:    liability.PayoffProgressPercent(),
		EstimatedMonthsRemaining: liability.EstimatedMonthsRemaining(),
	}
}

func toLiabilityPaymentResponse(payment *finance.LiabilityPayment) LiabilityPaymentResponse {
	return LiabilityPaymentResponse{
		ID:          payment.ID().Value(),
		LiabilityID: payment.LiabilityID().Value(),
		Amount:      payment.Amount(),
		PaidAt:      payment.PaidAt().Format("2006-01-02"),
		ExpenseID:   payment.ExpenseID(),
		Note:        payment.Note(),
		CreatedAt:   payment.CreatedAt().Format(time.RFC3339),
	}
}

func debtDetailsFromRequest(original *float64, apr *float64, minimum float64, nextDue *string) (finance.LiabilityDebtDetails, error) {
	due, err := parseOptionalDate(nextDue)
	if err != nil {
		return finance.LiabilityDebtDetails{}, err
	}
	return finance.LiabilityDebtDetails{
		OriginalPrincipal: original,
		InterestRateAPR:   apr,
		MinimumPayment:    minimum,
		NextDueDate:       due,
	}, nil
}

type CreateAssetUseCase struct {
	assetService *finance.AssetService
	entitlement  entitlement.Checker
}

func NewCreateAssetUseCase(s *finance.AssetService, checker entitlement.Checker) *CreateAssetUseCase {
	return &CreateAssetUseCase{assetService: s, entitlement: checker}
}

func (uc *CreateAssetUseCase) Execute(ctx context.Context, userID int, req CreateAssetRequest) (*AssetResponse, error) {
	existing, err := uc.assetService.List(ctx, finance.NewUserID(userID), false)
	if err != nil {
		return nil, err
	}
	if err := entitlement.EnforceCreateLimit(
		ctx, uc.entitlement, userID,
		entitlement.FeatureAssets, len(existing), entitlement.FreeAssets,
	); err != nil {
		return nil, err
	}

	assetType, err := finance.ParseAssetType(req.Type)
	if err != nil {
		return nil, err
	}
	asOf, err := parseOptionalDate(req.AsOfDate)
	if err != nil {
		return nil, err
	}
	asset, err := uc.assetService.Create(
		ctx,
		finance.NewUserID(userID),
		req.Name,
		assetType,
		finance.NewCurrencyID(req.CurrencyID),
		req.CurrentValue,
		req.Notes,
		asOf,
	)
	if err != nil {
		return nil, err
	}
	resp := toAssetResponse(asset)
	return &resp, nil
}

type GetAssetsUseCase struct{ assetService *finance.AssetService }

func NewGetAssetsUseCase(s *finance.AssetService) *GetAssetsUseCase {
	return &GetAssetsUseCase{assetService: s}
}

func (uc *GetAssetsUseCase) Execute(ctx context.Context, userID int, includeArchived bool) ([]AssetResponse, error) {
	assets, err := uc.assetService.List(ctx, finance.NewUserID(userID), includeArchived)
	if err != nil {
		return nil, err
	}
	result := make([]AssetResponse, 0, len(assets))
	for _, a := range assets {
		result = append(result, toAssetResponse(a))
	}
	return result, nil
}

type UpdateAssetUseCase struct{ assetService *finance.AssetService }

func NewUpdateAssetUseCase(s *finance.AssetService) *UpdateAssetUseCase {
	return &UpdateAssetUseCase{assetService: s}
}

func (uc *UpdateAssetUseCase) Execute(ctx context.Context, userID, id int, req UpdateAssetRequest) (*AssetResponse, error) {
	assetType, err := finance.ParseAssetType(req.Type)
	if err != nil {
		return nil, err
	}
	asOf, err := parseOptionalDate(req.AsOfDate)
	if err != nil {
		return nil, err
	}
	asset, err := uc.assetService.Update(
		ctx,
		finance.NewUserID(userID),
		finance.NewAssetID(id),
		req.Name,
		assetType,
		req.CurrentValue,
		req.Notes,
		asOf,
	)
	if err != nil {
		return nil, err
	}
	resp := toAssetResponse(asset)
	return &resp, nil
}

type ArchiveAssetUseCase struct{ assetService *finance.AssetService }

func NewArchiveAssetUseCase(s *finance.AssetService) *ArchiveAssetUseCase {
	return &ArchiveAssetUseCase{assetService: s}
}

func (uc *ArchiveAssetUseCase) Execute(ctx context.Context, userID, id int) (*AssetResponse, error) {
	asset, err := uc.assetService.Archive(ctx, finance.NewUserID(userID), finance.NewAssetID(id))
	if err != nil {
		return nil, err
	}
	resp := toAssetResponse(asset)
	return &resp, nil
}

type UnarchiveAssetUseCase struct{ assetService *finance.AssetService }

func NewUnarchiveAssetUseCase(s *finance.AssetService) *UnarchiveAssetUseCase {
	return &UnarchiveAssetUseCase{assetService: s}
}

func (uc *UnarchiveAssetUseCase) Execute(ctx context.Context, userID, id int) (*AssetResponse, error) {
	asset, err := uc.assetService.Unarchive(ctx, finance.NewUserID(userID), finance.NewAssetID(id))
	if err != nil {
		return nil, err
	}
	resp := toAssetResponse(asset)
	return &resp, nil
}

type CreateLiabilityUseCase struct {
	liabilityService *finance.LiabilityService
	entitlement      entitlement.Checker
}

func NewCreateLiabilityUseCase(s *finance.LiabilityService, checker entitlement.Checker) *CreateLiabilityUseCase {
	return &CreateLiabilityUseCase{liabilityService: s, entitlement: checker}
}

func (uc *CreateLiabilityUseCase) Execute(ctx context.Context, userID int, req CreateLiabilityRequest) (*LiabilityResponse, error) {
	existing, err := uc.liabilityService.List(ctx, finance.NewUserID(userID), false)
	if err != nil {
		return nil, err
	}
	if err := entitlement.EnforceCreateLimit(
		ctx, uc.entitlement, userID,
		entitlement.FeatureDebts, len(existing), entitlement.FreeDebts,
	); err != nil {
		return nil, err
	}

	liabilityType, err := finance.ParseLiabilityType(req.Type)
	if err != nil {
		return nil, err
	}
	asOf, err := parseOptionalDate(req.AsOfDate)
	if err != nil {
		return nil, err
	}
	debt, err := debtDetailsFromRequest(req.OriginalPrincipal, req.InterestRateAPR, req.MinimumPayment, req.NextDueDate)
	if err != nil {
		return nil, err
	}
	if debt.OriginalPrincipal == nil && req.CurrentBalance > 0 {
		principal := req.CurrentBalance
		debt.OriginalPrincipal = &principal
	}
	liability, err := uc.liabilityService.Create(
		ctx,
		finance.NewUserID(userID),
		req.Name,
		liabilityType,
		finance.NewCurrencyID(req.CurrencyID),
		req.CurrentBalance,
		req.Notes,
		asOf,
		debt,
	)
	if err != nil {
		return nil, err
	}
	resp := toLiabilityResponse(liability)
	return &resp, nil
}

type GetLiabilitiesUseCase struct{ liabilityService *finance.LiabilityService }

func NewGetLiabilitiesUseCase(s *finance.LiabilityService) *GetLiabilitiesUseCase {
	return &GetLiabilitiesUseCase{liabilityService: s}
}

func (uc *GetLiabilitiesUseCase) Execute(ctx context.Context, userID int, includeArchived bool) ([]LiabilityResponse, error) {
	items, err := uc.liabilityService.List(ctx, finance.NewUserID(userID), includeArchived)
	if err != nil {
		return nil, err
	}
	result := make([]LiabilityResponse, 0, len(items))
	for _, item := range items {
		result = append(result, toLiabilityResponse(item))
	}
	return result, nil
}

type UpdateLiabilityUseCase struct{ liabilityService *finance.LiabilityService }

func NewUpdateLiabilityUseCase(s *finance.LiabilityService) *UpdateLiabilityUseCase {
	return &UpdateLiabilityUseCase{liabilityService: s}
}

func (uc *UpdateLiabilityUseCase) Execute(ctx context.Context, userID, id int, req UpdateLiabilityRequest) (*LiabilityResponse, error) {
	liabilityType, err := finance.ParseLiabilityType(req.Type)
	if err != nil {
		return nil, err
	}
	asOf, err := parseOptionalDate(req.AsOfDate)
	if err != nil {
		return nil, err
	}
	debt, err := debtDetailsFromRequest(req.OriginalPrincipal, req.InterestRateAPR, req.MinimumPayment, req.NextDueDate)
	if err != nil {
		return nil, err
	}
	liability, err := uc.liabilityService.Update(
		ctx,
		finance.NewUserID(userID),
		finance.NewLiabilityID(id),
		req.Name,
		liabilityType,
		req.CurrentBalance,
		req.Notes,
		asOf,
		debt,
	)
	if err != nil {
		return nil, err
	}
	resp := toLiabilityResponse(liability)
	return &resp, nil
}

type ArchiveLiabilityUseCase struct{ liabilityService *finance.LiabilityService }

func NewArchiveLiabilityUseCase(s *finance.LiabilityService) *ArchiveLiabilityUseCase {
	return &ArchiveLiabilityUseCase{liabilityService: s}
}

func (uc *ArchiveLiabilityUseCase) Execute(ctx context.Context, userID, id int) (*LiabilityResponse, error) {
	liability, err := uc.liabilityService.Archive(ctx, finance.NewUserID(userID), finance.NewLiabilityID(id))
	if err != nil {
		return nil, err
	}
	resp := toLiabilityResponse(liability)
	return &resp, nil
}

type UnarchiveLiabilityUseCase struct{ liabilityService *finance.LiabilityService }

func NewUnarchiveLiabilityUseCase(s *finance.LiabilityService) *UnarchiveLiabilityUseCase {
	return &UnarchiveLiabilityUseCase{liabilityService: s}
}

func (uc *UnarchiveLiabilityUseCase) Execute(ctx context.Context, userID, id int) (*LiabilityResponse, error) {
	liability, err := uc.liabilityService.Unarchive(ctx, finance.NewUserID(userID), finance.NewLiabilityID(id))
	if err != nil {
		return nil, err
	}
	resp := toLiabilityResponse(liability)
	return &resp, nil
}

type GetNetWorthSummaryUseCase struct {
	walletSummaryUseCase *GetWalletSummaryUseCase
	assetService         *finance.AssetService
	liabilityService     *finance.LiabilityService
	currencyService      *finance.CurrencyService
}

func NewGetNetWorthSummaryUseCase(
	walletSummaryUseCase *GetWalletSummaryUseCase,
	assetService *finance.AssetService,
	liabilityService *finance.LiabilityService,
	currencyService *finance.CurrencyService,
) *GetNetWorthSummaryUseCase {
	return &GetNetWorthSummaryUseCase{
		walletSummaryUseCase: walletSummaryUseCase,
		assetService:         assetService,
		liabilityService:     liabilityService,
		currencyService:      currencyService,
	}
}

func (uc *GetNetWorthSummaryUseCase) Execute(ctx context.Context, userID int) (*NetWorthSummaryResponse, error) {
	walletSummary, err := uc.walletSummaryUseCase.Execute(ctx, userID)
	if err != nil {
		return nil, err
	}
	primaryID := walletSummary.CurrencyID

	assets, err := uc.assetService.List(ctx, finance.NewUserID(userID), false)
	if err != nil {
		return nil, err
	}
	liabilities, err := uc.liabilityService.List(ctx, finance.NewUserID(userID), false)
	if err != nil {
		return nil, err
	}

	var assetsTotal float64
	excludedAssets := 0
	for _, asset := range assets {
		if asset.CurrencyID().Value() != primaryID {
			excludedAssets++
			continue
		}
		assetsTotal += asset.CurrentValue()
	}

	var liabilitiesTotal float64
	excludedLiabilities := 0
	for _, liability := range liabilities {
		if liability.CurrencyID().Value() != primaryID {
			excludedLiabilities++
			continue
		}
		liabilitiesTotal += liability.CurrentBalance()
	}

	liquid := walletSummary.LiquidNetWorth
	return &NetWorthSummaryResponse{
		CurrencyID:             primaryID,
		LiquidNetWorth:         liquid,
		AssetsTotal:            assetsTotal,
		LiabilitiesTotal:       liabilitiesTotal,
		NetWorth:               liquid + assetsTotal - liabilitiesTotal,
		ExcludedAssetCount:     excludedAssets,
		ExcludedLiabilityCount: excludedLiabilities,
		ExcludedWalletCount:    walletSummary.ExcludedWalletCount,
	}, nil
}

type ListLiabilityPaymentsUseCase struct {
	liabilityService *finance.LiabilityService
}

func NewListLiabilityPaymentsUseCase(s *finance.LiabilityService) *ListLiabilityPaymentsUseCase {
	return &ListLiabilityPaymentsUseCase{liabilityService: s}
}

func (uc *ListLiabilityPaymentsUseCase) Execute(ctx context.Context, userID, liabilityID int) ([]LiabilityPaymentResponse, error) {
	payments, err := uc.liabilityService.ListPayments(ctx, finance.NewUserID(userID), finance.NewLiabilityID(liabilityID))
	if err != nil {
		return nil, err
	}
	result := make([]LiabilityPaymentResponse, 0, len(payments))
	for _, payment := range payments {
		result = append(result, toLiabilityPaymentResponse(payment))
	}
	return result, nil
}

type RecordLiabilityPaymentUseCase struct {
	liabilityService         *finance.LiabilityService
	createTransactionUseCase *CreateTransactionUseCase
	categoryService          *finance.CategoryService
}

func NewRecordLiabilityPaymentUseCase(
	liabilityService *finance.LiabilityService,
	createTransactionUseCase *CreateTransactionUseCase,
	categoryService *finance.CategoryService,
) *RecordLiabilityPaymentUseCase {
	return &RecordLiabilityPaymentUseCase{
		liabilityService:         liabilityService,
		createTransactionUseCase: createTransactionUseCase,
		categoryService:          categoryService,
	}
}

type RecordLiabilityPaymentResponse struct {
	Liability *LiabilityResponse        `json:"liability"`
	Payment   *LiabilityPaymentResponse `json:"payment"`
}

func (uc *RecordLiabilityPaymentUseCase) resolveDebtCategoryID(ctx context.Context, userID int, override *int) (int, error) {
	if override != nil && *override > 0 {
		return *override, nil
	}
	categories, err := uc.categoryService.GetCategoriesByUserAndType(ctx, finance.NewUserID(userID), finance.CategoryTypeExpense)
	if err != nil {
		return 0, err
	}
	for _, category := range categories {
		if category.Name() == "Debt" {
			return category.ID().Value(), nil
		}
	}
	for _, category := range categories {
		if category.Name() == "Bills" || category.Name() == "Other" {
			return category.ID().Value(), nil
		}
	}
	if len(categories) == 0 {
		return 0, errors.New("no expense category available for debt payment")
	}
	return categories[0].ID().Value(), nil
}

func (uc *RecordLiabilityPaymentUseCase) Execute(
	ctx context.Context,
	userID int,
	liabilityID int,
	req RecordLiabilityPaymentRequest,
) (*RecordLiabilityPaymentResponse, error) {
	paidAt, err := time.Parse("2006-01-02", req.PaidAt)
	if err != nil {
		return nil, errors.New("invalid paid_at format. Expected YYYY-MM-DD")
	}

	var expenseID *int
	if req.CreateExpense {
		if uc.createTransactionUseCase == nil || uc.categoryService == nil {
			return nil, errors.New("expense creation is not available")
		}
		categoryID, catErr := uc.resolveDebtCategoryID(ctx, userID, req.CategoryID)
		if catErr != nil {
			return nil, catErr
		}
		description := req.Note
		if description == "" {
			description = "Debt payment"
		}
		tx, txErr := uc.createTransactionUseCase.Execute(ctx, userID, CreateTransactionRequest{
			CategoryID:  categoryID,
			WalletID:    req.WalletID,
			Amount:      req.Amount,
			Description: description,
			Date:        req.PaidAt,
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
	}

	liability, payment, err := uc.liabilityService.RecordPayment(
		ctx,
		finance.NewUserID(userID),
		finance.NewLiabilityID(liabilityID),
		req.Amount,
		paidAt,
		expenseID,
		req.Note,
	)
	if err != nil {
		return nil, err
	}
	liabilityResp := toLiabilityResponse(liability)
	paymentResp := toLiabilityPaymentResponse(payment)
	return &RecordLiabilityPaymentResponse{
		Liability: &liabilityResp,
		Payment:   &paymentResp,
	}, nil
}
