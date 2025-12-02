package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// ErrorHandler Middleware Tests
// =============================================================================

func TestErrorHandler_NoError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

func TestErrorHandler_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		c.Error(&domain.ValidationError{
			Field:   "email",
			Message: "Invalid email format",
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid email format")
	assert.Contains(t, w.Body.String(), "email")
}

func TestErrorHandler_NotFoundError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		c.Error(&domain.NotFoundError{
			Resource: "task",
			ID:       "task-123",
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "task")
}

func TestErrorHandler_ConflictError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		c.Error(&domain.ConflictError{
			Resource: "user",
			Message:  "Email already exists",
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "Email already exists")
	assert.Contains(t, w.Body.String(), "user")
}

func TestErrorHandler_UnauthorizedError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		c.Error(&domain.UnauthorizedError{
			Message: "Invalid credentials",
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid credentials")
}

func TestErrorHandler_ForbiddenError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		c.Error(&domain.ForbiddenError{
			Resource: "task",
			Action:   "delete",
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "forbidden")
}

func TestErrorHandler_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		c.Error(&domain.InternalError{
			Message: "Database connection failed",
			Cause:   errors.New("connection timeout"),
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	// Should NOT expose internal error details
	assert.NotContains(t, w.Body.String(), "Database connection failed")
	assert.Contains(t, w.Body.String(), "internal error occurred")
}

func TestErrorHandler_UnknownError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		c.Error(errors.New("some unknown error"))
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "unexpected error occurred")
}

func TestErrorHandler_MultipleErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		// Add multiple errors - handler uses the last one
		c.Error(errors.New("first error"))
		c.Error(&domain.ValidationError{
			Field:   "name",
			Message: "Name is required",
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should use the last error (ValidationError)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Name is required")
}

// =============================================================================
// AbortWithError Helper Tests
// =============================================================================

func TestAbortWithError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		AbortWithError(c, &domain.ValidationError{
			Field:   "title",
			Message: "Title is required",
		})
		// In real code, you'd return after AbortWithError
		// Abort() prevents further middleware, not current handler code
		return
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Title is required")
}

func TestAbortWithError_PreventsSubsequentMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ErrorHandler())

	subsequentMiddlewareRan := false

	// First middleware calls abort
	router.Use(func(c *gin.Context) {
		AbortWithError(c, &domain.ValidationError{
			Field:   "auth",
			Message: "Not authorized",
		})
	})

	// Second middleware should NOT run
	router.Use(func(c *gin.Context) {
		subsequentMiddlewareRan = true
		c.Next()
	})

	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.False(t, subsequentMiddlewareRan, "Subsequent middleware should not run after abort")
}

// =============================================================================
// mapErrorToResponse Tests (via ErrorHandler)
// =============================================================================

func TestMapErrorToResponse_ValidationErrorWithDetails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/test", func(c *gin.Context) {
		c.Error(&domain.ValidationError{
			Field:   "password",
			Message: "Password must be at least 8 characters",
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "password")
	assert.Contains(t, w.Body.String(), "Password must be at least 8 characters")
}

func TestMapErrorToResponse_NotFoundErrorFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ErrorHandler())
	router.GET("/tasks/:id", func(c *gin.Context) {
		c.Error(&domain.NotFoundError{
			Resource: "task",
			ID:       c.Param("id"),
		})
	})

	req, _ := http.NewRequest(http.MethodGet, "/tasks/nonexistent-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestMapErrorToResponse_ConflictErrorWithResource(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ErrorHandler())
	router.POST("/test", func(c *gin.Context) {
		c.Error(&domain.ConflictError{
			Resource: "email",
			Message:  "This email is already registered",
		})
	})

	req, _ := http.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "This email is already registered")
	// Check for resource in details
	assert.Contains(t, w.Body.String(), "email")
}
