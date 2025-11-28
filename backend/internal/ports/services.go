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
	GetAtRiskTasks(ctx context.Context, userID string) ([]*domain.Task, error)
	GetCalendar(ctx context.Context, userID string, filter *domain.CalendarFilter) (*domain.CalendarResponse, error)
	RenameCategory(ctx context.Context, userID, oldName, newName string) (int, error)
	DeleteCategory(ctx context.Context, userID, categoryName string) (int, error)
}
