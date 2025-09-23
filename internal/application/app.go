package application

import (
	"net/http"
	appFinance "panda-pocket/internal/application/finance"
	appIdentity "panda-pocket/internal/application/identity"
	domainFinance "panda-pocket/internal/domain/finance"
	domainIdentity "panda-pocket/internal/domain/identity"
	"panda-pocket/internal/infrastructure/database"
	"panda-pocket/internal/interfaces/http/handlers"
	"panda-pocket/internal/interfaces/http/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// App represents the application with all its dependencies
type App struct {
	DB               *gorm.DB
	IdentityHandlers *handlers.IdentityHandlers
	FinanceHandlers  *handlers.FinanceHandlers
	AuthMiddleware   *middleware.AuthMiddleware
}

// NewApp creates a new application instance with all dependencies wired up
func NewApp(db *gorm.DB) *App {
	// Infrastructure layer - repositories (GORM)
	userRepo := database.NewGormUserRepository(db)
	categoryRepo := database.NewGormCategoryRepository(db)
	currencyRepo := database.NewGormCurrencyRepository(db)
	transactionRepo := database.NewGormTransactionRepository(db)
	budgetRepo := database.NewGormBudgetRepository(db)

	// Domain layer - services
	userService := domainIdentity.NewUserService(userRepo)
	transactionService := domainFinance.NewTransactionService(transactionRepo, categoryRepo, currencyRepo)
	categoryService := domainFinance.NewCategoryService(categoryRepo)
	currencyService := domainFinance.NewCurrencyService(currencyRepo)
	budgetService := domainFinance.NewBudgetService(budgetRepo, categoryRepo)

	// Application layer - use cases
	tokenService := appIdentity.NewTokenService()
	registerUserUseCase := appIdentity.NewRegisterUserUseCase(userService, tokenService)
	loginUserUseCase := appIdentity.NewLoginUserUseCase(userService, tokenService)
	createTransactionUseCase := appFinance.NewCreateTransactionUseCase(transactionService, currencyService)
	getTransactionsUseCase := appFinance.NewGetTransactionsUseCase(transactionService)
	getAllTransactionsUseCase := appFinance.NewGetAllTransactionsUseCase(transactionService)
	updateTransactionUseCase := appFinance.NewUpdateTransactionUseCase(transactionService)
	deleteTransactionUseCase := appFinance.NewDeleteTransactionUseCase(transactionService)
	createCategoryUseCase := appFinance.NewCreateCategoryUseCase(categoryService)
	updateCategoryUseCase := appFinance.NewUpdateCategoryUseCase(categoryService)
	deleteCategoryUseCase := appFinance.NewDeleteCategoryUseCase(categoryService)
	getCategoriesUseCase := appFinance.NewGetCategoriesUseCase(categoryService)
	getAnalyticsUseCase := appFinance.NewGetAnalyticsUseCase(transactionService)
	createBudgetUseCase := appFinance.NewCreateBudgetUseCase(budgetService, currencyService)
	getBudgetsUseCase := appFinance.NewGetBudgetsUseCase(budgetService)
	updateBudgetUseCase := appFinance.NewUpdateBudgetUseCase(budgetService)
	deleteBudgetUseCase := appFinance.NewDeleteBudgetUseCase(budgetService)
	createCurrencyUseCase := appFinance.NewCreateCurrencyUseCase(currencyService)
	getCurrenciesUseCase := appFinance.NewGetCurrenciesUseCase(currencyService)
	updateCurrencyUseCase := appFinance.NewUpdateCurrencyUseCase(currencyService)
	deleteCurrencyUseCase := appFinance.NewDeleteCurrencyUseCase(currencyService)
	setDefaultCurrencyUseCase := appFinance.NewSetDefaultCurrencyUseCase(currencyService)
	getDefaultCurrencyUseCase := appFinance.NewGetDefaultCurrencyUseCase(currencyService)

	// Interface layer - handlers and middleware
	identityHandlers := handlers.NewIdentityHandlers(registerUserUseCase, loginUserUseCase)
	financeHandlers := handlers.NewFinanceHandlers(
		createTransactionUseCase,
		getTransactionsUseCase,
		getAllTransactionsUseCase,
		updateTransactionUseCase,
		deleteTransactionUseCase,
		createCategoryUseCase,
		updateCategoryUseCase,
		deleteCategoryUseCase,
		getCategoriesUseCase,
		getAnalyticsUseCase,
		createBudgetUseCase,
		getBudgetsUseCase,
		updateBudgetUseCase,
		deleteBudgetUseCase,
		createCurrencyUseCase,
		getCurrenciesUseCase,
		updateCurrencyUseCase,
		deleteCurrencyUseCase,
		setDefaultCurrencyUseCase,
		getDefaultCurrencyUseCase,
	)
	authMiddleware := middleware.NewAuthMiddleware(tokenService)

	return &App{
		DB:               db,
		IdentityHandlers: identityHandlers,
		FinanceHandlers:  financeHandlers,
		AuthMiddleware:   authMiddleware,
	}
}

// SetupRoutes sets up all the HTTP routes
func (app *App) SetupRoutes() *gin.Engine {
	r := gin.Default()

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"http://localhost:3000",
		"http://localhost:3001",
		"http://localhost:3002",
		"http://localhost:3003",
		"http://localhost:3004", // Back office port
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
			auth.POST("/register", app.IdentityHandlers.Register)
			auth.POST("/login", app.IdentityHandlers.Login)
			auth.POST("/logout", app.IdentityHandlers.Logout)
		}

		// Protected routes
		protected := api.Group("")
		protected.Use(app.AuthMiddleware.RequireAuth())
		{
			// Categories
			protected.GET("/categories", app.FinanceHandlers.GetCategories)
			protected.POST("/categories", app.FinanceHandlers.CreateCategory)
			protected.PUT("/categories/:id", app.FinanceHandlers.UpdateCategory)
			protected.DELETE("/categories/:id", app.FinanceHandlers.DeleteCategory)

			// Expenses
			protected.GET("/expenses", app.FinanceHandlers.GetExpenses)
			protected.POST("/expenses", app.FinanceHandlers.CreateExpense)
			protected.PUT("/expenses/:id", app.FinanceHandlers.UpdateExpense)
			protected.DELETE("/expenses/:id", app.FinanceHandlers.DeleteExpense)

			// Incomes
			protected.GET("/incomes", app.FinanceHandlers.GetIncomes)
			protected.POST("/incomes", app.FinanceHandlers.CreateIncome)
			protected.PUT("/incomes/:id", app.FinanceHandlers.UpdateIncome)
			protected.DELETE("/incomes/:id", app.FinanceHandlers.DeleteIncome)

			// All Transactions (with filters)
			protected.GET("/transactions", app.FinanceHandlers.GetAllTransactions)

			// Budgets
			protected.GET("/budgets", app.FinanceHandlers.GetBudgets)
			protected.POST("/budgets", app.FinanceHandlers.CreateBudget)
			protected.PUT("/budgets/:id", app.FinanceHandlers.UpdateBudget)
			protected.DELETE("/budgets/:id", app.FinanceHandlers.DeleteBudget)

			// Currencies
			protected.GET("/currencies", app.FinanceHandlers.GetCurrencies)
			protected.POST("/currencies", app.FinanceHandlers.CreateCurrency)
			protected.GET("/currencies/default", app.FinanceHandlers.GetDefaultCurrency)
			protected.PUT("/currencies/:id/set-default", app.FinanceHandlers.SetDefaultCurrency)
			protected.PUT("/currencies/:id", app.FinanceHandlers.UpdateCurrency)
			protected.DELETE("/currencies/:id", app.FinanceHandlers.DeleteCurrency)

			// Analytics
			protected.GET("/analytics", app.FinanceHandlers.GetAnalytics)
		}
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return r
}
