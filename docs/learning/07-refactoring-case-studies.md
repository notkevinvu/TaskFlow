# Module 07: Refactoring Case Studies

## Learning Objectives

By the end of this module, you will:
- Learn from real refactoring decisions made during development
- Understand the N+1 query anti-pattern and how to fix it
- See synchronous → asynchronous transformations
- Analyze the impact of refactoring decisions

---

## Case Study 1: Raw SQL → sqlc

### The Problem

Phase 2 used manual SQL with error-prone row scanning:

```go
// BEFORE (Phase 2)
// backend/internal/repository/task_repository.go

func (r *TaskRepository) FindByID(ctx context.Context, id string) (*domain.Task, error) {
    query := `
        SELECT id, user_id, title, description, status, user_priority,
               priority_score, estimated_effort, due_date, category,
               bump_count, created_at, updated_at, completed_at
        FROM tasks
        WHERE id = $1 AND deleted_at IS NULL
    `

    row := r.db.QueryRow(ctx, query, id)

    var task domain.Task
    var description, category sql.NullString
    var effort sql.NullString
    var dueDate, completedAt sql.NullTime

    err := row.Scan(
        &task.ID,
        &task.UserID,
        &task.Title,
        &description,           // Handle NULL
        &task.Status,
        &task.UserPriority,
        &task.PriorityScore,
        &effort,                // Handle NULL
        &dueDate,               // Handle NULL
        &category,              // Handle NULL
        &task.BumpCount,
        &task.CreatedAt,
        &task.UpdatedAt,
        &completedAt,           // Handle NULL
    )
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, domain.NewNotFoundError("task", id)
        }
        return nil, err
    }

    // Convert NULLs to Go pointers
    if description.Valid {
        task.Description = &description.String
    }
    if category.Valid {
        task.Category = &category.String
    }
    if effort.Valid {
        e := domain.TaskEffort(effort.String)
        task.EstimatedEffort = &e
    }
    if dueDate.Valid {
        task.DueDate = &dueDate.Time
    }
    if completedAt.Valid {
        task.CompletedAt = &completedAt.Time
    }

    return &task, nil
}
```

**Problems:**
1. 40+ lines for one query
2. Column order must match SELECT order (brittle)
3. NULL handling is verbose and error-prone
4. No compile-time validation of SQL

### The Solution

sqlc generates type-safe Go code from SQL:

```sql
-- backend/internal/sqlc/queries/tasks.sql

-- name: GetTaskByID :one
SELECT id, user_id, title, description, status, user_priority,
       priority_score, estimated_effort, due_date, category,
       bump_count, task_type, parent_task_id, series_id,
       created_at, updated_at, completed_at, deleted_at
FROM tasks
WHERE id = $1 AND deleted_at IS NULL;
```

```go
// AFTER (Phase 3)
// backend/internal/repository/task_repository.go

func (r *TaskRepository) FindByID(ctx context.Context, id string) (*domain.Task, error) {
    row, err := r.queries.GetTaskByID(ctx, stringToPgtypeUUID(id))
    if err != nil {
        return nil, convertError(err)
    }
    return rowToTask(row), nil
}
```

### Impact

| Metric | Before | After | Change |
|--------|--------|-------|--------|
| Lines of code | ~40 per query | ~5 per query | -87% |
| NULL handling | Manual | Auto-generated | Eliminated |
| Type safety | Runtime | Compile-time | Bugs caught earlier |
| Refactoring risk | High | Low | Schema changes fail fast |

---

## Case Study 2: N+1 Query Elimination

### The Problem

When completing a blocker task, we need to find tasks that become unblocked:

