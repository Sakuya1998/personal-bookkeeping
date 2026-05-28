package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter 基于 IP 的滑动窗口限流。
// 零值可用：NewRateLimiter(rate, window) 创建。
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string][]time.Time
	rate     int           // 窗口内允许的最大请求数
	window   time.Duration // 滑动窗口时长
}

func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string][]time.Time),
		rate:     rate,
		window:   window,
	}
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		cutoff := time.Now().Add(-rl.window)
		for ip, times := range rl.visitors {
			j := 0
			for j < len(times) && times[j].Before(cutoff) {
				j++
			}
			if j == len(times) {
				delete(rl.visitors, ip)
			} else {
				rl.visitors[ip] = times[j:]
			}
		}
		rl.mu.Unlock()
	}
}

// Allow 检查 IP 是否在限流窗口内。
func (rl *RateLimiter) Allow(ip string) bool {
	now := time.Now()
	cutoff := now.Add(-rl.window)

	rl.mu.Lock()
	defer rl.mu.Unlock()

	times := rl.visitors[ip]
	j := 0
	for j < len(times) && times[j].Before(cutoff) {
		j++
	}
	times = times[j:]

	if len(times) >= rl.rate {
		rl.visitors[ip] = times
		return false
	}

	rl.visitors[ip] = append(times, now)
	return true
}

// LimitRate 返回 Gin 中间件，对匹配 pathPrefix 的路由进行限流。
func LimitRate(rl *RateLimiter, pathPrefix string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path != pathPrefix {
			c.Next()
			return
		}
		ip := c.ClientIP()
		if !rl.Allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "too many requests, please try later",
			})
			return
		}
		c.Next()
	}
}
