package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
	"time"
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
	ID             int     `json:"id"`
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	CurrencyID     int     `json:"currency_id"`
	CurrentBalance float64 `json:"current_balance"`
	Notes          string  `json:"notes"`
	IsArchived     bool    `json:"is_archived"`
	AsOfDate       *string `json:"as_of_date,omitempty"`
	CreatedAt      string  `json:"created_at"`
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
	Name           string  `json:"name" binding:"required"`
	Type           string  `json:"type" binding:"required,oneof=loan credit_card mortgage other"`
	CurrencyID     int     `json:"currency_id" binding:"required"`
	CurrentBalance float64 `json:"current_balance"`
	Notes          string  `json:"notes"`
	AsOfDate       *string `json:"as_of_date"`
}

type UpdateLiabilityRequest struct {
	Name           string  `json:"name" binding:"required"`
	Type           string  `json:"type" binding:"required,oneof=loan credit_card mortgage other"`
	CurrentBalance float64 `json:"current_balance"`
	Notes          string  `json:"notes"`
	AsOfDate       *string `json:"as_of_date"`
}

type NetWorthSummaryResponse struct {
	CurrencyID              int     `json:"currency_id"`
	LiquidNetWorth          float64 `json:"liquid_net_worth"`
	AssetsTotal             float64 `json:"assets_total"`
	LiabilitiesTotal        float64 `json:"liabilities_total"`
	NetWorth                float64 `json:"net_worth"`
	ExcludedAssetCount      int     `json:"excluded_asset_count"`
	ExcludedLiabilityCount  int     `json:"excluded_liability_count"`
	ExcludedWalletCount     int     `json:"excluded_wallet_count"`
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
		ID:             liability.ID().Value(),
		Name:           liability.Name(),
		Type:           string(liability.Type()),
		CurrencyID:     liability.CurrencyID().Value(),
		CurrentBalance: liability.CurrentBalance(),
		Notes:          liability.Notes(),
		IsArchived:     liability.IsArchived(),
		AsOfDate:       formatOptionalDate(liability.AsOfDate()),
		CreatedAt:      liability.CreatedAt().Format(time.RFC3339),
	}
}

type CreateAssetUseCase struct{ assetService *finance.AssetService }

func NewCreateAssetUseCase(s *finance.AssetService) *CreateAssetUseCase {
	return &CreateAssetUseCase{assetService: s}
}

func (uc *CreateAssetUseCase) Execute(ctx context.Context, userID int, req CreateAssetRequest) (*AssetResponse, error) {
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

type CreateLiabilityUseCase struct{ liabilityService *finance.LiabilityService }

func NewCreateLiabilityUseCase(s *finance.LiabilityService) *CreateLiabilityUseCase {
	return &CreateLiabilityUseCase{liabilityService: s}
}

func (uc *CreateLiabilityUseCase) Execute(ctx context.Context, userID int, req CreateLiabilityRequest) (*LiabilityResponse, error) {
	liabilityType, err := finance.ParseLiabilityType(req.Type)
	if err != nil {
		return nil, err
	}
	asOf, err := parseOptionalDate(req.AsOfDate)
	if err != nil {
		return nil, err
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
	liability, err := uc.liabilityService.Update(
		ctx,
		finance.NewUserID(userID),
		finance.NewLiabilityID(id),
		req.Name,
		liabilityType,
		req.CurrentBalance,
		req.Notes,
		asOf,
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