```go
// BEFORE (Phase 5B.2)
// backend/internal/service/dependency_service.go

func (s *DependencyService) GetBlockerCompletionInfo(ctx context.Context, blockerTaskID string) (*domain.BlockerCompletionInfo, error) {
    // Query 1: Get all tasks blocked by this task
    blockedTaskIDs, err := s.repo.GetTasksBlockedBy(ctx, blockerTaskID)
    if err != nil {
        return nil, err
    }

    var unblockedIDs []string

    // N queries: Check each blocked task's remaining blockers
    for _, taskID := range blockedTaskIDs {
        // This runs once PER BLOCKED TASK!
        incompleteCount, err := s.repo.CountIncompleteBlockers(ctx, taskID)
        if err != nil {
            return nil, err
        }
        if incompleteCount == 0 {
            unblockedIDs = append(unblockedIDs, taskID)
        }
    }

    return &domain.BlockerCompletionInfo{
        BlockedTasks:   blockedTaskIDs,
        UnblockedTasks: unblockedIDs,
    }, nil
}
```

**The N+1 Problem Visualized:**

```
Complete Task "Fix Bug" (blocks 5 tasks)
│
├── Query 1: Get blocked task IDs
│   └── Returns: [Task-A, Task-B, Task-C, Task-D, Task-E]
│
├── Query 2: Count blockers for Task-A
├── Query 3: Count blockers for Task-B
├── Query 4: Count blockers for Task-C
├── Query 5: Count blockers for Task-D
└── Query 6: Count blockers for Task-E

Total: 6 queries (1 + N where N=5)
```

If a task blocks 50 tasks, that's **51 queries** for one operation!

### The Solution

Batch query using PostgreSQL's `ANY()` operator:

```sql
-- backend/internal/sqlc/queries/dependencies.sql

-- name: CountIncompleteBlockersBatch :many
SELECT td.task_id, COUNT(*)::int as incomplete_count
FROM task_dependencies td
INNER JOIN tasks t ON t.id = td.blocked_by_id
WHERE td.task_id = ANY($1::uuid[])
  AND t.status != 'done'
  AND t.deleted_at IS NULL
GROUP BY td.task_id;
```

```go
// AFTER (PR #77 + #83)
// backend/internal/repository/dependency_repository.go

func (r *DependencyRepository) CountIncompleteBlockersBatch(ctx context.Context, taskIDs []string) (map[string]int, error) {
    // Initialize all IDs with 0 (important for tasks with no blockers)
    result := make(map[string]int)
    for _, id := range taskIDs {
        result[id] = 0
    }

    if len(taskIDs) == 0 {
        return result, nil
    }

    // Convert to pgtype UUIDs
    uuids := make([]pgtype.UUID, 0, len(taskIDs))
    for _, id := range taskIDs {
        uuid, err := stringToPgtypeUUID(id)
        if err != nil {
            return nil, err
        }
        uuids = append(uuids, uuid)
    }

    // SINGLE query for ALL task IDs
    rows, err := r.queries.CountIncompleteBlockersBatch(ctx, uuids)
    if err != nil {
        return nil, err
    }

    // Populate results
    for _, row := range rows {
        taskID := pgtypeUUIDToString(row.TaskID)
        result[taskID] = int(row.IncompleteCount)
    }

    return result, nil
}
```

### Impact

| Blocked Tasks | Before (N+1) | After (Batch) | Reduction |
|---------------|--------------|---------------|-----------|
| 5 | 6 queries | 2 queries | 67% |
| 10 | 11 queries | 2 queries | 82% |
| 50 | 51 queries | 2 queries | 96% |
| 100 | 101 queries | 2 queries | 98% |

**Always 2 queries** regardless of N!

---

## Case Study 3: Synchronous → Asynchronous Gamification

### The Problem

Task completion was slow because gamification ran synchronously:

```go
// BEFORE (Phase 5B.3)
// backend/internal/handler/task_handler.go

func (h *TaskHandler) Complete(c *gin.Context) {
    // ... validation ...

    task, err := h.taskService.Complete(c.Request.Context(), userID, taskID)
    if err != nil {
        handleError(c, err)
        return
    }

    // BLOCKING: Gamification adds 50-100ms to response
    var gamResult *domain.GamificationResult
    if h.gamificationService != nil {
        gamResult, _ = h.gamificationService.ProcessTaskCompletion(
            c.Request.Context(), userID, task,
        )
    }

    c.JSON(http.StatusOK, gin.H{
        "task":         task,
        "gamification": gamResult,
    })
}
```

