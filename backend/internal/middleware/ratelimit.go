package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xxbbzy/gonext-template/backend/pkg/response"
	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

type KeyFunc func(*gin.Context) string

// RateLimiter implements per-IP rate limiting using token bucket algorithm.
type RateLimiter struct {
	visitors map[string]*visitor
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a rate limiter with the given requests per duration.
func NewRateLimiter(requests int, duration time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		rate:     rate.Limit(float64(requests) / duration.Seconds()),
		burst:    requests,
	}

	// Cleanup old visitors every minute
	go rl.cleanup()

	return rl
}

func (rl *RateLimiter) cleanup() {
	for {
		time.Sleep(time.Minute)
		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[ip] = &visitor{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

// Middleware returns a Gin middleware for rate limiting.
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return rl.MiddlewareWithKey(IPKey)
}

// MiddlewareWithKey returns a Gin middleware for rate limiting using the given key.
func (rl *RateLimiter) MiddlewareWithKey(keyFunc KeyFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)
		if key == "" {
			c.Next()
			return
		}

		limiter := rl.getVisitor(key)
		if !limiter.Allow() {
			c.Header("Retry-After", "60")
			response.Error(c, 429, 429, "too many requests")
			c.Abort()
			return
		}
		c.Next()
	}
}

// IPKey returns the limiter key for anonymous traffic.
func IPKey(c *gin.Context) string {
	return "ip:" + c.ClientIP()
}

// UserKey returns the limiter key for authenticated traffic.
func UserKey(c *gin.Context) string {
	userID, exists := c.Get("user_id")
	if !exists {
		return ""
	}

	switch id := userID.(type) {
	case uint:
		return fmt.Sprintf("user:%d", id)
	case int:
		return fmt.Sprintf("user:%d", id)
	case int64:
		return fmt.Sprintf("user:%d", id)
	case uint64:
		return fmt.Sprintf("user:%d", id)
	default:
		return ""
	}
}
