package config

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server       ServerConfig       `mapstructure:"server"`
	DB           DBConfig           `mapstructure:"db"`
	JWT          JWTConfig          `mapstructure:"jwt"`
	CORS         CORSConfig         `mapstructure:"cors"`
	Log          LogConfig          `mapstructure:"log"`
	Cache        CacheConfig        `mapstructure:"cache"`
	Queue        QueueConfig        `mapstructure:"queue"`
	OTEL         OTELConfig         `mapstructure:"otel"`
	Scheduler    SchedulerConfig    `mapstructure:"scheduler"`
	ExchangeRate ExchangeRateConfig `mapstructure:"exchange_rate"`
	OCR          OCRConfig          `mapstructure:"ocr"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	SSLMode  string `mapstructure:"sslmode"`
}

type JWTConfig struct {
	Secret       string `mapstructure:"secret"`
	ExpireMinute int    `mapstructure:"expire_minutes"`
}

type CORSConfig struct {
	Origins string `mapstructure:"origins"`
}

type LogConfig struct {
	Dir        string `mapstructure:"dir"`
	Info       string `mapstructure:"info"`
	Warn       string `mapstructure:"warn"`
	Error      string `mapstructure:"error"`
	Target     string `mapstructure:"target"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
}

type CacheConfig struct {
	Type        string       `mapstructure:"type"`
	TTL         int          `mapstructure:"ttl"`
	L1TTL       int          `mapstructure:"l1_ttl"`
	MaxL1Items  int          `mapstructure:"max_l1_items"`
	Redis       RedisConfig  `mapstructure:"redis"`
}

// L1Duration returns the L1 TTL for tiered caching.
// If L1TTL is 0, defaults to TTL / 10 (min 10s).
func (c *CacheConfig) L1Duration() int {
	if c.L1TTL > 0 {
		return c.L1TTL
	}
	if c.TTL > 100 {
		return c.TTL / 10
	}
	return 30
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type QueueConfig struct {
	Enabled    bool              `mapstructure:"enabled"`
	Type       string            `mapstructure:"type"`       // redis | kafka
	Workers    int               `mapstructure:"workers"`
	MaxRetries int               `mapstructure:"max_retries"`
	Redis      QueueRedisConfig  `mapstructure:"redis"`
	Kafka      QueueKafkaConfig  `mapstructure:"kafka"`
}

type QueueRedisConfig struct {
	Addr          string `mapstructure:"addr"`
	Password      string `mapstructure:"password"`
	DB            int    `mapstructure:"db"`
	Stream        string `mapstructure:"stream"`
	ConsumerGroup string `mapstructure:"consumer_group"`
}

type QueueKafkaConfig struct {
	Brokers []string `mapstructure:"brokers"`
	Topic   string   `mapstructure:"topic"`
	GroupID string   `mapstructure:"group_id"`
}

type OTELConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	ServiceName    string `mapstructure:"service_name"`
	TracesExporter string `mapstructure:"traces_exporter"`
	MetricsPath    string `mapstructure:"metrics_path"`
}

type SchedulerConfig struct {
	RecurringCheckMinutes int `mapstructure:"recurring_check_minutes"`
}

type ExchangeRateConfig struct {
	Provider string `mapstructure:"provider"`   // exchangerate-api | frankfurter
	APIKey   string `mapstructure:"api_key"`
	Base     string `mapstructure:"base"`       // base currency for auto-fetch (e.g. USD)
}

type OCRConfig struct {
	Provider string `mapstructure:"provider"`   // paddleocr
	Endpoint string `mapstructure:"endpoint"`   // e.g. http://localhost:9000
}

func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DB.Host, c.DB.Port, c.DB.User, c.DB.Password, c.DB.Name, c.DB.SSLMode,
	)
}

