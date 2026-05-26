package cache

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestMemoryGetSet(t *testing.T) {
	c := NewMemory(time.Minute, 0)
	defer c.Close()

	ctx := context.Background()
	if err := c.Set(ctx, "key1", "val1", time.Minute); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	v, err := c.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if v != "val1" {
		t.Fatalf("got %q, want %q", v, "val1")
	}
}

func TestMemoryMiss(t *testing.T) {
	c := NewMemory(time.Minute, 0)
	defer c.Close()

	_, err := c.Get(context.Background(), "nonexistent")
	if err != ErrMiss {
		t.Fatalf("want ErrMiss, got %v", err)
	}
}

func TestMemoryDelete(t *testing.T) {
	c := NewMemory(time.Minute, 0)
	defer c.Close()

	ctx := context.Background()
	c.Set(ctx, "k", "v", time.Minute)
	c.Delete(ctx, "k")

	_, err := c.Get(ctx, "k")
	if err != ErrMiss {
		t.Fatalf("after delete: want ErrMiss, got %v", err)
	}
}

func TestMemoryExists(t *testing.T) {
	c := NewMemory(time.Minute, 0)
	defer c.Close()

	ctx := context.Background()
	ok, _ := c.Exists(ctx, "missing")
	if ok {
		t.Fatal("missing key should not exist")
	}

	c.Set(ctx, "k", "v", time.Minute)
	ok, _ = c.Exists(ctx, "k")
	if !ok {
		t.Fatal("existing key should exist")
	}
}

func TestMemoryExpiration(t *testing.T) {
	c := NewMemory(10*time.Millisecond, 0)
	defer c.Close()

	ctx := context.Background()
	c.Set(ctx, "k", "v", 10*time.Millisecond)
	time.Sleep(50 * time.Millisecond)

	_, err := c.Get(ctx, "k")
	if err != ErrMiss {
		t.Fatalf("expired key: want ErrMiss, got %v", err)
	}
}

func TestMemoryFlush(t *testing.T) {
	c := NewMemory(time.Minute, 0)
	defer c.Close()

	ctx := context.Background()
	c.Set(ctx, "a", "1", time.Minute)
	c.Set(ctx, "b", "2", time.Minute)
	c.Flush(ctx)

	_, err := c.Get(ctx, "a")
	if err != ErrMiss {
		t.Fatal("after flush, 'a' should be missing")
	}
	_, err = c.Get(ctx, "b")
	if err != ErrMiss {
		t.Fatal("after flush, 'b' should be missing")
	}
}

func TestMemorySetUpdatesExisting(t *testing.T) {
	c := NewMemory(time.Minute, 0)
	defer c.Close()

	ctx := context.Background()
	c.Set(ctx, "k", "old", time.Minute)
	c.Set(ctx, "k", "new", time.Minute)

	v, err := c.Get(ctx, "k")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if v != "new" {
		t.Fatalf("got %q, want %q", v, "new")
	}
}

func TestMemoryTTLOverride(t *testing.T) {
	c := NewMemory(time.Hour, 0)
	defer c.Close()

	ctx := context.Background()
	c.Set(ctx, "k", "v", 10*time.Millisecond)
	time.Sleep(50 * time.Millisecond)

	_, err := c.Get(ctx, "k")
	if err != ErrMiss {
		t.Fatal("short TTL should expire despite long default")
	}
}

func TestMemoryFIFOEviction(t *testing.T) {
	c := NewMemory(time.Hour, 3)
	defer c.Close()

	ctx := context.Background()
	c.Set(ctx, "a", "1", time.Hour)
	c.Set(ctx, "b", "2", time.Hour)
	c.Set(ctx, "c", "3", time.Hour)
	// Now at capacity (3). Adding "d" should evict "a" (oldest)
	c.Set(ctx, "d", "4", time.Hour)

	if _, err := c.Get(ctx, "a"); err != ErrMiss {
		t.Fatal("oldest entry 'a' should have been evicted")
	}
	if v, _ := c.Get(ctx, "d"); v != "4" {
		t.Fatalf("new entry 'd' should exist, got %v", v)
	}
}

func TestMemoryConcurrentAccess(t *testing.T) {
	c := NewMemory(time.Minute, 0)
	defer c.Close()

	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := "k" + string(rune('0'+n%10))
			c.Set(ctx, key, "v", time.Minute)
			c.Get(ctx, key)
			c.Exists(ctx, key)
		}(i)
	}
	wg.Wait()
}

func TestMemoryDefaultTTL(t *testing.T) {
	c := NewMemory(10*time.Millisecond, 0)
	defer c.Close()

	ctx := context.Background()
	c.Set(ctx, "k", "v", 0) // use default
	time.Sleep(50 * time.Millisecond)

	_, err := c.Get(ctx, "k")
	if err != ErrMiss {
		t.Fatal("with default TTL, key should expire")
	}
}
