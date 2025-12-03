package service

import (
	"context"

	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/repository"
	"github.com/stretchr/testify/mock"
)

// MockTaskRepository is a mock implementation of ports.TaskRepository
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

func (m *MockTaskRepository) BulkDelete(ctx context.Context, userID string, taskIDs []string) (int, []string, error) {
	args := m.Called(ctx, userID, taskIDs)
	var failedIDs []string
	if args.Get(1) != nil {
		failedIDs = args.Get(1).([]string)
	}
	return args.Int(0), failedIDs, args.Error(2)
}

func (m *MockTaskRepository) BulkUpdateStatus(ctx context.Context, userID string, taskIDs []string, newStatus domain.TaskStatus) (int, []string, error) {
	args := m.Called(ctx, userID, taskIDs, newStatus)
	var failedIDs []string
	if args.Get(1) != nil {
		failedIDs = args.Get(1).([]string)
	}
	return args.Int(0), failedIDs, args.Error(2)
}

// MockTaskHistoryRepository is a mock implementation of ports.TaskHistoryRepository
type MockTaskHistoryRepository struct {
	mock.Mock
}

func (m *MockTaskHistoryRepository) Create(ctx context.Context, history *domain.TaskHistory) error {
	args := m.Called(ctx, history)
	return args.Error(0)
}

func (m *MockTaskHistoryRepository) FindByTaskID(ctx context.Context, taskID string) ([]*domain.TaskHistory, error) {
	args := m.Called(ctx, taskID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.TaskHistory), args.Error(1)
}

// MockUserRepository is a mock implementation of ports.UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}
