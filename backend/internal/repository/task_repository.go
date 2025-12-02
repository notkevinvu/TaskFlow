package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/sqlc"
)

// TaskRepository handles database operations for tasks using sqlc
type TaskRepository struct {
	db      *pgxpool.Pool
	queries *sqlc.Queries
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(db *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{
		db:      db,
		queries: sqlc.New(db),
	}
}

// Helper: Convert domain.TaskStatus to sqlc.TaskStatus
func domainStatusToSqlc(s domain.TaskStatus) sqlc.TaskStatus {
	return sqlc.TaskStatus(string(s))
}

// Helper: Convert sqlc.TaskStatus to domain.TaskStatus
func sqlcStatusToDomain(s sqlc.TaskStatus) domain.TaskStatus {
	return domain.TaskStatus(string(s))
}

// Helper: Convert domain.TaskEffort to sqlc.NullTaskEffort
func domainEffortToSqlc(e *domain.TaskEffort) sqlc.NullTaskEffort {
	if e == nil {
		return sqlc.NullTaskEffort{Valid: false}
	}
	return sqlc.NullTaskEffort{
		TaskEffort: sqlc.TaskEffort(string(*e)),
		Valid:      true,
	}
}

// Helper: Convert sqlc.NullTaskEffort to domain.TaskEffort
func sqlcEffortToDomain(e sqlc.NullTaskEffort) *domain.TaskEffort {
	if !e.Valid {
		return nil
	}
	effort := domain.TaskEffort(string(e.TaskEffort))
	return &effort
}

// Helper: Convert *time.Time to pgtype.Timestamptz
func timePtrToPgtypeTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

// Helper: Convert pgtype.Timestamptz to *time.Time
func pgtypeTimestamptzToTimePtr(ts pgtype.Timestamptz) *time.Time {
	if !ts.Valid {
		return nil
	}
	return &ts.Time
}

// Helper: Convert sqlc.Task to domain.Task
func sqlcTaskToDomain(t sqlc.Task) domain.Task {
	return domain.Task{
		ID:              pgtypeUUIDToString(t.ID),
		UserID:          pgtypeUUIDToString(t.UserID),
		Title:           t.Title,
		Description:     t.Description,
		Status:          sqlcStatusToDomain(t.Status),
		UserPriority:    int(t.UserPriority),
		DueDate:         pgtypeTimestamptzToTimePtr(t.DueDate),
		EstimatedEffort: sqlcEffortToDomain(t.EstimatedEffort),
		Category:        t.Category,
		Context:         t.Context,
		RelatedPeople:   t.RelatedPeople,
		PriorityScore:   int(t.PriorityScore),
		BumpCount:       int(t.BumpCount),
		CreatedAt:       pgtypeTimestamptzToTime(t.CreatedAt),
		UpdatedAt:       pgtypeTimestamptzToTime(t.UpdatedAt),
		CompletedAt:     pgtypeTimestamptzToTimePtr(t.CompletedAt),
	}
}

// Helper: Convert sqlc row types to domain.Task (works for GetTaskByIDRow, GetAtRiskTasksRow, etc.)
func sqlcRowToDomain(id, userID pgtype.UUID, title string, description *string, status sqlc.TaskStatus,
	userPriority int32, dueDate pgtype.Timestamptz, estimatedEffort sqlc.NullTaskEffort,
	category, context *string, relatedPeople []string, priorityScore, bumpCount int32,
	createdAt, updatedAt, completedAt pgtype.Timestamptz) domain.Task {
	return domain.Task{
		ID:              pgtypeUUIDToString(id),
		UserID:          pgtypeUUIDToString(userID),
		Title:           title,
		Description:     description,
		Status:          sqlcStatusToDomain(status),
		UserPriority:    int(userPriority),
		DueDate:         pgtypeTimestamptzToTimePtr(dueDate),
		EstimatedEffort: sqlcEffortToDomain(estimatedEffort),
		Category:        category,
		Context:         context,
		RelatedPeople:   relatedPeople,
		PriorityScore:   int(priorityScore),
		BumpCount:       int(bumpCount),
		CreatedAt:       pgtypeTimestamptzToTime(createdAt),
		UpdatedAt:       pgtypeTimestamptzToTime(updatedAt),
		CompletedAt:     pgtypeTimestamptzToTimePtr(completedAt),
	}
}

// Create inserts a new task into the database
func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	id, err := stringToPgtypeUUID(task.ID)
	if err != nil {
		return err
	}
	userID, err := stringToPgtypeUUID(task.UserID)
	if err != nil {
		return err
	}

	params := sqlc.CreateTaskParams{
		ID:              id,
		UserID:          userID,
		Title:           task.Title,
		Description:     task.Description,
		Status:          domainStatusToSqlc(task.Status),
		UserPriority:    int32(task.UserPriority),
		DueDate:         timePtrToPgtypeTimestamptz(task.DueDate),
		EstimatedEffort: domainEffortToSqlc(task.EstimatedEffort),
		Category:        task.Category,
		Context:         task.Context,
		RelatedPeople:   task.RelatedPeople,
		PriorityScore:   int32(task.PriorityScore),
		BumpCount:       int32(task.BumpCount),
		CreatedAt:       timeToPgtypeTimestamptz(task.CreatedAt),
		UpdatedAt:       timeToPgtypeTimestamptz(task.UpdatedAt),
	}

	return r.queries.CreateTask(ctx, params)
}

