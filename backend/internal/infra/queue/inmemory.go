package queue

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
)

// InMemoryQueue is a simple in-memory task queue that dispatches tasks
// to registered handlers in a worker pool. Designed for development and
// single-instance deployments where Redis/Kafka are overkill.
type InMemoryQueue struct {
	mu         sync.RWMutex
	handlers   map[string]HandlerFunc
	taskCh     chan Task
	workers    int
	maxRetries int
	started    atomic.Bool
	wg         sync.WaitGroup
	cancel     context.CancelFunc

	pending   atomic.Int64
	running   atomic.Int64
	completed atomic.Int64
	failed    atomic.Int64
}

// NewInMemory creates an in-memory queue with the given number of workers.
// maxRetries controls how many times a failed task is retried (0 = no retry).
func NewInMemory(workers, maxRetries int) *InMemoryQueue {
	if workers <= 0 {
		workers = 1
	}
	return &InMemoryQueue{
		handlers:   make(map[string]HandlerFunc),
		taskCh:     make(chan Task, 1000),
		workers:    workers,
		maxRetries: maxRetries,
	}
}

// Register registers a handler for a given task type.
func (q *InMemoryQueue) Register(taskType string, fn HandlerFunc) {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.handlers[taskType] = fn
}

// Submit enqueues a task for processing.
func (q *InMemoryQueue) Submit(ctx context.Context, task Task) error {
	if !q.started.Load() {
		return fmt.Errorf("queue not started")
	}
	select {
	case q.taskCh <- task:
		q.pending.Add(1)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Start launches the worker goroutines.
func (q *InMemoryQueue) Start(ctx context.Context) {
	if q.started.Load() {
		return
	}
	q.started.Store(true)

	ctx, q.cancel = context.WithCancel(ctx)
	for i := range q.workers {
		q.wg.Add(1)
		go q.worker(ctx, i)
	}
	slog.Info("inmemory queue started", "workers", q.workers)
}

// Shutdown gracefully stops the queue, waiting for in-flight tasks to complete.
func (q *InMemoryQueue) Shutdown(ctx context.Context) error {
	if !q.started.Load() {
		return nil
	}
	if q.cancel != nil {
		q.cancel()
	}

	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Stats returns current queue statistics.
func (q *InMemoryQueue) Stats() Stats {
	return Stats{
		Pending:   int(q.pending.Load()),
		Running:   int(q.running.Load()),
		Completed: int(q.completed.Load()),
		Failed:    int(q.failed.Load()),
	}
}

// worker processes tasks from the queue.
func (q *InMemoryQueue) worker(ctx context.Context, id int) {
	defer q.wg.Done()
	slog.Debug("inmemory queue worker started", "worker_id", id)

	for {
		select {
		case <-ctx.Done():
			return
		case task := <-q.taskCh:
			q.pending.Add(-1)
			q.running.Add(1)
			q.processTask(ctx, task)
			q.running.Add(-1)
		}
	}
}

// processTask dispatches a single task to its registered handler with retry.
func (q *InMemoryQueue) processTask(ctx context.Context, task Task) {
	q.mu.RLock()
	handler, ok := q.handlers[task.Type]
	q.mu.RUnlock()

	if !ok {
		slog.Warn("inmemory queue: no handler registered for task type", "type", task.Type)
		q.failed.Add(1)
		return
	}

	var err error
	for attempt := 0; attempt <= q.maxRetries; attempt++ {
		if attempt > 0 {
			slog.Info("inmemory queue: retrying task", "type", task.Type, "id", task.ID, "attempt", attempt)
		}
		err = handler(ctx, task)
		if err == nil {
			q.completed.Add(1)
			return
		}
		select {
		case <-ctx.Done():
			return
		default:
		}
	}

	slog.Error("inmemory queue: task failed after retries", "type", task.Type, "id", task.ID, "error", err, "retries", q.maxRetries)
	q.failed.Add(1)
}
