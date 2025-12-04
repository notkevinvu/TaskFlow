package domain

import (
	"errors"
	"time"
)

var (
	ErrSeriesNotFound      = errors.New("task series not found")
	ErrInvalidPattern      = errors.New("invalid recurrence pattern")
	ErrInvalidInterval     = errors.New("interval must be between 1 and 365")
	ErrInvalidCalculation  = errors.New("invalid due date calculation mode")
	ErrSeriesNotActive     = errors.New("task series is not active")
	ErrSeriesEnded         = errors.New("task series has ended")
)

// RecurrencePattern represents how often a task recurs
type RecurrencePattern string

const (
	RecurrencePatternNone    RecurrencePattern = "none"
	RecurrencePatternDaily   RecurrencePattern = "daily"
	RecurrencePatternWeekly  RecurrencePattern = "weekly"
	RecurrencePatternMonthly RecurrencePattern = "monthly"
)

// Validate validates the recurrence pattern
func (p RecurrencePattern) Validate() error {
	switch p {
	case RecurrencePatternNone, RecurrencePatternDaily, RecurrencePatternWeekly, RecurrencePatternMonthly:
		return nil
	default:
		return ErrInvalidPattern
	}
}

// IsRecurring returns true if the pattern represents a recurring task
func (p RecurrencePattern) IsRecurring() bool {
	return p != RecurrencePatternNone
}

// DueDateCalculation represents how the next due date is calculated
type DueDateCalculation string

const (
	DueDateFromOriginal   DueDateCalculation = "from_original"
	DueDateFromCompletion DueDateCalculation = "from_completion"
)

// Validate validates the due date calculation mode
func (d DueDateCalculation) Validate() error {
	switch d {
	case DueDateFromOriginal, DueDateFromCompletion:
		return nil
	default:
		return ErrInvalidCalculation
	}
}

// TaskSeries represents a series of recurring tasks
type TaskSeries struct {
	ID                 string             `json:"id"`
	UserID             string             `json:"user_id"`
	OriginalTaskID     string             `json:"original_task_id"`
	Pattern            RecurrencePattern  `json:"pattern"`
	IntervalValue      int                `json:"interval_value"` // e.g., 2 for "every 2 weeks"
	EndDate            *time.Time         `json:"end_date,omitempty"`
	DueDateCalculation DueDateCalculation `json:"due_date_calculation"`
	IsActive           bool               `json:"is_active"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

// CanGenerateNext checks if a new task can be generated for this series
func (s *TaskSeries) CanGenerateNext(referenceDate time.Time) bool {
	if !s.IsActive {
		return false
	}
	if s.EndDate != nil && referenceDate.After(*s.EndDate) {
		return false
	}
	return true
}

// CalculateNextDueDate calculates the next due date based on the pattern
func (s *TaskSeries) CalculateNextDueDate(baseDate time.Time) time.Time {
	switch s.Pattern {
	case RecurrencePatternDaily:
		return baseDate.AddDate(0, 0, s.IntervalValue)
	case RecurrencePatternWeekly:
		return baseDate.AddDate(0, 0, s.IntervalValue*7)
	case RecurrencePatternMonthly:
		return baseDate.AddDate(0, s.IntervalValue, 0)
	default:
		return baseDate
	}
}

// CreateTaskSeriesDTO is used for creating a task series
type CreateTaskSeriesDTO struct {
	Pattern            RecurrencePattern  `json:"pattern" binding:"required"`
	IntervalValue      int                `json:"interval_value" binding:"required,min=1,max=365"`
	EndDate            *time.Time         `json:"end_date,omitempty"`
	DueDateCalculation DueDateCalculation `json:"due_date_calculation" binding:"required"`
}

// Validate validates the DTO
func (dto *CreateTaskSeriesDTO) Validate() error {
	if err := dto.Pattern.Validate(); err != nil {
		return err
	}
	if dto.IntervalValue < 1 || dto.IntervalValue > 365 {
		return ErrInvalidInterval
	}
	if err := dto.DueDateCalculation.Validate(); err != nil {
		return err
	}
	return nil
}

// UpdateTaskSeriesDTO is used for updating a task series
type UpdateTaskSeriesDTO struct {
	Pattern            *RecurrencePattern  `json:"pattern,omitempty"`
	IntervalValue      *int                `json:"interval_value,omitempty" binding:"omitempty,min=1,max=365"`
	EndDate            *time.Time          `json:"end_date,omitempty"`
	DueDateCalculation *DueDateCalculation `json:"due_date_calculation,omitempty"`
	IsActive           *bool               `json:"is_active,omitempty"`
}

// RecurrenceRule is embedded in task creation to enable recurrence
type RecurrenceRule struct {
	Pattern            RecurrencePattern  `json:"pattern" binding:"required"`
	IntervalValue      int                `json:"interval_value" binding:"required,min=1,max=365"`
	EndDate            *time.Time         `json:"end_date,omitempty"`
	DueDateCalculation DueDateCalculation `json:"due_date_calculation" binding:"required"`
}

// Validate validates the recurrence rule
func (r *RecurrenceRule) Validate() error {
	if err := r.Pattern.Validate(); err != nil {
		return err
	}
	if r.IntervalValue < 1 || r.IntervalValue > 365 {
		return ErrInvalidInterval
	}
	if err := r.DueDateCalculation.Validate(); err != nil {
		return err
	}
	return nil
}

// TaskCompletionRequest is sent when completing a recurring task
// It allows the user to specify how to calculate the next due date
type TaskCompletionRequest struct {
	DueDateCalculation     *DueDateCalculation `json:"due_date_calculation,omitempty"`
	SaveAsDefault          bool                `json:"save_as_default,omitempty"`
	SaveForCategory        bool                `json:"save_for_category,omitempty"`
	SkipNextOccurrence     bool                `json:"skip_next_occurrence,omitempty"`
	StopRecurrence         bool                `json:"stop_recurrence,omitempty"`
}

// TaskCompletionResponse is returned after completing a recurring task
type TaskCompletionResponse struct {
	CompletedTask *Task       `json:"completed_task"`
	NextTask      *Task       `json:"next_task,omitempty"`      // nil if series ended or stopped
	Series        *TaskSeries `json:"series,omitempty"`
	Message       string      `json:"message"`
}

// SeriesHistoryEntry represents a task in the series history
type SeriesHistoryEntry struct {
	TaskID      string     `json:"task_id"`
	Title       string     `json:"title"`
	Status      TaskStatus `json:"status"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// SeriesHistory represents the full history of a task series
type SeriesHistory struct {
	Series  *TaskSeries          `json:"series"`
	Tasks   []SeriesHistoryEntry `json:"tasks"`
	Total   int                  `json:"total"`
}
