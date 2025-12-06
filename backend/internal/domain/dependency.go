package domain

import (
	"errors"
	"time"
)

// Dependency-related errors
var (
	ErrDependencyCycle        = errors.New("adding this dependency would create a cycle")
	ErrSelfDependency         = errors.New("a task cannot depend on itself")
	ErrInvalidDependencyType  = errors.New("only regular tasks can have or be dependencies")
	ErrDependencyNotFound     = errors.New("dependency not found")
	ErrCannotCompleteBlocked  = errors.New("cannot complete task with unresolved blockers")
	ErrDependencyAlreadyExists = errors.New("dependency already exists")
)

// TaskDependency represents a blocked-by relationship between two tasks
type TaskDependency struct {
	TaskID      string    `json:"task_id"`
	BlockedByID string    `json:"blocked_by_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// DependencyWithTask includes the full task info for display
type DependencyWithTask struct {
	TaskID    string     `json:"task_id"`
	Title     string     `json:"title"`
	Status    TaskStatus `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
}

// DependencyInfo contains complete dependency information for a task
type DependencyInfo struct {
	TaskID      string                `json:"task_id"`
	Blockers    []*DependencyWithTask `json:"blockers"`     // Tasks blocking this task
	Blocking    []*DependencyWithTask `json:"blocking"`     // Tasks this task is blocking
	IsBlocked   bool                  `json:"is_blocked"`   // Has incomplete blockers
	CanComplete bool                  `json:"can_complete"` // All blockers are done
}

// AddDependencyDTO is used for creating a dependency
type AddDependencyDTO struct {
	TaskID      string `json:"task_id" binding:"required,uuid"`
	BlockedByID string `json:"blocked_by_id" binding:"required,uuid"`
}

// RemoveDependencyDTO is used for removing a dependency
type RemoveDependencyDTO struct {
	TaskID      string `json:"task_id" binding:"required,uuid"`
	BlockedByID string `json:"blocked_by_id" binding:"required,uuid"`
}

// DependencyEdge represents an edge in the dependency graph
// Used for cycle detection
type DependencyEdge struct {
	From string // Task that is blocked
	To   string // Task that is blocking
}

// BlockerCompletionInfo is returned when a blocker task is completed
type BlockerCompletionInfo struct {
	CompletedTaskID   string   `json:"completed_task_id"`
	UnblockedTaskIDs  []string `json:"unblocked_task_ids"`  // Tasks now unblocked
	UnblockedCount    int      `json:"unblocked_count"`
}

// CalculateDependencyInfo computes derived fields
func (d *DependencyInfo) CalculateDependencyInfo() {
	incompleteBlockers := 0
	for _, blocker := range d.Blockers {
		if blocker.Status != TaskStatusDone {
			incompleteBlockers++
		}
	}
	d.IsBlocked = incompleteBlockers > 0
	d.CanComplete = incompleteBlockers == 0
}

// NewDependencyInfo creates a DependencyInfo with calculated fields
func NewDependencyInfo(taskID string, blockers, blocking []*DependencyWithTask) *DependencyInfo {
	info := &DependencyInfo{
		TaskID:   taskID,
		Blockers: blockers,
		Blocking: blocking,
	}
	if info.Blockers == nil {
		info.Blockers = []*DependencyWithTask{}
	}
	if info.Blocking == nil {
		info.Blocking = []*DependencyWithTask{}
	}
	info.CalculateDependencyInfo()
	return info
}
