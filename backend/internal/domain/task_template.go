package domain

import (
	"errors"
	"time"
)

var (
	ErrTemplateNotFound      = errors.New("template not found")
	ErrTemplateDuplicateName = errors.New("template with this name already exists")
	ErrInvalidDueDateOffset  = errors.New("due date offset must be between 0 and 365 days")
)

// TaskTemplate represents a reusable task template
type TaskTemplate struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Name   string `json:"name"` // User-facing template name

	// Task field values
	Title           string      `json:"title"`
	Description     *string     `json:"description,omitempty"`
	Category        *string     `json:"category,omitempty"`
	EstimatedEffort *TaskEffort `json:"estimated_effort,omitempty"`
	UserPriority    int         `json:"user_priority"` // 1-10
	Context         *string     `json:"context,omitempty"`
	RelatedPeople   []string    `json:"related_people,omitempty"`

	// Relative due date (days from creation)
	DueDateOffset *int `json:"due_date_offset,omitempty"` // NULL = no due date

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateTaskTemplateDTO is used for creating templates
type CreateTaskTemplateDTO struct {
	Name            string      `json:"name" binding:"required,max=100"`
	Title           string      `json:"title" binding:"required,max=200"`
	Description     *string     `json:"description,omitempty" binding:"omitempty,max=2000"`
	Category        *string     `json:"category,omitempty" binding:"omitempty,max=50"`
	EstimatedEffort *TaskEffort `json:"estimated_effort,omitempty"`
	UserPriority    *int        `json:"user_priority,omitempty" binding:"omitempty,min=1,max=10"`
	Context         *string     `json:"context,omitempty" binding:"omitempty,max=500"`
	RelatedPeople   []string    `json:"related_people,omitempty"`
	DueDateOffset   *int        `json:"due_date_offset,omitempty" binding:"omitempty,min=0,max=365"`
}

// UpdateTaskTemplateDTO is used for updating templates
type UpdateTaskTemplateDTO struct {
	Name            *string     `json:"name,omitempty" binding:"omitempty,max=100"`
	Title           *string     `json:"title,omitempty" binding:"omitempty,max=200"`
	Description     *string     `json:"description,omitempty" binding:"omitempty,max=2000"`
	Category        *string     `json:"category,omitempty" binding:"omitempty,max=50"`
	EstimatedEffort *TaskEffort `json:"estimated_effort,omitempty"`
	UserPriority    *int        `json:"user_priority,omitempty" binding:"omitempty,min=1,max=10"`
	Context         *string     `json:"context,omitempty" binding:"omitempty,max=500"`
	RelatedPeople   []string    `json:"related_people,omitempty"`
	DueDateOffset   *int        `json:"due_date_offset,omitempty" binding:"omitempty,min=0,max=365"`
}

// TaskTemplateListResponse is the response for listing templates
type TaskTemplateListResponse struct {
	Templates  []*TaskTemplate `json:"templates"`
	TotalCount int             `json:"total_count"`
}

// ToCreateTaskDTO converts a template to a CreateTaskDTO for task creation
// Calculates absolute due date from offset
func (t *TaskTemplate) ToCreateTaskDTO() *CreateTaskDTO {
	dto := &CreateTaskDTO{
		Title:           t.Title,
		Description:     t.Description,
		Category:        t.Category,
		EstimatedEffort: t.EstimatedEffort,
		Context:         t.Context,
		RelatedPeople:   t.RelatedPeople,
	}

	// Set user priority if not default
	if t.UserPriority != 5 {
		dto.UserPriority = &t.UserPriority
	}

	// Calculate due date from offset
	if t.DueDateOffset != nil {
		dueDate := time.Now().AddDate(0, 0, *t.DueDateOffset)
		dto.DueDate = &dueDate
	}

	return dto
}

// Validate validates the template
func (t *TaskTemplate) Validate() error {
	if t.Name == "" {
		return NewValidationError("name", "template name is required")
	}
	if t.Title == "" {
		return NewValidationError("title", "template title is required")
	}
	if t.UserPriority < 1 || t.UserPriority > 10 {
		return NewValidationError("user_priority", "user priority must be between 1 and 10")
	}
	if t.DueDateOffset != nil && (*t.DueDateOffset < 0 || *t.DueDateOffset > 365) {
		return ErrInvalidDueDateOffset
	}
	if t.EstimatedEffort != nil {
		if err := t.EstimatedEffort.Validate(); err != nil {
			return err
		}
	}
	return nil
}
