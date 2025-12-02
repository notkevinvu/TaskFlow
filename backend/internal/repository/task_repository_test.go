package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// TaskRepository Integration Tests
// =============================================================================

// createTestTask creates a task for testing and returns it
func createTestTask(t *testing.T, ctx context.Context, repo *TaskRepository, userID string, title string) *domain.Task {
	t.Helper()
	task := &domain.Task{
		ID:            uuid.New().String(),
		UserID:        userID,
		Title:         title,
		Status:        domain.TaskStatusTodo,
		UserPriority:  5,
		PriorityScore: 50,
		BumpCount:     0,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	err := repo.Create(ctx, task)
	require.NoError(t, err)
	return task
}

func TestTaskRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	repo := NewTaskRepository(pool)
	ctx := context.Background()

	// Create test user
	userID := createTestUser(t, ctx, pool)

	t.Run("creates task successfully with minimal fields", func(t *testing.T) {
		task := &domain.Task{
			ID:            uuid.New().String(),
			UserID:        userID,
			Title:         "Minimal Task",
			Status:        domain.TaskStatusTodo,
			UserPriority:  5,
			PriorityScore: 50,
			BumpCount:     0,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}

		err := repo.Create(ctx, task)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, task.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, task.Title, found.Title)
	})

	t.Run("creates task with all fields", func(t *testing.T) {
		description := "Full description"
		category := "Work"
		contextText := "Office context"
		effort := domain.TaskEffortMedium
		dueDate := time.Now().Add(24 * time.Hour).UTC()

		task := &domain.Task{
			ID:              uuid.New().String(),
			UserID:          userID,
			Title:           "Full Task",
			Description:     &description,
			Status:          domain.TaskStatusInProgress,
			UserPriority:    8,
			DueDate:         &dueDate,
			EstimatedEffort: &effort,
			Category:        &category,
			Context:         &contextText,
			RelatedPeople:   []string{"Alice", "Bob"},
			PriorityScore:   75,
			BumpCount:       2,
			CreatedAt:       time.Now().UTC(),
			UpdatedAt:       time.Now().UTC(),
		}

		err := repo.Create(ctx, task)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, task.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "Full Task", found.Title)
		assert.Equal(t, &description, found.Description)
		assert.Equal(t, domain.TaskStatusInProgress, found.Status)
		assert.Equal(t, 8, found.UserPriority)
		assert.NotNil(t, found.DueDate)
		assert.NotNil(t, found.EstimatedEffort)
		assert.Equal(t, domain.TaskEffortMedium, *found.EstimatedEffort)
		assert.Equal(t, &category, found.Category)
		assert.Equal(t, 2, len(found.RelatedPeople))
	})

	t.Run("fails with invalid user UUID", func(t *testing.T) {
		task := &domain.Task{
			ID:            uuid.New().String(),
			UserID:        "invalid-user-id",
			Title:         "Invalid Task",
			Status:        domain.TaskStatusTodo,
			UserPriority:  5,
			PriorityScore: 50,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}

		err := repo.Create(ctx, task)
		assert.Error(t, err)
	})

	t.Run("fails with non-existent user (foreign key)", func(t *testing.T) {
		task := &domain.Task{
			ID:            uuid.New().String(),
			UserID:        uuid.New().String(), // Random UUID that doesn't exist
			Title:         "Orphan Task",
			Status:        domain.TaskStatusTodo,
			UserPriority:  5,
			PriorityScore: 50,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}

		err := repo.Create(ctx, task)
		assert.Error(t, err) // Foreign key violation
	})
}

func TestTaskRepository_FindByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	repo := NewTaskRepository(pool)
	ctx := context.Background()
	userID := createTestUser(t, ctx, pool)

	t.Run("finds existing task", func(t *testing.T) {
		task := createTestTask(t, ctx, repo, userID, "Find Me Task")

		found, err := repo.FindByID(ctx, task.ID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, task.ID, found.ID)
		assert.Equal(t, task.Title, found.Title)
	})

	t.Run("returns not found error for non-existent task", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		found, err := repo.FindByID(ctx, nonExistentID)
		assert.ErrorIs(t, err, domain.ErrTaskNotFound)
		assert.Nil(t, found)
	})

	t.Run("returns error for invalid UUID", func(t *testing.T) {
		_, err := repo.FindByID(ctx, "not-a-uuid")
		assert.Error(t, err)
	})
}

