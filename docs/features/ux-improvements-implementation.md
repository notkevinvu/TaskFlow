# UX Improvements: Implementation Plan

**Date:** 2025-12-08
**Status:** Ready for Implementation
**Spec:** [ux-improvements-spec.md](./ux-improvements-spec.md)

---

## Summary

| Phase | Description | Effort | Dependencies |
|-------|-------------|--------|--------------|
| 1 | Button Loading States | 1-2 hrs | None |
| 2 | Uncomplete API + Gamification Reversal | 2-3 hrs | None |
| 3 | Soft Delete + Undelete API | 1-2 hrs | None |
| 4 | Frontend Toast Undo Actions | 1 hr | Phases 2 & 3 |
| **Total** | | **6-8 hrs** | |

---

## Phase 1: Button Loading States

### Files to Modify

| File | Changes |
|------|---------|
| `frontend/components/ui/button.tsx` | Add `loading` prop |
| `frontend/components/TaskDetailsSidebar.tsx` | Use loading prop on buttons |
| `frontend/app/(dashboard)/page.tsx` | Use loading prop on quick actions |
| `frontend/components/CreateTaskDialog.tsx` | Add spinner to submit button |
| `frontend/components/EditTaskDialog.tsx` | Add spinner to save button |

### Implementation: button.tsx

```tsx
import * as React from "react"
import { Slot } from "@radix-ui/react-slot"
import { cva, type VariantProps } from "class-variance-authority"
import { Loader2 } from "lucide-react"
import { cn } from "@/lib/utils"

const buttonVariants = cva(
  // ... existing variants unchanged
)

function Button({
  className,
  variant,
  size,
  asChild = false,
  loading = false,
  children,
  disabled,
  ...props
}: React.ComponentProps<"button"> &
  VariantProps<typeof buttonVariants> & {
    asChild?: boolean
    loading?: boolean
  }) {
  const Comp = asChild ? Slot : "button"

  return (
    <Comp
      data-slot="button"
      className={cn(buttonVariants({ variant, size, className }))}
      disabled={disabled || loading}
      aria-busy={loading}
      {...props}
    >
      {loading && <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" />}
      {children}
    </Comp>
  )
}

export { Button, buttonVariants }
```

### Usage Examples

```tsx
// TaskDetailsSidebar.tsx - Complete button
<Button
  size="sm"
  onClick={() => completeTask.mutate(taskId)}
  loading={completeTask.isPending}
  disabled={hasAnyBlocker}
  className="transition-all hover:scale-105 hover:shadow-md cursor-pointer"
>
  Complete
</Button>

// Dashboard - Bump button
<Button
  variant="outline"
  size="sm"
  onClick={() => bumpTask.mutate({ id: taskId })}
  loading={bumpTask.isPending}
>
  Bump
</Button>

// CreateTaskDialog - Submit button
<Button type="submit" loading={createTask.isPending} disabled={!formData.title}>
  {createTask.isPending ? 'Creating...' : 'Create Task'}
</Button>
```

---

## Phase 2: Uncomplete API + Gamification Reversal

### New Endpoint

```
POST /api/v1/tasks/:id/uncomplete
```

**Response:** `200 OK` with updated task (status: "todo", completed_at: null)

### Files to Create/Modify

| File | Action | Changes |
|------|--------|---------|
| `backend/internal/domain/task_history.go` | Modify | Add `EventTaskUncompleted` constant |
| `backend/internal/ports/services.go` | Modify | Add `Uncomplete` to TaskService interface |
| `backend/internal/ports/services.go` | Modify | Add gamification reversal methods |
| `backend/internal/ports/repositories.go` | Modify | Add `DecrementCategoryMastery`, `RevokeAchievement` |
| `backend/internal/service/task_service.go` | Modify | Add `Uncomplete()` method |
| `backend/internal/service/gamification_service.go` | Modify | Add reversal methods |
| `backend/internal/repository/gamification_repository.go` | Modify | Add reversal queries |
| `backend/internal/handler/task_handler.go` | Modify | Add `Uncomplete()` handler |
| `backend/cmd/server/main.go` | Modify | Register route |

