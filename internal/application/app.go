package application

import (
	"database/sql"
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
)

// App represents the application with all its dependencies
type App struct {
	DB               *sql.DB
	IdentityHandlers *handlers.IdentityHandlers
	FinanceHandlers  *handlers.FinanceHandlers
	AuthMiddleware   *middleware.AuthMiddleware
}

// NewApp creates a new application instance with all dependencies wired up
func NewApp(db *sql.DB) *App {
	// Infrastructure layer - repositories (PostgreSQL only)
	userRepo := database.NewPostgresUserRepository(db)
	categoryRepo := database.NewPostgresCategoryRepository(db)
	currencyRepo := database.NewPostgresCurrencyRepository(db)
	transactionRepo := database.NewPostgresTransactionRepository(db)
	budgetRepo := database.NewPostgresBudgetRepository(db)

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
	createCategoryUseCase := appFinance.NewCreateCategoryUseCase(categoryService)
	getCategoriesUseCase := appFinance.NewGetCategoriesUseCase(categoryService)
	getAnalyticsUseCase := appFinance.NewGetAnalyticsUseCase(transactionService)
	createBudgetUseCase := appFinance.NewCreateBudgetUseCase(budgetService, currencyService)
	getBudgetsUseCase := appFinance.NewGetBudgetsUseCase(budgetService)

	// Interface layer - handlers and middleware
	identityHandlers := handlers.NewIdentityHandlers(registerUserUseCase, loginUserUseCase)
	financeHandlers := handlers.NewFinanceHandlers(
		createTransactionUseCase,
		getTransactionsUseCase,
		createCategoryUseCase,
		getCategoriesUseCase,
		getAnalyticsUseCase,
		createBudgetUseCase,
		getBudgetsUseCase,
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

			// Expenses
			protected.GET("/expenses", app.FinanceHandlers.GetExpenses)
			protected.POST("/expenses", app.FinanceHandlers.CreateExpense)
			protected.DELETE("/expenses/:id", app.FinanceHandlers.DeleteExpense)

			// Incomes
			protected.GET("/incomes", app.FinanceHandlers.GetIncomes)
			protected.POST("/incomes", app.FinanceHandlers.CreateIncome)
			protected.DELETE("/incomes/:id", app.FinanceHandlers.DeleteIncome)

			// Budgets
			protected.GET("/budgets", app.FinanceHandlers.GetBudgets)
			protected.POST("/budgets", app.FinanceHandlers.CreateBudget)

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
