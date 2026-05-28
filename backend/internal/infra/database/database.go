package database

import (
	"log/slog"

	"personal-bookkeeping/internal/infra/config"
	"personal-bookkeeping/internal/infra/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

var DB *gorm.DB

// Init 连接数据库并创建性能索引。
// 注意：AutoMigrate 由 main.go 在 Init 之后调用（app 层负责传递模型）。
func Init(cfg *config.Config) {
	var err error
	DB, err = gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.NewSlogLogger(gormlogger.Info),
	})
	if err != nil {
		slog.Error("failed to connect database", "error", err)
		return
	}

	// Ensure performance indexes
	createIndexes(DB)

	slog.Info("database connected successfully")
}

// AutoMigrate 自动迁移数据库表结构（dev 便利）。
func AutoMigrate(models ...interface{}) {
	if DB == nil {
		slog.Error("database not initialized, skipping migration")
		return
	}
	if err := DB.AutoMigrate(models...); err != nil {
		slog.Error("failed to migrate database", "error", err)
		return
	}
	slog.Info("database migrated successfully")
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

func GetDB() *gorm.DB {
	return DB
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
