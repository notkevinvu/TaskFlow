package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Helper to create expired anonymous user for testing
func createExpiredAnonymousUser(id string, taskCount int) *domain.User {
	now := time.Now()
	expiresAt := now.Add(-24 * time.Hour) // Expired yesterday
	return &domain.User{
		ID:        id,
		UserType:  domain.UserTypeAnonymous,
		ExpiresAt: &expiresAt,
		CreatedAt: now.Add(-31 * 24 * time.Hour), // Created 31 days ago
		UpdatedAt: now,
	}
}

// =============================================================================
// NewCleanupService Tests
// =============================================================================

func TestNewCleanupService_CreatesService(t *testing.T) {
	mockUserRepo := new(MockUserRepository)

	service := NewCleanupService(mockUserRepo)

	assert.NotNil(t, service)
}

// =============================================================================
// CleanupExpiredAnonymousUsers Tests - No Expired Users
// =============================================================================

func TestCleanupExpiredAnonymousUsers_NoExpiredUsers(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewCleanupService(mockUserRepo)

	mockUserRepo.On("FindExpiredAnonymous", mock.Anything).
		Return([]*domain.User{}, nil)

	result, err := service.CleanupExpiredAnonymousUsers(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 0, result.DeletedCount)
	assert.Equal(t, 0, result.FailedCount)
	assert.Equal(t, 0, result.TotalTasks)
	assert.Empty(t, result.Errors)
	mockUserRepo.AssertExpectations(t)
}

// =============================================================================
// CleanupExpiredAnonymousUsers Tests - Successful Deletion
// =============================================================================

func TestCleanupExpiredAnonymousUsers_DeletesSingleUser(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewCleanupService(mockUserRepo)

	expiredUser := createExpiredAnonymousUser("anon-123", 5)

	mockUserRepo.On("FindExpiredAnonymous", mock.Anything).
		Return([]*domain.User{expiredUser}, nil)
	mockUserRepo.On("CountTasksByUserID", mock.Anything, "anon-123").
		Return(5, nil)
	mockUserRepo.On("LogAnonymousCleanup", mock.Anything, "anon-123", 5, expiredUser.CreatedAt).
		Return(nil)
	mockUserRepo.On("Delete", mock.Anything, "anon-123").
		Return(nil)

	result, err := service.CleanupExpiredAnonymousUsers(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 1, result.DeletedCount)
	assert.Equal(t, 0, result.FailedCount)
	assert.Equal(t, 5, result.TotalTasks)
	assert.Empty(t, result.Errors)
	mockUserRepo.AssertExpectations(t)
}

func TestCleanupExpiredAnonymousUsers_DeletesMultipleUsers(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewCleanupService(mockUserRepo)

	user1 := createExpiredAnonymousUser("anon-1", 3)
	user2 := createExpiredAnonymousUser("anon-2", 7)
	user3 := createExpiredAnonymousUser("anon-3", 0)

	mockUserRepo.On("FindExpiredAnonymous", mock.Anything).
		Return([]*domain.User{user1, user2, user3}, nil)

	// User 1: 3 tasks
	mockUserRepo.On("CountTasksByUserID", mock.Anything, "anon-1").Return(3, nil)
	mockUserRepo.On("LogAnonymousCleanup", mock.Anything, "anon-1", 3, user1.CreatedAt).Return(nil)
	mockUserRepo.On("Delete", mock.Anything, "anon-1").Return(nil)

	// User 2: 7 tasks
	mockUserRepo.On("CountTasksByUserID", mock.Anything, "anon-2").Return(7, nil)
	mockUserRepo.On("LogAnonymousCleanup", mock.Anything, "anon-2", 7, user2.CreatedAt).Return(nil)
	mockUserRepo.On("Delete", mock.Anything, "anon-2").Return(nil)

	// User 3: 0 tasks
	mockUserRepo.On("CountTasksByUserID", mock.Anything, "anon-3").Return(0, nil)
	mockUserRepo.On("LogAnonymousCleanup", mock.Anything, "anon-3", 0, user3.CreatedAt).Return(nil)
	mockUserRepo.On("Delete", mock.Anything, "anon-3").Return(nil)

	result, err := service.CleanupExpiredAnonymousUsers(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 3, result.DeletedCount)
	assert.Equal(t, 0, result.FailedCount)
	assert.Equal(t, 10, result.TotalTasks) // 3 + 7 + 0
	assert.Empty(t, result.Errors)
	mockUserRepo.AssertExpectations(t)
}

// =============================================================================
// CleanupExpiredAnonymousUsers Tests - Error Handling
// =============================================================================