func TestTaskRepository_List(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	repo := NewTaskRepository(pool)
	ctx := context.Background()
	userID := createTestUser(t, ctx, pool)

	// Create some test tasks
	workCategory := "Work"
	personalCategory := "Personal"

	task1 := &domain.Task{
		ID:            uuid.New().String(),
		UserID:        userID,
		Title:         "Work Task 1",
		Category:      &workCategory,
		Status:        domain.TaskStatusTodo,
		UserPriority:  5,
		PriorityScore: 70,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	task2 := &domain.Task{
		ID:            uuid.New().String(),
		UserID:        userID,
		Title:         "Personal Task",
		Category:      &personalCategory,
		Status:        domain.TaskStatusInProgress,
		UserPriority:  3,
		PriorityScore: 50,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	task3 := &domain.Task{
		ID:            uuid.New().String(),
		UserID:        userID,
		Title:         "Work Task 2",
		Category:      &workCategory,
		Status:        domain.TaskStatusDone,
		UserPriority:  8,
		PriorityScore: 90,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	require.NoError(t, repo.Create(ctx, task1))
	require.NoError(t, repo.Create(ctx, task2))
	require.NoError(t, repo.Create(ctx, task3))

	t.Run("lists all tasks for user", func(t *testing.T) {
		filter := &domain.TaskListFilter{}
		tasks, err := repo.List(ctx, userID, filter)
		require.NoError(t, err)
		assert.Len(t, tasks, 3)
	})

	t.Run("filters by status", func(t *testing.T) {
		status := domain.TaskStatusTodo
		filter := &domain.TaskListFilter{Status: &status}
		tasks, err := repo.List(ctx, userID, filter)
		require.NoError(t, err)
		assert.Len(t, tasks, 1)
		assert.Equal(t, "Work Task 1", tasks[0].Title)
	})

	t.Run("filters by category", func(t *testing.T) {
		filter := &domain.TaskListFilter{Category: &workCategory}
		tasks, err := repo.List(ctx, userID, filter)
		require.NoError(t, err)
		assert.Len(t, tasks, 2)
	})

	t.Run("filters by min priority", func(t *testing.T) {
		minPriority := 60
		filter := &domain.TaskListFilter{MinPriority: &minPriority}
		tasks, err := repo.List(ctx, userID, filter)
		require.NoError(t, err)
		assert.Len(t, tasks, 2) // task1 (70) and task3 (90)
	})

	t.Run("applies limit", func(t *testing.T) {
		filter := &domain.TaskListFilter{Limit: 2}
		tasks, err := repo.List(ctx, userID, filter)
		require.NoError(t, err)
		assert.Len(t, tasks, 2)
	})

	t.Run("applies offset", func(t *testing.T) {
		filter := &domain.TaskListFilter{Limit: 10, Offset: 2}
		tasks, err := repo.List(ctx, userID, filter)
		require.NoError(t, err)
		assert.Len(t, tasks, 1) // Only 1 task after offset of 2
	})

	t.Run("orders by priority descending", func(t *testing.T) {
		filter := &domain.TaskListFilter{}
		tasks, err := repo.List(ctx, userID, filter)
		require.NoError(t, err)
		require.Len(t, tasks, 3)
		// Verify descending order by priority score
		assert.Equal(t, 90, tasks[0].PriorityScore) // task3
		assert.Equal(t, 70, tasks[1].PriorityScore) // task1
		assert.Equal(t, 50, tasks[2].PriorityScore) // task2
	})

	t.Run("returns empty list for user with no tasks", func(t *testing.T) {
		otherUserID := createTestUser(t, ctx, pool)
		filter := &domain.TaskListFilter{}
		tasks, err := repo.List(ctx, otherUserID, filter)
		require.NoError(t, err)
		assert.Empty(t, tasks)
	})
}

func TestTaskRepository_Update(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	repo := NewTaskRepository(pool)
	ctx := context.Background()
	userID := createTestUser(t, ctx, pool)

	t.Run("updates task successfully", func(t *testing.T) {
		task := createTestTask(t, ctx, repo, userID, "Original Title")

		// Update fields
		task.Title = "Updated Title"
		task.Status = domain.TaskStatusInProgress
		task.PriorityScore = 80
		task.UpdatedAt = time.Now().UTC()

		err := repo.Update(ctx, task)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, task.ID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Title", found.Title)
		assert.Equal(t, domain.TaskStatusInProgress, found.Status)
		assert.Equal(t, 80, found.PriorityScore)
	})

	t.Run("returns not found for non-existent task", func(t *testing.T) {
		task := &domain.Task{
			ID:            uuid.New().String(),
			UserID:        userID,
			Title:         "Non-existent",
			Status:        domain.TaskStatusTodo,
			UserPriority:  5,
			PriorityScore: 50,
			UpdatedAt:     time.Now().UTC(),
		}

		err := repo.Update(ctx, task)
		assert.ErrorIs(t, err, domain.ErrTaskNotFound)
	})

	t.Run("does not update other user's task", func(t *testing.T) {
		task := createTestTask(t, ctx, repo, userID, "User1 Task")

		// Try to update with different user ID
		otherUserID := createTestUser(t, ctx, pool)
		task.UserID = otherUserID
		task.Title = "Hacked Title"

		err := repo.Update(ctx, task)
		assert.ErrorIs(t, err, domain.ErrTaskNotFound) // Should fail because user_id doesn't match
	})
}

func TestTaskRepository_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	repo := NewTaskRepository(pool)
	ctx := context.Background()
	userID := createTestUser(t, ctx, pool)

	t.Run("deletes task successfully", func(t *testing.T) {
		task := createTestTask(t, ctx, repo, userID, "To Delete")

		err := repo.Delete(ctx, task.ID, userID)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, task.ID)
		assert.ErrorIs(t, err, domain.ErrTaskNotFound)
		assert.Nil(t, found)
	})

	t.Run("returns not found for non-existent task", func(t *testing.T) {
		err := repo.Delete(ctx, uuid.New().String(), userID)
		assert.ErrorIs(t, err, domain.ErrTaskNotFound)
	})

	t.Run("does not delete other user's task", func(t *testing.T) {
		task := createTestTask(t, ctx, repo, userID, "Protected Task")
		otherUserID := createTestUser(t, ctx, pool)

		err := repo.Delete(ctx, task.ID, otherUserID)
		assert.ErrorIs(t, err, domain.ErrTaskNotFound)

		// Verify task still exists
		found, err := repo.FindByID(ctx, task.ID)
		require.NoError(t, err)
		assert.NotNil(t, found)
	})
}

func TestTaskRepository_IncrementBumpCount(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	repo := NewTaskRepository(pool)
	ctx := context.Background()
	userID := createTestUser(t, ctx, pool)

	t.Run("increments bump count", func(t *testing.T) {
		task := createTestTask(t, ctx, repo, userID, "Bump Me")
		assert.Equal(t, 0, task.BumpCount)

		err := repo.IncrementBumpCount(ctx, task.ID, userID)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, task.ID)
		require.NoError(t, err)
		assert.Equal(t, 1, found.BumpCount)

		// Increment again
		err = repo.IncrementBumpCount(ctx, task.ID, userID)
		require.NoError(t, err)

		found, err = repo.FindByID(ctx, task.ID)
		require.NoError(t, err)
		assert.Equal(t, 2, found.BumpCount)
	})

	t.Run("returns not found for non-existent task", func(t *testing.T) {
		err := repo.IncrementBumpCount(ctx, uuid.New().String(), userID)
		assert.ErrorIs(t, err, domain.ErrTaskNotFound)
	})
}

