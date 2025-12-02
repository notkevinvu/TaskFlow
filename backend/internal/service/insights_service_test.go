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
// Test Helpers
// =============================================================================

func newInsightsService(mockRepo *MockTaskRepository) *InsightsService {
	return NewInsightsService(mockRepo)
}

// =============================================================================
// GetInsights Tests
// =============================================================================

func TestInsightsService_GetInsights_ReturnsEmptyWhenNoPatterns(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := newInsightsService(mockRepo)

	// All queries return empty/no matches
	mockRepo.On("GetCategoryBumpStats", mock.Anything, "user-123").Return([]domain.CategoryBumpStats{}, nil)
	mockRepo.On("GetCompletionByDayOfWeek", mock.Anything, "user-123", 90).Return([]domain.DayOfWeekStats{}, nil)
	mockRepo.On("GetAgingQuickWins", mock.Anything, "user-123", 5, 10).Return([]*domain.Task{}, nil)
	mockRepo.On("GetDeadlineClusters", mock.Anything, "user-123", 14).Return([]domain.DeadlineCluster{}, nil)
	mockRepo.On("FindAtRiskTasks", mock.Anything, "user-123").Return([]*domain.Task{}, nil)
	mockRepo.On("GetCategoryDistribution", mock.Anything, "user-123").Return([]domain.CategoryDistribution{}, nil)

	response, err := service.GetInsights(context.Background(), "user-123")

	require.NoError(t, err)
	assert.NotNil(t, response)
	assert.Empty(t, response.Insights)
	assert.False(t, response.CachedAt.IsZero())
	mockRepo.AssertExpectations(t)
}

func TestInsightsService_GetInsights_SortsByPriority(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := newInsightsService(mockRepo)

	// Setup: At-risk (Urgent priority) and Category overload (Medium priority)
	mockRepo.On("GetCategoryBumpStats", mock.Anything, "user-123").Return([]domain.CategoryBumpStats{}, nil)
	mockRepo.On("GetCompletionByDayOfWeek", mock.Anything, "user-123", 90).Return([]domain.DayOfWeekStats{}, nil)
	mockRepo.On("GetAgingQuickWins", mock.Anything, "user-123", 5, 10).Return([]*domain.Task{}, nil)
	mockRepo.On("GetDeadlineClusters", mock.Anything, "user-123", 14).Return([]domain.DeadlineCluster{}, nil)

	// At-risk task (Urgent priority = 4)
	atRiskTask := createTestTask("user-123", "task-1")
	atRiskTask.BumpCount = 5
	mockRepo.On("FindAtRiskTasks", mock.Anything, "user-123").Return([]*domain.Task{atRiskTask}, nil)

	// Category overload (Medium priority = 2)
	mockRepo.On("GetCategoryDistribution", mock.Anything, "user-123").Return([]domain.CategoryDistribution{
		{Category: "Work", TaskCount: 10, Percentage: 60},
	}, nil)

	response, err := service.GetInsights(context.Background(), "user-123")

	require.NoError(t, err)
	require.Len(t, response.Insights, 2)
	// Higher priority should come first
	assert.Equal(t, domain.InsightAtRiskAlert, response.Insights[0].Type)
	assert.Equal(t, domain.InsightCategoryOverload, response.Insights[1].Type)
	mockRepo.AssertExpectations(t)
}

func TestInsightsService_GetInsights_ContinuesOnIndividualErrors(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := newInsightsService(mockRepo)

	// First query fails, others succeed
	mockRepo.On("GetCategoryBumpStats", mock.Anything, "user-123").Return(nil, errors.New("db error"))
	mockRepo.On("GetCompletionByDayOfWeek", mock.Anything, "user-123", 90).Return([]domain.DayOfWeekStats{}, nil)
	mockRepo.On("GetAgingQuickWins", mock.Anything, "user-123", 5, 10).Return([]*domain.Task{}, nil)
	mockRepo.On("GetDeadlineClusters", mock.Anything, "user-123", 14).Return([]domain.DeadlineCluster{}, nil)
	mockRepo.On("FindAtRiskTasks", mock.Anything, "user-123").Return([]*domain.Task{}, nil)
	mockRepo.On("GetCategoryDistribution", mock.Anything, "user-123").Return([]domain.CategoryDistribution{
		{Category: "Work", TaskCount: 10, Percentage: 60},
	}, nil)

	response, err := service.GetInsights(context.Background(), "user-123")

	require.NoError(t, err)
	require.Len(t, response.Insights, 1)
	assert.Equal(t, domain.InsightCategoryOverload, response.Insights[0].Type)
	mockRepo.AssertExpectations(t)
}

