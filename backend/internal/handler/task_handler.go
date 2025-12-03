package handler

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/metrics"
	"github.com/notkevinvu/taskflow/backend/internal/middleware"
	"github.com/notkevinvu/taskflow/backend/internal/ports"
)

// TaskHandler handles HTTP requests for tasks
type TaskHandler struct {
	taskService ports.TaskService
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(taskService ports.TaskService) *TaskHandler {
	return &TaskHandler{taskService: taskService}
}

// Create handles task creation
// POST /api/v1/tasks
func (h *TaskHandler) Create(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	var dto domain.CreateTaskDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		middleware.AbortWithError(c, domain.NewValidationError("request_body", "invalid JSON format"))
		return
	}

	task, err := h.taskService.Create(c.Request.Context(), userID, &dto)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	// Record metrics
	category := ""
	if task.Category != nil {
		category = *task.Category
	}
	effort := "medium"
	if task.EstimatedEffort != nil {
		effort = string(*task.EstimatedEffort)
	}
	metrics.RecordTaskCreated(category, effort)

	c.JSON(http.StatusCreated, task)
}

// List handles task listing with filters
// GET /api/v1/tasks?status=&category=&search=&min_priority=&max_priority=&due_date_start=&due_date_end=&limit=&offset=
func (h *TaskHandler) List(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
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

	if minPriorityStr := c.Query("min_priority"); minPriorityStr != "" {
		minPriority, err := strconv.Atoi(minPriorityStr)
		if err != nil {
			middleware.AbortWithError(c, domain.NewValidationError("min_priority", "must be a number"))
			return
		}
		if minPriority < 0 || minPriority > 100 {
			middleware.AbortWithError(c, domain.NewValidationError("min_priority", "must be between 0 and 100"))
			return
		}
		filter.MinPriority = &minPriority
	}

	if maxPriorityStr := c.Query("max_priority"); maxPriorityStr != "" {
		maxPriority, err := strconv.Atoi(maxPriorityStr)
		if err != nil {
			middleware.AbortWithError(c, domain.NewValidationError("max_priority", "must be a number"))
			return
		}
		if maxPriority < 0 || maxPriority > 100 {
			middleware.AbortWithError(c, domain.NewValidationError("max_priority", "must be between 0 and 100"))
			return
		}
		filter.MaxPriority = &maxPriority
	}

	// Validate priority range if both are specified
	if filter.MinPriority != nil && filter.MaxPriority != nil {
		if *filter.MinPriority > *filter.MaxPriority {
			middleware.AbortWithError(c, domain.NewValidationError("min_priority", "cannot be greater than max_priority"))
			return
		}
	}

	if dueDateStartStr := c.Query("due_date_start"); dueDateStartStr != "" {
		dueDateStart, err := domain.ParseDate(dueDateStartStr)
		if err != nil {
			middleware.AbortWithError(c, domain.NewValidationError("due_date_start", "must be in YYYY-MM-DD format"))
			return
		}
		filter.DueDateStart = &dueDateStart
	}

	if dueDateEndStr := c.Query("due_date_end"); dueDateEndStr != "" {
		dueDateEnd, err := domain.ParseDate(dueDateEndStr)
		if err != nil {
			middleware.AbortWithError(c, domain.NewValidationError("due_date_end", "must be in YYYY-MM-DD format"))
			return
		}
		filter.DueDateEnd = &dueDateEnd
	}

	tasks, err := h.taskService.List(c.Request.Context(), userID, filter)
	if err != nil {
		middleware.AbortWithError(c, err)
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
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	taskID := c.Param("id")

	task, err := h.taskService.Get(c.Request.Context(), userID, taskID)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, task)
}

// Update handles task updates
// PUT /api/v1/tasks/:id
func (h *TaskHandler) Update(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	taskID := c.Param("id")

	var dto domain.UpdateTaskDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		middleware.AbortWithError(c, domain.NewValidationError("request_body", "invalid JSON format"))
		return
	}

	task, err := h.taskService.Update(c.Request.Context(), userID, taskID, &dto)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, task)
}

// Delete handles task deletion
// DELETE /api/v1/tasks/:id
func (h *TaskHandler) Delete(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	taskID := c.Param("id")

	// Get task before deletion to record metrics
	task, getErr := h.taskService.Get(c.Request.Context(), userID, taskID)
	if getErr != nil {
		// Log but don't fail - metrics are secondary to deletion
		slog.Warn("Failed to get task for deletion metrics",
			"taskID", taskID,
			"userID", userID,
			"error", getErr,
		)
	}

	err := h.taskService.Delete(c.Request.Context(), userID, taskID)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	// Record metrics if we got the task info
	if getErr == nil && task != nil {
		category := ""
		if task.Category != nil {
			category = *task.Category
		}
		metrics.RecordTaskDeleted(category)
	}

	c.Status(http.StatusNoContent)
}

// Bump handles bumping a task (incrementing bump counter)
// POST /api/v1/tasks/:id/bump
func (h *TaskHandler) Bump(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	taskID := c.Param("id")

	task, err := h.taskService.Bump(c.Request.Context(), userID, taskID)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	// Record metrics
	category := ""
	if task.Category != nil {
		category = *task.Category
	}
	metrics.RecordTaskBumped(category)

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
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	taskID := c.Param("id")

	task, err := h.taskService.Complete(c.Request.Context(), userID, taskID)
	if err != nil {
		middleware.AbortWithError(c, err)
		return
	}

	// Record metrics
	category := ""
	if task.Category != nil {
		category = *task.Category
	}
	effort := "medium"
	if task.EstimatedEffort != nil {
		effort = string(*task.EstimatedEffort)
	}
	metrics.RecordTaskCompleted(category, effort)

	c.JSON(http.StatusOK, task)
}

// GetCalendar handles calendar view requests
// GET /api/v1/tasks/calendar?start_date=2025-11-01&end_date=2025-11-30&status=todo,in_progress
func (h *TaskHandler) GetCalendar(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	// Parse start_date (required)
	startDateStr := c.Query("start_date")
	if startDateStr == "" {
		middleware.AbortWithError(c, domain.NewValidationError("start_date", "is required (format: YYYY-MM-DD)"))
		return
	}
	startDate, err := domain.ParseDate(startDateStr)
	if err != nil {
		middleware.AbortWithError(c, domain.NewValidationError("start_date", "invalid format (use YYYY-MM-DD)"))
		return
	}

	// Parse end_date (required)
	endDateStr := c.Query("end_date")
	if endDateStr == "" {
		middleware.AbortWithError(c, domain.NewValidationError("end_date", "is required (format: YYYY-MM-DD)"))
		return
	}
	endDate, err := domain.ParseDate(endDateStr)
	if err != nil {
		middleware.AbortWithError(c, domain.NewValidationError("end_date", "invalid format (use YYYY-MM-DD)"))
		return
	}

	// Validate date range (max 90 days)
	if endDate.Sub(startDate).Hours() > 90*24 {
		middleware.AbortWithError(c, domain.NewValidationError("date_range", "cannot exceed 90 days"))
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
				middleware.AbortWithError(c, domain.NewValidationError("status", "invalid status value: "+s))
				return
			}
			filter.Status = append(filter.Status, status)
		}
	}

	// Get calendar data
	calendar, err := h.taskService.GetCalendar(c.Request.Context(), userID, filter)
	if err != nil {
		middleware.AbortWithError(c, err)
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
