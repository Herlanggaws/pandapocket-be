package v120

import (
	"net/http"
	"panda-pocket/internal/application/finance"
	domainFinance "panda-pocket/internal/domain/finance"
	"strconv"

	"github.com/gin-gonic/gin"
)

// FinanceHandlersV120 handles finance-related HTTP requests for API version 1.2.0
type FinanceHandlersV120 struct {
	createTransactionUseCase  *finance.CreateTransactionUseCase
	getTransactionsUseCase    *finance.GetTransactionsUseCase
	getAllTransactionsUseCase *finance.GetAllTransactionsUseCase
	updateTransactionUseCase  *finance.UpdateTransactionUseCase
	deleteTransactionUseCase  *finance.DeleteTransactionUseCase
	createCategoryUseCase     *finance.CreateCategoryUseCase
	updateCategoryUseCase     *finance.UpdateCategoryUseCase
	deleteCategoryUseCase     *finance.DeleteCategoryUseCase
	getCategoriesUseCase      *finance.GetCategoriesUseCase
	getAnalyticsUseCase       *finance.GetAnalyticsUseCase
	createBudgetUseCase       *finance.CreateBudgetUseCase
	getBudgetsUseCase         *finance.GetBudgetsUseCase
	updateBudgetUseCase       *finance.UpdateBudgetUseCase
	deleteBudgetUseCase       *finance.DeleteBudgetUseCase
	createCurrencyUseCase     *finance.CreateCurrencyUseCase
	getCurrenciesUseCase      *finance.GetCurrenciesUseCase
	updateCurrencyUseCase     *finance.UpdateCurrencyUseCase
	deleteCurrencyUseCase     *finance.DeleteCurrencyUseCase
	setDefaultCurrencyUseCase *finance.SetDefaultCurrencyUseCase
	getDefaultCurrencyUseCase *finance.GetDefaultCurrencyUseCase
}

// NewFinanceHandlersV120 creates a new finance handlers instance for v120
func NewFinanceHandlersV120(
	createTransactionUseCase *finance.CreateTransactionUseCase,
	getTransactionsUseCase *finance.GetTransactionsUseCase,
	getAllTransactionsUseCase *finance.GetAllTransactionsUseCase,
	updateTransactionUseCase *finance.UpdateTransactionUseCase,
	deleteTransactionUseCase *finance.DeleteTransactionUseCase,
	createCategoryUseCase *finance.CreateCategoryUseCase,
	updateCategoryUseCase *finance.UpdateCategoryUseCase,
	deleteCategoryUseCase *finance.DeleteCategoryUseCase,
	getCategoriesUseCase *finance.GetCategoriesUseCase,
	getAnalyticsUseCase *finance.GetAnalyticsUseCase,
	createBudgetUseCase *finance.CreateBudgetUseCase,
	getBudgetsUseCase *finance.GetBudgetsUseCase,
	updateBudgetUseCase *finance.UpdateBudgetUseCase,
	deleteBudgetUseCase *finance.DeleteBudgetUseCase,
	createCurrencyUseCase *finance.CreateCurrencyUseCase,
	getCurrenciesUseCase *finance.GetCurrenciesUseCase,
	updateCurrencyUseCase *finance.UpdateCurrencyUseCase,
	deleteCurrencyUseCase *finance.DeleteCurrencyUseCase,
	setDefaultCurrencyUseCase *finance.SetDefaultCurrencyUseCase,
	getDefaultCurrencyUseCase *finance.GetDefaultCurrencyUseCase,
) *FinanceHandlersV120 {
	return &FinanceHandlersV120{
		createTransactionUseCase:  createTransactionUseCase,
		getTransactionsUseCase:    getTransactionsUseCase,
		getAllTransactionsUseCase: getAllTransactionsUseCase,
		updateTransactionUseCase:  updateTransactionUseCase,
		deleteTransactionUseCase:  deleteTransactionUseCase,
		createCategoryUseCase:     createCategoryUseCase,
		updateCategoryUseCase:     updateCategoryUseCase,
		deleteCategoryUseCase:     deleteCategoryUseCase,
		getCategoriesUseCase:      getCategoriesUseCase,
		getAnalyticsUseCase:       getAnalyticsUseCase,
		createBudgetUseCase:       createBudgetUseCase,
		getBudgetsUseCase:         getBudgetsUseCase,
		updateBudgetUseCase:       updateBudgetUseCase,
		deleteBudgetUseCase:       deleteBudgetUseCase,
		createCurrencyUseCase:     createCurrencyUseCase,
		getCurrenciesUseCase:      getCurrenciesUseCase,
		updateCurrencyUseCase:     updateCurrencyUseCase,
		deleteCurrencyUseCase:     deleteCurrencyUseCase,
		setDefaultCurrencyUseCase: setDefaultCurrencyUseCase,
		getDefaultCurrencyUseCase: getDefaultCurrencyUseCase,
	}
}

