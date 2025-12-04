package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Mock Repositories for Recurrence Tests
// =============================================================================

type MockTaskSeriesRepository struct {
	mock.Mock
}

func (m *MockTaskSeriesRepository) Create(ctx context.Context, series *domain.TaskSeries) error {
	args := m.Called(ctx, series)
	return args.Error(0)
}

func (m *MockTaskSeriesRepository) FindByID(ctx context.Context, id string) (*domain.TaskSeries, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TaskSeries), args.Error(1)
}

func (m *MockTaskSeriesRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.TaskSeries, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.TaskSeries), args.Error(1)
}

func (m *MockTaskSeriesRepository) FindActiveByUserID(ctx context.Context, userID string) ([]*domain.TaskSeries, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.TaskSeries), args.Error(1)
}

func (m *MockTaskSeriesRepository) Update(ctx context.Context, series *domain.TaskSeries) error {
	args := m.Called(ctx, series)
	return args.Error(0)
}

func (m *MockTaskSeriesRepository) Deactivate(ctx context.Context, id, userID string) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockTaskSeriesRepository) Delete(ctx context.Context, id, userID string) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockTaskSeriesRepository) CountTasksInSeries(ctx context.Context, seriesID string) (int, error) {
	args := m.Called(ctx, seriesID)
	return args.Int(0), args.Error(1)
}

func (m *MockTaskSeriesRepository) GetTasksBySeriesID(ctx context.Context, seriesID string) ([]*domain.Task, error) {
	args := m.Called(ctx, seriesID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Task), args.Error(1)
}

type MockUserPreferencesRepository struct {
	mock.Mock
}

func (m *MockUserPreferencesRepository) GetUserPreferences(ctx context.Context, userID string) (*domain.UserPreferences, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserPreferences), args.Error(1)
}

func (m *MockUserPreferencesRepository) UpsertUserPreferences(ctx context.Context, prefs *domain.UserPreferences) error {
	args := m.Called(ctx, prefs)
	return args.Error(0)
}

func (m *MockUserPreferencesRepository) DeleteUserPreferences(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserPreferencesRepository) GetCategoryPreference(ctx context.Context, userID, category string) (*domain.CategoryPreference, error) {
	args := m.Called(ctx, userID, category)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CategoryPreference), args.Error(1)
}

func (m *MockUserPreferencesRepository) GetCategoryPreferencesByUserID(ctx context.Context, userID string) ([]domain.CategoryPreference, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.CategoryPreference), args.Error(1)
}

func (m *MockUserPreferencesRepository) UpsertCategoryPreference(ctx context.Context, pref *domain.CategoryPreference) error {
	args := m.Called(ctx, pref)
	return args.Error(0)
}

func (m *MockUserPreferencesRepository) DeleteCategoryPreference(ctx context.Context, userID, category string) error {
	args := m.Called(ctx, userID, category)
	return args.Error(0)
}

func (m *MockUserPreferencesRepository) DeleteAllCategoryPreferences(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserPreferencesRepository) GetAllPreferences(ctx context.Context, userID string) (*domain.AllPreferences, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AllPreferences), args.Error(1)
}

// =============================================================================
// Test Helpers
// =============================================================================

func newRecurrenceService(
	taskRepo *MockTaskRepository,
	seriesRepo *MockTaskSeriesRepository,
	prefsRepo *MockUserPreferencesRepository,
	historyRepo *MockTaskHistoryRepository,
) *RecurrenceService {
	return NewRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)
}

