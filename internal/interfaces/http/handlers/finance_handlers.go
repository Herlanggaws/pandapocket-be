package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"panda-pocket/internal/application/finance"
	"panda-pocket/internal/domain/entitlement"
	domainFinance "panda-pocket/internal/domain/finance"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// FinanceHandlers handles finance-related HTTP requests
type FinanceHandlers struct {
	createTransactionUseCase      *finance.CreateTransactionUseCase
	getTransactionsUseCase        *finance.GetTransactionsUseCase
	getAllTransactionsUseCase     *finance.GetAllTransactionsUseCase
	updateTransactionUseCase      *finance.UpdateTransactionUseCase
	deleteTransactionUseCase      *finance.DeleteTransactionUseCase
	createCategoryUseCase         *finance.CreateCategoryUseCase
	updateCategoryUseCase         *finance.UpdateCategoryUseCase
	deleteCategoryUseCase         *finance.DeleteCategoryUseCase
	getCategoriesUseCase          *finance.GetCategoriesUseCase
	getAnalyticsUseCase           *finance.GetAnalyticsUseCase
	createBudgetUseCase           *finance.CreateBudgetUseCase
	getBudgetsUseCase             *finance.GetBudgetsUseCase
	updateBudgetUseCase           *finance.UpdateBudgetUseCase
	deleteBudgetUseCase           *finance.DeleteBudgetUseCase
	createCurrencyUseCase         *finance.CreateCurrencyUseCase
	getCurrenciesUseCase          *finance.GetCurrenciesUseCase
	updateCurrencyUseCase         *finance.UpdateCurrencyUseCase
	deleteCurrencyUseCase         *finance.DeleteCurrencyUseCase
	setDefaultCurrencyUseCase     *finance.SetDefaultCurrencyUseCase
	getDefaultCurrencyUseCase     *finance.GetDefaultCurrencyUseCase
	checkBudgetAlertsUseCase      *finance.CheckBudgetAlertsUseCase
	createRecurringUseCase        *finance.CreateRecurringTransactionUseCase
	getRecurringUseCase           *finance.GetRecurringTransactionsUseCase
	deleteRecurringUseCase        *finance.DeleteRecurringTransactionUseCase
	listPendingUseCase            *finance.ListPendingTransactionsUseCase
	confirmPendingUseCase         *finance.ConfirmPendingTransactionUseCase
	rejectPendingUseCase          *finance.RejectPendingTransactionUseCase
	createWalletUseCase           *finance.CreateWalletUseCase
	getWalletsUseCase             *finance.GetWalletsUseCase
	getWalletUseCase              *finance.GetWalletUseCase
	updateWalletUseCase           *finance.UpdateWalletUseCase
	setDefaultWalletUseCase       *finance.SetDefaultWalletUseCase
	archiveWalletUseCase          *finance.ArchiveWalletUseCase
	unarchiveWalletUseCase        *finance.UnarchiveWalletUseCase
	getWalletBalanceUseCase       *finance.GetWalletBalanceUseCase
	getWalletSummaryUseCase       *finance.GetWalletSummaryUseCase
	getHealthScoreUseCase         *finance.GetHealthScoreUseCase
	getHealthScoreHistoryUseCase  *finance.GetHealthScoreHistoryUseCase
	createGoalUseCase             *finance.CreateGoalUseCase
	getGoalsUseCase               *finance.GetGoalsUseCase
	getGoalUseCase                *finance.GetGoalUseCase
	updateGoalUseCase             *finance.UpdateGoalUseCase
	deleteGoalUseCase             *finance.DeleteGoalUseCase
	createAssetUseCase            *finance.CreateAssetUseCase
	getAssetsUseCase              *finance.GetAssetsUseCase
	updateAssetUseCase            *finance.UpdateAssetUseCase
	archiveAssetUseCase           *finance.ArchiveAssetUseCase
	unarchiveAssetUseCase         *finance.UnarchiveAssetUseCase
	createLiabilityUseCase        *finance.CreateLiabilityUseCase
	getLiabilitiesUseCase         *finance.GetLiabilitiesUseCase
	updateLiabilityUseCase        *finance.UpdateLiabilityUseCase
	archiveLiabilityUseCase       *finance.ArchiveLiabilityUseCase
	unarchiveLiabilityUseCase     *finance.UnarchiveLiabilityUseCase
	listLiabilityPaymentsUseCase  *finance.ListLiabilityPaymentsUseCase
	recordLiabilityPaymentUseCase *finance.RecordLiabilityPaymentUseCase
	completeOnboardingUseCase     *finance.CompleteOnboardingUseCase
	getNetWorthSummaryUseCase     *finance.GetNetWorthSummaryUseCase
	createTransferUseCase         *finance.CreateTransferUseCase
	getTransfersUseCase           *finance.GetTransfersUseCase
}