// =============================================================================
// checkAvoidancePattern Tests
// =============================================================================

func TestInsightsService_CheckAvoidancePattern_DetectsHighBumpCategory(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := newInsightsService(mockRepo)

	mockRepo.On("GetCategoryBumpStats", mock.Anything, "user-123").Return([]domain.CategoryBumpStats{
		{Category: "Documentation", TaskCount: 5, AvgBumps: 3.5},
	}, nil)

	// Call the private method through GetInsights
	mockRepo.On("GetCompletionByDayOfWeek", mock.Anything, "user-123", 90).Return([]domain.DayOfWeekStats{}, nil)
	mockRepo.On("GetAgingQuickWins", mock.Anything, "user-123", 5, 10).Return([]*domain.Task{}, nil)
	mockRepo.On("GetDeadlineClusters", mock.Anything, "user-123", 14).Return([]domain.DeadlineCluster{}, nil)
	mockRepo.On("FindAtRiskTasks", mock.Anything, "user-123").Return([]*domain.Task{}, nil)
	mockRepo.On("GetCategoryDistribution", mock.Anything, "user-123").Return([]domain.CategoryDistribution{}, nil)

	response, err := service.GetInsights(context.Background(), "user-123")

	require.NoError(t, err)
	require.Len(t, response.Insights, 1)
	insight := response.Insights[0]
	assert.Equal(t, domain.InsightAvoidancePattern, insight.Type)
	assert.Equal(t, "Avoidance Pattern Detected", insight.Title)
	assert.Contains(t, insight.Message, "Documentation")
	assert.Equal(t, domain.InsightPriorityHigh, insight.Priority)
	assert.Equal(t, "Documentation", insight.Data["category"])
	mockRepo.AssertExpectations(t)
}

func TestInsightsService_CheckAvoidancePattern_IgnoresBelowThreshold(t *testing.T) {
	tests := []struct {
		name      string
		stats     []domain.CategoryBumpStats
		wantMatch bool
	}{
		{
			name:      "avg bumps below 2.5",
			stats:     []domain.CategoryBumpStats{{Category: "Work", TaskCount: 5, AvgBumps: 2.4}},
			wantMatch: false,
		},
		{
			name:      "task count below 2",
			stats:     []domain.CategoryBumpStats{{Category: "Work", TaskCount: 1, AvgBumps: 3.0}},
			wantMatch: false,
		},
		{
			name:      "exactly at threshold",
			stats:     []domain.CategoryBumpStats{{Category: "Work", TaskCount: 2, AvgBumps: 2.5}},
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTaskRepository)
			service := newInsightsService(mockRepo)

			mockRepo.On("GetCategoryBumpStats", mock.Anything, "user-123").Return(tt.stats, nil)
			mockRepo.On("GetCompletionByDayOfWeek", mock.Anything, "user-123", 90).Return([]domain.DayOfWeekStats{}, nil)
			mockRepo.On("GetAgingQuickWins", mock.Anything, "user-123", 5, 10).Return([]*domain.Task{}, nil)
			mockRepo.On("GetDeadlineClusters", mock.Anything, "user-123", 14).Return([]domain.DeadlineCluster{}, nil)
			mockRepo.On("FindAtRiskTasks", mock.Anything, "user-123").Return([]*domain.Task{}, nil)
			mockRepo.On("GetCategoryDistribution", mock.Anything, "user-123").Return([]domain.CategoryDistribution{}, nil)

			response, err := service.GetInsights(context.Background(), "user-123")

			require.NoError(t, err)
			hasAvoidance := false
			for _, insight := range response.Insights {
				if insight.Type == domain.InsightAvoidancePattern {
					hasAvoidance = true
				}
			}
			assert.Equal(t, tt.wantMatch, hasAvoidance)
		})
	}
}

// =============================================================================
// checkPeakPerformance Tests
// =============================================================================

