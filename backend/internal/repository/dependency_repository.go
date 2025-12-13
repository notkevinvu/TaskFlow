package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/sqlc"
)

// DependencyRepository handles database operations for task dependencies using sqlc
type DependencyRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

// NewDependencyRepository creates a new dependency repository
func NewDependencyRepository(db *pgxpool.Pool) *DependencyRepository {
	return &DependencyRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

// Add creates a new dependency (taskID is blocked by blockedByID)
// Validates both tasks exist and belong to the user
func (r *DependencyRepository) Add(ctx context.Context, userID, taskID, blockedByID string) (*domain.TaskDependency, error) {
	// Convert string UUIDs to pgtype.UUID
	taskUUID, err := stringToPgtypeUUID(taskID)
	if err != nil {
		return nil, err
	}
	blockedByUUID, err := stringToPgtypeUUID(blockedByID)
	if err != nil {
		return nil, err
	}
	userUUID, err := stringToPgtypeUUID(userID)
	if err != nil {
		return nil, err
	}

	// Verify both tasks exist and belong to the user
	count, err := r.queries.VerifyTasksExistForUser(ctx, sqlc.VerifyTasksExistForUserParams{
		ID:     taskUUID,
		ID_2:   blockedByUUID,
		UserID: userUUID,
	})
	if err != nil {
		return nil, err
	}
	if count != 2 {
		return nil, domain.NewNotFoundError("task", "one or both tasks")
	}

	// Insert the dependency
	now := time.Now()
	err = r.queries.AddDependency(ctx, sqlc.AddDependencyParams{
		TaskID:      taskUUID,
		BlockedByID: blockedByUUID,
		CreatedAt:   timeToPgtypeTimestamptz(now),
	})
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
	// Convert string UUIDs to pgtype.UUID
	taskUUID, err := stringToPgtypeUUID(taskID)
	if err != nil {
		return err
	}
	blockedByUUID, err := stringToPgtypeUUID(blockedByID)
	if err != nil {
		return err
	}
	userUUID, err := stringToPgtypeUUID(userID)
	if err != nil {
		return err
	}

	rowsAffected, err := r.queries.RemoveDependency(ctx, sqlc.RemoveDependencyParams{
		TaskID:      taskUUID,
		BlockedByID: blockedByUUID,
		UserID:      userUUID,
	})
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrDependencyNotFound
	}

	return nil
}

// Exists checks if a specific dependency exists
func (r *DependencyRepository) Exists(ctx context.Context, taskID, blockedByID string) (bool, error) {
	taskUUID, err := stringToPgtypeUUID(taskID)
	if err != nil {
		return false, err
	}
	blockedByUUID, err := stringToPgtypeUUID(blockedByID)
	if err != nil {
		return false, err
	}

	return r.queries.CheckDependencyExists(ctx, sqlc.CheckDependencyExistsParams{
		TaskID:      taskUUID,
		BlockedByID: blockedByUUID,
	})
}

// GetBlockers returns all tasks blocking the given task (with task details)
func (r *DependencyRepository) GetBlockers(ctx context.Context, taskID string) ([]*domain.DependencyWithTask, error) {
	taskUUID, err := stringToPgtypeUUID(taskID)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.GetBlockerTasks(ctx, taskUUID)
	if err != nil {
		return nil, err
	}

	blockers := make([]*domain.DependencyWithTask, 0, len(rows))
	for _, row := range rows {
		blockers = append(blockers, &domain.DependencyWithTask{
			TaskID:    pgtypeUUIDToString(row.ID),
			Title:     row.Title,
			Status:    domain.TaskStatus(row.Status),
			CreatedAt: row.CreatedAt.Time,
		})
	}

	return blockers, nil
}

// GetBlocking returns all tasks that the given task is blocking (with task details)
func (r *DependencyRepository) GetBlocking(ctx context.Context, taskID string) ([]*domain.DependencyWithTask, error) {
	taskUUID, err := stringToPgtypeUUID(taskID)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.GetBlockingTasks(ctx, taskUUID)
	if err != nil {
		return nil, err
	}

	blocking := make([]*domain.DependencyWithTask, 0, len(rows))
	for _, row := range rows {
		blocking = append(blocking, &domain.DependencyWithTask{
			TaskID:    pgtypeUUIDToString(row.ID),
			Title:     row.Title,
			Status:    domain.TaskStatus(row.Status),
			CreatedAt: row.CreatedAt.Time,
		})
	}

	return blocking, nil
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
	taskUUID, err := stringToPgtypeUUID(taskID)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.GetBlockerIDs(ctx, taskUUID)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, pgtypeUUIDToString(row))
	}

	return ids, nil
}

// GetDependencyGraph returns all dependencies for cycle detection
// Returns map of task_id -> [blocked_by_ids]
func (r *DependencyRepository) GetDependencyGraph(ctx context.Context, userID string) (map[string][]string, error) {
	userUUID, err := stringToPgtypeUUID(userID)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.GetDependencyGraph(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	graph := make(map[string][]string)
	for _, row := range rows {
		taskID := pgtypeUUIDToString(row.TaskID)
		blockedByID := pgtypeUUIDToString(row.BlockedByID)
		graph[taskID] = append(graph[taskID], blockedByID)
	}

	return graph, nil
}

// CountIncompleteBlockers returns the number of incomplete blockers for a task
func (r *DependencyRepository) CountIncompleteBlockers(ctx context.Context, taskID string) (int, error) {
	taskUUID, err := stringToPgtypeUUID(taskID)
	if err != nil {
		return 0, err
	}

	count, err := r.queries.CountIncompleteBlockers(ctx, taskUUID)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

// GetTasksBlockedBy returns task IDs that are blocked by the given task
func (r *DependencyRepository) GetTasksBlockedBy(ctx context.Context, blockerTaskID string) ([]string, error) {
	blockerUUID, err := stringToPgtypeUUID(blockerTaskID)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.GetTasksBlockedByTask(ctx, blockerUUID)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, pgtypeUUIDToString(row))
	}

	return ids, nil
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

	// Convert string UUIDs to pgtype.UUID slice
	uuids := make([]pgtype.UUID, 0, len(taskIDs))
	for _, id := range taskIDs {
		uuid, err := stringToPgtypeUUID(id)
		if err != nil {
			return nil, err
		}
		uuids = append(uuids, uuid)
	}

	rows, err := r.queries.CountIncompleteBlockersBatch(ctx, uuids)
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		taskID := pgtypeUUIDToString(row.TaskID)
		result[taskID] = int(row.IncompleteCount)
	}

	return result, nil
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
