package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter is a simple in-memory fixed-window limiter keyed by client.
type RateLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	buckets map[string]*bucket
}

type bucket struct {
	count int
	reset time.Time
}

func NewRateLimiter(perHour int) *RateLimiter {
	return &RateLimiter{
		limit:   perHour,
		window:  time.Hour,
		buckets: make(map[string]*bucket),
	}
}

// Start begins the periodic cleanup of expired buckets.
func (rl *RateLimiter) Start() {
	go func() {
		ticker := time.NewTicker(rl.window)
		defer ticker.Stop()
		for range ticker.C {
			rl.prune()
		}
	}()
}

func (rl *RateLimiter) prune() {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()
	for k, b := range rl.buckets {
		if now.After(b.reset) {
			delete(rl.buckets, k)
		}
	}
}

// Limit returns a Gin middleware enforcing the per-hour limit.
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := clientKey(c)

		rl.mu.Lock()
		b, ok := rl.buckets[key]
		now := time.Now()
		if !ok || now.After(b.reset) {
			b = &bucket{count: 0, reset: now.Add(rl.window)}
			rl.buckets[key] = b
		}
		b.count++
		count := b.count
		remaining := rl.limit - count
		if remaining < 0 {
			remaining = 0
		}
		reset := b.reset.Unix()
		rl.mu.Unlock()

		c.Header("X-RateLimit-Limit", strconv.Itoa(rl.limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(reset, 10))

		if count > rl.limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "RATE_LIMIT",
					"message": "rate limit exceeded",
				},
			})
			return
		}
		c.Next()
	}
}

func clientKey(c *gin.Context) string {
	if k, ok := c.Get("api_key"); ok {
		if s, ok := k.(string); ok && s != "" {
			return s
		}
	}
	// Fall back to client IP when no authenticated API key is present.
	return c.ClientIP()
}