func TestInsightsService_CheckPeakPerformance_IdentifiesBestDay(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := newInsightsService(mockRepo)

	mockRepo.On("GetCategoryBumpStats", mock.Anything, "user-123").Return([]domain.CategoryBumpStats{}, nil)
	mockRepo.On("GetCompletionByDayOfWeek", mock.Anything, "user-123", 90).Return([]domain.DayOfWeekStats{
		{DayOfWeek: 1, CompletedCount: 15}, // Monday (highest)
		{DayOfWeek: 3, CompletedCount: 10}, // Wednesday
	}, nil)
	mockRepo.On("GetAgingQuickWins", mock.Anything, "user-123", 5, 10).Return([]*domain.Task{}, nil)
	mockRepo.On("GetDeadlineClusters", mock.Anything, "user-123", 14).Return([]domain.DeadlineCluster{}, nil)
	mockRepo.On("FindAtRiskTasks", mock.Anything, "user-123").Return([]*domain.Task{}, nil)
	mockRepo.On("GetCategoryDistribution", mock.Anything, "user-123").Return([]domain.CategoryDistribution{}, nil)

	response, err := service.GetInsights(context.Background(), "user-123")

	require.NoError(t, err)
	require.Len(t, response.Insights, 1)
	insight := response.Insights[0]
	assert.Equal(t, domain.InsightPeakPerformance, insight.Type)
	assert.Contains(t, insight.Message, "Monday")
	assert.Equal(t, 1, insight.Data["day_of_week"])
	assert.Equal(t, "Monday", insight.Data["day_name"])
}

func TestInsightsService_CheckPeakPerformance_RequiresMinimumCompletions(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := newInsightsService(mockRepo)

	mockRepo.On("GetCategoryBumpStats", mock.Anything, "user-123").Return([]domain.CategoryBumpStats{}, nil)
	mockRepo.On("GetCompletionByDayOfWeek", mock.Anything, "user-123", 90).Return([]domain.DayOfWeekStats{
		{DayOfWeek: 1, CompletedCount: 4}, // Below threshold of 5
	}, nil)
	mockRepo.On("GetAgingQuickWins", mock.Anything, "user-123", 5, 10).Return([]*domain.Task{}, nil)
	mockRepo.On("GetDeadlineClusters", mock.Anything, "user-123", 14).Return([]domain.DeadlineCluster{}, nil)
	mockRepo.On("FindAtRiskTasks", mock.Anything, "user-123").Return([]*domain.Task{}, nil)
	mockRepo.On("GetCategoryDistribution", mock.Anything, "user-123").Return([]domain.CategoryDistribution{}, nil)

	response, err := service.GetInsights(context.Background(), "user-123")

	require.NoError(t, err)
	assert.Empty(t, response.Insights)
}

func TestInsightsService_CheckPeakPerformance_HandlesInvalidDayOfWeek(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := newInsightsService(mockRepo)

	mockRepo.On("GetCategoryBumpStats", mock.Anything, "user-123").Return([]domain.CategoryBumpStats{}, nil)
	mockRepo.On("GetCompletionByDayOfWeek", mock.Anything, "user-123", 90).Return([]domain.DayOfWeekStats{
		{DayOfWeek: 7, CompletedCount: 10}, // Invalid: 7 is out of 0-6 range
	}, nil)
	mockRepo.On("GetAgingQuickWins", mock.Anything, "user-123", 5, 10).Return([]*domain.Task{}, nil)
	mockRepo.On("GetDeadlineClusters", mock.Anything, "user-123", 14).Return([]domain.DeadlineCluster{}, nil)
	mockRepo.On("FindAtRiskTasks", mock.Anything, "user-123").Return([]*domain.Task{}, nil)
	mockRepo.On("GetCategoryDistribution", mock.Anything, "user-123").Return([]domain.CategoryDistribution{}, nil)

	response, err := service.GetInsights(context.Background(), "user-123")

	require.NoError(t, err)
	// Should not panic, should return no peak performance insight
	hasPeak := false
	for _, insight := range response.Insights {
		if insight.Type == domain.InsightPeakPerformance {
			hasPeak = true
		}
	}
	assert.False(t, hasPeak)
}

// =============================================================================
// checkQuickWins Tests
// =============================================================================

