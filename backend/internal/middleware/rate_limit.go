package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/notkevinvu/taskflow/backend/internal/ratelimit"
	"golang.org/x/time/rate"
)

// InMemoryRateLimiterConfig holds the cleanup resources for graceful shutdown
type InMemoryRateLimiterConfig struct {
	stopCleanup chan struct{}
	stopped     bool
	mu          sync.Mutex
}

// Stop gracefully stops the cleanup goroutine
func (c *InMemoryRateLimiterConfig) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.stopped {
		close(c.stopCleanup)
		c.stopped = true
		slog.Info("In-memory rate limiter cleanup stopped")
	}
}

// RateLimiter creates a rate limiting middleware
// If Redis limiter is provided, uses Redis for horizontal scalability
// If Redis is nil, falls back to in-memory rate limiting (single instance only)
func RateLimiter(redisLimiter *ratelimit.RedisLimiter, requestsPerMinute int) gin.HandlerFunc {
	// If Redis is not available, create in-memory fallback
	if redisLimiter == nil {
		_, handler := inMemoryRateLimiter(context.Background(), requestsPerMinute)
		return handler
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

// RateLimiterWithContext creates a rate limiting middleware with context-aware cleanup
// Returns the config for stopping the cleanup goroutine on shutdown
func RateLimiterWithContext(ctx context.Context, redisLimiter *ratelimit.RedisLimiter, requestsPerMinute int) (*InMemoryRateLimiterConfig, gin.HandlerFunc) {
	// If Redis is available, no cleanup goroutine needed
	if redisLimiter != nil {
		return nil, RateLimiter(redisLimiter, requestsPerMinute)
	}

	// Create in-memory rate limiter with context-aware cleanup
	return inMemoryRateLimiter(ctx, requestsPerMinute)
}

// inMemoryRateLimiter provides in-memory rate limiting for single-instance deployments
// Returns the config for cleanup control and the middleware handler
func inMemoryRateLimiter(ctx context.Context, requestsPerMinute int) (*InMemoryRateLimiterConfig, gin.HandlerFunc) {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	config := &InMemoryRateLimiterConfig{
		stopCleanup: make(chan struct{}),
	}

	// Cleanup old clients every 3 minutes with proper cancellation
	go func() {
		ticker := time.NewTicker(3 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				slog.Info("Rate limiter cleanup stopped via context cancellation")
				return
			case <-config.stopCleanup:
				slog.Info("Rate limiter cleanup stopped via stop signal")
				return
			case <-ticker.C:
				mu.Lock()
				cleanedCount := 0
				for id, c := range clients {
					if time.Since(c.lastSeen) > 3*time.Minute {
						delete(clients, id)
						cleanedCount++
					}
				}
				if cleanedCount > 0 {
					slog.Debug("Rate limiter cleaned up stale clients", "count", cleanedCount)
				}
				mu.Unlock()
			}
		}
	}()

	handler := func(c *gin.Context) {
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

	return config, handler
}