// NewFinanceHandlers creates a new finance handlers instance
func NewFinanceHandlers(
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
	checkBudgetAlertsUseCase *finance.CheckBudgetAlertsUseCase,
	createRecurringUseCase *finance.CreateRecurringTransactionUseCase,
	getRecurringUseCase *finance.GetRecurringTransactionsUseCase,
	deleteRecurringUseCase *finance.DeleteRecurringTransactionUseCase,
	listPendingUseCase *finance.ListPendingTransactionsUseCase,
	confirmPendingUseCase *finance.ConfirmPendingTransactionUseCase,
	rejectPendingUseCase *finance.RejectPendingTransactionUseCase,
	createWalletUseCase *finance.CreateWalletUseCase,
	getWalletsUseCase *finance.GetWalletsUseCase,
	getWalletUseCase *finance.GetWalletUseCase,
	updateWalletUseCase *finance.UpdateWalletUseCase,
	setDefaultWalletUseCase *finance.SetDefaultWalletUseCase,
	archiveWalletUseCase *finance.ArchiveWalletUseCase,
	unarchiveWalletUseCase *finance.UnarchiveWalletUseCase,
	getWalletBalanceUseCase *finance.GetWalletBalanceUseCase,
	getWalletSummaryUseCase *finance.GetWalletSummaryUseCase,
	getHealthScoreUseCase *finance.GetHealthScoreUseCase,
	getHealthScoreHistoryUseCase *finance.GetHealthScoreHistoryUseCase,
	createGoalUseCase *finance.CreateGoalUseCase,
	getGoalsUseCase *finance.GetGoalsUseCase,
	getGoalUseCase *finance.GetGoalUseCase,
	updateGoalUseCase *finance.UpdateGoalUseCase,
	deleteGoalUseCase *finance.DeleteGoalUseCase,
	createAssetUseCase *finance.CreateAssetUseCase,
	getAssetsUseCase *finance.GetAssetsUseCase,
	updateAssetUseCase *finance.UpdateAssetUseCase,
	archiveAssetUseCase *finance.ArchiveAssetUseCase,
	unarchiveAssetUseCase *finance.UnarchiveAssetUseCase,
	createLiabilityUseCase *finance.CreateLiabilityUseCase,
	getLiabilitiesUseCase *finance.GetLiabilitiesUseCase,
	updateLiabilityUseCase *finance.UpdateLiabilityUseCase,
	archiveLiabilityUseCase *finance.ArchiveLiabilityUseCase,
	unarchiveLiabilityUseCase *finance.UnarchiveLiabilityUseCase,
	listLiabilityPaymentsUseCase *finance.ListLiabilityPaymentsUseCase,
	recordLiabilityPaymentUseCase *finance.RecordLiabilityPaymentUseCase,
	completeOnboardingUseCase *finance.CompleteOnboardingUseCase,
	getNetWorthSummaryUseCase *finance.GetNetWorthSummaryUseCase,
	createTransferUseCase *finance.CreateTransferUseCase,
	getTransfersUseCase *finance.GetTransfersUseCase,
) *FinanceHandlers {
	return &FinanceHandlers{
		createTransactionUseCase:      createTransactionUseCase,
		getTransactionsUseCase:        getTransactionsUseCase,
		getAllTransactionsUseCase:     getAllTransactionsUseCase,
		updateTransactionUseCase:      updateTransactionUseCase,
		deleteTransactionUseCase:      deleteTransactionUseCase,
		createCategoryUseCase:         createCategoryUseCase,
		updateCategoryUseCase:         updateCategoryUseCase,
		deleteCategoryUseCase:         deleteCategoryUseCase,
		getCategoriesUseCase:          getCategoriesUseCase,
		getAnalyticsUseCase:           getAnalyticsUseCase,
		createBudgetUseCase:           createBudgetUseCase,
		getBudgetsUseCase:             getBudgetsUseCase,
		updateBudgetUseCase:           updateBudgetUseCase,
		deleteBudgetUseCase:           deleteBudgetUseCase,
		createCurrencyUseCase:         createCurrencyUseCase,
		getCurrenciesUseCase:          getCurrenciesUseCase,
		updateCurrencyUseCase:         updateCurrencyUseCase,
		deleteCurrencyUseCase:         deleteCurrencyUseCase,
		setDefaultCurrencyUseCase:     setDefaultCurrencyUseCase,
		getDefaultCurrencyUseCase:     getDefaultCurrencyUseCase,
		checkBudgetAlertsUseCase:      checkBudgetAlertsUseCase,
		createRecurringUseCase:        createRecurringUseCase,
		getRecurringUseCase:           getRecurringUseCase,
		deleteRecurringUseCase:        deleteRecurringUseCase,
		listPendingUseCase:            listPendingUseCase,
		confirmPendingUseCase:         confirmPendingUseCase,
		rejectPendingUseCase:          rejectPendingUseCase,
		createWalletUseCase:           createWalletUseCase,
		getWalletsUseCase:             getWalletsUseCase,
		getWalletUseCase:              getWalletUseCase,
		updateWalletUseCase:           updateWalletUseCase,
		setDefaultWalletUseCase:       setDefaultWalletUseCase,
		archiveWalletUseCase:          archiveWalletUseCase,
		unarchiveWalletUseCase:        unarchiveWalletUseCase,
		getWalletBalanceUseCase:       getWalletBalanceUseCase,
		getWalletSummaryUseCase:       getWalletSummaryUseCase,
		getHealthScoreUseCase:         getHealthScoreUseCase,
		getHealthScoreHistoryUseCase:  getHealthScoreHistoryUseCase,
		createGoalUseCase:             createGoalUseCase,
		getGoalsUseCase:               getGoalsUseCase,
		getGoalUseCase:                getGoalUseCase,
		updateGoalUseCase:             updateGoalUseCase,
		deleteGoalUseCase:             deleteGoalUseCase,
		createAssetUseCase:            createAssetUseCase,
		getAssetsUseCase:              getAssetsUseCase,
		updateAssetUseCase:            updateAssetUseCase,
		archiveAssetUseCase:           archiveAssetUseCase,
		unarchiveAssetUseCase:         unarchiveAssetUseCase,
		createLiabilityUseCase:        createLiabilityUseCase,
		getLiabilitiesUseCase:         getLiabilitiesUseCase,
		updateLiabilityUseCase:        updateLiabilityUseCase,
		archiveLiabilityUseCase:       archiveLiabilityUseCase,
		unarchiveLiabilityUseCase:     unarchiveLiabilityUseCase,
		listLiabilityPaymentsUseCase:  listLiabilityPaymentsUseCase,
		recordLiabilityPaymentUseCase: recordLiabilityPaymentUseCase,
		completeOnboardingUseCase:     completeOnboardingUseCase,
		getNetWorthSummaryUseCase:     getNetWorthSummaryUseCase,
		createTransferUseCase:         createTransferUseCase,
		getTransfersUseCase:           getTransfersUseCase,
	}
}