**Timeline (before):**
```
Request Start ──────────────────────────────────────────► Response
              │                                          │
              ├── Complete task (20ms)                   │
              ├── Update priority (10ms)                 │
              ├── Create history (15ms)                  │
              ├── Calculate XP (50ms)     ← SLOW!        │
              ├── Check achievements (80ms) ← SLOW!      │
              ├── Update streak (30ms)                   │
              └── Format response (5ms)                  │
              │                                          │
              Total: ~210ms (user waits!)                │
```

### The Solution

Move gamification to a background goroutine:

```go
// AFTER (PR #67)
// backend/internal/service/gamification_service.go

// Async wrapper - fire and forget
func (s *GamificationService) ProcessTaskCompletionAsync(userID string, task *domain.Task) {
    go func() {
        // Background context (independent of request)
        ctx := context.Background()

        // Process gamification
        result, err := s.ProcessTaskCompletion(ctx, userID, task)
        if err != nil {
            slog.Error("Gamification processing failed",
                "user_id", userID,
                "task_id", task.ID,
                "error", err,
            )
            return
        }

        // Log achievements earned
        for _, achievement := range result.NewAchievements {
            slog.Info("Achievement earned",
                "user_id", userID,
                "achievement", achievement.Type,
            )
        }
    }()
}
```

```go
// backend/internal/service/task_service.go

func (s *TaskService) Complete(ctx context.Context, userID, taskID string) (*domain.Task, error) {
    // ... complete task logic ...

    // NON-BLOCKING: Fire and forget
    if s.gamificationService != nil {
        s.gamificationService.ProcessTaskCompletionAsync(userID, task)
    }

    return task, nil
}
```

**Timeline (after):**
```
Request Start ──────────────► Response (50ms)
              │                │
              ├── Complete (20ms)
              ├── Update (10ms)
              ├── History (15ms)
              └── Response (5ms)
                               │
                               │ (User sees response)
                               │
              Background ──────┼──────────────────►
                               │
                               ├── Calculate XP (50ms)
                               ├── Achievements (80ms)
                               └── Update streak (30ms)
```

### Impact

| Metric | Before | After | Change |
|--------|--------|-------|--------|
| Response time | ~210ms | ~50ms | -76% |
| User-perceived latency | 210ms | 50ms | 4x faster |
| Gamification processing | Same | Same | (still happens) |

**Tradeoff:** User doesn't see achievements immediately. Solution: Poll for new achievements or show on next page load.

---

## Case Study 4: Interface-Based Dependency Injection

### The Problem

Handlers depended on concrete types, making testing difficult:

```go
// BEFORE (Phase 2)
// backend/internal/handler/task_handler.go

type TaskHandler struct {
    taskService *service.TaskService  // Concrete type!
}

func NewTaskHandler(ts *service.TaskService) *TaskHandler {
    return &TaskHandler{taskService: ts}
}
```

**Testing problem:**
```go
// Can't test without real database!
func TestTaskHandler_Create(t *testing.T) {
    // Need real DB connection
    db := setupTestDatabase()
    repo := repository.NewTaskRepository(db)
    service := service.NewTaskService(repo, ...)
    handler := handler.NewTaskHandler(service)

    // Test is slow, flaky, needs cleanup
}
```

### The Solution

Depend on interfaces instead:

```go
// AFTER (Phase 3, PR #11)
// backend/internal/handler/task_handler.go

type TaskHandler struct {
    taskService ports.TaskService  // Interface!
}

func NewTaskHandler(ts ports.TaskService) *TaskHandler {
    return &TaskHandler{taskService: ts}
}
```

