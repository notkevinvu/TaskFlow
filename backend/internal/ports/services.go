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
	// Anonymous user methods
	CreateAnonymousUser(ctx context.Context) (*domain.AuthResponse, error)
	ConvertGuestToRegistered(ctx context.Context, userID string, dto *domain.ConvertGuestDTO) (*domain.AuthResponse, error)
}

// TaskService defines the interface for task business logic
type TaskService interface {
	Create(ctx context.Context, userID string, dto *domain.CreateTaskDTO) (*domain.Task, error)
	Get(ctx context.Context, userID, taskID string) (*domain.Task, error)
	List(ctx context.Context, userID string, filter *domain.TaskListFilter) ([]*domain.Task, error)
	Update(ctx context.Context, userID, taskID string, dto *domain.UpdateTaskDTO) (*domain.Task, error)
	Delete(ctx context.Context, userID, taskID string) error
	Restore(ctx context.Context, userID, taskID string) (*domain.Task, error)
	Bump(ctx context.Context, userID, taskID string) (*domain.Task, error)
	Complete(ctx context.Context, userID, taskID string) (*domain.Task, error)
	CompleteWithOptions(ctx context.Context, userID, taskID string, req *domain.TaskCompletionRequest) (*domain.TaskCompletionResponse, error)
	Uncomplete(ctx context.Context, userID, taskID string) (*domain.Task, error)
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

// SubtaskService defines the interface for subtask business logic
type SubtaskService interface {
	// Subtask CRUD operations
	Create(ctx context.Context, userID string, dto *domain.CreateSubtaskDTO) (*domain.Task, error)
	GetSubtasks(ctx context.Context, userID, parentTaskID string) ([]*domain.Task, error)
	GetSubtaskInfo(ctx context.Context, userID, parentTaskID string) (*domain.SubtaskInfo, error)
	GetTaskWithSubtasks(ctx context.Context, userID, taskID string, expandSubtasks bool) (*domain.TaskWithSubtasks, error)

	// Subtask completion with parent prompt
	CompleteSubtask(ctx context.Context, userID, subtaskID string) (*domain.SubtaskCompletionResponse, error)

	// Parent task completion validation
	CanCompleteParent(ctx context.Context, userID, parentTaskID string) (bool, error)
	ValidateParentCompletion(ctx context.Context, taskID string) error
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

// DependencyService defines the interface for task dependency business logic
type DependencyService interface {
	// Add creates a "blocked by" relationship between tasks
	AddDependency(ctx context.Context, userID string, dto *domain.AddDependencyDTO) (*domain.DependencyInfo, error)
	// Remove deletes a "blocked by" relationship
	RemoveDependency(ctx context.Context, userID string, dto *domain.RemoveDependencyDTO) error
	// GetDependencyInfo returns complete dependency information for a task
	GetDependencyInfo(ctx context.Context, userID, taskID string) (*domain.DependencyInfo, error)
	// ValidateCompletion checks if a task can be completed (no incomplete blockers)
	ValidateCompletion(ctx context.Context, taskID string) error
	// GetBlockerCompletionInfo returns info about tasks unblocked when a blocker completes
	GetBlockerCompletionInfo(ctx context.Context, blockerTaskID string) (*domain.BlockerCompletionInfo, error)
}

// TaskTemplateService defines the interface for task template business logic
type TaskTemplateService interface {
	// Create creates a new task template
	Create(ctx context.Context, userID string, dto *domain.CreateTaskTemplateDTO) (*domain.TaskTemplate, error)
	// Get retrieves a template by ID (with ownership verification)
	Get(ctx context.Context, userID, templateID string) (*domain.TaskTemplate, error)
	// List retrieves all templates for a user
	List(ctx context.Context, userID string) ([]*domain.TaskTemplate, error)
	// Update updates an existing template
	Update(ctx context.Context, userID, templateID string, dto *domain.UpdateTaskTemplateDTO) (*domain.TaskTemplate, error)
	// Delete removes a template
	Delete(ctx context.Context, userID, templateID string) error
	// CreateTaskFromTemplate converts a template to a CreateTaskDTO ready for task creation
	CreateTaskFromTemplate(ctx context.Context, userID, templateID string, overrides *domain.CreateTaskDTO) (*domain.CreateTaskDTO, error)
}

// GamificationService defines the interface for gamification business logic
type GamificationService interface {
	// Dashboard data
	GetDashboard(ctx context.Context, userID string) (*domain.GamificationDashboard, error)

	// Called when a task is completed - updates stats and checks achievements
	ProcessTaskCompletion(ctx context.Context, userID string, task *domain.Task) (*domain.TaskCompletionGamificationResult, error)

	// Async version that processes gamification in background goroutine
	// Allows API to return immediately while gamification processing continues
	ProcessTaskCompletionAsync(userID string, task *domain.Task)

	// Called when a task completion is reversed - decrements stats and revokes invalid achievements
	ProcessTaskUncompletion(ctx context.Context, userID string, task *domain.Task) error

	// Async version that processes uncompletion in background goroutine
	ProcessTaskUncompletionAsync(userID string, task *domain.Task)

	// Stats computation (can be called to refresh cache)
	ComputeStats(ctx context.Context, userID string) (*domain.GamificationStats, error)

	// Achievement checking
	CheckAndAwardAchievements(ctx context.Context, userID string, task *domain.Task, stats *domain.GamificationStats) ([]*domain.AchievementEarnedEvent, error)

	// Timezone management
	SetUserTimezone(ctx context.Context, userID, timezone string) error
	GetUserTimezone(ctx context.Context, userID string) (string, error)
}
