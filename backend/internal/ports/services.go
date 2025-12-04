package ports

import (
	"context"

	"github.com/notkevinvu/taskflow/backend/internal/domain"
)

// AuthService defines the interface for authentication business logic
type AuthService interface {
	Register(ctx context.Context, dto *domain.CreateUserDTO) (*domain.AuthResponse, error)
	Login(ctx context.Context, dto *domain.LoginDTO) (*domain.AuthResponse, error)
	GetUserByID(ctx context.Context, id string) (*domain.User, error)
}

// TaskService defines the interface for task business logic
type TaskService interface {
	Create(ctx context.Context, userID string, dto *domain.CreateTaskDTO) (*domain.Task, error)
	Get(ctx context.Context, userID, taskID string) (*domain.Task, error)
	List(ctx context.Context, userID string, filter *domain.TaskListFilter) ([]*domain.Task, error)
	Update(ctx context.Context, userID, taskID string, dto *domain.UpdateTaskDTO) (*domain.Task, error)
	Delete(ctx context.Context, userID, taskID string) error
	Bump(ctx context.Context, userID, taskID string) (*domain.Task, error)
	Complete(ctx context.Context, userID, taskID string) (*domain.Task, error)
	CompleteWithOptions(ctx context.Context, userID, taskID string, req *domain.TaskCompletionRequest) (*domain.TaskCompletionResponse, error)
	GetAtRiskTasks(ctx context.Context, userID string) ([]*domain.Task, error)
	GetCalendar(ctx context.Context, userID string, filter *domain.CalendarFilter) (*domain.CalendarResponse, error)
	RenameCategory(ctx context.Context, userID, oldName, newName string) (int, error)
	DeleteCategory(ctx context.Context, userID, categoryName string) (int, error)
	// Bulk operations
	BulkDelete(ctx context.Context, userID string, taskIDs []string) (*domain.BulkOperationResponse, error)
	BulkRestore(ctx context.Context, userID string, taskIDs []string) (*domain.BulkOperationResponse, error)
}

// InsightsService defines the interface for smart insights and suggestions
type InsightsService interface {
	GetInsights(ctx context.Context, userID string) (*domain.InsightResponse, error)
	EstimateCompletionTime(ctx context.Context, userID string, task *domain.Task) (*domain.TimeEstimate, error)
	SuggestCategory(ctx context.Context, userID string, req *domain.CategorySuggestionRequest) (*domain.CategorySuggestionResponse, error)
}

// RecurrenceService defines the interface for recurring task business logic
type RecurrenceService interface {
	// Task series management
	CreateTaskWithRecurrence(ctx context.Context, userID string, task *domain.Task, rule *domain.RecurrenceRule) (*domain.Task, *domain.TaskSeries, error)
	GenerateNextTask(ctx context.Context, completedTask *domain.Task, req *domain.TaskCompletionRequest) (*domain.Task, error)
	GetUserSeries(ctx context.Context, userID string, activeOnly bool) ([]*domain.TaskSeries, error)
	GetSeriesHistory(ctx context.Context, userID, seriesID string) (*domain.SeriesHistory, error)
	UpdateSeries(ctx context.Context, userID, seriesID string, dto *domain.UpdateTaskSeriesDTO) (*domain.TaskSeries, error)
	DeactivateSeries(ctx context.Context, userID, seriesID string) error

	// User preferences
	GetUserPreferences(ctx context.Context, userID string) (*domain.AllPreferences, error)
	GetEffectiveDueDateCalculation(ctx context.Context, userID string, category *string) domain.DueDateCalculation
	SetDefaultPreference(ctx context.Context, userID string, calc domain.DueDateCalculation) error
	SetCategoryPreference(ctx context.Context, userID, category string, calc domain.DueDateCalculation) error
	DeleteCategoryPreference(ctx context.Context, userID, category string) error
}
