package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

// RedisQueue 基于 Redis Streams 的队列实现。
// 使用消费者组实现消息分发，失败后重新入队带重试计数。
type RedisQueue struct {
	client     *goredis.Client
	stream     string
	group      string
	workers    int
	maxRetries int
	handlers   map[string]HandlerFunc
	mu         sync.RWMutex
	started    atomic.Bool
	stopCh     chan struct{}
	wg         sync.WaitGroup

	pending   atomic.Int64
	running   atomic.Int64
	completed atomic.Int64
	failed    atomic.Int64
}

func NewRedis(addr, password string, db int, stream, group string, workers, maxRetries int) *RedisQueue {
	return &RedisQueue{
		client: goredis.NewClient(&goredis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
		stream:     stream,
		group:      group,
		workers:    workers,
		maxRetries: maxRetries,
		handlers:   make(map[string]HandlerFunc),
		stopCh:     make(chan struct{}),
	}
}

func (q *RedisQueue) Register(taskType string, fn HandlerFunc) {
	q.mu.Lock()
	q.handlers[taskType] = fn
	q.mu.Unlock()
}

func (q *RedisQueue) Submit(ctx context.Context, task Task) error {
	payload, err := json.Marshal(task.Payload)
	if err != nil {
		return fmt.Errorf("redis queue: marshal payload: %w", err)
	}
	if task.ID == "" {
		task.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	}
	q.pending.Add(1)
	err = q.client.XAdd(ctx, &goredis.XAddArgs{
		Stream: q.stream,
		Values: map[string]any{
			"id":      task.ID,
			"type":    task.Type,
			"payload": string(payload),
			"retries": "0",
		},
	}).Err()
	if err != nil {
		q.pending.Add(-1)
		return fmt.Errorf("redis queue: xadd: %w", err)
	}
	return nil
}

func (q *RedisQueue) Start(ctx context.Context) {
	if q.started.Swap(true) {
		return
	}
	// Create consumer group (MKSTREAM auto-creates stream if it doesn't exist)
	if err := q.client.XGroupCreateMkStream(ctx, q.stream, q.group, "0").Err(); err != nil {
		// BUSYGROUP: group already exists — not an error
		slog.Debug("redis queue group create", "error", err)
	}
	for i := range q.workers {
		q.wg.Add(1)
		go q.worker(i)
	}
	slog.Info("redis queue started", "stream", q.stream, "group", q.group, "workers", q.workers)
}

func (q *RedisQueue) Shutdown(ctx context.Context) error {
	if !q.started.Load() {
		return nil
	}
	slog.Info("redis queue shutting down")
	q.started.Store(false)
	close(q.stopCh)

	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		slog.Info("redis queue shut down",
			"completed", q.completed.Load(),
			"failed", q.failed.Load(),
		)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (q *RedisQueue) Stats() Stats {
	return Stats{
		Pending:   int(q.pending.Load()),
		Running:   int(q.running.Load()),
		Completed: int(q.completed.Load()),
		Failed:    int(q.failed.Load()),
	}
}

func (q *RedisQueue) worker(id int) {
	defer q.wg.Done()
	consumer := fmt.Sprintf("worker-%d", id)

	for {
		select {
		case <-q.stopCh:
			return
		default:
		}
		result, err := q.client.XReadGroup(context.Background(), &goredis.XReadGroupArgs{
			Group:    q.group,
			Consumer: consumer,
			Streams:  []string{q.stream, ">"},
			Count:    1,
			Block:    2 * time.Second,
		}).Result()
		if err != nil {
			continue
		}
		for _, s := range result {
			for _, msg := range s.Messages {
				q.processMessage(msg)
			}
		}
	}
}

func (q *RedisQueue) processMessage(msg goredis.XMessage) {
	q.pending.Add(-1)
	q.running.Add(1)
	defer q.running.Add(-1)

	taskType, _ := msg.Values["type"].(string)
	payloadStr, _ := msg.Values["payload"].(string)
	retriesStr, _ := msg.Values["retries"].(string)

	q.mu.RLock()
	fn, ok := q.handlers[taskType]
	q.mu.RUnlock()

	// No handler registered — ack and skip
	if !ok {
		slog.Warn("redis queue: no handler for task type", "type", taskType, "msg_id", msg.ID)
		_ = q.client.XAck(context.Background(), q.stream, q.group, msg.ID)
		q.completed.Add(1)
		return
	}

	task := Task{Type: taskType}
	if id, ok := msg.Values["id"].(string); ok {
		task.ID = id
	}
	if payloadStr != "" {
		var payload any
		if err := json.Unmarshal([]byte(payloadStr), &payload); err == nil {
			task.Payload = payload
		}
	}

	err := fn(context.Background(), task)
	if err == nil {
		_ = q.client.XAck(context.Background(), q.stream, q.group, msg.ID)
		q.completed.Add(1)
		return
	}
	slog.Warn("redis queue: task failed", "type", taskType, "msg_id", msg.ID, "error", err)

	retries := 0
	if retriesStr != "" {
		_, _ = fmt.Sscanf(retriesStr, "%d", &retries)
	}
	retries++

	if retries > q.maxRetries {
		slog.Error("redis queue: task exceeded max retries", "type", taskType, "msg_id", msg.ID, "retries", retries)
		_ = q.client.XAck(context.Background(), q.stream, q.group, msg.ID)
		q.failed.Add(1)
		return
	}

	// Re-queue with incremented retry count
	_ = q.client.XAck(context.Background(), q.stream, q.group, msg.ID)
	if err := q.client.XAdd(context.Background(), &goredis.XAddArgs{
		Stream: q.stream,
		Values: map[string]any{
			"id":      task.ID,
			"type":    taskType,
			"payload": payloadStr,
			"retries": fmt.Sprintf("%d", retries),
		},
	}).Err(); err != nil {
		slog.Error("redis queue: failed to re-queue task", "type", taskType, "error", err)
	}
	q.pending.Add(1)
}
