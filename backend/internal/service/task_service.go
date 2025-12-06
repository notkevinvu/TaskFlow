package service

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/domain/priority"
	"github.com/notkevinvu/taskflow/backend/internal/ports"
	"github.com/notkevinvu/taskflow/backend/internal/validation"
)

// TaskService handles task business logic
type TaskService struct {
	taskRepo          ports.TaskRepository
	taskHistoryRepo   ports.TaskHistoryRepository
	priorityCalc      *priority.Calculator
	recurrenceService ports.RecurrenceService  // Optional: for recurring task support
	subtaskService    ports.SubtaskService     // Optional: for subtask validation
	dependencyService ports.DependencyService  // Optional: for dependency validation
}

// NewTaskService creates a new task service
func NewTaskService(taskRepo ports.TaskRepository, taskHistoryRepo ports.TaskHistoryRepository) *TaskService {
	return &TaskService{
		taskRepo:        taskRepo,
		taskHistoryRepo: taskHistoryRepo,
		priorityCalc:    priority.NewCalculator(),
	}
}

// SetRecurrenceService sets the optional recurrence service for recurring task support
func (s *TaskService) SetRecurrenceService(recurrenceService ports.RecurrenceService) {
	s.recurrenceService = recurrenceService
}

// SetSubtaskService sets the optional subtask service for parent completion validation
func (s *TaskService) SetSubtaskService(subtaskService ports.SubtaskService) {
	s.subtaskService = subtaskService
}

// SetDependencyService sets the optional dependency service for blocker validation
func (s *TaskService) SetDependencyService(dependencyService ports.DependencyService) {
	s.dependencyService = dependencyService
}

// Create creates a new task
func (s *TaskService) Create(ctx context.Context, userID string, dto *domain.CreateTaskDTO) (*domain.Task, error) {
	// Validate title
	title, err := validation.ValidateRequiredText(dto.Title, 200, "title")
	if err != nil {
		return nil, err
	}

	// Validate optional description
	description, err := validation.ValidateOptionalText(dto.Description, 2000, "description")
	if err != nil {
		return nil, err
	}

	// Validate optional category
	category, err := validation.ValidateCategory(dto.Category)
	if err != nil {
		return nil, err
	}

	// Validate optional context
	contextVal, err := validation.ValidateOptionalText(dto.Context, 500, "context")
	if err != nil {
		return nil, err
	}

	// Validate related people
	relatedPeople, err := validation.ValidateStringSlice(dto.RelatedPeople, 100, 20, "related_people")
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// Set default user priority if not provided (1-10 scale)
	userPriority := 5
	if dto.UserPriority != nil {
		if err := validation.ValidatePriority(*dto.UserPriority); err != nil {
			return nil, err
		}
		userPriority = *dto.UserPriority
	}

	// Create task
	task := &domain.Task{
		ID:              uuid.New().String(),
		UserID:          userID,
		Title:           title,
		Description:     description,
		Status:          domain.TaskStatusTodo,
		UserPriority:    userPriority,
		DueDate:         dto.DueDate,
		EstimatedEffort: dto.EstimatedEffort,
		Category:        category,
		Context:         contextVal,
		RelatedPeople:   relatedPeople,
		BumpCount:       0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Calculate initial priority score
	task.PriorityScore = s.priorityCalc.Calculate(task)

	// Save to database
	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, domain.NewInternalError("failed to create task", err)
	}

	// Log task creation in history
	if err := s.logHistory(ctx, userID, task.ID, domain.EventTaskCreated, nil, task); err != nil {
		// Log error but don't fail the request
		// In production, you might want to use a proper logger
	}

	return task, nil
}

// Get retrieves a task by ID
func (s *TaskService) Get(ctx context.Context, userID, taskID string) (*domain.Task, error) {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, domain.NewInternalError("failed to find task", err)
	}
	if task == nil {
		return nil, domain.NewNotFoundError("task", taskID)
	}

	// Verify task belongs to user
	if task.UserID != userID {
		return nil, domain.NewForbiddenError("task", "access")
	}

	// Populate priority breakdown for detailed view.
	// NOTE: We intentionally calculate breakdown for ALL tasks, including completed ones.
	// This provides historical context in the analytics dashboard and helps users understand
	// why a task was prioritized the way it was at completion time.
	_, task.PriorityBreakdown = s.priorityCalc.CalculateWithBreakdown(task)

	return task, nil
}

// List retrieves tasks with filters
func (s *TaskService) List(ctx context.Context, userID string, filter *domain.TaskListFilter) ([]*domain.Task, error) {
	return s.taskRepo.List(ctx, userID, filter)
}

