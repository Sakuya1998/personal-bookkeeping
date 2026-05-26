package queue

import (
	"context"
	"errors"
	"testing"

	"personal-bookkeeping/internal/infra/config"
)

// startGuardedQueue verifies the start-before-submit contract.
type startGuardedQueue struct {
	started bool
}

func (q *startGuardedQueue) Register(_ string, _ HandlerFunc) {}
func (q *startGuardedQueue) Submit(_ context.Context, _ Task) error {
	if !q.started {
		return errors.New("queue: not started")
	}
	return nil
}
func (q *startGuardedQueue) Start(_ context.Context) {
	q.started = true
}
func (q *startGuardedQueue) Shutdown(_ context.Context) error {
	q.started = false
	return nil
}
func (q *startGuardedQueue) Stats() Stats {
	return Stats{}
}

func TestQueueLifecycle_SubmitBeforeStart(t *testing.T) {
	q := &startGuardedQueue{}
	err := q.Submit(context.Background(), Task{ID: "t1", Type: "test"})
	if err == nil {
		t.Fatal("expected error when submitting before start, got nil")
	}
}

func TestQueueLifecycle_SubmitAfterStart(t *testing.T) {
	q := &startGuardedQueue{}
	q.Start(context.Background())
	err := q.Submit(context.Background(), Task{ID: "t1", Type: "test"})
	if err != nil {
		t.Fatalf("expected nil after start, got %v", err)
	}
}

func TestQueueLifecycle_ShutdownRejects(t *testing.T) {
	q := &startGuardedQueue{}
	q.Start(context.Background())
	if err := q.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
	err := q.Submit(context.Background(), Task{ID: "t1", Type: "test"})
	if err == nil {
		t.Fatal("expected error after shutdown, got nil")
	}
}

func TestQueueLifecycle_DoubleStart(t *testing.T) {
	q := &startGuardedQueue{}
	q.Start(context.Background())
	q.Start(context.Background()) // should not panic
	if !q.started {
		t.Fatal("queue should remain started after double start")
	}
}

func TestQueueLifecycle_DoubleShutdown(t *testing.T) {
	q := &startGuardedQueue{}
	q.Start(context.Background())
	_ = q.Shutdown(context.Background())
	err := q.Shutdown(context.Background()) // should not panic
	if err != nil {
		t.Logf("double shutdown error (acceptable): %v", err)
	}
}

// emptyQueue tests that all interface methods can be called on a minimal implementation.
type emptyQueue struct{}

func (e *emptyQueue) Register(_ string, _ HandlerFunc) {}
func (e *emptyQueue) Submit(_ context.Context, _ Task) error {
	return errors.New("not implemented")
}
func (e *emptyQueue) Start(_ context.Context)            {}
func (e *emptyQueue) Shutdown(_ context.Context) error   { return nil }
func (e *emptyQueue) Stats() Stats                       { return Stats{} }

func TestQueueEmptyInterface(t *testing.T) {
	var q Queue = &emptyQueue{}
	q.Register("test", func(ctx context.Context, task Task) error { return nil })
	q.Start(context.Background())
	_ = q.Stats()
	_ = q.Shutdown(context.Background())
	// No panic = pass
}

// ---------- Factory tests ----------

func TestQueueFactoryDisabled(t *testing.T) {
	// nil when disabled — no external dependencies
	q, err := NewFromConfig(&config.QueueConfig{Enabled: false})
	if err != nil {
		t.Fatalf("disabled queue: expected nil error, got %v", err)
	}
	if q != nil {
		t.Fatal("disabled queue: expected nil queue, got non-nil")
	}
}

func TestQueueFactoryUnknownType(t *testing.T) {
	q, err := NewFromConfig(&config.QueueConfig{
		Enabled: true,
		Type:    "unknown",
	})
	if err == nil {
		t.Fatal("unknown queue type: expected error, got nil")
	}
	if q != nil {
		t.Fatal("unknown queue type: expected nil queue, got non-nil")
	}
}
