// Package middleware 提供 HTTP 中间件，包括速率限制。
package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter 基于 IP 的滑动窗口速率限制器。
// 使用内存存储，适合单实例部署。
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*visitor
}

// visitor 记录单个 IP 的请求计数和窗口重置时间。
type visitor struct {
	count   int
	resetAt time.Time
}

// NewRateLimiter 创建速率限制器并启动后台清理协程。
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
	}
	go rl.cleanupLoop()
	return rl
}

// Limit 返回一个 Gin 中间件，限制每个 IP 在指定时间窗口内的最大请求数。
//
// 示例:
//
//	limiter := NewRateLimiter()
//	r.POST("/send-code", limiter.Limit(2, time.Minute), handler.SendCode)
func (rl *RateLimiter) Limit(maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		rl.mu.Lock()
		v, exists := rl.visitors[ip]
		now := time.Now()

		if !exists || now.After(v.resetAt) {
			// 新窗口
			rl.visitors[ip] = &visitor{count: 1, resetAt: now.Add(window)}
			rl.mu.Unlock()
			c.Next()
			return
		}

		v.count++
		remaining := maxRequests - v.count
		rl.mu.Unlock()

		if remaining < 0 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "请求过于频繁，请稍后再试",
			})
			return
		}

		c.Next()
	}
}

// cleanupLoop 定期清理过期的访问者记录，防止内存泄漏。
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, v := range rl.visitors {
			if now.After(v.resetAt) {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}
