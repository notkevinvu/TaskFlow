package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
)

// DependencyRepository handles database operations for task dependencies
type DependencyRepository struct {
	db *pgxpool.Pool
}

// NewDependencyRepository creates a new dependency repository
func NewDependencyRepository(db *pgxpool.Pool) *DependencyRepository {
	return &DependencyRepository{db: db}
}

// Add creates a new dependency (taskID is blocked by blockedByID)
// Validates both tasks exist and belong to the user
func (r *DependencyRepository) Add(ctx context.Context, userID, taskID, blockedByID string) (*domain.TaskDependency, error) {
	// Verify both tasks exist and belong to the user
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*) FROM tasks
		WHERE id IN ($1, $2) AND user_id = $3
	`, taskID, blockedByID, userID).Scan(&count)
	if err != nil {
		return nil, err
	}
	if count != 2 {
		return nil, domain.NewNotFoundError("task", "one or both tasks")
	}

	// Insert the dependency
	now := time.Now()
	_, err = r.db.Exec(ctx, `
		INSERT INTO task_dependencies (task_id, blocked_by_id, created_at)
		VALUES ($1, $2, $3)
	`, taskID, blockedByID, now)
	if err != nil {
		// Check for unique constraint violation
		if isPgUniqueViolation(err) {
			return nil, domain.ErrDependencyAlreadyExists
		}
		return nil, err
	}

	return &domain.TaskDependency{
		TaskID:      taskID,
		BlockedByID: blockedByID,
		CreatedAt:   now,
	}, nil
}

// Remove deletes a dependency relationship
func (r *DependencyRepository) Remove(ctx context.Context, userID, taskID, blockedByID string) error {
	// Verify the task belongs to the user before deleting
	result, err := r.db.Exec(ctx, `
		DELETE FROM task_dependencies td
		USING tasks t
		WHERE td.task_id = $1
		  AND td.blocked_by_id = $2
		  AND t.id = td.task_id
		  AND t.user_id = $3
	`, taskID, blockedByID, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrDependencyNotFound
	}

	return nil
}

// Exists checks if a specific dependency exists
func (r *DependencyRepository) Exists(ctx context.Context, taskID, blockedByID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM task_dependencies
			WHERE task_id = $1 AND blocked_by_id = $2
		)
	`, taskID, blockedByID).Scan(&exists)
	return exists, err
}

// GetBlockers returns all tasks blocking the given task (with task details)
func (r *DependencyRepository) GetBlockers(ctx context.Context, taskID string) ([]*domain.DependencyWithTask, error) {
	rows, err := r.db.Query(ctx, `
		SELECT t.id, t.title, t.status, td.created_at
		FROM task_dependencies td
		INNER JOIN tasks t ON t.id = td.blocked_by_id
		WHERE td.task_id = $1
		ORDER BY td.created_at ASC
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blockers []*domain.DependencyWithTask
	for rows.Next() {
		var d domain.DependencyWithTask
		var status string
		if err := rows.Scan(&d.TaskID, &d.Title, &status, &d.CreatedAt); err != nil {
			return nil, err
		}
		d.Status = domain.TaskStatus(status)
		blockers = append(blockers, &d)
	}

	if blockers == nil {
		blockers = []*domain.DependencyWithTask{}
	}

	return blockers, rows.Err()
}

// GetBlocking returns all tasks that the given task is blocking (with task details)
func (r *DependencyRepository) GetBlocking(ctx context.Context, taskID string) ([]*domain.DependencyWithTask, error) {
	rows, err := r.db.Query(ctx, `
		SELECT t.id, t.title, t.status, td.created_at
		FROM task_dependencies td
		INNER JOIN tasks t ON t.id = td.task_id
		WHERE td.blocked_by_id = $1
		ORDER BY td.created_at ASC
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocking []*domain.DependencyWithTask
	for rows.Next() {
		var d domain.DependencyWithTask
		var status string
		if err := rows.Scan(&d.TaskID, &d.Title, &status, &d.CreatedAt); err != nil {
			return nil, err
		}
		d.Status = domain.TaskStatus(status)
		blocking = append(blocking, &d)
	}

	if blocking == nil {
		blocking = []*domain.DependencyWithTask{}
	}

	return blocking, rows.Err()
}

// GetDependencyInfo returns complete dependency info for a task
func (r *DependencyRepository) GetDependencyInfo(ctx context.Context, taskID string) (*domain.DependencyInfo, error) {
	blockers, err := r.GetBlockers(ctx, taskID)
	if err != nil {
		return nil, err
	}

	blocking, err := r.GetBlocking(ctx, taskID)
	if err != nil {
		return nil, err
	}

	return domain.NewDependencyInfo(taskID, blockers, blocking), nil
}

// GetAllBlockerIDs returns just the blocker task IDs (for cycle detection)
func (r *DependencyRepository) GetAllBlockerIDs(ctx context.Context, taskID string) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT blocked_by_id FROM task_dependencies
		WHERE task_id = $1
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	if ids == nil {
		ids = []string{}
	}

	return ids, rows.Err()
}

