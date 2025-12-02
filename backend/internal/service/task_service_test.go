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

// Helper to create a valid task for testing
func createTestTask(userID, taskID string) *domain.Task {
	now := time.Now()
	desc := "Test Description"
	return &domain.Task{
		ID:            taskID,
		UserID:        userID,
		Title:         "Test Task",
		Description:   &desc,
		Status:        domain.TaskStatusTodo,
		UserPriority:  5,
		BumpCount:     0,
		PriorityScore: 50,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// =============================================================================
// TaskService.Create Tests
// =============================================================================

func TestTaskService_Create_Success(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	userID := "user-123"
	dto := &domain.CreateTaskDTO{
		Title:       "New Task",
		Description: stringPtr("Task description"),
	}

	mockTaskRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)
	mockHistoryRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.TaskHistory")).Return(nil)

	task, err := service.Create(context.Background(), userID, dto)

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "New Task", task.Title)
	assert.Equal(t, "Task description", *task.Description)
	assert.Equal(t, userID, task.UserID)
	assert.Equal(t, domain.TaskStatusTodo, task.Status)
	assert.Equal(t, 5, task.UserPriority) // Default priority
	assert.NotEmpty(t, task.ID)
	mockTaskRepo.AssertExpectations(t)
}

func TestTaskService_Create_WithAllFields(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	userID := "user-123"
	priority := 8
	effort := domain.TaskEffortMedium
	dueDate := time.Now().Add(24 * time.Hour)
	dto := &domain.CreateTaskDTO{
		Title:           "Full Task",
		Description:     stringPtr("Full description"),
		UserPriority:    &priority,
		EstimatedEffort: &effort,
		DueDate:         &dueDate,
		Category:        stringPtr("work"),
		Context:         stringPtr("office"),
		RelatedPeople:   []string{"John", "Jane"},
	}

	mockTaskRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)
	mockHistoryRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.TaskHistory")).Return(nil)

	task, err := service.Create(context.Background(), userID, dto)

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, 8, task.UserPriority)
	assert.Equal(t, domain.TaskEffortMedium, *task.EstimatedEffort)
	assert.Equal(t, "work", *task.Category)
	assert.Equal(t, "office", *task.Context)
	assert.Len(t, task.RelatedPeople, 2)
}

func TestTaskService_Create_EmptyTitle(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	dto := &domain.CreateTaskDTO{
		Title: "",
	}

	task, err := service.Create(context.Background(), "user-123", dto)

	assert.Error(t, err)
	assert.Nil(t, task)
	assert.Contains(t, err.Error(), "title")
}

func TestTaskService_Create_TitleTooLong(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	// Create title longer than 200 characters
	longTitle := make([]byte, 201)
	for i := range longTitle {
		longTitle[i] = 'a'
	}

	dto := &domain.CreateTaskDTO{
		Title: string(longTitle),
	}

	task, err := service.Create(context.Background(), "user-123", dto)

	assert.Error(t, err)
	assert.Nil(t, task)
}

func TestTaskService_Create_InvalidPriority(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	invalidPriority := 11 // Valid range is 1-10
	dto := &domain.CreateTaskDTO{
		Title:        "Test Task",
		UserPriority: &invalidPriority,
	}

	task, err := service.Create(context.Background(), "user-123", dto)

	assert.Error(t, err)
	assert.Nil(t, task)
}

func TestTaskService_Create_RepoError(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	dto := &domain.CreateTaskDTO{
		Title: "Test Task",
	}

	mockTaskRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Task")).
		Return(errors.New("database error"))

	task, err := service.Create(context.Background(), "user-123", dto)

	assert.Error(t, err)
	assert.Nil(t, task)
	mockTaskRepo.AssertExpectations(t)
}

// =============================================================================
// TaskService.Get Tests
// =============================================================================

func TestTaskService_Get_Success(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	userID := "user-123"
	taskID := "task-456"
	expectedTask := createTestTask(userID, taskID)

	mockTaskRepo.On("FindByID", mock.Anything, taskID).Return(expectedTask, nil)

	task, err := service.Get(context.Background(), userID, taskID)

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, taskID, task.ID)
	assert.Equal(t, userID, task.UserID)
	mockTaskRepo.AssertExpectations(t)
}

