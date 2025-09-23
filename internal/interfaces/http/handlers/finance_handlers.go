package handlers

import (
	"fmt"
	"net/http"
	"panda-pocket/internal/application/finance"
	domainFinance "panda-pocket/internal/domain/finance"
	"strconv"

	"github.com/gin-gonic/gin"
)

// FinanceHandlers handles finance-related HTTP requests
type FinanceHandlers struct {
	createTransactionUseCase  *finance.CreateTransactionUseCase
	getTransactionsUseCase    *finance.GetTransactionsUseCase
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

// NewFinanceHandlers creates a new finance handlers instance
func NewFinanceHandlers(
	createTransactionUseCase *finance.CreateTransactionUseCase,
	getTransactionsUseCase *finance.GetTransactionsUseCase,
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
) *FinanceHandlers {
	return &FinanceHandlers{
		createTransactionUseCase:  createTransactionUseCase,
		getTransactionsUseCase:    getTransactionsUseCase,
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

// CreateExpense handles expense creation
func (h *FinanceHandlers) CreateExpense(c *gin.Context) {
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

	c.JSON(http.StatusCreated, gin.H{
		"message": "Expense created successfully",
		"expense": response,
	})
}

// CreateIncome handles income creation
func (h *FinanceHandlers) CreateIncome(c *gin.Context) {
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

	c.JSON(http.StatusCreated, gin.H{
		"message": "Income created successfully",
		"income":  response,
	})
}

// GetExpenses handles getting expenses
func (h *FinanceHandlers) GetExpenses(c *gin.Context) {
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

	c.JSON(http.StatusOK, expenses)
}

// GetIncomes handles getting incomes
func (h *FinanceHandlers) GetIncomes(c *gin.Context) {
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

	c.JSON(http.StatusOK, incomes)
}

// CreateCategory handles category creation
func (h *FinanceHandlers) CreateCategory(c *gin.Context) {
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

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Category created successfully",
		"category": response,
	})
}

// GetCategories handles getting categories
func (h *FinanceHandlers) GetCategories(c *gin.Context) {
	userID := c.GetInt("user_id")
	categoryType := c.Query("type") // Optional filter by type

	response, err := h.getCategoriesUseCase.Execute(c.Request.Context(), userID, categoryType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch categories"})
		return
	}

	c.JSON(http.StatusOK, response.Categories)
}

// UpdateCategory handles category updates
func (h *FinanceHandlers) UpdateCategory(c *gin.Context) {
	userID := c.GetInt("user_id")
	categoryIDStr := c.Param("id")

	// Parse category ID
	var categoryID int
	if _, err := fmt.Sscanf(categoryIDStr, "%d", &categoryID); err != nil {
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

	c.JSON(http.StatusOK, gin.H{
		"message":  "Category updated successfully",
		"category": response,
	})
}

// DeleteCategory handles category deletion
func (h *FinanceHandlers) DeleteCategory(c *gin.Context) {
	userID := c.GetInt("user_id")
	categoryIDStr := c.Param("id")

	// Parse category ID
	var categoryID int
	if _, err := fmt.Sscanf(categoryIDStr, "%d", &categoryID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid category ID"})
		return
	}

	err := h.deleteCategoryUseCase.Execute(c.Request.Context(), userID, categoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Category deleted successfully",
	})
}

// DeleteExpense handles expense deletion
func (h *FinanceHandlers) DeleteExpense(c *gin.Context) {
	userID := c.GetInt("user_id")
	expenseID := c.Param("id")

	// Delete the expense transaction
	err := h.deleteTransactionUseCase.Execute(c.Request.Context(), expenseID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Expense deleted successfully",
	})
}

// DeleteIncome handles income deletion
func (h *FinanceHandlers) DeleteIncome(c *gin.Context) {
	userID := c.GetInt("user_id")
	incomeID := c.Param("id")

	// Delete the income transaction
	err := h.deleteTransactionUseCase.Execute(c.Request.Context(), incomeID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Income deleted successfully",
	})
}

// UpdateExpense handles expense updates
func (h *FinanceHandlers) UpdateExpense(c *gin.Context) {
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
	})
}

// UpdateIncome handles income updates
func (h *FinanceHandlers) UpdateIncome(c *gin.Context) {
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
	})
}

// GetAnalytics handles getting analytics data
func (h *FinanceHandlers) GetAnalytics(c *gin.Context) {
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

	c.JSON(http.StatusOK, response)
}

// CreateBudget handles budget creation
func (h *FinanceHandlers) CreateBudget(c *gin.Context) {
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

	c.JSON(http.StatusCreated, gin.H{
		"message": "Budget created successfully",
		"budget":  response,
	})
}

// GetBudgets handles getting budgets
func (h *FinanceHandlers) GetBudgets(c *gin.Context) {
	userID := c.GetInt("user_id")

	response, err := h.getBudgetsUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch budgets"})
		return
	}

	c.JSON(http.StatusOK, response.Budgets)
}

// UpdateBudget handles budget updates
func (h *FinanceHandlers) UpdateBudget(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{
		"message": "Budget updated successfully",
	})
}

// DeleteBudget handles budget deletion
func (h *FinanceHandlers) DeleteBudget(c *gin.Context) {
	userID := c.GetInt("user_id")
	budgetID := c.Param("id")

	// Delete the budget
	err := h.deleteBudgetUseCase.Execute(c.Request.Context(), budgetID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Budget deleted successfully",
	})
}

// GetCurrencies handles getting currencies
func (h *FinanceHandlers) GetCurrencies(c *gin.Context) {
	userID := c.GetInt("user_id")

	response, err := h.getCurrenciesUseCase.Execute(c.Request.Context(), domainFinance.NewUserID(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch currencies"})
		return
	}

	c.JSON(http.StatusOK, response.Currencies)
}

// CreateCurrency handles currency creation
func (h *FinanceHandlers) CreateCurrency(c *gin.Context) {
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

	c.JSON(http.StatusCreated, gin.H{
		"message":  response.Message,
		"currency": response.Currency,
	})
}

// UpdateCurrency handles currency updates
func (h *FinanceHandlers) UpdateCurrency(c *gin.Context) {
	userID := c.GetInt("user_id")
	currencyID := c.Param("id")

	// Convert string ID to CurrencyID
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

	c.JSON(http.StatusOK, gin.H{
		"message":  response.Message,
		"currency": response.Currency,
	})
}

// DeleteCurrency handles currency deletion
func (h *FinanceHandlers) DeleteCurrency(c *gin.Context) {
	userID := c.GetInt("user_id")
	currencyID := c.Param("id")

	// Convert string ID to CurrencyID
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

	c.JSON(http.StatusOK, gin.H{
		"message": response.Message,
	})
}

// SetDefaultCurrency handles setting the default currency for a user
func (h *FinanceHandlers) SetDefaultCurrency(c *gin.Context) {
	userID := c.GetInt("user_id")
	currencyID := c.Param("id")

	// Set the default currency
	err := h.setDefaultCurrencyUseCase.Execute(c.Request.Context(), userID, currencyID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Default currency set successfully",
	})
}

// GetDefaultCurrency handles getting the default currency for a user
func (h *FinanceHandlers) GetDefaultCurrency(c *gin.Context) {
	userID := c.GetInt("user_id")

	// Get the default currency
	currency, err := h.getDefaultCurrencyUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, currency)
}