// GetDependencyGraph returns all dependencies for cycle detection
// Returns map of task_id -> [blocked_by_ids]
func (r *DependencyRepository) GetDependencyGraph(ctx context.Context, userID string) (map[string][]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT td.task_id, td.blocked_by_id
		FROM task_dependencies td
		INNER JOIN tasks t ON t.id = td.task_id
		WHERE t.user_id = $1
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	graph := make(map[string][]string)
	for rows.Next() {
		var taskID, blockedByID string
		if err := rows.Scan(&taskID, &blockedByID); err != nil {
			return nil, err
		}
		graph[taskID] = append(graph[taskID], blockedByID)
	}

	return graph, rows.Err()
}

// CountIncompleteBlockers returns the number of incomplete blockers for a task
func (r *DependencyRepository) CountIncompleteBlockers(ctx context.Context, taskID string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM task_dependencies td
		INNER JOIN tasks t ON t.id = td.blocked_by_id
		WHERE td.task_id = $1
		  AND t.status != 'done'
	`, taskID).Scan(&count)
	return count, err
}

// GetTasksBlockedBy returns task IDs that are blocked by the given task
func (r *DependencyRepository) GetTasksBlockedBy(ctx context.Context, blockerTaskID string) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT task_id FROM task_dependencies
		WHERE blocked_by_id = $1
	`, blockerTaskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	if ids == nil {
		ids = []string{}
	}

	return ids, rows.Err()
}

// CountIncompleteBlockersBatch returns the count of incomplete blockers for multiple tasks in a single query.
// Returns a map of taskID -> incomplete blocker count. Tasks with no incomplete blockers will have count 0.
func (r *DependencyRepository) CountIncompleteBlockersBatch(ctx context.Context, taskIDs []string) (map[string]int, error) {
	result := make(map[string]int)

	// Initialize all task IDs with 0 count
	for _, id := range taskIDs {
		result[id] = 0
	}

	if len(taskIDs) == 0 {
		return result, nil
	}

	// Build query with placeholders for all task IDs
	rows, err := r.db.Query(ctx, `
		SELECT td.task_id, COUNT(*)::int as incomplete_count
		FROM task_dependencies td
		INNER JOIN tasks t ON t.id = td.blocked_by_id
		WHERE td.task_id = ANY($1)
		  AND t.status != 'done'
		GROUP BY td.task_id
	`, taskIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var taskID string
		var count int
		if err := rows.Scan(&taskID, &count); err != nil {
			return nil, err
		}
		result[taskID] = count
	}

	return result, rows.Err()
}

// isPgUniqueViolation checks if an error is a PostgreSQL unique violation
func isPgUniqueViolation(err error) bool {
	if err == nil {
		return false
	}

	// Try to unwrap to pgconn.PgError
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		// PostgreSQL unique violation error code is 23505
		return pgErr.Code == "23505"
	}

	// Fallback: check error message for unique constraint violation
	errStr := err.Error()
	return strings.Contains(errStr, "duplicate key") || strings.Contains(errStr, "unique constraint")
}