// CreateExpense handles expense creation
func (h *FinanceHandlers) CreateExpense(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req finance.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, err.Error())
		return
	}

	// Set transaction type to expense
	req.Type = "expense"

	response, err := h.createTransactionUseCase.Execute(c.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, entitlement.ErrPremiumRequired) {
			PremiumRequiredResponse(c, err)
			return
		}
		HandleError(c, err, http.StatusBadRequest)
		return
	}

	if h.checkBudgetAlertsUseCase != nil {
		h.checkBudgetAlertsUseCase.Execute(c.Request.Context(), userID, req.CategoryID)
	}

	SuccessResponse(c, http.StatusCreated, gin.H{
		"expense": response,
	})
}

// CreateIncome handles income creation
func (h *FinanceHandlers) CreateIncome(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req finance.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, err.Error())
		return
	}

	// Set transaction type to income
	req.Type = "income"

	response, err := h.createTransactionUseCase.Execute(c.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, entitlement.ErrPremiumRequired) {
			PremiumRequiredResponse(c, err)
			return
		}
		HandleError(c, err, http.StatusBadRequest)
		return
	}

	SuccessResponse(c, http.StatusCreated, gin.H{
		"income": response,
	})
}

// GetExpenses handles getting expenses
func (h *FinanceHandlers) GetExpenses(c *gin.Context) {
	userID := c.GetInt("user_id")

	response, err := h.getTransactionsUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		InternalServerErrorResponse(c, "FETCH_EXPENSES_ERROR", "Failed to fetch expenses")
		return
	}

	// Filter only expenses
	var expenses []finance.TransactionResponse
	for _, transaction := range response.Transactions {
		if transaction.Type == "expense" {
			expenses = append(expenses, transaction)
		}
	}

	SuccessResponse(c, http.StatusOK, expenses)
}

