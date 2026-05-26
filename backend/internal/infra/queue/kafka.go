package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/IBM/sarama"
)

// KafkaQueue 基于 Kafka 的队列实现。
// 使用 sarama 客户端，每个 task 作为一条 Kafka 消息。
type KafkaQueue struct {
	producer   sarama.SyncProducer
	consumer   sarama.ConsumerGroup
	topic      string
	groupID    string
	workers    int
	maxRetries int
	handlers   map[string]HandlerFunc
	mu         sync.RWMutex
	started    atomic.Bool
	cancel     context.CancelFunc
	wg         sync.WaitGroup

	pending   atomic.Int64
	running   atomic.Int64
	completed atomic.Int64
	failed    atomic.Int64
}

func NewKafka(brokers []string, topic, groupID string, workers, maxRetries int) (*KafkaQueue, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
		sarama.NewBalanceStrategyRoundRobin(),
	}
	config.Version = sarama.V2_6_0_0

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("kafka: create producer: %w", err)
	}
	consumer, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		producer.Close()
		return nil, fmt.Errorf("kafka: create consumer group: %w", err)
	}
	return &KafkaQueue{
		producer:   producer,
		consumer:   consumer,
		topic:      topic,
		groupID:    groupID,
		workers:    workers,
		maxRetries: maxRetries,
		handlers:   make(map[string]HandlerFunc),
	}, nil
}

func (q *KafkaQueue) Register(taskType string, fn HandlerFunc) {
	q.mu.Lock()
	q.handlers[taskType] = fn
	q.mu.Unlock()
}

func (q *KafkaQueue) Submit(ctx context.Context, task Task) error {
	body, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("kafka: marshal task: %w", err)
	}
	q.pending.Add(1)
	_, _, err = q.producer.SendMessage(&sarama.ProducerMessage{
		Topic: q.topic,
		Key:   sarama.StringEncoder(task.Type),
		Value: sarama.ByteEncoder(body),
	})
	if err != nil {
		q.pending.Add(-1)
		return fmt.Errorf("kafka: send: %w", err)
	}
	return nil
}

func (q *KafkaQueue) Start(ctx context.Context) {
	if q.started.Swap(true) {
		return
	}
	ctx, q.cancel = context.WithCancel(ctx)
	q.wg.Add(1)
	go func() {
		defer q.wg.Done()
		handler := &kafkaConsumerGroupHandler{
			queue: q,
		}
		for {
			if err := q.consumer.Consume(ctx, []string{q.topic}, handler); err != nil {
				slog.Error("kafka: consume error", "error", err)
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()
	slog.Info("kafka queue started", "topic", q.topic, "group", q.groupID, "workers", q.workers)
}

func (q *KafkaQueue) Shutdown(ctx context.Context) error {
	if !q.started.Load() {
		return nil
	}
	slog.Info("kafka queue shutting down")
	q.cancel()
	q.started.Store(false)

	done := make(chan struct{})
	go func() {
		q.wg.Wait()
		_ = q.consumer.Close()
		_ = q.producer.Close()
		close(done)
	}()
	select {
	case <-done:
		slog.Info("kafka queue shut down",
			"completed", q.completed.Load(),
			"failed", q.failed.Load(),
		)
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (q *KafkaQueue) Stats() Stats {
	return Stats{
		Pending:   int(q.pending.Load()),
		Running:   int(q.running.Load()),
		Completed: int(q.completed.Load()),
		Failed:    int(q.failed.Load()),
	}
}

// processMessage handles a single Kafka message with retry logic.
func (q *KafkaQueue) processMessage(msg *sarama.ConsumerMessage) {
	q.pending.Add(-1)
	q.running.Add(1)
	defer q.running.Add(-1)

	var task Task
	if err := json.Unmarshal(msg.Value, &task); err != nil {
		slog.Error("kafka: unmarshal task", "error", err)
		q.completed.Add(1)
		return
	}

	q.mu.RLock()
	fn, ok := q.handlers[task.Type]
	q.mu.RUnlock()
	if !ok {
		slog.Warn("kafka: no handler for task type", "type", task.Type)
		q.completed.Add(1)
		return
	}

	var lastErr error
	for attempt := 0; attempt <= q.maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*2) * time.Second)
		}
		lastErr = fn(context.Background(), task)
		if lastErr == nil {
			q.completed.Add(1)
			return
		}
		slog.Warn("kafka: task failed", "type", task.Type, "attempt", attempt, "error", lastErr)
	}
	q.failed.Add(1)
	slog.Error("kafka: task failed after retries", "type", task.Type, "max_retries", q.maxRetries, "error", lastErr)
}

// kafkaConsumerGroupHandler implements sarama.ConsumerGroupHandler
type kafkaConsumerGroupHandler struct {
	queue *KafkaQueue
}

func (h *kafkaConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *kafkaConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }
func (h *kafkaConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		h.queue.processMessage(msg)
		session.MarkMessage(msg, "")
	}
	return nil
}