func TestTaskService_Get_NotFound(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	mockTaskRepo.On("FindByID", mock.Anything, "nonexistent").Return(nil, nil)

	task, err := service.Get(context.Background(), "user-123", "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, task)
	var notFoundErr *domain.NotFoundError
	assert.True(t, errors.As(err, &notFoundErr))
}

func TestTaskService_Get_Forbidden(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	taskID := "task-456"
	// Task belongs to different user
	existingTask := createTestTask("other-user", taskID)

	mockTaskRepo.On("FindByID", mock.Anything, taskID).Return(existingTask, nil)

	task, err := service.Get(context.Background(), "user-123", taskID)

	assert.Error(t, err)
	assert.Nil(t, task)
	var forbiddenErr *domain.ForbiddenError
	assert.True(t, errors.As(err, &forbiddenErr))
}

// =============================================================================
// TaskService.Update Tests
// =============================================================================

func TestTaskService_Update_Success(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	userID := "user-123"
	taskID := "task-456"
	existingTask := createTestTask(userID, taskID)

	newTitle := "Updated Title"
	dto := &domain.UpdateTaskDTO{
		Title: &newTitle,
	}

	mockTaskRepo.On("FindByID", mock.Anything, taskID).Return(existingTask, nil)
	mockTaskRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)
	mockHistoryRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.TaskHistory")).Return(nil)

	task, err := service.Update(context.Background(), userID, taskID, dto)

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, "Updated Title", task.Title)
	mockTaskRepo.AssertExpectations(t)
}

func TestTaskService_Update_NotFound(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	mockTaskRepo.On("FindByID", mock.Anything, "nonexistent").Return(nil, nil)

	newTitle := "Updated Title"
	dto := &domain.UpdateTaskDTO{Title: &newTitle}

	task, err := service.Update(context.Background(), "user-123", "nonexistent", dto)

	assert.Error(t, err)
	assert.Nil(t, task)
}

func TestTaskService_Update_Forbidden(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	taskID := "task-456"
	existingTask := createTestTask("other-user", taskID)

	mockTaskRepo.On("FindByID", mock.Anything, taskID).Return(existingTask, nil)

	newTitle := "Updated Title"
	dto := &domain.UpdateTaskDTO{Title: &newTitle}

	task, err := service.Update(context.Background(), "user-123", taskID, dto)

	assert.Error(t, err)
	assert.Nil(t, task)
	var forbiddenErr *domain.ForbiddenError
	assert.True(t, errors.As(err, &forbiddenErr))
}

func TestTaskService_Update_SetStatus_Done(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	userID := "user-123"
	taskID := "task-456"
	existingTask := createTestTask(userID, taskID)

	status := domain.TaskStatusDone
	dto := &domain.UpdateTaskDTO{Status: &status}

	mockTaskRepo.On("FindByID", mock.Anything, taskID).Return(existingTask, nil)
	mockTaskRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)
	mockHistoryRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.TaskHistory")).Return(nil)

	task, err := service.Update(context.Background(), userID, taskID, dto)

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, domain.TaskStatusDone, task.Status)
	assert.NotNil(t, task.CompletedAt)
}

// =============================================================================
// TaskService.Delete Tests
// =============================================================================

func TestTaskService_Delete_Success(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	userID := "user-123"
	taskID := "task-456"
	existingTask := createTestTask(userID, taskID)

	mockTaskRepo.On("FindByID", mock.Anything, taskID).Return(existingTask, nil)
	mockHistoryRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.TaskHistory")).Return(nil)
	mockTaskRepo.On("Delete", mock.Anything, taskID, userID).Return(nil)

	err := service.Delete(context.Background(), userID, taskID)

	assert.NoError(t, err)
	mockTaskRepo.AssertExpectations(t)
}

func TestTaskService_Delete_NotFound(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	mockTaskRepo.On("FindByID", mock.Anything, "nonexistent").Return(nil, nil)

	err := service.Delete(context.Background(), "user-123", "nonexistent")

	assert.Error(t, err)
}