```go
// backend/internal/ports/services.go

type TaskService interface {
    Create(ctx context.Context, userID string, dto *domain.CreateTaskDTO) (*domain.Task, error)
    Get(ctx context.Context, userID, taskID string) (*domain.Task, error)
    List(ctx context.Context, userID string, filter *domain.TaskListFilter) ([]*domain.Task, error)
    Update(ctx context.Context, userID, taskID string, dto *domain.UpdateTaskDTO) (*domain.Task, error)
    Delete(ctx context.Context, userID, taskID string) error
    Complete(ctx context.Context, userID, taskID string) (*domain.Task, error)
    Bump(ctx context.Context, userID, taskID string) (*domain.Task, error)
}
```

**Testing with mocks:**
```go
// AFTER: Fast, isolated tests
type MockTaskService struct {
    tasks map[string]*domain.Task
    err   error
}

func (m *MockTaskService) Create(ctx context.Context, userID string, dto *domain.CreateTaskDTO) (*domain.Task, error) {
    if m.err != nil {
        return nil, m.err
    }
    task := &domain.Task{ID: "test-id", Title: dto.Title}
    m.tasks[task.ID] = task
    return task, nil
}

func TestTaskHandler_Create(t *testing.T) {
    mock := &MockTaskService{tasks: make(map[string]*domain.Task)}
    handler := NewTaskHandler(mock)

    // Fast, no database needed!
    // Test runs in milliseconds
}
```

### Impact

| Metric | Before | After |
|--------|--------|-------|
| Test speed | ~2s (with DB) | ~10ms (with mock) |
| Test isolation | Poor (shared DB) | Excellent (isolated) |
| Test reliability | Flaky | Stable |
| Flexibility | Low | High (can swap implementations) |

---

## Refactoring Pattern: Build → Use → Optimize

All case studies follow this pattern:

```
Phase 1-2: BUILD
├── Focus on functionality
├── "Make it work"
└── Accept some technical debt

Phase 3: USE
├── Discover pain points
├── Measure performance
└── Identify anti-patterns

Phase 4+: OPTIMIZE
├── Refactor with knowledge
├── "Make it right"
└── Measure improvement
```

**Example Timeline:**
- **Nov 22 (Build):** Raw SQL queries work fine
- **Dec 1 (Use):** Notice boilerplate pain, NULL handling bugs
- **Dec 1 (Optimize):** Migrate to sqlc, eliminate bugs

---

## Exercises

### 🔰 Beginner: Identify N+1

Find potential N+1 patterns in this code:
```go
func (h *Handler) GetDashboard(c *gin.Context) {
    tasks, _ := h.taskService.List(ctx, userID, nil)

    for _, task := range tasks {
        subtasks, _ := h.subtaskService.GetSubtasks(ctx, task.ID)
        task.Subtasks = subtasks
    }

    c.JSON(200, tasks)
}
```

How would you fix it?

### 🎯 Intermediate: Design Batch Query

Design a batch query for this N+1 pattern:
```go
for _, user := range users {
    stats, _ := repo.GetUserStats(ctx, user.ID)
    user.Stats = stats
}
```

Write the SQL and Go code.

### 🚀 Advanced: Async with Retries

Extend the async gamification pattern to include:
- Retry on failure (max 3 attempts)
- Exponential backoff
- Dead-letter logging for permanent failures

---

## Reflection Questions

1. **When should you NOT refactor?** What signals indicate refactoring is premature?

2. **How do you measure refactoring success?** What metrics matter?

3. **Async processing has tradeoffs.** What problems can arise? How do you mitigate them?

4. **Interface-based DI adds indirection.** When is the complexity not worth it?

---

## Key Takeaways

1. **Build → Use → Optimize.** Don't over-engineer upfront.

2. **N+1 queries scale linearly.** Batch queries scale constantly.

3. **Async for slow, non-critical operations.** But handle failures!

4. **Interfaces enable testing.** The abstraction cost is usually worth it.

5. **Measure before and after.** Refactoring without metrics is guessing.

---

## Next Module

Continue to **[Module 08: Technical Decisions](./08-technical-decisions.md)** to understand the tradeoffs behind major decisions.
