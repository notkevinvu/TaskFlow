package middleware

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
)

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Field   string                 `json:"field,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ErrorHandler middleware maps custom errors to HTTP status codes
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) == 0 {
			return
		}

		// Get the last error
		err := c.Errors.Last().Err

		// Map error to HTTP status and response
		statusCode, response := mapErrorToResponse(c, err)

		// Set status code and send JSON response
		c.JSON(statusCode, response)
	}
}

// mapErrorToResponse maps domain errors to HTTP status codes and responses
func mapErrorToResponse(c *gin.Context, err error) (int, ErrorResponse) {
	// Check for custom error types
	var validationErr *domain.ValidationError
	if errors.As(err, &validationErr) {
		return http.StatusBadRequest, ErrorResponse{
			Error: validationErr.Message,
			Field: validationErr.Field,
		}
	}

	var notFoundErr *domain.NotFoundError
	if errors.As(err, &notFoundErr) {
		return http.StatusNotFound, ErrorResponse{
			Error: notFoundErr.Error(),
		}
	}

	var conflictErr *domain.ConflictError
	if errors.As(err, &conflictErr) {
		return http.StatusConflict, ErrorResponse{
			Error: conflictErr.Message,
			Details: map[string]interface{}{
				"resource": conflictErr.Resource,
			},
		}
	}

	var unauthorizedErr *domain.UnauthorizedError
	if errors.As(err, &unauthorizedErr) {
		return http.StatusUnauthorized, ErrorResponse{
			Error: unauthorizedErr.Error(),
		}
	}

	var forbiddenErr *domain.ForbiddenError
	if errors.As(err, &forbiddenErr) {
		return http.StatusForbidden, ErrorResponse{
			Error: forbiddenErr.Error(),
		}
	}

	// Handle dependency sentinel errors
	if errors.Is(err, domain.ErrDependencyCycle) ||
		errors.Is(err, domain.ErrSelfDependency) ||
		errors.Is(err, domain.ErrInvalidDependencyType) {
		return http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		}
	}

	if errors.Is(err, domain.ErrDependencyNotFound) {
		return http.StatusNotFound, ErrorResponse{
			Error: err.Error(),
		}
	}

	if errors.Is(err, domain.ErrDependencyAlreadyExists) {
		return http.StatusConflict, ErrorResponse{
			Error: err.Error(),
		}
	}

	if errors.Is(err, domain.ErrCannotCompleteBlocked) {
		return http.StatusUnprocessableEntity, ErrorResponse{
			Error: err.Error(),
		}
	}

	var internalErr *domain.InternalError
	if errors.As(err, &internalErr) {
		// Log the internal error server-side with full details and request context
		slog.Error("Internal server error",
			"message", internalErr.Message,
			"cause", internalErr.Cause,
			"path", c.Request.URL.Path,
			"method", c.Request.Method,
		)
		// Don't expose internal error details to clients
		return http.StatusInternalServerError, ErrorResponse{
			Error: "An internal error occurred",
		}
	}

	// Log unexpected errors with request context
	slog.Error("Unexpected error",
		"error", err,
		"path", c.Request.URL.Path,
		"method", c.Request.Method,
	)

	// Default to internal server error
	return http.StatusInternalServerError, ErrorResponse{
		Error: "An unexpected error occurred",
	}
}

// AbortWithError is a helper to abort the request with a custom error
func AbortWithError(c *gin.Context, err error) {
	c.Error(err)
	c.Abort()
}
