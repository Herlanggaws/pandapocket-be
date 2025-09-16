package handlers

import (
	"net/http"
	"panda-pocket/internal/application/finance"

	"github.com/gin-gonic/gin"
)

// FinanceHandlers handles finance-related HTTP requests
type FinanceHandlers struct {
	createTransactionUseCase *finance.CreateTransactionUseCase
	getTransactionsUseCase   *finance.GetTransactionsUseCase
	createCategoryUseCase    *finance.CreateCategoryUseCase
	getCategoriesUseCase     *finance.GetCategoriesUseCase
	getAnalyticsUseCase      *finance.GetAnalyticsUseCase
	createBudgetUseCase      *finance.CreateBudgetUseCase
	getBudgetsUseCase        *finance.GetBudgetsUseCase
}

// NewFinanceHandlers creates a new finance handlers instance
func NewFinanceHandlers(
	createTransactionUseCase *finance.CreateTransactionUseCase,
	getTransactionsUseCase *finance.GetTransactionsUseCase,
	createCategoryUseCase *finance.CreateCategoryUseCase,
	getCategoriesUseCase *finance.GetCategoriesUseCase,
	getAnalyticsUseCase *finance.GetAnalyticsUseCase,
	createBudgetUseCase *finance.CreateBudgetUseCase,
	getBudgetsUseCase *finance.GetBudgetsUseCase,
) *FinanceHandlers {
	return &FinanceHandlers{
		createTransactionUseCase: createTransactionUseCase,
		getTransactionsUseCase:   getTransactionsUseCase,
		createCategoryUseCase:    createCategoryUseCase,
		getCategoriesUseCase:     getCategoriesUseCase,
		getAnalyticsUseCase:      getAnalyticsUseCase,
		createBudgetUseCase:      createBudgetUseCase,
		getBudgetsUseCase:        getBudgetsUseCase,
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

// DeleteExpense handles expense deletion
func (h *FinanceHandlers) DeleteExpense(c *gin.Context) {
	// This would need to be implemented with a delete transaction use case
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
}

// DeleteIncome handles income deletion
func (h *FinanceHandlers) DeleteIncome(c *gin.Context) {
	// This would need to be implemented with a delete transaction use case
	c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented yet"})
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
