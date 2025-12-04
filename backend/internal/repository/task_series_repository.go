package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/sqlc"
)

// TaskSeriesRepository handles database operations for task series
type TaskSeriesRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

// NewTaskSeriesRepository creates a new task series repository
func NewTaskSeriesRepository(db *pgxpool.Pool) *TaskSeriesRepository {
	return &TaskSeriesRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

// Helper: Convert domain.RecurrencePattern to sqlc.RecurrencePattern
func domainPatternToSqlc(p domain.RecurrencePattern) sqlc.RecurrencePattern {
	return sqlc.RecurrencePattern(string(p))
}

// Helper: Convert sqlc.RecurrencePattern to domain.RecurrencePattern
func sqlcPatternToDomain(p sqlc.RecurrencePattern) domain.RecurrencePattern {
	return domain.RecurrencePattern(string(p))
}

// Helper: Convert domain.DueDateCalculation to sqlc.DueDateCalculation
func domainCalculationToSqlc(d domain.DueDateCalculation) sqlc.DueDateCalculation {
	return sqlc.DueDateCalculation(string(d))
}

// Helper: Convert sqlc.DueDateCalculation to domain.DueDateCalculation
func sqlcCalculationToDomain(d sqlc.DueDateCalculation) domain.DueDateCalculation {
	return domain.DueDateCalculation(string(d))
}

// Helper: Convert sqlc.TaskSeries to domain.TaskSeries
func sqlcTaskSeriesToDomain(s sqlc.TaskSeries) domain.TaskSeries {
	return domain.TaskSeries{
		ID:                 pgtypeUUIDToString(s.ID),
		UserID:             pgtypeUUIDToString(s.UserID),
		OriginalTaskID:     pgtypeUUIDToString(s.OriginalTaskID),
		Pattern:            sqlcPatternToDomain(s.Pattern),
		IntervalValue:      int(s.IntervalValue),
		EndDate:            pgtypeTimestamptzToTimePtr(s.EndDate),
		DueDateCalculation: sqlcCalculationToDomain(s.DueDateCalculation),
		IsActive:           s.IsActive,
		CreatedAt:          pgtypeTimestamptzToTime(s.CreatedAt),
		UpdatedAt:          pgtypeTimestamptzToTime(s.UpdatedAt),
	}
}

// Create inserts a new task series into the database
func (r *TaskSeriesRepository) Create(ctx context.Context, series *domain.TaskSeries) error {
	id, err := stringToPgtypeUUID(series.ID)
	if err != nil {
		return err
	}
	userID, err := stringToPgtypeUUID(series.UserID)
	if err != nil {
		return err
	}
	originalTaskID, err := stringToPgtypeUUID(series.OriginalTaskID)
	if err != nil {
		return err
	}

	params := sqlc.CreateTaskSeriesParams{
		ID:                 id,
		UserID:             userID,
		OriginalTaskID:     originalTaskID,
		Pattern:            domainPatternToSqlc(series.Pattern),
		IntervalValue:      int32(series.IntervalValue),
		EndDate:            timePtrToPgtypeTimestamptz(series.EndDate),
		DueDateCalculation: domainCalculationToSqlc(series.DueDateCalculation),
		IsActive:           series.IsActive,
		CreatedAt:          timeToPgtypeTimestamptz(series.CreatedAt),
		UpdatedAt:          timeToPgtypeTimestamptz(series.UpdatedAt),
	}

	return r.queries.CreateTaskSeries(ctx, params)
}

// FindByID retrieves a task series by ID
func (r *TaskSeriesRepository) FindByID(ctx context.Context, id string) (*domain.TaskSeries, error) {
	pguuid, err := stringToPgtypeUUID(id)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.GetTaskSeriesByID(ctx, pguuid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSeriesNotFound
		}
		return nil, err
	}

	series := sqlcTaskSeriesToDomain(row)
	return &series, nil
}

// FindByUserID retrieves all task series for a user
func (r *TaskSeriesRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.TaskSeries, error) {
	pguuid, err := stringToPgtypeUUID(userID)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.GetTaskSeriesByUserID(ctx, pguuid)
	if err != nil {
		return nil, err
	}

	series := make([]*domain.TaskSeries, len(rows))
	for i, row := range rows {
		s := sqlcTaskSeriesToDomain(row)
		series[i] = &s
	}

	return series, nil
}

// FindActiveByUserID retrieves all active task series for a user
func (r *TaskSeriesRepository) FindActiveByUserID(ctx context.Context, userID string) ([]*domain.TaskSeries, error) {
	pguuid, err := stringToPgtypeUUID(userID)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.GetActiveTaskSeriesByUserID(ctx, pguuid)
	if err != nil {
		return nil, err
	}

	series := make([]*domain.TaskSeries, len(rows))
	for i, row := range rows {
		s := sqlcTaskSeriesToDomain(row)
		series[i] = &s
	}

	return series, nil
}

