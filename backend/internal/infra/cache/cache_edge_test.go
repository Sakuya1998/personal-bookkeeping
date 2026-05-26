package cache

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// ---------- Tiered edge cases ----------

func TestTieredConcurrentAccess(t *testing.T) {
	l1 := NewMemory(time.Minute, 100)
	l2 := NewMemory(time.Hour, 100)
	tc := NewTiered(l1, l2, time.Minute, time.Hour)
	defer tc.Close()

	ctx := context.Background()
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := "concurrent:" + string(rune('a'+n%26))
			_ = tc.Set(ctx, key, "val", time.Minute)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := "concurrent:" + string(rune('a'+n%26))
			_, _ = tc.Get(ctx, key)
		}(i)
	}

	// Concurrent deletes
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := "concurrent:" + string(rune('a'+n%26))
			_ = tc.Delete(ctx, key)
		}(i)
	}

	// Concurrent exists checks
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := "concurrent:" + string(rune('a'+n%26))
			_, _ = tc.Exists(ctx, key)
		}(i)
	}

	wg.Wait()
	// No data race should occur
}

func TestTieredL2ReturnsEmptyString(t *testing.T) {
	l1 := NewMemory(time.Minute, 0)
	l2 := NewMemory(time.Hour, 0)
	tc := NewTiered(l1, l2, time.Minute, time.Hour)
	defer tc.Close()

	ctx := context.Background()
	// Store empty string in L2
	l2.Set(ctx, "empty-key", "", time.Hour)

	val, err := tc.Get(ctx, "empty-key")
	if err != nil {
		t.Fatalf("Get empty value: expected nil, got %v", err)
	}
	if val != "" {
		t.Fatalf("Get empty value: expected '', got %q", val)
	}

	// Write back to L1 should have happened
	v, _ := l1.Get(ctx, "empty-key")
	if v != "" {
		t.Fatalf("L1 write-back: expected '', got %q", v)
	}
}

func TestTieredL2GetFailureDegradation(t *testing.T) {
	l1 := NewMemory(time.Minute, 0)
	broken := &panicCache{}
	tc := NewTiered(l1, broken, time.Minute, time.Hour)
	defer tc.Close()

	ctx := context.Background()
	l1.Set(ctx, "l1-only", "cached", time.Minute)

	// L2 is broken, but L1 hit should still work
	val, err := tc.Get(ctx, "l1-only")
	if err != nil {
		t.Fatalf("L1 hit with broken L2: expected nil error, got %v", err)
	}
	if val != "cached" {
		t.Fatalf("L1 hit with broken L2: expected 'cached', got %q", val)
	}
}

func TestTieredDeleteNonExistent(t *testing.T) {
	l1 := NewMemory(time.Minute, 0)
	l2 := NewMemory(time.Hour, 0)
	tc := NewTiered(l1, l2, time.Minute, time.Hour)
	defer tc.Close()

	ctx := context.Background()
	err := tc.Delete(ctx, "nonexistent-key")
	if err != nil {
		t.Fatalf("Delete non-existent: expected nil, got %v", err)
	}
}

func TestTieredFlushEmpty(t *testing.T) {
	l1 := NewMemory(time.Minute, 0)
	l2 := NewMemory(time.Hour, 0)
	tc := NewTiered(l1, l2, time.Minute, time.Hour)
	defer tc.Close()

	ctx := context.Background()
	err := tc.Flush(ctx)
	if err != nil {
		t.Fatalf("Flush empty: expected nil, got %v", err)
	}
}

func TestTieredSetWithZeroTTLUsesDefault(t *testing.T) {
	l1 := NewMemory(50*time.Millisecond, 0)
	l2 := NewMemory(time.Hour, 0)
	tc := NewTiered(l1, l2, 50*time.Millisecond, time.Hour)
	defer tc.Close()

	ctx := context.Background()
	// ttl=0 means use defaults: L2 gets hour, L1 gets 50ms
	tc.Set(ctx, "k", "v", 0)

	// L1 should expire quickly
	time.Sleep(100 * time.Millisecond)
	_, err := l1.Get(ctx, "k")
	if err != ErrMiss {
		t.Fatal("L1 should have expired with zero TTL (uses l1TTL=50ms)")
	}

	// L2 should still exist (hour TTL)
	v, err := l2.Get(ctx, "k")
	if err != nil {
		t.Fatalf("L2 should still have value, got err: %v", err)
	}
	if v != "v" {
		t.Fatalf("L2: expected 'v', got %q", v)
	}
}