func TestInsightsService_CheckQuickWins_DetectsAgingSmallTasks(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := newInsightsService(mockRepo)

	task1 := createTestTask("user-123", "task-1")
	task1.Title = "Quick task 1"
	task2 := createTestTask("user-123", "task-2")
	task2.Title = "Quick task 2"

	mockRepo.On("GetCategoryBumpStats", mock.Anything, "user-123").Return([]domain.CategoryBumpStats{}, nil)
	mockRepo.On("GetCompletionByDayOfWeek", mock.Anything, "user-123", 90).Return([]domain.DayOfWeekStats{}, nil)
	mockRepo.On("GetAgingQuickWins", mock.Anything, "user-123", 5, 10).Return([]*domain.Task{task1, task2}, nil)
	mockRepo.On("GetDeadlineClusters", mock.Anything, "user-123", 14).Return([]domain.DeadlineCluster{}, nil)
	mockRepo.On("FindAtRiskTasks", mock.Anything, "user-123").Return([]*domain.Task{}, nil)
	mockRepo.On("GetCategoryDistribution", mock.Anything, "user-123").Return([]domain.CategoryDistribution{}, nil)

	response, err := service.GetInsights(context.Background(), "user-123")

	require.NoError(t, err)
	require.Len(t, response.Insights, 1)
	insight := response.Insights[0]
	assert.Equal(t, domain.InsightQuickWins, insight.Type)
	assert.Contains(t, insight.Message, "2 quick wins")
	assert.NotNil(t, insight.ActionURL)
}

func TestInsightsService_CheckQuickWins_RequiresMinimumCount(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := newInsightsService(mockRepo)

	task1 := createTestTask("user-123", "task-1")

	mockRepo.On("GetCategoryBumpStats", mock.Anything, "user-123").Return([]domain.CategoryBumpStats{}, nil)
	mockRepo.On("GetCompletionByDayOfWeek", mock.Anything, "user-123", 90).Return([]domain.DayOfWeekStats{}, nil)
	mockRepo.On("GetAgingQuickWins", mock.Anything, "user-123", 5, 10).Return([]*domain.Task{task1}, nil) // Only 1 task
	mockRepo.On("GetDeadlineClusters", mock.Anything, "user-123", 14).Return([]domain.DeadlineCluster{}, nil)
	mockRepo.On("FindAtRiskTasks", mock.Anything, "user-123").Return([]*domain.Task{}, nil)
	mockRepo.On("GetCategoryDistribution", mock.Anything, "user-123").Return([]domain.CategoryDistribution{}, nil)

	response, err := service.GetInsights(context.Background(), "user-123")

	require.NoError(t, err)
	assert.Empty(t, response.Insights)
}

// =============================================================================
// checkDeadlineClustering Tests
// =============================================================================

func TestInsightsService_CheckDeadlineClustering_DetectsClusters(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := newInsightsService(mockRepo)

	dueDate := time.Now().Add(3 * 24 * time.Hour)

	mockRepo.On("GetCategoryBumpStats", mock.Anything, "user-123").Return([]domain.CategoryBumpStats{}, nil)
	mockRepo.On("GetCompletionByDayOfWeek", mock.Anything, "user-123", 90).Return([]domain.DayOfWeekStats{}, nil)
	mockRepo.On("GetAgingQuickWins", mock.Anything, "user-123", 5, 10).Return([]*domain.Task{}, nil)
	mockRepo.On("GetDeadlineClusters", mock.Anything, "user-123", 14).Return([]domain.DeadlineCluster{
		{DueDate: dueDate, TaskCount: 4, Titles: []string{"Task 1", "Task 2", "Task 3", "Task 4"}},
	}, nil)
	mockRepo.On("FindAtRiskTasks", mock.Anything, "user-123").Return([]*domain.Task{}, nil)
	mockRepo.On("GetCategoryDistribution", mock.Anything, "user-123").Return([]domain.CategoryDistribution{}, nil)

	response, err := service.GetInsights(context.Background(), "user-123")

	require.NoError(t, err)
	require.Len(t, response.Insights, 1)
	insight := response.Insights[0]
	assert.Equal(t, domain.InsightDeadlineClustering, insight.Type)
	assert.Contains(t, insight.Message, "4 tasks")
	assert.Equal(t, domain.InsightPriorityHigh, insight.Priority)
}

// =============================================================================
// checkAtRiskTasks Tests
// =============================================================================

