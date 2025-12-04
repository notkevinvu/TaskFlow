package domain

import (
	"errors"
	"time"
)

var (
	ErrTaskNotFound     = errors.New("task not found")
	ErrUnauthorized     = errors.New("unauthorized to access this task")
	ErrInvalidTaskStatus = errors.New("invalid task status")
	ErrInvalidEffort     = errors.New("invalid effort estimate")
)

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusTodo       TaskStatus = "todo"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusDone       TaskStatus = "done"
)

// TaskEffort represents the estimated effort for a task
type TaskEffort string

const (
	TaskEffortSmall  TaskEffort = "small"   // < 1 hour
	TaskEffortMedium TaskEffort = "medium"  // 1-2 hours
	TaskEffortLarge  TaskEffort = "large"   // 2-4 hours
	TaskEffortXLarge TaskEffort = "xlarge"  // > 4 hours
)

// Task represents a task in the system
type Task struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	Title           string     `json:"title"`
	Description     *string    `json:"description,omitempty"`
	Status          TaskStatus `json:"status"`
	UserPriority    int        `json:"user_priority"`     // 1-10
	DueDate         *time.Time `json:"due_date,omitempty"`
	EstimatedEffort *TaskEffort `json:"estimated_effort,omitempty"`
	Category        *string    `json:"category,omitempty"`
	Context         *string    `json:"context,omitempty"`
	RelatedPeople   []string   `json:"related_people,omitempty"`
	PriorityScore   int        `json:"priority_score"`    // 0-100, calculated
	BumpCount       int        `json:"bump_count"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	// PriorityBreakdown shows the individual components of the priority calculation
	// Optional: populated when detailed breakdown is requested
	PriorityBreakdown *PriorityBreakdown `json:"priority_breakdown,omitempty"`
}

// CreateTaskDTO is used for creating tasks
type CreateTaskDTO struct {
	Title           string      `json:"title" binding:"required,max=200"`
	Description     *string     `json:"description,omitempty" binding:"omitempty,max=2000"`
	UserPriority    *int        `json:"user_priority,omitempty" binding:"omitempty,min=1,max=10"`
	DueDate         *time.Time  `json:"due_date,omitempty"`
	EstimatedEffort *TaskEffort `json:"estimated_effort,omitempty"`
	Category        *string     `json:"category,omitempty" binding:"omitempty,max=50"`
	Context         *string     `json:"context,omitempty" binding:"omitempty,max=500"`
	RelatedPeople   []string    `json:"related_people,omitempty"`
}

// UpdateTaskDTO is used for updating tasks
type UpdateTaskDTO struct {
	Title           *string     `json:"title,omitempty" binding:"omitempty,max=200"`
	Description     *string     `json:"description,omitempty" binding:"omitempty,max=2000"`
	Status          *TaskStatus `json:"status,omitempty"`
	UserPriority    *int        `json:"user_priority,omitempty" binding:"omitempty,min=1,max=10"`
	DueDate         *time.Time  `json:"due_date,omitempty"`
	EstimatedEffort *TaskEffort `json:"estimated_effort,omitempty"`
	Category        *string     `json:"category,omitempty" binding:"omitempty,max=50"`
	Context         *string     `json:"context,omitempty" binding:"omitempty,max=500"`
	RelatedPeople   []string    `json:"related_people,omitempty"`
}

// TaskListFilter is used for filtering tasks
type TaskListFilter struct {
	Status         *TaskStatus
	Category       *string
	Search         *string
	MinPriority    *int       // Filter by minimum priority score (0-100)
	MaxPriority    *int       // Filter by maximum priority score (0-100)
	DueDateStart   *time.Time // Filter by due date >= this date
	DueDateEnd     *time.Time // Filter by due date <= this date
	Limit          int
	Offset         int
}

// CalendarFilter is used for filtering calendar tasks
type CalendarFilter struct {
	StartDate time.Time
	EndDate   time.Time
	Status    []TaskStatus // Optional: filter by status
}

// CalendarDayData represents tasks for a specific day
type CalendarDayData struct {
	Count      int     `json:"count"`
	BadgeColor string  `json:"badge_color"` // "red", "yellow", or "blue"
	Tasks      []*Task `json:"tasks"`
}

// CalendarResponse represents the calendar API response
type CalendarResponse struct {
	Dates map[string]*CalendarDayData `json:"dates"` // key format: "2025-11-23"
}

// Validate validates the task status
func (s TaskStatus) Validate() error {
	switch s {
	case TaskStatusTodo, TaskStatusInProgress, TaskStatusDone:
		return nil
	default:
		return ErrInvalidTaskStatus
	}
}

// Validate validates the task effort
func (e TaskEffort) Validate() error {
	switch e {
	case TaskEffortSmall, TaskEffortMedium, TaskEffortLarge, TaskEffortXLarge:
		return nil
	default:
		return ErrInvalidEffort
	}
}

// GetEffortMultiplier returns the priority multiplier for the effort
func (e TaskEffort) GetEffortMultiplier() float64 {
	switch e {
	case TaskEffortSmall:
		return 1.3
	case TaskEffortMedium:
		return 1.15
	case TaskEffortLarge:
		return 1.0
	case TaskEffortXLarge:
		return 1.0
	default:
		return 1.0
	}
}

// ParseDate parses a date string in YYYY-MM-DD format
func ParseDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

// BulkOperationRequest is used for bulk task operations
type BulkOperationRequest struct {
	TaskIDs []string `json:"task_ids" binding:"required,min=1,max=100"`
}

// BulkOperationResponse is the response for bulk operations
type BulkOperationResponse struct {
	SuccessCount int      `json:"success_count"`
	FailedIDs    []string `json:"failed_ids,omitempty"`
	Message      string   `json:"message"`
}

// PriorityBreakdown shows the individual components of the priority calculation
type PriorityBreakdown struct {
	// Raw component values (0-100 scale, except effort_boost which is 1.0-1.3)
	UserPriority    float64 `json:"user_priority"`     // User's set priority scaled to 0-100
	TimeDecay       float64 `json:"time_decay"`        // Age-based urgency (0-100)
	DeadlineUrgency float64 `json:"deadline_urgency"`  // Due date proximity (0-100)
	BumpPenalty     float64 `json:"bump_penalty"`      // Penalty for delays (0-50)
	EffortBoost     float64 `json:"effort_boost"`      // Effort multiplier (1.0-1.3)

	// Weighted contributions (after applying weights)
	UserPriorityWeighted    float64 `json:"user_priority_weighted"`    // × 0.4
	TimeDecayWeighted       float64 `json:"time_decay_weighted"`       // × 0.3
	DeadlineUrgencyWeighted float64 `json:"deadline_urgency_weighted"` // × 0.2
	BumpPenaltyWeighted     float64 `json:"bump_penalty_weighted"`     // × 0.1
}