func createRecurringTask(userID, taskID, seriesID string) *domain.Task {
	now := time.Now()
	dueDate := now.Add(24 * time.Hour)
	category := "Work"
	return &domain.Task{
		ID:           taskID,
		UserID:       userID,
		Title:        "Recurring Task",
		Status:       domain.TaskStatusTodo,
		UserPriority: 5,
		DueDate:      &dueDate,
		Category:     &category,
		SeriesID:     &seriesID,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

func createTaskSeries(userID, seriesID, originalTaskID string, pattern domain.RecurrencePattern) *domain.TaskSeries {
	now := time.Now()
	return &domain.TaskSeries{
		ID:                 seriesID,
		UserID:             userID,
		OriginalTaskID:     originalTaskID,
		Pattern:            pattern,
		IntervalValue:      1,
		DueDateCalculation: domain.DueDateFromOriginal,
		IsActive:           true,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
}

// =============================================================================
// CreateTaskWithRecurrence Tests
// =============================================================================

func TestRecurrenceService_CreateTaskWithRecurrence_NoRuleReturnsTask(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	task := createTestTask("user-123", "task-1")

	// No recurrence rule
	resultTask, resultSeries, err := service.CreateTaskWithRecurrence(context.Background(), "user-123", task, nil)

	require.NoError(t, err)
	assert.Equal(t, task, resultTask)
	assert.Nil(t, resultSeries)
}

func TestRecurrenceService_CreateTaskWithRecurrence_NonePatternReturnsTask(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	task := createTestTask("user-123", "task-1")
	rule := &domain.RecurrenceRule{
		Pattern:            domain.RecurrencePatternNone,
		IntervalValue:      1,
		DueDateCalculation: domain.DueDateFromOriginal,
	}

	resultTask, resultSeries, err := service.CreateTaskWithRecurrence(context.Background(), "user-123", task, rule)

	require.NoError(t, err)
	assert.Equal(t, task, resultTask)
	assert.Nil(t, resultSeries)
}

func TestRecurrenceService_CreateTaskWithRecurrence_CreatesSeriesSuccessfully(t *testing.T) {
	tests := []struct {
		name    string
		pattern domain.RecurrencePattern
	}{
		{"daily", domain.RecurrencePatternDaily},
		{"weekly", domain.RecurrencePatternWeekly},
		{"monthly", domain.RecurrencePatternMonthly},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			taskRepo := new(MockTaskRepository)
			seriesRepo := new(MockTaskSeriesRepository)
			prefsRepo := new(MockUserPreferencesRepository)
			historyRepo := new(MockTaskHistoryRepository)
			service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

			task := createTestTask("user-123", "task-1")
			rule := &domain.RecurrenceRule{
				Pattern:            tt.pattern,
				IntervalValue:      1,
				DueDateCalculation: domain.DueDateFromOriginal,
			}

			seriesRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.TaskSeries")).Return(nil)
			taskRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)

			resultTask, resultSeries, err := service.CreateTaskWithRecurrence(context.Background(), "user-123", task, rule)

			require.NoError(t, err)
			assert.NotNil(t, resultTask)
			assert.NotNil(t, resultSeries)
			assert.Equal(t, tt.pattern, resultSeries.Pattern)
			assert.NotNil(t, resultTask.SeriesID)
			seriesRepo.AssertExpectations(t)
			taskRepo.AssertExpectations(t)
		})
	}
}

func TestRecurrenceService_CreateTaskWithRecurrence_InvalidRuleReturnsError(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	task := createTestTask("user-123", "task-1")
	rule := &domain.RecurrenceRule{
		Pattern:            domain.RecurrencePatternDaily,
		IntervalValue:      0, // Invalid: must be >= 1
		DueDateCalculation: domain.DueDateFromOriginal,
	}

	_, _, err := service.CreateTaskWithRecurrence(context.Background(), "user-123", task, rule)

	assert.Error(t, err)
}

func TestRecurrenceService_CreateTaskWithRecurrence_SeriesCreationErrorReturnsError(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	task := createTestTask("user-123", "task-1")
	rule := &domain.RecurrenceRule{
		Pattern:            domain.RecurrencePatternDaily,
		IntervalValue:      1,
		DueDateCalculation: domain.DueDateFromOriginal,
	}

	seriesRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.TaskSeries")).Return(errors.New("db error"))

	_, _, err := service.CreateTaskWithRecurrence(context.Background(), "user-123", task, rule)

	assert.Error(t, err)
	seriesRepo.AssertExpectations(t)
}

// =============================================================================
// GenerateNextTask Tests
// =============================================================================

