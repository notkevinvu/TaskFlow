package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Context keys for storing user information
const (
	UserIDKey        = "user_id"
	UserEmailKey     = "user_email"
	UserIsAnonymous  = "user_is_anonymous"
)

// Claims represents the JWT claims
type Claims struct {
	UserID      string `json:"user_id"`
	Email       string `json:"email"`
	IsAnonymous bool   `json:"is_anonymous"`
	jwt.RegisteredClaims
}

// AuthRequired is a middleware that validates JWT tokens
func AuthRequired(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(*Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Set user information in context
		c.Set(UserIDKey, claims.UserID)
		c.Set(UserEmailKey, claims.Email)
		c.Set(UserIsAnonymous, claims.IsAnonymous)

		c.Next()
	}
}

// GetUserID retrieves the user ID from the context
func GetUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return "", false
	}
	return userID.(string), true
}

// IsAnonymousUser checks if the current user is anonymous
func IsAnonymousUser(c *gin.Context) bool {
	isAnon, exists := c.Get(UserIsAnonymous)
	if !exists {
		return false
	}
	return isAnon.(bool)
}