### Domain: New Event Type

```go
// backend/internal/domain/task_history.go
const (
    EventTaskCreated    TaskHistoryEventType = "created"
    EventTaskUpdated    TaskHistoryEventType = "updated"
    EventTaskCompleted  TaskHistoryEventType = "completed"
    EventTaskUncompleted TaskHistoryEventType = "uncompleted"  // NEW
    EventTaskBumped     TaskHistoryEventType = "bumped"
    EventTaskDeleted    TaskHistoryEventType = "deleted"
)
```

### Ports: Interface Updates

```go
// backend/internal/ports/services.go - TaskService interface
type TaskService interface {
    // ... existing methods ...
    Uncomplete(ctx context.Context, userID, taskID string) (*domain.Task, error)
}

// backend/internal/ports/services.go - GamificationService interface
type GamificationService interface {
    // ... existing methods ...
    ProcessTaskUncompletion(ctx context.Context, userID string, task *domain.Task) error
    ProcessTaskUncompletionAsync(userID string, task *domain.Task)
}

// backend/internal/ports/repositories.go - GamificationRepository interface
type GamificationRepository interface {
    // ... existing methods ...
    DecrementCategoryMastery(ctx context.Context, userID, category string) error
    RevokeAchievement(ctx context.Context, userID, achievementID string) error
}
```

### Service: TaskService.Uncomplete

```go
// backend/internal/service/task_service.go

// Uncomplete reverses a task completion
func (s *TaskService) Uncomplete(ctx context.Context, userID, taskID string) (*domain.Task, error) {
    task, err := s.taskRepo.FindByID(ctx, taskID)
    if err != nil {
        return nil, domain.NewInternalError("failed to find task", err)
    }
    if task == nil {
        return nil, domain.NewNotFoundError("task", taskID)
    }
    if task.UserID != userID {
        return nil, domain.NewForbiddenError("task", "access")
    }

    // Only allow uncomplete if task is done
    if task.Status != domain.TaskStatusDone {
        return nil, domain.NewValidationError("status", "task is not completed")
    }

    // Store previous state for history
    previousState := *task

    // Reverse gamification (async, before status change for accurate category)
    if s.gamificationService != nil {
        s.gamificationService.ProcessTaskUncompletionAsync(userID, task)
    }

    // Update task status
    task.Status = domain.TaskStatusTodo
    task.CompletedAt = nil
    task.UpdatedAt = time.Now()

    // Recalculate priority
    task.PriorityScore = s.priorityCalc.Calculate(task)

    // Save
    if err := s.taskRepo.Update(ctx, task); err != nil {
        return nil, domain.NewInternalError("failed to update task", err)
    }

    // Log history
    s.logHistory(ctx, userID, taskID, domain.EventTaskUncompleted, &previousState, task)

    return task, nil
}
```

### Service: GamificationService Reversal

