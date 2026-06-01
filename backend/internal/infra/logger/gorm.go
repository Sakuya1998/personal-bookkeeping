package logger

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

// SlogLogger adapts slog to GORM's logger interface.
type SlogLogger struct {
	logLevel          logger.LogLevel
	ignoreRecordNotFoundError bool
	slowThreshold             time.Duration
}

func NewSlogLogger(level logger.LogLevel) *SlogLogger {
	return &SlogLogger{
		logLevel:                   level,
		ignoreRecordNotFoundError:  true,
		slowThreshold:              200 * time.Millisecond,
	}
}

func (l *SlogLogger) LogMode(level logger.LogLevel) logger.Interface {
	l.logLevel = level
	return l
}

func (l *SlogLogger) Info(_ context.Context, msg string, args ...interface{}) {
	if l.logLevel >= logger.Info {
		slog.Info(fmt.Sprintf(msg, args...))
	}
}

func (l *SlogLogger) Warn(_ context.Context, msg string, args ...interface{}) {
	if l.logLevel >= logger.Warn {
		slog.Warn(fmt.Sprintf(msg, args...))
	}
}

func (l *SlogLogger) Error(_ context.Context, msg string, args ...interface{}) {
	if l.logLevel >= logger.Error {
		slog.Error(fmt.Sprintf(msg, args...))
	}
}

func (l *SlogLogger) Trace(_ context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.logLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	// Skip logging for ErrRecordNotFound if configured
	if err != nil && l.ignoreRecordNotFoundError && logger.ErrRecordNotFound.Error() == err.Error() {
		err = nil
	}

	attrs := []slog.Attr{
		slog.String("file", utils.FileWithLineNum()),
		slog.String("sql", sql),
		slog.Duration("elapsed", elapsed),
		slog.Int64("rows", rows),
	}

	switch {
	case err != nil && l.logLevel >= logger.Error:
		attrs = append(attrs, slog.Any("error", err))
		slog.LogAttrs(nil, slog.LevelError, "SQL query failed", attrs...)
	case elapsed > l.slowThreshold && l.logLevel >= logger.Warn:
		attrs = append(attrs, slog.Duration("threshold", l.slowThreshold))
		slog.LogAttrs(nil, slog.LevelWarn, "Slow SQL query", attrs...)
	case l.logLevel >= logger.Info:
		slog.LogAttrs(nil, slog.LevelDebug, "SQL query", attrs...)
	}
}