// Update updates a task series
func (r *TaskSeriesRepository) Update(ctx context.Context, series *domain.TaskSeries) error {
	id, err := stringToPgtypeUUID(series.ID)
	if err != nil {
		return err
	}
	userID, err := stringToPgtypeUUID(series.UserID)
	if err != nil {
		return err
	}

	params := sqlc.UpdateTaskSeriesParams{
		ID:                 id,
		UserID:             userID,
		Pattern:            domainPatternToSqlc(series.Pattern),
		IntervalValue:      int32(series.IntervalValue),
		EndDate:            timePtrToPgtypeTimestamptz(series.EndDate),
		DueDateCalculation: domainCalculationToSqlc(series.DueDateCalculation),
		IsActive:           series.IsActive,
	}

	return r.queries.UpdateTaskSeries(ctx, params)
}

// Deactivate deactivates a task series (stops generating new tasks)
func (r *TaskSeriesRepository) Deactivate(ctx context.Context, id, userID string) error {
	idUUID, err := stringToPgtypeUUID(id)
	if err != nil {
		return err
	}
	userUUID, err := stringToPgtypeUUID(userID)
	if err != nil {
		return err
	}

	return r.queries.DeactivateTaskSeries(ctx, sqlc.DeactivateTaskSeriesParams{
		ID:     idUUID,
		UserID: userUUID,
	})
}

// Delete removes a task series
func (r *TaskSeriesRepository) Delete(ctx context.Context, id, userID string) error {
	idUUID, err := stringToPgtypeUUID(id)
	if err != nil {
		return err
	}
	userUUID, err := stringToPgtypeUUID(userID)
	if err != nil {
		return err
	}

	return r.queries.DeleteTaskSeries(ctx, sqlc.DeleteTaskSeriesParams{
		ID:     idUUID,
		UserID: userUUID,
	})
}

// CountTasksInSeries counts the number of tasks in a series
func (r *TaskSeriesRepository) CountTasksInSeries(ctx context.Context, seriesID string) (int, error) {
	pguuid, err := stringToPgtypeUUID(seriesID)
	if err != nil {
		return 0, err
	}

	count, err := r.queries.CountTasksInSeries(ctx, pguuid)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

// GetTasksBySeriesID retrieves all tasks belonging to a series
func (r *TaskSeriesRepository) GetTasksBySeriesID(ctx context.Context, seriesID string) ([]*domain.Task, error) {
	pguuid, err := stringToPgtypeUUID(seriesID)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.GetTasksBySeriesID(ctx, pguuid)
	if err != nil {
		return nil, err
	}

	tasks := make([]*domain.Task, len(rows))
	for i, row := range rows {
		task := sqlcTaskWithRecurrenceToDomain(row)
		tasks[i] = &task
	}

	return tasks, nil
}

// Helper: Convert sqlc.GetTasksBySeriesIDRow to domain.Task
func sqlcTaskWithRecurrenceToDomain(row sqlc.GetTasksBySeriesIDRow) domain.Task {
	var seriesID *string
	if row.SeriesID.Valid {
		s := pgtypeUUIDToString(row.SeriesID)
		seriesID = &s
	}
	var parentTaskID *string
	if row.ParentTaskID.Valid {
		p := pgtypeUUIDToString(row.ParentTaskID)
		parentTaskID = &p
	}

	return domain.Task{
		ID:              pgtypeUUIDToString(row.ID),
		UserID:          pgtypeUUIDToString(row.UserID),
		Title:           row.Title,
		Description:     row.Description,
		Status:          domain.TaskStatus(string(row.Status)),
		UserPriority:    int(row.UserPriority),
		DueDate:         pgtypeTimestamptzToTimePtr(row.DueDate),
		EstimatedEffort: sqlcEffortToDomain(row.EstimatedEffort),
		Category:        row.Category,
		Context:         row.Context,
		RelatedPeople:   row.RelatedPeople,
		PriorityScore:   int(row.PriorityScore),
		BumpCount:       int(row.BumpCount),
		CreatedAt:       pgtypeTimestamptzToTime(row.CreatedAt),
		UpdatedAt:       pgtypeTimestamptzToTime(row.UpdatedAt),
		CompletedAt:     pgtypeTimestamptzToTimePtr(row.CompletedAt),
		SeriesID:        seriesID,
		ParentTaskID:    parentTaskID,
	}
}
