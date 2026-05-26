package cache

import (
	"context"
	"log/slog"
	"time"
)

// TieredCache implements Cache with L1 (memory) + L2 (Redis).
//
// Read path:  L1 → hit? return. L1 miss → L2 → write back to L1.
// Write path: L1 + L2 (write-through).
// Delete:     L1 + L2 (best-effort).
//
// L2 (Redis) failures are tolerated — L1 still serves cached data.
// L1 evictions happen based on L1 TTL + maxItems FIFO.
type TieredCache struct {
	l1    Cache
	l2    Cache
	l1TTL time.Duration // TTL for L1 entries
	l2TTL time.Duration // TTL for L2 entries
}

// NewTiered creates a multi-level cache.
// l1: local memory cache. l2: remote cache (Redis).
// l1TTL: how long entries stay in L1 (short). l2TTL: L2 expiration (long).
func NewTiered(l1 Cache, l2 Cache, l1TTL, l2TTL time.Duration) *TieredCache {
	slog.Info("tiered cache initialized",
		"l1_type", "memory",
		"l2_type", "redis",
		"l1_ttl", l1TTL,
		"l2_ttl", l2TTL,
	)
	return &TieredCache{l1: l1, l2: l2, l1TTL: l1TTL, l2TTL: l2TTL}
}

func (t *TieredCache) Get(ctx context.Context, key string) (string, error) {
	// L1 hit
	val, err := t.l1.Get(ctx, key)
	if err == nil {
		return val, nil
	}
	if err != ErrMiss {
		slog.Warn("tiered: L1 get error", "key", key, "error", err)
	}

	// L2 lookup
	val, err = t.l2.Get(ctx, key)
	if err != nil {
		return "", err // propagate ErrMiss or connection error
	}

	// Write back to L1 (best-effort)
	if setErr := t.l1.Set(ctx, key, val, t.l1TTL); setErr != nil {
		slog.Warn("tiered: L1 write-back failed", "key", key, "error", setErr)
	}
	return val, nil
}

func (t *TieredCache) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	// Use configured L2 TTL unless caller overrides with a specific ttl
	l2TTL := t.l2TTL
	if ttl > 0 {
		l2TTL = ttl
	}
	// L1 always uses the shorter l1TTL regardless of caller's ttl
	// (L1 is just a hot cache, L2 holds the authoritative TTL)
	if err := t.l1.Set(ctx, key, value, t.l1TTL); err != nil {
		slog.Warn("tiered: L1 set failed", "key", key, "error", err)
	}
	if err := t.l2.Set(ctx, key, value, l2TTL); err != nil {
		slog.Warn("tiered: L2 set failed", "key", key, "error", err)
	}
	return nil
}

func (t *TieredCache) Delete(ctx context.Context, key string) error {
	_ = t.l1.Delete(ctx, key) // best-effort
	_ = t.l2.Delete(ctx, key) // best-effort
	return nil
}

func (t *TieredCache) Exists(ctx context.Context, key string) (bool, error) {
	ok, err := t.l1.Exists(ctx, key)
	if err == nil && ok {
		return true, nil
	}
	return t.l2.Exists(ctx, key)
}

func (t *TieredCache) Flush(ctx context.Context) error {
	_ = t.l1.Flush(ctx)
	return t.l2.Flush(ctx)
}

func (t *TieredCache) Close() error {
	_ = t.l1.Close()
	return t.l2.Close()
}