// GetIncomes handles getting incomes
func (h *FinanceHandlers) GetIncomes(c *gin.Context) {
	userID := c.GetInt("user_id")

	response, err := h.getTransactionsUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		InternalServerErrorResponse(c, "FETCH_INCOMES_ERROR", "Failed to fetch incomes")
		return
	}

	// Filter only incomes
	var incomes []finance.TransactionResponse
	for _, transaction := range response.Transactions {
		if transaction.Type == "income" {
			incomes = append(incomes, transaction)
		}
	}

	SuccessResponse(c, http.StatusOK, incomes)
}

// GetAllTransactions handles getting all transactions with filters
func (h *FinanceHandlers) GetAllTransactions(c *gin.Context) {
	userID := c.GetInt("user_id")

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	now := time.Now()

	if startDateStr == "" && endDateStr == "" {
		startDateStr = now.AddDate(0, 0, -30).Format("2006-01-02")
		endDateStr = now.Format("2006-01-02")
	} else if startDateStr != "" && endDateStr == "" {
		endDateStr = now.Format("2006-01-02")
	} else if startDateStr == "" && endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			startDateStr = endDate.AddDate(0, 0, -30).Format("2006-01-02")
		} else {
			startDateStr = now.AddDate(0, 0, -30).Format("2006-01-02")
		}
	}

	// Parse query parameters
	req := finance.GetAllTransactionsRequest{
		Type:      c.Query("type"),
		StartDate: startDateStr,
		EndDate:   endDateStr,
	}

	if walletIDParam := c.Query("wallet_id"); walletIDParam != "" {
		if walletID, err := strconv.Atoi(walletIDParam); err == nil {
			req.WalletID = &walletID
		}
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
		InternalServerErrorResponse(c, "FETCH_TRANSACTIONS_ERROR", "Failed to fetch transactions")
		return
	}

	SuccessResponse(c, http.StatusOK, response)
}

// CreateCategory handles category creation
func (h *FinanceHandlers) CreateCategory(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req finance.CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, err.Error())
		return
	}

	response, err := h.createCategoryUseCase.Execute(c.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, entitlement.ErrPremiumRequired) {
			PremiumRequiredResponse(c, err)
			return
		}
		HandleError(c, err, http.StatusBadRequest)
		return
	}

	SuccessResponse(c, http.StatusCreated, gin.H{
		"category": response,
	})
}

// GetCategories handles getting categories
func (h *FinanceHandlers) GetCategories(c *gin.Context) {
	userID := c.GetInt("user_id")
	categoryType := c.Query("type") // Optional filter by type

	response, err := h.getCategoriesUseCase.Execute(c.Request.Context(), userID, categoryType)
	if err != nil {
		InternalServerErrorResponse(c, "FETCH_CATEGORIES_ERROR", "Failed to fetch categories")
		return
	}

	SuccessResponse(c, http.StatusOK, response.Categories)
}

