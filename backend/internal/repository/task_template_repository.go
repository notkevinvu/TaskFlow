package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
)

// TaskTemplateRepository handles database operations for task templates
type TaskTemplateRepository struct {
	db *pgxpool.Pool
}

// NewTaskTemplateRepository creates a new task template repository
func NewTaskTemplateRepository(db *pgxpool.Pool) *TaskTemplateRepository {
	return &TaskTemplateRepository{db: db}
}

// Create inserts a new task template into the database
func (r *TaskTemplateRepository) Create(ctx context.Context, template *domain.TaskTemplate) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO task_templates (
			id, user_id, name, title, description, category,
			estimated_effort, user_priority, context, related_people,
			due_date_offset, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`,
		template.ID,
		template.UserID,
		template.Name,
		template.Title,
		template.Description,
		template.Category,
		template.EstimatedEffort,
		template.UserPriority,
		template.Context,
		template.RelatedPeople,
		template.DueDateOffset,
		template.CreatedAt,
		template.UpdatedAt,
	)

	if err != nil {
		if isPgUniqueViolation(err) {
			return domain.ErrTemplateDuplicateName
		}
		return err
	}

	return nil
}

// FindByID retrieves a template by ID
func (r *TaskTemplateRepository) FindByID(ctx context.Context, id string) (*domain.TaskTemplate, error) {
	var template domain.TaskTemplate
	var estimatedEffort *string

	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, name, title, description, category,
		       estimated_effort, user_priority, context, related_people,
		       due_date_offset, created_at, updated_at
		FROM task_templates
		WHERE id = $1
	`, id).Scan(
		&template.ID,
		&template.UserID,
		&template.Name,
		&template.Title,
		&template.Description,
		&template.Category,
		&estimatedEffort,
		&template.UserPriority,
		&template.Context,
		&template.RelatedPeople,
		&template.DueDateOffset,
		&template.CreatedAt,
		&template.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrTemplateNotFound
		}
		return nil, err
	}

	// Convert effort string to TaskEffort
	if estimatedEffort != nil {
		effort := domain.TaskEffort(*estimatedEffort)
		template.EstimatedEffort = &effort
	}

	// Ensure RelatedPeople is not nil
	if template.RelatedPeople == nil {
		template.RelatedPeople = []string{}
	}

	return &template, nil
}

// FindByUserID retrieves all templates for a user
func (r *TaskTemplateRepository) FindByUserID(ctx context.Context, userID string) ([]*domain.TaskTemplate, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, user_id, name, title, description, category,
		       estimated_effort, user_priority, context, related_people,
		       due_date_offset, created_at, updated_at
		FROM task_templates
		WHERE user_id = $1
		ORDER BY name ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templates := []*domain.TaskTemplate{}
	for rows.Next() {
		var template domain.TaskTemplate
		var estimatedEffort *string

		err := rows.Scan(
			&template.ID,
			&template.UserID,
			&template.Name,
			&template.Title,
			&template.Description,
			&template.Category,
			&estimatedEffort,
			&template.UserPriority,
			&template.Context,
			&template.RelatedPeople,
			&template.DueDateOffset,
			&template.CreatedAt,
			&template.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Convert effort string to TaskEffort
		if estimatedEffort != nil {
			effort := domain.TaskEffort(*estimatedEffort)
			template.EstimatedEffort = &effort
		}

		// Ensure RelatedPeople is not nil
		if template.RelatedPeople == nil {
			template.RelatedPeople = []string{}
		}

		templates = append(templates, &template)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return templates, nil
}

// Update updates a template
func (r *TaskTemplateRepository) Update(ctx context.Context, template *domain.TaskTemplate) error {
	template.UpdatedAt = time.Now()

	result, err := r.db.Exec(ctx, `
		UPDATE task_templates
		SET name = $2, title = $3, description = $4, category = $5,
		    estimated_effort = $6, user_priority = $7, context = $8,
		    related_people = $9, due_date_offset = $10, updated_at = $11
		WHERE id = $1 AND user_id = $12
	`,
		template.ID,
		template.Name,
		template.Title,
		template.Description,
		template.Category,
		template.EstimatedEffort,
		template.UserPriority,
		template.Context,
		template.RelatedPeople,
		template.DueDateOffset,
		template.UpdatedAt,
		template.UserID,
	)

	if err != nil {
		if isPgUniqueViolation(err) {
			return domain.ErrTemplateDuplicateName
		}
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrTemplateNotFound
	}

	return nil
}

// Delete removes a template
func (r *TaskTemplateRepository) Delete(ctx context.Context, id, userID string) error {
	result, err := r.db.Exec(ctx, `
		DELETE FROM task_templates
		WHERE id = $1 AND user_id = $2
	`, id, userID)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return domain.ErrTemplateNotFound
	}

	return nil
}

// ExistsByName checks if a template with the given name exists for the user
func (r *TaskTemplateRepository) ExistsByName(ctx context.Context, userID, name string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM task_templates
			WHERE user_id = $1 AND name = $2
		)
	`, userID, name).Scan(&exists)
	return exists, err
}

// ExistsByNameExcludingID checks if a template with the given name exists for the user,
// excluding the template with the given ID (for update validation)
func (r *TaskTemplateRepository) ExistsByNameExcludingID(ctx context.Context, userID, name, excludeID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM task_templates
			WHERE user_id = $1 AND name = $2 AND id != $3
		)
	`, userID, name, excludeID).Scan(&exists)
	return exists, err
}
