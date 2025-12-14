# Module 05: Database Design

## Learning Objectives

By the end of this module, you will:
- Understand schema evolution through migrations
- Learn indexing strategies for query performance
- See soft delete and restore patterns
- Design migrations for new features

---

## Schema Overview

TaskFlow uses PostgreSQL 16 with 12 migrations that evolved the schema over time:

```
┌──────────────────────────────────────────────────────────────────┐
│                         Core Tables                               │
├──────────────┬───────────────┬───────────────┬───────────────────┤
│    users     │     tasks     │ task_history  │ task_dependencies │
│              │               │               │                   │
│ - id (PK)    │ - id (PK)     │ - id (PK)     │ - task_id (PK)    │
│ - email      │ - user_id (FK)│ - task_id(FK) │ - blocked_by(PK)  │
│ - name       │ - title       │ - event_type  │ - created_at      │
│ - password   │ - status      │ - old_value   │                   │
│ - user_type  │ - priority    │ - new_value   │                   │
│ - created_at │ - due_date    │ - created_at  │                   │
│              │ - category    │               │                   │
│              │ - parent_id   │               │                   │
│              │ - series_id   │               │                   │
└──────────────┴───────────────┴───────────────┴───────────────────┘

┌──────────────────────────────────────────────────────────────────┐
│                      Supporting Tables                            │
├──────────────┬───────────────┬───────────────┬───────────────────┤
│ task_series  │task_templates │user_preferences│ gamification     │
│              │               │               │                   │
│ - id (PK)    │ - id (PK)     │ - user_id(PK) │ - user_id (PK)   │
│ - user_id    │ - user_id     │ - defaults    │ - total_xp       │
│ - pattern    │ - name        │               │ - streak         │
│ - interval   │ - template    │               │ - level          │
│ - is_active  │               │               │                   │
└──────────────┴───────────────┴───────────────┴───────────────────┘
```

---

## Migration Timeline

| # | Migration | Date | Purpose |
|---|-----------|------|---------|
| 001 | initial_schema | Nov 22 | Users, tasks, task_history |
| 002 | update_priority_scale | Nov 25 | Change priority range |
| 003 | composite_indexes | Nov 27 | Performance optimization |
| 004 | recurring_tasks | Dec 1 | Task series support |
| 005 | subtasks_support | Dec 5 | Parent-child relationships |
| 006 | task_dependencies | Dec 5 | Blocked-by relationships |
| 007 | task_templates | Dec 6 | Template storage |
| 008 | gamification | Dec 6 | Achievements, streaks |
| 009 | anonymous_users | Dec 7 | Guest mode support |
| 010 | soft_delete | Dec 8 | deleted_at column |
| 011 | task_statuses | Dec 8 | Add new statuses |
| 012 | optimize_active_task_index | Dec 11 | Partial index for active tasks |

---

## Initial Schema (Migration 001)

```sql
-- backend/migrations/000001_initial_schema.up.sql

-- Custom enum types (PostgreSQL feature)
CREATE TYPE task_status AS ENUM ('todo', 'in_progress', 'done');
CREATE TYPE task_effort AS ENUM ('small', 'medium', 'large', 'xlarge');
CREATE TYPE event_type AS ENUM ('created', 'updated', 'deleted', 'status_changed', 'bumped');

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Tasks table
CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status task_status NOT NULL DEFAULT 'todo',
    user_priority INTEGER NOT NULL DEFAULT 5 CHECK (user_priority >= 1 AND user_priority <= 10),
    priority_score INTEGER NOT NULL DEFAULT 50 CHECK (priority_score >= 0 AND priority_score <= 100),
    estimated_effort task_effort,
    due_date TIMESTAMPTZ,
    category VARCHAR(50),
    bump_count INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    -- Full-text search vector
    search_vector TSVECTOR
);

-- Audit log
CREATE TABLE task_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    event_type event_type NOT NULL,
    old_value TEXT,
    new_value TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_tasks_user_id ON tasks(user_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_priority_score ON tasks(priority_score DESC);
CREATE INDEX idx_tasks_search_vector ON tasks USING GIN(search_vector);
CREATE INDEX idx_task_history_task_id ON task_history(task_id);

-- Trigger for search vector auto-update
CREATE OR REPLACE FUNCTION update_task_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW.title, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.description, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.category, '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_task_search_vector
    BEFORE INSERT OR UPDATE ON tasks
    FOR EACH ROW
    EXECUTE FUNCTION update_task_search_vector();
```

### Key Design Decisions

**1. UUIDs for Primary Keys:**
```sql
id UUID PRIMARY KEY DEFAULT gen_random_uuid()
```
- No sequential IDs (security: can't guess other user's task IDs)
- Works with distributed systems (no central ID server needed)

**2. PostgreSQL ENUMs:**
```sql
CREATE TYPE task_status AS ENUM ('todo', 'in_progress', 'done');
```
- Type-safe at database level
- Smaller storage than VARCHAR
- Fails fast on invalid values

