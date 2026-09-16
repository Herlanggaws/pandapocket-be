package application

import (
	"net/http"
	appFinance "panda-pocket/internal/application/finance"
	appIdentity "panda-pocket/internal/application/identity"
	appNotification "panda-pocket/internal/application/notification"
	domainFinance "panda-pocket/internal/domain/finance"
	domainIdentity "panda-pocket/internal/domain/identity"
	"panda-pocket/internal/infrastructure/database"
	"panda-pocket/internal/infrastructure/notification"
	"panda-pocket/internal/interfaces/http/handlers"
	"panda-pocket/internal/interfaces/http/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// App represents the application with all its dependencies
type App struct {
	DB                   *gorm.DB
	IdentityHandlers     *handlers.IdentityHandlers
	FinanceHandlers      *handlers.FinanceHandlers
	DashboardHandlers    *handlers.DashboardHandlers
	NotificationHandlers *handlers.NotificationHandlers
	AuthMiddleware       *middleware.AuthMiddleware
}

// NewApp creates a new application instance with all dependencies wired up
func NewApp(db *gorm.DB) *App {
	// Infrastructure layer - repositories (GORM)
	userRepo := database.NewGormUserRepository(db)
	categoryRepo := database.NewGormCategoryRepository(db)
	currencyRepo := database.NewGormCurrencyRepository(db)
	transactionRepo := database.NewGormTransactionRepository(db)
	budgetRepo := database.NewGormBudgetRepository(db)
	tokenRepo := database.NewGormPasswordResetTokenRepository(db)
	authTokenRepo := database.NewGormTokenRepository(db)
	prefsRepo := database.NewGormPreferencesRepository(db)
	notificationRepo := database.NewGormNotificationRepository(db)
	recurringRepo := database.NewGormRecurringTransactionRepository(db)

	// Domain layer - services
	userService := domainIdentity.NewUserService(userRepo)
	transactionService := domainFinance.NewTransactionService(transactionRepo, categoryRepo, currencyRepo)
	categoryService := domainFinance.NewCategoryService(categoryRepo, budgetRepo)
	currencyService := domainFinance.NewCurrencyService(currencyRepo)
	budgetService := domainFinance.NewBudgetService(budgetRepo, categoryRepo)

	// Application layer - use cases
	tokenService := appIdentity.NewTokenService(authTokenRepo)
	emailService := notification.NewSMTPEmailService()
	notificationHelper := appNotification.NewCreateNotificationHelper(notificationRepo)
	registerUserUseCase := appIdentity.NewRegisterUserUseCase(userService, tokenService)
	loginUserUseCase := appIdentity.NewLoginUserUseCase(userService, tokenService)
	getUsersUseCase := appIdentity.NewGetUsersUseCase(userService)
	forgotPasswordUseCase := appIdentity.NewForgotPasswordUseCase(userRepo, tokenRepo, emailService)
	resetPasswordUseCase := appIdentity.NewResetPasswordUseCase(userRepo, tokenRepo)
	refreshTokenUseCase := appIdentity.NewRefreshTokenUseCase(userService, tokenService)
	changePasswordUseCase := appIdentity.NewChangePasswordUseCase(userRepo)
	getPreferencesUseCase := appIdentity.NewGetPreferencesUseCase(prefsRepo, prefsRepo)
	updatePreferencesUseCase := appIdentity.NewUpdatePreferencesUseCase(prefsRepo, prefsRepo)
	getNotificationsUseCase := appNotification.NewGetNotificationsUseCase(notificationRepo)
	markNotificationReadUseCase := appNotification.NewMarkNotificationReadUseCase(notificationRepo)
	deleteNotificationUseCase := appNotification.NewDeleteNotificationUseCase(notificationRepo)
	getDashboardStatsUseCase := appIdentity.NewGetDashboardStatsUseCase(userRepo, budgetRepo, transactionRepo)
	createTransactionUseCase := appFinance.NewCreateTransactionUseCase(transactionService, currencyService)
	getTransactionsUseCase := appFinance.NewGetTransactionsUseCase(transactionService, categoryService)
	getAllTransactionsUseCase := appFinance.NewGetAllTransactionsUseCase(transactionService, categoryService)
	updateTransactionUseCase := appFinance.NewUpdateTransactionUseCase(transactionService)
	deleteTransactionUseCase := appFinance.NewDeleteTransactionUseCase(transactionService)
	createCategoryUseCase := appFinance.NewCreateCategoryUseCase(categoryService)
	updateCategoryUseCase := appFinance.NewUpdateCategoryUseCase(categoryService)
	deleteCategoryUseCase := appFinance.NewDeleteCategoryUseCase(categoryService)
	getCategoriesUseCase := appFinance.NewGetCategoriesUseCase(categoryService)
	getAnalyticsUseCase := appFinance.NewGetAnalyticsUseCase(transactionService, categoryService)
	createBudgetUseCase := appFinance.NewCreateBudgetUseCase(budgetService, currencyService, categoryService, transactionService)
	getBudgetsUseCase := appFinance.NewGetBudgetsUseCase(budgetService, categoryService, transactionService)
	updateBudgetUseCase := appFinance.NewUpdateBudgetUseCase(budgetService, categoryService, transactionService)
	deleteBudgetUseCase := appFinance.NewDeleteBudgetUseCase(budgetService)
	createCurrencyUseCase := appFinance.NewCreateCurrencyUseCase(currencyService)
	getCurrenciesUseCase := appFinance.NewGetCurrenciesUseCase(currencyService)
	updateCurrencyUseCase := appFinance.NewUpdateCurrencyUseCase(currencyService)
	deleteCurrencyUseCase := appFinance.NewDeleteCurrencyUseCase(currencyService)
	setDefaultCurrencyUseCase := appFinance.NewSetDefaultCurrencyUseCase(currencyService)
	getDefaultCurrencyUseCase := appFinance.NewGetDefaultCurrencyUseCase(currencyService)
	checkBudgetAlertsUseCase := appFinance.NewCheckBudgetAlertsUseCase(
		budgetService,
		categoryService,
		transactionService,
		prefsRepo,
		notificationHelper,
	)
	createRecurringUseCase := appFinance.NewCreateRecurringTransactionUseCase(recurringRepo, currencyService, categoryService)
	getRecurringUseCase := appFinance.NewGetRecurringTransactionsUseCase(
		recurringRepo,
		transactionService,
		categoryService,
		prefsRepo,
		notificationHelper,
	)
	deleteRecurringUseCase := appFinance.NewDeleteRecurringTransactionUseCase(recurringRepo)

	// Interface layer - handlers and middleware
	identityHandlers := handlers.NewIdentityHandlers(
		registerUserUseCase,
		loginUserUseCase,
		getUsersUseCase,
		forgotPasswordUseCase,
		resetPasswordUseCase,
		refreshTokenUseCase,
		tokenService,
		changePasswordUseCase,
		getPreferencesUseCase,
		updatePreferencesUseCase,
	)
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
		checkBudgetAlertsUseCase,
		createRecurringUseCase,
		getRecurringUseCase,
		deleteRecurringUseCase,
	)
	dashboardHandlers := handlers.NewDashboardHandlers(getDashboardStatsUseCase)
	notificationHandlers := handlers.NewNotificationHandlers(
		getNotificationsUseCase,
		markNotificationReadUseCase,
		deleteNotificationUseCase,
	)
	authMiddleware := middleware.NewAuthMiddleware(tokenService)

	return &App{
		DB:                   db,
		IdentityHandlers:     identityHandlers,
		FinanceHandlers:      financeHandlers,
		DashboardHandlers:    dashboardHandlers,
		NotificationHandlers: notificationHandlers,
		AuthMiddleware:       authMiddleware,
	}
}