func TestInsightsService_CheckAtRiskTasks_FindsHighestBumpTask(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := newInsightsService(mockRepo)

	task1 := createTestTask("user-123", "task-1")
	task1.BumpCount = 3
	task1.Title = "Low bump task"

	task2 := createTestTask("user-123", "task-2")
	task2.BumpCount = 7
	task2.Title = "High bump task"

	mockRepo.On("GetCategoryBumpStats", mock.Anything, "user-123").Return([]domain.CategoryBumpStats{}, nil)
	mockRepo.On("GetCompletionByDayOfWeek", mock.Anything, "user-123", 90).Return([]domain.DayOfWeekStats{}, nil)
	mockRepo.On("GetAgingQuickWins", mock.Anything, "user-123", 5, 10).Return([]*domain.Task{}, nil)
	mockRepo.On("GetDeadlineClusters", mock.Anything, "user-123", 14).Return([]domain.DeadlineCluster{}, nil)
	mockRepo.On("FindAtRiskTasks", mock.Anything, "user-123").Return([]*domain.Task{task1, task2}, nil)
	mockRepo.On("GetCategoryDistribution", mock.Anything, "user-123").Return([]domain.CategoryDistribution{}, nil)

	response, err := service.GetInsights(context.Background(), "user-123")

	require.NoError(t, err)
	require.Len(t, response.Insights, 1)
	insight := response.Insights[0]
	assert.Equal(t, domain.InsightAtRiskAlert, insight.Type)
	assert.Contains(t, insight.Message, "High bump task")
	assert.Contains(t, insight.Message, "7 times")
	assert.Equal(t, domain.InsightPriorityUrgent, insight.Priority)
	assert.Equal(t, 7, insight.Data["bump_count"])
}

func TestInsightsService_CheckAtRiskTasks_RequiresMinimumBumps(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := newInsightsService(mockRepo)

	task := createTestTask("user-123", "task-1")
	task.BumpCount = 2 // Below threshold of 3

	mockRepo.On("GetCategoryBumpStats", mock.Anything, "user-123").Return([]domain.CategoryBumpStats{}, nil)
	mockRepo.On("GetCompletionByDayOfWeek", mock.Anything, "user-123", 90).Return([]domain.DayOfWeekStats{}, nil)
	mockRepo.On("GetAgingQuickWins", mock.Anything, "user-123", 5, 10).Return([]*domain.Task{}, nil)
	mockRepo.On("GetDeadlineClusters", mock.Anything, "user-123", 14).Return([]domain.DeadlineCluster{}, nil)
	mockRepo.On("FindAtRiskTasks", mock.Anything, "user-123").Return([]*domain.Task{task}, nil)
	mockRepo.On("GetCategoryDistribution", mock.Anything, "user-123").Return([]domain.CategoryDistribution{}, nil)

	response, err := service.GetInsights(context.Background(), "user-123")

	require.NoError(t, err)
	assert.Empty(t, response.Insights)
}

// =============================================================================
// checkCategoryOverload Tests
// =============================================================================

func TestInsightsService_CheckCategoryOverload_DetectsDominantCategory(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := newInsightsService(mockRepo)

	mockRepo.On("GetCategoryBumpStats", mock.Anything, "user-123").Return([]domain.CategoryBumpStats{}, nil)
	mockRepo.On("GetCompletionByDayOfWeek", mock.Anything, "user-123", 90).Return([]domain.DayOfWeekStats{}, nil)
	mockRepo.On("GetAgingQuickWins", mock.Anything, "user-123", 5, 10).Return([]*domain.Task{}, nil)
	mockRepo.On("GetDeadlineClusters", mock.Anything, "user-123", 14).Return([]domain.DeadlineCluster{}, nil)
	mockRepo.On("FindAtRiskTasks", mock.Anything, "user-123").Return([]*domain.Task{}, nil)
	mockRepo.On("GetCategoryDistribution", mock.Anything, "user-123").Return([]domain.CategoryDistribution{
		{Category: "Bug Fix", TaskCount: 8, Percentage: 55},
	}, nil)

	response, err := service.GetInsights(context.Background(), "user-123")

	require.NoError(t, err)
	require.Len(t, response.Insights, 1)
	insight := response.Insights[0]
	assert.Equal(t, domain.InsightCategoryOverload, insight.Type)
	assert.Contains(t, insight.Message, "Bug Fix")
	assert.Contains(t, insight.Message, "55%")
}

