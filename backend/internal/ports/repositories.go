package ports

import (
	"context"

	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/repository"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
	EmailExists(ctx context.Context, email string) (bool, error)
}

// TaskRepository defines the interface for task data access
type TaskRepository interface {
	Create(ctx context.Context, task *domain.Task) error
	FindByID(ctx context.Context, id string) (*domain.Task, error)
	List(ctx context.Context, userID string, filter *domain.TaskListFilter) ([]*domain.Task, error)
	Update(ctx context.Context, task *domain.Task) error
	Delete(ctx context.Context, id, userID string) error
	IncrementBumpCount(ctx context.Context, id, userID string) error
	FindAtRiskTasks(ctx context.Context, userID string) ([]*domain.Task, error)
	GetCategories(ctx context.Context, userID string) ([]string, error)
	FindByDateRange(ctx context.Context, userID string, filter *domain.CalendarFilter) ([]*domain.Task, error)
	RenameCategoryForUser(ctx context.Context, userID, oldName, newName string) (int, error)
	DeleteCategoryForUser(ctx context.Context, userID, categoryName string) (int, error)
	// Analytics methods
	GetCompletionStats(ctx context.Context, userID string, daysBack int) (*repository.CompletionStats, error)
	GetBumpAnalytics(ctx context.Context, userID string) (*repository.BumpAnalytics, error)
	GetCategoryBreakdown(ctx context.Context, userID string, daysBack int) ([]repository.CategoryStats, error)
	GetVelocityMetrics(ctx context.Context, userID string, daysBack int) ([]repository.VelocityMetrics, error)
	GetPriorityDistribution(ctx context.Context, userID string) ([]repository.PriorityDistribution, error)
	// Insights analytics methods
	GetCategoryBumpStats(ctx context.Context, userID string) ([]domain.CategoryBumpStats, error)
	GetCompletionByDayOfWeek(ctx context.Context, userID string, daysBack int) ([]domain.DayOfWeekStats, error)
	GetAgingQuickWins(ctx context.Context, userID string, minAgeDays int, limit int) ([]*domain.Task, error)
	GetDeadlineClusters(ctx context.Context, userID string, windowDays int) ([]domain.DeadlineCluster, error)
	GetCompletionTimeStats(ctx context.Context, userID string, category *string, effort *domain.TaskEffort) (*domain.CompletionTimeStats, error)
	GetCategoryDistribution(ctx context.Context, userID string) ([]domain.CategoryDistribution, error)
	// Enhanced analytics methods
	GetProductivityHeatmap(ctx context.Context, userID string, daysBack int) (*domain.ProductivityHeatmap, error)
	GetCategoryTrends(ctx context.Context, userID string, daysBack int) (*domain.CategoryTrends, error)
	// Bulk operations
	BulkDelete(ctx context.Context, userID string, taskIDs []string) (int, []string, error)
	BulkUpdateStatus(ctx context.Context, userID string, taskIDs []string, newStatus domain.TaskStatus) (int, []string, error)
}

// TaskHistoryRepository defines the interface for task history data access
type TaskHistoryRepository interface {
	Create(ctx context.Context, history *domain.TaskHistory) error
	FindByTaskID(ctx context.Context, taskID string) ([]*domain.TaskHistory, error)
}
