package application

import (
	"net/http"
	appFinance "panda-pocket/internal/application/finance"
	appIdentity "panda-pocket/internal/application/identity"
	domainFinance "panda-pocket/internal/domain/finance"
	domainIdentity "panda-pocket/internal/domain/identity"
	"panda-pocket/internal/infrastructure/database"
	"panda-pocket/internal/interfaces/http/handlers"
	v100 "panda-pocket/internal/interfaces/http/handlers/v100"
	v120 "panda-pocket/internal/interfaces/http/handlers/v120"
	"panda-pocket/internal/interfaces/http/middleware"
	"panda-pocket/internal/interfaces/http/versioning"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// App represents the application with all its dependencies
type App struct {
	DB                  *gorm.DB
	IdentityHandlers    *handlers.IdentityHandlers
	FinanceHandlers     *handlers.FinanceHandlers
	FinanceHandlersV100 *v100.FinanceHandlersV100
	FinanceHandlersV120 *v120.FinanceHandlersV120
	DeprecationHandler  *handlers.DeprecationHandler
	AuthMiddleware      *middleware.AuthMiddleware
	VersionMiddleware   *middleware.VersionMiddleware
	VersionManager      *versioning.VersionManager
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

	// Version-specific handlers
	financeHandlersV100 := v100.NewFinanceHandlersV100(
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

	financeHandlersV120 := v120.NewFinanceHandlersV120(
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

	// Version management
	versionManager := versioning.NewVersionManager()
	versionMiddleware := middleware.NewVersionMiddleware()
	deprecationHandler := handlers.NewDeprecationHandler(versionManager)
	authMiddleware := middleware.NewAuthMiddleware(tokenService)

	return &App{
		DB:                  db,
		IdentityHandlers:    identityHandlers,
		FinanceHandlers:     financeHandlers,
		FinanceHandlersV100: financeHandlersV100,
		FinanceHandlersV120: financeHandlersV120,
		DeprecationHandler:  deprecationHandler,
		AuthMiddleware:      authMiddleware,
		VersionMiddleware:   versionMiddleware,
		VersionManager:      versionManager,
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
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-API-Version"}
	config.AllowCredentials = false
	r.Use(cors.New(config))

	// Version middleware
	r.Use(app.VersionMiddleware.ExtractVersion())
	r.Use(app.VersionMiddleware.ValidateVersion())
	r.Use(app.VersionMiddleware.AddDeprecationWarning())

	// Legacy routes (for backward compatibility)
	legacy := r.Group("/api")
	{
		// Auth routes
		auth := legacy.Group("/auth")
		{
			auth.POST("/register", app.IdentityHandlers.Register)
			auth.POST("/login", app.IdentityHandlers.Login)
			auth.POST("/logout", app.IdentityHandlers.Logout)
		}

		// Protected routes
		protected := legacy.Group("")
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

	// Versioned routes
	versioned := r.Group("/api")
	{
		// v120 routes (latest)
		v120 := versioned.Group("/v120")
		{
			// Auth routes
			auth := v120.Group("/auth")
			{
				auth.POST("/register", app.IdentityHandlers.Register)
				auth.POST("/login", app.IdentityHandlers.Login)
				auth.POST("/logout", app.IdentityHandlers.Logout)
			}

			// Protected routes
			protected := v120.Group("")
			protected.Use(app.AuthMiddleware.RequireAuth())
			{
				// Transactions (enhanced with analytics)
				protected.GET("/transactions", app.FinanceHandlersV120.GetTransactions)
				protected.POST("/transactions", app.FinanceHandlersV120.CreateTransaction)
				protected.PUT("/transactions/:id", app.FinanceHandlersV120.UpdateTransaction)
				protected.DELETE("/transactions/:id", app.FinanceHandlersV120.DeleteTransaction)
				protected.GET("/transactions/analytics", app.FinanceHandlersV120.GetTransactionsWithAnalytics)

				// Categories
				protected.GET("/categories", app.FinanceHandlersV120.GetCategories)
				protected.POST("/categories", app.FinanceHandlersV120.CreateCategory)
				protected.PUT("/categories/:id", app.FinanceHandlersV120.UpdateCategory)
				protected.DELETE("/categories/:id", app.FinanceHandlersV120.DeleteCategory)

				// Budgets
				protected.GET("/budgets", app.FinanceHandlersV120.GetBudgets)
				protected.POST("/budgets", app.FinanceHandlersV120.CreateBudget)
				protected.PUT("/budgets/:id", app.FinanceHandlersV120.UpdateBudget)
				protected.DELETE("/budgets/:id", app.FinanceHandlersV120.DeleteBudget)

				// Currencies
				protected.GET("/currencies", app.FinanceHandlersV120.GetCurrencies)
				protected.POST("/currencies", app.FinanceHandlersV120.CreateCurrency)
				protected.GET("/currencies/default", app.FinanceHandlersV120.GetDefaultCurrency)
				protected.PUT("/currencies/:id/set-default", app.FinanceHandlersV120.SetDefaultCurrency)
				protected.PUT("/currencies/:id", app.FinanceHandlersV120.UpdateCurrency)
				protected.DELETE("/currencies/:id", app.FinanceHandlersV120.DeleteCurrency)

				// Analytics (enhanced)
				protected.GET("/analytics", app.FinanceHandlersV120.GetAnalytics)
			}
		}

		// v100 routes (legacy - deprecated)
		v100 := versioned.Group("/v100")
		{
			// Auth routes
			auth := v100.Group("/auth")
			{
				auth.POST("/register", app.IdentityHandlers.Register)
				auth.POST("/login", app.IdentityHandlers.Login)
				auth.POST("/logout", app.IdentityHandlers.Logout)
			}

			// Protected routes
			protected := v100.Group("")
			protected.Use(app.AuthMiddleware.RequireAuth())
			{
				// Expenses (legacy)
				protected.GET("/expenses", app.FinanceHandlersV100.GetExpenses)
				protected.POST("/expenses", app.FinanceHandlersV100.CreateExpense)
				protected.PUT("/expenses/:id", app.FinanceHandlersV100.UpdateExpense)
				protected.DELETE("/expenses/:id", app.FinanceHandlersV100.DeleteExpense)

				// Incomes (legacy)
				protected.GET("/incomes", app.FinanceHandlersV100.GetIncomes)
				protected.POST("/incomes", app.FinanceHandlersV100.CreateIncome)
				protected.PUT("/incomes/:id", app.FinanceHandlersV100.UpdateIncome)
				protected.DELETE("/incomes/:id", app.FinanceHandlersV100.DeleteIncome)

				// All Transactions (legacy)
				protected.GET("/transactions", app.FinanceHandlersV100.GetAllTransactions)

				// Categories (legacy)
				protected.GET("/categories", app.FinanceHandlersV100.GetCategories)
				protected.POST("/categories", app.FinanceHandlersV100.CreateCategory)
				protected.PUT("/categories/:id", app.FinanceHandlersV100.UpdateCategory)
				protected.DELETE("/categories/:id", app.FinanceHandlersV100.DeleteCategory)

				// Budgets (legacy)
				protected.GET("/budgets", app.FinanceHandlersV100.GetBudgets)
				protected.POST("/budgets", app.FinanceHandlersV100.CreateBudget)
				protected.PUT("/budgets/:id", app.FinanceHandlersV100.UpdateBudget)
				protected.DELETE("/budgets/:id", app.FinanceHandlersV100.DeleteBudget)

				// Currencies (legacy)
				protected.GET("/currencies", app.FinanceHandlersV100.GetCurrencies)
				protected.POST("/currencies", app.FinanceHandlersV100.CreateCurrency)
				protected.GET("/currencies/default", app.FinanceHandlersV100.GetDefaultCurrency)
				protected.PUT("/currencies/:id/set-default", app.FinanceHandlersV100.SetDefaultCurrency)
				protected.PUT("/currencies/:id", app.FinanceHandlersV100.UpdateCurrency)
				protected.DELETE("/currencies/:id", app.FinanceHandlersV100.DeleteCurrency)

				// Analytics (legacy)
				protected.GET("/analytics", app.FinanceHandlersV100.GetAnalytics)
			}
		}

		// Version management routes
		version := versioned.Group("/version")
		{
			version.GET("/info/:version", app.DeprecationHandler.GetDeprecationInfo)
			version.GET("/status/:version", app.DeprecationHandler.GetVersionStatus)
			version.GET("/matrix", app.DeprecationHandler.GetVersionMatrix)
			version.GET("/migration/:version", app.DeprecationHandler.GetMigrationPath)
			version.GET("/compare", app.DeprecationHandler.CompareVersions)
			version.GET("/features/:version", app.DeprecationHandler.GetVersionFeatures)
			version.GET("/validate", app.DeprecationHandler.ValidateVersionTransition)
			version.GET("/timeline", app.DeprecationHandler.GetDeprecationTimeline)
			version.GET("/recommendations/:version", app.DeprecationHandler.GetUpgradeRecommendations)
		}
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return r
}
