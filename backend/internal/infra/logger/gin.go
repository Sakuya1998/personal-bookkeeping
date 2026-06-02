package logger

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// GinSlogMiddleware 用 slog 记录 Gin 的请求日志，JSON 格式，按级别分文件。
func GinSlogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		errors := c.Errors.ByType(gin.ErrorTypeAny).String()

		attrs := []slog.Attr{
			slog.Int("status", status),
			slog.String("method", method),
			slog.String("path", path),
			slog.String("ip", clientIP),
			slog.Duration("latency", latency),
			slog.Int("size", c.Writer.Size()),
		}
		if errors != "" {
			attrs = append(attrs, slog.String("error", errors))
		}

		if status >= 500 {
			slog.LogAttrs(c.Request.Context(), slog.LevelError, "request", attrs...)
		} else if status >= 400 {
			slog.LogAttrs(c.Request.Context(), slog.LevelWarn, "request", attrs...)
		} else {
			slog.LogAttrs(c.Request.Context(), slog.LevelInfo, "request", attrs...)
		}
	}
}