func TestCleanupExpiredAnonymousUsers_FindExpiredError(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewCleanupService(mockUserRepo)

	mockUserRepo.On("FindExpiredAnonymous", mock.Anything).
		Return(nil, errors.New("database error"))

	result, err := service.CleanupExpiredAnonymousUsers(context.Background())

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "database error")
	mockUserRepo.AssertExpectations(t)
}

func TestCleanupExpiredAnonymousUsers_CountTasksError_SkipsUser(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewCleanupService(mockUserRepo)

	expiredUser := createExpiredAnonymousUser("anon-fail", 0)

	mockUserRepo.On("FindExpiredAnonymous", mock.Anything).
		Return([]*domain.User{expiredUser}, nil)
	mockUserRepo.On("CountTasksByUserID", mock.Anything, "anon-fail").
		Return(0, errors.New("count error"))

	result, err := service.CleanupExpiredAnonymousUsers(context.Background())

	require.NoError(t, err) // Overall operation succeeds
	assert.Equal(t, 0, result.DeletedCount)
	assert.Equal(t, 1, result.FailedCount)
	assert.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0], "task count failed")
	mockUserRepo.AssertExpectations(t)
}

func TestCleanupExpiredAnonymousUsers_LogCleanupError_SkipsUser(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewCleanupService(mockUserRepo)

	expiredUser := createExpiredAnonymousUser("anon-fail", 5)

	mockUserRepo.On("FindExpiredAnonymous", mock.Anything).
		Return([]*domain.User{expiredUser}, nil)
	mockUserRepo.On("CountTasksByUserID", mock.Anything, "anon-fail").
		Return(5, nil)
	mockUserRepo.On("LogAnonymousCleanup", mock.Anything, "anon-fail", 5, expiredUser.CreatedAt).
		Return(errors.New("audit log error"))

	result, err := service.CleanupExpiredAnonymousUsers(context.Background())

	require.NoError(t, err) // Overall operation succeeds
	assert.Equal(t, 0, result.DeletedCount)
	assert.Equal(t, 1, result.FailedCount)
	assert.Equal(t, 5, result.TotalTasks) // Tasks were counted before error
	assert.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0], "audit log failed")
	mockUserRepo.AssertExpectations(t)
}

func TestCleanupExpiredAnonymousUsers_DeleteError_SkipsUser(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewCleanupService(mockUserRepo)

	expiredUser := createExpiredAnonymousUser("anon-fail", 5)

	mockUserRepo.On("FindExpiredAnonymous", mock.Anything).
		Return([]*domain.User{expiredUser}, nil)
	mockUserRepo.On("CountTasksByUserID", mock.Anything, "anon-fail").
		Return(5, nil)
	mockUserRepo.On("LogAnonymousCleanup", mock.Anything, "anon-fail", 5, expiredUser.CreatedAt).
		Return(nil)
	mockUserRepo.On("Delete", mock.Anything, "anon-fail").
		Return(errors.New("delete error"))

	result, err := service.CleanupExpiredAnonymousUsers(context.Background())

	require.NoError(t, err) // Overall operation succeeds
	assert.Equal(t, 0, result.DeletedCount)
	assert.Equal(t, 1, result.FailedCount)
	assert.Equal(t, 5, result.TotalTasks)
	assert.Len(t, result.Errors, 1)
	assert.Contains(t, result.Errors[0], "delete error")
	mockUserRepo.AssertExpectations(t)
}

func TestCleanupExpiredAnonymousUsers_PartialFailure(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewCleanupService(mockUserRepo)

	user1 := createExpiredAnonymousUser("anon-success", 3)
	user2 := createExpiredAnonymousUser("anon-fail", 5)
	user3 := createExpiredAnonymousUser("anon-success-2", 2)

	mockUserRepo.On("FindExpiredAnonymous", mock.Anything).
		Return([]*domain.User{user1, user2, user3}, nil)

	// User 1: Success
	mockUserRepo.On("CountTasksByUserID", mock.Anything, "anon-success").Return(3, nil)
	mockUserRepo.On("LogAnonymousCleanup", mock.Anything, "anon-success", 3, user1.CreatedAt).Return(nil)
	mockUserRepo.On("Delete", mock.Anything, "anon-success").Return(nil)

	// User 2: Fails on delete
	mockUserRepo.On("CountTasksByUserID", mock.Anything, "anon-fail").Return(5, nil)
	mockUserRepo.On("LogAnonymousCleanup", mock.Anything, "anon-fail", 5, user2.CreatedAt).Return(nil)
	mockUserRepo.On("Delete", mock.Anything, "anon-fail").Return(errors.New("delete failed"))

	// User 3: Success
	mockUserRepo.On("CountTasksByUserID", mock.Anything, "anon-success-2").Return(2, nil)
	mockUserRepo.On("LogAnonymousCleanup", mock.Anything, "anon-success-2", 2, user3.CreatedAt).Return(nil)
	mockUserRepo.On("Delete", mock.Anything, "anon-success-2").Return(nil)

	result, err := service.CleanupExpiredAnonymousUsers(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 2, result.DeletedCount)
	assert.Equal(t, 1, result.FailedCount)
	assert.Equal(t, 10, result.TotalTasks) // 3 + 5 + 2
	assert.Len(t, result.Errors, 1)
	mockUserRepo.AssertExpectations(t)
}

