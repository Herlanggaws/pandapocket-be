package finance

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"panda-pocket/internal/domain/entitlement"
	"panda-pocket/internal/domain/finance"

	"github.com/phpdave11/gofpdf"
)

const exportMaxRows = 5000

var (
	ErrExportTooLarge     = errors.New("export exceeds maximum of 5000 transactions; narrow the date range")
	ErrInvalidExportFormat = errors.New("format must be csv or pdf")
)

type ExportTransactionsRequest struct {
	Format      string
	Type        string
	CategoryIDs []string
	WalletID    *int
	StartDate   string
	EndDate     string
}

type ExportTransactionsResult struct {
	Content     []byte
	ContentType string
	Filename    string
}

type exportRow struct {
	Date        string
	Type        string
	Category    string
	Description string
	Amount      float64
	CurrencyID  int
	WalletID    int
}

type ExportTransactionsUseCase struct {
	transactionService *finance.TransactionService
	categoryService    *finance.CategoryService
	entitlements       entitlement.Checker
}

func NewExportTransactionsUseCase(
	transactionService *finance.TransactionService,
	categoryService *finance.CategoryService,
	entitlements entitlement.Checker,
) *ExportTransactionsUseCase {
	return &ExportTransactionsUseCase{
		transactionService: transactionService,
		categoryService:    categoryService,
		entitlements:       entitlements,
	}
}

func (uc *ExportTransactionsUseCase) Execute(
	ctx context.Context,
	userID int,
	req ExportTransactionsRequest,
) (*ExportTransactionsResult, error) {
	format := strings.ToLower(strings.TrimSpace(req.Format))
	if format != "csv" && format != "pdf" {
		return nil, ErrInvalidExportFormat
	}

	isPro, err := uc.entitlements.IsPro(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !isPro {
		return nil, entitlement.ErrPremiumRequired
	}

	filters, err := buildExportFilters(req)
	if err != nil {
		return nil, err
	}

	transactions, totalCount, err := uc.transactionService.GetTransactionsByUserWithFilters(
		ctx,
		finance.NewUserID(userID),
		filters,
	)
	if err != nil {
		return nil, err
	}
	if totalCount > exportMaxRows {
		return nil, ErrExportTooLarge
	}

	rows := make([]exportRow, 0, len(transactions))
	var totalIncomes, totalExpenses float64
	for _, transaction := range transactions {
		categoryName := ""
		if category, catErr := uc.categoryService.GetCategoryByID(ctx, transaction.CategoryID()); catErr == nil && category != nil {
			categoryName = category.Name()
		}

		amt := transaction.Amount().Amount()
		if transaction.Type() == finance.TransactionTypeIncome {
			totalIncomes += amt
		} else {
			totalExpenses += amt
		}

		rows = append(rows, exportRow{
			Date:        transaction.Date().Format("2006-01-02"),
			Type:        string(transaction.Type()),
			Category:    categoryName,
			Description: transaction.Description(),
			Amount:      amt,
			CurrencyID:  transaction.CurrencyID().Value(),
			WalletID:    transaction.WalletID().Value(),
		})
	}

	startLabel := req.StartDate
	endLabel := req.EndDate
	filenameBase := fmt.Sprintf("berbudget-transactions-%s-%s", startLabel, endLabel)

	switch format {
	case "csv":
		content, csvErr := buildTransactionsCSV(rows)
		if csvErr != nil {
			return nil, csvErr
		}
		return &ExportTransactionsResult{
			Content:     content,
			ContentType: "text/csv; charset=utf-8",
			Filename:    filenameBase + ".csv",
		}, nil
	default:
		content, pdfErr := buildTransactionsPDF(rows, startLabel, endLabel, totalIncomes, totalExpenses)
		if pdfErr != nil {
			return nil, pdfErr
		}
		return &ExportTransactionsResult{
			Content:     content,
			ContentType: "application/pdf",
			Filename:    filenameBase + ".pdf",
		}, nil
	}
}

func buildExportFilters(req ExportTransactionsRequest) (finance.TransactionFilters, error) {
	filters := finance.TransactionFilters{
		Limit:  exportMaxRows,
		Offset: 0,
	}

	if req.Type != "" {
		switch req.Type {
		case "income":
			t := finance.TransactionTypeIncome
			filters.TransactionType = &t
		case "expense":
			t := finance.TransactionTypeExpense
			filters.TransactionType = &t
		}
	}

	if len(req.CategoryIDs) > 0 {
		categoryIDs := make([]finance.CategoryID, 0)
		for _, categoryIDStr := range req.CategoryIDs {
			if categoryIDStr == "" {
				continue
			}
			parts := strings.Split(categoryIDStr, ",")
			for _, part := range parts {
				if id, err := strconv.Atoi(strings.TrimSpace(part)); err == nil {
					categoryIDs = append(categoryIDs, finance.NewCategoryID(id))
				}
			}
		}
		filters.CategoryIDs = categoryIDs
	}

	if req.WalletID != nil && *req.WalletID > 0 {
		walletID := finance.NewWalletID(*req.WalletID)
		filters.WalletID = &walletID
	}

	if req.StartDate != "" {
		startDate, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return filters, fmt.Errorf("invalid start_date: %w", err)
		}
		filters.StartDate = &startDate
	}
	if req.EndDate != "" {
		endDate, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return filters, fmt.Errorf("invalid end_date: %w", err)
		}
		filters.EndDate = &endDate
	}

	return filters, nil
}

