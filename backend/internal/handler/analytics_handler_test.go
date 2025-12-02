package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/handler/testutil"
	"github.com/notkevinvu/taskflow/backend/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTaskRepository is a mock implementation of ports.TaskRepository for analytics tests
type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) Create(ctx context.Context, task *domain.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) FindByID(ctx context.Context, id string) (*domain.Task, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Task), args.Error(1)
}

func (m *MockTaskRepository) List(ctx context.Context, userID string, filter *domain.TaskListFilter) ([]*domain.Task, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Task), args.Error(1)
}

func (m *MockTaskRepository) Update(ctx context.Context, task *domain.Task) error {
	args := m.Called(ctx, task)
	return args.Error(0)
}

func (m *MockTaskRepository) Delete(ctx context.Context, id, userID string) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockTaskRepository) IncrementBumpCount(ctx context.Context, id, userID string) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockTaskRepository) FindAtRiskTasks(ctx context.Context, userID string) ([]*domain.Task, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Task), args.Error(1)
}

func (m *MockTaskRepository) GetCategories(ctx context.Context, userID string) ([]string, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockTaskRepository) FindByDateRange(ctx context.Context, userID string, filter *domain.CalendarFilter) ([]*domain.Task, error) {
	args := m.Called(ctx, userID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Task), args.Error(1)
}

func (m *MockTaskRepository) RenameCategoryForUser(ctx context.Context, userID, oldName, newName string) (int, error) {
	args := m.Called(ctx, userID, oldName, newName)
	return args.Int(0), args.Error(1)
}

func (m *MockTaskRepository) DeleteCategoryForUser(ctx context.Context, userID, categoryName string) (int, error) {
	args := m.Called(ctx, userID, categoryName)
	return args.Int(0), args.Error(1)
}

func (m *MockTaskRepository) GetCompletionStats(ctx context.Context, userID string, daysBack int) (*repository.CompletionStats, error) {
	args := m.Called(ctx, userID, daysBack)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.CompletionStats), args.Error(1)
}

func (m *MockTaskRepository) GetBumpAnalytics(ctx context.Context, userID string) (*repository.BumpAnalytics, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.BumpAnalytics), args.Error(1)
}

func (m *MockTaskRepository) GetCategoryBreakdown(ctx context.Context, userID string, daysBack int) ([]repository.CategoryStats, error) {
	args := m.Called(ctx, userID, daysBack)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.CategoryStats), args.Error(1)
}

func (m *MockTaskRepository) GetVelocityMetrics(ctx context.Context, userID string, daysBack int) ([]repository.VelocityMetrics, error) {
	args := m.Called(ctx, userID, daysBack)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.VelocityMetrics), args.Error(1)
}

func (m *MockTaskRepository) GetPriorityDistribution(ctx context.Context, userID string) ([]repository.PriorityDistribution, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repository.PriorityDistribution), args.Error(1)
}

// Insights-related methods

func (m *MockTaskRepository) GetCategoryBumpStats(ctx context.Context, userID string) ([]domain.CategoryBumpStats, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.CategoryBumpStats), args.Error(1)
}

func (m *MockTaskRepository) GetCompletionByDayOfWeek(ctx context.Context, userID string, daysBack int) ([]domain.DayOfWeekStats, error) {
	args := m.Called(ctx, userID, daysBack)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.DayOfWeekStats), args.Error(1)
}

func (m *MockTaskRepository) GetAgingQuickWins(ctx context.Context, userID string, minAgeDays int, limit int) ([]*domain.Task, error) {
	args := m.Called(ctx, userID, minAgeDays, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Task), args.Error(1)
}

func (m *MockTaskRepository) GetDeadlineClusters(ctx context.Context, userID string, windowDays int) ([]domain.DeadlineCluster, error) {
	args := m.Called(ctx, userID, windowDays)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.DeadlineCluster), args.Error(1)
}

