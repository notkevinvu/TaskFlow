package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/middleware"
	"github.com/notkevinvu/taskflow/backend/internal/ports"
)

// DependencyHandler handles HTTP requests for task dependencies
type DependencyHandler struct {
	dependencyService ports.DependencyService
}

// NewDependencyHandler creates a new dependency handler
func NewDependencyHandler(dependencyService ports.DependencyService) *DependencyHandler {
	return &DependencyHandler{dependencyService: dependencyService}
}

// AddDependency creates a "blocked by" relationship between two tasks
// POST /api/v1/tasks/:id/dependencies
// Request body: { "blocked_by_id": "uuid" }
func (h *DependencyHandler) AddDependency(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	taskID := c.Param("id")

	var req struct {
		BlockedByID string `json:"blocked_by_id" binding:"required,uuid"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, domain.NewValidationError("blocked_by_id", "must be a valid UUID"))
		return
	}

	dto := &domain.AddDependencyDTO{
		TaskID:      taskID,
		BlockedByID: req.BlockedByID,
	}

	info, err := h.dependencyService.AddDependency(c.Request.Context(), userID, dto)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, info)
}

// RemoveDependency removes a "blocked by" relationship
// DELETE /api/v1/tasks/:id/dependencies/:blocker_id
func (h *DependencyHandler) RemoveDependency(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	taskID := c.Param("id")
	blockerID := c.Param("blocker_id")

	dto := &domain.RemoveDependencyDTO{
		TaskID:      taskID,
		BlockedByID: blockerID,
	}

	if err := h.dependencyService.RemoveDependency(c.Request.Context(), userID, dto); err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "dependency removed",
	})
}

// GetDependencyInfo retrieves complete dependency information for a task
// GET /api/v1/tasks/:id/dependencies
func (h *DependencyHandler) GetDependencyInfo(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	taskID := c.Param("id")

	info, err := h.dependencyService.GetDependencyInfo(c.Request.Context(), userID, taskID)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, info)
}

// CheckCanComplete checks if a task can be completed (no incomplete blockers)
// GET /api/v1/tasks/:id/can-complete-dependencies
func (h *DependencyHandler) CheckCanComplete(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	taskID := c.Param("id")

	// Use GetDependencyInfo which includes CanComplete
	info, err := h.dependencyService.GetDependencyInfo(c.Request.Context(), userID, taskID)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"can_complete":       info.CanComplete,
		"is_blocked":         info.IsBlocked,
		"incomplete_blockers": countIncompleteBlockers(info.Blockers),
	})
}

// countIncompleteBlockers counts blockers that are not done
func countIncompleteBlockers(blockers []*domain.DependencyWithTask) int {
	count := 0
	for _, b := range blockers {
		if b.Status != domain.TaskStatusDone {
			count++
		}
	}
	return count
}
