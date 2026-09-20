package finance

import (
	"context"
	"fmt"
	"panda-pocket/internal/domain/entitlement"
	"panda-pocket/internal/domain/finance"
	"sort"
	"time"
)

// GetAnalyticsRequest represents the request for analytics
type GetAnalyticsRequest struct {
	Period    string `json:"period"` // "monthly", "weekly", "yearly", or "custom" when dates set
	WalletID  *int   `json:"wallet_id,omitempty"`
	StartDate string `json:"start_date,omitempty"` // YYYY-MM-DD, Pro only with end_date
	EndDate   string `json:"end_date,omitempty"`
}

type SpendingByCategoryItem struct {
	CategoryID    int     `json:"category_id"`
	CategoryName  string  `json:"category_name"`
	CategoryColor string  `json:"category_color"`
	Amount        float64 `json:"amount"`
	Percentage    float64 `json:"percentage"`
}

type SpendingByPeriodItem struct {
	Period string  `json:"period"`
	Amount float64 `json:"amount"`
	Date   string  `json:"date"`
}

// GetAnalyticsResponse represents the analytics response
type GetAnalyticsResponse struct {
	TotalIncome               float64                  `json:"total_income"`
	TotalSpent                float64                  `json:"total_spent"`
	NetAmount                 float64                  `json:"net_amount"`
	Period                    string                   `json:"period"`
	CurrencyID                int                      `json:"currency_id"`
	TransactionCount          int                      `json:"transaction_count"`
	ExcludedTransactionCount  int                      `json:"excluded_transaction_count"`
	SpendingByCategory        []SpendingByCategoryItem `json:"spending_by_category"`
	SpendingByPeriod          []SpendingByPeriodItem   `json:"spending_by_period"`
}

// GetAnalyticsUseCase handles getting analytics data
type GetAnalyticsUseCase struct {
	transactionService *finance.TransactionService
	categoryService    *finance.CategoryService
	currencyService    *finance.CurrencyService
	entitlements       entitlement.Checker
}

// NewGetAnalyticsUseCase creates a new get analytics use case
func NewGetAnalyticsUseCase(
	transactionService *finance.TransactionService,
	categoryService *finance.CategoryService,
	currencyService *finance.CurrencyService,
	entitlements entitlement.Checker,
) *GetAnalyticsUseCase {
	return &GetAnalyticsUseCase{
		transactionService: transactionService,
		categoryService:    categoryService,
		currencyService:    currencyService,
		entitlements:       entitlements,
	}
}

func parseAnalyticsDate(value string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02", value, time.Local)
}

