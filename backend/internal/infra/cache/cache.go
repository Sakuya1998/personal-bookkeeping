package cache

import (
	"context"
	"time"
)

// Cache 通用缓存接口，支持 Redis 和内存两种后端
type Cache interface {
	// Get 获取缓存值，key 不存在返回 ErrMiss
	Get(ctx context.Context, key string) (string, error)

	// Set 写入缓存
	Set(ctx context.Context, key string, value string, ttl time.Duration) error

	// Delete 删除缓存键
	Delete(ctx context.Context, key string) error

	// Exists 检查键是否存在
	Exists(ctx context.Context, key string) (bool, error)

	// Flush 清空所有缓存（谨慎使用）
	Flush(ctx context.Context) error

	// Close 释放资源
	Close() error
}

var ErrMiss = errMiss{}

type errMiss struct{}

func (e errMiss) Error() string { return "cache: key not found" }
func (e errMiss) Is(target error) bool {
	_, ok := target.(errMiss)
	return ok
}

// Key helpers — 统一键命名
func KeyExchangeRate(from, to, date string) string {
	return "exchange:rate:" + from + ":" + to + ":" + date
}

func KeyCategoryList(userID string) string {
	return "categories:list:" + userID
}

func KeyTokenBlacklist(jti string) string {
	return "token:blacklist:" + jti
}

// NullSentinel is cached when a DB lookup returns no result, preventing
// cache penetration (repeated DB queries for non-existent keys).
const NullSentinel = "__NULL__"

// NullCacheTTL is how long null entries are kept to absorb burst traffic.
const NullCacheTTL = 30 * time.Second

// defaultCache is the package-level default cache instance.
var defaultCache Cache

// SetDefault sets the package-level default cache instance.
func SetDefault(c Cache) { defaultCache = c }

// GetDefault returns the package-level default cache instance.
func GetDefault() Cache { return defaultCache }
