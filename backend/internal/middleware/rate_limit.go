package middleware

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/notkevinvu/taskflow/backend/internal/ratelimit"
	"golang.org/x/time/rate"
)

// RateLimiter creates a rate limiting middleware
// If Redis limiter is provided, uses Redis for horizontal scalability
// If Redis is nil, falls back to in-memory rate limiting (single instance only)
func RateLimiter(redisLimiter *ratelimit.RedisLimiter, requestsPerMinute int) gin.HandlerFunc {
	// If Redis is not available, create in-memory fallback
	if redisLimiter == nil {
		return inMemoryRateLimiter(requestsPerMinute)
	}

	// Redis-backed rate limiting
	return func(c *gin.Context) {
		// Use user_id if authenticated, otherwise use IP
		identifier := c.ClientIP()
		if userID, exists := GetUserID(c); exists {
			identifier = userID
		}

		// Check rate limit with 1-minute window
		allowed, err := redisLimiter.Allow(c.Request.Context(), identifier, requestsPerMinute, time.Minute)
		if err != nil {
			// Log error but fail open (allow request) to prevent Redis outages from blocking all traffic
			slog.Warn("Rate limiter error, failing open", "error", err, "identifier", identifier)
			c.Next()
			return
		}

		// Get rate limit info for headers
		limitInfo, err := redisLimiter.GetLimitInfo(c.Request.Context(), identifier, requestsPerMinute, time.Minute)
		if err == nil {
			// Add standard rate limit headers
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", limitInfo.Limit))
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", limitInfo.Remaining))
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", limitInfo.ResetAt.Unix()))
		}

		if !allowed {
			// Add Retry-After header for rate limit exceeded
			if limitInfo != nil {
				retryAfter := int(time.Until(limitInfo.ResetAt).Seconds())
				if retryAfter < 0 {
					retryAfter = 1
				}
				c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			}

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// inMemoryRateLimiter provides in-memory rate limiting for single-instance deployments
func inMemoryRateLimiter(requestsPerMinute int) gin.HandlerFunc {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Cleanup old clients every 3 minutes
	go func() {
		for {
			time.Sleep(3 * time.Minute)
			mu.Lock()
			for id, c := range clients {
				if time.Since(c.lastSeen) > 3*time.Minute {
					delete(clients, id)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		// Use user_id if authenticated, otherwise use IP
		identifier := c.ClientIP()
		if userID, exists := GetUserID(c); exists {
			identifier = userID
		}

		mu.Lock()
		if _, exists := clients[identifier]; !exists {
			// Create new rate limiter: requestsPerMinute tokens, burst of 10
			clients[identifier] = &client{
				limiter: rate.NewLimiter(rate.Limit(requestsPerMinute)/60, 10),
			}
		}
		clients[identifier].lastSeen = time.Now()
		limiter := clients[identifier].limiter
		mu.Unlock()

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