```go
// backend/internal/service/gamification_service.go

// ProcessTaskUncompletion reverses gamification effects of a task completion
func (s *GamificationService) ProcessTaskUncompletion(
    ctx context.Context,
    userID string,
    task *domain.Task,
) error {
    // 1. Decrement category mastery if task had category
    if task.Category != nil && *task.Category != "" {
        if err := s.gamificationRepo.DecrementCategoryMastery(ctx, userID, *task.Category); err != nil {
            slog.Warn("Failed to decrement category mastery",
                "user_id", userID,
                "category", *task.Category,
                "error", err,
            )
        }
    }

    // 2. Recompute stats (will recalculate totals, streaks, etc. from actual data)
    stats, err := s.ComputeStats(ctx, userID)
    if err != nil {
        slog.Error("Failed to recompute stats during uncompletion",
            "user_id", userID,
            "error", err,
        )
        return err
    }

    // 3. Check achievements that may need revocation
    if err := s.RevokeInvalidAchievements(ctx, userID, stats); err != nil {
        slog.Warn("Failed to revoke invalid achievements",
            "user_id", userID,
            "error", err,
        )
    }

    return nil
}

// ProcessTaskUncompletionAsync runs uncompletion processing in background
func (s *GamificationService) ProcessTaskUncompletionAsync(userID string, task *domain.Task) {
    go func() {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        defer cancel()

        if err := s.ProcessTaskUncompletion(ctx, userID, task); err != nil {
            slog.Error("Async gamification uncompletion failed",
                "user_id", userID,
                "task_id", task.ID,
                "error", err,
            )
        }
    }()
}

// RevokeInvalidAchievements removes achievements user no longer qualifies for
func (s *GamificationService) RevokeInvalidAchievements(
    ctx context.Context,
    userID string,
    stats *domain.GamificationStats,
) error {
    achievements, err := s.gamificationRepo.GetAchievements(ctx, userID)
    if err != nil {
        return fmt.Errorf("failed to get achievements: %w", err)
    }

    for _, achievement := range achievements {
        def := domain.GetAchievementDefinition(achievement.AchievementID)
        if def == nil {
            continue
        }

        // Check if user still qualifies based on current stats
        stillQualifies := s.checkAchievementQualification(def, stats)

        if !stillQualifies {
            slog.Info("Revoking achievement - user no longer qualifies",
                "user_id", userID,
                "achievement_id", achievement.AchievementID,
            )
            if err := s.gamificationRepo.RevokeAchievement(ctx, userID, achievement.AchievementID); err != nil {
                slog.Warn("Failed to revoke achievement",
                    "achievement_id", achievement.AchievementID,
                    "error", err,
                )
            }
        }
    }

    return nil
}

// checkAchievementQualification checks if user still qualifies for an achievement
func (s *GamificationService) checkAchievementQualification(
    def *domain.AchievementDefinition,
    stats *domain.GamificationStats,
) bool {
    switch def.ID {
    case "first_task":
        return stats.TotalCompleted >= 1
    case "ten_tasks":
        return stats.TotalCompleted >= 10
    case "fifty_tasks":
        return stats.TotalCompleted >= 50
    case "hundred_tasks":
        return stats.TotalCompleted >= 100
    case "streak_3":
        return stats.LongestStreak >= 3
    case "streak_7":
        return stats.LongestStreak >= 7
    case "streak_14":
        return stats.LongestStreak >= 14
    case "streak_30":
        return stats.LongestStreak >= 30
    // Add other achievement checks as needed
    default:
        // For unknown achievements, keep them (conservative)
        return true
    }
}
```

### Repository: Gamification Reversal Queries

```go
// backend/internal/repository/gamification_repository.go

// DecrementCategoryMastery decreases the completion count for a category
func (r *GamificationRepository) DecrementCategoryMastery(ctx context.Context, userID, category string) error {
    query := `
        UPDATE category_mastery
        SET tasks_completed = GREATEST(tasks_completed - 1, 0),
            updated_at = NOW()
        WHERE user_id = $1 AND category = $2
    `
    _, err := r.db.Exec(ctx, query, userID, category)
    if err != nil {
        return fmt.Errorf("failed to decrement category mastery: %w", err)
    }
    return nil
}

// RevokeAchievement removes an earned achievement from a user
func (r *GamificationRepository) RevokeAchievement(ctx context.Context, userID, achievementID string) error {
    query := `
        DELETE FROM user_achievements
        WHERE user_id = $1 AND achievement_id = $2
    `
    _, err := r.db.Exec(ctx, query, userID, achievementID)
    if err != nil {
        return fmt.Errorf("failed to revoke achievement: %w", err)
    }
    return nil
}
```

### Handler: Uncomplete Endpoint

```go
// backend/internal/handler/task_handler.go

// Uncomplete handles reverting a completed task back to todo
// POST /api/v1/tasks/:id/uncomplete
func (h *TaskHandler) Uncomplete(c *gin.Context) {
    userID, exists := middleware.GetUserID(c)
    if !exists {
        middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
        return
    }

    taskID := c.Param("id")

    task, err := h.taskService.Uncomplete(c.Request.Context(), userID, taskID)
    if err != nil {
        middleware.AbortWithError(c, err)
        return
    }

    c.JSON(http.StatusOK, task)
}
```