func TestInsightsService_CheckCategoryOverload_RequiresThresholds(t *testing.T) {
	tests := []struct {
		name      string
		dist      []domain.CategoryDistribution
		wantMatch bool
	}{
		{
			name:      "below 40% threshold",
			dist:      []domain.CategoryDistribution{{Category: "Work", TaskCount: 10, Percentage: 35}},
			wantMatch: false,
		},
		{
			name:      "below 5 tasks threshold",
			dist:      []domain.CategoryDistribution{{Category: "Work", TaskCount: 4, Percentage: 50}},
			wantMatch: false,
		},
		{
			name:      "exactly at both thresholds",
			dist:      []domain.CategoryDistribution{{Category: "Work", TaskCount: 5, Percentage: 40}},
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTaskRepository)
			service := newInsightsService(mockRepo)

			mockRepo.On("GetCategoryBumpStats", mock.Anything, "user-123").Return([]domain.CategoryBumpStats{}, nil)
			mockRepo.On("GetCompletionByDayOfWeek", mock.Anything, "user-123", 90).Return([]domain.DayOfWeekStats{}, nil)
			mockRepo.On("GetAgingQuickWins", mock.Anything, "user-123", 5, 10).Return([]*domain.Task{}, nil)
			mockRepo.On("GetDeadlineClusters", mock.Anything, "user-123", 14).Return([]domain.DeadlineCluster{}, nil)
			mockRepo.On("FindAtRiskTasks", mock.Anything, "user-123").Return([]*domain.Task{}, nil)
			mockRepo.On("GetCategoryDistribution", mock.Anything, "user-123").Return(tt.dist, nil)

			response, err := service.GetInsights(context.Background(), "user-123")

			require.NoError(t, err)
			hasOverload := false
			for _, insight := range response.Insights {
				if insight.Type == domain.InsightCategoryOverload {
					hasOverload = true
				}
			}
			assert.Equal(t, tt.wantMatch, hasOverload)
		})
	}
}

// =============================================================================
// EstimateCompletionTime Tests
// =============================================================================

func TestInsightsService_EstimateCompletionTime_CalculatesCorrectEstimate(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := newInsightsService(mockRepo)

	category := "Work"
	effort := domain.TaskEffortMedium
	task := &domain.Task{
		ID:              "task-1",
		UserID:          "user-123",
		Category:        &category,
		EstimatedEffort: &effort,
		BumpCount:       0,
	}

	// Base stats: median 5 days
	mockRepo.On("GetCompletionTimeStats", mock.Anything, "user-123", (*string)(nil), (*domain.TaskEffort)(nil)).
		Return(&domain.CompletionTimeStats{MedianDays: 5.0, AvgDays: 5.5, SampleSize: 25}, nil)

	// Category stats: median 4 days (faster category)
	mockRepo.On("GetCompletionTimeStats", mock.Anything, "user-123", &category, (*domain.TaskEffort)(nil)).
		Return(&domain.CompletionTimeStats{MedianDays: 4.0, AvgDays: 4.2, SampleSize: 10}, nil)

	estimate, err := service.EstimateCompletionTime(context.Background(), "user-123", task)

	require.NoError(t, err)
	assert.NotNil(t, estimate)
	// Base=5, CategoryFactor=4/5=0.8, EffortFactor=1.0, BumpFactor=1.0
	// Estimate = 5 * 0.8 * 1.0 * 1.0 = 4.0
	assert.InDelta(t, 4.0, estimate.EstimatedDays, 0.01)
	assert.Equal(t, "high", estimate.ConfidenceLevel) // 25 samples >= 20
	assert.Equal(t, 25, estimate.BasedOn)
}

func TestInsightsService_EstimateCompletionTime_AppliesEffortFactor(t *testing.T) {
	tests := []struct {
		effort         domain.TaskEffort
		expectedFactor float64
	}{
		{domain.TaskEffortSmall, 0.5},
		{domain.TaskEffortMedium, 1.0},
		{domain.TaskEffortLarge, 2.0},
		{domain.TaskEffortXLarge, 3.5},
	}

	for _, tt := range tests {
		t.Run(string(tt.effort), func(t *testing.T) {
			mockRepo := new(MockTaskRepository)
			service := newInsightsService(mockRepo)

			task := &domain.Task{
				ID:              "task-1",
				UserID:          "user-123",
				EstimatedEffort: &tt.effort,
				BumpCount:       0,
			}

			mockRepo.On("GetCompletionTimeStats", mock.Anything, "user-123", (*string)(nil), (*domain.TaskEffort)(nil)).
				Return(&domain.CompletionTimeStats{MedianDays: 4.0, SampleSize: 10}, nil)

			estimate, err := service.EstimateCompletionTime(context.Background(), "user-123", task)

			require.NoError(t, err)
			// Base=4, no category, EffortFactor=tt.expectedFactor, BumpFactor=1.0
			expectedDays := 4.0 * tt.expectedFactor
			assert.InDelta(t, expectedDays, estimate.EstimatedDays, 0.01)
			assert.Equal(t, tt.expectedFactor, estimate.Factors.EffortFactor)
		})
	}
}