// Execute executes the get analytics use case
func (uc *GetAnalyticsUseCase) Execute(ctx context.Context, userID int, req GetAnalyticsRequest) (*GetAnalyticsResponse, error) {
	period := req.Period
	if period == "" {
		period = "monthly"
	}

	hasCustom := req.StartDate != "" || req.EndDate != ""
	needsPro := period == "yearly" || hasCustom
	if needsPro {
		isPro, err := uc.entitlements.IsPro(ctx, userID)
		if err != nil {
			return nil, err
		}
		if !isPro {
			return nil, entitlement.RequirePro(entitlement.FeatureInsights)
		}
	}

	var startDate, endDate time.Time
	now := time.Now()
	bucketMode := period

	if hasCustom {
		if req.StartDate == "" || req.EndDate == "" {
			return nil, fmt.Errorf("start_date and end_date are both required for custom range")
		}
		parsedStart, err := parseAnalyticsDate(req.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date")
		}
		parsedEnd, err := parseAnalyticsDate(req.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date")
		}
		if parsedEnd.Before(parsedStart) {
			return nil, fmt.Errorf("end_date must be on or after start_date")
		}
		startDate = parsedStart
		endDate = parsedEnd.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		period = "custom"
		if endDate.Sub(startDate) <= 31*24*time.Hour {
			bucketMode = "weekly"
		} else {
			bucketMode = "yearly"
		}
	} else {
		switch period {
		case "weekly":
			weekday := int(now.Weekday())
			if weekday == 0 {
				weekday = 7
			}
			startDate = now.AddDate(0, 0, -weekday+1).Truncate(24 * time.Hour)
			endDate = startDate.AddDate(0, 0, 6).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		case "yearly":
			startDate = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location())
			endDate = time.Date(now.Year(), 12, 31, 23, 59, 59, 999999999, now.Location())
		default:
			period = "monthly"
			bucketMode = "monthly"
			startDate = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
			endDate = startDate.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		}
	}

	transactions, err := uc.transactionService.GetTransactionsByUserAndDateRange(ctx, finance.NewUserID(userID), startDate, endDate)
	if err != nil {
		return nil, err
	}

	if req.WalletID != nil && *req.WalletID > 0 {
		filtered := make([]*finance.Transaction, 0, len(transactions))
		for _, transaction := range transactions {
			if transaction.WalletID().Value() == *req.WalletID {
				filtered = append(filtered, transaction)
			}
		}
		transactions = filtered
	}

	primary, err := uc.currencyService.GetPrimaryCurrency(ctx, finance.NewUserID(userID))
	if err != nil {
		return nil, err
	}
	primaryCurrencyID := primary.ID().Value()

	excludedCount := 0
	primaryOnly := make([]*finance.Transaction, 0, len(transactions))
	for _, transaction := range transactions {
		if transaction.CurrencyID().Value() != primaryCurrencyID {
			excludedCount++
			continue
		}
		primaryOnly = append(primaryOnly, transaction)
	}
	transactions = primaryOnly

	var totalIncome, totalSpent float64
	categoryTotals := map[int]float64{}
	periodTotals := map[string]float64{}
	periodDates := map[string]string{}

	for _, transaction := range transactions {
		amount := transaction.Amount().Amount()
		if transaction.Type() == finance.TransactionTypeIncome {
			totalIncome += amount
			continue
		}
		if transaction.Type() != finance.TransactionTypeExpense {
			continue
		}

		totalSpent += amount
		categoryTotals[transaction.CategoryID().Value()] += amount

		var bucketKey, bucketLabel, bucketDate string
		d := transaction.Date()
		switch bucketMode {
		case "weekly":
			bucketKey = d.Format("2006-01-02")
			bucketLabel = d.Format("Mon")
			bucketDate = bucketKey
		case "yearly":
			bucketKey = d.Format("2006-01")
			bucketLabel = d.Format("Jan")
			bucketDate = bucketKey + "-01"
		default:
			bucketKey = fmt.Sprintf("%d", ((d.Day()-1)/7)+1)
			bucketLabel = fmt.Sprintf("Week %s", bucketKey)
			bucketDate = d.Format("2006-01-02")
		}
		periodTotals[bucketKey] += amount
		if _, ok := periodDates[bucketKey]; !ok {
			periodDates[bucketKey] = bucketDate
		}
		periodDates[bucketKey+"__label"] = bucketLabel
	}

	spendingByCategory := make([]SpendingByCategoryItem, 0, len(categoryTotals))
	for categoryID, amount := range categoryTotals {
		name := "Unknown"
		color := "#94a3b8"
		if cat, err := uc.categoryService.GetCategoryByID(ctx, finance.NewCategoryID(categoryID)); err == nil {
			name = cat.Name()
			if cat.Color() != "" {
				color = cat.Color()
			}
		}
		percentage := 0.0
		if totalSpent > 0 {
			percentage = (amount / totalSpent) * 100
		}
		spendingByCategory = append(spendingByCategory, SpendingByCategoryItem{
			CategoryID:    categoryID,
			CategoryName:  name,
			CategoryColor: color,
			Amount:        amount,
			Percentage:    percentage,
		})
	}
	sort.Slice(spendingByCategory, func(i, j int) bool {
		return spendingByCategory[i].Amount > spendingByCategory[j].Amount
	})

	keys := make([]string, 0, len(periodTotals))
	for k := range periodTotals {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	spendingByPeriod := make([]SpendingByPeriodItem, 0, len(keys))
	for _, k := range keys {
		spendingByPeriod = append(spendingByPeriod, SpendingByPeriodItem{
			Period: periodDates[k+"__label"],
			Amount: periodTotals[k],
			Date:   periodDates[k],
		})
	}

	return &GetAnalyticsResponse{
		TotalIncome:              totalIncome,
		TotalSpent:               totalSpent,
		NetAmount:                totalIncome - totalSpent,
		Period:                   period,
		CurrencyID:               primaryCurrencyID,
		TransactionCount:         len(transactions),
		ExcludedTransactionCount: excludedCount,
		SpendingByCategory:       spendingByCategory,
		SpendingByPeriod:         spendingByPeriod,
	}, nil
}
