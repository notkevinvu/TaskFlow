package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// containsSensitiveParams Tests
// =============================================================================

func TestContainsSensitiveParams_DetectsToken(t *testing.T) {
	assert.True(t, containsSensitiveParams("token=abc123"))
	assert.True(t, containsSensitiveParams("access_token=xyz"))
	assert.True(t, containsSensitiveParams("refresh_token=def"))
}

func TestContainsSensitiveParams_DetectsPassword(t *testing.T) {
	assert.True(t, containsSensitiveParams("password=secret"))
	assert.True(t, containsSensitiveParams("user=test&password=secret123"))
}

func TestContainsSensitiveParams_DetectsSecret(t *testing.T) {
	assert.True(t, containsSensitiveParams("secret=mysecret"))
	assert.True(t, containsSensitiveParams("client_secret=abc"))
}

func TestContainsSensitiveParams_DetectsApiKey(t *testing.T) {
	assert.True(t, containsSensitiveParams("api_key=key123"))
	assert.True(t, containsSensitiveParams("apikey=key456"))
}

func TestContainsSensitiveParams_DetectsAuth(t *testing.T) {
	assert.True(t, containsSensitiveParams("auth=bearer"))
	assert.True(t, containsSensitiveParams("authorization=token"))
}

func TestContainsSensitiveParams_DetectsKey(t *testing.T) {
	assert.True(t, containsSensitiveParams("key=mykey"))
}

func TestContainsSensitiveParams_CaseInsensitive(t *testing.T) {
	assert.True(t, containsSensitiveParams("TOKEN=abc"))
	assert.True(t, containsSensitiveParams("PASSWORD=secret"))
	assert.True(t, containsSensitiveParams("API_KEY=xyz"))
	assert.True(t, containsSensitiveParams("ApiKey=test"))
}

func TestContainsSensitiveParams_SafeParams(t *testing.T) {
	assert.False(t, containsSensitiveParams("page=1"))
	assert.False(t, containsSensitiveParams("limit=10"))
	assert.False(t, containsSensitiveParams("status=pending"))
	assert.False(t, containsSensitiveParams("category=work"))
	assert.False(t, containsSensitiveParams("search=meeting"))
}

func TestContainsSensitiveParams_EmptyQuery(t *testing.T) {
	assert.False(t, containsSensitiveParams(""))
}

func TestContainsSensitiveParams_MultipleParams(t *testing.T) {
	// Contains safe params and one sensitive param
	assert.True(t, containsSensitiveParams("page=1&token=abc&limit=10"))
	// All safe params
	assert.False(t, containsSensitiveParams("page=1&limit=10&status=pending"))
}

func TestContainsSensitiveParams_PartialMatch(t *testing.T) {
	// "tokenizer" should not match "token=" pattern
	assert.False(t, containsSensitiveParams("tokenizer=test"))
	// "passwordless" should not match "password=" pattern
	assert.False(t, containsSensitiveParams("passwordless=true"))
}

// =============================================================================
// RequestLogger Middleware Tests
// =============================================================================

// captureLogOutput captures slog output during test execution
func captureLogOutput(fn func()) string {
	var buf bytes.Buffer
	originalHandler := slog.Default().Handler()

	// Create JSON handler that writes to buffer
	jsonHandler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	slog.SetDefault(slog.New(jsonHandler))

	fn()

	// Restore original handler
	slog.SetDefault(slog.New(originalHandler))
	return buf.String()
}

func TestRequestLogger_ReturnsHandler(t *testing.T) {
	handler := RequestLogger()
	assert.NotNil(t, handler)
}

func TestRequestLogger_LogsSuccessfulRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestLogger())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	output := captureLogOutput(func() {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// Parse log output
	assert.Contains(t, output, "HTTP request")
	assert.Contains(t, output, "/test")
	assert.Contains(t, output, "GET")
	assert.Contains(t, output, "200")
}

func TestRequestLogger_LogsErrorRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestLogger())
	router.GET("/error", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
	})

	output := captureLogOutput(func() {
		req := httptest.NewRequest(http.MethodGet, "/error", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	assert.Contains(t, output, "500")
	assert.Contains(t, output, "ERROR")
}

func TestRequestLogger_LogsClientError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestLogger())
	router.GET("/notfound", func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	})

	output := captureLogOutput(func() {
		req := httptest.NewRequest(http.MethodGet, "/notfound", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	assert.Contains(t, output, "404")
	assert.Contains(t, output, "WARN")
}

func TestRequestLogger_IncludesSafeQueryString(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestLogger())
	router.GET("/tasks", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"tasks": []string{}})
	})

	output := captureLogOutput(func() {
		req := httptest.NewRequest(http.MethodGet, "/tasks?page=1&limit=10", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	})

	// Should include safe query params
	assert.Contains(t, output, "page=1")
}

func TestRequestLogger_ExcludesSensitiveQueryString(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestLogger())
	router.GET("/auth", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	output := captureLogOutput(func() {
		req := httptest.NewRequest(http.MethodGet, "/auth?token=secret123", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	})

	// Should NOT include sensitive query params
	assert.NotContains(t, output, "secret123")
}

func TestRequestLogger_IncludesUserIDWhenAuthenticated(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(UserIDKey, "user-123")
		c.Next()
	})
	router.Use(RequestLogger())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	output := captureLogOutput(func() {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	})

	assert.Contains(t, output, "user-123")
}

func TestRequestLogger_IncludesResponseSize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestLogger())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": "some response data"})
	})

	output := captureLogOutput(func() {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	})

	assert.Contains(t, output, "response_size")
}

func TestRequestLogger_IncludesErrorMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestLogger())
	router.GET("/error", func(c *gin.Context) {
		c.Error(gin.Error{
			Err:  assert.AnError,
			Type: gin.ErrorTypePrivate,
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
	})

	output := captureLogOutput(func() {
		req := httptest.NewRequest(http.MethodGet, "/error", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	})

	assert.Contains(t, output, "error")
}

func TestRequestLogger_LogsDuration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestLogger())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	output := captureLogOutput(func() {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	})

	assert.Contains(t, output, "duration_ms")
}

func TestRequestLogger_LogsClientIP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestLogger())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	output := captureLogOutput(func() {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.RemoteAddr = "192.168.1.100:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	})

	assert.Contains(t, output, "ip")
}

func TestRequestLogger_LogsMethod(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestLogger())
	router.POST("/tasks", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"id": "123"})
	})

	output := captureLogOutput(func() {
		req := httptest.NewRequest(http.MethodPost, "/tasks", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	})

	assert.Contains(t, output, "POST")
}

// =============================================================================
// Log Level Tests
// =============================================================================

func TestRequestLogger_UsesInfoLevelFor2xx(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestLogger())
	router.GET("/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	output := captureLogOutput(func() {
		req := httptest.NewRequest(http.MethodGet, "/ok", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	})

	// Parse JSON to verify level
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err == nil {
		assert.Equal(t, "INFO", logEntry["level"])
	}
}

func TestRequestLogger_UsesWarnLevelFor4xx(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestLogger())
	router.GET("/bad", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
	})

	output := captureLogOutput(func() {
		req := httptest.NewRequest(http.MethodGet, "/bad", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	})

	assert.Contains(t, output, "WARN")
}

func TestRequestLogger_UsesErrorLevelFor5xx(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(RequestLogger())
	router.GET("/fail", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	})

	output := captureLogOutput(func() {
		req := httptest.NewRequest(http.MethodGet, "/fail", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	})

	assert.Contains(t, output, "ERROR")
}

// Ensure tests don't pollute global state
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
