package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
	"strconv"
	"strings"
	"time"
)

// GetAllTransactionsRequest represents the request for getting all transactions with filters
type GetAllTransactionsRequest struct {
	Type        string   `json:"type,omitempty"`         // "income", "expense", or empty for both
	CategoryIDs []string `json:"category_ids,omitempty"` // Comma-separated category IDs
	StartDate   string   `json:"start_date,omitempty"`   // Date in YYYY-MM-DD format
	EndDate     string   `json:"end_date,omitempty"`     // Date in YYYY-MM-DD format
	Page        int      `json:"page,omitempty"`         // Page number (1-based)
	Limit       int      `json:"limit,omitempty"`        // Number of items per page
}

// GetAllTransactionsResponse represents the response for getting all transactions
type GetAllTransactionsResponse struct {
	Transactions []TransactionResponse     `json:"transactions"`
	Total        int64                     `json:"total"`
	Page         int                       `json:"page"`
	Limit        int                       `json:"limit"`
	TotalPages   int                       `json:"total_pages"`
	Filters      GetAllTransactionsRequest `json:"filters"`
}

// GetAllTransactionsUseCase handles getting all transactions for a user with filters
type GetAllTransactionsUseCase struct {
	transactionService *finance.TransactionService
	categoryService    *finance.CategoryService
}

// NewGetAllTransactionsUseCase creates a new get all transactions use case
func NewGetAllTransactionsUseCase(transactionService *finance.TransactionService, categoryService *finance.CategoryService) *GetAllTransactionsUseCase {
	return &GetAllTransactionsUseCase{
		transactionService: transactionService,
		categoryService:    categoryService,
	}
}

// Execute executes the get all transactions use case
func (uc *GetAllTransactionsUseCase) Execute(ctx context.Context, userID int, req GetAllTransactionsRequest) (*GetAllTransactionsResponse, error) {
	// Build filters
	filters := finance.TransactionFilters{}

	// Parse transaction type
	if req.Type != "" {
		switch req.Type {
		case "income":
			filters.TransactionType = &[]finance.TransactionType{finance.TransactionTypeIncome}[0]
		case "expense":
			filters.TransactionType = &[]finance.TransactionType{finance.TransactionTypeExpense}[0]
		}
	}

	// Parse category IDs
	if len(req.CategoryIDs) > 0 {
		categoryIDs := make([]finance.CategoryID, 0, len(req.CategoryIDs))
		for _, categoryIDStr := range req.CategoryIDs {
			if categoryIDStr == "" {
				continue
			}
			// Handle comma-separated values
			if strings.Contains(categoryIDStr, ",") {
				parts := strings.Split(categoryIDStr, ",")
				for _, part := range parts {
					if id, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
						categoryIDs = append(categoryIDs, finance.NewCategoryID(id))
					}
				}
			} else {
				if id, err := strconv.Atoi(categoryIDStr); err == nil {
					categoryIDs = append(categoryIDs, finance.NewCategoryID(id))
				}
			}
		}
		filters.CategoryIDs = categoryIDs
	}

	// Parse date range
	if req.StartDate != "" {
		if startDate, err := time.Parse("2006-01-02", req.StartDate); err == nil {
			filters.StartDate = &startDate
		}
	}
	if req.EndDate != "" {
		if endDate, err := time.Parse("2006-01-02", req.EndDate); err == nil {
			filters.EndDate = &endDate
		}
	}

	// Parse pagination parameters
	page := req.Page
	if page <= 0 {
		page = 1
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20 // Default limit
	}
	if limit > 100 {
		limit = 100 // Maximum limit
	}

	// Calculate offset (0-based)
	offset := (page - 1) * limit
	filters.Limit = limit
	filters.Offset = offset

	// Get transactions with filters
	transactions, totalCount, err := uc.transactionService.GetTransactionsByUserWithFilters(ctx, finance.NewUserID(userID), filters)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	transactionResponses := make([]TransactionResponse, len(transactions))
	for i, transaction := range transactions {
		// Fetch category details
		category, err := uc.categoryService.GetCategoryByID(ctx, transaction.CategoryID())
		if err != nil {
			// If category not found, create a default response
			category = &finance.Category{}
		}

		transactionResponses[i] = TransactionResponse{
			ID:     transaction.ID().Value(),
			UserID: transaction.UserID().Value(),
			Category: CategoryResponse{
				ID:        category.ID().Value(),
				Name:      category.Name(),
				Color:     category.Color(),
				Type:      string(category.Type()),
				IsDefault: category.IsDefault(),
			},
			CurrencyID:  transaction.CurrencyID().Value(),
			Amount:      transaction.Amount().Amount(),
			Description: transaction.Description(),
			Date:        transaction.Date().Format("2006-01-02"),
			Type:        string(transaction.Type()),
			CreatedAt:   transaction.CreatedAt().Format(time.RFC3339),
		}
	}

	// Calculate total pages
	totalPages := int((totalCount + int64(limit) - 1) / int64(limit))
	if totalPages == 0 {
		totalPages = 1
	}

	return &GetAllTransactionsResponse{
		Transactions: transactionResponses,
		Total:        totalCount,
		Page:         page,
		Limit:        limit,
		TotalPages:   totalPages,
		Filters:      req,
	}, nil
}
