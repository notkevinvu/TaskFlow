package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
)

func TestTaskRepository_Create(t *testing.T) {
	pool := setupTestDB(t)
	defer cleanupTestDB(t, pool)

	repo := NewTaskRepository(pool)
	ctx := context.Background()

	// Create test user
	userID := createTestUser(t, ctx, pool)
	defer pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)

	// Create a test task
	taskID := uuid.New().String()
	task := &domain.Task{
		ID:            taskID,
		UserID:        userID,
		Title:         "Test Task",
		Description:   stringPtr("Test description"),
		Status:        domain.TaskStatusTodo,
		UserPriority:  5,
		PriorityScore: 50,
		BumpCount:     0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Test Create
	err := repo.Create(ctx, task)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}

	// Cleanup
	_, err = pool.Exec(ctx, "DELETE FROM tasks WHERE id = $1", taskID)
	if err != nil {
		t.Logf("Warning: Failed to cleanup test task: %v", err)
	}
}

func TestTaskRepository_FindByID(t *testing.T) {
	pool := setupTestDB(t)
	defer cleanupTestDB(t, pool)

	repo := NewTaskRepository(pool)
	ctx := context.Background()

	// Create test user
	userID := createTestUser(t, ctx, pool)
	defer pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)

	// Create a test task
	taskID := uuid.New().String()
	title := "Test Task"
	task := &domain.Task{
		ID:            taskID,
		UserID:        userID,
		Title:         title,
		Status:        domain.TaskStatusTodo,
		UserPriority:  5,
		PriorityScore: 50,
		BumpCount:     0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := repo.Create(ctx, task)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}
	defer pool.Exec(ctx, "DELETE FROM tasks WHERE id = $1", taskID)

	// Test FindByID
	found, err := repo.FindByID(ctx, taskID)
	if err != nil {
		t.Fatalf("Failed to find task by ID: %v", err)
	}

	if found == nil {
		t.Fatal("Expected to find task, got nil")
	}

	if found.ID != taskID {
		t.Errorf("Expected ID %s, got %s", taskID, found.ID)
	}

	if found.Title != title {
		t.Errorf("Expected title %s, got %s", title, found.Title)
	}
}

func TestTaskRepository_Update(t *testing.T) {
	pool := setupTestDB(t)
	defer cleanupTestDB(t, pool)

	repo := NewTaskRepository(pool)
	ctx := context.Background()

	// Create test user
	userID := createTestUser(t, ctx, pool)
	defer pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)

	// Create a test task
	taskID := uuid.New().String()
	task := &domain.Task{
		ID:            taskID,
		UserID:        userID,
		Title:         "Original Title",
		Status:        domain.TaskStatusTodo,
		UserPriority:  5,
		PriorityScore: 50,
		BumpCount:     0,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := repo.Create(ctx, task)
	if err != nil {
		t.Fatalf("Failed to create task: %v", err)
	}
	defer pool.Exec(ctx, "DELETE FROM tasks WHERE id = $1", taskID)

	// Update the task
	task.Title = "Updated Title"
	task.Status = domain.TaskStatusInProgress
	task.UpdatedAt = time.Now()

	err = repo.Update(ctx, task)
	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	// Verify update
	found, err := repo.FindByID(ctx, taskID)
	if err != nil {
		t.Fatalf("Failed to find updated task: %v", err)
	}

	if found.Title != "Updated Title" {
		t.Errorf("Expected title 'Updated Title', got %s", found.Title)
	}

	if found.Status != domain.TaskStatusInProgress {
		t.Errorf("Expected status in_progress, got %s", found.Status)
	}
}

func TestTaskRepository_GetCategories(t *testing.T) {
	pool := setupTestDB(t)
	defer cleanupTestDB(t, pool)

	repo := NewTaskRepository(pool)
	ctx := context.Background()

	// Create test user
	userID := createTestUser(t, ctx, pool)
	defer pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)

	// Create tasks with different categories
	categories := []string{"Work", "Personal", "Work"} // Duplicate to test DISTINCT
	taskIDs := make([]string, len(categories))

	for i, cat := range categories {
		taskID := uuid.New().String()
		taskIDs[i] = taskID

		task := &domain.Task{
			ID:            taskID,
			UserID:        userID,
			Title:         "Task " + cat,
			Category:      &cat,
			Status:        domain.TaskStatusTodo,
			UserPriority:  5,
			PriorityScore: 50,
			BumpCount:     0,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		err := repo.Create(ctx, task)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}
	}
	defer func() {
		for _, id := range taskIDs {
			pool.Exec(ctx, "DELETE FROM tasks WHERE id = $1", id)
		}
	}()

	// Test GetCategories
	cats, err := repo.GetCategories(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to get categories: %v", err)
	}

	// Should return 2 unique categories: "Personal" and "Work"
	if len(cats) != 2 {
		t.Errorf("Expected 2 unique categories, got %d", len(cats))
	}
}

func TestTaskRepository_GetCompletionStats(t *testing.T) {
	pool := setupTestDB(t)
	defer cleanupTestDB(t, pool)

	repo := NewTaskRepository(pool)
	ctx := context.Background()

	// Create test user
	userID := createTestUser(t, ctx, pool)
	defer pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)

	// Create some test tasks
	taskIDs := make([]string, 3)
	statuses := []domain.TaskStatus{domain.TaskStatusDone, domain.TaskStatusTodo, domain.TaskStatusInProgress}

	for i, status := range statuses {
		taskID := uuid.New().String()
		taskIDs[i] = taskID

		task := &domain.Task{
			ID:            taskID,
			UserID:        userID,
			Title:         "Task " + string(status),
			Status:        status,
			UserPriority:  5,
			PriorityScore: 50,
			BumpCount:     0,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		err := repo.Create(ctx, task)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}
	}
	defer func() {
		for _, id := range taskIDs {
			pool.Exec(ctx, "DELETE FROM tasks WHERE id = $1", id)
		}
	}()

	// Test GetCompletionStats
	stats, err := repo.GetCompletionStats(ctx, userID, 30)
	if err != nil {
		t.Fatalf("Failed to get completion stats: %v", err)
	}

	if stats.TotalTasks != 3 {
		t.Errorf("Expected 3 total tasks, got %d", stats.TotalTasks)
	}

	if stats.CompletedTasks != 1 {
		t.Errorf("Expected 1 completed task, got %d", stats.CompletedTasks)
	}

	if stats.PendingTasks != 2 {
		t.Errorf("Expected 2 pending tasks, got %d", stats.PendingTasks)
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
