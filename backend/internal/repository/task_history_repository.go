package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/sqlc"
)

// TaskHistoryRepository handles database operations for task history using sqlc
type TaskHistoryRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

// NewTaskHistoryRepository creates a new task history repository
func NewTaskHistoryRepository(db *pgxpool.Pool) *TaskHistoryRepository {
	return &TaskHistoryRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

// Helper: Convert domain.TaskHistoryEventType to sqlc.TaskHistoryEventType
func domainEventTypeToSqlc(t domain.TaskHistoryEventType) sqlc.TaskHistoryEventType {
	return sqlc.TaskHistoryEventType(string(t))
}

// Helper: Convert sqlc.TaskHistoryEventType to domain.TaskHistoryEventType
func sqlcEventTypeToDomain(t sqlc.TaskHistoryEventType) domain.TaskHistoryEventType {
	return domain.TaskHistoryEventType(string(t))
}

// Helper: Convert sqlc row to domain.TaskHistory
func sqlcHistoryRowToDomain(id, userID, taskID string, eventType sqlc.TaskHistoryEventType,
	oldValue, newValue *string, createdAt string) domain.TaskHistory {
	// Parse createdAt string to time.Time if needed
	// For now, assuming it's already handled by pgtype conversion
	return domain.TaskHistory{
		ID:        id,
		UserID:    userID,
		TaskID:    taskID,
		EventType: sqlcEventTypeToDomain(eventType),
		OldValue:  oldValue,
		NewValue:  newValue,
		// CreatedAt: will be set separately
	}
}

// Create inserts a new task history entry
func (r *TaskHistoryRepository) Create(ctx context.Context, history *domain.TaskHistory) error {
	id, err := stringToPgtypeUUID(history.ID)
	if err != nil {
		return err
	}
	userID, err := stringToPgtypeUUID(history.UserID)
	if err != nil {
		return err
	}
	taskID, err := stringToPgtypeUUID(history.TaskID)
	if err != nil {
		return err
	}

	params := sqlc.CreateTaskHistoryParams{
		ID:        id,
		UserID:    userID,
		TaskID:    taskID,
		EventType: domainEventTypeToSqlc(history.EventType),
		OldValue:  history.OldValue,
		NewValue:  history.NewValue,
		CreatedAt: timeToPgtypeTimestamptz(history.CreatedAt),
	}

	return r.queries.CreateTaskHistory(ctx, params)
}

// FindByTaskID retrieves all history entries for a task
func (r *TaskHistoryRepository) FindByTaskID(ctx context.Context, taskID string) ([]*domain.TaskHistory, error) {
	taskUUID, err := stringToPgtypeUUID(taskID)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.GetTaskHistoryByTaskID(ctx, taskUUID)
	if err != nil {
		return nil, err
	}

	history := make([]*domain.TaskHistory, len(rows))
	for i, row := range rows {
		h := domain.TaskHistory{
			ID:        pgtypeUUIDToString(row.ID),
			UserID:    pgtypeUUIDToString(row.UserID),
			TaskID:    pgtypeUUIDToString(row.TaskID),
			EventType: sqlcEventTypeToDomain(row.EventType),
			OldValue:  row.OldValue,
			NewValue:  row.NewValue,
			CreatedAt: pgtypeTimestamptzToTime(row.CreatedAt),
		}
		history[i] = &h
	}

	return history, nil
}