func TestTieredExistsAfterDelete(t *testing.T) {
	l1 := NewMemory(time.Minute, 0)
	l2 := NewMemory(time.Hour, 0)
	tc := NewTiered(l1, l2, time.Minute, time.Hour)
	defer tc.Close()

	ctx := context.Background()
	tc.Set(ctx, "k", "v", 0)
	tc.Delete(ctx, "k")

	ok, _ := tc.Exists(ctx, "k")
	if ok {
		t.Fatal("key should not exist after delete")
	}
}

func TestTieredL2SetFailureDegradation(t *testing.T) {
	l1 := NewMemory(time.Minute, 0)
	broken := &panicCache{}
	tc := NewTiered(l1, broken, time.Minute, time.Hour)
	defer tc.Close()

	ctx := context.Background()
	// Set should not panic even if L2 is broken (best-effort)
	err := tc.Set(ctx, "k", "v", time.Minute)
	if err != nil {
		t.Fatalf("Set with broken L2: expected nil (best-effort), got %v", err)
	}

	// L1 should still have the value
	v, err := l1.Get(ctx, "k")
	if err != nil {
		t.Fatalf("L1 should have value despite broken L2: %v", err)
	}
	if v != "v" {
		t.Fatalf("L1: expected 'v', got %q", v)
	}
}

// panicCache simulates a completely broken cache (panics on every call).
type panicCache struct{}

func (p *panicCache) Get(_ context.Context, _ string) (string, error) {
	return "", errors.New("broken L2")
}
func (p *panicCache) Set(_ context.Context, _ string, _ string, _ time.Duration) error {
	return errors.New("broken L2")
}
func (p *panicCache) Delete(_ context.Context, _ string) error {
	return errors.New("broken L2")
}
func (p *panicCache) Exists(_ context.Context, _ string) (bool, error) {
	return false, errors.New("broken L2")
}
func (p *panicCache) Flush(_ context.Context) error {
	return errors.New("broken L2")
}
func (p *panicCache) Close() error {
	return errors.New("broken L2")
}

// ---------- Memory edge cases ----------

func TestMemorySetGetLargeValue(t *testing.T) {
	c := NewMemory(time.Minute, 0)
	defer c.Close()

	ctx := context.Background()
	largeVal := string(make([]byte, 10000))
	for i := range largeVal {
		largeVal = largeVal[:i] + "x" + largeVal[i+1:]
	}

	if err := c.Set(ctx, "large", largeVal, time.Minute); err != nil {
		t.Fatalf("Set large value: %v", err)
	}
	v, err := c.Get(ctx, "large")
	if err != nil {
		t.Fatalf("Get large value: %v", err)
	}
	if len(v) != 10000 {
		t.Fatalf("expected 10000 bytes, got %d", len(v))
	}
}

func TestMemoryDeleteNonExistent(t *testing.T) {
	c := NewMemory(time.Minute, 0)
	defer c.Close()
	err := c.Delete(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("Delete non-existent: expected nil, got %v", err)
	}
}

func TestMemoryFlushEmpty(t *testing.T) {
	c := NewMemory(time.Minute, 0)
	defer c.Close()
	err := c.Flush(context.Background())
	if err != nil {
		t.Fatalf("Flush empty: expected nil, got %v", err)
	}
}

func TestMemoryExistsAfterExpiration(t *testing.T) {
	c := NewMemory(10*time.Millisecond, 0)
	defer c.Close()

	ctx := context.Background()
	c.Set(ctx, "k", "v", 10*time.Millisecond)
	time.Sleep(50 * time.Millisecond)

	ok, _ := c.Exists(ctx, "k")
	if ok {
		t.Fatal("expired key should not exist")
	}
}
