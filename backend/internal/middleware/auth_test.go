package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

const testJWTSecret = "test-secret-key-for-jwt-testing-minimum-32-chars"

// Helper to generate a valid JWT token
func generateValidToken(userID, email string, expiresAt time.Time) string {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(testJWTSecret))
	return tokenString
}

// Helper to set up test router with auth middleware
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

// =============================================================================
// AuthRequired Middleware Tests
// =============================================================================

func TestAuthRequired_ValidToken(t *testing.T) {
	router := setupTestRouter()
	router.Use(AuthRequired(testJWTSecret))
	router.GET("/protected", func(c *gin.Context) {
		userID, _ := GetUserID(c)
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	token := generateValidToken("user-123", "test@example.com", time.Now().Add(time.Hour))

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "user-123")
}

func TestAuthRequired_MissingAuthHeader(t *testing.T) {
	router := setupTestRouter()
	router.Use(AuthRequired(testJWTSecret))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	// No Authorization header

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authorization header required")
}

func TestAuthRequired_InvalidHeaderFormat_NoBearerPrefix(t *testing.T) {
	router := setupTestRouter()
	router.Use(AuthRequired(testJWTSecret))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "some-token-without-bearer")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid authorization header format")
}

func TestAuthRequired_InvalidHeaderFormat_WrongPrefix(t *testing.T) {
	router := setupTestRouter()
	router.Use(AuthRequired(testJWTSecret))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	token := generateValidToken("user-123", "test@example.com", time.Now().Add(time.Hour))

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Basic "+token) // Wrong prefix

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid authorization header format")
}

func TestAuthRequired_ExpiredToken(t *testing.T) {
	router := setupTestRouter()
	router.Use(AuthRequired(testJWTSecret))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Token expired 1 hour ago
	token := generateValidToken("user-123", "test@example.com", time.Now().Add(-time.Hour))

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired token")
}

func TestAuthRequired_InvalidToken(t *testing.T) {
	router := setupTestRouter()
	router.Use(AuthRequired(testJWTSecret))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired token")
}

func TestAuthRequired_WrongSecret(t *testing.T) {
	router := setupTestRouter()
	router.Use(AuthRequired(testJWTSecret))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// Generate token with different secret
	claims := &Claims{
		UserID: "user-123",
		Email:  "test@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("different-secret-key"))

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired token")
}

func TestAuthRequired_MalformedToken(t *testing.T) {
	router := setupTestRouter()
	router.Use(AuthRequired(testJWTSecret))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer not.a.valid.jwt.token.format")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthRequired_EmptyToken(t *testing.T) {
	router := setupTestRouter()
	router.Use(AuthRequired(testJWTSecret))
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer ")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// =============================================================================
// GetUserID Tests
// =============================================================================

func TestGetUserID_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set(UserIDKey, "user-123")

	userID, exists := GetUserID(c)

	assert.True(t, exists)
	assert.Equal(t, "user-123", userID)
}

func TestGetUserID_NotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	// Don't set UserIDKey

	userID, exists := GetUserID(c)

	assert.False(t, exists)
	assert.Equal(t, "", userID)
}

// =============================================================================
// Integration Tests - Full Request Flow
// =============================================================================

func TestAuthRequired_FullFlow_MultipleRequests(t *testing.T) {
	router := setupTestRouter()
	router.Use(AuthRequired(testJWTSecret))
	router.GET("/me", func(c *gin.Context) {
		userID, _ := GetUserID(c)
		email, _ := c.Get(UserEmailKey)
		c.JSON(http.StatusOK, gin.H{
			"user_id": userID,
			"email":   email,
		})
	})

	// Test 1: Valid token
	token := generateValidToken("user-123", "test@example.com", time.Now().Add(time.Hour))
	req1, _ := http.NewRequest(http.MethodGet, "/me", nil)
	req1.Header.Set("Authorization", "Bearer "+token)
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)
	assert.Contains(t, w1.Body.String(), "user-123")
	assert.Contains(t, w1.Body.String(), "test@example.com")

	// Test 2: Different user with valid token
	token2 := generateValidToken("user-456", "other@example.com", time.Now().Add(time.Hour))
	req2, _ := http.NewRequest(http.MethodGet, "/me", nil)
	req2.Header.Set("Authorization", "Bearer "+token2)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
	assert.Contains(t, w2.Body.String(), "user-456")
	assert.Contains(t, w2.Body.String(), "other@example.com")

	// Test 3: No token (should fail)
	req3, _ := http.NewRequest(http.MethodGet, "/me", nil)
	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)
	assert.Equal(t, http.StatusUnauthorized, w3.Code)
}
