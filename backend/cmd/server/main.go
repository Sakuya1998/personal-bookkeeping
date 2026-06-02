package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"personal-bookkeeping/internal/infra/database"
	"personal-bookkeeping/internal/app/router"
	service "personal-bookkeeping/internal/app/service"
	"personal-bookkeeping/internal/app/task"
	"personal-bookkeeping/internal/infra/cache"
	"personal-bookkeeping/internal/infra/config"
	"personal-bookkeeping/internal/infra/logger"
	"personal-bookkeeping/internal/infra/migrate"
	"personal-bookkeeping/internal/infra/otel"
	"personal-bookkeeping/internal/infra/queue"

	"github.com/gin-gonic/gin"
)

// @title           Personal Bookkeeping API
// @version         1.0.0
// @description     个人记账应用后端 API
// @termsOfService  https://github.com/your-repo

// @contact.name   Developer
// @contact.email  dev@example.com

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8000
// @BasePath  /api/v1

// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                JWT token，格式: Bearer <token>

func main() {
	cfg := config.Load()
	if cfg == nil {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
		slog.Error("failed to load config")
		os.Exit(1)
	}

	// — slog —
	logHandler, err := logger.NewHandler(logger.LevelFileConfig{
		Target:     cfg.Log.Target,
		Dir:        cfg.Log.Dir,
		Info:       cfg.Log.Info,
		Warn:       cfg.Log.Warn,
		Error:      cfg.Log.Error,
		MaxSize:    cfg.Log.MaxSize,
		MaxAge:     cfg.Log.MaxAge,
		MaxBackups: cfg.Log.MaxBackups,
		Compress:   cfg.Log.Compress,
	}, &slog.HandlerOptions{Level: slog.LevelDebug})
	if err != nil {
		slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stderr, nil)))
		slog.Error("failed to init logger", "error", err)
		os.Exit(1)
	}
	slog.SetDefault(slog.New(logHandler))

	// — OTEL —
	otelCfg := &otel.Config{
		Enabled:        cfg.OTEL.Enabled,
		ServiceName:    cfg.OTEL.ServiceName,
		TracesExporter: cfg.OTEL.TracesExporter,
		MetricsPath:    cfg.OTEL.MetricsPath,
	}
	o, err := otel.Init(otelCfg)
	if err != nil {
		slog.Error("failed to init otel", "error", err)
		os.Exit(1)
	}

	// — Cache —
	var cch cache.Cache
	cch, err = cache.NewFromConfig(&cfg.Cache)
	if err != nil {
		slog.Error("failed to init cache", "error", err)
		os.Exit(1)
	}
	cache.SetDefault(cch)
	slog.Info("cache initialized", "type", cfg.Cache.Type)

	// — Queue —
	var q queue.Queue
	q, err = queue.NewFromConfig(&cfg.Queue)
	if err != nil {
		slog.Error("failed to init queue", "error", err)
		os.Exit(1)
	}
	queue.SetDefault(q)
	if q != nil {
		task.RegisterAll(q)
		q.Start(context.Background())
		slog.Info("queue started", "type", cfg.Queue.Type, "workers", cfg.Queue.Workers)

		// Use a cancellable context so schedulers stop on shutdown
		schedCtx, schedCancel := context.WithCancel(context.Background())

		// Start recurring transaction scheduler
		interval := time.Duration(cfg.Scheduler.RecurringCheckMinutes) * time.Minute
		task.StartRecurringScheduler(schedCtx, q, interval)

		// Start exchange rate auto-update scheduler (daily)
		task.StartExchangeRateScheduler(schedCtx, q)

		// Cancel schedulers on shutdown (before the 30s timeout)
		defer schedCancel()
	}

	// Connect database
	database.Init(cfg)
	if database.GetDB() == nil {
		os.Exit(1)
	}

	// Run database migrations
	if err := migrate.Up(cfg.DSN()); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}
	slog.Info("database migration completed")

	// Migrate existing ledger owners into ledger_members table
	svc := service.NewService()
	memberSvc := service.NewMemberService(svc)
	if err := memberSvc.MigrateLedgerOwnership(); err != nil {
		slog.Warn("ledger member migration (non-fatal)", "error", err)
	}

	// Initialize exchange rate provider with DI-injected DB and cache
	service.InitExchangeRateProvider(database.GetDB(), cch)

	// Immediate exchange rate fetch on startup
	if cfg.ExchangeRate.APIKey != "" {
		go func() {
			slog.Info("exchange rates: initial fetch on startup")
			if err := service.UpdateExchangeRates(database.GetDB(), &cfg.ExchangeRate); err != nil {
				slog.Warn("exchange rates: initial fetch failed (will retry via scheduler)", "error", err)
			}
		}()
	}

	// — Gin —
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(o.GinMiddleware(cfg.OTEL.ServiceName))
	r.Use(logger.GinSlogMiddleware())

	// Metrics endpoint
	r.GET(cfg.OTEL.MetricsPath, func(c *gin.Context) {
		if o.MetricsHandler != nil {
			o.MetricsHandler.ServeHTTP(c.Writer, c.Request)
		} else {
			c.JSON(200, gin.H{"status": "otel disabled"})
		}
	})

	router.Setup(r, cfg)

	addr := ":" + cfg.Server.Port
	slog.Info("server starting", "addr", addr)

	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// — Graceful shutdown —
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	slog.Info("shutting down", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("http server shutdown error", "error", err)
	}

	// shutdown queue first (wait for in-flight tasks that may access DB)
	if q != nil {
		if err := q.Shutdown(ctx); err != nil {
			slog.Error("queue shutdown error", "error", err)
		}
	}

	if err := database.Close(); err != nil {
		slog.Error("db close error", "error", err)
	}

	// shutdown cache
	if err := cch.Close(); err != nil {
		slog.Error("cache close error", "error", err)
	}

	o.Shutdown()

	if c, ok := logHandler.(interface{ Close() error }); ok {
		if err := c.Close(); err != nil {
			slog.Error("log close error", "error", err)
		}
	}

	slog.Info("server stopped")
}
