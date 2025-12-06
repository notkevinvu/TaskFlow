package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/middleware"
	"github.com/notkevinvu/taskflow/backend/internal/ports"
)

// TaskTemplateHandler handles HTTP requests for task templates
type TaskTemplateHandler struct {
	templateService ports.TaskTemplateService
}

// NewTaskTemplateHandler creates a new task template handler
func NewTaskTemplateHandler(templateService ports.TaskTemplateService) *TaskTemplateHandler {
	return &TaskTemplateHandler{templateService: templateService}
}

// CreateTemplate creates a new task template
// POST /api/v1/templates
func (h *TaskTemplateHandler) CreateTemplate(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	var dto domain.CreateTaskTemplateDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		middleware.AbortWithError(c, domain.NewValidationError("request", err.Error()))
		return
	}

	template, err := h.templateService.Create(c.Request.Context(), userID, &dto)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusCreated, template)
}

// GetTemplate retrieves a specific template by ID
// GET /api/v1/templates/:id
func (h *TaskTemplateHandler) GetTemplate(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	templateID := c.Param("id")

	template, err := h.templateService.Get(c.Request.Context(), userID, templateID)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, template)
}

// ListTemplates retrieves all templates for the authenticated user
// GET /api/v1/templates
func (h *TaskTemplateHandler) ListTemplates(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	templates, err := h.templateService.List(c.Request.Context(), userID)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, domain.TaskTemplateListResponse{
		Templates:  templates,
		TotalCount: len(templates),
	})
}

// UpdateTemplate updates an existing template
// PUT /api/v1/templates/:id
func (h *TaskTemplateHandler) UpdateTemplate(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	templateID := c.Param("id")

	var dto domain.UpdateTaskTemplateDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		middleware.AbortWithError(c, domain.NewValidationError("request", err.Error()))
		return
	}

	template, err := h.templateService.Update(c.Request.Context(), userID, templateID, &dto)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, template)
}

// DeleteTemplate removes a template
// DELETE /api/v1/templates/:id
func (h *TaskTemplateHandler) DeleteTemplate(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	templateID := c.Param("id")

	if err := h.templateService.Delete(c.Request.Context(), userID, templateID); err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "template deleted",
	})
}

// UseTemplate converts a template to a CreateTaskDTO ready for task creation
// Returns the pre-filled task DTO without creating the task (frontend handles final creation)
// POST /api/v1/templates/:id/use
func (h *TaskTemplateHandler) UseTemplate(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	templateID := c.Param("id")

	// Optionally accept overrides in the request body
	var overrides *domain.CreateTaskDTO
	if c.Request.ContentLength > 0 {
		var dto domain.CreateTaskDTO
		if err := c.ShouldBindJSON(&dto); err != nil {
			middleware.AbortWithError(c, domain.NewValidationError("request", err.Error()))
			return
		}
		overrides = &dto
	}

	taskDTO, err := h.templateService.CreateTaskFromTemplate(c.Request.Context(), userID, templateID, overrides)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, taskDTO)
}
