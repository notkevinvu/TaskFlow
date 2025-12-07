package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
)

// RequireRegistered is a middleware that ensures the user is a registered user.
// Anonymous users will receive a 403 Forbidden response with upgrade instructions.
func RequireRegistered() gin.HandlerFunc {
	return func(c *gin.Context) {
		if IsAnonymousUser(c) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "This feature requires a registered account",
				"code":    "FEATURE_GATED",
				"upgrade": true,
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireFeature creates a middleware that checks if the current user can access
// a specific feature. Registered users always have access, anonymous users are
// checked against the allowed features list.
func RequireFeature(feature domain.Feature) gin.HandlerFunc {
	return func(c *gin.Context) {
		isAnonymous := IsAnonymousUser(c)

		// Determine user type
		userType := domain.UserTypeRegistered
		if isAnonymous {
			userType = domain.UserTypeAnonymous
		}

		// Check feature access
		if !domain.CanAccessFeature(userType, feature) {
			err := domain.NewFeatureGatedError(feature)
			c.JSON(http.StatusForbidden, gin.H{
				"error":   err.Error(),
				"code":    "FEATURE_GATED",
				"feature": feature,
				"upgrade": true,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