// FindByID retrieves a task by ID
func (r *TaskRepository) FindByID(ctx context.Context, id string) (*domain.Task, error) {
	pguuid, err := stringToPgtypeUUID(id)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.GetTaskByID(ctx, pguuid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTaskNotFound
		}
		return nil, err
	}

	domainTask := sqlcRowToDomain(row.ID, row.UserID, row.Title, row.Description, row.Status,
		row.UserPriority, row.DueDate, row.EstimatedEffort, row.Category, row.Context,
		row.RelatedPeople, row.PriorityScore, row.BumpCount, row.CreatedAt, row.UpdatedAt, row.CompletedAt)
	return &domainTask, nil
}

// List retrieves tasks with filters (kept as manual SQL due to dynamic query building)
func (r *TaskRepository) List(ctx context.Context, userID string, filter *domain.TaskListFilter) ([]*domain.Task, error) {
	query := `
		SELECT id, user_id, title, description, status, user_priority,
			   due_date, estimated_effort, category, context, related_people,
			   priority_score, bump_count, created_at, updated_at, completed_at
		FROM tasks
		WHERE user_id = $1
	`
	args := []interface{}{userID}
	argNum := 2

	// Apply filters
	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", argNum)
		args = append(args, *filter.Status)
		argNum++
	}

	if filter.Category != nil {
		query += fmt.Sprintf(" AND category = $%d", argNum)
		args = append(args, *filter.Category)
		argNum++
	}

	if filter.Search != nil && *filter.Search != "" {
		query += fmt.Sprintf(" AND search_vector @@ plainto_tsquery('english', $%d)", argNum)
		args = append(args, *filter.Search)
		argNum++
	}

	if filter.MinPriority != nil {
		query += fmt.Sprintf(" AND priority_score >= $%d", argNum)
		args = append(args, *filter.MinPriority)
		argNum++
	}

	if filter.MaxPriority != nil {
		query += fmt.Sprintf(" AND priority_score <= $%d", argNum)
		args = append(args, *filter.MaxPriority)
		argNum++
	}

	if filter.DueDateStart != nil {
		query += fmt.Sprintf(" AND due_date >= $%d", argNum)
		args = append(args, *filter.DueDateStart)
		argNum++
	}

	if filter.DueDateEnd != nil {
		query += fmt.Sprintf(" AND due_date <= $%d", argNum)
		args = append(args, *filter.DueDateEnd)
		argNum++
	}

	// Order by priority score descending
	query += " ORDER BY priority_score DESC, created_at DESC"

	// Apply limit and offset
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argNum)
		args = append(args, filter.Limit)
		argNum++
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argNum)
		args = append(args, filter.Offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		var task domain.Task
		err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.UserPriority,
			&task.DueDate,
			&task.EstimatedEffort,
			&task.Category,
			&task.Context,
			&task.RelatedPeople,
			&task.PriorityScore,
			&task.BumpCount,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &task)
	}

	return tasks, rows.Err()
}