**3. Full-Text Search with Weights:**
```sql
setweight(to_tsvector('english', title), 'A') ||  -- Title matches rank highest
setweight(to_tsvector('english', description), 'B') ||  -- Then description
setweight(to_tsvector('english', category), 'C')  -- Then category
```

---

## Subtasks Migration (005)

```sql
-- backend/migrations/000005_subtasks_support.up.sql

-- Add task type enum
ALTER TYPE task_status ADD VALUE IF NOT EXISTS 'on_hold';
ALTER TYPE task_status ADD VALUE IF NOT EXISTS 'blocked';

CREATE TYPE task_type AS ENUM ('regular', 'recurring', 'subtask');

-- Add columns to tasks
ALTER TABLE tasks
ADD COLUMN task_type task_type NOT NULL DEFAULT 'regular',
ADD COLUMN parent_task_id UUID REFERENCES tasks(id) ON DELETE CASCADE;

-- Index for subtask lookups
CREATE INDEX idx_tasks_parent_task_id ON tasks(parent_task_id) WHERE parent_task_id IS NOT NULL;

-- Function to count incomplete subtasks
CREATE OR REPLACE FUNCTION count_incomplete_subtasks(parent_id UUID)
RETURNS INTEGER AS $$
BEGIN
    RETURN (
        SELECT COUNT(*)
        FROM tasks
        WHERE parent_task_id = parent_id
          AND status != 'done'
          AND deleted_at IS NULL
    );
END;
$$ LANGUAGE plpgsql;
```

### Subtask Constraints

**Single-level nesting:** Subtasks can't have subtasks (prevents complexity):
```sql
-- Enforced in application layer, not database
-- backend/internal/service/task_service.go
if dto.ParentTaskID != nil {
    parent, _ := s.taskRepo.FindByID(ctx, *dto.ParentTaskID)
    if parent.TaskType == domain.TaskTypeSubtask {
        return nil, domain.NewValidationError("parent_task_id", "subtasks cannot have subtasks")
    }
}
```

---

## Dependencies Migration (006)

```sql
-- backend/migrations/000006_task_dependencies.up.sql

CREATE TABLE task_dependencies (
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    blocked_by_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Composite primary key
    PRIMARY KEY (task_id, blocked_by_id),

    -- Prevent self-dependency at database level
    CONSTRAINT no_self_dependency CHECK (task_id != blocked_by_id)
);

-- Indexes for lookups in both directions
CREATE INDEX idx_dependencies_task_id ON task_dependencies(task_id);
CREATE INDEX idx_dependencies_blocked_by ON task_dependencies(blocked_by_id);
```

### Cycle Detection

Cycle detection happens in application layer (DFS algorithm):

```go
// backend/internal/service/dependency_service.go

func (s *DependencyService) wouldCreateCycle(ctx context.Context, taskID, blockedByID string) (bool, error) {
    // Get entire dependency graph
    graph, err := s.repo.GetDependencyGraph(ctx, userID)
    if err != nil {
        return false, err
    }

    // Add proposed edge
    if graph[taskID] == nil {
        graph[taskID] = []string{}
    }
    graph[taskID] = append(graph[taskID], blockedByID)

    // DFS to detect cycle
    visited := make(map[string]bool)
    recStack := make(map[string]bool)

    var hasCycle func(node string) bool
    hasCycle = func(node string) bool {
        visited[node] = true
        recStack[node] = true

        for _, neighbor := range graph[node] {
            if !visited[neighbor] {
                if hasCycle(neighbor) {
                    return true
                }
            } else if recStack[neighbor] {
                return true  // Back edge found = cycle!
            }
        }

        recStack[node] = false
        return false
    }

    return hasCycle(taskID), nil
}
```

---

## Partial Index Optimization (012)

```sql
-- backend/migrations/000012_optimize_active_task_index.up.sql

-- Partial index: only includes non-deleted tasks
CREATE INDEX idx_tasks_user_active_priority
ON tasks (user_id, priority_score DESC)
WHERE deleted_at IS NULL;
```

### Why Partial Indexes?

```
Full Index (all rows):
┌─────────────────────────────────────┐
│ 100,000 rows indexed                │
│ - 95,000 active tasks               │
│ -  5,000 deleted tasks (never queried!) │
└─────────────────────────────────────┘

Partial Index (active only):
┌─────────────────────────────────────┐
│ 95,000 rows indexed                 │
│ - Smaller index size                │
│ - Faster scans                      │
│ - Less maintenance overhead         │
└─────────────────────────────────────┘
```

Most queries filter by `WHERE deleted_at IS NULL`, so indexing deleted tasks is waste.

---

## Soft Delete Pattern

```sql
-- Migration 010: Add deleted_at column
ALTER TABLE tasks ADD COLUMN deleted_at TIMESTAMPTZ;
```

### Application Layer Handling

