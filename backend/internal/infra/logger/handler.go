package logger

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"sync"

	"go.opentelemetry.io/otel/trace"
	"gopkg.in/natefinch/lumberjack.v2"
)

// LevelFileConfig 日志配置。
type LevelFileConfig struct {
	Target     string // "file" | "stdout" | "stderr" | "both"
	Dir        string // 日志目录（仅 file / both）
	Info       string // Info 文件名
	Warn       string // Warn 文件名
	Error      string // Error 文件名
	MaxSize    int    // 单个文件最大 MB（仅 file / both）
	MaxAge     int    // 保留天数
	MaxBackups int    // 保留旧文件数（0 = 按天数）
	Compress   bool   // 是否 gzip 压缩旧文件
}

// LevelFileHandler 按日志级别将 JSON 日志写入不同 Writer。
type LevelFileHandler struct {
	mu       sync.Mutex
	minLevel slog.Leveler
	infoH    slog.Handler
	warnH    slog.Handler
	errorH   slog.Handler
	closers  []io.Closer // 退出时关闭的文件句柄
}

// NewHandler 根据配置创建 slog.Handler。
func NewHandler(cfg LevelFileConfig, opts *slog.HandlerOptions) (slog.Handler, error) {
	if opts == nil {
		opts = &slog.HandlerOptions{}
	}

	switch cfg.Target {
	case "stdout":
		return slog.NewJSONHandler(os.Stdout, opts), nil
	case "stderr":
		return slog.NewJSONHandler(os.Stderr, opts), nil
	case "both":
		return newMultiHandler(cfg, opts)
	default: // "file"
		return newLevelFileHandler(cfg, opts)
	}
}

// newLevelFileHandler 创建按级别分文件的 handler（带轮转）。
func newLevelFileHandler(cfg LevelFileConfig, opts *slog.HandlerOptions) (*LevelFileHandler, error) {
	if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
		return nil, err
	}

	jsonOpts := &slog.HandlerOptions{
		Level:       slog.LevelDebug,
		AddSource:   opts.AddSource,
		ReplaceAttr: opts.ReplaceAttr,
	}

	h := &LevelFileHandler{
		minLevel: slog.LevelDebug,
		closers:  make([]io.Closer, 0, 3),
	}

	if opts.Level != nil {
		h.minLevel = opts.Level.Level()
	}

	infoW := newLumberjack(cfg.Dir, cfg.Info, cfg)
	warnW := newLumberjack(cfg.Dir, cfg.Warn, cfg)
	errW := newLumberjack(cfg.Dir, cfg.Error, cfg)

	h.closers = append(h.closers, infoW, warnW, errW)
	h.infoH = slog.NewJSONHandler(infoW, jsonOpts)
	h.warnH = slog.NewJSONHandler(warnW, jsonOpts)
	h.errorH = slog.NewJSONHandler(errW, jsonOpts)

	return h, nil
}

// newMultiHandler 创建同时写文件和 stdout/stderr 的 handler。
func newMultiHandler(cfg LevelFileConfig, opts *slog.HandlerOptions) (*LevelFileHandler, error) {
	if err := os.MkdirAll(cfg.Dir, 0755); err != nil {
		return nil, err
	}

	jsonOpts := &slog.HandlerOptions{
		Level:       slog.LevelDebug,
		AddSource:   opts.AddSource,
		ReplaceAttr: opts.ReplaceAttr,
	}

	h := &LevelFileHandler{
		minLevel: slog.LevelDebug,
		closers:  make([]io.Closer, 0, 3),
	}
	if opts.Level != nil {
		h.minLevel = opts.Level.Level()
	}

	infoF := newLumberjack(cfg.Dir, cfg.Info, cfg)
	warnF := newLumberjack(cfg.Dir, cfg.Warn, cfg)
	errF := newLumberjack(cfg.Dir, cfg.Error, cfg)

	h.closers = append(h.closers, infoF, warnF, errF)

	h.infoH = slog.NewJSONHandler(io.MultiWriter(infoF, os.Stdout), jsonOpts)
	h.warnH = slog.NewJSONHandler(io.MultiWriter(warnF, os.Stdout), jsonOpts)
	h.errorH = slog.NewJSONHandler(io.MultiWriter(errF, os.Stderr), jsonOpts)

	return h, nil
}

func (h *LevelFileHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.minLevel.Level()
}

func (h *LevelFileHandler) Handle(ctx context.Context, r slog.Record) error {
	// 从 context 提取 trace_id / span_id（如果有）
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		r.AddAttrs(
			slog.String("trace_id", sc.TraceID().String()),
			slog.String("span_id", sc.SpanID().String()),
		)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	var errs []error
	switch {
	case r.Level >= slog.LevelError:
		errs = append(errs, h.errorH.Handle(ctx, r))
	case r.Level >= slog.LevelWarn:
		errs = append(errs, h.warnH.Handle(ctx, r))
	default:
		errs = append(errs, h.infoH.Handle(ctx, r))
	}
	return errors.Join(errs...)
}

func (h *LevelFileHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &LevelFileHandler{
		mu:       sync.Mutex{},
		minLevel: h.minLevel,
		infoH:    h.infoH.WithAttrs(attrs),
		warnH:    h.warnH.WithAttrs(attrs),
		errorH:   h.errorH.WithAttrs(attrs),
	}
}

func (h *LevelFileHandler) WithGroup(name string) slog.Handler {
	return &LevelFileHandler{
		mu:       sync.Mutex{},
		minLevel: h.minLevel,
		infoH:    h.infoH.WithGroup(name),
		warnH:    h.warnH.WithGroup(name),
		errorH:   h.errorH.WithGroup(name),
	}
}

// Close 关闭所有日志文件。
func (h *LevelFileHandler) Close() error {
	var errs []error
	for _, c := range h.closers {
		errs = append(errs, c.Close())
	}
	return errors.Join(errs...)
}

func newLumberjack(dir, name string, cfg LevelFileConfig) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   dir + "/" + name,
		MaxSize:    cfg.MaxSize,
		MaxAge:     cfg.MaxAge,
		MaxBackups: cfg.MaxBackups,
		Compress:   cfg.Compress,
		LocalTime:  true,
	}
}