// Update updates a task in the database
func (r *TaskRepository) Update(ctx context.Context, task *domain.Task) error {
	id, err := stringToPgtypeUUID(task.ID)
	if err != nil {
		return err
	}
	userID, err := stringToPgtypeUUID(task.UserID)
	if err != nil {
		return err
	}

	params := sqlc.UpdateTaskParams{
		Title:           task.Title,
		Description:     task.Description,
		Status:          domainStatusToSqlc(task.Status),
		UserPriority:    int32(task.UserPriority),
		DueDate:         timePtrToPgtypeTimestamptz(task.DueDate),
		EstimatedEffort: domainEffortToSqlc(task.EstimatedEffort),
		Category:        task.Category,
		Context:         task.Context,
		RelatedPeople:   task.RelatedPeople,
		PriorityScore:   int32(task.PriorityScore),
		BumpCount:       int32(task.BumpCount),
		UpdatedAt:       timeToPgtypeTimestamptz(task.UpdatedAt),
		CompletedAt:     timePtrToPgtypeTimestamptz(task.CompletedAt),
		ID:              id,
		UserID:          userID,
	}

	// Since sqlc :exec doesn't return rows affected, we need to use manual query to check
	query := `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, user_priority = $4,
			due_date = $5, estimated_effort = $6, category = $7, context = $8,
			related_people = $9, priority_score = $10, bump_count = $11,
			updated_at = $12, completed_at = $13
		WHERE id = $14 AND user_id = $15
	`
	result, err := r.db.Exec(ctx, query,
		params.Title,
		params.Description,
		params.Status,
		params.UserPriority,
		params.DueDate,
		params.EstimatedEffort,
		params.Category,
		params.Context,
		params.RelatedPeople,
		params.PriorityScore,
		params.BumpCount,
		params.UpdatedAt,
		params.CompletedAt,
		params.ID,
		params.UserID,
	)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrTaskNotFound
	}

	return nil
}

// Delete deletes a task from the database
func (r *TaskRepository) Delete(ctx context.Context, id, userID string) error {
	idUUID, err := stringToPgtypeUUID(id)
	if err != nil {
		return err
	}
	userUUID, err := stringToPgtypeUUID(userID)
	if err != nil {
		return err
	}

	// Use manual query to check rows affected
	result, err := r.db.Exec(ctx, "DELETE FROM tasks WHERE id = $1 AND user_id = $2", idUUID, userUUID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrTaskNotFound
	}

	return nil
}

// IncrementBumpCount increments the bump counter for a task
func (r *TaskRepository) IncrementBumpCount(ctx context.Context, id, userID string) error {
	idUUID, err := stringToPgtypeUUID(id)
	if err != nil {
		return err
	}
	userUUID, err := stringToPgtypeUUID(userID)
	if err != nil {
		return err
	}

	// Use manual query to check rows affected
	result, err := r.db.Exec(ctx,
		"UPDATE tasks SET bump_count = bump_count + 1, updated_at = NOW() WHERE id = $1 AND user_id = $2",
		idUUID, userUUID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrTaskNotFound
	}

	return nil
}

