package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/middleware"
	"github.com/notkevinvu/taskflow/backend/internal/ports"
)

// GamificationHandler handles gamification HTTP requests
type GamificationHandler struct {
	gamificationService ports.GamificationService
}

// NewGamificationHandler creates a new gamification handler
func NewGamificationHandler(gamificationService ports.GamificationService) *GamificationHandler {
	return &GamificationHandler{gamificationService: gamificationService}
}

// GetDashboard retrieves gamification dashboard data
// GET /api/v1/gamification/dashboard
func (h *GamificationHandler) GetDashboard(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	dashboard, err := h.gamificationService.GetDashboard(c.Request.Context(), userID)
	if err != nil {
		middleware.AbortWithError(c, domain.NewInternalError("failed to get gamification dashboard", err))
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

// GetStats retrieves current gamification stats only
// GET /api/v1/gamification/stats
func (h *GamificationHandler) GetStats(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	stats, err := h.gamificationService.ComputeStats(c.Request.Context(), userID)
	if err != nil {
		middleware.AbortWithError(c, domain.NewInternalError("failed to get gamification stats", err))
		return
	}

	c.JSON(http.StatusOK, stats)
}

// SetTimezone updates user's timezone for streak calculation
// PUT /api/v1/gamification/timezone
func (h *GamificationHandler) SetTimezone(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	var dto domain.UpdateTimezoneDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		middleware.AbortWithError(c, domain.NewValidationError("timezone", "invalid timezone format"))
		return
	}

	if err := h.gamificationService.SetUserTimezone(c.Request.Context(), userID, dto.Timezone); err != nil {
		if err == domain.ErrInvalidTimezone {
			middleware.AbortWithError(c, domain.NewValidationError("timezone", "invalid IANA timezone"))
			return
		}
		middleware.AbortWithError(c, domain.NewInternalError("failed to set timezone", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Timezone updated successfully", "timezone": dto.Timezone})
}

// GetTimezone retrieves user's current timezone setting
// GET /api/v1/gamification/timezone
func (h *GamificationHandler) GetTimezone(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	timezone, err := h.gamificationService.GetUserTimezone(c.Request.Context(), userID)
	if err != nil {
		middleware.AbortWithError(c, domain.NewInternalError("failed to get timezone", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{"timezone": timezone})
}