func TestRecurrenceService_GenerateNextTask_NonRecurringTaskReturnsNil(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	task := createTestTask("user-123", "task-1")
	task.SeriesID = nil // Not recurring

	nextTask, err := service.GenerateNextTask(context.Background(), task, nil)

	require.NoError(t, err)
	assert.Nil(t, nextTask)
}

func TestRecurrenceService_GenerateNextTask_StopRecurrenceDeactivatesSeries(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	seriesID := "series-1"
	task := createRecurringTask("user-123", "task-1", seriesID)

	req := &domain.TaskCompletionRequest{
		StopRecurrence: true,
	}

	seriesRepo.On("Deactivate", mock.Anything, seriesID, "user-123").Return(nil)

	nextTask, err := service.GenerateNextTask(context.Background(), task, req)

	require.NoError(t, err)
	assert.Nil(t, nextTask)
	seriesRepo.AssertExpectations(t)
}

func TestRecurrenceService_GenerateNextTask_SkipOccurrenceReturnsNil(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	seriesID := "series-1"
	task := createRecurringTask("user-123", "task-1", seriesID)

	req := &domain.TaskCompletionRequest{
		SkipNextOccurrence: true,
	}

	nextTask, err := service.GenerateNextTask(context.Background(), task, req)

	require.NoError(t, err)
	assert.Nil(t, nextTask)
}

func TestRecurrenceService_GenerateNextTask_CreatesNextTaskSuccessfully(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	seriesID := "series-1"
	task := createRecurringTask("user-123", "task-1", seriesID)
	series := createTaskSeries("user-123", seriesID, "task-1", domain.RecurrencePatternDaily)

	seriesRepo.On("FindByID", mock.Anything, seriesID).Return(series, nil)
	taskRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)
	historyRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.TaskHistory")).Return(nil)

	nextTask, err := service.GenerateNextTask(context.Background(), task, nil)

	require.NoError(t, err)
	require.NotNil(t, nextTask)
	assert.Equal(t, task.Title, nextTask.Title)
	assert.Equal(t, task.UserID, nextTask.UserID)
	assert.Equal(t, domain.TaskStatusTodo, nextTask.Status)
	assert.Equal(t, 0, nextTask.BumpCount) // Reset bump count
	assert.Equal(t, &task.ID, nextTask.ParentTaskID)
	seriesRepo.AssertExpectations(t)
	taskRepo.AssertExpectations(t)
}

func TestRecurrenceService_GenerateNextTask_CalculatesDueDateFromOriginal(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	seriesID := "series-1"
	originalDueDate := time.Now().Add(24 * time.Hour).Truncate(time.Second)
	task := createRecurringTask("user-123", "task-1", seriesID)
	task.DueDate = &originalDueDate

	series := createTaskSeries("user-123", seriesID, "task-1", domain.RecurrencePatternDaily)
	series.DueDateCalculation = domain.DueDateFromOriginal

	seriesRepo.On("FindByID", mock.Anything, seriesID).Return(series, nil)
	taskRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)
	historyRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.TaskHistory")).Return(nil)

	nextTask, err := service.GenerateNextTask(context.Background(), task, nil)

	require.NoError(t, err)
	require.NotNil(t, nextTask.DueDate)
	// Next due date should be original + 1 day
	expectedDueDate := originalDueDate.AddDate(0, 0, 1)
	assert.Equal(t, expectedDueDate.Day(), nextTask.DueDate.Day())
}

func TestRecurrenceService_GenerateNextTask_InactiveSeriesReturnsNil(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	seriesID := "series-1"
	task := createRecurringTask("user-123", "task-1", seriesID)
	series := createTaskSeries("user-123", seriesID, "task-1", domain.RecurrencePatternDaily)
	series.IsActive = false

	seriesRepo.On("FindByID", mock.Anything, seriesID).Return(series, nil)

	nextTask, err := service.GenerateNextTask(context.Background(), task, nil)

	require.NoError(t, err)
	assert.Nil(t, nextTask)
}

