package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/ports"
)

// TaskTemplateService handles task template business logic
type TaskTemplateService struct {
	templateRepo ports.TaskTemplateRepository
}

// NewTaskTemplateService creates a new task template service
func NewTaskTemplateService(templateRepo ports.TaskTemplateRepository) *TaskTemplateService {
	return &TaskTemplateService{
		templateRepo: templateRepo,
	}
}

// Create creates a new task template for the user
func (s *TaskTemplateService) Create(ctx context.Context, userID string, dto *domain.CreateTaskTemplateDTO) (*domain.TaskTemplate, error) {
	// Check for duplicate name
	exists, err := s.templateRepo.ExistsByName(ctx, userID, dto.Name)
	if err != nil {
		return nil, domain.NewInternalError("failed to check template name", err)
	}
	if exists {
		return nil, domain.ErrTemplateDuplicateName
	}

	// Set default user priority if not provided
	userPriority := 5
	if dto.UserPriority != nil {
		userPriority = *dto.UserPriority
	}

	// Create template entity
	now := time.Now()
	template := &domain.TaskTemplate{
		ID:              uuid.New().String(),
		UserID:          userID,
		Name:            dto.Name,
		Title:           dto.Title,
		Description:     dto.Description,
		Category:        dto.Category,
		EstimatedEffort: dto.EstimatedEffort,
		UserPriority:    userPriority,
		Context:         dto.Context,
		RelatedPeople:   dto.RelatedPeople,
		DueDateOffset:   dto.DueDateOffset,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Ensure RelatedPeople is not nil (empty slice instead)
	if template.RelatedPeople == nil {
		template.RelatedPeople = []string{}
	}

	// Validate template
	if err := template.Validate(); err != nil {
		return nil, err
	}

	// Persist to database
	if err := s.templateRepo.Create(ctx, template); err != nil {
		if err == domain.ErrTemplateDuplicateName {
			return nil, err
		}
		return nil, domain.NewInternalError("failed to create template", err)
	}

	return template, nil
}

// Get retrieves a specific template by ID, verifying ownership
func (s *TaskTemplateService) Get(ctx context.Context, userID, templateID string) (*domain.TaskTemplate, error) {
	template, err := s.templateRepo.FindByID(ctx, templateID)
	if err != nil {
		if err == domain.ErrTemplateNotFound {
			return nil, err
		}
		return nil, domain.NewInternalError("failed to find template", err)
	}

	// Verify ownership
	if template.UserID != userID {
		return nil, domain.NewForbiddenError("template", "access")
	}

	return template, nil
}

// List retrieves all templates for a user
func (s *TaskTemplateService) List(ctx context.Context, userID string) ([]*domain.TaskTemplate, error) {
	templates, err := s.templateRepo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, domain.NewInternalError("failed to list templates", err)
	}

	return templates, nil
}

// Update updates an existing template
func (s *TaskTemplateService) Update(ctx context.Context, userID, templateID string, dto *domain.UpdateTaskTemplateDTO) (*domain.TaskTemplate, error) {
	// Get existing template
	template, err := s.Get(ctx, userID, templateID)
	if err != nil {
		return nil, err
	}

	// Check for duplicate name if name is being changed
	if dto.Name != nil && *dto.Name != template.Name {
		exists, err := s.templateRepo.ExistsByNameExcludingID(ctx, userID, *dto.Name, templateID)
		if err != nil {
			return nil, domain.NewInternalError("failed to check template name", err)
		}
		if exists {
			return nil, domain.ErrTemplateDuplicateName
		}
		template.Name = *dto.Name
	}

	// Apply updates (only update fields that are provided)
	if dto.Title != nil {
		template.Title = *dto.Title
	}
	if dto.Description != nil {
		template.Description = dto.Description
	}
	if dto.Category != nil {
		template.Category = dto.Category
	}
	if dto.EstimatedEffort != nil {
		template.EstimatedEffort = dto.EstimatedEffort
	}
	if dto.UserPriority != nil {
		template.UserPriority = *dto.UserPriority
	}
	if dto.Context != nil {
		template.Context = dto.Context
	}
	if dto.RelatedPeople != nil {
		template.RelatedPeople = dto.RelatedPeople
	}
	if dto.DueDateOffset != nil {
		template.DueDateOffset = dto.DueDateOffset
	}

	// Ensure RelatedPeople is not nil
	if template.RelatedPeople == nil {
		template.RelatedPeople = []string{}
	}

	// Validate updated template
	if err := template.Validate(); err != nil {
		return nil, err
	}

	// Persist changes
	if err := s.templateRepo.Update(ctx, template); err != nil {
		if err == domain.ErrTemplateDuplicateName {
			return nil, err
		}
		if err == domain.ErrTemplateNotFound {
			return nil, err
		}
		return nil, domain.NewInternalError("failed to update template", err)
	}

	return template, nil
}

// Delete removes a template
func (s *TaskTemplateService) Delete(ctx context.Context, userID, templateID string) error {
	// Verify ownership by attempting to get the template
	_, err := s.Get(ctx, userID, templateID)
	if err != nil {
		return err
	}

	if err := s.templateRepo.Delete(ctx, templateID, userID); err != nil {
		if err == domain.ErrTemplateNotFound {
			return err
		}
		return domain.NewInternalError("failed to delete template", err)
	}

	return nil
}

// CreateTaskFromTemplate converts a template to a CreateTaskDTO ready for task creation
// Optionally allows overriding fields from the template
func (s *TaskTemplateService) CreateTaskFromTemplate(ctx context.Context, userID, templateID string, overrides *domain.CreateTaskDTO) (*domain.CreateTaskDTO, error) {
	// Get the template
	template, err := s.Get(ctx, userID, templateID)
	if err != nil {
		return nil, err
	}

	// Convert template to CreateTaskDTO
	dto := template.ToCreateTaskDTO()

	// Apply any overrides
	if overrides != nil {
		if overrides.Title != "" {
			dto.Title = overrides.Title
		}
		if overrides.Description != nil {
			dto.Description = overrides.Description
		}
		if overrides.Category != nil {
			dto.Category = overrides.Category
		}
		if overrides.EstimatedEffort != nil {
			dto.EstimatedEffort = overrides.EstimatedEffort
		}
		if overrides.UserPriority != nil {
			dto.UserPriority = overrides.UserPriority
		}
		if overrides.Context != nil {
			dto.Context = overrides.Context
		}
		if overrides.RelatedPeople != nil {
			dto.RelatedPeople = overrides.RelatedPeople
		}
		if overrides.DueDate != nil {
			dto.DueDate = overrides.DueDate
		}
	}

	return dto, nil
}
