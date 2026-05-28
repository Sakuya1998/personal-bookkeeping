package middleware

import (
	"testing"
	"time"
)

func TestRateLimiter_AllowWithinLimit(t *testing.T) {
	rl := NewRateLimiter(5, time.Minute)
	for i := 0; i < 5; i++ {
		if !rl.Allow("192.168.1.1") {
			t.Fatalf("request %d: expected allowed, got rejected", i+1)
		}
	}
}

func TestRateLimiter_RejectWhenExceeded(t *testing.T) {
	rl := NewRateLimiter(3, time.Minute)
	for i := 0; i < 3; i++ {
		rl.Allow("10.0.0.1")
	}
	if rl.Allow("10.0.0.1") {
		t.Fatal("expected rejected after rate exceeded")
	}
}

func TestRateLimiter_DifferentIPsIndependent(t *testing.T) {
	rl := NewRateLimiter(2, time.Minute)

	// IP A uses all its budget
	rl.Allow("10.0.0.1")
	rl.Allow("10.0.0.1")
	if rl.Allow("10.0.0.1") {
		t.Fatal("IP A should be rejected")
	}

	// IP B should still be allowed
	if !rl.Allow("10.0.0.2") {
		t.Fatal("IP B should be allowed (different from IP A)")
	}
	if !rl.Allow("10.0.0.2") {
		t.Fatal("IP B second request should be allowed")
	}
	if rl.Allow("10.0.0.2") {
		t.Fatal("IP B should be rejected after exceeding limit")
	}
}

func TestRateLimiter_WindowSliding(t *testing.T) {
	// Use a very short window so we can test sliding
	rl := NewRateLimiter(2, 50*time.Millisecond)

	rl.Allow("10.0.0.1")
	rl.Allow("10.0.0.1")
	if rl.Allow("10.0.0.1") {
		t.Fatal("expected rejected")
	}

	time.Sleep(60 * time.Millisecond)

	// After window slides, should be allowed again
	if !rl.Allow("10.0.0.1") {
		t.Fatal("expected allowed after window slides")
	}
}

func TestRateLimiter_ZeroRate(t *testing.T) {
	rl := NewRateLimiter(0, time.Minute)
	if rl.Allow("10.0.0.1") {
		t.Fatal("with rate=0, no request should be allowed")
	}
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	rl := NewRateLimiter(100, time.Minute)
	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				rl.Allow("10.0.0.1")
			}
			done <- struct{}{}
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
	// Should not panic (no data race)
}

func TestRateLimiter_SingleBurst(t *testing.T) {
	rl := NewRateLimiter(1, time.Minute)
	if !rl.Allow("10.0.0.1") {
		t.Fatal("first request should be allowed")
	}
	if rl.Allow("10.0.0.1") {
		t.Fatal("second request should be rejected")
	}
}