func TestRecurrenceService_GenerateNextTask_ExpiredEndDateReturnsNil(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	seriesID := "series-1"
	task := createRecurringTask("user-123", "task-1", seriesID)

	// Set end date in the past - CanGenerateNext will return false
	pastEndDate := time.Now().Add(-24 * time.Hour)
	series := createTaskSeries("user-123", seriesID, "task-1", domain.RecurrencePatternDaily)
	series.EndDate = &pastEndDate

	seriesRepo.On("FindByID", mock.Anything, seriesID).Return(series, nil)

	nextTask, err := service.GenerateNextTask(context.Background(), task, nil)

	require.NoError(t, err)
	assert.Nil(t, nextTask) // CanGenerateNext returns false when end date passed
}

func TestRecurrenceService_GenerateNextTask_NextDueDateExceedsEndDateDeactivates(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	seriesID := "series-1"
	task := createRecurringTask("user-123", "task-1", seriesID)

	// End date is in the future but next due date will exceed it
	endDate := time.Now().Add(12 * time.Hour)     // Ends in 12 hours
	dueDate := time.Now().Add(6 * time.Hour)       // Due in 6 hours
	task.DueDate = &dueDate

	series := createTaskSeries("user-123", seriesID, "task-1", domain.RecurrencePatternDaily)
	series.EndDate = &endDate // Next due = dueDate + 1 day, which exceeds endDate

	seriesRepo.On("FindByID", mock.Anything, seriesID).Return(series, nil)
	seriesRepo.On("Deactivate", mock.Anything, seriesID, "user-123").Return(nil)

	nextTask, err := service.GenerateNextTask(context.Background(), task, nil)

	require.NoError(t, err)
	assert.Nil(t, nextTask)
	seriesRepo.AssertExpectations(t)
}

// =============================================================================
// GetSeriesHistory Tests
// =============================================================================

func TestRecurrenceService_GetSeriesHistory_ReturnsHistorySuccessfully(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	seriesID := "series-1"
	series := createTaskSeries("user-123", seriesID, "task-1", domain.RecurrencePatternDaily)

	tasks := []*domain.Task{
		createRecurringTask("user-123", "task-1", seriesID),
		createRecurringTask("user-123", "task-2", seriesID),
	}
	tasks[1].Status = domain.TaskStatusDone

	seriesRepo.On("FindByID", mock.Anything, seriesID).Return(series, nil)
	seriesRepo.On("GetTasksBySeriesID", mock.Anything, seriesID).Return(tasks, nil)

	history, err := service.GetSeriesHistory(context.Background(), "user-123", seriesID)

	require.NoError(t, err)
	assert.NotNil(t, history)
	assert.Equal(t, series, history.Series)
	assert.Len(t, history.Tasks, 2)
	assert.Equal(t, 2, history.Total)
}

func TestRecurrenceService_GetSeriesHistory_UnauthorizedReturnsError(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	seriesID := "series-1"
	series := createTaskSeries("other-user", seriesID, "task-1", domain.RecurrencePatternDaily)

	seriesRepo.On("FindByID", mock.Anything, seriesID).Return(series, nil)

	_, err := service.GetSeriesHistory(context.Background(), "user-123", seriesID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrUnauthorized))
}

func TestRecurrenceService_GetSeriesHistory_NotFoundReturnsError(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	seriesRepo.On("FindByID", mock.Anything, "non-existent").Return(nil, domain.ErrSeriesNotFound)

	_, err := service.GetSeriesHistory(context.Background(), "user-123", "non-existent")

	assert.Error(t, err)
}

// =============================================================================
// GetUserSeries Tests
// =============================================================================

