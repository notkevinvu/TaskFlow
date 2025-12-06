package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/middleware"
	"github.com/notkevinvu/taskflow/backend/internal/ports"
)

// SubtaskHandler handles HTTP requests for subtasks
type SubtaskHandler struct {
	subtaskService ports.SubtaskService
}

// NewSubtaskHandler creates a new subtask handler
func NewSubtaskHandler(subtaskService ports.SubtaskService) *SubtaskHandler {
	return &SubtaskHandler{subtaskService: subtaskService}
}

// CreateSubtask handles subtask creation under a parent task
// POST /api/v1/tasks/:id/subtasks
func (h *SubtaskHandler) CreateSubtask(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	parentTaskID := c.Param("id")

	var dto domain.CreateSubtaskDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		middleware.AbortWithError(c, domain.NewValidationError("request_body", "invalid JSON format"))
		return
	}

	// Set parent task ID from URL parameter
	dto.ParentTaskID = parentTaskID

	subtask, err := h.subtaskService.Create(c.Request.Context(), userID, &dto)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, subtask)
}

// GetSubtasks retrieves all subtasks for a parent task
// GET /api/v1/tasks/:id/subtasks
func (h *SubtaskHandler) GetSubtasks(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	parentTaskID := c.Param("id")

	subtasks, err := h.subtaskService.GetSubtasks(c.Request.Context(), userID, parentTaskID)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	// Return empty array instead of null
	if subtasks == nil {
		subtasks = []*domain.Task{}
	}

	c.JSON(http.StatusOK, gin.H{
		"subtasks":    subtasks,
		"total_count": len(subtasks),
	})
}

// GetSubtaskInfo retrieves aggregated subtask statistics for a parent task
// GET /api/v1/tasks/:id/subtask-info
func (h *SubtaskHandler) GetSubtaskInfo(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	parentTaskID := c.Param("id")

	info, err := h.subtaskService.GetSubtaskInfo(c.Request.Context(), userID, parentTaskID)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, info)
}

// GetTaskExpanded retrieves a task with subtask info and optionally expanded subtasks
// GET /api/v1/tasks/:id/expanded?include_subtasks=true
func (h *SubtaskHandler) GetTaskExpanded(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	taskID := c.Param("id")

	// Check if subtasks should be expanded
	expandSubtasks := c.Query("include_subtasks") == "true"

	taskWithSubtasks, err := h.subtaskService.GetTaskWithSubtasks(c.Request.Context(), userID, taskID, expandSubtasks)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, taskWithSubtasks)
}

// CompleteSubtask marks a subtask as complete and returns parent completion prompt info
// POST /api/v1/subtasks/:id/complete
func (h *SubtaskHandler) CompleteSubtask(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	subtaskID := c.Param("id")

	response, err := h.subtaskService.CompleteSubtask(c.Request.Context(), userID, subtaskID)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

// CanCompleteParent checks if a parent task can be completed (all subtasks done)
// GET /api/v1/tasks/:id/can-complete
func (h *SubtaskHandler) CanCompleteParent(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	parentTaskID := c.Param("id")

	canComplete, err := h.subtaskService.CanCompleteParent(c.Request.Context(), userID, parentTaskID)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"can_complete": canComplete,
	})
}
