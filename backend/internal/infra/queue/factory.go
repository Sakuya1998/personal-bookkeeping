package queue

import (
	"fmt"

	"personal-bookkeeping/internal/infra/config"
)

// NewFromConfig 根据配置创建队列实例。
// queue.enabled=false 时返回 nil（调用方跳过初始化）。
func NewFromConfig(cfg *config.QueueConfig) (Queue, error) {
	if !cfg.Enabled {
		return nil, nil
	}
	switch cfg.Type {
	case "inmemory":
		return NewInMemory(cfg.Workers, cfg.MaxRetries), nil
	case "redis":
		return NewRedis(
			cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB,
			cfg.Redis.Stream, cfg.Redis.ConsumerGroup,
			cfg.Workers, cfg.MaxRetries,
		), nil
	case "kafka":
		return NewKafka(
			cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.GroupID,
			cfg.Workers, cfg.MaxRetries,
		)
	default:
		return nil, fmt.Errorf("unknown queue type: %q (expected inmemory, redis, or kafka)", cfg.Type)
	}
}