// UpdateCategory handles category updates
func (h *FinanceHandlers) UpdateCategory(c *gin.Context) {
	userID := c.GetInt("user_id")
	categoryIDStr := c.Param("id")

	// Parse category ID
	var categoryID int
	if _, err := fmt.Sscanf(categoryIDStr, "%d", &categoryID); err != nil {
		BadRequestResponse(c, "INVALID_CATEGORY_ID", "Invalid category ID")
		return
	}

	var req finance.UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, err.Error())
		return
	}

	response, err := h.updateCategoryUseCase.Execute(c.Request.Context(), userID, categoryID, req)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
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
		BadRequestResponse(c, "INVALID_CATEGORY_ID", "Invalid category ID")
		return
	}

	err := h.deleteCategoryUseCase.Execute(c.Request.Context(), userID, categoryID)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
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
		HandleError(c, err, http.StatusBadRequest)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
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
		HandleError(c, err, http.StatusBadRequest)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
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
		ValidationErrorResponse(c, err.Error())
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
		domainFinance.TransactionTypeExpense,
	)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}

	if h.checkBudgetAlertsUseCase != nil {
		h.checkBudgetAlertsUseCase.Execute(c.Request.Context(), userID, req.CategoryID)
	}

	SuccessResponse(c, http.StatusOK, gin.H{
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
		ValidationErrorResponse(c, err.Error())
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
		domainFinance.TransactionTypeIncome,
	)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
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
	if walletIDParam := c.Query("wallet_id"); walletIDParam != "" {
		if walletID, err := strconv.Atoi(walletIDParam); err == nil {
			req.WalletID = &walletID
		}
	}

	response, err := h.getAnalyticsUseCase.Execute(c.Request.Context(), userID, req)
	if err != nil {
		InternalServerErrorResponse(c, "FETCH_ANALYTICS_ERROR", "Failed to fetch analytics")
		return
	}

	SuccessResponse(c, http.StatusOK, response)
}

// CreateBudget handles budget creation
func (h *FinanceHandlers) CreateBudget(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req finance.CreateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, err.Error())
		return
	}

	response, err := h.createBudgetUseCase.Execute(c.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, entitlement.ErrPremiumRequired) {
			PremiumRequiredResponse(c, err)
			return
		}
		HandleError(c, err, http.StatusBadRequest)
		return
	}

	SuccessResponse(c, http.StatusCreated, response)
}

// GetBudgets handles getting budgets
func (h *FinanceHandlers) GetBudgets(c *gin.Context) {
	userID := c.GetInt("user_id")

	response, err := h.getBudgetsUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		InternalServerErrorResponse(c, "FETCH_BUDGETS_ERROR", "Failed to fetch budgets")
		return
	}

	SuccessResponse(c, http.StatusOK, response.Budgets)
}

// UpdateBudget handles budget updates
func (h *FinanceHandlers) UpdateBudget(c *gin.Context) {
	userID := c.GetInt("user_id")
	budgetID := c.Param("id")

	var req finance.UpdateBudgetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, err.Error())
		return
	}

	response, err := h.updateBudgetUseCase.Execute(
		c.Request.Context(),
		budgetID,
		userID,
		req,
	)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}

	SuccessResponse(c, http.StatusOK, response)
}

// DeleteBudget handles budget deletion
func (h *FinanceHandlers) DeleteBudget(c *gin.Context) {
	userID := c.GetInt("user_id")
	budgetID := c.Param("id")

	// Delete the budget
	err := h.deleteBudgetUseCase.Execute(c.Request.Context(), budgetID, userID)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Budget deleted successfully",
	})
}

// GetCurrencies handles getting currencies
func (h *FinanceHandlers) GetCurrencies(c *gin.Context) {
	userID := c.GetInt("user_id")

	response, err := h.getCurrenciesUseCase.Execute(c.Request.Context(), domainFinance.NewUserID(userID))
	if err != nil {
		InternalServerErrorResponse(c, "FETCH_CURRENCIES_ERROR", "Failed to fetch currencies")
		return
	}

	SuccessResponse(c, http.StatusOK, response.Currencies)
}

