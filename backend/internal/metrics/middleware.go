package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Middleware returns a Gin middleware that records HTTP metrics
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Track in-flight requests
		HTTPRequestsInFlight.Inc()
		defer HTTPRequestsInFlight.Dec()

		// Record start time
		start := time.Now()

		// Get path template for consistent labeling
		// Use the route pattern if available, otherwise use a constant
		// to prevent cardinality explosion from unmatched/404 paths
		path := c.FullPath()
		if path == "" {
			path = "/unmatched"
		}

		// Process request
		c.Next()

		// Record metrics after request completes
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method

		// Record request count
		HTTPRequestsTotal.WithLabelValues(method, path, status).Inc()

		// Record request duration
		HTTPRequestDuration.WithLabelValues(method, path).Observe(duration)
	}
}
