package testutil

import (
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/notkevinvu/taskflow/backend/internal/middleware"
)

// SetupTestRouter creates a test Gin router with error handling middleware
func SetupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	// Add error handler middleware to properly map errors to HTTP status codes
	router.Use(middleware.ErrorHandler())
	return router
}

// NewResponseRecorder creates a new httptest.ResponseRecorder for each test
func NewResponseRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

// WithAuthContext creates a test context with user authentication set
func WithAuthContext(router *gin.Engine, userID string, handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(middleware.UserIDKey, userID)
		handler(c)
	}
}
