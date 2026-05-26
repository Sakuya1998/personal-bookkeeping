package task

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"personal-bookkeeping/internal/infra/queue"

	"github.com/google/uuid"
)

// StartRecurringScheduler 启动周期性交易调度器。
// 定期 dispatch TypeProcessRecurring 任务到队列。
// handler 本身是幂等的（按 next_run_date 筛选），多次调度不产生重复。
func StartRecurringScheduler(ctx context.Context, q queue.Queue, interval time.Duration) {
	if q == nil {
		slog.Warn("recurring scheduler: queue disabled, skipping")
		return
	}
	if interval <= 0 {
		interval = 1 * time.Hour
	}

	var (
		mu         sync.Mutex
		lastDate   string // YYYY-MM-DD，避免同一天多次调度
		firstRun   = true
	)

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		slog.Info("recurring scheduler started", "check_interval", interval)

		for {
			select {
			case <-ctx.Done():
				slog.Info("recurring scheduler stopped")
				return
			case <-ticker.C:
				today := time.Now().Format("2006-01-02")

				mu.Lock()
				if today == lastDate && !firstRun {
					mu.Unlock()
					continue // 同一日期已调度过
				}
				firstRun = false
				lastDate = today
				mu.Unlock()

				task := queue.Task{
					ID:      uuid.New().String(),
					Type:    TypeProcessRecurring,
					Payload: nil, // handler 无额外参数
				}
				if err := q.Submit(ctx, task); err != nil {
					slog.Error("recurring scheduler: submit failed", "error", err)
				} else {
					slog.Info("recurring scheduler: dispatched process_recurring task", "task_id", task.ID, "date", today)
				}
			}
		}
	}()
}

// StartExchangeRateScheduler 启动汇率自动更新调度器。
// 每天 UTC 02:00 dispatch TypeUpdateExchangeRates 任务。
func StartExchangeRateScheduler(ctx context.Context, q queue.Queue) {
	if q == nil {
		slog.Warn("exchange rate scheduler: queue disabled, skipping")
		return
	}

	go func() {
		now := time.Now().UTC()
		next := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, time.UTC)
		if !next.After(now) {
			next = next.AddDate(0, 0, 1)
		}
		firstDelay := next.Sub(now)

		slog.Info("exchange rate scheduler started", "first_run_in", firstDelay)

		select {
		case <-ctx.Done():
			slog.Info("exchange rate scheduler stopped")
			return
		case <-time.After(firstDelay):
			dispatchExchangeRateUpdate(ctx, q)
		}

		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				slog.Info("exchange rate scheduler stopped")
				return
			case <-ticker.C:
				dispatchExchangeRateUpdate(ctx, q)
			}
		}
	}()
}

func dispatchExchangeRateUpdate(ctx context.Context, q queue.Queue) {
	task := queue.Task{
		ID:      uuid.New().String(),
		Type:    TypeUpdateExchangeRates,
		Payload: nil,
	}
	if err := q.Submit(ctx, task); err != nil {
		slog.Error("exchange rate scheduler: submit failed", "error", err)
	} else {
		slog.Info("exchange rate scheduler: dispatched update_exchange_rates task", "task_id", task.ID)
	}
}
