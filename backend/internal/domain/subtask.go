package domain

import "errors"

// Subtask-related errors
var (
	ErrSubtaskDepthExceeded = errors.New("subtasks cannot have subtasks (single-level nesting only)")
	ErrParentNotFound       = errors.New("parent task not found")
	ErrCannotCompleteParent = errors.New("cannot complete task with incomplete subtasks")
	ErrCannotCreateSubtask  = errors.New("cannot create subtask for this task type")
)

// SubtaskInfo contains aggregated subtask statistics for a parent task
type SubtaskInfo struct {
	TotalCount      int     `json:"total_count"`
	CompletedCount  int     `json:"completed_count"`
	InProgressCount int     `json:"in_progress_count"`
	TodoCount       int     `json:"todo_count"`
	CompletionRate  float64 `json:"completion_rate"` // 0.0 - 1.0
	AllComplete     bool    `json:"all_complete"`
}

// CalculateCompletionRate calculates the completion percentage and sets AllComplete flag
func (s *SubtaskInfo) CalculateCompletionRate() {
	if s.TotalCount == 0 {
		s.CompletionRate = 0
		s.AllComplete = true // No subtasks means "all complete" for parent completion purposes
		return
	}
	s.CompletionRate = float64(s.CompletedCount) / float64(s.TotalCount)
	s.AllComplete = s.CompletedCount == s.TotalCount
}

// TaskWithSubtasks extends Task with subtask aggregation and optional subtask list
type TaskWithSubtasks struct {
	*Task
	SubtaskInfo *SubtaskInfo `json:"subtask_info,omitempty"` // Aggregated statistics
	Subtasks    []*Task      `json:"subtasks,omitempty"`     // Populated when expanded
}

// CreateSubtaskDTO is used specifically for creating subtasks
type CreateSubtaskDTO struct {
	ParentTaskID    string      `json:"parent_task_id" binding:"required,uuid"`
	Title           string      `json:"title" binding:"required,max=200"`
	Description     *string     `json:"description,omitempty" binding:"omitempty,max=2000"`
	UserPriority    *int        `json:"user_priority,omitempty" binding:"omitempty,min=1,max=10"`
	DueDate         *string     `json:"due_date,omitempty"`
	EstimatedEffort *TaskEffort `json:"estimated_effort,omitempty"`
	Context         *string     `json:"context,omitempty" binding:"omitempty,max=500"`
}

// SubtaskCompletionResponse is returned when completing a subtask
// Includes whether all subtasks are now complete (to prompt parent completion)
type SubtaskCompletionResponse struct {
	CompletedTask      *Task  `json:"completed_task"`
	AllSubtasksComplete bool  `json:"all_subtasks_complete"`
	ParentTask         *Task  `json:"parent_task,omitempty"` // Included when all subtasks complete
	Message            string `json:"message,omitempty"`
}