// FindAtRiskTasks retrieves tasks that are at risk
func (r *TaskRepository) FindAtRiskTasks(ctx context.Context, userID string) ([]*domain.Task, error) {
	userUUID, err := stringToPgtypeUUID(userID)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.GetAtRiskTasks(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	tasks := make([]*domain.Task, len(rows))
	for i, row := range rows {
		dt := sqlcRowToDomain(row.ID, row.UserID, row.Title, row.Description, row.Status,
			row.UserPriority, row.DueDate, row.EstimatedEffort, row.Category, row.Context,
			row.RelatedPeople, row.PriorityScore, row.BumpCount, row.CreatedAt, row.UpdatedAt, row.CompletedAt)
		tasks[i] = &dt
	}

	return tasks, nil
}

// GetCategories retrieves all unique categories for a user
func (r *TaskRepository) GetCategories(ctx context.Context, userID string) ([]string, error) {
	userUUID, err := stringToPgtypeUUID(userID)
	if err != nil {
		return nil, err
	}

	categories, err := r.queries.GetCategories(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	// Convert []*string to []string
	result := make([]string, 0, len(categories))
	for _, cat := range categories {
		if cat != nil {
			result = append(result, *cat)
		}
	}

	return result, nil
}

// FindByDateRange retrieves tasks within a date range (kept as manual SQL due to dynamic query building)
func (r *TaskRepository) FindByDateRange(ctx context.Context, userID string, filter *domain.CalendarFilter) ([]*domain.Task, error) {
	query := `
		SELECT id, user_id, title, description, status, user_priority,
			   due_date, estimated_effort, category, context, related_people,
			   priority_score, bump_count, created_at, updated_at, completed_at
		FROM tasks
		WHERE user_id = $1
		  AND due_date IS NOT NULL
		  AND due_date >= $2
		  AND due_date <= $3
	`
	args := []interface{}{userID, filter.StartDate, filter.EndDate}
	argNum := 4

	// Optional status filter
	if len(filter.Status) > 0 {
		// Build OR conditions for each status
		statusConditions := ""
		for i, status := range filter.Status {
			if i == 0 {
				statusConditions = fmt.Sprintf("status = $%d", argNum)
			} else {
				statusConditions += fmt.Sprintf(" OR status = $%d", argNum)
			}
			args = append(args, status)
			argNum++
		}
		query += fmt.Sprintf(" AND (%s)", statusConditions)
	}

	// Order by due date, then priority
	query += " ORDER BY due_date ASC, priority_score DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		var task domain.Task
		err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.Title,
			&task.Description,
			&task.Status,
			&task.UserPriority,
			&task.DueDate,
			&task.EstimatedEffort,
			&task.Category,
			&task.Context,
			&task.RelatedPeople,
			&task.PriorityScore,
			&task.BumpCount,
			&task.CreatedAt,
			&task.UpdatedAt,
			&task.CompletedAt,
		)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, &task)
	}

	return tasks, rows.Err()
}

// RenameCategoryForUser renames a category for all tasks belonging to a user
func (r *TaskRepository) RenameCategoryForUser(ctx context.Context, userID, oldName, newName string) (int, error) {
	userUUID, err := stringToPgtypeUUID(userID)
	if err != nil {
		return 0, err
	}

	// Use manual query to get rows affected
	result, err := r.db.Exec(ctx,
		"UPDATE tasks SET category = $1, updated_at = NOW() WHERE user_id = $2 AND category = $3 AND status != 'done'",
		newName, userUUID, oldName)
	if err != nil {
		return 0, err
	}

	return int(result.RowsAffected()), nil
}

// DeleteCategoryForUser removes a category from all tasks belonging to a user
func (r *TaskRepository) DeleteCategoryForUser(ctx context.Context, userID, categoryName string) (int, error) {
	userUUID, err := stringToPgtypeUUID(userID)
	if err != nil {
		return 0, err
	}

	// Use manual query to get rows affected
	result, err := r.db.Exec(ctx,
		"UPDATE tasks SET category = NULL, updated_at = NOW() WHERE user_id = $1 AND category = $2 AND status != 'done'",
		userUUID, categoryName)
	if err != nil {
		return 0, err
	}

	return int(result.RowsAffected()), nil
}

// Analytics-related repository methods