### Route Registration

```go
// backend/cmd/server/main.go - in route setup
tasks.POST("/:id/uncomplete", taskHandler.Uncomplete)
```

---

## Phase 3: Soft Delete + Undelete API

### Database Migration

```sql
-- backend/migrations/000010_soft_delete.up.sql

-- Add soft delete column to tasks
ALTER TABLE tasks ADD COLUMN deleted_at TIMESTAMPTZ DEFAULT NULL;

-- Partial index for efficient queries on non-deleted tasks
CREATE INDEX idx_tasks_not_deleted ON tasks(user_id, status)
WHERE deleted_at IS NULL;

-- Index for finding deleted tasks (for admin/cleanup)
CREATE INDEX idx_tasks_deleted ON tasks(deleted_at)
WHERE deleted_at IS NOT NULL;

-- Update existing indexes to be partial (optional optimization)
-- This can be done in a follow-up migration if needed
```

```sql
-- backend/migrations/000010_soft_delete.down.sql

-- Hard delete any soft-deleted tasks first
DELETE FROM tasks WHERE deleted_at IS NOT NULL;

-- Remove column
ALTER TABLE tasks DROP COLUMN deleted_at;

-- Drop indexes (will be auto-dropped with column, but explicit is cleaner)
DROP INDEX IF EXISTS idx_tasks_not_deleted;
DROP INDEX IF EXISTS idx_tasks_deleted;
```

### Domain: Update Task Model

```go
// backend/internal/domain/task.go

type Task struct {
    ID              string      `json:"id"`
    UserID          string      `json:"user_id"`
    Title           string      `json:"title"`
    Description     *string     `json:"description,omitempty"`
    Status          TaskStatus  `json:"status"`
    // ... other existing fields ...
    CompletedAt     *time.Time  `json:"completed_at,omitempty"`
    DeletedAt       *time.Time  `json:"deleted_at,omitempty"` // NEW: Soft delete timestamp
}
```

### Ports: Interface Updates

```go
// backend/internal/ports/services.go
type TaskService interface {
    // ... existing methods ...
    Undelete(ctx context.Context, userID, taskID string) (*domain.Task, error)
}

// backend/internal/ports/repositories.go
type TaskRepository interface {
    // ... existing methods ...
    FindByIDIncludingDeleted(ctx context.Context, id string) (*domain.Task, error)
}
```

### Repository: Update Queries

All existing queries that fetch tasks need `WHERE deleted_at IS NULL` added:

```go
// backend/internal/repository/task_repository.go

// FindByID - add deleted_at check
func (r *TaskRepository) FindByID(ctx context.Context, id string) (*domain.Task, error) {
    query := `
        SELECT id, user_id, title, description, status, ...
        FROM tasks
        WHERE id = $1 AND deleted_at IS NULL  -- ADD THIS
    `
    // ...
}

// FindByIDIncludingDeleted - new method for undelete
func (r *TaskRepository) FindByIDIncludingDeleted(ctx context.Context, id string) (*domain.Task, error) {
    query := `
        SELECT id, user_id, title, description, status, ..., deleted_at
        FROM tasks
        WHERE id = $1  -- No deleted_at check
    `
    // ...
}

// List - add deleted_at check
func (r *TaskRepository) List(ctx context.Context, userID string, filter *domain.TaskListFilter) ([]*domain.Task, error) {
    // Add to WHERE clause: AND deleted_at IS NULL
}

// Similar updates needed for:
// - ListCompleted
// - GetCalendar
// - GetAtRisk
// - etc.
```

### Service: Update Delete + Add Undelete

