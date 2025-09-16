package finance

import (
	"context"
	"panda-pocket/internal/domain/finance"
	"time"
)

// GetAnalyticsRequest represents the request for analytics
type GetAnalyticsRequest struct {
	Period string `json:"period"` // "monthly", "weekly", "yearly"
}

// GetAnalyticsResponse represents the analytics response
type GetAnalyticsResponse struct {
	TotalIncome      float64 `json:"total_income"`
	TotalSpent       float64 `json:"total_spent"`
	NetAmount        float64 `json:"net_amount"`
	Period           string  `json:"period"`
	TransactionCount int     `json:"transaction_count"`
}

// GetAnalyticsUseCase handles getting analytics data
type GetAnalyticsUseCase struct {
	transactionService *finance.TransactionService
}

// NewGetAnalyticsUseCase creates a new get analytics use case
func NewGetAnalyticsUseCase(transactionService *finance.TransactionService) *GetAnalyticsUseCase {
	return &GetAnalyticsUseCase{
		transactionService: transactionService,
	}
}

// Execute executes the get analytics use case
func (uc *GetAnalyticsUseCase) Execute(ctx context.Context, userID int, req GetAnalyticsRequest) (*GetAnalyticsResponse, error) {
	// Determine the date range based on period
	var startDate, endDate time.Time
	now := time.Now()

	switch req.Period {
	case "weekly":
		// Get current week (Monday to Sunday)
		weekday := int(now.Weekday())
		if weekday == 0 { // Sunday
			weekday = 7
		}
		startDate = now.AddDate(0, 0, -weekday+1).Truncate(24 * time.Hour)
		endDate = startDate.AddDate(0, 0, 6).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	case "yearly":
		// Get current year
		startDate = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
		endDate = time.Date(now.Year(), 12, 31, 23, 59, 59, 999999999, now.Location())
	default: // "monthly" or any other value defaults to monthly
		// Get current month
		startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		endDate = startDate.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
	}

	// Get transactions for the period
	transactions, err := uc.transactionService.GetTransactionsByUserAndDateRange(ctx, finance.NewUserID(userID), startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Calculate analytics
	var totalIncome, totalSpent float64
	transactionCount := len(transactions)

	for _, transaction := range transactions {
		if transaction.Type() == finance.TransactionTypeIncome {
			totalIncome += transaction.Amount().Amount()
		} else if transaction.Type() == finance.TransactionTypeExpense {
			totalSpent += transaction.Amount().Amount()
		}
	}

	netAmount := totalIncome - totalSpent

	return &GetAnalyticsResponse{
		TotalIncome:      totalIncome,
		TotalSpent:       totalSpent,
		NetAmount:        netAmount,
		Period:           req.Period,
		TransactionCount: transactionCount,
	}, nil
}