func Load() *Config {
	v := viper.NewWithOptions(
		viper.EnvKeyReplacer(strings.NewReplacer(".", "_")),
	)

	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")

	// — Defaults —
	v.SetDefault("server.port", "8000")

	v.SetDefault("db.host", "localhost")
	v.SetDefault("db.port", "5432")
	v.SetDefault("db.user", "bookkeeper")
	v.SetDefault("db.password", "bookkeeper_dev")
	v.SetDefault("db.name", "bookkeeping")
	v.SetDefault("db.sslmode", "disable")

	v.SetDefault("jwt.secret", "dev-secret-key-change-in-production")
	v.SetDefault("jwt.expire_minutes", 10080)

	v.SetDefault("cors.origins", "http://localhost:5173,http://localhost:3000")

	v.SetDefault("log.target", "file")
	v.SetDefault("log.dir", "logs")
	v.SetDefault("log.info", "app.log")
	v.SetDefault("log.warn", "warn.log")
	v.SetDefault("log.error", "error.log")
	v.SetDefault("log.date_format", "2006-01-02")
	v.SetDefault("log.time_format", "15:04:05.000")
	v.SetDefault("log.max_size", 100)
	v.SetDefault("log.max_age", 30)
	v.SetDefault("log.max_backups", 10)
	v.SetDefault("log.compress", true)

	v.SetDefault("cache.type", "memory")
	v.SetDefault("cache.ttl", 300)
	v.SetDefault("cache.l1_ttl", 30)
	v.SetDefault("cache.max_l1_items", 10000)
	v.SetDefault("cache.redis.addr", "localhost:6379")
	v.SetDefault("cache.redis.password", "")
	v.SetDefault("cache.redis.db", 0)

	v.SetDefault("queue.enabled", true)
	v.SetDefault("queue.type", "inmemory")
	v.SetDefault("queue.workers", 5)
	v.SetDefault("queue.max_retries", 3)
	v.SetDefault("queue.redis.addr", "localhost:6379")
	v.SetDefault("queue.redis.password", "")
	v.SetDefault("queue.redis.db", 0)
	v.SetDefault("queue.redis.stream", "task-queue")
	v.SetDefault("queue.redis.consumer_group", "workers")
	v.SetDefault("queue.kafka.brokers", []string{"localhost:9092"})
	v.SetDefault("queue.kafka.topic", "task-queue")
	v.SetDefault("queue.kafka.group_id", "workers")

	v.SetDefault("otel.enabled", false)
	v.SetDefault("otel.service_name", "personal-bookkeeping")
	v.SetDefault("otel.traces_exporter", "none")
	v.SetDefault("otel.metrics_path", "/metrics")

	v.SetDefault("scheduler.recurring_check_minutes", 60)

	v.SetDefault("exchange_rate.provider", "exchangerate-api")
	v.SetDefault("exchange_rate.api_key", "")
	v.SetDefault("exchange_rate.base", "CNY")

	v.SetDefault("ocr.provider", "paddleocr")
	v.SetDefault("ocr.endpoint", "http://localhost:9000")

	// — Environment variable bindings (backward-compatible names) —
	envBindings := map[string]string{
		"server.port":         "SERVER_PORT",
		"db.host":             "DB_HOST",
		"db.port":             "DB_PORT",
		"db.user":             "DB_USER",
		"db.password":         "DB_PASSWORD",
		"db.name":             "DB_NAME",
		"db.sslmode":          "DB_SSLMODE",
		"jwt.secret":          "JWT_SECRET",
		"jwt.expire_minutes":  "JWT_EXPIRE_MINUTES",
		"cors.origins":        "CORS_ORIGINS",
		"cache.type":          "CACHE_TYPE",
		"cache.redis.addr":    "CACHE_REDIS_ADDR",
		"cache.redis.password": "CACHE_REDIS_PASSWORD",
		"cache.redis.db":            "CACHE_REDIS_DB",
		"queue.type":                "QUEUE_TYPE",
		"queue.redis.addr":          "QUEUE_REDIS_ADDR",
		"queue.redis.password":      "QUEUE_REDIS_PASSWORD",
		"queue.redis.db":            "QUEUE_REDIS_DB",
		"queue.redis.stream":        "QUEUE_REDIS_STREAM",
		"queue.redis.consumer_group": "QUEUE_REDIS_CONSUMER_GROUP",
		"ocr.endpoint":              "OCR_ENDPOINT",
		"exchange_rate.api_key":     "EXCHANGE_RATE_API_KEY",
	}
	for key, env := range envBindings {
		if err := v.BindEnv(key, env); err != nil {
			slog.Warn("failed to bind env var", "key", key, "env", env, "error", err)
		}
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			slog.Warn("failed to read config file", "error", err)
		}
	} else {
		slog.Info("loaded config file", "path", v.ConfigFileUsed())
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		slog.Error("failed to unmarshal config", "error", err)
		return nil
	}

	// — Startup validation —
	if cfg.JWT.Secret == "" || cfg.JWT.Secret == "dev-secret-key-change-in-production" {
		slog.Warn("JWT secret is using default or empty value, please set JWT_SECRET in production")
	}
	if cfg.ExchangeRate.APIKey == "" {
		slog.Warn("ExchangeRate API key is empty, exchange rate auto-fetch will fail")
	}

	slog.Info("config loaded",
		"server.port", cfg.Server.Port,
		"db.host", cfg.DB.Host,
		"cache.type", cfg.Cache.Type,
		"queue.enabled", cfg.Queue.Enabled,
	)

	return &cfg
}
