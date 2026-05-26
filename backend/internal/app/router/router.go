package routes

import (
	"personal-bookkeeping/internal/app/handler"
	"personal-bookkeeping/internal/app/middleware"
	"personal-bookkeeping/internal/app/repository"
	"personal-bookkeeping/internal/infra/config"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "personal-bookkeeping/internal/infra/swagger"
)

func Setup(r *gin.Engine, cfg *config.Config) {
	// CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", cfg.CORS.Origins)
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

	// Auth routes (no auth required)
	auth := handlers.NewAuthHandler(cfg)
	r.POST("/api/v1/auth/register", auth.Register)
	r.POST("/api/v1/auth/login", auth.Login)

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
		ledger := handlers.NewLedgerHandler()
		protected.GET("/ledgers", ledger.List)
		protected.POST("/ledgers", ledger.Create)
		protected.GET("/ledgers/:ledger_id", ledger.Get)
		protected.PUT("/ledgers/:ledger_id", ledger.Update)
		protected.DELETE("/ledgers/:ledger_id", ledger.Delete)
		protected.GET("/ledgers/:ledger_id/summary", ledger.Summary)
		protected.GET("/ledgers/:ledger_id/monthly-trend", ledger.MonthlyTrend)
		protected.GET("/ledgers/:ledger_id/category-breakdown", ledger.CategoryBreakdown)
		protected.GET("/ledgers/:ledger_id/daily-transactions", ledger.DailyTransactions)
		protected.GET("/ledgers/:ledger_id/export", ledger.Export)
		protected.GET("/ledgers/:ledger_id/tags", ledger.Tags)

		// Categories
		cat := handlers.NewCategoryHandler()
		protected.GET("/ledgers/:ledger_id/categories", cat.List)
		protected.POST("/categories", cat.Create)
		protected.PUT("/categories/:id", cat.Update)
		protected.DELETE("/categories/:id", cat.Delete)

		// Transactions
		txn := handlers.NewTransactionHandler()
		protected.GET("/ledgers/:ledger_id/transactions", txn.List)
		protected.POST("/transactions", txn.Create)
		protected.PUT("/transactions/:id", txn.Update)
		protected.DELETE("/transactions/:id", txn.Delete)
		protected.POST("/transactions/batch-delete", txn.BatchDelete)
		protected.PUT("/transactions/batch-update", txn.BatchUpdate)

		// Exchange rates
		rate := handlers.NewExchangeRateHandler()
		protected.GET("/exchange-rates", rate.List)
		protected.POST("/exchange-rates", rate.Create)
		protected.GET("/exchange-rates/latest", rate.Latest)
		protected.DELETE("/exchange-rates/:id", rate.Delete)

		// Recurring rules
		rec := handlers.NewRecurringHandler()
		protected.GET("/recurring", rec.List)
		protected.POST("/recurring", rec.Create)
		protected.PUT("/recurring/:id", rec.Update)
		protected.DELETE("/recurring/:id", rec.Delete)
		protected.GET("/recurring/upcoming", rec.Upcoming)

		// Budgets
		bgt := handlers.NewBudgetHandler()
		protected.POST("/budgets", bgt.Upsert)
		protected.GET("/budgets", bgt.List)
		protected.GET("/budgets/status", bgt.Status)
		protected.DELETE("/budgets/:id", bgt.Delete)

		// Report
		rpt := handlers.NewReportHandler()
		protected.GET("/ledgers/:ledger_id/report", rpt.GenerateReport)
		protected.GET("/ledgers/:ledger_id/report/preview", rpt.ReportPreview)

		// OCR
		ocr := handlers.NewOCRHandler(cfg.OCR.Endpoint)
		protected.POST("/ocr/receipt", ocr.RecognizeReceipt)
	}
}
