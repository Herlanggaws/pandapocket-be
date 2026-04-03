package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
)

type GetWalletsRequest struct {
	Search string `json:"search,omitempty"`
	Page   int    `json:"page,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

type GetWalletsResponse struct {
	Wallets    []*WalletResponse `json:"wallets"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	Limit      int               `json:"limit"`
	TotalPages int               `json:"total_pages"`
}

type GetWalletsUseCase struct {
	walletService *finance.WalletService
}

func NewGetWalletsUseCase(walletService *finance.WalletService) *GetWalletsUseCase {
	return &GetWalletsUseCase{
		walletService: walletService,
	}
}

func (uc *GetWalletsUseCase) Execute(ctx context.Context, userID int, req GetWalletsRequest) (*GetWalletsResponse, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	wallets, totalCount, err := uc.walletService.GetWalletsByUserIDWithFilters(ctx, userID, req.Search, limit, offset)
	if err != nil {
		return nil, err
	}

	responses := make([]*WalletResponse, 0, len(wallets))
	for _, wallet := range wallets {
		responses = append(responses, &WalletResponse{
			ID:        wallet.ID().Value(),
			UserID:    wallet.UserID().Value(),
			Name:      wallet.Name(),
			Amount:    wallet.Amount(),
			CreatedAt: wallet.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: wallet.UpdatedAt().Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	totalPages := int((totalCount + int64(limit) - 1) / int64(limit))
	if totalPages == 0 && totalCount > 0 {
		totalPages = 1
	}

	return &GetWalletsResponse{
		Wallets:    responses,
		Total:      totalCount,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}
