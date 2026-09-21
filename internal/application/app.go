package application

import (
	"context"
	"log"
	"net/http"
	"time"

	appBilling "panda-pocket/internal/application/billing"
	appFeedback "panda-pocket/internal/application/feedback"
	appFinance "panda-pocket/internal/application/finance"
	appIdentity "panda-pocket/internal/application/identity"
	appNotification "panda-pocket/internal/application/notification"
	appTicket "panda-pocket/internal/application/ticket"
	"panda-pocket/internal/domain/entitlement"
	domainFinance "panda-pocket/internal/domain/finance"
	domainIdentity "panda-pocket/internal/domain/identity"
	"panda-pocket/internal/infrastructure/database"
	"panda-pocket/internal/infrastructure/doit"
	"panda-pocket/internal/infrastructure/notification"
	"panda-pocket/internal/interfaces/http/handlers"
	"panda-pocket/internal/interfaces/http/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// App represents the application with all its dependencies
type App struct {
	DB                                 *gorm.DB
	IdentityHandlers                   *handlers.IdentityHandlers
	FinanceHandlers                    *handlers.FinanceHandlers
	ExportHandlers                     *handlers.ExportHandlers
	DashboardHandlers                  *handlers.DashboardHandlers
	NotificationHandlers               *handlers.NotificationHandlers
	FeedbackHandlers                   *handlers.FeedbackHandlers
	TicketHandlers                     *handlers.TicketHandlers
	BillingHandlers                    *handlers.BillingHandlers
	AuthMiddleware                     *middleware.AuthMiddleware
	purgeDeletedAccountsUseCase        *appIdentity.PurgeDeletedAccountsUseCase
	cleanupExpiredTokensUseCase        *appIdentity.CleanupExpiredTokensUseCase
	checkGoalDeadlineAlertsUseCase     *appFinance.CheckGoalDeadlineAlertsUseCase
	processBillingSubscriptionsUseCase *appBilling.ProcessBillingSubscriptionsUseCase
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
	subscriptionRepo := database.NewGormSubscriptionRepository(db)
	billingWebhookEventRepo := database.NewGormBillingWebhookEventRepository(db)
	notificationRepo := database.NewGormNotificationRepository(db)
	feedbackRepo := database.NewGormFeedbackRepository(db)
	ticketRepo := database.NewGormTicketRepository(db)
	walletRepo := database.NewGormWalletRepository(db)
	transferRepo := database.NewGormTransferRepository(db)
	goalRepo := database.NewGormGoalRepository(db)
	assetRepo := database.NewGormAssetRepository(db)
	liabilityRepo := database.NewGormLiabilityRepository(db)
	liabilityPaymentRepo := database.NewGormLiabilityPaymentRepository(db)
	healthSnapshotRepo := database.NewGormHealthScoreSnapshotRepository(db)
	recurringRepo := database.NewGormRecurringTransactionRepository(db)
	pendingRepo := database.NewGormPendingTransactionRepository(db)

	// Domain layer - services
	userService := domainIdentity.NewUserService(userRepo)
	transactionService := domainFinance.NewTransactionService(transactionRepo, categoryRepo, currencyRepo, walletRepo)
	categoryService := domainFinance.NewCategoryService(categoryRepo, budgetRepo)
	currencyService := domainFinance.NewCurrencyService(currencyRepo)
	budgetService := domainFinance.NewBudgetService(budgetRepo, categoryRepo)
	walletService := domainFinance.NewWalletService(walletRepo, currencyRepo)
	transferService := domainFinance.NewTransferService(transferRepo, walletRepo)
	goalService := domainFinance.NewGoalService(goalRepo)
	assetService := domainFinance.NewAssetService(assetRepo)
	liabilityService := domainFinance.NewLiabilityService(liabilityRepo, liabilityPaymentRepo)

	// Application layer - use cases
	tokenService := appIdentity.NewTokenService(authTokenRepo)
	emailService := notification.NewSMTPEmailService()
	notificationHelper := appNotification.NewCreateNotificationHelper(notificationRepo)
	registerUserUseCase := appIdentity.NewRegisterUserUseCase(userService, tokenService, subscriptionRepo)
	loginUserUseCase := appIdentity.NewLoginUserUseCase(userService, tokenService)
	getUsersUseCase := appIdentity.NewGetUsersUseCase(userService)
	forgotPasswordUseCase := appIdentity.NewForgotPasswordUseCase(userRepo, tokenRepo, emailService)
	resetPasswordUseCase := appIdentity.NewResetPasswordUseCase(userRepo, tokenRepo)
	refreshTokenUseCase := appIdentity.NewRefreshTokenUseCase(userService, tokenService)
	changePasswordUseCase := appIdentity.NewChangePasswordUseCase(userRepo)
	getPreferencesUseCase := appIdentity.NewGetPreferencesUseCase(prefsRepo, prefsRepo)
	updatePreferencesUseCase := appIdentity.NewUpdatePreferencesUseCase(prefsRepo, prefsRepo)
	accountResetChallengeRepo := database.NewGormAccountResetChallengeRepository(db)
	userDataWiper := database.NewGormUserDataWiper(db)
	resetAccountDataUseCase := appIdentity.NewResetAccountDataUseCase(accountResetChallengeRepo, userDataWiper)
	deleteAccountUseCase := appIdentity.NewDeleteAccountUseCase(userRepo, tokenService)
	purgeDeletedAccountsUseCase := appIdentity.NewPurgeDeletedAccountsUseCase(userRepo, userDataWiper)
	cleanupExpiredTokensUseCase := appIdentity.NewCleanupExpiredTokensUseCase(authTokenRepo, tokenRepo)
	getNotificationsUseCase := appNotification.NewGetNotificationsUseCase(notificationRepo)
	markNotificationReadUseCase := appNotification.NewMarkNotificationReadUseCase(notificationRepo)
	deleteNotificationUseCase := appNotification.NewDeleteNotificationUseCase(notificationRepo)
	submitFeedbackUseCase := appFeedback.NewSubmitFeedbackUseCase(feedbackRepo)
	entitlementChecker := entitlement.NewSubscriptionChecker(subscriptionRepo)
	doitClient := doit.NewClient()
	getSubscriptionUseCase := appBilling.NewGetSubscriptionUseCase(subscriptionRepo)
	createCheckoutUseCase := appBilling.NewCreateCheckoutUseCase(doitClient)
	handleDoitWebhookUseCase := appBilling.NewHandleDoitWebhookUseCase(billingWebhookEventRepo, subscriptionRepo)
	cancelSubscriptionUseCase := appBilling.NewCancelSubscriptionUseCase(subscriptionRepo)
	processBillingSubscriptionsUseCase := appBilling.NewProcessBillingSubscriptionsUseCase(subscriptionRepo)
	createTicketUseCase := appTicket.NewCreateTicketUseCase(ticketRepo, entitlementChecker)
	listTicketsUseCase := appTicket.NewListTicketsUseCase(ticketRepo)
	getTicketUseCase := appTicket.NewGetTicketUseCase(ticketRepo)
	reopenTicketUseCase := appTicket.NewReopenTicketUseCase(ticketRepo, userRepo, emailService)
	listAdminTicketsUseCase := appTicket.NewListAdminTicketsUseCase(ticketRepo)
	getAdminTicketUseCase := appTicket.NewGetAdminTicketUseCase(ticketRepo, userRepo)
	updateTicketStatusUseCase := appTicket.NewUpdateTicketStatusUseCase(ticketRepo, userRepo, emailService)
	getDashboardStatsUseCase := appIdentity.NewGetDashboardStatsUseCase(userRepo, budgetRepo, transactionRepo)
	createTransactionUseCase := appFinance.NewCreateTransactionUseCase(transactionService, walletService, entitlementChecker)
	getTransactionsUseCase := appFinance.NewGetTransactionsUseCase(transactionService, categoryService)
	getAllTransactionsUseCase := appFinance.NewGetAllTransactionsUseCase(transactionService, categoryService, currencyService)
	exportTransactionsUseCase := appFinance.NewExportTransactionsUseCase(transactionService, categoryService, entitlementChecker)
	updateTransactionUseCase := appFinance.NewUpdateTransactionUseCase(transactionService)
	deleteTransactionUseCase := appFinance.NewDeleteTransactionUseCase(transactionService)
	createCategoryUseCase := appFinance.NewCreateCategoryUseCase(categoryService, entitlementChecker)
	updateCategoryUseCase := appFinance.NewUpdateCategoryUseCase(categoryService)
	deleteCategoryUseCase := appFinance.NewDeleteCategoryUseCase(categoryService)
	getCategoriesUseCase := appFinance.NewGetCategoriesUseCase(categoryService)
	getAnalyticsUseCase := appFinance.NewGetAnalyticsUseCase(transactionService, categoryService, currencyService, entitlementChecker)
	getHealthScoreUseCase := appFinance.NewGetHealthScoreUseCase(budgetService, categoryService, transactionService, getAnalyticsUseCase, healthSnapshotRepo)
	getHealthScoreHistoryUseCase := appFinance.NewGetHealthScoreHistoryUseCase(healthSnapshotRepo)
	createBudgetUseCase := appFinance.NewCreateBudgetUseCase(budgetService, currencyService, categoryService, transactionService, entitlementChecker)
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
	createRecurringUseCase := appFinance.NewCreateRecurringTransactionUseCase(recurringRepo, walletService, categoryService, entitlementChecker)
	enqueueDueRecurringUseCase := appFinance.NewEnqueueDueRecurringUseCase(
		recurringRepo,
		pendingRepo,
		prefsRepo,
		notificationHelper,
	)
	getRecurringUseCase := appFinance.NewGetRecurringTransactionsUseCase(
		recurringRepo,
		categoryService,
		enqueueDueRecurringUseCase,
	)
	deleteRecurringUseCase := appFinance.NewDeleteRecurringTransactionUseCase(recurringRepo)
	listPendingUseCase := appFinance.NewListPendingTransactionsUseCase(
		enqueueDueRecurringUseCase,
		pendingRepo,
		categoryService,
	)
	confirmPendingUseCase := appFinance.NewConfirmPendingTransactionUseCase(pendingRepo, transactionService)
	rejectPendingUseCase := appFinance.NewRejectPendingTransactionUseCase(pendingRepo)
	createWalletUseCase := appFinance.NewCreateWalletUseCase(walletService, entitlementChecker)
	getWalletsUseCase := appFinance.NewGetWalletsUseCase(walletService)
	getWalletUseCase := appFinance.NewGetWalletUseCase(walletService)
	updateWalletUseCase := appFinance.NewUpdateWalletUseCase(walletService)
	setDefaultWalletUseCase := appFinance.NewSetDefaultWalletUseCase(walletService)
	archiveWalletUseCase := appFinance.NewArchiveWalletUseCase(walletService, goalService)
	unarchiveWalletUseCase := appFinance.NewUnarchiveWalletUseCase(walletService)
	getWalletBalanceUseCase := appFinance.NewGetWalletBalanceUseCase(walletService)
	getWalletSummaryUseCase := appFinance.NewGetWalletSummaryUseCase(walletService, currencyService)
	createGoalUseCase := appFinance.NewCreateGoalUseCase(goalService, currencyService, walletService, notificationHelper)
	getGoalsUseCase := appFinance.NewGetGoalsUseCase(goalService, walletService, notificationHelper)
	getGoalUseCase := appFinance.NewGetGoalUseCase(goalService, walletService, notificationHelper)
	updateGoalUseCase := appFinance.NewUpdateGoalUseCase(goalService, walletService, notificationHelper)
	deleteGoalUseCase := appFinance.NewDeleteGoalUseCase(goalService)
	checkGoalDeadlineAlertsUseCase := appFinance.NewCheckGoalDeadlineAlertsUseCase(
		goalService,
		walletService,
		prefsRepo,
		notificationHelper,
	)
	createAssetUseCase := appFinance.NewCreateAssetUseCase(assetService, entitlementChecker)
	getAssetsUseCase := appFinance.NewGetAssetsUseCase(assetService)
	updateAssetUseCase := appFinance.NewUpdateAssetUseCase(assetService)
	archiveAssetUseCase := appFinance.NewArchiveAssetUseCase(assetService)
	unarchiveAssetUseCase := appFinance.NewUnarchiveAssetUseCase(assetService)
	createLiabilityUseCase := appFinance.NewCreateLiabilityUseCase(liabilityService, entitlementChecker)
	getLiabilitiesUseCase := appFinance.NewGetLiabilitiesUseCase(liabilityService)
	updateLiabilityUseCase := appFinance.NewUpdateLiabilityUseCase(liabilityService)
	archiveLiabilityUseCase := appFinance.NewArchiveLiabilityUseCase(liabilityService)
	unarchiveLiabilityUseCase := appFinance.NewUnarchiveLiabilityUseCase(liabilityService)
	listLiabilityPaymentsUseCase := appFinance.NewListLiabilityPaymentsUseCase(liabilityService)
	recordLiabilityPaymentUseCase := appFinance.NewRecordLiabilityPaymentUseCase(liabilityService, createTransactionUseCase, categoryService)
	completeOnboardingUseCase := appFinance.NewCompleteOnboardingUseCase(
		prefsRepo,
		walletService,
		currencyService,
		categoryService,
		createRecurringUseCase,
		enqueueDueRecurringUseCase,
		createBudgetUseCase,
		liabilityService,
		getHealthScoreUseCase,
	)
	getNetWorthSummaryUseCase := appFinance.NewGetNetWorthSummaryUseCase(getWalletSummaryUseCase, assetService, liabilityService, currencyService)
	createTransferUseCase := appFinance.NewCreateTransferUseCase(transferService)
	getTransfersUseCase := appFinance.NewGetTransfersUseCase(transferService)

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
		resetAccountDataUseCase,
		deleteAccountUseCase,
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
		listPendingUseCase,
		confirmPendingUseCase,
		rejectPendingUseCase,
		createWalletUseCase,
		getWalletsUseCase,
		getWalletUseCase,
		updateWalletUseCase,
		setDefaultWalletUseCase,
		archiveWalletUseCase,
		unarchiveWalletUseCase,
		getWalletBalanceUseCase,
		getWalletSummaryUseCase,
		getHealthScoreUseCase,
		getHealthScoreHistoryUseCase,
		createGoalUseCase,
		getGoalsUseCase,
		getGoalUseCase,
		updateGoalUseCase,
		deleteGoalUseCase,
		createAssetUseCase,
		getAssetsUseCase,
		updateAssetUseCase,
		archiveAssetUseCase,
		unarchiveAssetUseCase,
		createLiabilityUseCase,
		getLiabilitiesUseCase,
		updateLiabilityUseCase,
		archiveLiabilityUseCase,
		unarchiveLiabilityUseCase,
		listLiabilityPaymentsUseCase,
		recordLiabilityPaymentUseCase,
		completeOnboardingUseCase,
		getNetWorthSummaryUseCase,
		createTransferUseCase,
		getTransfersUseCase,
	)
	exportHandlers := handlers.NewExportHandlers(exportTransactionsUseCase)
	dashboardHandlers := handlers.NewDashboardHandlers(getDashboardStatsUseCase)
	notificationHandlers := handlers.NewNotificationHandlers(
		getNotificationsUseCase,
		markNotificationReadUseCase,
		deleteNotificationUseCase,
	)
	feedbackHandlers := handlers.NewFeedbackHandlers(submitFeedbackUseCase)
	ticketHandlers := handlers.NewTicketHandlers(
		createTicketUseCase,
		listTicketsUseCase,
		getTicketUseCase,
		reopenTicketUseCase,
		listAdminTicketsUseCase,
		getAdminTicketUseCase,
		updateTicketStatusUseCase,
	)
	billingHandlers := handlers.NewBillingHandlers(
		getSubscriptionUseCase,
		createCheckoutUseCase,
		handleDoitWebhookUseCase,
		cancelSubscriptionUseCase,
	)
	authMiddleware := middleware.NewAuthMiddleware(tokenService, userRepo)

	return &App{
		DB:                                 db,
		IdentityHandlers:                   identityHandlers,
		FinanceHandlers:                    financeHandlers,
		ExportHandlers:                     exportHandlers,
		DashboardHandlers:                  dashboardHandlers,
		NotificationHandlers:               notificationHandlers,
		FeedbackHandlers:                   feedbackHandlers,
		TicketHandlers:                     ticketHandlers,
		BillingHandlers:                    billingHandlers,
		AuthMiddleware:                     authMiddleware,
		purgeDeletedAccountsUseCase:        purgeDeletedAccountsUseCase,
		cleanupExpiredTokensUseCase:        cleanupExpiredTokensUseCase,
		checkGoalDeadlineAlertsUseCase:     checkGoalDeadlineAlertsUseCase,
		processBillingSubscriptionsUseCase: processBillingSubscriptionsUseCase,
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
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
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
			auth.DELETE("/account", app.AuthMiddleware.RequireAuth(), app.IdentityHandlers.DeleteAccount)
		}

		protected := api.Group("")
		protected.Use(app.AuthMiddleware.RequireAuth())
		{
			adminOnly := protected.Group("")
			adminOnly.Use(app.AuthMiddleware.RequireRole("admin"))
			{
				adminOnly.GET("/users", app.IdentityHandlers.GetUsers)
				adminOnly.GET("/dashboard/stats", app.DashboardHandlers.GetDashboardStats)
				adminOnly.GET("/admin/tickets", app.TicketHandlers.ListAdminTickets)
				adminOnly.GET("/admin/tickets/:id", app.TicketHandlers.GetAdminTicket)
				adminOnly.PATCH("/admin/tickets/:id/status", app.TicketHandlers.UpdateTicketStatus)
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
			protected.GET("/export/transactions", app.ExportHandlers.ExportTransactions)

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
			protected.GET("/health-score", app.FinanceHandlers.GetHealthScore)
			protected.GET("/health-score/history", app.FinanceHandlers.GetHealthScoreHistory)

			protected.GET("/goals", app.FinanceHandlers.GetGoals)
			protected.POST("/goals", app.FinanceHandlers.CreateGoal)
			protected.GET("/goals/:id", app.FinanceHandlers.GetGoal)
			protected.PUT("/goals/:id", app.FinanceHandlers.UpdateGoal)
			protected.DELETE("/goals/:id", app.FinanceHandlers.DeleteGoal)

			protected.GET("/assets", app.FinanceHandlers.GetAssets)
			protected.POST("/assets", app.FinanceHandlers.CreateAsset)
			protected.PUT("/assets/:id", app.FinanceHandlers.UpdateAsset)
			protected.POST("/assets/:id/archive", app.FinanceHandlers.ArchiveAsset)
			protected.POST("/assets/:id/unarchive", app.FinanceHandlers.UnarchiveAsset)

			protected.GET("/liabilities", app.FinanceHandlers.GetLiabilities)
			protected.POST("/liabilities", app.FinanceHandlers.CreateLiability)
			protected.PUT("/liabilities/:id", app.FinanceHandlers.UpdateLiability)
			protected.POST("/liabilities/:id/archive", app.FinanceHandlers.ArchiveLiability)
			protected.POST("/liabilities/:id/unarchive", app.FinanceHandlers.UnarchiveLiability)
			protected.GET("/liabilities/:id/payments", app.FinanceHandlers.GetLiabilityPayments)
			protected.POST("/liabilities/:id/payments", app.FinanceHandlers.RecordLiabilityPayment)

			protected.GET("/net-worth/summary", app.FinanceHandlers.GetNetWorthSummary)

			protected.GET("/preferences", app.IdentityHandlers.GetPreferences)
			protected.PUT("/preferences", app.IdentityHandlers.UpdatePreferences)
			protected.GET("/me/subscription", app.BillingHandlers.GetSubscription)
			protected.POST("/billing/checkout", app.BillingHandlers.Checkout)
			protected.POST("/billing/cancel", app.BillingHandlers.CancelSubscription)
			protected.POST("/account/reset/challenge", app.IdentityHandlers.CreateAccountResetChallenge)
			protected.POST("/account/reset", app.IdentityHandlers.ResetAccountData)
			protected.POST("/onboarding/complete", app.FinanceHandlers.CompleteOnboarding)

			protected.GET("/notifications", app.NotificationHandlers.GetNotifications)
			protected.PUT("/notifications/:id/read", app.NotificationHandlers.MarkNotificationRead)
			protected.DELETE("/notifications/:id", app.NotificationHandlers.DeleteNotification)

			protected.POST("/feedback", app.FeedbackHandlers.SubmitFeedback)

			protected.POST("/tickets", app.TicketHandlers.CreateTicket)
			protected.GET("/tickets", app.TicketHandlers.ListTickets)
			protected.GET("/tickets/:id", app.TicketHandlers.GetTicket)
			protected.POST("/tickets/:id/reopen", app.TicketHandlers.ReopenTicket)

			protected.GET("/recurring-transactions", app.FinanceHandlers.GetRecurringTransactions)
			protected.POST("/recurring-transactions", app.FinanceHandlers.CreateRecurringTransaction)
			protected.DELETE("/recurring-transactions/:id", app.FinanceHandlers.DeleteRecurringTransaction)

			protected.GET("/pending-transactions", app.FinanceHandlers.GetPendingTransactions)
			protected.POST("/pending-transactions/:id/confirm", app.FinanceHandlers.ConfirmPendingTransaction)
			protected.POST("/pending-transactions/:id/reject", app.FinanceHandlers.RejectPendingTransaction)

			protected.GET("/wallets", app.FinanceHandlers.GetWallets)
			protected.POST("/wallets", app.FinanceHandlers.CreateWallet)
			protected.GET("/wallets/summary", app.FinanceHandlers.GetWalletSummary)
			protected.GET("/wallets/:id", app.FinanceHandlers.GetWallet)
			protected.PUT("/wallets/:id", app.FinanceHandlers.UpdateWallet)
			protected.POST("/wallets/:id/default", app.FinanceHandlers.SetDefaultWallet)
			protected.POST("/wallets/:id/archive", app.FinanceHandlers.ArchiveWallet)
			protected.POST("/wallets/:id/unarchive", app.FinanceHandlers.UnarchiveWallet)
			protected.GET("/wallets/:id/balance", app.FinanceHandlers.GetWalletBalance)

			protected.GET("/transfers", app.FinanceHandlers.GetTransfers)
			protected.POST("/transfers", app.FinanceHandlers.CreateTransfer)
		}
	}

	r.GET("/health", func(c *gin.Context) {
		handlers.SuccessResponse(c, http.StatusOK, gin.H{"status": "ok"})
	})
	r.POST("/webhooks/doit", app.BillingHandlers.HandleDoitWebhook)

	return r
}

