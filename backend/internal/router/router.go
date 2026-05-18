package router

import (
	"github.com/finance-os/backend/internal/config"
	"github.com/finance-os/backend/internal/handler"
	"github.com/finance-os/backend/internal/middleware"
	"github.com/finance-os/backend/internal/repository"
	"github.com/finance-os/backend/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

// Setup configura todas as rotas da aplicação.
func Setup(e *echo.Echo, db *pgxpool.Pool, cfg *config.Config) {
	// Inicializa Repositories
	userRepo := repository.NewUserRepository(db)
	txRepo := repository.NewTransactionRepository(db)

	// Inicializa Services
	jwtService := service.NewJWTService(cfg)
	authService := service.NewAuthService(userRepo, jwtService)

	encryptionService, err := service.NewEncryptionService(cfg.EncryptionKey)
	if err != nil {
		panic("erro ao inicializar encryption service: " + err.Error())
	}

	installmentService := service.NewInstallmentsService(db)
	subscriptionService := service.NewSubscriptionService(db)
	classifierService := service.NewClassifierService(db, cfg)
	feedService := service.NewFeedService(db)
	syncService := service.NewSyncService(db, installmentService, classifierService, feedService)
	goalsService := service.NewGoalsService(db)
	survivalModeService := service.NewSurvivalModeService(db)
	impulseRadarService := service.NewImpulseRadarService(db)

	// Inicializa handlers
	authH := handler.NewAuthHandler(authService, cfg)
	accountsH := handler.NewAccountsHandler(db, syncService, encryptionService, userRepo, cfg)
	transactionsH := handler.NewTransactionsHandler(txRepo, classifierService)
	reportsH := handler.NewReportsHandler(db, cfg, survivalModeService, impulseRadarService)
	cardsH := handler.NewCardsHandler(installmentService, subscriptionService)
	chatH := handler.NewChatHandler(cfg)
	categoriesH := handler.NewCategoriesHandler(db)
	feedH := handler.NewFeedHandler(feedService)
	goalsH := handler.NewGoalsHandler(goalsService, cfg)
	simulatorH := handler.NewSimulatorHandler()

	// Grupo API v1
	v1 := e.Group("/api/v1")

	// Rotas públicas
	auth := v1.Group("/auth")
	auth.POST("/register", authH.Register)
	auth.POST("/login", authH.Login)
	auth.POST("/refresh", authH.Refresh)

	// Rotas protegidas (Middleware JWT customizado)
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware(jwtService))

	// User
	protected.GET("/me", authH.Me)

	// Accounts
	accounts := protected.Group("/accounts")
	accounts.GET("", accountsH.ListAccounts)
	accounts.POST("/connect-token", accountsH.ConnectToken)
	accounts.POST("/sync", accountsH.Sync)
	accounts.POST("/keys", accountsH.SavePluggyKeys)
	accounts.PATCH("/:id/settings", accountsH.UpdateAccountSettings)
	accounts.DELETE("/:id", accountsH.DeleteAccount)

	// Transactions
	transactions := protected.Group("/transactions")
	transactions.GET("", transactionsH.ListTransactions)
	transactions.PATCH("/:id/category", transactionsH.UpdateCategory)
	transactions.GET("/summary", transactionsH.Summary)

	// Categories
	protected.GET("/categories", categoriesH.ListCategories)

	// Reports
	reports := protected.Group("/reports")
	reports.GET("", reportsH.GetReports)
	reports.GET("/cashflow", reportsH.GetCashflow)
	reports.GET("/behavioral", reportsH.GetBehavioralInsights)
	reports.GET("/comparison", reportsH.GetComparison)
	reports.GET("/invisible-spending", reportsH.GetInvisibleSpending)
	reports.GET("/projection", reportsH.GetProjections)
	reports.GET("/health-score", reportsH.GetHealthScore)
	reports.GET("/stress-score", reportsH.GetStressScore)
	reports.GET("/survival-mode", reportsH.GetSurvivalMode)
	reports.GET("/upcoming-expenses", reportsH.GetUpcomingExpenses)
	reports.GET("/narrative", reportsH.GetNarrativeReport)
	reports.GET("/timeline", reportsH.GetTimeline)
	reports.GET("/lifestyle-drift", reportsH.GetLifestyleDrift)
	reports.GET("/financial-memory", reportsH.GetFinancialMemory)
	reports.GET("/dangerous-days", reportsH.GetDangerousDays)
	reports.GET("/behavioral-prediction", reportsH.GetBehavioralPrediction)
	reports.GET("/micro-spending", reportsH.GetMicroSpending)
	reports.GET("/impulse-radar", reportsH.GetImpulseRadar)
	reports.GET("/personal-inflation", reportsH.GetPersonalInflation)
	reports.GET("/silent-growth", reportsH.GetSilentGrowth)
	reports.GET("/weekly-profile", reportsH.GetWeeklyProfile)
	reports.GET("/weekday-weekend", reportsH.GetWeekdayWeekend)
	reports.GET("/salary-effect", reportsH.GetSalaryEffect)
	reports.GET("/monthly-weeks", reportsH.GetMonthlyWeeks)
	reports.GET("/impulse", reportsH.GetImpulse)
	reports.GET("/compensation-pattern", reportsH.GetCompensationPattern)
	reports.GET("/meal-cost", reportsH.GetMealCost)
	reports.GET("/convenience-index", reportsH.GetConvenienceIndex)
	reports.GET("/ticket-analysis", reportsH.GetTicketAnalysis)
	reports.GET("/loyalty", reportsH.GetLoyalty)
	reports.POST("/trigger/:type", reportsH.TriggerAgent)

	// Merchants
	merchants := protected.Group("/merchants")
	merchants.GET("", reportsH.GetTopMerchants)
	merchants.GET("/:name", reportsH.GetMerchantProfile)

	// Cards & Installments
	cards := protected.Group("/cards")
	cards.GET("/installments", cardsH.ListInstallments)
	cards.GET("/invoice/:account_id", cardsH.GetInvoice)
	cards.GET("/subscriptions", cardsH.ListSubscriptions)

	// Chat
	protected.POST("/chat", chatH.SendMessage)

	// Feed
	feed := protected.Group("/feed")
	feed.GET("", feedH.GetFeed)
	feed.GET("/unread-count", feedH.GetUnreadCount)
	feed.PATCH("/:id/read", feedH.MarkAsRead)
	feed.PATCH("/read-all", feedH.MarkAllAsRead)

	// Goals
	goals := protected.Group("/goals")
	goals.GET("", goalsH.List)
	goals.POST("", goalsH.Create)
	goals.GET("/suggest", goalsH.Suggest)

	// Simulator
	simulator := protected.Group("/simulator")
	simulator.POST("/purchase", simulatorH.SimulatePurchase)
	simulator.POST("/cut-subscription", simulatorH.SimulateCut)
}