```go
// backend/internal/service/task_service.go

// Delete now performs soft delete
func (s *TaskService) Delete(ctx context.Context, userID, taskID string) error {
    task, err := s.taskRepo.FindByID(ctx, taskID)
    if err != nil {
        return domain.NewInternalError("failed to find task", err)
    }
    if task == nil {
        return domain.NewNotFoundError("task", taskID)
    }
    if task.UserID != userID {
        return domain.NewForbiddenError("task", "access")
    }

    // Soft delete instead of hard delete
    now := time.Now()
    task.DeletedAt = &now
    task.UpdatedAt = now

    if err := s.taskRepo.Update(ctx, task); err != nil {
        return domain.NewInternalError("failed to soft delete task", err)
    }

    // Log history
    s.logHistory(ctx, userID, taskID, domain.EventTaskDeleted, task, nil)

    return nil
}

// Undelete restores a soft-deleted task
func (s *TaskService) Undelete(ctx context.Context, userID, taskID string) (*domain.Task, error) {
    task, err := s.taskRepo.FindByIDIncludingDeleted(ctx, taskID)
    if err != nil {
        return nil, domain.NewInternalError("failed to find task", err)
    }
    if task == nil {
        return nil, domain.NewNotFoundError("task", taskID)
    }
    if task.UserID != userID {
        return nil, domain.NewForbiddenError("task", "access")
    }
    if task.DeletedAt == nil {
        return nil, domain.NewValidationError("task", "task is not deleted")
    }

    // Restore
    task.DeletedAt = nil
    task.UpdatedAt = time.Now()

    if err := s.taskRepo.Update(ctx, task); err != nil {
        return nil, domain.NewInternalError("failed to restore task", err)
    }

    return task, nil
}
```

### Handler: Undelete Endpoint

```go
// backend/internal/handler/task_handler.go

// Undelete handles restoring a soft-deleted task
// POST /api/v1/tasks/:id/undelete
func (h *TaskHandler) Undelete(c *gin.Context) {
    userID, exists := middleware.GetUserID(c)
    if !exists {
        middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
        return
    }

    taskID := c.Param("id")

    task, err := h.taskService.Undelete(c.Request.Context(), userID, taskID)
    if err != nil {
        middleware.AbortWithError(c, err)
        return
    }

    c.JSON(http.StatusOK, task)
}
```

### Route Registration

```go
// backend/cmd/server/main.go
tasks.POST("/:id/undelete", taskHandler.Undelete)
```

---

## Phase 4: Frontend Toast Undo Actions

### API Client Updates

```typescript
// frontend/lib/api.ts

export const taskAPI = {
  // ... existing methods ...

  uncomplete: (id: string) =>
    api.post<TaskResponse>(`/tasks/${id}/uncomplete`),

  undelete: (id: string) =>
    api.post<TaskResponse>(`/tasks/${id}/undelete`),
};
```

### Hook: useCompleteTask with Undo

```typescript
// frontend/hooks/useTasks.ts

export function useCompleteTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => taskAPI.complete(id),
    onMutate: async (taskId) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries({ queryKey: taskKeys.lists() });

      // Snapshot previous value
      const previousTasks = queryClient.getQueryData(taskKeys.lists());

      // Optimistic update
      queryClient.setQueriesData(
        { queryKey: taskKeys.lists() },
        (old: TaskResponse[] | undefined) =>
          old?.map((task) =>
            task.id === taskId ? { ...task, status: 'done' as const } : task
          )
      );

      return { previousTasks };
    },
    onSuccess: (response, taskId) => {
      // Invalidate relevant queries
      queryClient.invalidateQueries({ queryKey: taskKeys.lists() });
      queryClient.invalidateQueries({ queryKey: taskKeys.completed() });
      queryClient.invalidateQueries({ queryKey: taskKeys.atRisk() });
      queryClient.invalidateQueries({ queryKey: gamificationKeys.all });
      queryClient.invalidateQueries({ queryKey: analyticsKeys.all });

      // Show toast with undo action
      toast.success('Task completed!', {
        duration: 5000,
        action: {
          label: 'Undo',
          onClick: async () => {
            try {
              await taskAPI.uncomplete(taskId);
              queryClient.invalidateQueries({ queryKey: taskKeys.lists() });
              queryClient.invalidateQueries({ queryKey: taskKeys.completed() });
              queryClient.invalidateQueries({ queryKey: gamificationKeys.all });
              toast.success('Task restored to active');
            } catch (err) {
              toast.error('Failed to undo completion');
            }
          },
        },
      });

      // Handle gamification feedback (existing logic)
      // ... achievement toasts, streak notifications, etc.
    },
    onError: (err, taskId, context) => {
      // Rollback on error
      if (context?.previousTasks) {
        queryClient.setQueriesData(
          { queryKey: taskKeys.lists() },
          context.previousTasks
        );
      }
      toast.error(getApiErrorMessage(err, 'Failed to complete task', 'Task Complete'));
    },
  });
}
```

