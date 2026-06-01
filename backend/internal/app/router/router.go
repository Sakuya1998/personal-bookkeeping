package router

import (
	"strings"
	"time"

	"personal-bookkeeping/internal/app/handler"
	"personal-bookkeeping/internal/app/middleware"
	inframiddleware "personal-bookkeeping/internal/infra/middleware"
	"personal-bookkeeping/internal/infra/database"
	"personal-bookkeeping/internal/app/service"
	"personal-bookkeeping/internal/infra/config"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "personal-bookkeeping/internal/infra/swagger"
)

func Setup(r *gin.Engine, cfg *config.Config) {
	// CORS — split comma-separated origins, echo back the matching one
	allowedOrigins := strings.Split(cfg.CORS.Origins, ",")
	for i := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
	}
	r.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		allowed := ""
		for _, o := range allowedOrigins {
			if o == origin || o == "*" {
				allowed = origin
				break
			}
		}
		// fallback to first allowed origin for non-browser clients
		if allowed == "" && len(allowedOrigins) > 0 {
			allowed = allowedOrigins[0]
		}
		if allowed != "" {
			c.Header("Access-Control-Allow-Origin", allowed)
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check
	r.GET("/api/v1/health", func(c *gin.Context) {
		dbOK := "ok"
		if err := database.Ping(); err != nil {
			dbOK = "error: " + err.Error()
		}
		c.JSON(200, gin.H{
			"status": "ok",
			"db":     dbOK,
		})
	})

	// Swagger docs
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Auth routes (no auth required, rate limited)
	s := service.NewService()
	auth := handler.NewAuthHandler(service.NewAuthService(s, cfg.JWT.Secret, cfg.JWT.ExpireMinute))
	loginLimiter := inframiddleware.NewRateLimiter(10, 1*time.Minute)
	r.POST("/api/v1/auth/register", inframiddleware.LimitRate(loginLimiter, "/api/v1/auth/register"), auth.Register)
	r.POST("/api/v1/auth/login", inframiddleware.LimitRate(loginLimiter, "/api/v1/auth/login"), auth.Login)

	// Protected routes
	protected := r.Group("/api/v1")
	protected.Use(middleware.AuthRequired(cfg.JWT.Secret))
	{
		// Auth
		protected.GET("/auth/me", auth.Me)
		protected.POST("/auth/logout", auth.Logout)
		protected.PUT("/auth/password", auth.ChangePassword)
		protected.PUT("/auth/email", auth.ChangeEmail)

		// Ledgers
		ledger := handler.NewLedgerHandler(service.NewLedgerService(s))
		protected.GET("/ledgers", ledger.List)
		protected.POST("/ledgers", ledger.Create)
		protected.GET("/ledgers/:ledger_id", ledger.Get)
		protected.PUT("/ledgers/:ledger_id", ledger.Update)
		protected.DELETE("/ledgers/:ledger_id", ledger.Delete)
		protected.GET("/ledgers/:ledger_id/summary", ledger.Summary)
		protected.GET("/ledgers/:ledger_id/monthly-trend", ledger.MonthlyTrend)
		protected.GET("/ledgers/:ledger_id/category-breakdown", ledger.CategoryBreakdown)
		protected.GET("/ledgers/:ledger_id/daily-transactions", ledger.DailyTransactions)
		protected.GET("/ledgers/:ledger_id/tag-stats", ledger.TagStats)
		protected.GET("/ledgers/:ledger_id/export", ledger.Export)
		protected.GET("/ledgers/:ledger_id/tags", ledger.Tags)

		// Categories
		cat := handler.NewCategoryHandler(service.NewCategoryService(s))
		protected.GET("/ledgers/:ledger_id/categories", cat.List)
		protected.POST("/categories", cat.Create)
		protected.PUT("/categories/:id", cat.Update)
		protected.DELETE("/categories/:id", cat.Delete)

		// Transactions
		txn := handler.NewTransactionHandler(service.NewTransactionService(s))
		protected.GET("/ledgers/:ledger_id/transactions", txn.List)
		protected.POST("/transactions", txn.Create)
		protected.PUT("/transactions/:id", txn.Update)
		protected.DELETE("/transactions/:id", txn.Delete)
		protected.POST("/transactions/batch-delete", txn.BatchDelete)
		protected.PUT("/transactions/batch-update", txn.BatchUpdate)

		// Exchange rates
		rate := handler.NewExchangeRateHandler(service.NewExchangeRateService(s))
		protected.GET("/exchange-rates", rate.List)
		protected.POST("/exchange-rates/sync", rate.Sync)
		protected.GET("/exchange-rates/latest", rate.Latest)
		protected.DELETE("/exchange-rates/:id", rate.Delete)

		// Recurring rules
		rec := handler.NewRecurringHandler(service.NewRecurringService(s))
		protected.GET("/recurring", rec.List)
		protected.POST("/recurring", rec.Create)
		protected.PUT("/recurring/:id", rec.Update)
		protected.DELETE("/recurring/:id", rec.Delete)
		protected.GET("/recurring/upcoming", rec.Upcoming)

		// Budgets
		bgt := handler.NewBudgetHandler(service.NewBudgetService(s))
		protected.POST("/budgets", bgt.Upsert)
		protected.GET("/budgets", bgt.List)
		protected.GET("/budgets/status", bgt.Status)
		protected.DELETE("/budgets/:id", bgt.Delete)

		// Report
		rpt := handler.NewReportHandler(service.NewReportService(s))
		protected.GET("/ledgers/:ledger_id/report", rpt.GenerateReport)
		protected.GET("/ledgers/:ledger_id/report/preview", rpt.ReportPreview)

		// OCR
		ocr := handler.NewOCRHandler(cfg.OCR.Endpoint, service.NewOCRService(s))
		protected.POST("/ocr/receipt", ocr.RecognizeReceipt)
	}
}
