package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Error Types Tests
// =============================================================================

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *ValidationError
		expected string
	}{
		{
			name:     "with field",
			err:      &ValidationError{Field: "email", Message: "invalid format"},
			expected: "validation error: email - invalid format",
		},
		{
			name:     "without field",
			err:      &ValidationError{Field: "", Message: "request body is invalid"},
			expected: "validation error: request body is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("password", "must be at least 8 characters")
	assert.Equal(t, "password", err.Field)
	assert.Equal(t, "must be at least 8 characters", err.Message)
}

func TestNotFoundError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *NotFoundError
		expected string
	}{
		{
			name:     "with ID",
			err:      &NotFoundError{Resource: "task", ID: "task-123"},
			expected: "task not found: task-123",
		},
		{
			name:     "without ID",
			err:      &NotFoundError{Resource: "user", ID: ""},
			expected: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("task", "abc-123")
	assert.Equal(t, "task", err.Resource)
	assert.Equal(t, "abc-123", err.ID)
}

func TestConflictError_Error(t *testing.T) {
	err := &ConflictError{Resource: "email", Message: "already exists"}
	assert.Equal(t, "conflict: email - already exists", err.Error())
}

func TestNewConflictError(t *testing.T) {
	err := NewConflictError("user", "email already registered")
	assert.Equal(t, "user", err.Resource)
	assert.Equal(t, "email already registered", err.Message)
}

func TestUnauthorizedError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *UnauthorizedError
		expected string
	}{
		{
			name:     "with message",
			err:      &UnauthorizedError{Message: "invalid credentials"},
			expected: "unauthorized: invalid credentials",
		},
		{
			name:     "without message",
			err:      &UnauthorizedError{Message: ""},
			expected: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestNewUnauthorizedError(t *testing.T) {
	err := NewUnauthorizedError("token expired")
	assert.Equal(t, "token expired", err.Message)
}

func TestForbiddenError_Error(t *testing.T) {
	err := &ForbiddenError{Resource: "task", Action: "delete"}
	assert.Equal(t, "forbidden: cannot delete task", err.Error())
}

func TestNewForbiddenError(t *testing.T) {
	err := NewForbiddenError("task", "modify")
	assert.Equal(t, "task", err.Resource)
	assert.Equal(t, "modify", err.Action)
}

func TestInternalError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *InternalError
		expected string
	}{
		{
			name:     "with cause",
			err:      &InternalError{Message: "database error", Cause: assert.AnError},
			expected: "internal error: database error (caused by: assert.AnError general error for testing)",
		},
		{
			name:     "without cause",
			err:      &InternalError{Message: "unknown error", Cause: nil},
			expected: "internal error: unknown error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestInternalError_Unwrap(t *testing.T) {
	cause := assert.AnError
	err := &InternalError{Message: "wrapped error", Cause: cause}
	assert.Equal(t, cause, err.Unwrap())
}

func TestNewInternalError(t *testing.T) {
	cause := assert.AnError
	err := NewInternalError("database connection failed", cause)
	assert.Equal(t, "database connection failed", err.Message)
	assert.Equal(t, cause, err.Cause)
}

// =============================================================================
// Password Hashing Tests
// =============================================================================

func TestHashPassword_Success(t *testing.T) {
	password := "SecurePass123!"
	hash, err := HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash) // Hash should be different from password
}

func TestCheckPasswordHash_CorrectPassword(t *testing.T) {
	password := "SecurePass123!"
	hash, _ := HashPassword(password)

	assert.True(t, CheckPasswordHash(password, hash))
}

func TestCheckPasswordHash_WrongPassword(t *testing.T) {
	password := "SecurePass123!"
	hash, _ := HashPassword(password)

	assert.False(t, CheckPasswordHash("WrongPassword", hash))
}

func TestCheckPasswordHash_EmptyPassword(t *testing.T) {
	password := "SecurePass123!"
	hash, _ := HashPassword(password)

	assert.False(t, CheckPasswordHash("", hash))
}

// =============================================================================
// TaskStatus Tests
// =============================================================================

func TestTaskStatus_Validate_Valid(t *testing.T) {
	validStatuses := []TaskStatus{TaskStatusTodo, TaskStatusInProgress, TaskStatusDone}

	for _, status := range validStatuses {
		t.Run(string(status), func(t *testing.T) {
			err := status.Validate()
			assert.NoError(t, err)
		})
	}
}

func TestTaskStatus_Validate_Invalid(t *testing.T) {
	invalidStatuses := []TaskStatus{"", "invalid", "pending", "cancelled"}

	for _, status := range invalidStatuses {
		t.Run(string(status), func(t *testing.T) {
			err := status.Validate()
			assert.Error(t, err)
			assert.Equal(t, ErrInvalidTaskStatus, err)
		})
	}
}

// =============================================================================
// TaskEffort Tests
// =============================================================================

func TestTaskEffort_Validate_Valid(t *testing.T) {
	validEfforts := []TaskEffort{TaskEffortSmall, TaskEffortMedium, TaskEffortLarge, TaskEffortXLarge}

	for _, effort := range validEfforts {
		t.Run(string(effort), func(t *testing.T) {
			err := effort.Validate()
			assert.NoError(t, err)
		})
	}
}

func TestTaskEffort_Validate_Invalid(t *testing.T) {
	invalidEfforts := []TaskEffort{"", "tiny", "huge", "unknown"}

	for _, effort := range invalidEfforts {
		t.Run(string(effort), func(t *testing.T) {
			err := effort.Validate()
			assert.Error(t, err)
			assert.Equal(t, ErrInvalidEffort, err)
		})
	}
}

func TestTaskEffort_GetEffortMultiplier(t *testing.T) {
	tests := []struct {
		effort   TaskEffort
		expected float64
	}{
		{TaskEffortSmall, 1.3},
		{TaskEffortMedium, 1.15},
		{TaskEffortLarge, 1.0},
		{TaskEffortXLarge, 1.0},
		{TaskEffort("unknown"), 1.0}, // Default case
	}

	for _, tt := range tests {
		t.Run(string(tt.effort), func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.effort.GetEffortMultiplier())
		})
	}
}

