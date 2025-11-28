package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/middleware"
	"github.com/notkevinvu/taskflow/backend/internal/ports"
)

// CategoryHandler handles HTTP requests for category management
type CategoryHandler struct {
	taskService ports.TaskService
}

// NewCategoryHandler creates a new category handler
func NewCategoryHandler(taskService ports.TaskService) *CategoryHandler {
	return &CategoryHandler{taskService: taskService}
}

// RenameRequest represents a category rename request
type RenameRequest struct {
	OldName string `json:"old_name" binding:"required"`
	NewName string `json:"new_name" binding:"required"`
}

// Rename handles category renaming
// PUT /api/v1/categories/rename
func (h *CategoryHandler) Rename(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	var req RenameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.AbortWithError(c, domain.NewValidationError("request_body", "invalid JSON format"))
		return
	}

	// Service layer will handle validation
	updatedCount, err := h.taskService.RenameCategory(c.Request.Context(), userID, req.OldName, req.NewName)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Category renamed successfully",
		"updated_count": updatedCount,
	})
}

// Delete handles category deletion
// DELETE /api/v1/categories/:name
func (h *CategoryHandler) Delete(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	categoryName := c.Param("name")

	// Service layer will handle validation
	updatedCount, err := h.taskService.DeleteCategory(c.Request.Context(), userID, categoryName)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Category deleted successfully",
		"updated_count": updatedCount,
	})
}