### Hook: useDeleteTask with Undo

```typescript
// frontend/hooks/useTasks.ts

export function useDeleteTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => taskAPI.delete(id),
    onMutate: async (taskId) => {
      await queryClient.cancelQueries({ queryKey: taskKeys.lists() });

      const previousTasks = queryClient.getQueryData(taskKeys.lists());

      // Optimistic removal
      queryClient.setQueriesData(
        { queryKey: taskKeys.lists() },
        (old: TaskResponse[] | undefined) =>
          old?.filter((task) => task.id !== taskId)
      );

      return { previousTasks };
    },
    onSuccess: (_, taskId) => {
      queryClient.invalidateQueries({ queryKey: taskKeys.lists() });
      queryClient.invalidateQueries({ queryKey: analyticsKeys.all });

      toast.success('Task deleted', {
        duration: 5000,
        action: {
          label: 'Undo',
          onClick: async () => {
            try {
              await taskAPI.undelete(taskId);
              queryClient.invalidateQueries({ queryKey: taskKeys.lists() });
              toast.success('Task restored');
            } catch (err) {
              toast.error('Failed to restore task');
            }
          },
        },
      });
    },
    onError: (err, taskId, context) => {
      if (context?.previousTasks) {
        queryClient.setQueriesData(
          { queryKey: taskKeys.lists() },
          context.previousTasks
        );
      }
      toast.error(getApiErrorMessage(err, 'Failed to delete task', 'Task Delete'));
    },
  });
}
```

---

## Testing Checklist

### Phase 1: Button Loading
- [ ] Button shows spinner when `loading={true}`
- [ ] Button is disabled during loading
- [ ] Spinner animates correctly
- [ ] Text remains visible alongside spinner
- [ ] Works with all button variants (default, destructive, outline, etc.)

### Phase 2: Uncomplete API
- [ ] `POST /tasks/:id/uncomplete` returns 200 with updated task
- [ ] Task status changes from "done" to "todo"
- [ ] `completed_at` is cleared (null)
- [ ] Priority score is recalculated
- [ ] Task history shows "uncompleted" event
- [ ] Category mastery decremented
- [ ] Stats recomputed correctly
- [ ] Invalid achievements revoked
- [ ] Error returned if task not completed

### Phase 3: Soft Delete
- [ ] Migration applies cleanly
- [ ] Delete sets `deleted_at` instead of removing row
- [ ] Deleted tasks excluded from list queries
- [ ] Deleted tasks excluded from completed queries
- [ ] `POST /tasks/:id/undelete` restores task
- [ ] Error returned if task not deleted

### Phase 4: Toast Undo
- [ ] Completion toast shows "Undo" button
- [ ] Toast visible for 5 seconds
- [ ] Clicking Undo calls uncomplete API
- [ ] Success toast shown after undo
- [ ] Error toast shown if undo fails
- [ ] Same behavior for delete toast

---

## Rollback Plan

### If Button Loading causes issues:
- Remove `loading` prop from Button component
- Revert to `disabled={isPending}` pattern

### If Uncomplete API causes issues:
- Remove route from main.go
- Keep service code for future use

### If Soft Delete causes issues:
- Run down migration: `000010_soft_delete.down.sql`
- Revert repository changes
- Note: This will hard-delete any soft-deleted tasks

### If Toast Undo causes issues:
- Remove `action` prop from toast calls
- Revert to simple toast.success() messages
