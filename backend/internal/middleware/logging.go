package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger middleware logs HTTP requests with structured logging
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate request duration
		duration := time.Since(start)

		// Get status code and response size
		statusCode := c.Writer.Status()
		responseSize := c.Writer.Size()

		// Determine log level based on status code
		logLevel := slog.LevelInfo
		if statusCode >= 500 {
			logLevel = slog.LevelError
		} else if statusCode >= 400 {
			logLevel = slog.LevelWarn
		}

		// Build log attributes
		attrs := []any{
			"method", c.Request.Method,
			"path", path,
			"status", statusCode,
			"duration_ms", duration.Milliseconds(),
			"ip", c.ClientIP(),
		}

		// Add query string if present
		if query != "" {
			attrs = append(attrs, "query", query)
		}

		// Add response size if available
		if responseSize > 0 {
			attrs = append(attrs, "response_size", responseSize)
		}

		// Add user ID if authenticated
		if userID, exists := GetUserID(c); exists {
			attrs = append(attrs, "user_id", userID)
		}

		// Add error message if present
		if len(c.Errors) > 0 {
			attrs = append(attrs, "error", c.Errors.Last().Error())
		}

		// Log the request
		slog.Log(c.Request.Context(), logLevel, "HTTP request", attrs...)
	}
}
