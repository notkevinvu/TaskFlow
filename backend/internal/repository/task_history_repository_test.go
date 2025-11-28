package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
)

func TestTaskHistoryRepository_Create(t *testing.T) {
	pool := setupTestDB(t)
	defer cleanupTestDB(t, pool)

	repo := NewTaskHistoryRepository(pool)
	ctx := context.Background()

	// Create test user
	userID := createTestUser(t, ctx, pool)
	defer pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)

	// Create test task
	taskRepo := NewTaskRepository(pool)
	taskID := uuid.New().String()
	task := &domain.Task{
		ID:            taskID,
		UserID:        userID,
		Title:         "Test Task",
		Status:        domain.TaskStatusTodo,
		UserPriority:  5,
		PriorityScore: 50,
		BumpCount:     0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := taskRepo.Create(ctx, task)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}
	defer pool.Exec(ctx, "DELETE FROM tasks WHERE id = $1", taskID)

	// Create a test history entry
	historyID := uuid.New().String()
	history := &domain.TaskHistory{
		ID:        historyID,
		UserID:    userID,
		TaskID:    taskID,
		EventType: domain.EventTaskCreated,
		NewValue:  stringPtr("Task created"),
		CreatedAt: time.Now(),
	}

	// Test Create
	err = repo.Create(ctx, history)
	if err != nil {
		t.Fatalf("Failed to create task history: %v", err)
	}

	// Cleanup
	_, err = pool.Exec(ctx, "DELETE FROM task_history WHERE id = $1", historyID)
	if err != nil {
		t.Logf("Warning: Failed to cleanup test history: %v", err)
	}
}

func TestTaskHistoryRepository_FindByTaskID(t *testing.T) {
	pool := setupTestDB(t)
	defer cleanupTestDB(t, pool)

	repo := NewTaskHistoryRepository(pool)
	ctx := context.Background()

	// Create test user
	userID := createTestUser(t, ctx, pool)
	defer pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)

	// Create test task
	taskRepo := NewTaskRepository(pool)
	taskID := uuid.New().String()
	task := &domain.Task{
		ID:            taskID,
		UserID:        userID,
		Title:         "Test Task",
		Status:        domain.TaskStatusTodo,
		UserPriority:  5,
		PriorityScore: 50,
		BumpCount:     0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := taskRepo.Create(ctx, task)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}
	defer pool.Exec(ctx, "DELETE FROM tasks WHERE id = $1", taskID)

	// Create multiple history entries
	historyIDs := make([]string, 2)
	events := []domain.TaskHistoryEventType{domain.EventTaskCreated, domain.EventTaskUpdated}

	for i, event := range events {
		historyID := uuid.New().String()
		historyIDs[i] = historyID

		history := &domain.TaskHistory{
			ID:        historyID,
			UserID:    userID,
			TaskID:    taskID,
			EventType: event,
			NewValue:  stringPtr(string(event)),
			CreatedAt: time.Now(),
		}

		err := repo.Create(ctx, history)
		if err != nil {
			t.Fatalf("Failed to create history entry: %v", err)
		}
	}
	defer func() {
		for _, id := range historyIDs {
			pool.Exec(ctx, "DELETE FROM task_history WHERE id = $1", id)
		}
	}()

	// Test FindByTaskID
	histories, err := repo.FindByTaskID(ctx, taskID)
	if err != nil {
		t.Fatalf("Failed to find task histories: %v", err)
	}

	if len(histories) != 2 {
		t.Errorf("Expected 2 history entries, got %d", len(histories))
	}

	// Verify they're ordered by created_at DESC
	if len(histories) == 2 {
		// The second event (updated) should come first due to DESC ordering
		if histories[0].EventType != domain.EventTaskUpdated {
			t.Errorf("Expected first entry to be 'updated', got %s", histories[0].EventType)
		}
	}
}
