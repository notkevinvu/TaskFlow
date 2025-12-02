package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// In-Memory Rate Limiter Tests
// =============================================================================

func TestInMemoryRateLimiter_AllowsRequestsUnderLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	// Use nil Redis limiter to trigger in-memory fallback
	// High limit to ensure requests pass
	router.Use(RateLimiter(nil, 100))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Make a few requests - should all pass
	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}
}

func TestInMemoryRateLimiter_BlocksExcessiveRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	// Very low limit: 1 request per minute with burst of 10
	// The in-memory limiter uses rate.NewLimiter(requestsPerMinute/60, 10)
	// So with 1 request per minute, after burst of 10, it should block
	router.Use(RateLimiter(nil, 1))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	var blocked bool
	// Make many requests to exhaust the burst
	for i := 0; i < 20; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code == http.StatusTooManyRequests {
			blocked = true
			assert.Contains(t, w.Body.String(), "Rate limit exceeded")
			break
		}
	}

	assert.True(t, blocked, "Expected rate limiter to block excessive requests")
}

func TestInMemoryRateLimiter_SeparateLimitsPerClient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimiter(nil, 100))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ip": c.ClientIP()})
	})

	// Client 1
	req1, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req1.RemoteAddr = "10.0.0.1:12345"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Client 2 (different IP)
	req2, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req2.RemoteAddr = "10.0.0.2:12345"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestInMemoryRateLimiter_UsesUserIDWhenAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// First apply a middleware that sets user ID
	router.Use(func(c *gin.Context) {
		c.Set(UserIDKey, "user-123")
		c.Next()
	})
	router.Use(RateLimiter(nil, 100))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestInMemoryRateLimiter_Returns429WithCorrectMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	// Very restrictive limit
	router.Use(RateLimiter(nil, 1))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Exhaust the limit
	for i := 0; i < 50; i++ {
		req, _ := http.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.50.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code == http.StatusTooManyRequests {
			assert.Contains(t, w.Body.String(), "Rate limit exceeded")
			assert.Contains(t, w.Body.String(), "Please try again later")
			return
		}
	}

	t.Error("Expected to receive 429 status code")
}

// =============================================================================
// RateLimiter nil Redis fallback Tests
// =============================================================================

func TestRateLimiter_FallsBackToInMemoryWhenRedisNil(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	// Pass nil for Redis limiter - should fall back to in-memory
	router.Use(RateLimiter(nil, 60))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should work without Redis
	assert.Equal(t, http.StatusOK, w.Code)
}

// =============================================================================
// Integration-style Tests
// =============================================================================

func TestRateLimiter_MultipleEndpointsSameClient(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimiter(nil, 100))
	router.GET("/tasks", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"endpoint": "tasks"})
	})
	router.GET("/analytics", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"endpoint": "analytics"})
	})

	// Same client accessing different endpoints
	req1, _ := http.NewRequest(http.MethodGet, "/tasks", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	req2, _ := http.NewRequest(http.MethodGet, "/analytics", nil)
	req2.RemoteAddr = "192.168.1.1:12345"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestRateLimiter_DifferentHTTPMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RateLimiter(nil, 100))
	router.GET("/tasks", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"method": "GET"})
	})
	router.POST("/tasks", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"method": "POST"})
	})

	// GET request
	req1, _ := http.NewRequest(http.MethodGet, "/tasks", nil)
	req1.RemoteAddr = "10.1.1.1:12345"
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// POST request from same client
	req2, _ := http.NewRequest(http.MethodPost, "/tasks", nil)
	req2.RemoteAddr = "10.1.1.1:12345"
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}
