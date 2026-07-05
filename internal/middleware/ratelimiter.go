package middleware

import (
	"net/http"
	"sync"
	"time"

	"devhelp/internal/services/dto"

	"github.com/gin-gonic/gin"
)

// bucket holds the token count and last refill timestamp for a single IP address.
type bucket struct {
	mu       sync.Mutex
	tokens   float64
	lastSeen time.Time
}

// RateLimiter returns a Gin middleware that enforces IP-based rate limiting using
// a token-bucket algorithm.
// maxRPM is the maximum number of requests allowed per minute per IP address.
// Clients that exceed this limit receive HTTP 429 with a standard error envelope.
func RateLimiter(maxRPM int) gin.HandlerFunc {
	var (
		buckets sync.Map
		rate    = float64(maxRPM) / 60.0 // tokens per second
		max     = float64(maxRPM)
	)

	// Background goroutine that periodically cleans up stale buckets (inactive > 5 min).
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		for range ticker.C {
			now := time.Now()
			buckets.Range(func(k, v interface{}) bool {
				b := v.(*bucket)
				b.mu.Lock()
				idle := now.Sub(b.lastSeen)
				b.mu.Unlock()
				if idle > 5*time.Minute {
					buckets.Delete(k)
				}
				return true
			})
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()

		val, _ := buckets.LoadOrStore(ip, &bucket{
			tokens:   max,
			lastSeen: time.Now(),
		})
		b := val.(*bucket)

		b.mu.Lock()
		now := time.Now()
		elapsed := now.Sub(b.lastSeen).Seconds()
		b.tokens += elapsed * rate
		if b.tokens > max {
			b.tokens = max
		}
		b.lastSeen = now

		if b.tokens < 1 {
			b.mu.Unlock()
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, dto.APIResponse{
				Success: false,
				Error: &dto.APIError{
					Code:    "RATE_LIMITED",
					Message: "Rate limit exceeded. Try again later.",
					Details: "100 requests per minute allowed per IP",
				},
			})
			return
		}

		b.tokens--
		b.mu.Unlock()

		c.Next()
	}
}
