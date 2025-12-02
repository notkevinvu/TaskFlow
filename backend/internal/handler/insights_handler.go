package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/middleware"
	"github.com/notkevinvu/taskflow/backend/internal/ports"
)

// InsightsHandler handles HTTP requests for insights and suggestions
type InsightsHandler struct {
	insightsService ports.InsightsService
	taskService     ports.TaskService
}

// NewInsightsHandler creates a new insights handler
func NewInsightsHandler(insightsService ports.InsightsService, taskService ports.TaskService) *InsightsHandler {
	return &InsightsHandler{
		insightsService: insightsService,
		taskService:     taskService,
	}
}

// GetInsights returns smart suggestions for the user
// GET /api/v1/insights
func (h *InsightsHandler) GetInsights(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	insights, err := h.insightsService.GetInsights(c.Request.Context(), userID)
	if err != nil {
		middleware.AbortWithError(c, domain.NewInternalError("failed to generate insights", err))
		return
	}

	c.JSON(http.StatusOK, insights)
}

// GetTimeEstimate returns a completion time estimate for a specific task
// GET /api/v1/tasks/:id/estimate
func (h *InsightsHandler) GetTimeEstimate(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	taskID := c.Param("id")
	if taskID == "" {
		middleware.AbortWithError(c, domain.NewValidationError("id", "task ID is required"))
		return
	}

	// Get the task first
	task, err := h.taskService.Get(c.Request.Context(), userID, taskID)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	// Generate estimate
	estimate, err := h.insightsService.EstimateCompletionTime(c.Request.Context(), userID, task)
	if err != nil {
		middleware.AbortWithError(c, domain.NewInternalError("failed to estimate completion time", err))
		return
	}

	c.JSON(http.StatusOK, estimate)
}

// SuggestCategoryRequest is the request body for category suggestion
type SuggestCategoryRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

// SuggestCategory returns category suggestions based on task content
// POST /api/v1/tasks/suggest-category
func (h *InsightsHandler) SuggestCategory(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	var req SuggestCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, domain.NewValidationError("body", "invalid request body"))
		return
	}

	suggestion, err := h.insightsService.SuggestCategory(c.Request.Context(), userID, &domain.CategorySuggestionRequest{
		Title:       req.Title,
		Description: req.Description,
	})
	if err != nil {
		middleware.AbortWithError(c, domain.NewInternalError("failed to suggest category", err))
		return
	}

	c.JSON(http.StatusOK, suggestion)
}