func TestRecurrenceService_GetUserSeries_ReturnsAllSeries(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	allSeries := []*domain.TaskSeries{
		createTaskSeries("user-123", "series-1", "task-1", domain.RecurrencePatternDaily),
		createTaskSeries("user-123", "series-2", "task-2", domain.RecurrencePatternWeekly),
	}
	allSeries[1].IsActive = false

	seriesRepo.On("FindByUserID", mock.Anything, "user-123").Return(allSeries, nil)

	result, err := service.GetUserSeries(context.Background(), "user-123", false)

	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestRecurrenceService_GetUserSeries_ReturnsActiveOnly(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	activeSeries := []*domain.TaskSeries{
		createTaskSeries("user-123", "series-1", "task-1", domain.RecurrencePatternDaily),
	}

	seriesRepo.On("FindActiveByUserID", mock.Anything, "user-123").Return(activeSeries, nil)

	result, err := service.GetUserSeries(context.Background(), "user-123", true)

	require.NoError(t, err)
	assert.Len(t, result, 1)
}

// =============================================================================
// UpdateSeries Tests
// =============================================================================

func TestRecurrenceService_UpdateSeries_UpdatesPatternSuccessfully(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	seriesID := "series-1"
	series := createTaskSeries("user-123", seriesID, "task-1", domain.RecurrencePatternDaily)

	newPattern := domain.RecurrencePatternWeekly
	dto := &domain.UpdateTaskSeriesDTO{
		Pattern: &newPattern,
	}

	seriesRepo.On("FindByID", mock.Anything, seriesID).Return(series, nil)
	seriesRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.TaskSeries")).Return(nil)

	result, err := service.UpdateSeries(context.Background(), "user-123", seriesID, dto)

	require.NoError(t, err)
	assert.Equal(t, domain.RecurrencePatternWeekly, result.Pattern)
}

func TestRecurrenceService_UpdateSeries_UnauthorizedReturnsError(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	seriesID := "series-1"
	series := createTaskSeries("other-user", seriesID, "task-1", domain.RecurrencePatternDaily)

	dto := &domain.UpdateTaskSeriesDTO{}

	seriesRepo.On("FindByID", mock.Anything, seriesID).Return(series, nil)

	_, err := service.UpdateSeries(context.Background(), "user-123", seriesID, dto)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrUnauthorized))
}

func TestRecurrenceService_UpdateSeries_InvalidIntervalReturnsError(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	seriesID := "series-1"
	series := createTaskSeries("user-123", seriesID, "task-1", domain.RecurrencePatternDaily)

	invalidInterval := 0
	dto := &domain.UpdateTaskSeriesDTO{
		IntervalValue: &invalidInterval,
	}

	seriesRepo.On("FindByID", mock.Anything, seriesID).Return(series, nil)

	_, err := service.UpdateSeries(context.Background(), "user-123", seriesID, dto)

	assert.Error(t, err)
}

// =============================================================================
// DeactivateSeries Tests
// =============================================================================

func TestRecurrenceService_DeactivateSeries_DeactivatesSuccessfully(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	seriesID := "series-1"
	series := createTaskSeries("user-123", seriesID, "task-1", domain.RecurrencePatternDaily)

	seriesRepo.On("FindByID", mock.Anything, seriesID).Return(series, nil)
	seriesRepo.On("Deactivate", mock.Anything, seriesID, "user-123").Return(nil)

	err := service.DeactivateSeries(context.Background(), "user-123", seriesID)

	require.NoError(t, err)
	seriesRepo.AssertExpectations(t)
}

func TestRecurrenceService_DeactivateSeries_UnauthorizedReturnsError(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	seriesID := "series-1"
	series := createTaskSeries("other-user", seriesID, "task-1", domain.RecurrencePatternDaily)

	seriesRepo.On("FindByID", mock.Anything, seriesID).Return(series, nil)

	err := service.DeactivateSeries(context.Background(), "user-123", seriesID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrUnauthorized))
}

// =============================================================================
// GetEffectiveDueDateCalculation Tests
// =============================================================================

func TestRecurrenceService_GetEffectiveDueDateCalculation_CategoryPreferenceTakesPriority(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	category := "Work"
	catPref := &domain.CategoryPreference{
		UserID:             "user-123",
		Category:           category,
		DueDateCalculation: domain.DueDateFromCompletion,
	}

	prefsRepo.On("GetCategoryPreference", mock.Anything, "user-123", category).Return(catPref, nil)

	result := service.GetEffectiveDueDateCalculation(context.Background(), "user-123", &category)

	assert.Equal(t, domain.DueDateFromCompletion, result)
}

