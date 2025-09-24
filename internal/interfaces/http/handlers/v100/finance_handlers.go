package v100

import (
	"net/http"
	"panda-pocket/internal/application/finance"
	domainFinance "panda-pocket/internal/domain/finance"
	"strconv"

	"github.com/gin-gonic/gin"
)

// FinanceHandlersV100 handles finance-related HTTP requests for API version 1.0.0 (Legacy)
type FinanceHandlersV100 struct {
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

// NewFinanceHandlersV100 creates a new finance handlers instance for v100 (Legacy)
func NewFinanceHandlersV100(
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
) *FinanceHandlersV100 {
	return &FinanceHandlersV100{
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

// CreateExpense handles expense creation (v100 - Legacy)
func (h *FinanceHandlersV100) CreateExpense(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req finance.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set transaction type to expense
	req.Type = "expense"

	response, err := h.createTransactionUseCase.Execute(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Legacy response format for v100
	c.JSON(http.StatusCreated, gin.H{
		"message":     "Expense created successfully",
		"expense":     response,
		"version":     "v100",
		"deprecated":  true,
		"sunset_date": "2024-06-01",
		"upgrade_url": "https://docs.pandapocket.com/upgrade",
	})
}

// CreateIncome handles income creation (v100 - Legacy)
func (h *FinanceHandlersV100) CreateIncome(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req finance.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set transaction type to income
	req.Type = "income"

	response, err := h.createTransactionUseCase.Execute(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Legacy response format for v100
	c.JSON(http.StatusCreated, gin.H{
		"message":     "Income created successfully",
		"income":      response,
		"version":     "v100",
		"deprecated":  true,
		"sunset_date": "2024-06-01",
		"upgrade_url": "https://docs.pandapocket.com/upgrade",
	})
}

// GetExpenses handles getting expenses (v100 - Legacy)
func (h *FinanceHandlersV100) GetExpenses(c *gin.Context) {
	userID := c.GetInt("user_id")

	response, err := h.getTransactionsUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch expenses"})
		return
	}

	// Filter only expenses
	var expenses []finance.TransactionResponse
	for _, transaction := range response.Transactions {
		if transaction.Type == "expense" {
			expenses = append(expenses, transaction)
		}
	}

	// Legacy response format for v100
	c.JSON(http.StatusOK, gin.H{
		"expenses":    expenses,
		"version":     "v100",
		"deprecated":  true,
		"sunset_date": "2024-06-01",
		"upgrade_url": "https://docs.pandapocket.com/upgrade",
	})
}

// GetIncomes handles getting incomes (v100 - Legacy)
func (h *FinanceHandlersV100) GetIncomes(c *gin.Context) {
	userID := c.GetInt("user_id")

	response, err := h.getTransactionsUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch incomes"})
		return
	}

	// Filter only incomes
	var incomes []finance.TransactionResponse
	for _, transaction := range response.Transactions {
		if transaction.Type == "income" {
			incomes = append(incomes, transaction)
		}
	}

	// Legacy response format for v100
	c.JSON(http.StatusOK, gin.H{
		"incomes":     incomes,
		"version":     "v100",
		"deprecated":  true,
		"sunset_date": "2024-06-01",
		"upgrade_url": "https://docs.pandapocket.com/upgrade",
	})
}

// GetAllTransactions handles getting all transactions (v100 - Legacy)
func (h *FinanceHandlersV100) GetAllTransactions(c *gin.Context) {
	userID := c.GetInt("user_id")

	// Parse query parameters
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

	// Legacy response format for v100
	c.JSON(http.StatusOK, gin.H{
		"transactions": response.Transactions,
		"version":      "v100",
		"deprecated":   true,
		"sunset_date":  "2024-06-01",
		"upgrade_url":  "https://docs.pandapocket.com/upgrade",
	})
}

// CreateCategory handles category creation (v100 - Legacy)
func (h *FinanceHandlersV100) CreateCategory(c *gin.Context) {
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

	// Legacy response format for v100
	c.JSON(http.StatusCreated, gin.H{
		"message":     "Category created successfully",
		"category":    response,
		"version":     "v100",
		"deprecated":  true,
		"sunset_date": "2024-06-01",
		"upgrade_url": "https://docs.pandapocket.com/upgrade",
	})
}

// GetCategories handles getting categories (v100 - Legacy)
func (h *FinanceHandlersV100) GetCategories(c *gin.Context) {
	userID := c.GetInt("user_id")
	categoryType := c.Query("type")

	response, err := h.getCategoriesUseCase.Execute(c.Request.Context(), userID, categoryType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}

	// Legacy response format for v100
	c.JSON(http.StatusOK, gin.H{
		"categories":  response.Categories,
		"version":     "v100",
		"deprecated":  true,
		"sunset_date": "2024-06-01",
		"upgrade_url": "https://docs.pandapocket.com/upgrade",
	})
}

// UpdateCategory handles category updates (v100 - Legacy)
func (h *FinanceHandlersV100) UpdateCategory(c *gin.Context) {
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

	// Legacy response format for v100
	c.JSON(http.StatusOK, gin.H{
		"message":     "Category updated successfully",
		"category":    response,
		"version":     "v100",
		"deprecated":  true,
		"sunset_date": "2024-06-01",
		"upgrade_url": "https://docs.pandapocket.com/upgrade",
	})
}

// DeleteCategory handles category deletion (v100 - Legacy)
func (h *FinanceHandlersV100) DeleteCategory(c *gin.Context) {
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

	// Legacy response format for v100
	c.JSON(http.StatusOK, gin.H{
		"message":     "Category deleted successfully",
		"version":     "v100",
		"deprecated":  true,
		"sunset_date": "2024-06-01",
		"upgrade_url": "https://docs.pandapocket.com/upgrade",
	})
}

// DeleteExpense handles expense deletion (v100 - Legacy)
func (h *FinanceHandlersV100) DeleteExpense(c *gin.Context) {
	userID := c.GetInt("user_id")
	expenseID := c.Param("id")

	err := h.deleteTransactionUseCase.Execute(c.Request.Context(), expenseID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Legacy response format for v100
	c.JSON(http.StatusOK, gin.H{
		"message":     "Expense deleted successfully",
		"version":     "v100",
		"deprecated":  true,
		"sunset_date": "2024-06-01",
		"upgrade_url": "https://docs.pandapocket.com/upgrade",
	})
}

// DeleteIncome handles income deletion (v100 - Legacy)
func (h *FinanceHandlersV100) DeleteIncome(c *gin.Context) {
	userID := c.GetInt("user_id")
	incomeID := c.Param("id")

	err := h.deleteTransactionUseCase.Execute(c.Request.Context(), incomeID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Legacy response format for v100
	c.JSON(http.StatusOK, gin.H{
		"message":     "Income deleted successfully",
		"version":     "v100",
		"deprecated":  true,
		"sunset_date": "2024-06-01",
		"upgrade_url": "https://docs.pandapocket.com/upgrade",
	})
}

// UpdateExpense handles expense updates (v100 - Legacy)
func (h *FinanceHandlersV100) UpdateExpense(c *gin.Context) {
	userID := c.GetInt("user_id")
	expenseID := c.Param("id")

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

	// Update the expense transaction
	transaction, err := h.updateTransactionUseCase.Execute(
		c.Request.Context(),
		expenseID,
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

	// Legacy response format for v100
	c.JSON(http.StatusOK, gin.H{
		"message": "Expense updated successfully",
		"expense": gin.H{
			"id":          transaction.ID().Value(),
			"user_id":     transaction.UserID().Value(),
			"category_id": transaction.CategoryID().Value(),
			"currency_id": transaction.CurrencyID().Value(),
			"amount":      transaction.Amount().Amount(),
			"description": transaction.Description(),
			"date":        transaction.Date().Format("2006-01-02"),
			"type":        "expense",
		},
		"version":     "v100",
		"deprecated":  true,
		"sunset_date": "2024-06-01",
		"upgrade_url": "https://docs.pandapocket.com/upgrade",
	})
}

// UpdateIncome handles income updates (v100 - Legacy)
func (h *FinanceHandlersV100) UpdateIncome(c *gin.Context) {
	userID := c.GetInt("user_id")
	incomeID := c.Param("id")

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

	// Update the income transaction
	transaction, err := h.updateTransactionUseCase.Execute(
		c.Request.Context(),
		incomeID,
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

	// Legacy response format for v100
	c.JSON(http.StatusOK, gin.H{
		"message": "Income updated successfully",
		"income": gin.H{
			"id":          transaction.ID().Value(),
			"user_id":     transaction.UserID().Value(),
			"category_id": transaction.CategoryID().Value(),
			"currency_id": transaction.CurrencyID().Value(),
			"amount":      transaction.Amount().Amount(),
			"description": transaction.Description(),
			"date":        transaction.Date().Format("2006-01-02"),
			"type":        "income",
		},
		"version":     "v100",
		"deprecated":  true,
		"sunset_date": "2024-06-01",
		"upgrade_url": "https://docs.pandapocket.com/upgrade",
	})
}

// GetAnalytics handles getting analytics data (v100 - Legacy)
func (h *FinanceHandlersV100) GetAnalytics(c *gin.Context) {
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

	// Legacy response format for v100
	c.JSON(http.StatusOK, gin.H{
		"analytics":   response,
		"version":     "v100",
		"deprecated":  true,
		"sunset_date": "2024-06-01",
		"upgrade_url": "https://docs.pandapocket.com/upgrade",
	})
}

// CreateBudget handles budget creation (v100 - Legacy)
func (h *FinanceHandlersV100) CreateBudget(c *gin.Context) {
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

	// Legacy response format for v100
	c.JSON(http.StatusCreated, gin.H{
		"message":     "Budget created successfully",
		"budget":      response,
		"version":     "v100",
		"deprecated":  true,
		"sunset_date": "2024-06-01",
		"upgrade_url": "https://docs.pandapocket.com/upgrade",
	})
}

// GetBudgets handles getting budgets (v100 - Legacy)
func (h *FinanceHandlersV100) GetBudgets(c *gin.Context) {
	userID := c.GetInt("user_id")

	response, err := h.getBudgetsUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch budgets"})
		return
	}

	// Legacy response format for v100
	c.JSON(http.StatusOK, gin.H{
		"budgets":     response.Budgets,
		"version":     "v100",
		"deprecated":  true,
		"sunset_date": "2024-06-01",
		"upgrade_url": "https://docs.pandapocket.com/upgrade",
	})
}

// UpdateBudget handles budget updates (v100 - Legacy)
func (h *FinanceHandlersV100) UpdateBudget(c *gin.Context) {
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

	// Legacy response format for v100
	c.JSON(http.StatusOK, gin.H{
		"message":     "Budget updated successfully",
		"version":     "v100",
		"deprecated":  true,
		"sunset_date": "2024-06-01",
		"upgrade_url": "https://docs.pandapocket.com/upgrade",
	})
}

// DeleteBudget handles budget deletion (v100 - Legacy)
func (h *FinanceHandlersV100) DeleteBudget(c *gin.Context) {
	userID := c.GetInt("user_id")
	budgetID := c.Param("id")

	err := h.deleteBudgetUseCase.Execute(c.Request.Context(), budgetID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Legacy response format for v100
	c.JSON(http.StatusOK, gin.H{
		"message":     "Budget deleted successfully",
		"version":     "v100",
		"deprecated":  true,
		"sunset_date": "2024-06-01",
		"upgrade_url": "https://docs.pandapocket.com/upgrade",
	})
}

// GetCurrencies handles getting currencies (v100 - Legacy)
func (h *FinanceHandlersV100) GetCurrencies(c *gin.Context) {
	userID := c.GetInt("user_id")

	response, err := h.getCurrenciesUseCase.Execute(c.Request.Context(), domainFinance.NewUserID(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch currencies"})
		return
	}

	// Legacy response format for v100
	c.JSON(http.StatusOK, gin.H{
		"currencies":  response.Currencies,
		"version":     "v100",
		"deprecated":  true,
		"sunset_date": "2024-06-01",
		"upgrade_url": "https://docs.pandapocket.com/upgrade",
	})
}

// CreateCurrency handles currency creation (v100 - Legacy)
func (h *FinanceHandlersV100) CreateCurrency(c *gin.Context) {
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

	// Legacy response format for v100
	c.JSON(http.StatusCreated, gin.H{
		"message":     response.Message,
		"currency":    response.Currency,
		"version":     "v100",
		"deprecated":  true,
		"sunset_date": "2024-06-01",
		"upgrade_url": "https://docs.pandapocket.com/upgrade",
	})
}

// UpdateCurrency handles currency updates (v100 - Legacy)
func (h *FinanceHandlersV100) UpdateCurrency(c *gin.Context) {
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

	// Legacy response format for v100
	c.JSON(http.StatusOK, gin.H{
		"message":     response.Message,
		"currency":    response.Currency,
		"version":     "v100",
		"deprecated":  true,
		"sunset_date": "2024-06-01",
		"upgrade_url": "https://docs.pandapocket.com/upgrade",
	})
}

// DeleteCurrency handles currency deletion (v100 - Legacy)
func (h *FinanceHandlersV100) DeleteCurrency(c *gin.Context) {
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

	// Legacy response format for v100
	c.JSON(http.StatusOK, gin.H{
		"message":     response.Message,
		"version":     "v100",
		"deprecated":  true,
		"sunset_date": "2024-06-01",
		"upgrade_url": "https://docs.pandapocket.com/upgrade",
	})
}

// SetDefaultCurrency handles setting the default currency (v100 - Legacy)
func (h *FinanceHandlersV100) SetDefaultCurrency(c *gin.Context) {
	userID := c.GetInt("user_id")
	currencyID := c.Param("id")

	err := h.setDefaultCurrencyUseCase.Execute(c.Request.Context(), userID, currencyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Legacy response format for v100
	c.JSON(http.StatusOK, gin.H{
		"message":     "Default currency set successfully",
		"version":     "v100",
		"deprecated":  true,
		"sunset_date": "2024-06-01",
		"upgrade_url": "https://docs.pandapocket.com/upgrade",
	})
}

// GetDefaultCurrency handles getting the default currency (v100 - Legacy)
func (h *FinanceHandlersV100) GetDefaultCurrency(c *gin.Context) {
	userID := c.GetInt("user_id")

	currency, err := h.getDefaultCurrencyUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Legacy response format for v100
	c.JSON(http.StatusOK, gin.H{
		"currency":    currency,
		"version":     "v100",
		"deprecated":  true,
		"sunset_date": "2024-06-01",
		"upgrade_url": "https://docs.pandapocket.com/upgrade",
	})
}
