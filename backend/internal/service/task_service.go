package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/domain/priority"
	"github.com/notkevinvu/taskflow/backend/internal/repository"
)

// TaskService handles task business logic
type TaskService struct {
	taskRepo        *repository.TaskRepository
	taskHistoryRepo *repository.TaskHistoryRepository
	priorityCalc    *priority.Calculator
}

// NewTaskService creates a new task service
func NewTaskService(taskRepo *repository.TaskRepository, taskHistoryRepo *repository.TaskHistoryRepository) *TaskService {
	return &TaskService{
		taskRepo:        taskRepo,
		taskHistoryRepo: taskHistoryRepo,
		priorityCalc:    priority.NewCalculator(),
	}
}

// Create creates a new task
func (s *TaskService) Create(ctx context.Context, userID string, dto *domain.CreateTaskDTO) (*domain.Task, error) {
	now := time.Now()

	// Set default user priority if not provided (1-10 scale)
	userPriority := 5
	if dto.UserPriority != nil {
		userPriority = *dto.UserPriority
	}

	// Create task
	task := &domain.Task{
		ID:              uuid.New().String(),
		UserID:          userID,
		Title:           dto.Title,
		Description:     dto.Description,
		Status:          domain.TaskStatusTodo,
		UserPriority:    userPriority,
		DueDate:         dto.DueDate,
		EstimatedEffort: dto.EstimatedEffort,
		Category:        dto.Category,
		Context:         dto.Context,
		RelatedPeople:   dto.RelatedPeople,
		BumpCount:       0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Calculate initial priority score
	task.PriorityScore = s.priorityCalc.Calculate(task)

	// Save to database
	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, err
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
		return nil, err
	}

	// Verify task belongs to user
	if task.UserID != userID {
		return nil, domain.ErrUnauthorized
	}

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
		return nil, err
	}

	// Verify ownership
	if task.UserID != userID {
		return nil, domain.ErrUnauthorized
	}

	// Store old task for history
	oldTask := *task

	// Apply updates
	if dto.Title != nil {
		task.Title = *dto.Title
	}
	if dto.Description != nil {
		task.Description = dto.Description
	}
	if dto.Status != nil {
		task.Status = *dto.Status
		if *dto.Status == domain.TaskStatusDone && task.CompletedAt == nil {
			now := time.Now()
			task.CompletedAt = &now
		}
	}
	if dto.UserPriority != nil {
		task.UserPriority = *dto.UserPriority
	}
	if dto.DueDate != nil {
		task.DueDate = dto.DueDate
	}
	if dto.EstimatedEffort != nil {
		task.EstimatedEffort = dto.EstimatedEffort
	}
	if dto.Category != nil {
		task.Category = dto.Category
	}
	if dto.Context != nil {
		task.Context = dto.Context
	}
	if dto.RelatedPeople != nil {
		task.RelatedPeople = dto.RelatedPeople
	}

	task.UpdatedAt = time.Now()

	// Recalculate priority
	task.PriorityScore = s.priorityCalc.Calculate(task)

	// Save to database
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, err
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
		return err
	}

	if task.UserID != userID {
		return domain.ErrUnauthorized
	}

	// Log deletion in history (before deleting)
	if err := s.logHistory(ctx, userID, taskID, domain.EventTaskDeleted, task, nil); err != nil {
		// Log error but continue with deletion
	}

	return s.taskRepo.Delete(ctx, taskID, userID)
}

// Bump increments the bump counter for a task
func (s *TaskService) Bump(ctx context.Context, userID, taskID string) (*domain.Task, error) {
	// Get task
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if task.UserID != userID {
		return nil, domain.ErrUnauthorized
	}

	// Store old bump count
	oldBumpCount := task.BumpCount

	// Increment bump count
	if err := s.taskRepo.IncrementBumpCount(ctx, taskID, userID); err != nil {
		return nil, err
	}

	// Get updated task
	task, err = s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Recalculate priority
	task.PriorityScore = s.priorityCalc.Calculate(task)
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, err
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
	// Get task
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if task.UserID != userID {
		return nil, domain.ErrUnauthorized
	}

	// Update task
	task.Status = domain.TaskStatusDone
	now := time.Now()
	task.CompletedAt = &now
	task.UpdatedAt = now

	// Save to database
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, err
	}

	// Log completion in history
	if err := s.logHistorySimple(ctx, userID, taskID, domain.EventTaskCompleted, nil, nil); err != nil {
		// Log error but don't fail the request
	}

	return task, nil
}

// GetAtRiskTasks retrieves tasks that are at risk
func (s *TaskService) GetAtRiskTasks(ctx context.Context, userID string) ([]*domain.Task, error) {
	return s.taskRepo.FindAtRiskTasks(ctx, userID)
}

// GetCalendar retrieves tasks grouped by date for calendar view
func (s *TaskService) GetCalendar(ctx context.Context, userID string, filter *domain.CalendarFilter) (*domain.CalendarResponse, error) {
	// Fetch tasks in date range
	tasks, err := s.taskRepo.FindByDateRange(ctx, userID, filter)
	if err != nil {
		return nil, err
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
	return s.taskRepo.RenameCategoryForUser(ctx, userID, oldName, newName)
}

// DeleteCategory removes a category from all tasks belonging to the user
// Returns the number of tasks updated
func (s *TaskService) DeleteCategory(ctx context.Context, userID, categoryName string) (int, error) {
	return s.taskRepo.DeleteCategoryForUser(ctx, userID, categoryName)
}