func TestInsightsService_EstimateCompletionTime_AppliesBumpFactor(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := newInsightsService(mockRepo)

	task := &domain.Task{
		ID:        "task-1",
		UserID:    "user-123",
		BumpCount: 3, // 3 bumps = 1 + (0.2 * 3) = 1.6 factor
	}

	mockRepo.On("GetCompletionTimeStats", mock.Anything, "user-123", (*string)(nil), (*domain.TaskEffort)(nil)).
		Return(&domain.CompletionTimeStats{MedianDays: 5.0, SampleSize: 10}, nil)

	estimate, err := service.EstimateCompletionTime(context.Background(), "user-123", task)

	require.NoError(t, err)
	// Base=5, EffortFactor=1.0, BumpFactor=1.6
	assert.InDelta(t, 8.0, estimate.EstimatedDays, 0.01)
	assert.InDelta(t, 1.6, estimate.Factors.BumpFactor, 0.01)
}

func TestInsightsService_EstimateCompletionTime_UsesDefaultWhenNoData(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := newInsightsService(mockRepo)

	task := &domain.Task{
		ID:        "task-1",
		UserID:    "user-123",
		BumpCount: 0,
	}

	mockRepo.On("GetCompletionTimeStats", mock.Anything, "user-123", (*string)(nil), (*domain.TaskEffort)(nil)).
		Return(&domain.CompletionTimeStats{MedianDays: 0, SampleSize: 0}, nil)

	estimate, err := service.EstimateCompletionTime(context.Background(), "user-123", task)

	require.NoError(t, err)
	// Should use default of 3.0 days
	assert.InDelta(t, 3.0, estimate.EstimatedDays, 0.01)
	assert.Equal(t, "low", estimate.ConfidenceLevel) // 0 samples < 5
}

func TestInsightsService_EstimateCompletionTime_DeterminesConfidenceLevel(t *testing.T) {
	tests := []struct {
		sampleSize int
		expected   string
	}{
		{0, "low"},
		{4, "low"},
		{5, "medium"},
		{19, "medium"},
		{20, "high"},
		{100, "high"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			mockRepo := new(MockTaskRepository)
			service := newInsightsService(mockRepo)

			task := &domain.Task{ID: "task-1", UserID: "user-123"}

			mockRepo.On("GetCompletionTimeStats", mock.Anything, "user-123", (*string)(nil), (*domain.TaskEffort)(nil)).
				Return(&domain.CompletionTimeStats{MedianDays: 5.0, SampleSize: tt.sampleSize}, nil)

			estimate, err := service.EstimateCompletionTime(context.Background(), "user-123", task)

			require.NoError(t, err)
			assert.Equal(t, tt.expected, estimate.ConfidenceLevel)
		})
	}
}

func TestInsightsService_EstimateCompletionTime_ReturnsErrorOnRepoFailure(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := newInsightsService(mockRepo)

	task := &domain.Task{ID: "task-1", UserID: "user-123"}

	mockRepo.On("GetCompletionTimeStats", mock.Anything, "user-123", (*string)(nil), (*domain.TaskEffort)(nil)).
		Return(nil, errors.New("db error"))

	estimate, err := service.EstimateCompletionTime(context.Background(), "user-123", task)

	assert.Error(t, err)
	assert.Nil(t, estimate)
}

// =============================================================================
// SuggestCategory Tests
// =============================================================================

func TestInsightsService_SuggestCategory_MatchesKnownPatterns(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := newInsightsService(mockRepo)

	mockRepo.On("GetCategories", mock.Anything, "user-123").Return([]string{}, nil)

	req := &domain.CategorySuggestionRequest{
		Title:       "Review PR for authentication module",
		Description: "Need to check the code changes",
	}

	response, err := service.SuggestCategory(context.Background(), "user-123", req)

	require.NoError(t, err)
	require.NotEmpty(t, response.Suggestions)

	// Should match "Code Review" pattern (contains "review", "pr", "code")
	foundCodeReview := false
	for _, s := range response.Suggestions {
		if s.Category == "Code Review" {
			foundCodeReview = true
			assert.True(t, s.Confidence > 0)
			assert.NotEmpty(t, s.MatchedKeywords)
		}
	}
	assert.True(t, foundCodeReview, "Should suggest Code Review category")
}

