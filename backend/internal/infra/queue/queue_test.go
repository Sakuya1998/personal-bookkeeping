package queue

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestStats(t *testing.T) {
	var s Stats
	s.Pending = 1
	s.Running = 2
	s.Completed = 3
	s.Failed = 4

	if s.Pending != 1 || s.Running != 2 || s.Completed != 3 || s.Failed != 4 {
		t.Fatal("Stats fields mismatch")
	}
}

func TestTask(t *testing.T) {
	task := Task{
		ID:      "task-1",
		Type:    "test",
		Payload: "hello",
	}
	if task.ID != "task-1" || task.Type != "test" {
		t.Fatal("Task fields mismatch")
	}
	payload, ok := task.Payload.(string)
	if !ok || payload != "hello" {
		t.Fatal("Task payload mismatch")
	}
}

// testQueue is a simplified queue for testing the interface contract.
type testQueue struct {
	registered map[string]HandlerFunc
	submitted  atomic.Int64
	started    atomic.Bool
}

func (q *testQueue) Register(t string, fn HandlerFunc) {
	q.registered[t] = fn
}

func (q *testQueue) Submit(_ context.Context, task Task) error {
	if !q.started.Load() {
		return errors.New("not started")
	}
	q.submitted.Add(1)
	return nil
}

func (q *testQueue) Start(_ context.Context) {
	q.started.Store(true)
}

func (q *testQueue) Shutdown(_ context.Context) error {
	q.started.Store(false)
	return nil
}

func (q *testQueue) Stats() Stats {
	return Stats{Pending: 0, Running: 0, Completed: int(q.submitted.Load()), Failed: 0}
}

func TestQueueInterface(t *testing.T) {
	q := &testQueue{registered: make(map[string]HandlerFunc)}

	// Test Register
	called := make(chan struct{})
	q.Register("greet", func(_ context.Context, _ Task) error {
		close(called)
		return nil
	})

	if fn, ok := q.registered["greet"]; !ok {
		t.Fatal("handler not registered")
	} else {
		fn(context.Background(), Task{Type: "greet"})
		select {
		case <-called:
		case <-time.After(time.Second):
			t.Fatal("handler was not called")
		}
	}

	// Test Start
	q.Start(context.Background())
	if !q.started.Load() {
		t.Fatal("queue should be started")
	}

	// Test Submit
	if err := q.Submit(context.Background(), Task{ID: "t1", Type: "greet"}); err != nil {
		t.Fatalf("Submit failed: %v", err)
	}
	if q.submitted.Load() != 1 {
		t.Fatalf("want 1 submitted, got %d", q.submitted.Load())
	}

	// Test Stats
	s := q.Stats()
	if s.Completed != 1 {
		t.Fatalf("want 1 completed, got %d", s.Completed)
	}

	// Test Shutdown
	if err := q.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown failed: %v", err)
	}
	if q.started.Load() {
		t.Fatal("queue should be stopped after shutdown")
	}
}

func TestHandlerFuncSignature(t *testing.T) {
	// Verify HandlerFunc can be called with context and task
	fn := HandlerFunc(func(ctx context.Context, task Task) error {
		if task.Type == "fail" {
			return errors.New("failed")
		}
		return nil
	})

	if err := fn(context.Background(), Task{Type: "ok"}); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}
	if err := fn(context.Background(), Task{Type: "fail"}); err == nil {
		t.Fatal("expected error, got nil")
	}
}