// CompletionStats represents completion statistics for a time period
type CompletionStats struct {
	TotalTasks     int     `json:"total_tasks"`
	CompletedTasks int     `json:"completed_tasks"`
	CompletionRate float64 `json:"completion_rate"`
	PendingTasks   int     `json:"pending_tasks"`
}

// GetCompletionStats retrieves completion statistics for a user within a time period
func (r *TaskRepository) GetCompletionStats(ctx context.Context, userID string, daysBack int) (*CompletionStats, error) {
	userUUID, err := stringToPgtypeUUID(userID)
	if err != nil {
		return nil, err
	}

	result, err := r.queries.GetCompletionStats(ctx, sqlc.GetCompletionStatsParams{
		UserID:   userUUID,
		Column2:  int32(daysBack),
	})
	if err != nil {
		return nil, err
	}

	stats := &CompletionStats{
		TotalTasks:     int(result.TotalTasks),
		CompletedTasks: int(result.CompletedTasks),
		PendingTasks:   int(result.PendingTasks),
	}

	if stats.TotalTasks > 0 {
		stats.CompletionRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
	}

	return stats, nil
}

// BumpAnalytics represents bump-related analytics
type BumpAnalytics struct {
	AverageBumpCount float64     `json:"average_bump_count"`
	TasksByBumpCount map[int]int `json:"tasks_by_bump_count"`
	AtRiskCount      int         `json:"at_risk_count"` // bump_count >= 3
}

