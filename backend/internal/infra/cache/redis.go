package cache

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// RedisCache 基于 go-redis 的缓存实现
type RedisCache struct {
	client     *goredis.Client
	defaultTTL time.Duration
}

func NewRedis(addr, password string, db int, defaultTTL time.Duration) *RedisCache {
	rdb := goredis.NewClient(&goredis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &RedisCache{client: rdb, defaultTTL: defaultTTL}
}

func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == goredis.Nil {
		return "", ErrMiss
	}
	return val, err
}

func (r *RedisCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	return n > 0, err
}

func (r *RedisCache) Flush(ctx context.Context) error {
	return r.client.FlushDB(ctx).Err()
}

func (r *RedisCache) Close() error {
	return r.client.Close()
}

// Ping 健康检查
func (r *RedisCache) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
