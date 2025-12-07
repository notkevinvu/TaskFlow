package testutil

import (
	"time"

	"github.com/google/uuid"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
)

// strPtr returns a pointer to the given string
func strPtr(s string) *string {
	return &s
}

// UserBuilder provides a fluent API for building test User objects
type UserBuilder struct {
	user *domain.User
}

// NewUserBuilder creates a new UserBuilder with sensible defaults (registered user)
func NewUserBuilder() *UserBuilder {
	return &UserBuilder{
		user: &domain.User{
			ID:           uuid.New().String(),
			UserType:     domain.UserTypeRegistered,
			Email:        strPtr("test@example.com"),
			Name:         strPtr("Test User"),
			PasswordHash: strPtr("$2a$10$hashedpassword"),
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
	}
}

func (b *UserBuilder) WithID(id string) *UserBuilder {
	b.user.ID = id
	return b
}

func (b *UserBuilder) WithEmail(email string) *UserBuilder {
	b.user.Email = strPtr(email)
	return b
}

func (b *UserBuilder) WithName(name string) *UserBuilder {
	b.user.Name = strPtr(name)
	return b
}

func (b *UserBuilder) WithUserType(userType domain.UserType) *UserBuilder {
	b.user.UserType = userType
	return b
}

func (b *UserBuilder) AsAnonymous() *UserBuilder {
	b.user.UserType = domain.UserTypeAnonymous
	b.user.Email = nil
	b.user.Name = nil
	b.user.PasswordHash = nil
	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	b.user.ExpiresAt = &expiresAt
	return b
}

func (b *UserBuilder) Build() *domain.User {
	return b.user
}

// TaskBuilder provides a fluent API for building test Task objects
type TaskBuilder struct {
	task *domain.Task
}

// NewTaskBuilder creates a new TaskBuilder with sensible defaults
func NewTaskBuilder() *TaskBuilder {
	now := time.Now()
	return &TaskBuilder{
		task: &domain.Task{
			ID:            uuid.New().String(),
			UserID:        uuid.New().String(),
			Title:         "Test Task",
			Description:   stringPtr("Test description"),
			Status:        domain.TaskStatusTodo,
			UserPriority:  5,
			PriorityScore: 50,
			BumpCount:     0,
			CreatedAt:     now,
			UpdatedAt:     now,
		},
	}
}

func (b *TaskBuilder) WithID(id string) *TaskBuilder {
	b.task.ID = id
	return b
}

func (b *TaskBuilder) WithUserID(userID string) *TaskBuilder {
	b.task.UserID = userID
	return b
}

func (b *TaskBuilder) WithTitle(title string) *TaskBuilder {
	b.task.Title = title
	return b
}

func (b *TaskBuilder) WithDescription(desc string) *TaskBuilder {
	b.task.Description = stringPtr(desc)
	return b
}

func (b *TaskBuilder) WithStatus(status domain.TaskStatus) *TaskBuilder {
	b.task.Status = status
	return b
}

func (b *TaskBuilder) WithUserPriority(priority int) *TaskBuilder {
	b.task.UserPriority = priority
	return b
}

func (b *TaskBuilder) WithPriorityScore(score int) *TaskBuilder {
	b.task.PriorityScore = score
	return b
}

func (b *TaskBuilder) WithDueDate(dueDate time.Time) *TaskBuilder {
	b.task.DueDate = &dueDate
	return b
}

func (b *TaskBuilder) WithEstimatedEffort(effort domain.TaskEffort) *TaskBuilder {
	b.task.EstimatedEffort = &effort
	return b
}

func (b *TaskBuilder) WithCategory(category string) *TaskBuilder {
	b.task.Category = stringPtr(category)
	return b
}

func (b *TaskBuilder) WithBumpCount(count int) *TaskBuilder {
	b.task.BumpCount = count
	return b
}

func (b *TaskBuilder) Completed() *TaskBuilder {
	now := time.Now()
	b.task.Status = domain.TaskStatusDone
	b.task.CompletedAt = &now
	return b
}

func (b *TaskBuilder) Build() *domain.Task {
	return b.task
}

// AuthResponseBuilder provides a fluent API for building test AuthResponse objects
type AuthResponseBuilder struct {
	response *domain.AuthResponse
}

// NewAuthResponseBuilder creates a new AuthResponseBuilder with sensible defaults
func NewAuthResponseBuilder() *AuthResponseBuilder {
	return &AuthResponseBuilder{
		response: &domain.AuthResponse{
			User:        *NewUserBuilder().Build(),
			AccessToken: "test-jwt-token-" + uuid.New().String(),
		},
	}
}

func (b *AuthResponseBuilder) WithUser(user domain.User) *AuthResponseBuilder {
	b.response.User = user
	return b
}

func (b *AuthResponseBuilder) WithAccessToken(token string) *AuthResponseBuilder {
	b.response.AccessToken = token
	return b
}

func (b *AuthResponseBuilder) Build() *domain.AuthResponse {
	return b.response
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
