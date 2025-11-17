package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
)

// TaskHistoryRepository handles database operations for task history
type TaskHistoryRepository struct {
	db *pgxpool.Pool
}

// NewTaskHistoryRepository creates a new task history repository
func NewTaskHistoryRepository(db *pgxpool.Pool) *TaskHistoryRepository {
	return &TaskHistoryRepository{db: db}
}

// Create inserts a new task history entry
func (r *TaskHistoryRepository) Create(ctx context.Context, history *domain.TaskHistory) error {
	query := `
		INSERT INTO task_history (id, user_id, task_id, event_type, old_value, new_value, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Exec(ctx, query,
		history.ID,
		history.UserID,
		history.TaskID,
		history.EventType,
		history.OldValue,
		history.NewValue,
		history.CreatedAt,
	)
	return err
}

// FindByTaskID retrieves all history entries for a task
func (r *TaskHistoryRepository) FindByTaskID(ctx context.Context, taskID string) ([]*domain.TaskHistory, error) {
	query := `
		SELECT id, user_id, task_id, event_type, old_value, new_value, created_at
		FROM task_history
		WHERE task_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*domain.TaskHistory
	for rows.Next() {
		var h domain.TaskHistory
		err := rows.Scan(
			&h.ID,
			&h.UserID,
			&h.TaskID,
			&h.EventType,
			&h.OldValue,
			&h.NewValue,
			&h.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		history = append(history, &h)
	}

	return history, rows.Err()
}