func buildTransactionsCSV(rows []exportRow) ([]byte, error) {
	var buf bytes.Buffer
	// UTF-8 BOM helps Excel open CSV correctly
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	writer := csv.NewWriter(&buf)
	if err := writer.Write([]string{
		"date", "type", "category", "description", "amount", "currency_id", "wallet_id",
	}); err != nil {
		return nil, err
	}
	for _, row := range rows {
		if err := writer.Write([]string{
			row.Date,
			row.Type,
			row.Category,
			row.Description,
			strconv.FormatFloat(row.Amount, 'f', -1, 64),
			strconv.Itoa(row.CurrencyID),
			strconv.Itoa(row.WalletID),
		}); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func buildTransactionsPDF(
	rows []exportRow,
	startDate, endDate string,
	totalIncomes, totalExpenses float64,
) ([]byte, error) {
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.SetTitle("Berbudget Transactions", false)
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(0, 8, "Berbudget — Transactions Export")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, fmt.Sprintf("Period: %s to %s", startDate, endDate))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf(
		"Total income: %.2f  |  Total expense: %.2f  |  Rows: %d",
		totalIncomes,
		totalExpenses,
		len(rows),
	))
	pdf.Ln(10)

	headers := []string{"Date", "Type", "Category", "Description", "Amount", "Currency", "Wallet"}
	widths := []float64{28, 22, 40, 90, 30, 24, 24}

	pdf.SetFont("Arial", "B", 9)
	for i, header := range headers {
		pdf.CellFormat(widths[i], 7, header, "1", 0, "L", false, 0, "")
	}
	pdf.Ln(-1)

	pdf.SetFont("Arial", "", 8)
	for _, row := range rows {
		desc := row.Description
		if len(desc) > 60 {
			desc = desc[:57] + "..."
		}
		values := []string{
			row.Date,
			row.Type,
			row.Category,
			desc,
			strconv.FormatFloat(row.Amount, 'f', 2, 64),
			strconv.Itoa(row.CurrencyID),
			strconv.Itoa(row.WalletID),
		}
		for i, value := range values {
			pdf.CellFormat(widths[i], 6, value, "1", 0, "L", false, 0, "")
		}
		pdf.Ln(-1)
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DefaultExportDateRange returns start/end YYYY-MM-DD matching list transactions defaults.
func DefaultExportDateRange(startDateStr, endDateStr string) (string, string) {
	now := time.Now()
	if startDateStr == "" && endDateStr == "" {
		return now.AddDate(0, 0, -30).Format("2006-01-02"), now.Format("2006-01-02")
	}
	if startDateStr != "" && endDateStr == "" {
		return startDateStr, now.Format("2006-01-02")
	}
	if startDateStr == "" && endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			return endDate.AddDate(0, 0, -30).Format("2006-01-02"), endDateStr
		}
		return now.AddDate(0, 0, -30).Format("2006-01-02"), endDateStr
	}
	return startDateStr, endDateStr
}
