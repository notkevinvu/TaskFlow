package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/handler/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTaskService is a mock implementation of ports.TaskService
type MockTaskService struct {
	mock.Mock
}

func (m *MockTaskService) Create(ctx context.Context, userID string, dto *domain.CreateTaskDTO) (*domain.Task, error) {
	args := m.Called(ctx, userID, dto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *MockTaskService) Get(ctx context.Context, userID, taskID string) (*domain.Task, error) {
	args := m.Called(ctx, userID, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *MockTaskService) List(ctx context.Context, userID string, filter *domain.TaskListFilter) ([]*domain.Task, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Task), args.Error(1)
}

func (m *MockTaskService) Update(ctx context.Context, userID, taskID string, dto *domain.UpdateTaskDTO) (*domain.Task, error) {
	args := m.Called(ctx, userID, taskID, dto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *MockTaskService) Delete(ctx context.Context, userID, taskID string) error {
	args := m.Called(ctx, userID, taskID)
	return args.Error(0)
}

func (m *MockTaskService) Bump(ctx context.Context, userID, taskID string) (*domain.Task, error) {
	args := m.Called(ctx, userID, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *MockTaskService) Complete(ctx context.Context, userID, taskID string) (*domain.Task, error) {
	args := m.Called(ctx, userID, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *MockTaskService) GetAtRiskTasks(ctx context.Context, userID string) ([]*domain.Task, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Task), args.Error(1)
}

func (m *MockTaskService) GetCalendar(ctx context.Context, userID string, filter *domain.CalendarFilter) (*domain.CalendarResponse, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CalendarResponse), args.Error(1)
}

func (m *MockTaskService) RenameCategory(ctx context.Context, userID, oldName, newName string) (int, error) {
	args := m.Called(ctx, userID, oldName, newName)
	return args.Int(0), args.Error(1)
}

func (m *MockTaskService) DeleteCategory(ctx context.Context, userID, categoryName string) (int, error) {
	args := m.Called(ctx, userID, categoryName)
	return args.Int(0), args.Error(1)
}

// setupTaskTest creates a test router and mock service
func setupTaskTest() (*gin.Engine, *MockTaskService) {
	router := testutil.SetupTestRouter()
	mockService := new(MockTaskService)
	return router, mockService
}

// TestTaskHandler_Create_Success tests successful task creation
func TestTaskHandler_Create_Success(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	// Setup authenticated route
	router.POST("/tasks", testutil.WithAuthContext(router, "user-123", handler.Create))

	// Mock expectations
	expectedTask := testutil.NewTaskBuilder().
		WithID("task-123").
		WithUserID("user-123").
		WithTitle("Test Task").
		Build()

	mockService.On("Create", mock.Anything, "user-123", mock.MatchedBy(func(dto *domain.CreateTaskDTO) bool {
		return dto.Title == "Test Task"
	})).Return(expectedTask, nil)

	// Create request
	reqBody := map[string]interface{}{
		"title": "Test Task",
	}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/tasks", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)

	// Parse response
	var task domain.Task
	err := json.Unmarshal(w.Body.Bytes(), &task)
	assert.NoError(t, err)
	assert.Equal(t, "task-123", task.ID)
	assert.Equal(t, "Test Task", task.Title)
}

// TestTaskHandler_Create_Unauthenticated tests task creation without auth
func TestTaskHandler_Create_Unauthenticated(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	// Setup route WITHOUT auth
	router.POST("/tasks", handler.Create)

	// Create request
	reqBody := map[string]interface{}{
		"title": "Test Task",
	}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/tasks", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertNotCalled(t, "Create")
}

// TestTaskHandler_Create_InvalidJSON tests task creation with invalid JSON
func TestTaskHandler_Create_InvalidJSON(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	router.POST("/tasks", testutil.WithAuthContext(router, "user-123", handler.Create))

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/tasks", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "Create")
}

// TestTaskHandler_Create_ValidationError tests task creation with validation error from service
func TestTaskHandler_Create_ValidationError(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	router.POST("/tasks", testutil.WithAuthContext(router, "user-123", handler.Create))

	// Mock validation error from service (e.g., business rule violation)
	// Note: We send a valid title to pass Gin binding, but the service rejects it
	mockService.On("Create", mock.Anything, "user-123", mock.Anything).
		Return(nil, domain.NewValidationError("title", "title contains prohibited content"))

	// Create request with valid title (passes binding, but service rejects)
	reqBody := map[string]interface{}{
		"title": "Valid Title",
	}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/tasks", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

// TestTaskHandler_List_Success tests successful task listing
func TestTaskHandler_List_Success(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	router.GET("/tasks", testutil.WithAuthContext(router, "user-123", handler.List))

	// Mock expectations
	expectedTasks := []*domain.Task{
		testutil.NewTaskBuilder().WithID("task-1").WithTitle("Task 1").Build(),
		testutil.NewTaskBuilder().WithID("task-2").WithTitle("Task 2").Build(),
	}

	mockService.On("List", mock.Anything, "user-123", mock.MatchedBy(func(filter *domain.TaskListFilter) bool {
		return filter.Limit == 20 && filter.Offset == 0
	})).Return(expectedTasks, nil)

	// Create request
	req := httptest.NewRequest("GET", "/tasks", nil)

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	// Parse response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(2), response["total_count"])
}

// TestTaskHandler_List_WithFilters tests task listing with filters
func TestTaskHandler_List_WithFilters(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	router.GET("/tasks", testutil.WithAuthContext(router, "user-123", handler.List))

	// Mock expectations with specific filters
	expectedTasks := []*domain.Task{
		testutil.NewTaskBuilder().WithID("task-1").WithStatus(domain.TaskStatusTodo).Build(),
	}

	mockService.On("List", mock.Anything, "user-123", mock.MatchedBy(func(filter *domain.TaskListFilter) bool {
		return filter.Status != nil && *filter.Status == domain.TaskStatusTodo &&
			filter.Category != nil && *filter.Category == "work" &&
			filter.Limit == 10 && filter.Offset == 5
	})).Return(expectedTasks, nil)

	// Create request with filters
	req := httptest.NewRequest("GET", "/tasks?status=todo&category=work&limit=10&offset=5", nil)

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// TestTaskHandler_List_EmptyResult tests task listing with no tasks
func TestTaskHandler_List_EmptyResult(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	router.GET("/tasks", testutil.WithAuthContext(router, "user-123", handler.List))

	// Mock empty result
	mockService.On("List", mock.Anything, "user-123", mock.Anything).
		Return(nil, nil)

	// Create request
	req := httptest.NewRequest("GET", "/tasks", nil)

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response - should return empty array, not null
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	tasks := response["tasks"].([]interface{})
	assert.NotNil(t, tasks)
	assert.Equal(t, 0, len(tasks))
}

// TestTaskHandler_List_InvalidMinPriority tests validation of min_priority parameter
func TestTaskHandler_List_InvalidMinPriority(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	router.GET("/tasks", testutil.WithAuthContext(router, "user-123", handler.List))

	// Create request with invalid min_priority
	req := httptest.NewRequest("GET", "/tasks?min_priority=150", nil)

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "List")
}

// TestTaskHandler_List_InvalidPriorityRange tests validation of priority range
func TestTaskHandler_List_InvalidPriorityRange(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	router.GET("/tasks", testutil.WithAuthContext(router, "user-123", handler.List))

	// Create request with min > max
	req := httptest.NewRequest("GET", "/tasks?min_priority=80&max_priority=20", nil)

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "List")
}

// TestTaskHandler_Get_Success tests successful task retrieval
func TestTaskHandler_Get_Success(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	router.GET("/tasks/:id", testutil.WithAuthContext(router, "user-123", handler.Get))

	// Mock expectations
	expectedTask := testutil.NewTaskBuilder().
		WithID("task-123").
		WithUserID("user-123").
		WithTitle("Test Task").
		Build()

	mockService.On("Get", mock.Anything, "user-123", "task-123").
		Return(expectedTask, nil)

	// Create request
	req := httptest.NewRequest("GET", "/tasks/task-123", nil)

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	// Parse response
	var task domain.Task
	err := json.Unmarshal(w.Body.Bytes(), &task)
	assert.NoError(t, err)
	assert.Equal(t, "task-123", task.ID)
}

// TestTaskHandler_Get_NotFound tests getting non-existent task
func TestTaskHandler_Get_NotFound(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	router.GET("/tasks/:id", testutil.WithAuthContext(router, "user-123", handler.Get))

	// Mock not found error
	mockService.On("Get", mock.Anything, "user-123", "nonexistent").
		Return(nil, domain.NewNotFoundError("task", "nonexistent"))

	// Create request
	req := httptest.NewRequest("GET", "/tasks/nonexistent", nil)

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// TestTaskHandler_Get_Unauthenticated tests getting task without auth
func TestTaskHandler_Get_Unauthenticated(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	router.GET("/tasks/:id", handler.Get)

	// Create request
	req := httptest.NewRequest("GET", "/tasks/task-123", nil)

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertNotCalled(t, "Get")
}

// TestTaskHandler_Update_Success tests successful task update
func TestTaskHandler_Update_Success(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	router.PUT("/tasks/:id", testutil.WithAuthContext(router, "user-123", handler.Update))

	// Mock expectations
	updatedTask := testutil.NewTaskBuilder().
		WithID("task-123").
		WithTitle("Updated Task").
		WithStatus(domain.TaskStatusInProgress).
		Build()

	mockService.On("Update", mock.Anything, "user-123", "task-123", mock.MatchedBy(func(dto *domain.UpdateTaskDTO) bool {
		return dto.Title != nil && *dto.Title == "Updated Task"
	})).Return(updatedTask, nil)

	// Create request
	newTitle := "Updated Task"
	reqBody := map[string]interface{}{
		"title": newTitle,
	}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/tasks/task-123", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	// Parse response
	var task domain.Task
	err := json.Unmarshal(w.Body.Bytes(), &task)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Task", task.Title)
}

// TestTaskHandler_Update_InvalidJSON tests update with invalid JSON
func TestTaskHandler_Update_InvalidJSON(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	router.PUT("/tasks/:id", testutil.WithAuthContext(router, "user-123", handler.Update))

	// Create request with invalid JSON
	req := httptest.NewRequest("PUT", "/tasks/task-123", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "Update")
}

// TestTaskHandler_Update_NotFound tests updating non-existent task
func TestTaskHandler_Update_NotFound(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	router.PUT("/tasks/:id", testutil.WithAuthContext(router, "user-123", handler.Update))

	// Mock not found error
	mockService.On("Update", mock.Anything, "user-123", "nonexistent", mock.Anything).
		Return(nil, domain.NewNotFoundError("task", "nonexistent"))

	// Create request
	reqBody := map[string]interface{}{"title": "Updated"}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("PUT", "/tasks/nonexistent", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// TestTaskHandler_Delete_Success tests successful task deletion
func TestTaskHandler_Delete_Success(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	router.DELETE("/tasks/:id", testutil.WithAuthContext(router, "user-123", handler.Delete))

	// Mock expectations - Get is called first for metrics recording
	category := "Work"
	mockService.On("Get", mock.Anything, "user-123", "task-123").
		Return(&domain.Task{
			ID:       "task-123",
			Category: &category,
		}, nil)
	mockService.On("Delete", mock.Anything, "user-123", "task-123").
		Return(nil)

	// Create request
	req := httptest.NewRequest("DELETE", "/tasks/task-123", nil)

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code)
	mockService.AssertExpectations(t)
}

// TestTaskHandler_Delete_NotFound tests deleting non-existent task
func TestTaskHandler_Delete_NotFound(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	router.DELETE("/tasks/:id", testutil.WithAuthContext(router, "user-123", handler.Delete))

	// Mock Get returning not found (still attempt delete for consistency)
	mockService.On("Get", mock.Anything, "user-123", "nonexistent").
		Return(nil, domain.NewNotFoundError("task", "nonexistent"))
	// Mock not found error
	mockService.On("Delete", mock.Anything, "user-123", "nonexistent").
		Return(domain.NewNotFoundError("task", "nonexistent"))

	// Create request
	req := httptest.NewRequest("DELETE", "/tasks/nonexistent", nil)

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// TestTaskHandler_Delete_Unauthenticated tests deletion without auth
func TestTaskHandler_Delete_Unauthenticated(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	router.DELETE("/tasks/:id", handler.Delete)

	// Create request
	req := httptest.NewRequest("DELETE", "/tasks/task-123", nil)

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertNotCalled(t, "Delete")
}

// TestTaskHandler_Bump_Success tests successful task bump
func TestTaskHandler_Bump_Success(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	router.POST("/tasks/:id/bump", testutil.WithAuthContext(router, "user-123", handler.Bump))

	// Mock expectations
	bumpedTask := testutil.NewTaskBuilder().
		WithID("task-123").
		WithBumpCount(1).
		Build()

	mockService.On("Bump", mock.Anything, "user-123", "task-123").
		Return(bumpedTask, nil)

	// Create request
	req := httptest.NewRequest("POST", "/tasks/task-123/bump", nil)

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	// Parse response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Task bumped successfully", response["message"])
	assert.NotNil(t, response["task"])
}

// TestTaskHandler_Bump_NotFound tests bumping non-existent task
func TestTaskHandler_Bump_NotFound(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	router.POST("/tasks/:id/bump", testutil.WithAuthContext(router, "user-123", handler.Bump))

	// Mock not found error
	mockService.On("Bump", mock.Anything, "user-123", "nonexistent").
		Return(nil, domain.NewNotFoundError("task", "nonexistent"))

	// Create request
	req := httptest.NewRequest("POST", "/tasks/nonexistent/bump", nil)

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// TestTaskHandler_Complete_Success tests successful task completion
func TestTaskHandler_Complete_Success(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	router.POST("/tasks/:id/complete", testutil.WithAuthContext(router, "user-123", handler.Complete))

	// Mock expectations
	completedTask := testutil.NewTaskBuilder().
		WithID("task-123").
		Completed().
		Build()

	mockService.On("Complete", mock.Anything, "user-123", "task-123").
		Return(completedTask, nil)

	// Create request
	req := httptest.NewRequest("POST", "/tasks/task-123/complete", nil)

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	// Parse response
	var task domain.Task
	err := json.Unmarshal(w.Body.Bytes(), &task)
	assert.NoError(t, err)
	assert.Equal(t, domain.TaskStatusDone, task.Status)
	assert.NotNil(t, task.CompletedAt)
}

// TestTaskHandler_Complete_NotFound tests completing non-existent task
func TestTaskHandler_Complete_NotFound(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	router.POST("/tasks/:id/complete", testutil.WithAuthContext(router, "user-123", handler.Complete))

	// Mock not found error
	mockService.On("Complete", mock.Anything, "user-123", "nonexistent").
		Return(nil, domain.NewNotFoundError("task", "nonexistent"))

	// Create request
	req := httptest.NewRequest("POST", "/tasks/nonexistent/complete", nil)

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

// TestTaskHandler_GetCalendar_Success tests successful calendar retrieval
func TestTaskHandler_GetCalendar_Success(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	router.GET("/tasks/calendar", testutil.WithAuthContext(router, "user-123", handler.GetCalendar))

	// Mock expectations
	startDate := time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2025, 11, 30, 0, 0, 0, 0, time.UTC)

	expectedResponse := &domain.CalendarResponse{
		Dates: map[string]*domain.CalendarDayData{
			"2025-11-01": {
				Count:      1,
				BadgeColor: "blue",
				Tasks: []*domain.Task{
					testutil.NewTaskBuilder().WithID("task-1").WithDueDate(startDate).Build(),
				},
			},
		},
	}

	mockService.On("GetCalendar", mock.Anything, "user-123", mock.MatchedBy(func(filter *domain.CalendarFilter) bool {
		return filter.StartDate.Equal(startDate) && filter.EndDate.Equal(endDate)
	})).Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest("GET", "/tasks/calendar?start_date=2025-11-01&end_date=2025-11-30", nil)

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

// TestTaskHandler_GetCalendar_MissingStartDate tests calendar without start_date
func TestTaskHandler_GetCalendar_MissingStartDate(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	router.GET("/tasks/calendar", testutil.WithAuthContext(router, "user-123", handler.GetCalendar))

	// Create request without start_date
	req := httptest.NewRequest("GET", "/tasks/calendar?end_date=2025-11-30", nil)

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "GetCalendar")
}

// TestTaskHandler_GetCalendar_InvalidDateFormat tests calendar with invalid date format
func TestTaskHandler_GetCalendar_InvalidDateFormat(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	router.GET("/tasks/calendar", testutil.WithAuthContext(router, "user-123", handler.GetCalendar))

	// Create request with invalid date format
	req := httptest.NewRequest("GET", "/tasks/calendar?start_date=11/01/2025&end_date=2025-11-30", nil)

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "GetCalendar")
}

// TestTaskHandler_GetCalendar_ExceedsMaxRange tests calendar with date range > 90 days
func TestTaskHandler_GetCalendar_ExceedsMaxRange(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	router.GET("/tasks/calendar", testutil.WithAuthContext(router, "user-123", handler.GetCalendar))

	// Create request with 91 days range
	req := httptest.NewRequest("GET", "/tasks/calendar?start_date=2025-11-01&end_date=2026-02-01", nil)

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "GetCalendar")
}

// TestTaskHandler_GetCalendar_WithStatusFilter tests calendar with status filter
func TestTaskHandler_GetCalendar_WithStatusFilter(t *testing.T) {
	router, mockService := setupTaskTest()
	handler := NewTaskHandler(mockService)

	router.GET("/tasks/calendar", testutil.WithAuthContext(router, "user-123", handler.GetCalendar))

	// Mock expectations with status filter
	expectedResponse := &domain.CalendarResponse{
		Dates: map[string]*domain.CalendarDayData{},
	}

	mockService.On("GetCalendar", mock.Anything, "user-123", mock.MatchedBy(func(filter *domain.CalendarFilter) bool {
		return len(filter.Status) == 2 &&
			filter.Status[0] == domain.TaskStatusTodo &&
			filter.Status[1] == domain.TaskStatusInProgress
	})).Return(expectedResponse, nil)

	// Create request with status filter
	req := httptest.NewRequest("GET", "/tasks/calendar?start_date=2025-11-01&end_date=2025-11-30&status=todo,in_progress", nil)

	// Execute request
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}