func TestInsightsService_SuggestCategory_BoostsExistingCategories(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := newInsightsService(mockRepo)

	// User has existing "frontend" category
	mockRepo.On("GetCategories", mock.Anything, "user-123").Return([]string{"Frontend"}, nil)

	req := &domain.CategorySuggestionRequest{
		Title:       "Fix frontend bug in login page",
		Description: "",
	}

	response, err := service.SuggestCategory(context.Background(), "user-123", req)

	require.NoError(t, err)
	require.NotEmpty(t, response.Suggestions)

	// Should have "Bug Fix" and "Frontend" as suggestions
	var bugFixConfidence, frontendConfidence float64
	for _, s := range response.Suggestions {
		if s.Category == "Bug Fix" {
			bugFixConfidence = s.Confidence
		}
		if s.Category == "Frontend" {
			frontendConfidence = s.Confidence
		}
	}

	assert.True(t, bugFixConfidence > 0, "Should suggest Bug Fix from keyword matching")
	assert.True(t, frontendConfidence > 0, "Should suggest Frontend from existing categories")
}

func TestInsightsService_SuggestCategory_LimitsToTopThree(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := newInsightsService(mockRepo)

	mockRepo.On("GetCategories", mock.Anything, "user-123").Return([]string{}, nil)

	// Request with multiple keyword matches
	req := &domain.CategorySuggestionRequest{
		Title:       "Review PR and fix bug in documentation for meeting notes deploy pipeline",
		Description: "Also test the research spike",
	}

	response, err := service.SuggestCategory(context.Background(), "user-123", req)

	require.NoError(t, err)
	assert.LessOrEqual(t, len(response.Suggestions), 3)
}

func TestInsightsService_SuggestCategory_CalculatesConfidence(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := newInsightsService(mockRepo)

	mockRepo.On("GetCategories", mock.Anything, "user-123").Return([]string{}, nil)

	req := &domain.CategorySuggestionRequest{
		Title:       "Fix the bug causing crash error",
		Description: "",
	}

	response, err := service.SuggestCategory(context.Background(), "user-123", req)

	require.NoError(t, err)

	// Bug Fix pattern has keywords: bug, fix, issue, error, crash, broken
	// Match: bug, fix, crash, error = 4 out of 6 = 0.666...
	for _, s := range response.Suggestions {
		if s.Category == "Bug Fix" {
			assert.InDelta(t, 0.666, s.Confidence, 0.01)
		}
	}
}

func TestInsightsService_SuggestCategory_ClampsConfidenceToMax(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := newInsightsService(mockRepo)

	// User has existing "bugfix" category that will match and get boosted
	mockRepo.On("GetCategories", mock.Anything, "user-123").Return([]string{"Bug Fix"}, nil)

	// All keywords match for Bug Fix
	req := &domain.CategorySuggestionRequest{
		Title:       "bug fix issue error crash broken",
		Description: "",
	}

	response, err := service.SuggestCategory(context.Background(), "user-123", req)

	require.NoError(t, err)

	for _, s := range response.Suggestions {
		assert.LessOrEqual(t, s.Confidence, 1.0, "Confidence should be clamped to 1.0")
	}
}

func TestInsightsService_SuggestCategory_HandlesGetCategoriesError(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := newInsightsService(mockRepo)

	// GetCategories fails but should not fail the whole operation
	mockRepo.On("GetCategories", mock.Anything, "user-123").Return(nil, errors.New("db error"))

	req := &domain.CategorySuggestionRequest{
		Title:       "Review code changes",
		Description: "",
	}

	response, err := service.SuggestCategory(context.Background(), "user-123", req)

	require.NoError(t, err) // Should not error
	require.NotEmpty(t, response.Suggestions) // Should still have pattern-based suggestions
}

func TestInsightsService_SuggestCategory_ReturnsEmptyForNoMatches(t *testing.T) {
	mockRepo := new(MockTaskRepository)
	service := newInsightsService(mockRepo)

	mockRepo.On("GetCategories", mock.Anything, "user-123").Return([]string{}, nil)

	req := &domain.CategorySuggestionRequest{
		Title:       "Random task with no keywords",
		Description: "",
	}

	response, err := service.SuggestCategory(context.Background(), "user-123", req)

	require.NoError(t, err)
	assert.Empty(t, response.Suggestions)
}
