package otel

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"

	"github.com/gin-gonic/gin"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	otelgin "go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

// Config OTEL 配置。
type Config struct {
	Enabled        bool
	ServiceName    string
	TracesExporter string // "none" | "stdout"
	MetricsPath    string
}

// OTEL 持有所有 OTEL 组件，关闭时需统一清理。
type OTEL struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *metric.MeterProvider
	Meter          api.Meter
	MetricsHandler http.Handler // /metrics 端点

	// 内置的 HTTP 指标
	reqCounter  api.Int64Counter
	durHist     api.Float64Histogram
	activeReq   api.Int64UpDownCounter

	shutdownFuncs []func(context.Context) error
}

// Init 初始化 OTEL SDK。可以多次调用 Init(nil) 返回空实现。
func Init(cfg *Config) (*OTEL, error) {
	if cfg == nil || !cfg.Enabled {
		return &OTEL{}, nil
	}

	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("resource: %w", err)
	}

	o := &OTEL{}

	// — Tracing —
	if cfg.TracesExporter == "stdout" {
		texp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("trace exporter: %w", err)
		}
		tp := sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(texp),
			sdktrace.WithResource(res),
		)
		o.TracerProvider = tp
		o.shutdownFuncs = append(o.shutdownFuncs, tp.Shutdown)
		otel.SetTracerProvider(tp)
	}

	// — 全局传播器：支持 trace context 在 HTTP 间传递 —
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// — Metrics（Prometheus）—
	reg := prom.NewRegistry()
	mexp, err := prometheus.New(
		prometheus.WithRegisterer(reg),
	)
	if err != nil {
		return nil, fmt.Errorf("prometheus exporter: %w", err)
	}
	mp := metric.NewMeterProvider(
		metric.WithReader(mexp),
		metric.WithResource(res),
	)
	o.MeterProvider = mp
	o.Meter = mp.Meter(cfg.ServiceName)
	o.MetricsHandler = promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	o.shutdownFuncs = append(o.shutdownFuncs, mp.Shutdown)

	// — 内置 HTTP 指标 —
	o.reqCounter, _ = o.Meter.Int64Counter(
		"http.server.request_count",
		api.WithDescription("HTTP request count"),
		api.WithUnit("{request}"),
	)
	o.durHist, _ = o.Meter.Float64Histogram(
		"http.server.request_duration_seconds",
		api.WithDescription("HTTP request duration"),
		api.WithUnit("s"),
	)
	o.activeReq, _ = o.Meter.Int64UpDownCounter(
		"http.server.active_requests",
		api.WithDescription("Active HTTP requests"),
		api.WithUnit("{request}"),
	)

	return o, nil
}

// Shutdown 清理所有 OTEL 资源。
func (o *OTEL) Shutdown() {
	if o == nil {
		return
	}
	for _, fn := range o.shutdownFuncs {
		if err := fn(context.Background()); err != nil {
			slog.Warn("otel shutdown error", "error", err)
		}
	}
}

// GinMiddleware 返回 Gin 中间件：trace 传播 + HTTP 指标。
func (o *OTEL) GinMiddleware(serviceName string) gin.HandlerFunc {
	if o.Meter == nil {
		return func(c *gin.Context) { c.Next() } // noop
	}

	// trace 中间件
	traceMw := otelgin.Middleware(serviceName)

	return func(c *gin.Context) {
		// 先执行 trace 中间件（注入 span 到 context）
		traceMw(c)
		if c.IsAborted() {
			return
		}

		// 指标收集
		start := time.Now()
		o.activeReq.Add(c.Request.Context(), 1)
		c.Next()
		o.activeReq.Add(c.Request.Context(), -1)

		status := c.Writer.Status()
		dur := time.Since(start).Seconds()
		attrs := []attribute.KeyValue{
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.route", c.FullPath()),
			attribute.Int("http.status_code", status),
		}

		o.reqCounter.Add(c.Request.Context(), 1, api.WithAttributes(attrs...))
		o.durHist.Record(c.Request.Context(), dur, api.WithAttributes(attrs...))
	}
}
