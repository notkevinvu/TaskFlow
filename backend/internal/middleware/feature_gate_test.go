package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// Helper to create a test context with user claims
func setupTestContextWithUser(isAnonymous bool) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	// Set user claims using the correct context keys from auth.go
	c.Set(UserIDKey, "test-user-123")
	c.Set(UserIsAnonymous, isAnonymous)

	return c, w
}

// Helper to parse error response
type errorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Feature string `json:"feature,omitempty"`
	Upgrade bool   `json:"upgrade"`
}

func parseErrorResponse(t *testing.T, body []byte) errorResponse {
	var resp errorResponse
	err := json.Unmarshal(body, &resp)
	require.NoError(t, err, "Failed to parse error response")
	return resp
}

// =============================================================================
// RequireRegistered Tests
// =============================================================================

func TestRequireRegistered_AllowsRegisteredUser(t *testing.T) {
	c, w := setupTestContextWithUser(false) // Not anonymous = registered

	called := false
	handler := func(c *gin.Context) {
		called = true
		c.Status(http.StatusOK)
	}

	middleware := RequireRegistered()

	// Execute middleware then handler
	middleware(c)
	if !c.IsAborted() {
		handler(c)
	}

	assert.True(t, called, "Handler should be called for registered users")
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestRequireRegistered_BlocksAnonymousUser(t *testing.T) {
	c, w := setupTestContextWithUser(true) // Anonymous user

	called := false
	handler := func(c *gin.Context) {
		called = true
		c.Status(http.StatusOK)
	}

	middleware := RequireRegistered()

	// Execute middleware then handler
	middleware(c)
	if !c.IsAborted() {
		handler(c)
	}

	assert.False(t, called, "Handler should NOT be called for anonymous users")
	assert.Equal(t, http.StatusForbidden, w.Code)

	resp := parseErrorResponse(t, w.Body.Bytes())
	assert.Equal(t, "FEATURE_GATED", resp.Code)
	assert.True(t, resp.Upgrade)
	assert.Contains(t, resp.Error, "registered account")
}

func TestRequireRegistered_AbortsContext(t *testing.T) {
	c, _ := setupTestContextWithUser(true) // Anonymous user

	middleware := RequireRegistered()
	middleware(c)

	assert.True(t, c.IsAborted(), "Context should be aborted for anonymous users")
}

// =============================================================================
// RequireFeature Tests - Core Features (Allowed for Anonymous)
// =============================================================================

func TestRequireFeature_AllowsCoreFeatureForAnonymous(t *testing.T) {
	coreFeatures := []domain.Feature{
		domain.FeatureTaskCRUD,
		domain.FeatureTaskPriority,
		domain.FeatureTaskCategories,
		domain.FeatureTaskBump,
		domain.FeatureAnalytics,
		domain.FeatureInsights,
	}

	for _, feature := range coreFeatures {
		t.Run(string(feature), func(t *testing.T) {
			c, w := setupTestContextWithUser(true) // Anonymous user

			called := false
			handler := func(c *gin.Context) {
				called = true
				c.Status(http.StatusOK)
			}

			middleware := RequireFeature(feature)
			middleware(c)
			if !c.IsAborted() {
				handler(c)
			}

			assert.True(t, called, "Handler should be called for core feature %s", feature)
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

// =============================================================================
// RequireFeature Tests - Advanced Features (Restricted for Anonymous)
// =============================================================================

func TestRequireFeature_BlocksAdvancedFeatureForAnonymous(t *testing.T) {
	advancedFeatures := []domain.Feature{
		domain.FeatureSubtasks,
		domain.FeatureDependencies,
		domain.FeatureTemplates,
		domain.FeatureGamification,
		domain.FeatureRecurring,
	}

	for _, feature := range advancedFeatures {
		t.Run(string(feature), func(t *testing.T) {
			c, w := setupTestContextWithUser(true) // Anonymous user

			called := false
			handler := func(c *gin.Context) {
				called = true
				c.Status(http.StatusOK)
			}

			middleware := RequireFeature(feature)
			middleware(c)
			if !c.IsAborted() {
				handler(c)
			}

			assert.False(t, called, "Handler should NOT be called for advanced feature %s", feature)
			assert.Equal(t, http.StatusForbidden, w.Code)

			resp := parseErrorResponse(t, w.Body.Bytes())
			assert.Equal(t, "FEATURE_GATED", resp.Code)
			assert.Equal(t, string(feature), resp.Feature)
			assert.True(t, resp.Upgrade)
		})
	}
}

// =============================================================================
// RequireFeature Tests - Registered Users (Always Allowed)
// =============================================================================

func TestRequireFeature_AllowsAllFeaturesForRegisteredUser(t *testing.T) {
	allFeatures := []domain.Feature{
		// Core features
		domain.FeatureTaskCRUD,
		domain.FeatureTaskPriority,
		domain.FeatureTaskCategories,
		domain.FeatureTaskBump,
		domain.FeatureAnalytics,
		domain.FeatureInsights,
		// Advanced features
		domain.FeatureSubtasks,
		domain.FeatureDependencies,
		domain.FeatureTemplates,
		domain.FeatureGamification,
		domain.FeatureRecurring,
	}

	for _, feature := range allFeatures {
		t.Run(string(feature), func(t *testing.T) {
			c, w := setupTestContextWithUser(false) // Registered user

			called := false
			handler := func(c *gin.Context) {
				called = true
				c.Status(http.StatusOK)
			}

			middleware := RequireFeature(feature)
			middleware(c)
			if !c.IsAborted() {
				handler(c)
			}

			assert.True(t, called, "Handler should be called for registered user with feature %s", feature)
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}

// =============================================================================
// RequireFeature Tests - Unknown Features
// =============================================================================

func TestRequireFeature_BlocksUnknownFeatureForAnonymous(t *testing.T) {
	c, w := setupTestContextWithUser(true) // Anonymous user
	unknownFeature := domain.Feature("unknown_feature")

	called := false
	handler := func(c *gin.Context) {
		called = true
		c.Status(http.StatusOK)
	}

	middleware := RequireFeature(unknownFeature)
	middleware(c)
	if !c.IsAborted() {
		handler(c)
	}

	assert.False(t, called, "Handler should NOT be called for unknown feature")
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireFeature_AllowsUnknownFeatureForRegistered(t *testing.T) {
	c, w := setupTestContextWithUser(false) // Registered user
	unknownFeature := domain.Feature("unknown_feature")

	called := false
	handler := func(c *gin.Context) {
		called = true
		c.Status(http.StatusOK)
	}

	middleware := RequireFeature(unknownFeature)
	middleware(c)
	if !c.IsAborted() {
		handler(c)
	}

	assert.True(t, called, "Handler should be called for registered user with unknown feature")
	assert.Equal(t, http.StatusOK, w.Code)
}

// =============================================================================
// Integration Tests - Middleware Chain
// =============================================================================

func TestRequireFeature_DoesNotAbortOnSuccess(t *testing.T) {
	c, _ := setupTestContextWithUser(false) // Registered user

	middleware := RequireFeature(domain.FeatureSubtasks)
	middleware(c)

	assert.False(t, c.IsAborted(), "Context should not be aborted for allowed access")
}

func TestRequireFeature_CallsNextOnSuccess(t *testing.T) {
	c, w := setupTestContextWithUser(false) // Registered user

	// Chain: RequireFeature -> handler
	var order []string

	middleware := RequireFeature(domain.FeatureSubtasks)
	middleware(c)
	order = append(order, "middleware")

	if !c.IsAborted() {
		c.Status(http.StatusOK)
		order = append(order, "handler")
	}

	assert.Equal(t, []string{"middleware", "handler"}, order)
	assert.Equal(t, http.StatusOK, w.Code)
}

// =============================================================================
// Response Format Tests
// =============================================================================

func TestRequireFeature_ErrorResponseFormat(t *testing.T) {
	c, w := setupTestContextWithUser(true) // Anonymous user

	middleware := RequireFeature(domain.FeatureSubtasks)
	middleware(c)

	// Verify response is valid JSON
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err, "Response should be valid JSON")

	// Verify required fields
	assert.Contains(t, resp, "error")
	assert.Contains(t, resp, "code")
	assert.Contains(t, resp, "feature")
	assert.Contains(t, resp, "upgrade")

	// Verify field types
	assert.IsType(t, "", resp["error"])
	assert.IsType(t, "", resp["code"])
	assert.IsType(t, "", resp["feature"])
	assert.IsType(t, true, resp["upgrade"])
}

func TestRequireRegistered_ErrorResponseFormat(t *testing.T) {
	c, w := setupTestContextWithUser(true) // Anonymous user

	middleware := RequireRegistered()
	middleware(c)

	// Verify response is valid JSON
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err, "Response should be valid JSON")

	// Verify required fields
	assert.Contains(t, resp, "error")
	assert.Contains(t, resp, "code")
	assert.Contains(t, resp, "upgrade")
}