// CreateCurrency handles currency creation
func (h *FinanceHandlers) CreateCurrency(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req finance.CreateCurrencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, err.Error())
		return
	}

	response, err := h.createCurrencyUseCase.Execute(c.Request.Context(), domainFinance.NewUserID(userID), req)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}

	SuccessResponse(c, http.StatusCreated, gin.H{
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
		BadRequestResponse(c, "INVALID_CURRENCY_ID", "Invalid currency ID")
		return
	}

	var req finance.UpdateCurrencyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, err.Error())
		return
	}

	response, err := h.updateCurrencyUseCase.Execute(c.Request.Context(), domainFinance.NewUserID(userID), domainFinance.NewCurrencyID(currencyIDInt), req)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
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
		BadRequestResponse(c, "INVALID_CURRENCY_ID", "Invalid currency ID")
		return
	}

	response, err := h.deleteCurrencyUseCase.Execute(c.Request.Context(), domainFinance.NewUserID(userID), domainFinance.NewCurrencyID(currencyIDInt))
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
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
		HandleError(c, err, http.StatusBadRequest)
		return
	}

	SuccessResponse(c, http.StatusOK, gin.H{
		"message": "Default currency set successfully",
	})
}

// GetDefaultCurrency handles getting the default currency for a user
func (h *FinanceHandlers) GetDefaultCurrency(c *gin.Context) {
	userID := c.GetInt("user_id")

	// Get the default currency
	currency, err := h.getDefaultCurrencyUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}

	SuccessResponse(c, http.StatusOK, currency)
}

// GetRecurringTransactions lists recurring rules and enqueues due items as pending
func (h *FinanceHandlers) GetRecurringTransactions(c *gin.Context) {
	userID := c.GetInt("user_id")
	response, err := h.getRecurringUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusOK, response)
}

// CreateRecurringTransaction creates a recurring rule
func (h *FinanceHandlers) CreateRecurringTransaction(c *gin.Context) {
	userID := c.GetInt("user_id")
	var req finance.CreateRecurringTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, err.Error())
		return
	}
	response, err := h.createRecurringUseCase.Execute(c.Request.Context(), userID, req)
	if err != nil {
		if errors.Is(err, entitlement.ErrPremiumRequired) {
			PremiumRequiredResponse(c, err)
			return
		}
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusCreated, gin.H{
		"message":               "Recurring transaction created successfully",
		"recurring_transaction": response,
	})
}

// DeleteRecurringTransaction deletes a recurring rule
func (h *FinanceHandlers) DeleteRecurringTransaction(c *gin.Context) {
	userID := c.GetInt("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ValidationErrorResponse(c, "invalid recurring transaction id")
		return
	}
	if err := h.deleteRecurringUseCase.Execute(c.Request.Context(), userID, id); err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"message": "Recurring transaction deleted"})
}

// GetPendingTransactions lists open pending recurring occurrences (also enqueues dues)
func (h *FinanceHandlers) GetPendingTransactions(c *gin.Context) {
	userID := c.GetInt("user_id")
	response, err := h.listPendingUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusOK, response)
}

// ConfirmPendingTransaction confirms a pending occurrence into a real transaction
func (h *FinanceHandlers) ConfirmPendingTransaction(c *gin.Context) {
	userID := c.GetInt("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ValidationErrorResponse(c, "invalid pending transaction id")
		return
	}
	response, err := h.confirmPendingUseCase.Execute(c.Request.Context(), userID, id)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	if h.checkBudgetAlertsUseCase != nil && response != nil {
		h.checkBudgetAlertsUseCase.Execute(c.Request.Context(), userID, response.CategoryID)
	}
	SuccessResponse(c, http.StatusOK, gin.H{
		"message":             "Pending transaction confirmed",
		"pending_transaction": response,
	})
}

// RejectPendingTransaction rejects a pending occurrence without creating a transaction
func (h *FinanceHandlers) RejectPendingTransaction(c *gin.Context) {
	userID := c.GetInt("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ValidationErrorResponse(c, "invalid pending transaction id")
		return
	}
	response, err := h.rejectPendingUseCase.Execute(c.Request.Context(), userID, id)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{
		"message":             "Pending transaction rejected",
		"pending_transaction": response,
	})
}

func (h *FinanceHandlers) GetHealthScore(c *gin.Context) {
	userID := c.GetInt("user_id")
	response, err := h.getHealthScoreUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		InternalServerErrorResponse(c, "HEALTH_SCORE_ERROR", "Failed to compute health score")
		return
	}
	SuccessResponse(c, http.StatusOK, response)
}