func TestTaskRepository_FindAtRiskTasks(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	repo := NewTaskRepository(pool)
	ctx := context.Background()
	userID := createTestUser(t, ctx, pool)

	// Create tasks with various bump counts
	lowBumpTask := createTestTask(t, ctx, repo, userID, "Low Bump")
	highBumpTask := &domain.Task{
		ID:            uuid.New().String(),
		UserID:        userID,
		Title:         "High Bump (At Risk)",
		Status:        domain.TaskStatusTodo,
		UserPriority:  5,
		PriorityScore: 50,
		BumpCount:     3, // At risk threshold
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	require.NoError(t, repo.Create(ctx, highBumpTask))

	t.Run("returns tasks with 3+ bumps", func(t *testing.T) {
		atRisk, err := repo.FindAtRiskTasks(ctx, userID)
		require.NoError(t, err)
		require.Len(t, atRisk, 1)
		assert.Equal(t, "High Bump (At Risk)", atRisk[0].Title)
		assert.GreaterOrEqual(t, atRisk[0].BumpCount, 3)
	})

	t.Run("excludes low bump tasks", func(t *testing.T) {
		atRisk, err := repo.FindAtRiskTasks(ctx, userID)
		require.NoError(t, err)
		for _, task := range atRisk {
			assert.NotEqual(t, lowBumpTask.ID, task.ID)
		}
	})
}

func TestTaskRepository_GetCategories(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	repo := NewTaskRepository(pool)
	ctx := context.Background()
	userID := createTestUser(t, ctx, pool)

	t.Run("returns unique categories", func(t *testing.T) {
		cat1 := "Work"
		cat2 := "Personal"

		task1 := &domain.Task{
			ID:            uuid.New().String(),
			UserID:        userID,
			Title:         "Work 1",
			Category:      &cat1,
			Status:        domain.TaskStatusTodo,
			UserPriority:  5,
			PriorityScore: 50,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}
		task2 := &domain.Task{
			ID:            uuid.New().String(),
			UserID:        userID,
			Title:         "Work 2",
			Category:      &cat1, // Duplicate category
			Status:        domain.TaskStatusTodo,
			UserPriority:  5,
			PriorityScore: 50,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}
		task3 := &domain.Task{
			ID:            uuid.New().String(),
			UserID:        userID,
			Title:         "Personal",
			Category:      &cat2,
			Status:        domain.TaskStatusTodo,
			UserPriority:  5,
			PriorityScore: 50,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}

		require.NoError(t, repo.Create(ctx, task1))
		require.NoError(t, repo.Create(ctx, task2))
		require.NoError(t, repo.Create(ctx, task3))

		categories, err := repo.GetCategories(ctx, userID)
		require.NoError(t, err)
		assert.Len(t, categories, 2)
		assert.Contains(t, categories, "Work")
		assert.Contains(t, categories, "Personal")
	})

	t.Run("excludes null categories", func(t *testing.T) {
		// Task without category
		task := createTestTask(t, ctx, repo, userID, "No Category")

		categories, err := repo.GetCategories(ctx, userID)
		require.NoError(t, err)
		// Should not contain empty string
		for _, cat := range categories {
			assert.NotEmpty(t, cat)
		}
		_ = task // Silence unused variable warning
	})
}

func TestTaskRepository_FindByDateRange(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	repo := NewTaskRepository(pool)
	ctx := context.Background()
	userID := createTestUser(t, ctx, pool)

	// Create tasks with different due dates
	now := time.Now().UTC()
	tomorrow := now.Add(24 * time.Hour)
	nextWeek := now.Add(7 * 24 * time.Hour)
	nextMonth := now.Add(30 * 24 * time.Hour)

	taskToday := &domain.Task{
		ID:            uuid.New().String(),
		UserID:        userID,
		Title:         "Due Today",
		DueDate:       &now,
		Status:        domain.TaskStatusTodo,
		UserPriority:  5,
		PriorityScore: 50,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	taskNextWeek := &domain.Task{
		ID:            uuid.New().String(),
		UserID:        userID,
		Title:         "Due Next Week",
		DueDate:       &nextWeek,
		Status:        domain.TaskStatusTodo,
		UserPriority:  5,
		PriorityScore: 50,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	taskNextMonth := &domain.Task{
		ID:            uuid.New().String(),
		UserID:        userID,
		Title:         "Due Next Month",
		DueDate:       &nextMonth,
		Status:        domain.TaskStatusTodo,
		UserPriority:  5,
		PriorityScore: 50,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	require.NoError(t, repo.Create(ctx, taskToday))
	require.NoError(t, repo.Create(ctx, taskNextWeek))
	require.NoError(t, repo.Create(ctx, taskNextMonth))

	t.Run("finds tasks within date range", func(t *testing.T) {
		filter := &domain.CalendarFilter{
			StartDate: now.Add(-time.Hour),
			EndDate:   tomorrow,
		}

		tasks, err := repo.FindByDateRange(ctx, userID, filter)
		require.NoError(t, err)
		require.Len(t, tasks, 1)
		assert.Equal(t, "Due Today", tasks[0].Title)
	})

	t.Run("filters by status in date range", func(t *testing.T) {
		filter := &domain.CalendarFilter{
			StartDate: now.Add(-time.Hour),
			EndDate:   nextMonth.Add(time.Hour),
			Status:    []domain.TaskStatus{domain.TaskStatusTodo},
		}

		tasks, err := repo.FindByDateRange(ctx, userID, filter)
		require.NoError(t, err)
		assert.Len(t, tasks, 3)
	})
}

func TestTaskRepository_RenameCategoryForUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	repo := NewTaskRepository(pool)
	ctx := context.Background()
	userID := createTestUser(t, ctx, pool)

	t.Run("renames category for all user tasks", func(t *testing.T) {
		oldCat := "OldCategory"
		newCat := "NewCategory"

		task1 := &domain.Task{
			ID:            uuid.New().String(),
			UserID:        userID,
			Title:         "Task 1",
			Category:      &oldCat,
			Status:        domain.TaskStatusTodo,
			UserPriority:  5,
			PriorityScore: 50,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}
		task2 := &domain.Task{
			ID:            uuid.New().String(),
			UserID:        userID,
			Title:         "Task 2",
			Category:      &oldCat,
			Status:        domain.TaskStatusTodo,
			UserPriority:  5,
			PriorityScore: 50,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}

		require.NoError(t, repo.Create(ctx, task1))
		require.NoError(t, repo.Create(ctx, task2))

		count, err := repo.RenameCategoryForUser(ctx, userID, oldCat, newCat)
		require.NoError(t, err)
		assert.Equal(t, 2, count)

		// Verify rename
		categories, err := repo.GetCategories(ctx, userID)
		require.NoError(t, err)
		assert.Contains(t, categories, "NewCategory")
		assert.NotContains(t, categories, "OldCategory")
	})

	t.Run("does not rename completed tasks", func(t *testing.T) {
		cat := "DoneCategory"
		task := &domain.Task{
			ID:            uuid.New().String(),
			UserID:        userID,
			Title:         "Done Task",
			Category:      &cat,
			Status:        domain.TaskStatusDone,
			UserPriority:  5,
			PriorityScore: 50,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}
		require.NoError(t, repo.Create(ctx, task))

		count, err := repo.RenameCategoryForUser(ctx, userID, cat, "RenamedDone")
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func TestTaskRepository_DeleteCategoryForUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	repo := NewTaskRepository(pool)
	ctx := context.Background()
	userID := createTestUser(t, ctx, pool)

	t.Run("removes category from tasks", func(t *testing.T) {
		cat := "ToDelete"
		task := &domain.Task{
			ID:            uuid.New().String(),
			UserID:        userID,
			Title:         "Has Category",
			Category:      &cat,
			Status:        domain.TaskStatusTodo,
			UserPriority:  5,
			PriorityScore: 50,
			CreatedAt:     time.Now().UTC(),
			UpdatedAt:     time.Now().UTC(),
		}
		require.NoError(t, repo.Create(ctx, task))

		count, err := repo.DeleteCategoryForUser(ctx, userID, cat)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// Verify category is removed
		found, err := repo.FindByID(ctx, task.ID)
		require.NoError(t, err)
		assert.Nil(t, found.Category)
	})
}

func TestTaskRepository_GetCompletionStats(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	repo := NewTaskRepository(pool)
	ctx := context.Background()
	userID := createTestUser(t, ctx, pool)

	// Create tasks with various statuses
	task1 := &domain.Task{
		ID:            uuid.New().String(),
		UserID:        userID,
		Title:         "Done Task",
		Status:        domain.TaskStatusDone,
		UserPriority:  5,
		PriorityScore: 50,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	task2 := &domain.Task{
		ID:            uuid.New().String(),
		UserID:        userID,
		Title:         "Todo Task",
		Status:        domain.TaskStatusTodo,
		UserPriority:  5,
		PriorityScore: 50,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	task3 := &domain.Task{
		ID:            uuid.New().String(),
		UserID:        userID,
		Title:         "In Progress Task",
		Status:        domain.TaskStatusInProgress,
		UserPriority:  5,
		PriorityScore: 50,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	require.NoError(t, repo.Create(ctx, task1))
	require.NoError(t, repo.Create(ctx, task2))
	require.NoError(t, repo.Create(ctx, task3))

	t.Run("returns correct completion statistics", func(t *testing.T) {
		stats, err := repo.GetCompletionStats(ctx, userID, 30)
		require.NoError(t, err)
		assert.Equal(t, 3, stats.TotalTasks)
		assert.Equal(t, 1, stats.CompletedTasks)
		assert.Equal(t, 2, stats.PendingTasks)
		assert.InDelta(t, 33.33, stats.CompletionRate, 0.1)
	})
}

func TestTaskRepository_GetBumpAnalytics(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	repo := NewTaskRepository(pool)
	ctx := context.Background()
	userID := createTestUser(t, ctx, pool)

	// Create tasks with various bump counts
	task1 := &domain.Task{
		ID:            uuid.New().String(),
		UserID:        userID,
		Title:         "No Bumps",
		Status:        domain.TaskStatusTodo,
		UserPriority:  5,
		PriorityScore: 50,
		BumpCount:     0,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	task2 := &domain.Task{
		ID:            uuid.New().String(),
		UserID:        userID,
		Title:         "Two Bumps",
		Status:        domain.TaskStatusTodo,
		UserPriority:  5,
		PriorityScore: 50,
		BumpCount:     2,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	task3 := &domain.Task{
		ID:            uuid.New().String(),
		UserID:        userID,
		Title:         "At Risk",
		Status:        domain.TaskStatusTodo,
		UserPriority:  5,
		PriorityScore: 50,
		BumpCount:     4,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	require.NoError(t, repo.Create(ctx, task1))
	require.NoError(t, repo.Create(ctx, task2))
	require.NoError(t, repo.Create(ctx, task3))

	t.Run("returns bump analytics", func(t *testing.T) {
		analytics, err := repo.GetBumpAnalytics(ctx, userID)
		require.NoError(t, err)
		assert.NotNil(t, analytics.TasksByBumpCount)
		assert.GreaterOrEqual(t, analytics.AtRiskCount, 1)
	})
}

func TestTaskRepository_GetPriorityDistribution(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	repo := NewTaskRepository(pool)
	ctx := context.Background()
	userID := createTestUser(t, ctx, pool)

	// Create tasks with different priority scores
	lowPriority := &domain.Task{
		ID:            uuid.New().String(),
		UserID:        userID,
		Title:         "Low Priority",
		Status:        domain.TaskStatusTodo,
		UserPriority:  2,
		PriorityScore: 20,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	highPriority := &domain.Task{
		ID:            uuid.New().String(),
		UserID:        userID,
		Title:         "High Priority",
		Status:        domain.TaskStatusTodo,
		UserPriority:  9,
		PriorityScore: 90,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	require.NoError(t, repo.Create(ctx, lowPriority))
	require.NoError(t, repo.Create(ctx, highPriority))

	t.Run("returns priority distribution", func(t *testing.T) {
		distribution, err := repo.GetPriorityDistribution(ctx, userID)
		require.NoError(t, err)
		assert.NotEmpty(t, distribution)
	})
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
