package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/middleware"
	"github.com/notkevinvu/taskflow/backend/internal/ports"
)

// AuthHandler handles HTTP requests for authentication
type AuthHandler struct {
	authService ports.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService ports.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register handles user registration
// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var dto domain.CreateUserDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		middleware.AbortWithError(c, domain.NewValidationError("request_body", "invalid JSON format"))
		return
	}

	response, err := h.authService.Register(c.Request.Context(), &dto)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

// Login handles user login
// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var dto domain.LoginDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		middleware.AbortWithError(c, domain.NewValidationError("request_body", "invalid JSON format"))
		return
	}

	response, err := h.authService.Login(c.Request.Context(), &dto)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// Me returns the current authenticated user
// GET /api/v1/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

// Guest creates a new anonymous guest user
// POST /api/v1/auth/guest
func (h *AuthHandler) Guest(c *gin.Context) {
	response, err := h.authService.CreateAnonymousUser(c.Request.Context())
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, response)
}

// ConvertGuest converts an anonymous user to a registered user
// POST /api/v1/auth/convert
func (h *AuthHandler) ConvertGuest(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	// Early check: only anonymous users can convert (fail fast using JWT claim)
	if !middleware.IsAnonymousUser(c) {
		middleware.AbortWithError(c, domain.NewValidationError("user", "only guest users can convert to registered accounts"))
		return
	}

	var dto domain.ConvertGuestDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		middleware.AbortWithError(c, domain.NewValidationError("request_body", "invalid JSON format"))
		return
	}

	response, err := h.authService.ConvertGuestToRegistered(c.Request.Context(), userID, &dto)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}
