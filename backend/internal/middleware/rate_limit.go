package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter creates a rate limiting middleware
// Limits requests per minute based on user_id (after auth) or IP (before auth)
func RateLimiter(requestsPerMinute int) gin.HandlerFunc {
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
		if userID, exists := c.Get("user_id"); exists {
			identifier = userID.(string)
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