// Update updates a task
func (s *TaskService) Update(ctx context.Context, userID, taskID string, dto *domain.UpdateTaskDTO) (*domain.Task, error) {
	// Get existing task
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, domain.NewInternalError("failed to find task", err)
	}
	if task == nil {
		return nil, domain.NewNotFoundError("task", taskID)
	}

	// Verify ownership
	if task.UserID != userID {
		return nil, domain.NewForbiddenError("task", "update")
	}

	// Store old task for history
	oldTask := *task

	// Apply updates with validation
	if dto.Title != nil {
		validated, err := validation.ValidateRequiredText(*dto.Title, 200, "title")
		if err != nil {
			return nil, err
		}
		task.Title = validated
	}
	if dto.Description != nil {
		validated, err := validation.ValidateOptionalText(dto.Description, 2000, "description")
		if err != nil {
			return nil, err
		}
		task.Description = validated
	}
	if dto.Status != nil {
		task.Status = *dto.Status
		if *dto.Status == domain.TaskStatusDone && task.CompletedAt == nil {
			now := time.Now()
			task.CompletedAt = &now
		}
	}
	if dto.UserPriority != nil {
		if err := validation.ValidatePriority(*dto.UserPriority); err != nil {
			return nil, err
		}
		task.UserPriority = *dto.UserPriority
	}
	if dto.DueDate != nil {
		task.DueDate = dto.DueDate
	}
	if dto.EstimatedEffort != nil {
		task.EstimatedEffort = dto.EstimatedEffort
	}
	if dto.Category != nil {
		validated, err := validation.ValidateCategory(dto.Category)
		if err != nil {
			return nil, err
		}
		task.Category = validated
	}
	if dto.Context != nil {
		validated, err := validation.ValidateOptionalText(dto.Context, 500, "context")
		if err != nil {
			return nil, err
		}
		task.Context = validated
	}
	if dto.RelatedPeople != nil {
		validated, err := validation.ValidateStringSlice(dto.RelatedPeople, 100, 20, "related_people")
		if err != nil {
			return nil, err
		}
		task.RelatedPeople = validated
	}

	task.UpdatedAt = time.Now()

	// Recalculate priority
	task.PriorityScore = s.priorityCalc.Calculate(task)

	// Save to database
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, domain.NewInternalError("failed to update task", err)
	}

	// Log update in history
	if err := s.logHistory(ctx, userID, task.ID, domain.EventTaskUpdated, &oldTask, task); err != nil {
		// Log error but don't fail the request
	}

	return task, nil
}

// Delete deletes a task
func (s *TaskService) Delete(ctx context.Context, userID, taskID string) error {
	// Verify task exists and belongs to user
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return domain.NewInternalError("failed to find task", err)
	}
	if task == nil {
		return domain.NewNotFoundError("task", taskID)
	}

	if task.UserID != userID {
		return domain.NewForbiddenError("task", "delete")
	}

	// Log deletion in history (before deleting)
	if err := s.logHistory(ctx, userID, taskID, domain.EventTaskDeleted, task, nil); err != nil {
		// Log error but continue with deletion
	}

	if err := s.taskRepo.Delete(ctx, taskID, userID); err != nil {
		return domain.NewInternalError("failed to delete task", err)
	}

	return nil
}

// Bump increments the bump counter for a task
func (s *TaskService) Bump(ctx context.Context, userID, taskID string) (*domain.Task, error) {
	// Get task
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, domain.NewInternalError("failed to find task", err)
	}
	if task == nil {
		return nil, domain.NewNotFoundError("task", taskID)
	}

	// Verify ownership
	if task.UserID != userID {
		return nil, domain.NewForbiddenError("task", "bump")
	}

	// Store old bump count
	oldBumpCount := task.BumpCount

	// Increment bump count
	if err := s.taskRepo.IncrementBumpCount(ctx, taskID, userID); err != nil {
		return nil, domain.NewInternalError("failed to increment bump count", err)
	}

	// Get updated task
	task, err = s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, domain.NewInternalError("failed to retrieve updated task", err)
	}

	// Recalculate priority
	task.PriorityScore = s.priorityCalc.Calculate(task)
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, domain.NewInternalError("failed to update task priority", err)
	}

	// Log bump in history
	oldValue := string(rune(oldBumpCount + '0'))
	newValue := string(rune(task.BumpCount + '0'))
	if err := s.logHistorySimple(ctx, userID, taskID, domain.EventTaskBumped, &oldValue, &newValue); err != nil {
		// Log error but don't fail the request
	}

	return task, nil
}

// Complete marks a task as complete
func (s *TaskService) Complete(ctx context.Context, userID, taskID string) (*domain.Task, error) {
	response, err := s.CompleteWithOptions(ctx, userID, taskID, nil)
	if err != nil {
		return nil, err
	}
	return response.CompletedTask, nil
}