```go
// backend/internal/repository/task_repository.go

// Default: exclude deleted tasks
func (r *TaskRepository) FindByID(ctx context.Context, id string) (*domain.Task, error) {
    return r.queries.GetTaskByID(ctx, id)
    // Query includes: WHERE deleted_at IS NULL
}

// Explicit: include deleted tasks (for restore)
func (r *TaskRepository) FindByIDIncludingDeleted(ctx context.Context, id string) (*domain.Task, error) {
    return r.queries.GetTaskByIDIncludingDeleted(ctx, id)
    // Query omits the deleted_at filter
}

// Soft delete
func (r *TaskRepository) Delete(ctx context.Context, id, userID string) error {
    return r.queries.SoftDeleteTask(ctx, sqlc.SoftDeleteTaskParams{
        ID:        stringToPgtypeUUID(id),
        UserID:    stringToPgtypeUUID(userID),
        DeletedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
    })
}

// Restore
func (r *TaskRepository) Restore(ctx context.Context, id, userID string) error {
    return r.queries.RestoreTask(ctx, sqlc.RestoreTaskParams{
        ID:     stringToPgtypeUUID(id),
        UserID: stringToPgtypeUUID(userID),
    })
    // Query: UPDATE tasks SET deleted_at = NULL WHERE ...
}
```

---

## Index Strategy

### Single Column Indexes

```sql
-- High-cardinality columns frequently filtered
CREATE INDEX idx_tasks_user_id ON tasks(user_id);
CREATE INDEX idx_tasks_status ON tasks(status);
```

### Composite Indexes

```sql
-- For queries that filter by user AND sort by priority
CREATE INDEX idx_tasks_user_priority ON tasks(user_id, priority_score DESC);

-- For queries that filter by user AND date range
CREATE INDEX idx_tasks_user_created ON tasks(user_id, created_at DESC);
```

### GIN Indexes (Full-Text Search)

```sql
CREATE INDEX idx_tasks_search_vector ON tasks USING GIN(search_vector);
```

### Partial Indexes (Filtered Subsets)

```sql
-- Only active tasks (most common query pattern)
CREATE INDEX idx_tasks_user_active_priority
ON tasks (user_id, priority_score DESC)
WHERE deleted_at IS NULL;

-- Only incomplete subtasks
CREATE INDEX idx_tasks_parent_task_id
ON tasks(parent_task_id)
WHERE parent_task_id IS NOT NULL;
```

---

## Entity Relationship Diagram

```
┌─────────────┐       ┌─────────────────┐       ┌───────────────────┐
│    users    │──────<│      tasks      │──────<│   task_history    │
│             │   1:N │                 │   1:N │                   │
│ id (PK)     │       │ id (PK)         │       │ id (PK)           │
│ email       │       │ user_id (FK)    │       │ task_id (FK)      │
│ name        │       │ parent_task_id  │──┐    │ event_type        │
│ password    │       │ series_id (FK) ─┼──┼──> │ old_value         │
│ user_type   │       │ title           │  │    │ new_value         │
└─────────────┘       │ status          │  │    │ created_at        │
      │               │ priority_score  │  │    └───────────────────┘
      │               │ ...             │  │
      │               └────────┬────────┘  │
      │                        │           │
      │                   Self-Reference   │
      │                  (subtasks)        │
      │                                    │
      │               ┌────────────────────┘
      │               │
      │         ┌─────┴─────────┐
      │         │  task_series  │
      │         │               │
      │         │ id (PK)       │
      └────────>│ user_id (FK)  │
                │ pattern       │
                │ interval      │
                │ is_active     │
                └───────────────┘

┌───────────────────────────────────────┐
│         task_dependencies             │
│                                       │
│ task_id (PK, FK) ───────> tasks.id    │
│ blocked_by_id (PK, FK) ─> tasks.id    │
│                                       │
│ (Many-to-Many relationship)           │
└───────────────────────────────────────┘
```

---

## Exercises

### 🔰 Beginner: Read the Migrations

1. Open `backend/migrations/` folder
2. Read migrations 001, 005, and 006
3. List all the indexes created and their purpose

### 🎯 Intermediate: Design a Migration

Design a migration for "task labels" feature:
- Tasks can have multiple labels
- Labels have a name and color
- Users can create/delete their own labels

What tables would you create? What indexes?

### 🚀 Advanced: Query Optimization

Given this slow query:
```sql
SELECT * FROM tasks
WHERE user_id = $1
  AND status != 'done'
  AND deleted_at IS NULL
ORDER BY priority_score DESC
LIMIT 50;
```

Design an optimal index. Explain your reasoning.

---

## Reflection Questions

1. **Why UUIDs over auto-increment?** What are the tradeoffs?

2. **Why enforce constraints in app layer vs database?** (e.g., cycle detection)

3. **When would you NOT use soft delete?** What are its costs?

4. **Why partial indexes?** When are they better than full indexes?

---

## Key Takeaways

1. **Migrations are version control for your database.** Never modify existing migrations.

2. **ENUMs provide type safety.** Invalid values fail at database level.

3. **Partial indexes are powerful.** Only index what you actually query.

4. **Soft delete enables restoration.** But adds complexity to every query.

5. **Full-text search is built-in.** PostgreSQL's tsvector + GIN beats external solutions for simple cases.

---

## Next Module

Continue to **[Module 07: Refactoring Case Studies](./07-refactoring-case-studies.md)** to see how the code evolved through refactoring.