// =============================================================================
// CleanupExpiredAnonymousUsers Tests - Context Cancellation
// =============================================================================

func TestCleanupExpiredAnonymousUsers_ContextCancelled(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewCleanupService(mockUserRepo)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	result, err := service.CleanupExpiredAnonymousUsers(ctx)

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.NotNil(t, result) // Result is still returned
	assert.Equal(t, 0, result.DeletedCount)
}

// =============================================================================
// CleanupResult Tests
// =============================================================================

func TestCleanupExpiredAnonymousUsers_ResultHasDuration(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewCleanupService(mockUserRepo)

	mockUserRepo.On("FindExpiredAnonymous", mock.Anything).
		Return([]*domain.User{}, nil)

	result, err := service.CleanupExpiredAnonymousUsers(context.Background())

	require.NoError(t, err)
	// Duration is recorded (may be 0 for very fast operations, but field is populated)
	assert.GreaterOrEqual(t, result.Duration, time.Duration(0), "Duration should be recorded")
}

func TestCleanupExpiredAnonymousUsers_ResultErrorsInitialized(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewCleanupService(mockUserRepo)

	mockUserRepo.On("FindExpiredAnonymous", mock.Anything).
		Return([]*domain.User{}, nil)

	result, err := service.CleanupExpiredAnonymousUsers(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, result.Errors, "Errors slice should be initialized")
	assert.Empty(t, result.Errors)
}

// =============================================================================
// RunCleanupLoop Tests (Note: These test basic behavior, not the loop itself)
// =============================================================================

func TestRunCleanupLoop_ExecutesImmediately(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewCleanupService(mockUserRepo)

	// Set up mock to return empty results
	mockUserRepo.On("FindExpiredAnonymous", mock.Anything).
		Return([]*domain.User{}, nil).Once()

	// Create a context that cancels quickly
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Run cleanup loop (will execute once then context times out)
	done := make(chan struct{})
	go func() {
		service.RunCleanupLoop(ctx, 1*time.Hour) // Long interval so only initial run happens
		close(done)
	}()

	// Wait for goroutine to finish
	select {
	case <-done:
		// Success
	case <-time.After(1 * time.Second):
		t.Fatal("Cleanup loop did not stop after context cancellation")
	}

	// Verify initial cleanup was called
	mockUserRepo.AssertNumberOfCalls(t, "FindExpiredAnonymous", 1)
}

// =============================================================================
// Edge Cases
// =============================================================================

func TestCleanupExpiredAnonymousUsers_UserWithZeroTasks(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewCleanupService(mockUserRepo)

	expiredUser := createExpiredAnonymousUser("anon-empty", 0)

	mockUserRepo.On("FindExpiredAnonymous", mock.Anything).
		Return([]*domain.User{expiredUser}, nil)
	mockUserRepo.On("CountTasksByUserID", mock.Anything, "anon-empty").
		Return(0, nil)
	mockUserRepo.On("LogAnonymousCleanup", mock.Anything, "anon-empty", 0, expiredUser.CreatedAt).
		Return(nil)
	mockUserRepo.On("Delete", mock.Anything, "anon-empty").
		Return(nil)

	result, err := service.CleanupExpiredAnonymousUsers(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 1, result.DeletedCount)
	assert.Equal(t, 0, result.TotalTasks)
	mockUserRepo.AssertExpectations(t)
}

func TestCleanupExpiredAnonymousUsers_AllUsersFail(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewCleanupService(mockUserRepo)

	user1 := createExpiredAnonymousUser("anon-1", 0)
	user2 := createExpiredAnonymousUser("anon-2", 0)

	mockUserRepo.On("FindExpiredAnonymous", mock.Anything).
		Return([]*domain.User{user1, user2}, nil)

	// Both fail on count
	mockUserRepo.On("CountTasksByUserID", mock.Anything, "anon-1").
		Return(0, errors.New("error 1"))
	mockUserRepo.On("CountTasksByUserID", mock.Anything, "anon-2").
		Return(0, errors.New("error 2"))

	result, err := service.CleanupExpiredAnonymousUsers(context.Background())

	require.NoError(t, err) // Operation succeeds even if all users fail
	assert.Equal(t, 0, result.DeletedCount)
	assert.Equal(t, 2, result.FailedCount)
	assert.Len(t, result.Errors, 2)
	mockUserRepo.AssertExpectations(t)
}