func (m *MockTaskRepository) GetCompletionTimeStats(ctx context.Context, userID string, category *string, effort *domain.TaskEffort) (*domain.CompletionTimeStats, error) {
	args := m.Called(ctx, userID, category, effort)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CompletionTimeStats), args.Error(1)
}

func (m *MockTaskRepository) GetCategoryDistribution(ctx context.Context, userID string) ([]domain.CategoryDistribution, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.CategoryDistribution), args.Error(1)
}

func (m *MockTaskRepository) GetProductivityHeatmap(ctx context.Context, userID string, daysBack int) (*domain.ProductivityHeatmap, error) {
	args := m.Called(ctx, userID, daysBack)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.ProductivityHeatmap), args.Error(1)
}

func (m *MockTaskRepository) GetCategoryTrends(ctx context.Context, userID string, daysBack int) (*domain.CategoryTrends, error) {
	args := m.Called(ctx, userID, daysBack)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CategoryTrends), args.Error(1)
}

// setupAnalyticsTest creates a test router and mock repository
func setupAnalyticsTest() (*gin.Engine, *MockTaskRepository) {
	router := testutil.SetupTestRouter()
	mockRepo := new(MockTaskRepository)
	return router, mockRepo
}

// =============================================================================
// AnalyticsHandler Constructor Tests
// =============================================================================

func TestNewAnalyticsHandler(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	handler := NewAnalyticsHandler(mockRepo)

	assert.NotNil(t, handler)
}

// =============================================================================
// GetSummary Tests
// =============================================================================

func TestAnalyticsHandler_GetSummary_Success(t *testing.T) {
	router, mockRepo := setupAnalyticsTest()
	handler := NewAnalyticsHandler(mockRepo)

	router.GET("/analytics/summary", testutil.WithAuthContext(router, "user-123", handler.GetSummary))

	// Mock expectations
	mockRepo.On("GetCompletionStats", mock.Anything, "user-123", 30).Return(&repository.CompletionStats{
		TotalTasks:     100,
		CompletedTasks: 75,
		CompletionRate: 0.75,
		PendingTasks:   25,
	}, nil)

	mockRepo.On("GetBumpAnalytics", mock.Anything, "user-123").Return(&repository.BumpAnalytics{
		AverageBumpCount: 1.5,
		TasksByBumpCount: map[int]int{0: 50, 1: 25, 2: 15, 3: 5, 4: 3, 6: 2},
		AtRiskCount:      10,
	}, nil)

	mockRepo.On("GetCategoryBreakdown", mock.Anything, "user-123", 30).Return([]repository.CategoryStats{
		{Category: "Work", TotalCount: 50, CompletedCount: 40, CompletionRate: 0.8},
		{Category: "Personal", TotalCount: 30, CompletedCount: 20, CompletionRate: 0.67},
	}, nil)

	mockRepo.On("GetPriorityDistribution", mock.Anything, "user-123").Return([]repository.PriorityDistribution{
		{Range: "high", Count: 20},
		{Range: "medium", Count: 50},
		{Range: "low", Count: 30},
	}, nil)

	req := httptest.NewRequest("GET", "/analytics/summary", nil)
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)

	// Verify response structure
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(30), response["period_days"])
	assert.NotNil(t, response["completion_stats"])
	assert.NotNil(t, response["bump_analytics"])
	assert.NotNil(t, response["category_breakdown"])
	assert.NotNil(t, response["priority_distribution"])
}