// CompleteWithOptions marks a task as complete with optional recurrence handling
func (s *TaskService) CompleteWithOptions(ctx context.Context, userID, taskID string, req *domain.TaskCompletionRequest) (*domain.TaskCompletionResponse, error) {
	// Get task
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, domain.NewInternalError("failed to find task", err)
	}
	if task == nil {
		return nil, domain.NewNotFoundError("task", taskID)
	}

	// Verify ownership
	if task.UserID != userID {
		return nil, domain.NewForbiddenError("task", "complete")
	}

	// Block parent task completion if subtasks are incomplete (if subtask service is available)
	if s.subtaskService != nil {
		if err := s.subtaskService.ValidateParentCompletion(ctx, taskID); err != nil {
			return nil, err
		}
	}

	// Block task completion if it has incomplete blockers (if dependency service is available)
	if s.dependencyService != nil {
		if err := s.dependencyService.ValidateCompletion(ctx, taskID); err != nil {
			return nil, err
		}
	}

	// Update task
	task.Status = domain.TaskStatusDone
	now := time.Now()
	task.CompletedAt = &now
	task.UpdatedAt = now

	// Save to database
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, domain.NewInternalError("failed to update task", err)
	}

	// Log completion in history
	if err := s.logHistorySimple(ctx, userID, taskID, domain.EventTaskCompleted, nil, nil); err != nil {
		// Log error but don't fail the request
	}

	// Build response
	response := &domain.TaskCompletionResponse{
		CompletedTask: task,
	}

	// Handle recurring task generation if this is a recurring task and recurrence service is available
	if task.SeriesID != nil && s.recurrenceService != nil {
		nextTask, err := s.recurrenceService.GenerateNextTask(ctx, task, req)
		if err != nil {
			// Log error but don't fail the completion - the task is already completed
			// In production, use a proper logger
		} else if nextTask != nil {
			response.NextTask = nextTask
		}
	}

	return response, nil
}

// GetAtRiskTasks retrieves tasks that are at risk
func (s *TaskService) GetAtRiskTasks(ctx context.Context, userID string) ([]*domain.Task, error) {
	tasks, err := s.taskRepo.FindAtRiskTasks(ctx, userID)
	if err != nil {
		return nil, domain.NewInternalError("failed to retrieve at-risk tasks", err)
	}
	return tasks, nil
}

// GetCalendar retrieves tasks grouped by date for calendar view
func (s *TaskService) GetCalendar(ctx context.Context, userID string, filter *domain.CalendarFilter) (*domain.CalendarResponse, error) {
	// Fetch tasks in date range
	tasks, err := s.taskRepo.FindByDateRange(ctx, userID, filter)
	if err != nil {
		return nil, domain.NewInternalError("failed to retrieve calendar tasks", err)
	}

	// Group tasks by date
	dateMap := make(map[string]*domain.CalendarDayData)
	for _, task := range tasks {
		if task.DueDate == nil {
			continue
		}

		// Format date as YYYY-MM-DD
		dateKey := task.DueDate.Format("2006-01-02")

		// Initialize day data if not exists
		if _, exists := dateMap[dateKey]; !exists {
			dateMap[dateKey] = &domain.CalendarDayData{
				Count:      0,
				BadgeColor: "blue",
				Tasks:      []*domain.Task{},
			}
		}

		// Add task to day
		dateMap[dateKey].Tasks = append(dateMap[dateKey].Tasks, task)
		dateMap[dateKey].Count++
	}

	// Calculate badge color for each day based on highest priority
	for _, dayData := range dateMap {
		dayData.BadgeColor = s.calculateBadgeColor(dayData.Tasks)
	}

	return &domain.CalendarResponse{
		Dates: dateMap,
	}, nil
}

// calculateBadgeColor determines badge color based on highest priority task
func (s *TaskService) calculateBadgeColor(tasks []*domain.Task) string {
	if len(tasks) == 0 {
		return "blue"
	}

	maxPriority := 0
	for _, task := range tasks {
		if task.PriorityScore > maxPriority {
			maxPriority = task.PriorityScore
		}
	}

	// Color thresholds
	if maxPriority >= 70 {
		return "red"
	}
	if maxPriority >= 40 {
		return "yellow"
	}
	return "blue"
}

// logHistory creates a history entry with full task data
func (s *TaskService) logHistory(ctx context.Context, userID, taskID string, eventType domain.TaskHistoryEventType, oldTask, newTask *domain.Task) error {
	var oldValue, newValue *string

	if oldTask != nil {
		data, _ := json.Marshal(oldTask)
		str := string(data)
		oldValue = &str
	}

	if newTask != nil {
		data, _ := json.Marshal(newTask)
		str := string(data)
		newValue = &str
	}

	return s.logHistorySimple(ctx, userID, taskID, eventType, oldValue, newValue)
}