// StartBackgroundJobs starts periodic maintenance such as account purge and token hygiene.
func (app *App) StartBackgroundJobs(ctx context.Context) {
	go func() {
		runPurge := func() {
			if app.purgeDeletedAccountsUseCase == nil {
				return
			}
			purged, err := app.purgeDeletedAccountsUseCase.Execute(context.Background())
			if err != nil {
				log.Printf("account purge job failed: %v", err)
				return
			}
			if purged > 0 {
				log.Printf("account purge job removed %d account(s)", purged)
			}
		}

		runTokenCleanup := func() {
			if app.cleanupExpiredTokensUseCase == nil {
				return
			}
			result, err := app.cleanupExpiredTokensUseCase.Execute(context.Background())
			if err != nil {
				log.Printf("token hygiene job failed: %v", err)
				return
			}
			if result.SessionTokensDeleted > 0 || result.PasswordResetTokensDeleted > 0 {
				log.Printf(
					"token hygiene job deleted %d session token(s) and %d password-reset token(s)",
					result.SessionTokensDeleted,
					result.PasswordResetTokensDeleted,
				)
			}
		}

		runGoalDeadlineAlerts := func() {
			if app.checkGoalDeadlineAlertsUseCase == nil {
				return
			}
			app.checkGoalDeadlineAlertsUseCase.Execute(context.Background())
		}

		runBillingMaintenance := func() {
			if app.processBillingSubscriptionsUseCase == nil {
				return
			}
			updated, err := app.processBillingSubscriptionsUseCase.Execute(context.Background())
			if err != nil {
				log.Printf("billing maintenance job failed: %v", err)
				return
			}
			if updated > 0 {
				log.Printf("billing maintenance job updated %d subscription(s)", updated)
			}
		}

		runPurge()
		runTokenCleanup()
		runGoalDeadlineAlerts()
		runBillingMaintenance()

		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runPurge()
				runTokenCleanup()
				runGoalDeadlineAlerts()
				runBillingMaintenance()
			}
		}
	}()
}
