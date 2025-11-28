package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req RenameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Validate names
	if req.OldName == "" || req.NewName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category names cannot be empty"})
		return
	}

	// Trim whitespace and validate length (matching domain validation: max 50 chars)
	req.NewName = strings.TrimSpace(req.NewName)
	req.OldName = strings.TrimSpace(req.OldName)

	if len(req.NewName) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category name too long (maximum 50 characters)"})
		return
	}

	if req.NewName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category name cannot be empty or whitespace only"})
		return
	}

	if req.OldName == req.NewName {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Old and new category names are the same"})
		return
	}

	// Update all tasks with the old category to the new category
	updatedCount, err := h.taskService.RenameCategory(c.Request.Context(), userID, req.OldName, req.NewName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to rename category"})
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	categoryName := strings.TrimSpace(c.Param("name"))
	if categoryName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category name is required"})
		return
	}

	if len(categoryName) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category name too long (maximum 50 characters)"})
		return
	}

	// Remove category from all tasks (set to null/empty)
	updatedCount, err := h.taskService.DeleteCategory(c.Request.Context(), userID, categoryName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Category deleted successfully",
		"updated_count": updatedCount,
	})
}