func (h *FinanceHandlers) GetWalletSummary(c *gin.Context) {
	userID := c.GetInt("user_id")
	response, err := h.getWalletSummaryUseCase.Execute(c.Request.Context(), userID)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusOK, response)
}

func (h *FinanceHandlers) CreateWallet(c *gin.Context) {
	userID := c.GetInt("user_id")
	var req finance.CreateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, err.Error())
		return
	}
	response, err := h.createWalletUseCase.Execute(c.Request.Context(), userID, req)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusCreated, gin.H{"wallet": response})
}

func (h *FinanceHandlers) GetWallets(c *gin.Context) {
	userID := c.GetInt("user_id")
	includeArchived := c.Query("include_archived") == "true" || c.Query("include_archived") == "1"
	response, err := h.getWalletsUseCase.Execute(c.Request.Context(), userID, includeArchived)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"wallets": response})
}

func (h *FinanceHandlers) GetWallet(c *gin.Context) {
	userID := c.GetInt("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ValidationErrorResponse(c, "invalid wallet id")
		return
	}
	response, err := h.getWalletUseCase.Execute(c.Request.Context(), userID, id)
	if err != nil {
		HandleError(c, err, http.StatusNotFound)
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"wallet": response})
}

func (h *FinanceHandlers) UpdateWallet(c *gin.Context) {
	userID := c.GetInt("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ValidationErrorResponse(c, "invalid wallet id")
		return
	}
	var req finance.UpdateWalletRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, err.Error())
		return
	}
	response, err := h.updateWalletUseCase.Execute(c.Request.Context(), userID, id, req)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"wallet": response})
}

func (h *FinanceHandlers) SetDefaultWallet(c *gin.Context) {
	userID := c.GetInt("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ValidationErrorResponse(c, "invalid wallet id")
		return
	}
	response, err := h.setDefaultWalletUseCase.Execute(c.Request.Context(), userID, id)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"wallet": response})
}

func (h *FinanceHandlers) ArchiveWallet(c *gin.Context) {
	userID := c.GetInt("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ValidationErrorResponse(c, "invalid wallet id")
		return
	}
	response, err := h.archiveWalletUseCase.Execute(c.Request.Context(), userID, id)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"wallet": response})
}

func (h *FinanceHandlers) UnarchiveWallet(c *gin.Context) {
	userID := c.GetInt("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ValidationErrorResponse(c, "invalid wallet id")
		return
	}
	response, err := h.unarchiveWalletUseCase.Execute(c.Request.Context(), userID, id)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"wallet": response})
}

func (h *FinanceHandlers) GetWalletBalance(c *gin.Context) {
	userID := c.GetInt("user_id")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ValidationErrorResponse(c, "invalid wallet id")
		return
	}
	response, err := h.getWalletBalanceUseCase.Execute(c.Request.Context(), userID, id)
	if err != nil {
		HandleError(c, err, http.StatusNotFound)
		return
	}
	SuccessResponse(c, http.StatusOK, response)
}

func (h *FinanceHandlers) CreateTransfer(c *gin.Context) {
	userID := c.GetInt("user_id")
	var req finance.CreateTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ValidationErrorResponse(c, err.Error())
		return
	}
	response, err := h.createTransferUseCase.Execute(c.Request.Context(), userID, req)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusCreated, gin.H{"transfer": response})
}

func (h *FinanceHandlers) GetTransfers(c *gin.Context) {
	userID := c.GetInt("user_id")
	var walletID *int
	if walletIDParam := c.Query("wallet_id"); walletIDParam != "" {
		if id, err := strconv.Atoi(walletIDParam); err == nil {
			walletID = &id
		}
	}
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	var startPtr, endPtr *string
	if startDate != "" {
		startPtr = &startDate
	}
	if endDate != "" {
		endPtr = &endDate
	}
	response, err := h.getTransfersUseCase.Execute(c.Request.Context(), userID, walletID, startPtr, endPtr)
	if err != nil {
		HandleError(c, err, http.StatusBadRequest)
		return
	}
	SuccessResponse(c, http.StatusOK, gin.H{"transfers": response})
}