func TestAnalyticsHandler_GetSummary_WithCustomDays(t *testing.T) {
	router, mockRepo := setupAnalyticsTest()
	handler := NewAnalyticsHandler(mockRepo)

	router.GET("/analytics/summary", testutil.WithAuthContext(router, "user-123", handler.GetSummary))

	// Mock expectations with custom 7-day period
	mockRepo.On("GetCompletionStats", mock.Anything, "user-123", 7).Return(&repository.CompletionStats{
		TotalTasks:     10,
		CompletedTasks: 8,
		CompletionRate: 0.8,
		PendingTasks:   2,
	}, nil)

	mockRepo.On("GetBumpAnalytics", mock.Anything, "user-123").Return(&repository.BumpAnalytics{
		AverageBumpCount: 0.5,
		TasksByBumpCount: map[int]int{0: 8, 1: 2},
		AtRiskCount:      0,
	}, nil)

	mockRepo.On("GetCategoryBreakdown", mock.Anything, "user-123", 7).Return([]repository.CategoryStats{}, nil)
	mockRepo.On("GetPriorityDistribution", mock.Anything, "user-123").Return([]repository.PriorityDistribution{}, nil)

	req := httptest.NewRequest("GET", "/analytics/summary?days=7", nil)
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(7), response["period_days"])
}

func TestAnalyticsHandler_GetSummary_InvalidDaysParameter(t *testing.T) {
	router, mockRepo := setupAnalyticsTest()
	handler := NewAnalyticsHandler(mockRepo)

	router.GET("/analytics/summary", testutil.WithAuthContext(router, "user-123", handler.GetSummary))

	// Mock expectations - should default to 30 days for invalid parameter
	mockRepo.On("GetCompletionStats", mock.Anything, "user-123", 30).Return(&repository.CompletionStats{}, nil)
	mockRepo.On("GetBumpAnalytics", mock.Anything, "user-123").Return(&repository.BumpAnalytics{
		TasksByBumpCount: map[int]int{},
	}, nil)
	mockRepo.On("GetCategoryBreakdown", mock.Anything, "user-123", 30).Return([]repository.CategoryStats{}, nil)
	mockRepo.On("GetPriorityDistribution", mock.Anything, "user-123").Return([]repository.PriorityDistribution{}, nil)

	req := httptest.NewRequest("GET", "/analytics/summary?days=invalid", nil)
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestAnalyticsHandler_GetSummary_DaysOutOfRange(t *testing.T) {
	router, mockRepo := setupAnalyticsTest()
	handler := NewAnalyticsHandler(mockRepo)

	router.GET("/analytics/summary", testutil.WithAuthContext(router, "user-123", handler.GetSummary))

	// Mock expectations - days > 365 should default to 30
	mockRepo.On("GetCompletionStats", mock.Anything, "user-123", 30).Return(&repository.CompletionStats{}, nil)
	mockRepo.On("GetBumpAnalytics", mock.Anything, "user-123").Return(&repository.BumpAnalytics{
		TasksByBumpCount: map[int]int{},
	}, nil)
	mockRepo.On("GetCategoryBreakdown", mock.Anything, "user-123", 30).Return([]repository.CategoryStats{}, nil)
	mockRepo.On("GetPriorityDistribution", mock.Anything, "user-123").Return([]repository.PriorityDistribution{}, nil)

	req := httptest.NewRequest("GET", "/analytics/summary?days=500", nil)
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestAnalyticsHandler_GetSummary_DaysBoundaryValues(t *testing.T) {
	// Tests boundary values for days parameter: days > 0 && days <= 365
	testCases := []struct {
		name         string
		daysParam    string
		expectedDays int
	}{
		{"zero", "0", 30},     // 0 is not > 0, defaults to 30
		{"one", "1", 1},       // minimum valid value
		{"max_valid", "365", 365},  // maximum valid value
		{"over_max", "366", 30},    // 366 > 365, defaults to 30
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router, mockRepo := setupAnalyticsTest()
			handler := NewAnalyticsHandler(mockRepo)

			router.GET("/analytics/summary", testutil.WithAuthContext(router, "user-123", handler.GetSummary))

			mockRepo.On("GetCompletionStats", mock.Anything, "user-123", tc.expectedDays).Return(&repository.CompletionStats{}, nil)
			mockRepo.On("GetBumpAnalytics", mock.Anything, "user-123").Return(&repository.BumpAnalytics{
				TasksByBumpCount: map[int]int{},
			}, nil)
			mockRepo.On("GetCategoryBreakdown", mock.Anything, "user-123", tc.expectedDays).Return([]repository.CategoryStats{}, nil)
			mockRepo.On("GetPriorityDistribution", mock.Anything, "user-123").Return([]repository.PriorityDistribution{}, nil)

			req := httptest.NewRequest("GET", "/analytics/summary?days="+tc.daysParam, nil)
			w := testutil.NewResponseRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			mockRepo.AssertExpectations(t)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, float64(tc.expectedDays), response["period_days"])
		})
	}
}

