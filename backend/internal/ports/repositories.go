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
	// Subtask operations
	GetSubtasks(ctx context.Context, parentTaskID string) ([]*domain.Task, error)
	GetSubtaskInfo(ctx context.Context, parentTaskID string) (*domain.SubtaskInfo, error)
	CountIncompleteSubtasks(ctx context.Context, parentTaskID string) (int, error)
}

// TaskHistoryRepository defines the interface for task history data access
type TaskHistoryRepository interface {
	Create(ctx context.Context, history *domain.TaskHistory) error
	FindByTaskID(ctx context.Context, taskID string) ([]*domain.TaskHistory, error)
}

// TaskSeriesRepository defines the interface for task series data access
type TaskSeriesRepository interface {
	Create(ctx context.Context, series *domain.TaskSeries) error
	FindByID(ctx context.Context, id string) (*domain.TaskSeries, error)
	FindByUserID(ctx context.Context, userID string) ([]*domain.TaskSeries, error)
	FindActiveByUserID(ctx context.Context, userID string) ([]*domain.TaskSeries, error)
	Update(ctx context.Context, series *domain.TaskSeries) error
	Deactivate(ctx context.Context, id, userID string) error
	Delete(ctx context.Context, id, userID string) error
	CountTasksInSeries(ctx context.Context, seriesID string) (int, error)
	GetTasksBySeriesID(ctx context.Context, seriesID string) ([]*domain.Task, error)
}

// UserPreferencesRepository defines the interface for user preferences data access
type UserPreferencesRepository interface {
	GetUserPreferences(ctx context.Context, userID string) (*domain.UserPreferences, error)
	UpsertUserPreferences(ctx context.Context, prefs *domain.UserPreferences) error
	DeleteUserPreferences(ctx context.Context, userID string) error
	GetCategoryPreference(ctx context.Context, userID, category string) (*domain.CategoryPreference, error)
	GetCategoryPreferencesByUserID(ctx context.Context, userID string) ([]domain.CategoryPreference, error)
	UpsertCategoryPreference(ctx context.Context, pref *domain.CategoryPreference) error
	DeleteCategoryPreference(ctx context.Context, userID, category string) error
	DeleteAllCategoryPreferences(ctx context.Context, userID string) error
	GetAllPreferences(ctx context.Context, userID string) (*domain.AllPreferences, error)
}

// TaskTemplateRepository defines the interface for task template data access
type TaskTemplateRepository interface {
	Create(ctx context.Context, template *domain.TaskTemplate) error
	FindByID(ctx context.Context, id string) (*domain.TaskTemplate, error)
	FindByUserID(ctx context.Context, userID string) ([]*domain.TaskTemplate, error)
	Update(ctx context.Context, template *domain.TaskTemplate) error
	Delete(ctx context.Context, id, userID string) error
	ExistsByName(ctx context.Context, userID, name string) (bool, error)
	ExistsByNameExcludingID(ctx context.Context, userID, name, excludeID string) (bool, error)
}

// DependencyRepository defines the interface for task dependency data access
type DependencyRepository interface {
	// Add creates a new dependency (taskID is blocked by blockedByID)
	Add(ctx context.Context, userID, taskID, blockedByID string) (*domain.TaskDependency, error)
	// Remove deletes a dependency relationship
	Remove(ctx context.Context, userID, taskID, blockedByID string) error
	// Exists checks if a specific dependency exists
	Exists(ctx context.Context, taskID, blockedByID string) (bool, error)
	// GetBlockers returns all tasks blocking the given task (with task details)
	GetBlockers(ctx context.Context, taskID string) ([]*domain.DependencyWithTask, error)
	// GetBlocking returns all tasks that the given task is blocking (with task details)
	GetBlocking(ctx context.Context, taskID string) ([]*domain.DependencyWithTask, error)
	// GetDependencyInfo returns complete dependency info for a task
	GetDependencyInfo(ctx context.Context, taskID string) (*domain.DependencyInfo, error)
	// GetAllBlockerIDs returns just the blocker task IDs (for cycle detection)
	GetAllBlockerIDs(ctx context.Context, taskID string) ([]string, error)
	// GetDependencyGraph returns all dependencies for cycle detection
	GetDependencyGraph(ctx context.Context, userID string) (map[string][]string, error)
	// CountIncompleteBlockers returns the number of incomplete blockers for a task
	CountIncompleteBlockers(ctx context.Context, taskID string) (int, error)
	// GetTasksBlockedBy returns tasks that will be unblocked when this task completes
	GetTasksBlockedBy(ctx context.Context, blockerTaskID string) ([]string, error)
}

// GamificationRepository defines the interface for gamification data access
type GamificationRepository interface {
	// Stats operations
	GetStats(ctx context.Context, userID string) (*domain.GamificationStats, error)
	UpsertStats(ctx context.Context, stats *domain.GamificationStats) error

	// Achievement operations
	CreateAchievement(ctx context.Context, achievement *domain.UserAchievement) error
	GetAchievements(ctx context.Context, userID string) ([]*domain.UserAchievement, error)
	GetRecentAchievements(ctx context.Context, userID string, limit int) ([]*domain.UserAchievement, error)
	HasAchievement(ctx context.Context, userID string, achievementType domain.AchievementType, category *string) (bool, error)

	// Category mastery operations
	GetCategoryMastery(ctx context.Context, userID, category string) (*domain.CategoryMastery, error)
	GetAllCategoryMastery(ctx context.Context, userID string) ([]*domain.CategoryMastery, error)
	IncrementCategoryMastery(ctx context.Context, userID, category string) (*domain.CategoryMastery, error)

	// Analytics queries for stats computation
	GetTotalCompletedTasks(ctx context.Context, userID string) (int, error)
	GetCompletionsByDate(ctx context.Context, userID string, timezone string, daysBack int) (map[string]int, error)
	GetOnTimeCompletionRate(ctx context.Context, userID string, daysBack int) (float64, error)
	GetEffortDistribution(ctx context.Context, userID string, daysBack int) (map[domain.TaskEffort]int, error)
	GetSpeedCompletions(ctx context.Context, userID string) (int, error)
	GetWeeklyCompletionDays(ctx context.Context, userID string, weekStart string, timezone string) (int, error)

	// User timezone
	GetUserTimezone(ctx context.Context, userID string) (string, error)
	SetUserTimezone(ctx context.Context, userID, timezone string) error
}
