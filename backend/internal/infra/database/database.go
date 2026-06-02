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

// Init 连接数据库。
func Init(cfg *config.Config) {
	var err error
	DB, err = gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.NewSlogLogger(gormlogger.Info),
	})
	if err != nil {
		slog.Error("failed to connect database", "error", err)
		return
	}

	slog.Info("database connected successfully")
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
