package domain

import "fmt"

// Custom error types for structured error handling

// ValidationError represents a validation failure
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

// NotFoundError represents a resource not found error
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	if e.ID != "" {
		return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
	}
	return fmt.Sprintf("%s not found", e.Resource)
}

// NewNotFoundError creates a new not found error
func NewNotFoundError(resource, id string) *NotFoundError {
	return &NotFoundError{
		Resource: resource,
		ID:       id,
	}
}

// ConflictError represents a conflict error (e.g., duplicate resource)
type ConflictError struct {
	Resource string
	Message  string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("conflict: %s - %s", e.Resource, e.Message)
}

// NewConflictError creates a new conflict error
func NewConflictError(resource, message string) *ConflictError {
	return &ConflictError{
		Resource: resource,
		Message:  message,
	}
}

// UnauthorizedError represents an authentication failure
type UnauthorizedError struct {
	Message string
}

func (e *UnauthorizedError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("unauthorized: %s", e.Message)
	}
	return "unauthorized"
}

// NewUnauthorizedError creates a new unauthorized error
func NewUnauthorizedError(message string) *UnauthorizedError {
	return &UnauthorizedError{Message: message}
}

// ForbiddenError represents an authorization failure
type ForbiddenError struct {
	Resource string
	Action   string
}

func (e *ForbiddenError) Error() string {
	return fmt.Sprintf("forbidden: cannot %s %s", e.Action, e.Resource)
}

// NewForbiddenError creates a new forbidden error
func NewForbiddenError(resource, action string) *ForbiddenError {
	return &ForbiddenError{
		Resource: resource,
		Action:   action,
	}
}

// InternalError represents an internal server error
type InternalError struct {
	Message string
	Cause   error
}

func (e *InternalError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("internal error: %s (caused by: %v)", e.Message, e.Cause)
	}
	return fmt.Sprintf("internal error: %s", e.Message)
}

func (e *InternalError) Unwrap() error {
	return e.Cause
}

// NewInternalError creates a new internal error
func NewInternalError(message string, cause error) *InternalError {
	return &InternalError{
		Message: message,
		Cause:   cause,
	}
}