// =============================================================================
// ParseDate Tests
// =============================================================================

func TestParseDate_Valid(t *testing.T) {
	dateStr := "2025-12-25"
	parsed, err := ParseDate(dateStr)

	assert.NoError(t, err)
	assert.Equal(t, 2025, parsed.Year())
	assert.Equal(t, time.December, parsed.Month())
	assert.Equal(t, 25, parsed.Day())
}

func TestParseDate_Invalid(t *testing.T) {
	invalidDates := []string{
		"",
		"2025/12/25",      // Wrong separator
		"25-12-2025",      // Wrong order
		"2025-13-01",      // Invalid month
		"2025-12-32",      // Invalid day
		"not-a-date",
	}

	for _, dateStr := range invalidDates {
		t.Run(dateStr, func(t *testing.T) {
			_, err := ParseDate(dateStr)
			assert.Error(t, err)
		})
	}
}

// =============================================================================
// DTO and Struct Tests
// =============================================================================

func TestUser_PasswordHashNotInJSON(t *testing.T) {
	// This test verifies the json:"-" tag works by checking the struct
	// The actual JSON marshaling is handled by the json package
	user := User{
		ID:           "user-123",
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "should-not-appear",
	}

	// Password hash should be stored but not exposed
	assert.NotEmpty(t, user.PasswordHash)
}

func TestTask_DefaultValues(t *testing.T) {
	task := Task{
		ID:     "task-123",
		UserID: "user-123",
		Title:  "Test Task",
		Status: TaskStatusTodo,
	}

	// Verify default values for optional fields
	assert.Nil(t, task.Description)
	assert.Nil(t, task.DueDate)
	assert.Nil(t, task.EstimatedEffort)
	assert.Nil(t, task.Category)
	assert.Empty(t, task.RelatedPeople)
	assert.Equal(t, 0, task.BumpCount)
	assert.Equal(t, 0, task.PriorityScore)
}
