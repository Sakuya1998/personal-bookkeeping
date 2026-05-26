package cache

import (
	"context"
	"fmt"
	"time"

	"personal-bookkeeping/internal/infra/config"
)

// NewFromConfig 根据配置创建 Cache 实例。
//
// cache.type = "memory"  → 本地内存缓存（无外部依赖）
// cache.type = "redis"   → Redis 缓存，启动时 ping 验证
// cache.type = "tiered"  → L1 memory + L2 Redis 多级缓存
func NewFromConfig(cfg *config.CacheConfig) (Cache, error) {
	l2TTL := time.Duration(cfg.TTL) * time.Second
	l1TTL := time.Duration(cfg.L1Duration()) * time.Second

	switch cfg.Type {
	case "redis":
		r := NewRedis(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB, l2TTL)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := r.Ping(ctx); err != nil {
			return nil, fmt.Errorf("redis ping failed: %w", err)
		}
		return r, nil

	case "memory":
		return NewMemory(l2TTL, cfg.MaxL1Items), nil

	case "tiered":
		memory := NewMemory(l1TTL, cfg.MaxL1Items)
		redis := NewRedis(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB, l2TTL)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := redis.Ping(ctx); err != nil {
			return nil, fmt.Errorf("tiered: redis ping failed: %w", err)
		}
		return NewTiered(memory, redis, l1TTL, l2TTL), nil

	default:
		return nil, fmt.Errorf("unknown cache type: %q (expected memory, redis, or tiered)", cfg.Type)
	}
}