func TestTaskService_Delete_Forbidden(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	taskID := "task-456"
	existingTask := createTestTask("other-user", taskID)

	mockTaskRepo.On("FindByID", mock.Anything, taskID).Return(existingTask, nil)

	err := service.Delete(context.Background(), "user-123", taskID)

	assert.Error(t, err)
	var forbiddenErr *domain.ForbiddenError
	assert.True(t, errors.As(err, &forbiddenErr))
}

// =============================================================================
// TaskService.Bump Tests
// =============================================================================

func TestTaskService_Bump_Success(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	userID := "user-123"
	taskID := "task-456"
	existingTask := createTestTask(userID, taskID)
	bumpedTask := createTestTask(userID, taskID)
	bumpedTask.BumpCount = 1

	mockTaskRepo.On("FindByID", mock.Anything, taskID).Return(existingTask, nil).Once()
	mockTaskRepo.On("IncrementBumpCount", mock.Anything, taskID, userID).Return(nil)
	mockTaskRepo.On("FindByID", mock.Anything, taskID).Return(bumpedTask, nil).Once()
	mockTaskRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)
	mockHistoryRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.TaskHistory")).Return(nil)

	task, err := service.Bump(context.Background(), userID, taskID)

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, 1, task.BumpCount)
	mockTaskRepo.AssertExpectations(t)
}

func TestTaskService_Bump_NotFound(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	mockTaskRepo.On("FindByID", mock.Anything, "nonexistent").Return(nil, nil)

	task, err := service.Bump(context.Background(), "user-123", "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, task)
}

func TestTaskService_Bump_Forbidden(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	taskID := "task-456"
	existingTask := createTestTask("other-user", taskID)

	mockTaskRepo.On("FindByID", mock.Anything, taskID).Return(existingTask, nil)

	task, err := service.Bump(context.Background(), "user-123", taskID)

	assert.Error(t, err)
	assert.Nil(t, task)
}

// =============================================================================
// TaskService.Complete Tests
// =============================================================================

func TestTaskService_Complete_Success(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	userID := "user-123"
	taskID := "task-456"
	existingTask := createTestTask(userID, taskID)

	mockTaskRepo.On("FindByID", mock.Anything, taskID).Return(existingTask, nil)
	mockTaskRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Task")).Return(nil)
	mockHistoryRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.TaskHistory")).Return(nil)

	task, err := service.Complete(context.Background(), userID, taskID)

	assert.NoError(t, err)
	assert.NotNil(t, task)
	assert.Equal(t, domain.TaskStatusDone, task.Status)
	assert.NotNil(t, task.CompletedAt)
	mockTaskRepo.AssertExpectations(t)
}

func TestTaskService_Complete_NotFound(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	mockTaskRepo.On("FindByID", mock.Anything, "nonexistent").Return(nil, nil)

	task, err := service.Complete(context.Background(), "user-123", "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, task)
}

// =============================================================================
// TaskService.GetAtRiskTasks Tests
// =============================================================================

func TestTaskService_GetAtRiskTasks_Success(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	userID := "user-123"
	atRiskTask := createTestTask(userID, "task-1")
	atRiskTask.BumpCount = 3

	mockTaskRepo.On("FindAtRiskTasks", mock.Anything, userID).Return([]*domain.Task{atRiskTask}, nil)

	tasks, err := service.GetAtRiskTasks(context.Background(), userID)

	assert.NoError(t, err)
	assert.Len(t, tasks, 1)
	assert.Equal(t, 3, tasks[0].BumpCount)
	mockTaskRepo.AssertExpectations(t)
}

// =============================================================================
// TaskService.GetCalendar Tests
// =============================================================================

func TestTaskService_GetCalendar_Success(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	userID := "user-123"
	dueDate := time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC)
	task := createTestTask(userID, "task-1")
	task.DueDate = &dueDate
	task.PriorityScore = 50

	filter := &domain.CalendarFilter{
		StartDate: time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
	}

	mockTaskRepo.On("FindByDateRange", mock.Anything, userID, filter).Return([]*domain.Task{task}, nil)

	response, err := service.GetCalendar(context.Background(), userID, filter)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Contains(t, response.Dates, "2025-12-15")
	assert.Equal(t, 1, response.Dates["2025-12-15"].Count)
	assert.Equal(t, "yellow", response.Dates["2025-12-15"].BadgeColor) // Priority 50 = yellow
	mockTaskRepo.AssertExpectations(t)
}

func TestTaskService_GetCalendar_BadgeColors(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	userID := "user-123"

	tests := []struct {
		name           string
		priorityScore  int
		expectedColor  string
	}{
		{"High priority (red)", 75, "red"},
		{"Medium priority (yellow)", 50, "yellow"},
		{"Low priority (blue)", 30, "blue"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dueDate := time.Date(2025, 12, 15, 0, 0, 0, 0, time.UTC)
			task := createTestTask(userID, "task-1")
			task.DueDate = &dueDate
			task.PriorityScore = tt.priorityScore

			filter := &domain.CalendarFilter{
				StartDate: time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
				EndDate:   time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
			}

			mockTaskRepo.On("FindByDateRange", mock.Anything, userID, filter).Return([]*domain.Task{task}, nil).Once()

			response, err := service.GetCalendar(context.Background(), userID, filter)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedColor, response.Dates["2025-12-15"].BadgeColor)
		})
	}
}

