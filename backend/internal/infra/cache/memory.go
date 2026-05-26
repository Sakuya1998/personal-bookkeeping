package cache

import (
	"context"
	"sync"
	"time"
)

type memEntry struct {
	value   string
	expires time.Time
}

// MemoryCache 内存缓存，无外部依赖
type MemoryCache struct {
	mu          sync.RWMutex
	entries     map[string]memEntry
	entryOrder  []string // FIFO order for maxItems eviction
	defaultTTL  time.Duration
	maxItems    int
	stopCh      chan struct{}
}

// NewMemory 创建内存缓存。
// defaultTTL: 默认过期时间。maxItems: 最大条目数（0 = 无限制）。
func NewMemory(defaultTTL time.Duration, maxItems int) *MemoryCache {
	c := &MemoryCache{
		entries:    make(map[string]memEntry),
		defaultTTL: defaultTTL,
		maxItems:   maxItems,
		stopCh:     make(chan struct{}),
	}
	go c.evictLoop()
	return c
}

func (c *MemoryCache) Get(_ context.Context, key string) (string, error) {
	c.mu.RLock()
	e, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok {
		return "", ErrMiss
	}
	if !e.expires.IsZero() && e.expires.Before(time.Now()) {
		c.mu.Lock()
		delete(c.entries, key)
		c.removeOrder(key)
		c.mu.Unlock()
		return "", ErrMiss
	}
	return e.value, nil
}

func (c *MemoryCache) Set(_ context.Context, key string, value string, ttl time.Duration) error {
	expires := time.Now().Add(ttl)

	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict oldest entries if over limit
	if c.maxItems > 0 && len(c.entries) >= c.maxItems && c.entries[key] == (memEntry{}) {
		c.evictFIFO()
	}

	// If key already exists, update in place (no order change)
	if _, exists := c.entries[key]; !exists {
		c.entryOrder = append(c.entryOrder, key)
	}
	c.entries[key] = memEntry{value: value, expires: expires}
	return nil
}

func (c *MemoryCache) Delete(_ context.Context, key string) error {
	c.mu.Lock()
	delete(c.entries, key)
	c.removeOrder(key)
	c.mu.Unlock()
	return nil
}

func (c *MemoryCache) Exists(_ context.Context, key string) (bool, error) {
	c.mu.RLock()
	e, ok := c.entries[key]
	c.mu.RUnlock()
	if !ok {
		return false, nil
	}
	if !e.expires.IsZero() && e.expires.Before(time.Now()) {
		c.mu.Lock()
		delete(c.entries, key)
		c.removeOrder(key)
		c.mu.Unlock()
		return false, nil
	}
	return true, nil
}

func (c *MemoryCache) Flush(_ context.Context) error {
	c.mu.Lock()
	c.entries = make(map[string]memEntry)
	c.entryOrder = c.entryOrder[:0]
	c.mu.Unlock()
	return nil
}

func (c *MemoryCache) Close() error {
	close(c.stopCh)
	return nil
}

// removeOrder removes key from entryOrder. Must hold c.mu.
func (c *MemoryCache) removeOrder(key string) {
	for i, k := range c.entryOrder {
		if k == key {
			c.entryOrder = append(c.entryOrder[:i], c.entryOrder[i+1:]...)
			return
		}
	}
}

// evictFIFO removes the oldest entries until under maxItems. Must hold c.mu.
func (c *MemoryCache) evictFIFO() {
	for len(c.entries) >= c.maxItems && len(c.entryOrder) > 0 {
		oldest := c.entryOrder[0]
		c.entryOrder = c.entryOrder[1:]
		delete(c.entries, oldest)
	}
}

func (c *MemoryCache) evictLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			c.mu.Lock()
			for k, e := range c.entries {
				if !e.expires.IsZero() && e.expires.Before(now) {
					delete(c.entries, k)
					c.removeOrder(k)
				}
			}
			c.mu.Unlock()
		case <-c.stopCh:
			return
		}
	}
}
