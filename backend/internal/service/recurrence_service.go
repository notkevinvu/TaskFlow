package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/domain/priority"
	"github.com/notkevinvu/taskflow/backend/internal/ports"
)

// RecurrenceService handles business logic for recurring tasks
type RecurrenceService struct {
	taskRepo        ports.TaskRepository
	seriesRepo      ports.TaskSeriesRepository
	prefsRepo       ports.UserPreferencesRepository
	taskHistoryRepo ports.TaskHistoryRepository
	priorityCalc    *priority.Calculator
}

// NewRecurrenceService creates a new recurrence service
func NewRecurrenceService(
	taskRepo ports.TaskRepository,
	seriesRepo ports.TaskSeriesRepository,
	prefsRepo ports.UserPreferencesRepository,
	taskHistoryRepo ports.TaskHistoryRepository,
) *RecurrenceService {
	return &RecurrenceService{
		taskRepo:        taskRepo,
		seriesRepo:      seriesRepo,
		prefsRepo:       prefsRepo,
		taskHistoryRepo: taskHistoryRepo,
		priorityCalc:    priority.NewCalculator(),
	}
}

// CreateTaskWithRecurrence creates a task and optionally sets up recurrence
func (s *RecurrenceService) CreateTaskWithRecurrence(
	ctx context.Context,
	userID string,
	task *domain.Task,
	rule *domain.RecurrenceRule,
) (*domain.Task, *domain.TaskSeries, error) {
	// If no recurrence rule, just return the task as-is
	if rule == nil || !rule.Pattern.IsRecurring() {
		return task, nil, nil
	}

	// Validate recurrence rule
	if err := rule.Validate(); err != nil {
		return nil, nil, err
	}

	// Create task series
	now := time.Now()
	series := &domain.TaskSeries{
		ID:                 uuid.New().String(),
		UserID:             userID,
		OriginalTaskID:     task.ID,
		Pattern:            rule.Pattern,
		IntervalValue:      rule.IntervalValue,
		EndDate:            rule.EndDate,
		DueDateCalculation: rule.DueDateCalculation,
		IsActive:           true,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	// Save series
	if err := s.seriesRepo.Create(ctx, series); err != nil {
		return nil, nil, domain.NewInternalError("failed to create task series", err)
	}

	// Update task with series reference
	task.SeriesID = &series.ID

	// Update the task in the database
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return nil, nil, domain.NewInternalError("failed to update task with series", err)
	}

	return task, series, nil
}

