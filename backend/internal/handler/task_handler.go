package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/middleware"
	"github.com/notkevinvu/taskflow/backend/internal/service"
)

// TaskHandler handles HTTP requests for tasks
type TaskHandler struct {
	taskService *service.TaskService
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

// Create handles task creation
// POST /api/v1/tasks
func (h *TaskHandler) Create(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var dto domain.CreateTaskDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	task, err := h.taskService.Create(c.Request.Context(), userID, &dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// List handles task listing with filters
// GET /api/v1/tasks?status=&category=&search=&limit=&offset=
func (h *TaskHandler) List(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse query parameters
	filter := &domain.TaskListFilter{
		Limit:  20, // Default limit
		Offset: 0,
	}

	if statusStr := c.Query("status"); statusStr != "" {
		status := domain.TaskStatus(statusStr)
		filter.Status = &status
	}

	if category := c.Query("category"); category != "" {
		filter.Category = &category
	}

	if search := c.Query("search"); search != "" {
		filter.Search = &search
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filter.Offset = offset
		}
	}

	tasks, err := h.taskService.List(c.Request.Context(), userID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	// Return empty array instead of null if no tasks
	if tasks == nil {
		tasks = []*domain.Task{}
	}

	c.JSON(http.StatusOK, gin.H{
		"tasks":       tasks,
		"total_count": len(tasks),
	})
}

// Get handles fetching a single task
// GET /api/v1/tasks/:id
func (h *TaskHandler) Get(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	taskID := c.Param("id")

	task, err := h.taskService.Get(c.Request.Context(), userID, taskID)
	if err != nil {
		switch err {
		case domain.ErrTaskNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		case domain.ErrUnauthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch task"})
		}
		return
	}

	c.JSON(http.StatusOK, task)
}

// Update handles task updates
// PUT /api/v1/tasks/:id
func (h *TaskHandler) Update(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	taskID := c.Param("id")

	var dto domain.UpdateTaskDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	task, err := h.taskService.Update(c.Request.Context(), userID, taskID, &dto)
	if err != nil {
		switch err {
		case domain.ErrTaskNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		case domain.ErrUnauthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		}
		return
	}

	c.JSON(http.StatusOK, task)
}

// Delete handles task deletion
// DELETE /api/v1/tasks/:id
func (h *TaskHandler) Delete(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	taskID := c.Param("id")

	err := h.taskService.Delete(c.Request.Context(), userID, taskID)
	if err != nil {
		switch err {
		case domain.ErrTaskNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		case domain.ErrUnauthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete task"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}

// Bump handles bumping a task (incrementing bump counter)
// POST /api/v1/tasks/:id/bump
func (h *TaskHandler) Bump(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	taskID := c.Param("id")

	task, err := h.taskService.Bump(c.Request.Context(), userID, taskID)
	if err != nil {
		switch err {
		case domain.ErrTaskNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		case domain.ErrUnauthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to bump task"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Task bumped successfully",
		"task":    task,
	})
}

// Complete handles marking a task as complete
// POST /api/v1/tasks/:id/complete
func (h *TaskHandler) Complete(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	taskID := c.Param("id")

	task, err := h.taskService.Complete(c.Request.Context(), userID, taskID)
	if err != nil {
		switch err {
		case domain.ErrTaskNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		case domain.ErrUnauthorized:
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete task"})
		}
		return
	}

	c.JSON(http.StatusOK, task)
}

// GetCalendar handles calendar view requests
// GET /api/v1/tasks/calendar?start_date=2025-11-01&end_date=2025-11-30&status=todo,in_progress
func (h *TaskHandler) GetCalendar(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse start_date (required)
	startDateStr := c.Query("start_date")
	if startDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_date is required (format: YYYY-MM-DD)"})
		return
	}
	startDate, err := domain.ParseDate(startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format (use YYYY-MM-DD)"})
		return
	}

	// Parse end_date (required)
	endDateStr := c.Query("end_date")
	if endDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_date is required (format: YYYY-MM-DD)"})
		return
	}
	endDate, err := domain.ParseDate(endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format (use YYYY-MM-DD)"})
		return
	}

	// Validate date range (max 90 days)
	if endDate.Sub(startDate).Hours() > 90*24 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Date range cannot exceed 90 days"})
		return
	}

	// Parse optional status filter (comma-separated)
	filter := &domain.CalendarFilter{
		StartDate: startDate,
		EndDate:   endDate,
		Status:    []domain.TaskStatus{},
	}

	if statusStr := c.Query("status"); statusStr != "" {
		// Split by comma and parse each status
		statusParts := splitAndTrim(statusStr, ",")
		for _, s := range statusParts {
			status := domain.TaskStatus(s)
			if err := status.Validate(); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status value: " + s})
				return
			}
			filter.Status = append(filter.Status, status)
		}
	}

	// Get calendar data
	calendar, err := h.taskService.GetCalendar(c.Request.Context(), userID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch calendar data"})
		return
	}

	c.JSON(http.StatusOK, calendar)
}

// splitAndTrim splits a string by delimiter and trims whitespace
func splitAndTrim(s, delimiter string) []string {
	if s == "" {
		return []string{}
	}
	parts := []string{}
	for _, part := range splitString(s, delimiter) {
		trimmed := trimSpace(part)
		if trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return parts
}

// splitString splits a string by delimiter
func splitString(s, delimiter string) []string {
	result := []string{}
	current := ""
	for _, char := range s {
		if string(char) == delimiter {
			result = append(result, current)
			current = ""
		} else {
			current += string(char)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// trimSpace removes leading/trailing whitespace
func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < len(s) && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n') {
		end--
	}
	return s[start:end]
}
