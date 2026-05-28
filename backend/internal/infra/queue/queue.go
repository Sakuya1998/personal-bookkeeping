package queue

import (
	"context"
	"errors"
)

// HandlerFunc 处理单个任务的函数签名
type HandlerFunc func(ctx context.Context, task Task) error

// Task 队列任务
type Task struct {
	ID      string
	Type    string
	Payload any
}

// Queue 任务队列接口
type Queue interface {
	// Register 注册任务类型对应的处理函数
	Register(taskType string, fn HandlerFunc)

	// Submit 提交任务
	Submit(ctx context.Context, task Task) error

	// Start 启动 worker
	Start(ctx context.Context)

	// Shutdown 优雅关闭，等待进行中的任务完成
	Shutdown(ctx context.Context) error

	// Stats 返回队列统计信息
	Stats() Stats
}

// Stats 队列统计
type Stats struct {
	Pending   int
	Running   int
	Completed int
	Failed    int
}

// ErrQueueStopped 队列已停止
var ErrQueueStopped = errors.New("queue: stopped")

// defaultQueue is the package-level default queue instance.
var defaultQueue Queue

// SetDefault sets the package-level default queue instance.
func SetDefault(q Queue) { defaultQueue = q }

// GetDefault returns the package-level default queue instance.
func GetDefault() Queue { return defaultQueue }