func TestRecurrenceService_GetEffectiveDueDateCalculation_FallsBackToUserDefault(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	category := "Work"
	userPrefs := &domain.UserPreferences{
		UserID:                    "user-123",
		DefaultDueDateCalculation: domain.DueDateFromCompletion,
	}

	prefsRepo.On("GetCategoryPreference", mock.Anything, "user-123", category).Return(nil, errors.New("not found"))
	prefsRepo.On("GetUserPreferences", mock.Anything, "user-123").Return(userPrefs, nil)

	result := service.GetEffectiveDueDateCalculation(context.Background(), "user-123", &category)

	assert.Equal(t, domain.DueDateFromCompletion, result)
}

func TestRecurrenceService_GetEffectiveDueDateCalculation_FallsBackToSystemDefault(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	category := "Work"

	prefsRepo.On("GetCategoryPreference", mock.Anything, "user-123", category).Return(nil, errors.New("not found"))
	prefsRepo.On("GetUserPreferences", mock.Anything, "user-123").Return(nil, errors.New("not found"))

	result := service.GetEffectiveDueDateCalculation(context.Background(), "user-123", &category)

	assert.Equal(t, domain.DueDateFromOriginal, result) // System default
}

func TestRecurrenceService_GetEffectiveDueDateCalculation_NoCategoryChecksUserDefault(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	userPrefs := &domain.UserPreferences{
		UserID:                    "user-123",
		DefaultDueDateCalculation: domain.DueDateFromCompletion,
	}

	prefsRepo.On("GetUserPreferences", mock.Anything, "user-123").Return(userPrefs, nil)

	result := service.GetEffectiveDueDateCalculation(context.Background(), "user-123", nil)

	assert.Equal(t, domain.DueDateFromCompletion, result)
}

// =============================================================================
// Preference Management Tests
// =============================================================================

func TestRecurrenceService_SetDefaultPreference_SetsPreferenceSuccessfully(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	prefsRepo.On("UpsertUserPreferences", mock.Anything, mock.AnythingOfType("*domain.UserPreferences")).Return(nil)

	err := service.SetDefaultPreference(context.Background(), "user-123", domain.DueDateFromCompletion)

	require.NoError(t, err)
	prefsRepo.AssertExpectations(t)
}

func TestRecurrenceService_SetCategoryPreference_SetsPreferenceSuccessfully(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	prefsRepo.On("UpsertCategoryPreference", mock.Anything, mock.AnythingOfType("*domain.CategoryPreference")).Return(nil)

	err := service.SetCategoryPreference(context.Background(), "user-123", "Work", domain.DueDateFromCompletion)

	require.NoError(t, err)
	prefsRepo.AssertExpectations(t)
}

func TestRecurrenceService_DeleteCategoryPreference_DeletesSuccessfully(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	prefsRepo.On("DeleteCategoryPreference", mock.Anything, "user-123", "Work").Return(nil)

	err := service.DeleteCategoryPreference(context.Background(), "user-123", "Work")

	require.NoError(t, err)
	prefsRepo.AssertExpectations(t)
}

func TestRecurrenceService_GetUserPreferences_ReturnsPreferences(t *testing.T) {
	taskRepo := new(MockTaskRepository)
	seriesRepo := new(MockTaskSeriesRepository)
	prefsRepo := new(MockUserPreferencesRepository)
	historyRepo := new(MockTaskHistoryRepository)
	service := newRecurrenceService(taskRepo, seriesRepo, prefsRepo, historyRepo)

	allPrefs := &domain.AllPreferences{
		UserPreferences: &domain.UserPreferences{
			UserID:                    "user-123",
			DefaultDueDateCalculation: domain.DueDateFromCompletion,
		},
		CategoryPreferences: []domain.CategoryPreference{
			{UserID: "user-123", Category: "Work", DueDateCalculation: domain.DueDateFromOriginal},
		},
	}

	prefsRepo.On("GetAllPreferences", mock.Anything, "user-123").Return(allPrefs, nil)

	result, err := service.GetUserPreferences(context.Background(), "user-123")

	require.NoError(t, err)
	assert.NotNil(t, result.UserPreferences)
	assert.Len(t, result.CategoryPreferences, 1)
}