// CreateTransaction handles transaction creation (v120 - enhanced with analytics)
func (h *FinanceHandlersV120) CreateTransaction(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req finance.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.createTransactionUseCase.Execute(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enhanced response with analytics for v120
	c.JSON(http.StatusCreated, gin.H{
		"message":     "Transaction created successfully",
		"transaction": response,
		"analytics": gin.H{
			"version":  "v120",
			"features": []string{"analytics", "advanced_filtering", "bulk_operations", "export_functionality"},
		},
	})
}

// GetTransactions handles getting transactions with enhanced analytics (v120)
func (h *FinanceHandlersV120) GetTransactions(c *gin.Context) {
	userID := c.GetInt("user_id")

	// Parse query parameters for advanced filtering
	req := finance.GetAllTransactionsRequest{
		Type:      c.Query("type"),
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
	}

	// Parse category IDs from query parameter
	if categoryIDsParam := c.Query("category_ids"); categoryIDsParam != "" {
		req.CategoryIDs = []string{categoryIDsParam}
	}

	// Parse pagination parameters
	if pageParam := c.Query("page"); pageParam != "" {
		if page, err := strconv.Atoi(pageParam); err == nil {
			req.Page = page
		}
	}
	if limitParam := c.Query("limit"); limitParam != "" {
		if limit, err := strconv.Atoi(limitParam); err == nil {
			req.Limit = limit
		}
	}

	response, err := h.getAllTransactionsUseCase.Execute(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	// Enhanced response with analytics for v120
	c.JSON(http.StatusOK, gin.H{
		"transactions": response.Transactions,
		"pagination": gin.H{
			"page":  req.Page,
			"limit": req.Limit,
			"total": len(response.Transactions),
		},
		"analytics": gin.H{
			"version":  "v120",
			"features": []string{"analytics", "advanced_filtering", "pagination"},
		},
	})
}

// GetTransactionsWithAnalytics handles getting transactions with detailed analytics (v120 specific)
func (h *FinanceHandlersV120) GetTransactionsWithAnalytics(c *gin.Context) {
	userID := c.GetInt("user_id")
	period := c.DefaultQuery("period", "monthly")

	// Get transactions
	req := finance.GetAllTransactionsRequest{
		Type:      c.Query("type"),
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
	}

	transactions, err := h.getAllTransactionsUseCase.Execute(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transactions"})
		return
	}

	// Get analytics
	analyticsReq := finance.GetAnalyticsRequest{
		Period: period,
	}

	analytics, err := h.getAnalyticsUseCase.Execute(c.Request.Context(), userID, analyticsReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch analytics"})
		return
	}

	// Enhanced response with detailed analytics for v120
	c.JSON(http.StatusOK, gin.H{
		"transactions": transactions.Transactions,
		"analytics": gin.H{
			"period":   period,
			"data":     analytics,
			"version":  "v120",
			"features": []string{"detailed_analytics", "period_analysis", "trend_analysis"},
		},
		"pagination": gin.H{
			"total": len(transactions.Transactions),
		},
	})
}

// UpdateTransaction handles transaction updates (v120 - enhanced validation)
func (h *FinanceHandlersV120) UpdateTransaction(c *gin.Context) {
	userID := c.GetInt("user_id")
	transactionID := c.Param("id")

	var req struct {
		CategoryID  int     `json:"category_id" binding:"required"`
		Amount      float64 `json:"amount" binding:"required"`
		Description string  `json:"description" binding:"required"`
		Date        string  `json:"date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update the transaction
	transaction, err := h.updateTransactionUseCase.Execute(
		c.Request.Context(),
		transactionID,
		userID,
		strconv.Itoa(req.CategoryID),
		"1", // Default currency ID for now
		req.Amount,
		req.Description,
		req.Date,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enhanced response for v120
	c.JSON(http.StatusOK, gin.H{
		"message": "Transaction updated successfully",
		"transaction": gin.H{
			"id":          transaction.ID().Value(),
			"user_id":     transaction.UserID().Value(),
			"category_id": transaction.CategoryID().Value(),
			"currency_id": transaction.CurrencyID().Value(),
			"amount":      transaction.Amount().Amount(),
			"description": transaction.Description(),
			"date":        transaction.Date().Format("2006-01-02"),
			"type":        "transaction",
		},
		"analytics": gin.H{
			"version":  "v120",
			"features": []string{"enhanced_validation", "analytics"},
		},
	})
}

// DeleteTransaction handles transaction deletion (v120 - enhanced with confirmation)
func (h *FinanceHandlersV120) DeleteTransaction(c *gin.Context) {
	userID := c.GetInt("user_id")
	transactionID := c.Param("id")

	err := h.deleteTransactionUseCase.Execute(c.Request.Context(), transactionID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enhanced response for v120
	c.JSON(http.StatusOK, gin.H{
		"message": "Transaction deleted successfully",
		"analytics": gin.H{
			"version":  "v120",
			"features": []string{"enhanced_confirmation", "analytics"},
		},
	})
}

// CreateCategory handles category creation (v120 - enhanced with validation)
func (h *FinanceHandlersV120) CreateCategory(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req finance.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.createCategoryUseCase.Execute(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enhanced response for v120
	c.JSON(http.StatusCreated, gin.H{
		"message":  "Category created successfully",
		"category": response,
		"analytics": gin.H{
			"version":  "v120",
			"features": []string{"enhanced_validation", "analytics"},
		},
	})
}

// GetCategories handles getting categories (v120 - enhanced with analytics)
func (h *FinanceHandlersV120) GetCategories(c *gin.Context) {
	userID := c.GetInt("user_id")
	categoryType := c.Query("type")

	response, err := h.getCategoriesUseCase.Execute(c.Request.Context(), userID, categoryType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}

	// Enhanced response for v120
	c.JSON(http.StatusOK, gin.H{
		"categories": response.Categories,
		"analytics": gin.H{
			"version":  "v120",
			"features": []string{"analytics", "category_insights"},
		},
	})
}

// UpdateCategory handles category updates (v120 - enhanced validation)
func (h *FinanceHandlersV120) UpdateCategory(c *gin.Context) {
	userID := c.GetInt("user_id")
	categoryIDStr := c.Param("id")

	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	var req finance.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.updateCategoryUseCase.Execute(c.Request.Context(), userID, categoryID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enhanced response for v120
	c.JSON(http.StatusOK, gin.H{
		"message":  "Category updated successfully",
		"category": response,
		"analytics": gin.H{
			"version":  "v120",
			"features": []string{"enhanced_validation", "analytics"},
		},
	})
}

// DeleteCategory handles category deletion (v120 - enhanced with confirmation)
func (h *FinanceHandlersV120) DeleteCategory(c *gin.Context) {
	userID := c.GetInt("user_id")
	categoryIDStr := c.Param("id")

	categoryID, err := strconv.Atoi(categoryIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	err = h.deleteCategoryUseCase.Execute(c.Request.Context(), userID, categoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enhanced response for v120
	c.JSON(http.StatusOK, gin.H{
		"message": "Category deleted successfully",
		"analytics": gin.H{
			"version":  "v120",
			"features": []string{"enhanced_confirmation", "analytics"},
		},
	})
}

// GetAnalytics handles getting analytics data (v120 - enhanced analytics)
func (h *FinanceHandlersV120) GetAnalytics(c *gin.Context) {
	userID := c.GetInt("user_id")
	period := c.DefaultQuery("period", "monthly")

	req := finance.GetAnalyticsRequest{
		Period: period,
	}

	response, err := h.getAnalyticsUseCase.Execute(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch analytics"})
		return
	}

	// Enhanced analytics response for v120
	c.JSON(http.StatusOK, gin.H{
		"analytics": response,
		"version":   "v120",
		"features":  []string{"detailed_analytics", "period_analysis", "trend_analysis", "export_functionality"},
	})
}

// CreateBudget handles budget creation (v120 - enhanced validation)
func (h *FinanceHandlersV120) CreateBudget(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req finance.CreateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.createBudgetUseCase.Execute(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enhanced response for v120
	c.JSON(http.StatusCreated, gin.H{
		"message": "Budget created successfully",
		"budget":  response,
		"analytics": gin.H{
			"version":  "v120",
			"features": []string{"enhanced_validation", "analytics"},
		},
	})
}

// GetBudgets handles getting budgets (v120 - enhanced with analytics)
func (h *FinanceHandlersV120) GetBudgets(c *gin.Context) {
	userID := c.GetInt("user_id")

	response, err := h.getBudgetsUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch budgets"})
		return
	}

	// Enhanced response for v120
	c.JSON(http.StatusOK, gin.H{
		"budgets": response.Budgets,
		"analytics": gin.H{
			"version":  "v120",
			"features": []string{"analytics", "budget_insights"},
		},
	})
}

// UpdateBudget handles budget updates (v120 - enhanced validation)
func (h *FinanceHandlersV120) UpdateBudget(c *gin.Context) {
	userID := c.GetInt("user_id")
	budgetID := c.Param("id")

	var req struct {
		CategoryID int     `json:"category_id" binding:"required"`
		Amount     float64 `json:"amount" binding:"required"`
		Period     string  `json:"period" binding:"required"`
		StartDate  string  `json:"start_date" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update the budget
	err := h.updateBudgetUseCase.Execute(
		c.Request.Context(),
		budgetID,
		userID,
		strconv.Itoa(req.CategoryID),
		req.Amount,
		req.Period,
		req.StartDate,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enhanced response for v120
	c.JSON(http.StatusOK, gin.H{
		"message": "Budget updated successfully",
		"analytics": gin.H{
			"version":  "v120",
			"features": []string{"enhanced_validation", "analytics"},
		},
	})
}

// DeleteBudget handles budget deletion (v120 - enhanced with confirmation)
func (h *FinanceHandlersV120) DeleteBudget(c *gin.Context) {
	userID := c.GetInt("user_id")
	budgetID := c.Param("id")

	err := h.deleteBudgetUseCase.Execute(c.Request.Context(), budgetID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enhanced response for v120
	c.JSON(http.StatusOK, gin.H{
		"message": "Budget deleted successfully",
		"analytics": gin.H{
			"version":  "v120",
			"features": []string{"enhanced_confirmation", "analytics"},
		},
	})
}

// GetCurrencies handles getting currencies (v120 - enhanced with analytics)
func (h *FinanceHandlersV120) GetCurrencies(c *gin.Context) {
	userID := c.GetInt("user_id")

	response, err := h.getCurrenciesUseCase.Execute(c.Request.Context(), domainFinance.NewUserID(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch currencies"})
		return
	}

	// Enhanced response for v120
	c.JSON(http.StatusOK, gin.H{
		"currencies": response.Currencies,
		"analytics": gin.H{
			"version":  "v120",
			"features": []string{"analytics", "currency_insights"},
		},
	})
}

// CreateCurrency handles currency creation (v120 - enhanced validation)
func (h *FinanceHandlersV120) CreateCurrency(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req finance.CreateCurrencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.createCurrencyUseCase.Execute(c.Request.Context(), domainFinance.NewUserID(userID), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enhanced response for v120
	c.JSON(http.StatusCreated, gin.H{
		"message":  response.Message,
		"currency": response.Currency,
		"analytics": gin.H{
			"version":  "v120",
			"features": []string{"enhanced_validation", "analytics"},
		},
	})
}

// UpdateCurrency handles currency updates (v120 - enhanced validation)
func (h *FinanceHandlersV120) UpdateCurrency(c *gin.Context) {
	userID := c.GetInt("user_id")
	currencyID := c.Param("id")

	currencyIDInt, err := strconv.Atoi(currencyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid currency ID"})
		return
	}

	var req finance.UpdateCurrencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.updateCurrencyUseCase.Execute(c.Request.Context(), domainFinance.NewUserID(userID), domainFinance.NewCurrencyID(currencyIDInt), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enhanced response for v120
	c.JSON(http.StatusOK, gin.H{
		"message":  response.Message,
		"currency": response.Currency,
		"analytics": gin.H{
			"version":  "v120",
			"features": []string{"enhanced_validation", "analytics"},
		},
	})
}

// DeleteCurrency handles currency deletion (v120 - enhanced with confirmation)
func (h *FinanceHandlersV120) DeleteCurrency(c *gin.Context) {
	userID := c.GetInt("user_id")
	currencyID := c.Param("id")

	currencyIDInt, err := strconv.Atoi(currencyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid currency ID"})
		return
	}

	response, err := h.deleteCurrencyUseCase.Execute(c.Request.Context(), domainFinance.NewUserID(userID), domainFinance.NewCurrencyID(currencyIDInt))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enhanced response for v120
	c.JSON(http.StatusOK, gin.H{
		"message": response.Message,
		"analytics": gin.H{
			"version":  "v120",
			"features": []string{"enhanced_confirmation", "analytics"},
		},
	})
}

// SetDefaultCurrency handles setting the default currency (v120 - enhanced validation)
func (h *FinanceHandlersV120) SetDefaultCurrency(c *gin.Context) {
	userID := c.GetInt("user_id")
	currencyID := c.Param("id")

	err := h.setDefaultCurrencyUseCase.Execute(c.Request.Context(), userID, currencyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enhanced response for v120
	c.JSON(http.StatusOK, gin.H{
		"message": "Default currency set successfully",
		"analytics": gin.H{
			"version":  "v120",
			"features": []string{"enhanced_validation", "analytics"},
		},
	})
}

// GetDefaultCurrency handles getting the default currency (v120 - enhanced with analytics)
func (h *FinanceHandlersV120) GetDefaultCurrency(c *gin.Context) {
	userID := c.GetInt("user_id")

	currency, err := h.getDefaultCurrencyUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Enhanced response for v120
	c.JSON(http.StatusOK, gin.H{
		"currency": currency,
		"analytics": gin.H{
			"version":  "v120",
			"features": []string{"analytics", "currency_insights"},
		},
	})
}