// SetupRoutes sets up all the HTTP routes
func (app *App) SetupRoutes() *gin.Engine {
	r := gin.Default()

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"*",
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.AllowCredentials = false
	r.Use(cors.New(config))

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", app.IdentityHandlers.Register)
			auth.POST("/login", app.IdentityHandlers.Login)
			auth.POST("/refresh", app.IdentityHandlers.RefreshToken)
			auth.POST("/logout", app.IdentityHandlers.Logout)
			auth.POST("/forgot", app.IdentityHandlers.ForgotPassword)
			auth.POST("/reset-password", app.IdentityHandlers.ResetPassword)

			auth.POST("/change-password", app.AuthMiddleware.RequireAuth(), app.IdentityHandlers.ChangePassword)
		}

		protected := api.Group("")
		protected.Use(app.AuthMiddleware.RequireAuth())
		{
			adminOnly := protected.Group("")
			adminOnly.Use(app.AuthMiddleware.RequireRole("admin"))
			{
				adminOnly.GET("/users", app.IdentityHandlers.GetUsers)
				adminOnly.GET("/dashboard/stats", app.DashboardHandlers.GetDashboardStats)
			}

			protected.GET("/categories", app.FinanceHandlers.GetCategories)
			protected.POST("/categories", app.FinanceHandlers.CreateCategory)
			protected.PUT("/categories/:id", app.FinanceHandlers.UpdateCategory)
			protected.DELETE("/categories/:id", app.FinanceHandlers.DeleteCategory)

			protected.GET("/expenses", app.FinanceHandlers.GetExpenses)
			protected.POST("/expenses", app.FinanceHandlers.CreateExpense)
			protected.PUT("/expenses/:id", app.FinanceHandlers.UpdateExpense)
			protected.DELETE("/expenses/:id", app.FinanceHandlers.DeleteExpense)

			protected.GET("/incomes", app.FinanceHandlers.GetIncomes)
			protected.POST("/incomes", app.FinanceHandlers.CreateIncome)
			protected.PUT("/incomes/:id", app.FinanceHandlers.UpdateIncome)
			protected.DELETE("/incomes/:id", app.FinanceHandlers.DeleteIncome)

			protected.GET("/transactions", app.FinanceHandlers.GetAllTransactions)

			protected.GET("/budgets", app.FinanceHandlers.GetBudgets)
			protected.POST("/budgets", app.FinanceHandlers.CreateBudget)
			protected.PUT("/budgets/:id", app.FinanceHandlers.UpdateBudget)
			protected.DELETE("/budgets/:id", app.FinanceHandlers.DeleteBudget)

			protected.GET("/currencies", app.FinanceHandlers.GetCurrencies)
			protected.POST("/currencies", app.FinanceHandlers.CreateCurrency)
			protected.GET("/currencies/default", app.FinanceHandlers.GetDefaultCurrency)
			protected.PUT("/currencies/:id/set-default", app.FinanceHandlers.SetDefaultCurrency)
			protected.PUT("/currencies/:id", app.FinanceHandlers.UpdateCurrency)
			protected.DELETE("/currencies/:id", app.FinanceHandlers.DeleteCurrency)

			protected.GET("/analytics", app.FinanceHandlers.GetAnalytics)

			protected.GET("/preferences", app.IdentityHandlers.GetPreferences)
			protected.PUT("/preferences", app.IdentityHandlers.UpdatePreferences)

			protected.GET("/notifications", app.NotificationHandlers.GetNotifications)
			protected.PUT("/notifications/:id/read", app.NotificationHandlers.MarkNotificationRead)
			protected.DELETE("/notifications/:id", app.NotificationHandlers.DeleteNotification)

			protected.GET("/recurring-transactions", app.FinanceHandlers.GetRecurringTransactions)
			protected.POST("/recurring-transactions", app.FinanceHandlers.CreateRecurringTransaction)
			protected.DELETE("/recurring-transactions/:id", app.FinanceHandlers.DeleteRecurringTransaction)
		}
	}

	r.GET("/health", func(c *gin.Context) {
		handlers.SuccessResponse(c, http.StatusOK, gin.H{"status": "ok"})
	})

	return r
}
