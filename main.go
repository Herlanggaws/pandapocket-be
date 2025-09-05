package main

import (
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	db, err := initDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Create default categories and currencies
	createDefaultCategories(db)
	createDefaultCurrencies(db)

	// Setup Gin router
	r := gin.Default()

	// CORS configuration - Allow all localhost ports
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"http://localhost:3000",
		"http://localhost:3001",
		"http://localhost:3002",
		"http://localhost:3003",
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.AllowCredentials = false
	r.Use(cors.New(config))

	// API routes
	api := r.Group("/api")
	{
		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", registerHandler(db))
			auth.POST("/login", loginHandler(db))
			auth.POST("/logout", logoutHandler())
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(authMiddleware())
		{
			// Categories
			protected.GET("/categories", getCategoriesHandler(db))
			protected.POST("/categories", createCategoryHandler(db))

			// Expenses
			protected.GET("/expenses", getExpensesHandler(db))
			protected.POST("/expenses", createExpenseHandler(db))
			protected.DELETE("/expenses/:id", deleteExpenseHandler(db))

			// Incomes
			protected.GET("/incomes", getIncomesHandler(db))
			protected.POST("/incomes", createIncomeHandler(db))
			protected.DELETE("/incomes/:id", deleteIncomeHandler(db))
		}

		// Currency routes (separate group for testing)
		currencies := api.Group("")
		currencies.Use(authMiddleware())
		{
			currencies.GET("/currencies", getCurrenciesHandler(db))
			currencies.POST("/currencies", createCurrencyHandler(db))
			currencies.PUT("/currencies/:id", updateCurrencyHandler(db))
			currencies.DELETE("/currencies/:id", deleteCurrencyHandler(db))
		}

		// Phase 2 routes
		phase2 := api.Group("")
		phase2.Use(authMiddleware())
		{
			// Budgets
			phase2.GET("/budgets", getBudgetsHandler(db))
			phase2.POST("/budgets", createBudgetHandler(db))
			phase2.PUT("/budgets/:id", updateBudgetHandler(db))
			phase2.DELETE("/budgets/:id", deleteBudgetHandler(db))

			// Recurring Transactions
			phase2.GET("/recurring-transactions", getRecurringTransactionsHandler(db))
			phase2.POST("/recurring-transactions", createRecurringTransactionHandler(db))
			phase2.DELETE("/recurring-transactions/:id", deleteRecurringTransactionHandler(db))

			// Analytics
			phase2.GET("/analytics", getAnalyticsHandler(db))

			// User Preferences
			phase2.GET("/preferences", getUserPreferencesHandler(db))
			phase2.PUT("/preferences", updateUserPreferencesHandler(db))

			// Notifications
			phase2.GET("/notifications", getNotificationsHandler(db))
			phase2.PUT("/notifications/:id/read", markNotificationReadHandler(db))
			phase2.DELETE("/notifications/:id", deleteNotificationHandler(db))
		}
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	log.Println("Server starting on :8080")
	r.Run(":8080")
}