// =============================================================================
// TaskService.RenameCategory Tests
// =============================================================================

func TestTaskService_RenameCategory_Success(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	userID := "user-123"

	mockTaskRepo.On("RenameCategoryForUser", mock.Anything, userID, "old-name", "new-name").Return(5, nil)

	count, err := service.RenameCategory(context.Background(), userID, "old-name", "new-name")

	assert.NoError(t, err)
	assert.Equal(t, 5, count)
	mockTaskRepo.AssertExpectations(t)
}

func TestTaskService_RenameCategory_EmptyOldName(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	count, err := service.RenameCategory(context.Background(), "user-123", "", "new-name")

	assert.Error(t, err)
	assert.Equal(t, 0, count)
}

func TestTaskService_RenameCategory_EmptyNewName(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	count, err := service.RenameCategory(context.Background(), "user-123", "old-name", "")

	assert.Error(t, err)
	assert.Equal(t, 0, count)
}

// =============================================================================
// TaskService.DeleteCategory Tests
// =============================================================================

func TestTaskService_DeleteCategory_Success(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	userID := "user-123"

	mockTaskRepo.On("DeleteCategoryForUser", mock.Anything, userID, "category-to-delete").Return(3, nil)

	count, err := service.DeleteCategory(context.Background(), userID, "category-to-delete")

	assert.NoError(t, err)
	assert.Equal(t, 3, count)
	mockTaskRepo.AssertExpectations(t)
}

func TestTaskService_DeleteCategory_EmptyName(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	count, err := service.DeleteCategory(context.Background(), "user-123", "")

	assert.Error(t, err)
	assert.Equal(t, 0, count)
}

// =============================================================================
// TaskService.List Tests
// =============================================================================

func TestTaskService_List_Success(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	desc := "Task description"
	expectedTasks := []*domain.Task{
		{ID: "task-1", UserID: "user-123", Title: "Task 1", Status: domain.TaskStatusTodo},
		{ID: "task-2", UserID: "user-123", Title: "Task 2", Description: &desc, Status: domain.TaskStatusInProgress},
	}

	mockTaskRepo.On("List", mock.Anything, "user-123", (*domain.TaskListFilter)(nil)).Return(expectedTasks, nil)

	tasks, err := service.List(context.Background(), "user-123", nil)

	require.NoError(t, err)
	require.Len(t, tasks, 2)
	assert.Equal(t, "task-1", tasks[0].ID)
	assert.Equal(t, "task-2", tasks[1].ID)
	mockTaskRepo.AssertExpectations(t)
}

func TestTaskService_List_WithStatusFilter(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	status := domain.TaskStatusTodo
	filter := &domain.TaskListFilter{Status: &status}

	expectedTasks := []*domain.Task{
		{ID: "task-1", UserID: "user-123", Title: "Task 1", Status: domain.TaskStatusTodo},
	}

	mockTaskRepo.On("List", mock.Anything, "user-123", filter).Return(expectedTasks, nil)

	tasks, err := service.List(context.Background(), "user-123", filter)

	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, domain.TaskStatusTodo, tasks[0].Status)
	mockTaskRepo.AssertExpectations(t)
}