// GetBumpAnalytics retrieves bump statistics for a user
func (r *TaskRepository) GetBumpAnalytics(ctx context.Context, userID string) (*BumpAnalytics, error) {
	userUUID, err := stringToPgtypeUUID(userID)
	if err != nil {
		return nil, err
	}

	avgBumpCountRaw, err := r.queries.GetAverageBumpCount(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	tasksByBump, err := r.queries.GetTasksByBumpCount(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	// Convert interface{} to float64
	avgBumpCount := 0.0
	if avgBumpCountRaw != nil {
		switch v := avgBumpCountRaw.(type) {
		case float64:
			avgBumpCount = v
		case string:
			// PostgreSQL numeric type might be returned as string
			fmt.Sscanf(v, "%f", &avgBumpCount)
		}
	}

	analytics := &BumpAnalytics{
		AverageBumpCount: avgBumpCount,
		TasksByBumpCount: make(map[int]int),
	}

	for _, row := range tasksByBump {
		bumpCount := int(row.BumpCount)
		taskCount := int(row.TaskCount)
		analytics.TasksByBumpCount[bumpCount] = taskCount

		if bumpCount >= 3 {
			analytics.AtRiskCount += taskCount
		}
	}

	return analytics, nil
}

// CategoryStats represents statistics for a single category
type CategoryStats struct {
	Category       string  `json:"category"`
	TotalCount     int     `json:"task_count"`
	CompletedCount int     `json:"completed_count"`
	CompletionRate float64 `json:"completion_rate"`
}

// GetCategoryBreakdown retrieves task statistics grouped by category
func (r *TaskRepository) GetCategoryBreakdown(ctx context.Context, userID string, daysBack int) ([]CategoryStats, error) {
	userUUID, err := stringToPgtypeUUID(userID)
	if err != nil {
		return nil, err
	}

	results, err := r.queries.GetCategoryBreakdown(ctx, sqlc.GetCategoryBreakdownParams{
		UserID:  userUUID,
		Column2: int32(daysBack),
	})
	if err != nil {
		return nil, err
	}

	stats := make([]CategoryStats, len(results))
	for i, r := range results {
		s := CategoryStats{
			Category:       r.Category,
			TotalCount:     int(r.TotalCount),
			CompletedCount: int(r.CompletedCount),
		}

		if s.TotalCount > 0 {
			s.CompletionRate = float64(s.CompletedCount) / float64(s.TotalCount) * 100
		}

		stats[i] = s
	}

	return stats, nil
}

// VelocityMetrics represents task completion velocity
type VelocityMetrics struct {
	Date           string `json:"date"`
	CompletedCount int    `json:"completed_count"`
}

// GetVelocityMetrics retrieves daily task completion counts
func (r *TaskRepository) GetVelocityMetrics(ctx context.Context, userID string, daysBack int) ([]VelocityMetrics, error) {
	userUUID, err := stringToPgtypeUUID(userID)
	if err != nil {
		return nil, err
	}

	results, err := r.queries.GetVelocityMetrics(ctx, sqlc.GetVelocityMetricsParams{
		UserID:  userUUID,
		Column2: int32(daysBack),
	})
	if err != nil {
		return nil, err
	}

	metrics := make([]VelocityMetrics, len(results))
	for i, r := range results {
		dateStr := ""
		if r.CompletionDate.Valid {
			dateStr = r.CompletionDate.Time.Format("2006-01-02")
		}

		metrics[i] = VelocityMetrics{
			Date:           dateStr,
			CompletedCount: int(r.CompletedCount),
		}
	}

	return metrics, nil
}

// PriorityDistribution represents task count by priority range
type PriorityDistribution struct {
	Range string `json:"priority_range"`
	Count int    `json:"task_count"`
}

// GetPriorityDistribution retrieves task distribution across priority ranges
func (r *TaskRepository) GetPriorityDistribution(ctx context.Context, userID string) ([]PriorityDistribution, error) {
	userUUID, err := stringToPgtypeUUID(userID)
	if err != nil {
		return nil, err
	}

	results, err := r.queries.GetPriorityDistribution(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	distribution := make([]PriorityDistribution, len(results))
	for i, r := range results {
		distribution[i] = PriorityDistribution{
			Range: r.PriorityRange,
			Count: int(r.TaskCount),
		}
	}

	return distribution, nil
}

// Insights analytics methods

// GetCategoryBumpStats retrieves average bump counts per category for avoidance pattern detection
func (r *TaskRepository) GetCategoryBumpStats(ctx context.Context, userID string) ([]domain.CategoryBumpStats, error) {
	userUUID, err := stringToPgtypeUUID(userID)
	if err != nil {
		return nil, err
	}

	results, err := r.queries.GetCategoryBumpStats(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	stats := make([]domain.CategoryBumpStats, len(results))
	for i, row := range results {
		stats[i] = domain.CategoryBumpStats{
			Category:  row.Category,
			AvgBumps:  row.AvgBumps,
			TaskCount: int(row.TaskCount),
		}
	}

	return stats, nil
}

// GetCompletionByDayOfWeek retrieves task completion counts grouped by day of week
func (r *TaskRepository) GetCompletionByDayOfWeek(ctx context.Context, userID string, daysBack int) ([]domain.DayOfWeekStats, error) {
	userUUID, err := stringToPgtypeUUID(userID)
	if err != nil {
		return nil, err
	}

	results, err := r.queries.GetCompletionByDayOfWeek(ctx, sqlc.GetCompletionByDayOfWeekParams{
		UserID:  userUUID,
		Column2: int32(daysBack),
	})
	if err != nil {
		return nil, err
	}

	stats := make([]domain.DayOfWeekStats, len(results))
	for i, row := range results {
		stats[i] = domain.DayOfWeekStats{
			DayOfWeek:      int(row.DayOfWeek),
			CompletedCount: int(row.CompletedCount),
		}
	}

	return stats, nil
}

// GetAgingQuickWins retrieves small effort tasks that are aging
func (r *TaskRepository) GetAgingQuickWins(ctx context.Context, userID string, minAgeDays int, limit int) ([]*domain.Task, error) {
	userUUID, err := stringToPgtypeUUID(userID)
	if err != nil {
		return nil, err
	}

	results, err := r.queries.GetAgingQuickWins(ctx, sqlc.GetAgingQuickWinsParams{
		UserID:  userUUID,
		Column2: int32(minAgeDays),
		Limit:   int32(limit),
	})
	if err != nil {
		return nil, err
	}

	tasks := make([]*domain.Task, len(results))
	for i, row := range results {
		dt := sqlcRowToDomain(row.ID, row.UserID, row.Title, row.Description, row.Status,
			row.UserPriority, row.DueDate, row.EstimatedEffort, row.Category, row.Context,
			row.RelatedPeople, row.PriorityScore, row.BumpCount, row.CreatedAt, row.UpdatedAt, row.CompletedAt)
		tasks[i] = &dt
	}

	return tasks, nil
}

// GetDeadlineClusters retrieves dates with multiple tasks due
func (r *TaskRepository) GetDeadlineClusters(ctx context.Context, userID string, windowDays int) ([]domain.DeadlineCluster, error) {
	userUUID, err := stringToPgtypeUUID(userID)
	if err != nil {
		return nil, err
	}

	results, err := r.queries.GetDeadlineClusters(ctx, sqlc.GetDeadlineClustersParams{
		UserID:  userUUID,
		Column2: windowDays, // interface{} accepts int
	})
	if err != nil {
		return nil, err
	}

	clusters := make([]domain.DeadlineCluster, len(results))
	for i, row := range results {
		// Convert interface{} titles to []string
		var titles []string
		if row.Titles != nil {
			if arr, ok := row.Titles.([]interface{}); ok {
				titles = make([]string, len(arr))
				for j, v := range arr {
					if s, ok := v.(string); ok {
						titles[j] = s
					}
				}
			}
		}

		// Convert pgtype.Date to time.Time
		var dueDate time.Time
		if row.DueDate.Valid {
			dueDate = row.DueDate.Time
		}

		clusters[i] = domain.DeadlineCluster{
			DueDate:   dueDate,
			TaskCount: int(row.TaskCount),
			Titles:    titles,
		}
	}

	return clusters, nil
}

// GetCompletionTimeStats retrieves completion time statistics for time estimation
// Uses manual SQL due to optional parameters that sqlc doesn't handle well
func (r *TaskRepository) GetCompletionTimeStats(ctx context.Context, userID string, category *string, effort *domain.TaskEffort) (*domain.CompletionTimeStats, error) {
	query := `
		SELECT
			COUNT(*)::int as sample_size,
			COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY
				EXTRACT(EPOCH FROM (completed_at - created_at)) / 86400
			), 0)::float as median_days,
			COALESCE(AVG(EXTRACT(EPOCH FROM (completed_at - created_at)) / 86400), 0)::float as avg_days
		FROM tasks
		WHERE user_id = $1
		  AND status = 'done'
		  AND completed_at IS NOT NULL
	`
	args := []interface{}{userID}
	argNum := 2

	if category != nil {
		query += fmt.Sprintf(" AND category = $%d", argNum)
		args = append(args, *category)
		argNum++
	}

	if effort != nil {
		query += fmt.Sprintf(" AND estimated_effort = $%d", argNum)
		args = append(args, string(*effort))
	}

	var stats domain.CompletionTimeStats
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&stats.SampleSize,
		&stats.MedianDays,
		&stats.AvgDays,
	)
	if err != nil {
		return nil, err
	}

	stats.CategoryMedian = stats.MedianDays // Use same median for now
	return &stats, nil
}

// GetCategoryDistribution retrieves distribution of pending tasks by category
func (r *TaskRepository) GetCategoryDistribution(ctx context.Context, userID string) ([]domain.CategoryDistribution, error) {
	userUUID, err := stringToPgtypeUUID(userID)
	if err != nil {
		return nil, err
	}

	results, err := r.queries.GetCategoryDistribution(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	dist := make([]domain.CategoryDistribution, len(results))
	for i, row := range results {
		dist[i] = domain.CategoryDistribution{
			Category:   row.Category,
			TaskCount:  int(row.TaskCount),
			Percentage: row.Percentage,
		}
	}

	return dist, nil
}