// logHistorySimple creates a history entry with simple values
func (s *TaskService) logHistorySimple(ctx context.Context, userID, taskID string, eventType domain.TaskHistoryEventType, oldValue, newValue *string) error {
	history := &domain.TaskHistory{
		ID:        uuid.New().String(),
		UserID:    userID,
		TaskID:    taskID,
		EventType: eventType,
		OldValue:  oldValue,
		NewValue:  newValue,
		CreatedAt: time.Now(),
	}

	return s.taskHistoryRepo.Create(ctx, history)
}

// RenameCategory renames a category for all tasks belonging to the user
// Returns the number of tasks updated
func (s *TaskService) RenameCategory(ctx context.Context, userID, oldName, newName string) (int, error) {
	// Validate old category name
	validatedOld, err := validation.ValidateCategory(&oldName)
	if err != nil {
		return 0, err
	}
	if validatedOld == nil {
		return 0, domain.NewValidationError("old_name", "is required")
	}

	// Validate new category name
	validatedNew, err := validation.ValidateCategory(&newName)
	if err != nil {
		return 0, err
	}
	if validatedNew == nil {
		return 0, domain.NewValidationError("new_name", "is required")
	}

	count, err := s.taskRepo.RenameCategoryForUser(ctx, userID, *validatedOld, *validatedNew)
	if err != nil {
		return 0, domain.NewInternalError("failed to rename category", err)
	}
	return count, nil
}

// DeleteCategory removes a category from all tasks belonging to the user
// Returns the number of tasks updated
func (s *TaskService) DeleteCategory(ctx context.Context, userID, categoryName string) (int, error) {
	// Validate category name
	validated, err := validation.ValidateCategory(&categoryName)
	if err != nil {
		return 0, err
	}
	if validated == nil {
		return 0, domain.NewValidationError("category_name", "is required")
	}

	count, err := s.taskRepo.DeleteCategoryForUser(ctx, userID, *validated)
	if err != nil {
		return 0, domain.NewInternalError("failed to delete category", err)
	}
	return count, nil
}

// BulkDelete permanently deletes multiple tasks by their IDs
// Returns a response with success count, failed IDs, and a message
func (s *TaskService) BulkDelete(ctx context.Context, userID string, taskIDs []string) (*domain.BulkOperationResponse, error) {
	if len(taskIDs) == 0 {
		return nil, domain.NewValidationError("task_ids", "must contain at least 1 item")
	}
	if len(taskIDs) > 100 {
		return nil, domain.NewValidationError("task_ids", "cannot contain more than 100 items")
	}

	successCount, failedIDs, err := s.taskRepo.BulkDelete(ctx, userID, taskIDs)
	if err != nil {
		return nil, domain.NewInternalError("failed to bulk delete tasks", err)
	}

	var message string
	if len(failedIDs) == 0 {
		message = "Successfully deleted " + strconv.Itoa(successCount) + " tasks"
	} else {
		message = "Deleted " + strconv.Itoa(successCount) + " tasks. " + strconv.Itoa(len(failedIDs)) + " task(s) not found or not owned."
	}

	return &domain.BulkOperationResponse{
		SuccessCount: successCount,
		FailedIDs:    failedIDs,
		Message:      message,
	}, nil
}

// BulkRestore restores multiple completed tasks to "todo" status
// Returns a response with success count, failed IDs, and a message
func (s *TaskService) BulkRestore(ctx context.Context, userID string, taskIDs []string) (*domain.BulkOperationResponse, error) {
	if len(taskIDs) == 0 {
		return nil, domain.NewValidationError("task_ids", "must contain at least 1 item")
	}
	if len(taskIDs) > 100 {
		return nil, domain.NewValidationError("task_ids", "cannot contain more than 100 items")
	}

	successCount, failedIDs, err := s.taskRepo.BulkUpdateStatus(ctx, userID, taskIDs, domain.TaskStatusTodo)
	if err != nil {
		return nil, domain.NewInternalError("failed to bulk restore tasks", err)
	}

	var message string
	if len(failedIDs) == 0 {
		message = "Successfully restored " + strconv.Itoa(successCount) + " tasks to active status"
	} else {
		message = "Restored " + strconv.Itoa(successCount) + " tasks. " + strconv.Itoa(len(failedIDs)) + " task(s) not found, not owned, or not completed."
	}

	return &domain.BulkOperationResponse{
		SuccessCount: successCount,
		FailedIDs:    failedIDs,
		Message:      message,
	}, nil
}
