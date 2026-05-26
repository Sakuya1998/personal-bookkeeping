package cache

import (
	"context"
	"testing"
	"time"
)

// TestTieredL1Hit verifies that L1 returns immediately without hitting L2.
func TestTieredL1Hit(t *testing.T) {
	l1 := NewMemory(time.Minute, 0)
	l2 := NewMemory(time.Hour, 0)
	tc := NewTiered(l1, l2, time.Minute, time.Hour)
	defer tc.Close()

	ctx := context.Background()
	// Write to L1 directly (simulate write-back scenario)
	l1.Set(ctx, "k", "from-l1", time.Minute)

	val, err := tc.Get(ctx, "k")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "from-l1" {
		t.Fatalf("got %q, want %q", val, "from-l1")
	}
}

// TestTieredL1MissL2Hit verifies L1 miss → L2 hit → write back to L1.
func TestTieredL1MissL2Hit(t *testing.T) {
	l1 := NewMemory(time.Minute, 0)
	l2 := NewMemory(time.Hour, 0)
	tc := NewTiered(l1, l2, time.Minute, time.Hour)
	defer tc.Close()

	ctx := context.Background()
	// Write to L2 directly (simulate another instance cached it)
	l2.Set(ctx, "k", "from-l2", time.Hour)

	val, err := tc.Get(ctx, "k")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "from-l2" {
		t.Fatalf("got %q, want %q", val, "from-l2")
	}

	// Verify write-back to L1
	v, _ := l1.Get(ctx, "k")
	if v != "from-l2" {
		t.Fatalf("L1 write-back: got %q, want %q", v, "from-l2")
	}
}

// TestTieredMiss verifies both L1 and L2 miss returns ErrMiss.
func TestTieredMiss(t *testing.T) {
	l1 := NewMemory(time.Minute, 0)
	l2 := NewMemory(time.Hour, 0)
	tc := NewTiered(l1, l2, time.Minute, time.Hour)
	defer tc.Close()

	_, err := tc.Get(context.Background(), "nonexistent")
	if err != ErrMiss {
		t.Fatalf("want ErrMiss, got %v", err)
	}
}

// TestTieredSetWriteThrough verifies Set writes to both layers.
func TestTieredSetWriteThrough(t *testing.T) {
	l1 := NewMemory(time.Minute, 0)
	l2 := NewMemory(time.Hour, 0)
	tc := NewTiered(l1, l2, time.Minute, time.Hour)
	defer tc.Close()

	ctx := context.Background()
	tc.Set(ctx, "k", "v", 0)

	// Both layers should have it
	v1, _ := l1.Get(ctx, "k")
	if v1 != "v" {
		t.Fatalf("L1: got %q, want %q", v1, "v")
	}
	v2, _ := l2.Get(ctx, "k")
	if v2 != "v" {
		t.Fatalf("L2: got %q, want %q", v2, "v")
	}
}

// TestTieredSetWithTTL verifies caller-specified TTL is used for L2.
func TestTieredSetWithTTL(t *testing.T) {
	l1 := NewMemory(10*time.Millisecond, 0)
	l2 := NewMemory(time.Hour, 0)
	tc := NewTiered(l1, l2, 10*time.Millisecond, time.Hour)
	defer tc.Close()

	ctx := context.Background()
	tc.Set(ctx, "k", "v", 10*time.Millisecond)
	time.Sleep(50 * time.Millisecond)

	// L1 should be expired
	_, err := l1.Get(ctx, "k")
	if err != ErrMiss {
		t.Fatal("L1 should have expired")
	}

	// L2 should still exist (L2 was set with the caller's TTL but in our implementation
	// we always use l2TTL from config unless ttl > 0, and here ttl=10ms which is >0
	// Wait, let me check the implementation...
	// In TieredCache.Set: l2TTL is the configured one unless ttl > 0
	// ttl=10ms IS >0, so it overrides l2TTL.
	// So L2 should also be expired. That's expected behavior.
	_, err = l2.Get(ctx, "k")
	if err != ErrMiss {
		t.Fatal("L2 should also have expired since caller TTL overrode l2TTL")
	}
}

// TestTieredDelete removes from both layers.
func TestTieredDelete(t *testing.T) {
	l1 := NewMemory(time.Minute, 0)
	l2 := NewMemory(time.Hour, 0)
	tc := NewTiered(l1, l2, time.Minute, time.Hour)
	defer tc.Close()

	ctx := context.Background()
	l1.Set(ctx, "k", "v", time.Minute)
	l2.Set(ctx, "k", "v", time.Hour)

	tc.Delete(ctx, "k")

	if _, err := l1.Get(ctx, "k"); err != ErrMiss {
		t.Fatal("L1 should be empty after delete")
	}
	if _, err := l2.Get(ctx, "k"); err != ErrMiss {
		t.Fatal("L2 should be empty after delete")
	}
}

// TestTieredExists checks both layers.
func TestTieredExists(t *testing.T) {
	l1 := NewMemory(time.Minute, 0)
	l2 := NewMemory(time.Hour, 0)
	tc := NewTiered(l1, l2, time.Minute, time.Hour)
	defer tc.Close()

	ctx := context.Background()
	l2.Set(ctx, "k", "v", time.Hour)

	ok, _ := tc.Exists(ctx, "k")
	if !ok {
		t.Fatal("key in L2 should exist via tiered")
	}
}

// TestTieredFlush cleans both layers.
func TestTieredFlush(t *testing.T) {
	l1 := NewMemory(time.Minute, 0)
	l2 := NewMemory(time.Hour, 0)
	tc := NewTiered(l1, l2, time.Minute, time.Hour)
	defer tc.Close()

	ctx := context.Background()
	tc.Set(ctx, "k", "v", 0)
	tc.Flush(ctx)

	if _, err := l1.Get(ctx, "k"); err != ErrMiss {
		t.Fatal("L1 should be empty after flush")
	}
	if _, err := l2.Get(ctx, "k"); err != ErrMiss {
		t.Fatal("L2 should be empty after flush")
	}
}

// TestTieredL2FailureDegradation verifies L1 still works when L2 is down.
func TestTieredL2FailureDegradation(t *testing.T) {
	l1 := NewMemory(time.Minute, 0)
	// Use a broken L2
	l2 := &brokenCache{}
	tc := NewTiered(l1, l2, time.Minute, time.Hour)
	defer tc.Close()

	ctx := context.Background()
	l1.Set(ctx, "k", "l1-val", time.Minute)

	// Even though L2 is broken, L1 hit should work
	val, err := tc.Get(ctx, "k")
	if err != nil {
		t.Fatalf("Get should fall back to L1: %v", err)
	}
	if val != "l1-val" {
		t.Fatalf("got %q, want %q", val, "l1-val")
	}
}

// brokenCache simulates a broken L2 (all operations fail).
type brokenCache struct{}

func (b *brokenCache) Get(_ context.Context, _ string) (string, error) {
	return "", assertErr{msg: "broken"}
}
func (b *brokenCache) Set(_ context.Context, _ string, _ string, _ time.Duration) error {
	return assertErr{msg: "broken"}
}
func (b *brokenCache) Delete(_ context.Context, _ string) error          { return assertErr{msg: "broken"} }
func (b *brokenCache) Exists(_ context.Context, _ string) (bool, error)  { return false, assertErr{msg: "broken"} }
func (b *brokenCache) Flush(_ context.Context) error                     { return assertErr{msg: "broken"} }
func (b *brokenCache) Close() error                                      { return assertErr{msg: "broken"} }

type assertErr struct{ msg string }

func (a assertErr) Error() string { return a.msg }
