package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/domain/priority"
	"github.com/notkevinvu/taskflow/backend/internal/ports"
	"github.com/notkevinvu/taskflow/backend/internal/validation"
)

// SubtaskService handles subtask-specific business logic
type SubtaskService struct {
	taskRepo        ports.TaskRepository
	taskHistoryRepo ports.TaskHistoryRepository
	priorityCalc    *priority.Calculator
}

// NewSubtaskService creates a new subtask service
func NewSubtaskService(taskRepo ports.TaskRepository, taskHistoryRepo ports.TaskHistoryRepository) *SubtaskService {
	return &SubtaskService{
		taskRepo:        taskRepo,
		taskHistoryRepo: taskHistoryRepo,
		priorityCalc:    priority.NewCalculator(),
	}
}

// Create creates a new subtask under a parent task
// Validates single-level nesting and inherits parent's category
func (s *SubtaskService) Create(ctx context.Context, userID string, dto *domain.CreateSubtaskDTO) (*domain.Task, error) {
	// Validate parent task exists and user owns it
	parentTask, err := s.taskRepo.FindByID(ctx, dto.ParentTaskID)
	if err != nil {
		return nil, domain.NewInternalError("failed to find parent task", err)
	}
	if parentTask == nil {
		return nil, domain.ErrParentNotFound
	}
	if parentTask.UserID != userID {
		return nil, domain.NewForbiddenError("parent task", "access")
	}

	// Enforce single-level nesting: parent cannot be a subtask
	if !parentTask.CanHaveSubtasks() {
		return nil, domain.ErrSubtaskDepthExceeded
	}

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

	// Validate optional context
	contextVal, err := validation.ValidateOptionalText(dto.Context, 500, "context")
	if err != nil {
		return nil, err
	}

	now := time.Now()

	// Set default user priority if not provided
	userPriority := 5
	if dto.UserPriority != nil {
		if err := validation.ValidatePriority(*dto.UserPriority); err != nil {
			return nil, err
		}
		userPriority = *dto.UserPriority
	}

	// Parse due date if provided
	var dueDate *time.Time
	if dto.DueDate != nil {
		parsed, err := domain.ParseDate(*dto.DueDate)
		if err != nil {
			return nil, domain.NewValidationError("due_date", "must be in YYYY-MM-DD format")
		}
		dueDate = &parsed
	}

	// Create subtask - inherits category from parent
	subtask := &domain.Task{
		ID:              uuid.New().String(),
		UserID:          userID,
		Title:           title,
		Description:     description,
		Status:          domain.TaskStatusTodo,
		TaskType:        domain.TaskTypeSubtask,
		UserPriority:    userPriority,
		DueDate:         dueDate,
		EstimatedEffort: dto.EstimatedEffort,
		Category:        parentTask.Category, // Inherit parent's category
		Context:         contextVal,
		ParentTaskID:    &dto.ParentTaskID, // Link to parent
		BumpCount:       0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Calculate priority with parent boost (15% of parent's priority score)
	subtask.PriorityScore = s.priorityCalc.CalculateForSubtask(subtask, parentTask.PriorityScore)

	// Save to database
	if err := s.taskRepo.Create(ctx, subtask); err != nil {
		return nil, domain.NewInternalError("failed to create subtask", err)
	}

	// Log subtask creation in history
	if err := s.logHistorySimple(ctx, userID, subtask.ID, domain.EventTaskCreated, nil, nil); err != nil {
		// Log error but don't fail the request
	}

	return subtask, nil
}

// GetSubtasks retrieves all subtasks for a parent task
func (s *SubtaskService) GetSubtasks(ctx context.Context, userID, parentTaskID string) ([]*domain.Task, error) {
	// Verify parent task exists and user owns it
	parentTask, err := s.taskRepo.FindByID(ctx, parentTaskID)
	if err != nil {
		return nil, domain.NewInternalError("failed to find parent task", err)
	}
	if parentTask == nil {
		return nil, domain.ErrParentNotFound
	}
	if parentTask.UserID != userID {
		return nil, domain.NewForbiddenError("parent task", "access")
	}

	subtasks, err := s.taskRepo.GetSubtasks(ctx, parentTaskID)
	if err != nil {
		return nil, domain.NewInternalError("failed to retrieve subtasks", err)
	}

	return subtasks, nil
}

// GetSubtaskInfo retrieves aggregated subtask statistics for a parent task
func (s *SubtaskService) GetSubtaskInfo(ctx context.Context, userID, parentTaskID string) (*domain.SubtaskInfo, error) {
	// Verify parent task exists and user owns it
	parentTask, err := s.taskRepo.FindByID(ctx, parentTaskID)
	if err != nil {
		return nil, domain.NewInternalError("failed to find parent task", err)
	}
	if parentTask == nil {
		return nil, domain.ErrParentNotFound
	}
	if parentTask.UserID != userID {
		return nil, domain.NewForbiddenError("parent task", "access")
	}

	info, err := s.taskRepo.GetSubtaskInfo(ctx, parentTaskID)
	if err != nil {
		return nil, domain.NewInternalError("failed to retrieve subtask info", err)
	}

	return info, nil
}

// GetTaskWithSubtasks retrieves a parent task with its subtask info and optionally expanded subtasks
func (s *SubtaskService) GetTaskWithSubtasks(ctx context.Context, userID, taskID string, expandSubtasks bool) (*domain.TaskWithSubtasks, error) {
	// Get the task
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, domain.NewInternalError("failed to find task", err)
	}
	if task == nil {
		return nil, domain.NewNotFoundError("task", taskID)
	}
	if task.UserID != userID {
		return nil, domain.NewForbiddenError("task", "access")
	}

	result := &domain.TaskWithSubtasks{
		Task: task,
	}

	// Only regular tasks can have subtasks
	if task.TaskType == domain.TaskTypeRegular {
		// Get subtask info
		info, err := s.taskRepo.GetSubtaskInfo(ctx, taskID)
		if err != nil {
			return nil, domain.NewInternalError("failed to retrieve subtask info", err)
		}
		result.SubtaskInfo = info

		// Optionally expand subtasks
		if expandSubtasks && info.TotalCount > 0 {
			subtasks, err := s.taskRepo.GetSubtasks(ctx, taskID)
			if err != nil {
				return nil, domain.NewInternalError("failed to retrieve subtasks", err)
			}
			result.Subtasks = subtasks
		}
	}

	return result, nil
}

// CompleteSubtask marks a subtask as complete and checks if all subtasks are done
// Returns special response to prompt user for parent completion if this was the last subtask
func (s *SubtaskService) CompleteSubtask(ctx context.Context, userID, subtaskID string) (*domain.SubtaskCompletionResponse, error) {
	// Get the subtask
	subtask, err := s.taskRepo.FindByID(ctx, subtaskID)
	if err != nil {
		return nil, domain.NewInternalError("failed to find subtask", err)
	}
	if subtask == nil {
		return nil, domain.NewNotFoundError("subtask", subtaskID)
	}
	if subtask.UserID != userID {
		return nil, domain.NewForbiddenError("subtask", "complete")
	}

	// Verify this is actually a subtask
	if !subtask.IsSubtask() {
		return nil, domain.NewValidationError("task", "is not a subtask")
	}

	// Complete the subtask
	subtask.Status = domain.TaskStatusDone
	now := time.Now()
	subtask.CompletedAt = &now
	subtask.UpdatedAt = now

	if err := s.taskRepo.Update(ctx, subtask); err != nil {
		return nil, domain.NewInternalError("failed to update subtask", err)
	}

	// Log completion
	if err := s.logHistorySimple(ctx, userID, subtaskID, domain.EventTaskCompleted, nil, nil); err != nil {
		// Log error but don't fail the request
	}

	// Check if all subtasks are now complete
	incompleteCount, err := s.taskRepo.CountIncompleteSubtasks(ctx, *subtask.ParentTaskID)
	if err != nil {
		return nil, domain.NewInternalError("failed to count incomplete subtasks", err)
	}

	allComplete := incompleteCount == 0

	response := &domain.SubtaskCompletionResponse{
		CompletedTask:       subtask,
		AllSubtasksComplete: allComplete,
	}

	// If all subtasks are complete, include parent task info for UI prompt
	if allComplete {
		parentTask, err := s.taskRepo.FindByID(ctx, *subtask.ParentTaskID)
		if err == nil && parentTask != nil {
			response.ParentTask = parentTask
			response.Message = "All subtasks complete! Would you like to mark the parent task as done?"
		}
	}

	return response, nil
}

// CanCompleteParent checks if a parent task can be completed (all subtasks done)
func (s *SubtaskService) CanCompleteParent(ctx context.Context, userID, parentTaskID string) (bool, error) {
	// Verify parent task exists and user owns it
	parentTask, err := s.taskRepo.FindByID(ctx, parentTaskID)
	if err != nil {
		return false, domain.NewInternalError("failed to find parent task", err)
	}
	if parentTask == nil {
		return false, domain.ErrParentNotFound
	}
	if parentTask.UserID != userID {
		return false, domain.NewForbiddenError("parent task", "access")
	}

	// Check for incomplete subtasks
	incompleteCount, err := s.taskRepo.CountIncompleteSubtasks(ctx, parentTaskID)
	if err != nil {
		return false, domain.NewInternalError("failed to count incomplete subtasks", err)
	}

	return incompleteCount == 0, nil
}

// ValidateParentCompletion validates that a task can be completed
// Called by TaskService before completing a task with subtasks
func (s *SubtaskService) ValidateParentCompletion(ctx context.Context, taskID string) error {
	// Get the task
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return domain.NewInternalError("failed to find task", err)
	}
	if task == nil {
		return nil // Not found is OK - TaskService handles this
	}

	// Only check regular tasks (they can have subtasks)
	if task.TaskType != domain.TaskTypeRegular {
		return nil
	}

	// Check for incomplete subtasks
	incompleteCount, err := s.taskRepo.CountIncompleteSubtasks(ctx, taskID)
	if err != nil {
		return domain.NewInternalError("failed to check subtask status", err)
	}

	if incompleteCount > 0 {
		return domain.ErrCannotCompleteParent
	}

	return nil
}

// logHistorySimple creates a history entry with simple values
func (s *SubtaskService) logHistorySimple(ctx context.Context, userID, taskID string, eventType domain.TaskHistoryEventType, oldValue, newValue *string) error {
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
