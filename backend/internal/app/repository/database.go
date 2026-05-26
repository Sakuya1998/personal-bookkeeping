package database

import (
	"log/slog"

	"personal-bookkeeping/internal/app/model"
	"personal-bookkeeping/internal/infra/cache"
	"personal-bookkeeping/internal/infra/config"
	"personal-bookkeeping/internal/infra/logger"
	"personal-bookkeeping/internal/infra/queue"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var (
	DB  *gorm.DB
	cch cache.Cache
	q   queue.Queue
)

func Init(cfg *config.Config) {
	var err error
	DB, err = gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.NewSlogLogger(gormlogger.Info),
	})
	if err != nil {
		slog.Error("failed to connect database", "error", err)
		return
	}

	// Auto migrate (dev convenience)
	if err := DB.AutoMigrate(
		&models.User{},
		&models.Ledger{},
		&models.Category{},
		&models.Transaction{},
		&models.ExchangeRate{},
	); err != nil {
		slog.Error("failed to migrate database", "error", err)
		return
	}

	// Ensure performance indexes
	createIndexes(DB)

	slog.Info("database connected and migrated successfully")
}

func createIndexes(db *gorm.DB) {
	indexes := []string{
		`CREATE INDEX IF NOT EXISTS idx_transactions_ledger_user_date ON transactions (ledger_id, user_id, transaction_date)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_user_type ON transactions (user_id, type)`,
	}
	for _, idx := range indexes {
		if err := db.Exec(idx).Error; err != nil {
			slog.Warn("failed to create index", "error", err)
		}
	}
}

func InitCache(cfg cache.Cache) {
	cch = cfg
}

func InitQueue(qq queue.Queue) {
	q = qq
}

func GetDB() *gorm.DB {
	return DB
}

func GetCache() cache.Cache {
	return cch
}

func GetQueue() queue.Queue {
	return q
}

// Ping 检查数据库是否可达。
func Ping() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

// Close 关闭数据库连接。
func Close() error {
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