func TestAnalyticsHandler_GetSummary_Unauthenticated(t *testing.T) {
	router, mockRepo := setupAnalyticsTest()
	handler := NewAnalyticsHandler(mockRepo)

	router.GET("/analytics/summary", handler.GetSummary)

	req := httptest.NewRequest("GET", "/analytics/summary", nil)
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockRepo.AssertNotCalled(t, "GetCompletionStats")
}

func TestAnalyticsHandler_GetSummary_CompletionStatsError(t *testing.T) {
	router, mockRepo := setupAnalyticsTest()
	handler := NewAnalyticsHandler(mockRepo)

	router.GET("/analytics/summary", testutil.WithAuthContext(router, "user-123", handler.GetSummary))

	mockRepo.On("GetCompletionStats", mock.Anything, "user-123", 30).
		Return(nil, domain.NewInternalError("database error", nil))

	req := httptest.NewRequest("GET", "/analytics/summary", nil)
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestAnalyticsHandler_GetSummary_BumpAnalyticsError(t *testing.T) {
	router, mockRepo := setupAnalyticsTest()
	handler := NewAnalyticsHandler(mockRepo)

	router.GET("/analytics/summary", testutil.WithAuthContext(router, "user-123", handler.GetSummary))

	mockRepo.On("GetCompletionStats", mock.Anything, "user-123", 30).Return(&repository.CompletionStats{}, nil)
	mockRepo.On("GetBumpAnalytics", mock.Anything, "user-123").
		Return(nil, domain.NewInternalError("database error", nil))

	req := httptest.NewRequest("GET", "/analytics/summary", nil)
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestAnalyticsHandler_GetSummary_CategoryBreakdownError(t *testing.T) {
	router, mockRepo := setupAnalyticsTest()
	handler := NewAnalyticsHandler(mockRepo)

	router.GET("/analytics/summary", testutil.WithAuthContext(router, "user-123", handler.GetSummary))

	mockRepo.On("GetCompletionStats", mock.Anything, "user-123", 30).Return(&repository.CompletionStats{}, nil)
	mockRepo.On("GetBumpAnalytics", mock.Anything, "user-123").Return(&repository.BumpAnalytics{
		TasksByBumpCount: map[int]int{},
	}, nil)
	mockRepo.On("GetCategoryBreakdown", mock.Anything, "user-123", 30).
		Return(nil, domain.NewInternalError("database error", nil))

	req := httptest.NewRequest("GET", "/analytics/summary", nil)
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestAnalyticsHandler_GetSummary_PriorityDistributionError(t *testing.T) {
	router, mockRepo := setupAnalyticsTest()
	handler := NewAnalyticsHandler(mockRepo)

	router.GET("/analytics/summary", testutil.WithAuthContext(router, "user-123", handler.GetSummary))

	mockRepo.On("GetCompletionStats", mock.Anything, "user-123", 30).Return(&repository.CompletionStats{}, nil)
	mockRepo.On("GetBumpAnalytics", mock.Anything, "user-123").Return(&repository.BumpAnalytics{
		TasksByBumpCount: map[int]int{},
	}, nil)
	mockRepo.On("GetCategoryBreakdown", mock.Anything, "user-123", 30).Return([]repository.CategoryStats{}, nil)
	mockRepo.On("GetPriorityDistribution", mock.Anything, "user-123").
		Return(nil, domain.NewInternalError("database error", nil))

	req := httptest.NewRequest("GET", "/analytics/summary", nil)
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestAnalyticsHandler_GetSummary_BumpDistributionMapping(t *testing.T) {
	router, mockRepo := setupAnalyticsTest()
	handler := NewAnalyticsHandler(mockRepo)

	router.GET("/analytics/summary", testutil.WithAuthContext(router, "user-123", handler.GetSummary))

	// Test bump distribution mapping with various bump counts
	mockRepo.On("GetCompletionStats", mock.Anything, "user-123", 30).Return(&repository.CompletionStats{}, nil)
	mockRepo.On("GetBumpAnalytics", mock.Anything, "user-123").Return(&repository.BumpAnalytics{
		AverageBumpCount: 2.5,
		TasksByBumpCount: map[int]int{
			0: 10, // "0 bumps"
			1: 5,  // "1-2 bumps"
			2: 5,  // "1-2 bumps"
			3: 3,  // "3-5 bumps"
			4: 2,  // "3-5 bumps"
			5: 1,  // "3-5 bumps"
			6: 2,  // "6+ bumps"
			7: 1,  // "6+ bumps"
		},
		AtRiskCount: 9,
	}, nil)
	mockRepo.On("GetCategoryBreakdown", mock.Anything, "user-123", 30).Return([]repository.CategoryStats{}, nil)
	mockRepo.On("GetPriorityDistribution", mock.Anything, "user-123").Return([]repository.PriorityDistribution{}, nil)

	req := httptest.NewRequest("GET", "/analytics/summary", nil)
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	bumpAnalytics := response["bump_analytics"].(map[string]interface{})
	bumpDist := bumpAnalytics["bump_distribution"].(map[string]interface{})

	assert.Equal(t, float64(10), bumpDist["0 bumps"])
	assert.Equal(t, float64(10), bumpDist["1-2 bumps"])  // 5 + 5
	assert.Equal(t, float64(6), bumpDist["3-5 bumps"])   // 3 + 2 + 1
	assert.Equal(t, float64(3), bumpDist["6+ bumps"])    // 2 + 1
}

// =============================================================================
// GetTrends Tests
// =============================================================================

func TestAnalyticsHandler_GetTrends_Success(t *testing.T) {
	router, mockRepo := setupAnalyticsTest()
	handler := NewAnalyticsHandler(mockRepo)

	router.GET("/analytics/trends", testutil.WithAuthContext(router, "user-123", handler.GetTrends))

	mockRepo.On("GetVelocityMetrics", mock.Anything, "user-123", 30).Return([]repository.VelocityMetrics{
		{Date: "2025-12-01", CompletedCount: 5},
		{Date: "2025-11-30", CompletedCount: 3},
		{Date: "2025-11-29", CompletedCount: 7},
	}, nil)

	req := httptest.NewRequest("GET", "/analytics/trends", nil)
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(30), response["period_days"])
	assert.NotNil(t, response["velocity_metrics"])
}

func TestAnalyticsHandler_GetTrends_WithCustomDays(t *testing.T) {
	router, mockRepo := setupAnalyticsTest()
	handler := NewAnalyticsHandler(mockRepo)

	router.GET("/analytics/trends", testutil.WithAuthContext(router, "user-123", handler.GetTrends))

	mockRepo.On("GetVelocityMetrics", mock.Anything, "user-123", 14).Return([]repository.VelocityMetrics{
		{Date: "2025-12-01", CompletedCount: 2},
	}, nil)

	req := httptest.NewRequest("GET", "/analytics/trends?days=14", nil)
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(14), response["period_days"])
}

func TestAnalyticsHandler_GetTrends_InvalidDaysParameter(t *testing.T) {
	router, mockRepo := setupAnalyticsTest()
	handler := NewAnalyticsHandler(mockRepo)

	router.GET("/analytics/trends", testutil.WithAuthContext(router, "user-123", handler.GetTrends))

	// Invalid parameter should default to 30 days
	mockRepo.On("GetVelocityMetrics", mock.Anything, "user-123", 30).Return([]repository.VelocityMetrics{}, nil)

	req := httptest.NewRequest("GET", "/analytics/trends?days=abc", nil)
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestAnalyticsHandler_GetTrends_NegativeDays(t *testing.T) {
	router, mockRepo := setupAnalyticsTest()
	handler := NewAnalyticsHandler(mockRepo)

	router.GET("/analytics/trends", testutil.WithAuthContext(router, "user-123", handler.GetTrends))

	// Negative days should default to 30
	mockRepo.On("GetVelocityMetrics", mock.Anything, "user-123", 30).Return([]repository.VelocityMetrics{}, nil)

	req := httptest.NewRequest("GET", "/analytics/trends?days=-5", nil)
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestAnalyticsHandler_GetTrends_DaysBoundaryValues(t *testing.T) {
	// Tests boundary values for days parameter: days > 0 && days <= 365
	testCases := []struct {
		name         string
		daysParam    string
		expectedDays int
	}{
		{"zero", "0", 30},          // 0 is not > 0, defaults to 30
		{"one", "1", 1},            // minimum valid value
		{"max_valid", "365", 365},  // maximum valid value
		{"over_max", "366", 30},    // 366 > 365, defaults to 30
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			router, mockRepo := setupAnalyticsTest()
			handler := NewAnalyticsHandler(mockRepo)

			router.GET("/analytics/trends", testutil.WithAuthContext(router, "user-123", handler.GetTrends))

			mockRepo.On("GetVelocityMetrics", mock.Anything, "user-123", tc.expectedDays).Return([]repository.VelocityMetrics{}, nil)

			req := httptest.NewRequest("GET", "/analytics/trends?days="+tc.daysParam, nil)
			w := testutil.NewResponseRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			mockRepo.AssertExpectations(t)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, float64(tc.expectedDays), response["period_days"])
		})
	}
}

func TestAnalyticsHandler_GetTrends_Unauthenticated(t *testing.T) {
	router, mockRepo := setupAnalyticsTest()
	handler := NewAnalyticsHandler(mockRepo)

	router.GET("/analytics/trends", handler.GetTrends)

	req := httptest.NewRequest("GET", "/analytics/trends", nil)
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockRepo.AssertNotCalled(t, "GetVelocityMetrics")
}

func TestAnalyticsHandler_GetTrends_VelocityMetricsError(t *testing.T) {
	router, mockRepo := setupAnalyticsTest()
	handler := NewAnalyticsHandler(mockRepo)

	router.GET("/analytics/trends", testutil.WithAuthContext(router, "user-123", handler.GetTrends))

	mockRepo.On("GetVelocityMetrics", mock.Anything, "user-123", 30).
		Return(nil, domain.NewInternalError("database error", nil))

	req := httptest.NewRequest("GET", "/analytics/trends", nil)
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockRepo.AssertExpectations(t)
}

func TestAnalyticsHandler_GetTrends_EmptyMetrics(t *testing.T) {
	router, mockRepo := setupAnalyticsTest()
	handler := NewAnalyticsHandler(mockRepo)

	router.GET("/analytics/trends", testutil.WithAuthContext(router, "user-123", handler.GetTrends))

	mockRepo.On("GetVelocityMetrics", mock.Anything, "user-123", 30).Return([]repository.VelocityMetrics{}, nil)

	req := httptest.NewRequest("GET", "/analytics/trends", nil)
	w := testutil.NewResponseRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	metrics := response["velocity_metrics"].([]interface{})
	assert.Empty(t, metrics)
}
