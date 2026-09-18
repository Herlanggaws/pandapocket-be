package handlers

import (
	"errors"
	"strconv"

	"panda-pocket/internal/application/finance"
	"panda-pocket/internal/domain/entitlement"

	"github.com/gin-gonic/gin"
)

type ExportHandlers struct {
	exportTransactionsUseCase *finance.ExportTransactionsUseCase
}

func NewExportHandlers(exportTransactionsUseCase *finance.ExportTransactionsUseCase) *ExportHandlers {
	return &ExportHandlers{exportTransactionsUseCase: exportTransactionsUseCase}
}

// ExportTransactions handles GET /api/export/transactions
func (h *ExportHandlers) ExportTransactions(c *gin.Context) {
	userID := c.GetInt("user_id")

	startDate, endDate := finance.DefaultExportDateRange(c.Query("start_date"), c.Query("end_date"))

	req := finance.ExportTransactionsRequest{
		Format:    c.Query("format"),
		Type:      c.Query("type"),
		StartDate: startDate,
		EndDate:   endDate,
	}

	if walletIDParam := c.Query("wallet_id"); walletIDParam != "" {
		if walletID, err := strconv.Atoi(walletIDParam); err == nil {
			req.WalletID = &walletID
		}
	}
	if categoryIDsParam := c.Query("category_ids"); categoryIDsParam != "" {
		req.CategoryIDs = []string{categoryIDsParam}
	}

	result, err := h.exportTransactionsUseCase.Execute(c.Request.Context(), userID, req)
	if err != nil {
		switch {
		case errors.Is(err, entitlement.ErrPremiumRequired):
			ForbiddenResponse(c, "PREMIUM_REQUIRED", err.Error())
		case errors.Is(err, finance.ErrInvalidExportFormat):
			ValidationErrorResponse(c, err.Error())
		case errors.Is(err, finance.ErrExportTooLarge):
			BadRequestResponse(c, "EXPORT_TOO_LARGE", err.Error())
		default:
			InternalServerErrorResponse(c, "EXPORT_TRANSACTIONS_ERROR", "Failed to export transactions")
		}
		return
	}

	FileDownloadResponse(c, result.ContentType, result.Filename, result.Content)
}