// GenerateNextTask creates the next task in a recurring series
func (s *RecurrenceService) GenerateNextTask(
	ctx context.Context,
	completedTask *domain.Task,
	req *domain.TaskCompletionRequest,
) (*domain.Task, error) {
	if completedTask.SeriesID == nil {
		return nil, nil // Not a recurring task
	}

	// Check if user wants to stop recurrence
	if req != nil && req.StopRecurrence {
		if err := s.seriesRepo.Deactivate(ctx, *completedTask.SeriesID, completedTask.UserID); err != nil {
			return nil, domain.NewInternalError("failed to deactivate series", err)
		}
		return nil, nil
	}

	// Check if user wants to skip this occurrence
	if req != nil && req.SkipNextOccurrence {
		return nil, nil
	}

	// Get the series
	series, err := s.seriesRepo.FindByID(ctx, *completedTask.SeriesID)
	if err != nil {
		if errors.Is(err, domain.ErrSeriesNotFound) {
			return nil, nil // Series deleted, nothing to do
		}
		return nil, err
	}

	// Check if series can generate next task
	now := time.Now()
	if !series.CanGenerateNext(now) {
		return nil, nil
	}

	// Determine due date calculation mode
	dueDateCalc := series.DueDateCalculation
	if req != nil && req.DueDateCalculation != nil {
		dueDateCalc = *req.DueDateCalculation

		// Save preference if requested
		if req.SaveAsDefault {
			if err := s.saveDefaultPreference(ctx, completedTask.UserID, dueDateCalc); err != nil {
				// Log but don't fail the operation
				// In production, we'd use a proper logger
			}
		}
		if req.SaveForCategory && completedTask.Category != nil {
			if err := s.saveCategoryPreference(ctx, completedTask.UserID, *completedTask.Category, dueDateCalc); err != nil {
				// Log but don't fail the operation
			}
		}
	}

	// Calculate next due date
	var nextDueDate *time.Time
	if completedTask.DueDate != nil {
		var baseDate time.Time
		if dueDateCalc == domain.DueDateFromCompletion {
			baseDate = now
		} else {
			baseDate = *completedTask.DueDate
		}
		next := series.CalculateNextDueDate(baseDate)
		nextDueDate = &next
	}

	// Check if next due date exceeds end date
	if series.EndDate != nil && nextDueDate != nil && nextDueDate.After(*series.EndDate) {
		// Series has ended
		if err := s.seriesRepo.Deactivate(ctx, series.ID, completedTask.UserID); err != nil {
			return nil, domain.NewInternalError("failed to deactivate ended series", err)
		}
		return nil, nil
	}

	// Create next task (inherits all properties except bump count)
	nextTask := &domain.Task{
		ID:              uuid.New().String(),
		UserID:          completedTask.UserID,
		Title:           completedTask.Title,
		Description:     completedTask.Description,
		Status:          domain.TaskStatusTodo,
		UserPriority:    completedTask.UserPriority,
		DueDate:         nextDueDate,
		EstimatedEffort: completedTask.EstimatedEffort,
		Category:        completedTask.Category,
		Context:         completedTask.Context,
		RelatedPeople:   completedTask.RelatedPeople,
		BumpCount:       0, // Reset bump count for new instance
		SeriesID:        completedTask.SeriesID,
		ParentTaskID:    &completedTask.ID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Calculate priority score for new task
	nextTask.PriorityScore = s.priorityCalc.Calculate(nextTask)

	// Save new task
	if err := s.taskRepo.Create(ctx, nextTask); err != nil {
		return nil, domain.NewInternalError("failed to create next recurring task", err)
	}

	// Log task creation in history
	s.taskHistoryRepo.Create(ctx, &domain.TaskHistory{
		ID:        uuid.New().String(),
		UserID:    nextTask.UserID,
		TaskID:    nextTask.ID,
		EventType: domain.EventTaskCreated,
		NewValue:  strPtr("Auto-generated from recurring series"),
		CreatedAt: now,
	})

	return nextTask, nil
}

// GetSeriesHistory retrieves the history of a task series
func (s *RecurrenceService) GetSeriesHistory(ctx context.Context, userID, seriesID string) (*domain.SeriesHistory, error) {
	// Get series
	series, err := s.seriesRepo.FindByID(ctx, seriesID)
	if err != nil {
		return nil, err
	}

	// Verify ownership
	if series.UserID != userID {
		return nil, domain.ErrUnauthorized
	}

	// Get all tasks in series
	tasks, err := s.seriesRepo.GetTasksBySeriesID(ctx, seriesID)
	if err != nil {
		return nil, err
	}

	// Convert to history entries
	entries := make([]domain.SeriesHistoryEntry, len(tasks))
	for i, task := range tasks {
		entries[i] = domain.SeriesHistoryEntry{
			TaskID:      task.ID,
			Title:       task.Title,
			Status:      task.Status,
			DueDate:     task.DueDate,
			CompletedAt: task.CompletedAt,
			CreatedAt:   task.CreatedAt,
		}
	}

	return &domain.SeriesHistory{
		Series: series,
		Tasks:  entries,
		Total:  len(entries),
	}, nil
}

// GetUserSeries retrieves all task series for a user
func (s *RecurrenceService) GetUserSeries(ctx context.Context, userID string, activeOnly bool) ([]*domain.TaskSeries, error) {
	if activeOnly {
		return s.seriesRepo.FindActiveByUserID(ctx, userID)
	}
	return s.seriesRepo.FindByUserID(ctx, userID)
}

// UpdateSeries updates a task series settings
func (s *RecurrenceService) UpdateSeries(ctx context.Context, userID, seriesID string, dto *domain.UpdateTaskSeriesDTO) (*domain.TaskSeries, error) {
	series, err := s.seriesRepo.FindByID(ctx, seriesID)
	if err != nil {
		return nil, err
	}

	if series.UserID != userID {
		return nil, domain.ErrUnauthorized
	}

	// Apply updates
	if dto.Pattern != nil {
		if err := dto.Pattern.Validate(); err != nil {
			return nil, err
		}
		series.Pattern = *dto.Pattern
	}
	if dto.IntervalValue != nil {
		if *dto.IntervalValue < 1 || *dto.IntervalValue > 365 {
			return nil, domain.ErrInvalidInterval
		}
		series.IntervalValue = *dto.IntervalValue
	}
	if dto.EndDate != nil {
		series.EndDate = dto.EndDate
	}
	if dto.DueDateCalculation != nil {
		if err := dto.DueDateCalculation.Validate(); err != nil {
			return nil, err
		}
		series.DueDateCalculation = *dto.DueDateCalculation
	}
	if dto.IsActive != nil {
		series.IsActive = *dto.IsActive
	}

	series.UpdatedAt = time.Now()

	if err := s.seriesRepo.Update(ctx, series); err != nil {
		return nil, domain.NewInternalError("failed to update series", err)
	}

	return series, nil
}

// DeactivateSeries stops a recurring series
func (s *RecurrenceService) DeactivateSeries(ctx context.Context, userID, seriesID string) error {
	series, err := s.seriesRepo.FindByID(ctx, seriesID)
	if err != nil {
		return err
	}

	if series.UserID != userID {
		return domain.ErrUnauthorized
	}

	return s.seriesRepo.Deactivate(ctx, seriesID, userID)
}

// GetEffectiveDueDateCalculation returns the effective due date calculation mode
// based on user preferences (category preference > user default > system default)
func (s *RecurrenceService) GetEffectiveDueDateCalculation(ctx context.Context, userID string, category *string) domain.DueDateCalculation {
	// Try category preference first
	if category != nil {
		catPref, err := s.prefsRepo.GetCategoryPreference(ctx, userID, *category)
		if err == nil && catPref != nil {
			return catPref.DueDateCalculation
		}
	}

	// Try user default
	userPrefs, err := s.prefsRepo.GetUserPreferences(ctx, userID)
	if err == nil && userPrefs != nil {
		return userPrefs.DefaultDueDateCalculation
	}

	// System default
	return domain.DueDateFromOriginal
}

// GetUserPreferences retrieves all preferences for a user
func (s *RecurrenceService) GetUserPreferences(ctx context.Context, userID string) (*domain.AllPreferences, error) {
	return s.prefsRepo.GetAllPreferences(ctx, userID)
}

// SetDefaultPreference sets the user's default due date calculation mode
func (s *RecurrenceService) SetDefaultPreference(ctx context.Context, userID string, calc domain.DueDateCalculation) error {
	return s.saveDefaultPreference(ctx, userID, calc)
}

// SetCategoryPreference sets the due date calculation mode for a category
func (s *RecurrenceService) SetCategoryPreference(ctx context.Context, userID, category string, calc domain.DueDateCalculation) error {
	return s.saveCategoryPreference(ctx, userID, category, calc)
}

// DeleteCategoryPreference removes a category preference
func (s *RecurrenceService) DeleteCategoryPreference(ctx context.Context, userID, category string) error {
	return s.prefsRepo.DeleteCategoryPreference(ctx, userID, category)
}

// Helper: Save default preference
func (s *RecurrenceService) saveDefaultPreference(ctx context.Context, userID string, calc domain.DueDateCalculation) error {
	prefs := &domain.UserPreferences{
		UserID:                    userID,
		DefaultDueDateCalculation: calc,
		CreatedAt:                 time.Now(),
		UpdatedAt:                 time.Now(),
	}
	return s.prefsRepo.UpsertUserPreferences(ctx, prefs)
}

// Helper: Save category preference
func (s *RecurrenceService) saveCategoryPreference(ctx context.Context, userID, category string, calc domain.DueDateCalculation) error {
	pref := &domain.CategoryPreference{
		UserID:             userID,
		Category:           category,
		DueDateCalculation: calc,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
	return s.prefsRepo.UpsertCategoryPreference(ctx, pref)
}

// Helper: string pointer
func strPtr(s string) *string {
	return &s
}