func TestTaskService_List_WithCategoryFilter(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	category := "work"
	filter := &domain.TaskListFilter{Category: &category}

	expectedTasks := []*domain.Task{
		{ID: "task-1", UserID: "user-123", Title: "Task 1", Category: &category},
	}

	mockTaskRepo.On("List", mock.Anything, "user-123", filter).Return(expectedTasks, nil)

	tasks, err := service.List(context.Background(), "user-123", filter)

	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, "work", *tasks[0].Category)
	mockTaskRepo.AssertExpectations(t)
}

func TestTaskService_List_WithPagination(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	filter := &domain.TaskListFilter{Limit: 10, Offset: 20}

	expectedTasks := []*domain.Task{
		{ID: "task-21", UserID: "user-123", Title: "Task 21"},
	}

	mockTaskRepo.On("List", mock.Anything, "user-123", filter).Return(expectedTasks, nil)

	tasks, err := service.List(context.Background(), "user-123", filter)

	require.NoError(t, err)
	require.Len(t, tasks, 1)
	mockTaskRepo.AssertExpectations(t)
}

func TestTaskService_List_EmptyResult(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	mockTaskRepo.On("List", mock.Anything, "user-123", (*domain.TaskListFilter)(nil)).Return([]*domain.Task{}, nil)

	tasks, err := service.List(context.Background(), "user-123", nil)

	require.NoError(t, err)
	assert.Empty(t, tasks)
	mockTaskRepo.AssertExpectations(t)
}

func TestTaskService_List_RepoError(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	mockTaskRepo.On("List", mock.Anything, "user-123", (*domain.TaskListFilter)(nil)).
		Return(nil, errors.New("database error"))

	tasks, err := service.List(context.Background(), "user-123", nil)

	assert.Error(t, err)
	assert.Nil(t, tasks)
	mockTaskRepo.AssertExpectations(t)
}

// =============================================================================
// TaskService.Complete Additional Tests
// =============================================================================

func TestTaskService_Complete_Forbidden(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	existingTask := &domain.Task{
		ID:     "task-123",
		UserID: "other-user", // Different user owns this task
		Title:  "Other user's task",
		Status: domain.TaskStatusTodo,
	}

	mockTaskRepo.On("FindByID", mock.Anything, "task-123").Return(existingTask, nil)

	task, err := service.Complete(context.Background(), "user-123", "task-123")

	assert.Error(t, err)
	assert.Nil(t, task)
	var forbiddenErr *domain.ForbiddenError
	assert.True(t, errors.As(err, &forbiddenErr))
	mockTaskRepo.AssertExpectations(t)
}

// =============================================================================
// TaskService.GetAtRiskTasks Additional Tests
// =============================================================================

func TestTaskService_GetAtRiskTasks_RepoError(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	mockTaskRepo.On("FindAtRiskTasks", mock.Anything, "user-123").
		Return(nil, errors.New("database error"))

	tasks, err := service.GetAtRiskTasks(context.Background(), "user-123")

	assert.Error(t, err)
	assert.Nil(t, tasks)
	mockTaskRepo.AssertExpectations(t)
}

// =============================================================================
// TaskService.GetCalendar Additional Tests
// =============================================================================

func TestTaskService_GetCalendar_RepoError(t *testing.T) {
	mockTaskRepo := new(MockTaskRepository)
	mockHistoryRepo := new(MockTaskHistoryRepository)
	service := NewTaskService(mockTaskRepo, mockHistoryRepo)

	filter := &domain.CalendarFilter{
		StartDate: time.Now(),
		EndDate:   time.Now().Add(7 * 24 * time.Hour),
	}

	mockTaskRepo.On("FindByDateRange", mock.Anything, "user-123", filter).
		Return(nil, errors.New("database error"))

	response, err := service.GetCalendar(context.Background(), "user-123", filter)

	assert.Error(t, err)
	assert.Nil(t, response)
	mockTaskRepo.AssertExpectations(t)
}

// =============================================================================
// Helper Functions
// =============================================================================

func stringPtr(s string) *string {
	return &s
}
