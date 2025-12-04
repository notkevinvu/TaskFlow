package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/middleware"
	"github.com/notkevinvu/taskflow/backend/internal/ports"
)

// RecurrenceHandler handles HTTP requests for recurring tasks and preferences
type RecurrenceHandler struct {
	recurrenceService ports.RecurrenceService
}

// NewRecurrenceHandler creates a new recurrence handler
func NewRecurrenceHandler(recurrenceService ports.RecurrenceService) *RecurrenceHandler {
	return &RecurrenceHandler{recurrenceService: recurrenceService}
}

// ListSeries returns all task series for the authenticated user
// GET /api/v1/series?active_only=true
func (h *RecurrenceHandler) ListSeries(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	activeOnly := false
	if activeOnlyStr := c.Query("active_only"); activeOnlyStr != "" {
		activeOnly, _ = strconv.ParseBool(activeOnlyStr)
	}

	series, err := h.recurrenceService.GetUserSeries(c.Request.Context(), userID, activeOnly)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	// Return empty array instead of null
	if series == nil {
		series = []*domain.TaskSeries{}
	}

	c.JSON(http.StatusOK, gin.H{
		"series":      series,
		"total_count": len(series),
	})
}

// GetSeriesHistory returns the history of a task series
// GET /api/v1/series/:id/history
func (h *RecurrenceHandler) GetSeriesHistory(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	seriesID := c.Param("id")

	history, err := h.recurrenceService.GetSeriesHistory(c.Request.Context(), userID, seriesID)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, history)
}

// UpdateSeries updates a task series
// PUT /api/v1/series/:id
func (h *RecurrenceHandler) UpdateSeries(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	seriesID := c.Param("id")

	var dto domain.UpdateTaskSeriesDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		middleware.AbortWithError(c, domain.NewValidationError("request_body", "invalid JSON format"))
		return
	}

	series, err := h.recurrenceService.UpdateSeries(c.Request.Context(), userID, seriesID, &dto)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, series)
}

// DeactivateSeries stops a recurring series
// POST /api/v1/series/:id/deactivate
func (h *RecurrenceHandler) DeactivateSeries(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	seriesID := c.Param("id")

	err := h.recurrenceService.DeactivateSeries(c.Request.Context(), userID, seriesID)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Series deactivated successfully",
	})
}

// GetPreferences returns all user preferences for recurring tasks
// GET /api/v1/preferences/recurrence
func (h *RecurrenceHandler) GetPreferences(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	prefs, err := h.recurrenceService.GetUserPreferences(c.Request.Context(), userID)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	// Return empty preferences if none exist
	if prefs == nil {
		prefs = &domain.AllPreferences{
			UserPreferences:     nil,
			CategoryPreferences: []domain.CategoryPreference{},
		}
	}

	c.JSON(http.StatusOK, prefs)
}

// SetDefaultPreference sets the user's default due date calculation mode
// PUT /api/v1/preferences/recurrence/default
func (h *RecurrenceHandler) SetDefaultPreference(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	var req struct {
		DueDateCalculation domain.DueDateCalculation `json:"due_date_calculation"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, domain.NewValidationError("request_body", "invalid JSON format"))
		return
	}

	// Validate calculation mode
	if err := req.DueDateCalculation.Validate(); err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	err := h.recurrenceService.SetDefaultPreference(c.Request.Context(), userID, req.DueDateCalculation)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":              "Default preference saved",
		"due_date_calculation": req.DueDateCalculation,
	})
}

// SetCategoryPreference sets the due date calculation mode for a specific category
// PUT /api/v1/preferences/recurrence/category/:category
func (h *RecurrenceHandler) SetCategoryPreference(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	category := c.Param("category")
	if category == "" {
		middleware.AbortWithError(c, domain.NewValidationError("category", "is required"))
		return
	}

	var req struct {
		DueDateCalculation domain.DueDateCalculation `json:"due_date_calculation"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, domain.NewValidationError("request_body", "invalid JSON format"))
		return
	}

	// Validate calculation mode
	if err := req.DueDateCalculation.Validate(); err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	err := h.recurrenceService.SetCategoryPreference(c.Request.Context(), userID, category, req.DueDateCalculation)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":              "Category preference saved",
		"category":             category,
		"due_date_calculation": req.DueDateCalculation,
	})
}

// DeleteCategoryPreference removes a category preference
// DELETE /api/v1/preferences/recurrence/category/:category
func (h *RecurrenceHandler) DeleteCategoryPreference(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	category := c.Param("category")
	if category == "" {
		middleware.AbortWithError(c, domain.NewValidationError("category", "is required"))
		return
	}

	err := h.recurrenceService.DeleteCategoryPreference(c.Request.Context(), userID, category)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// GetEffectiveCalculation returns the effective due date calculation for a task
// GET /api/v1/preferences/recurrence/effective?category=Work
func (h *RecurrenceHandler) GetEffectiveCalculation(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	var category *string
	if cat := c.Query("category"); cat != "" {
		category = &cat
	}

	calc := h.recurrenceService.GetEffectiveDueDateCalculation(c.Request.Context(), userID, category)

	c.JSON(http.StatusOK, gin.H{
		"due_date_calculation": calc,
		"category":             category,
	})
}
