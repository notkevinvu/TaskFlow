package domain

import "time"

// TaskHistoryEventType represents the type of event in task history
type TaskHistoryEventType string

const (
	EventTaskCreated     TaskHistoryEventType = "created"
	EventTaskUpdated     TaskHistoryEventType = "updated"
	EventTaskBumped      TaskHistoryEventType = "bumped"
	EventTaskCompleted   TaskHistoryEventType = "completed"
	EventTaskUncompleted TaskHistoryEventType = "uncompleted"
	EventTaskDeleted     TaskHistoryEventType = "deleted"
	EventTaskRestored    TaskHistoryEventType = "restored"
	EventStatusChanged   TaskHistoryEventType = "status_changed"
)

// TaskHistory represents an audit log entry for task changes
type TaskHistory struct {
	ID        string               `json:"id"`
	UserID    string               `json:"user_id"`
	TaskID    string               `json:"task_id"`
	EventType TaskHistoryEventType `json:"event_type"`
	OldValue  *string              `json:"old_value,omitempty"`
	NewValue  *string              `json:"new_value,omitempty"`
	CreatedAt time.Time            `json:"created_at"`
}
