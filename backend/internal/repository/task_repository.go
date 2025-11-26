package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
)

// TaskRepository handles database operations for tasks
type TaskRepository struct {
	db *pgxpool.Pool
}

// NewTaskRepository creates a new task repository
func NewTaskRepository(db *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{db: db}
}

// Create inserts a new task into the database
func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
	query := `
		INSERT INTO tasks (
			id, user_id, title, description, status, user_priority,
			due_date, estimated_effort, category, context, related_people,
			priority_score, bump_count, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	_, err := r.db.Exec(ctx, query,
		task.ID,
		task.UserID,
		task.Title,
		task.Description,
		task.Status,
		task.UserPriority,
		task.DueDate,
		task.EstimatedEffort,
		task.Category,
		task.Context,
		task.RelatedPeople,
		task.PriorityScore,
		task.BumpCount,
		task.CreatedAt,
		task.UpdatedAt,
	)
	return err
}

// FindByID retrieves a task by ID
func (r *TaskRepository) FindByID(ctx context.Context, id string) (*domain.Task, error) {
	query := `
		SELECT id, user_id, title, description, status, user_priority,
			   due_date, estimated_effort, category, context, related_people,
			   priority_score, bump_count, created_at, updated_at, completed_at
		FROM tasks
		WHERE id = $1
	`
	var task domain.Task
	err := r.db.QueryRow(ctx, query, id).Scan(
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
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTaskNotFound
		}
		return nil, err
	}
	return &task, nil
}

// List retrieves tasks with filters
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
	query := `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, user_priority = $4,
			due_date = $5, estimated_effort = $6, category = $7, context = $8,
			related_people = $9, priority_score = $10, bump_count = $11,
			updated_at = $12, completed_at = $13
		WHERE id = $14 AND user_id = $15
	`
	result, err := r.db.Exec(ctx, query,
		task.Title,
		task.Description,
		task.Status,
		task.UserPriority,
		task.DueDate,
		task.EstimatedEffort,
		task.Category,
		task.Context,
		task.RelatedPeople,
		task.PriorityScore,
		task.BumpCount,
		task.UpdatedAt,
		task.CompletedAt,
		task.ID,
		task.UserID,
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
	query := `DELETE FROM tasks WHERE id = $1 AND user_id = $2`
	result, err := r.db.Exec(ctx, query, id, userID)
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
	query := `
		UPDATE tasks
		SET bump_count = bump_count + 1, updated_at = NOW()
		WHERE id = $1 AND user_id = $2
	`
	result, err := r.db.Exec(ctx, query, id, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrTaskNotFound
	}

	return nil
}

// FindAtRiskTasks retrieves tasks that are at risk (bumped 3+ times or overdue by 3+ days)
func (r *TaskRepository) FindAtRiskTasks(ctx context.Context, userID string) ([]*domain.Task, error) {
	query := `
		SELECT id, user_id, title, description, status, user_priority,
			   due_date, estimated_effort, category, context, related_people,
			   priority_score, bump_count, created_at, updated_at, completed_at
		FROM tasks
		WHERE user_id = $1
		  AND (
			  bump_count >= 3
			  OR (due_date IS NOT NULL AND due_date < NOW() - INTERVAL '3 days')
		  )
		  AND status != 'done'
		ORDER BY priority_score DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
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

// GetCategories retrieves all unique categories for a user
func (r *TaskRepository) GetCategories(ctx context.Context, userID string) ([]string, error) {
	query := `
		SELECT DISTINCT category
		FROM tasks
		WHERE user_id = $1 AND category IS NOT NULL
		ORDER BY category
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		if err := rows.Scan(&category); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, rows.Err()
}

// FindByDateRange retrieves tasks within a date range
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
// Returns the number of tasks updated
func (r *TaskRepository) RenameCategoryForUser(ctx context.Context, userID, oldName, newName string) (int, error) {
	query := `
		UPDATE tasks
		SET category = $1, updated_at = NOW()
		WHERE user_id = $2 AND category = $3 AND status != 'done'
	`

	result, err := r.db.Exec(ctx, query, newName, userID, oldName)
	if err != nil {
		return 0, err
	}

	return int(result.RowsAffected()), nil
}

// DeleteCategoryForUser removes a category from all tasks belonging to a user
// Sets category to NULL for affected tasks
// Returns the number of tasks updated
func (r *TaskRepository) DeleteCategoryForUser(ctx context.Context, userID, categoryName string) (int, error) {
	query := `
		UPDATE tasks
		SET category = NULL, updated_at = NOW()
		WHERE user_id = $1 AND category = $2 AND status != 'done'
	`

	result, err := r.db.Exec(ctx, query, userID, categoryName)
	if err != nil {
		return 0, err
	}

	return int(result.RowsAffected()), nil
}
