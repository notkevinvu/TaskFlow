package service

import (
	"context"
	"log/slog"

	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/ports"
	"github.com/notkevinvu/taskflow/backend/internal/utils/graph"
)

// DependencyService handles task dependency business logic
type DependencyService struct {
	dependencyRepo ports.DependencyRepository
	taskRepo       ports.TaskRepository
}

// NewDependencyService creates a new dependency service
func NewDependencyService(dependencyRepo ports.DependencyRepository, taskRepo ports.TaskRepository) *DependencyService {
	return &DependencyService{
		dependencyRepo: dependencyRepo,
		taskRepo:       taskRepo,
	}
}

// AddDependency creates a "blocked by" relationship between tasks
// Validates task types, ownership, and prevents cycles
func (s *DependencyService) AddDependency(ctx context.Context, userID string, dto *domain.AddDependencyDTO) (*domain.DependencyInfo, error) {
	// Prevent self-dependency
	if dto.TaskID == dto.BlockedByID {
		return nil, domain.ErrSelfDependency
	}

	// Get both tasks to validate type and ownership
	task, err := s.taskRepo.FindByID(ctx, dto.TaskID)
	if err != nil {
		return nil, domain.NewInternalError("failed to find task", err)
	}
	if task == nil {
		return nil, domain.NewNotFoundError("task", dto.TaskID)
	}
	if task.UserID != userID {
		return nil, domain.NewForbiddenError("task", "access")
	}

	blocker, err := s.taskRepo.FindByID(ctx, dto.BlockedByID)
	if err != nil {
		return nil, domain.NewInternalError("failed to find blocker task", err)
	}
	if blocker == nil {
		return nil, domain.NewNotFoundError("blocker task", dto.BlockedByID)
	}
	if blocker.UserID != userID {
		return nil, domain.NewForbiddenError("blocker task", "access")
	}

	// Only regular tasks can have or be dependencies
	if task.TaskType != domain.TaskTypeRegular {
		return nil, domain.ErrInvalidDependencyType
	}
	if blocker.TaskType != domain.TaskTypeRegular {
		return nil, domain.ErrInvalidDependencyType
	}

	// Check for existing dependency
	exists, err := s.dependencyRepo.Exists(ctx, dto.TaskID, dto.BlockedByID)
	if err != nil {
		return nil, domain.NewInternalError("failed to check dependency", err)
	}
	if exists {
		return nil, domain.ErrDependencyAlreadyExists
	}

	// Check for cycle - get current dependency graph and check if adding this edge creates a cycle
	dependencyGraph, err := s.dependencyRepo.GetDependencyGraph(ctx, userID)
	if err != nil {
		return nil, domain.NewInternalError("failed to get dependency graph", err)
	}

	cycleDetector := graph.NewCycleDetector(dependencyGraph)
	if cycleDetector.WouldCreateCycle(dto.TaskID, dto.BlockedByID) {
		return nil, domain.ErrDependencyCycle
	}

	// Add the dependency
	_, err = s.dependencyRepo.Add(ctx, userID, dto.TaskID, dto.BlockedByID)
	if err != nil {
		if err == domain.ErrDependencyAlreadyExists {
			return nil, err
		}
		return nil, domain.NewInternalError("failed to add dependency", err)
	}

	// Return updated dependency info
	return s.dependencyRepo.GetDependencyInfo(ctx, dto.TaskID)
}

// RemoveDependency deletes a "blocked by" relationship
func (s *DependencyService) RemoveDependency(ctx context.Context, userID string, dto *domain.RemoveDependencyDTO) error {
	// Verify ownership via the task
	task, err := s.taskRepo.FindByID(ctx, dto.TaskID)
	if err != nil {
		return domain.NewInternalError("failed to find task", err)
	}
	if task == nil {
		return domain.NewNotFoundError("task", dto.TaskID)
	}
	if task.UserID != userID {
		return domain.NewForbiddenError("task", "access")
	}

	err = s.dependencyRepo.Remove(ctx, userID, dto.TaskID, dto.BlockedByID)
	if err != nil {
		if err == domain.ErrDependencyNotFound {
			return err
		}
		return domain.NewInternalError("failed to remove dependency", err)
	}

	return nil
}

// GetDependencyInfo returns complete dependency information for a task
func (s *DependencyService) GetDependencyInfo(ctx context.Context, userID, taskID string) (*domain.DependencyInfo, error) {
	// Verify ownership
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, domain.NewInternalError("failed to find task", err)
	}
	if task == nil {
		return nil, domain.NewNotFoundError("task", taskID)
	}
	if task.UserID != userID {
		return nil, domain.NewForbiddenError("task", "access")
	}

	info, err := s.dependencyRepo.GetDependencyInfo(ctx, taskID)
	if err != nil {
		return nil, domain.NewInternalError("failed to get dependency info", err)
	}

	return info, nil
}

// ValidateCompletion checks if a task can be completed (no incomplete blockers)
// Called by TaskService before completing a task
func (s *DependencyService) ValidateCompletion(ctx context.Context, taskID string) error {
	// Get the task
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return domain.NewInternalError("failed to find task", err)
	}
	if task == nil {
		return nil // Not found is OK - TaskService handles this
	}

	// Only check regular tasks (they can have dependencies)
	if task.TaskType != domain.TaskTypeRegular {
		return nil
	}

	// Check for incomplete blockers
	incompleteCount, err := s.dependencyRepo.CountIncompleteBlockers(ctx, taskID)
	if err != nil {
		return domain.NewInternalError("failed to check blocker status", err)
	}

	if incompleteCount > 0 {
		return domain.ErrCannotCompleteBlocked
	}

	return nil
}

// GetBlockerCompletionInfo returns info about tasks unblocked when a blocker completes
func (s *DependencyService) GetBlockerCompletionInfo(ctx context.Context, blockerTaskID string) (*domain.BlockerCompletionInfo, error) {
	// Get tasks that were blocked by this task
	blockedTaskIDs, err := s.dependencyRepo.GetTasksBlockedBy(ctx, blockerTaskID)
	if err != nil {
		return nil, domain.NewInternalError("failed to get blocked tasks", err)
	}

	// Use batch query to check all blocked tasks at once (eliminates N+1 pattern)
	var unblockedIDs []string
	if len(blockedTaskIDs) > 0 {
		blockerCounts, err := s.dependencyRepo.CountIncompleteBlockersBatch(ctx, blockedTaskIDs)
		if err != nil {
			// Log error but continue with empty list - graceful degradation
			slog.Warn("failed to count incomplete blockers in batch",
				"blocker_task_id", blockerTaskID,
				"blocked_task_count", len(blockedTaskIDs),
				"error", err,
			)
		} else {
			// Find tasks with no remaining incomplete blockers
			for _, taskID := range blockedTaskIDs {
				if blockerCounts[taskID] == 0 {
					unblockedIDs = append(unblockedIDs, taskID)
				}
			}
		}
	}

	if unblockedIDs == nil {
		unblockedIDs = []string{}
	}

	return &domain.BlockerCompletionInfo{
		CompletedTaskID:  blockerTaskID,
		UnblockedTaskIDs: unblockedIDs,
		UnblockedCount:   len(unblockedIDs),
	}, nil
}
