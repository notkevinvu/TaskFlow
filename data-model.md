# Data Model Specification
## Intelligent Task Prioritization System

**Version:** 1.0
**Last Updated:** 2025-01-15

---

## Overview

This document specifies the complete database schema for the Intelligent Task Prioritization System. The data model supports task management, smart prioritization, bump tracking, and analytics.

**Database:** PostgreSQL 16
**Access Pattern:** sqlc for type-safe queries

---

## Entity-Relationship Diagram

```
┌─────────────┐
│    Users    │
└──────┬──────┘
       │
       │ 1:N
       ▼
┌─────────────┐      1:N     ┌──────────────────┐
│    Tasks    │─────────────▶│   TaskHistory    │
└──────┬──────┘              └──────────────────┘
       │
       │ 1:N
       ▼
┌──────────────┐
│  Categories  │
└──────────────┘
```

---

## Table Schemas

### 1. Users Table

Stores user authentication and profile information.

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,

    -- User preferences
    default_priority INT DEFAULT 50 CHECK (default_priority BETWEEN 0 AND 100),
    time_decay_enabled BOOLEAN DEFAULT TRUE,
    bump_penalty_enabled BOOLEAN DEFAULT TRUE,

    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMPTZ -- Soft delete
);

-- Indexes
CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_created_at ON users(created_at DESC);
```

**Fields:**
- `id`: Unique identifier (UUID)
- `email`: User email (unique, used for login)
- `name`: Display name
- `password_hash`: Bcrypt hashed password
- `default_priority`: Default priority for new tasks (0-100)
- `time_decay_enabled`: User preference for time-based priority decay
- `bump_penalty_enabled`: User preference for bump penalties
- `created_at`: Account creation timestamp
- `updated_at`: Last profile update
- `deleted_at`: Soft delete timestamp (NULL = active)

---

### 2. Tasks Table

Core task data with all attributes for prioritization and tracking.

```sql
CREATE TYPE task_status AS ENUM ('todo', 'in_progress', 'done');
CREATE TYPE task_effort AS ENUM ('small', 'medium', 'large', 'xlarge');
CREATE TYPE task_source AS ENUM ('meeting', 'email', 'spontaneous', 'other');

CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Core task data
    title VARCHAR(200) NOT NULL,
    description TEXT,
    status task_status DEFAULT 'todo' NOT NULL,

    -- Priority inputs
    user_priority INT DEFAULT 50 CHECK (user_priority BETWEEN 0 AND 100),
    due_date TIMESTAMPTZ,
    estimated_effort task_effort,
    category VARCHAR(50),
    source task_source DEFAULT 'other',

    -- Context (optional)
    context TEXT, -- Where did this come from? Why is it important?
    related_people TEXT[], -- Names or emails of people involved

    -- Computed priority
    priority_score DECIMAL(5,2) DEFAULT 50.00 CHECK (priority_score BETWEEN 0 AND 100),
    priority_last_calculated_at TIMESTAMPTZ DEFAULT NOW(),

    -- Tracking
    bump_count INT DEFAULT 0,
    time_to_complete_minutes INT, -- Actual time taken (filled on completion)

    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    completed_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

-- Indexes
CREATE INDEX idx_tasks_user_id ON tasks(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_status ON tasks(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_priority_score ON tasks(priority_score DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_due_date ON tasks(due_date) WHERE due_date IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_tasks_category ON tasks(category) WHERE category IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_tasks_created_at ON tasks(created_at DESC);
CREATE INDEX idx_tasks_bump_count ON tasks(bump_count DESC) WHERE deleted_at IS NULL;

-- Composite index for main query (user's priority-sorted tasks)
CREATE INDEX idx_tasks_user_priority
ON tasks(user_id, priority_score DESC, created_at DESC)
WHERE deleted_at IS NULL AND status != 'done';

-- Full-text search on title and description
ALTER TABLE tasks ADD COLUMN search_vector tsvector;

CREATE INDEX idx_tasks_search ON tasks USING GIN(search_vector);

CREATE FUNCTION tasks_search_trigger() RETURNS trigger AS $$
BEGIN
  NEW.search_vector :=
    setweight(to_tsvector('english', coalesce(NEW.title, '')), 'A') ||
    setweight(to_tsvector('english', coalesce(NEW.description, '')), 'B') ||
    setweight(to_tsvector('english', coalesce(NEW.context, '')), 'C');
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tasks_search_update
BEFORE INSERT OR UPDATE ON tasks
FOR EACH ROW EXECUTE FUNCTION tasks_search_trigger();
```

**Fields:**
- `id`: Unique task identifier
- `user_id`: Owner of the task
- `title`: Short task description (required, max 200 chars)
- `description`: Detailed description (optional)
- `status`: Current state (todo, in_progress, done)
- `user_priority`: User-set importance (0-100, default 50)
- `due_date`: Optional deadline
- `estimated_effort`: Size estimate (small < 1h, medium 1-4h, large 4-8h, xlarge > 8h)
- `category`: User-defined category (e.g., "Code Review", "Documentation")
- `source`: Where the task originated
- `context`: Optional notes about task origin/importance
- `related_people`: Array of names/emails of people involved
- `priority_score`: Computed priority (0-100, recalculated periodically)
- `priority_last_calculated_at`: Timestamp of last priority calculation
- `bump_count`: Number of times task was delayed
- `time_to_complete_minutes`: Actual completion time (for estimation analysis)
- `created_at`: Task creation timestamp
- `updated_at`: Last modification
- `completed_at`: When marked as done
- `deleted_at`: Soft delete timestamp

**Effort Mapping:**
- `small`: < 1 hour
- `medium`: 1-4 hours
- `large`: 4-8 hours
- `xlarge`: > 8 hours

---

### 3. TaskHistory Table

Tracks all changes to tasks for analytics and audit trail.

```sql
CREATE TYPE history_event_type AS ENUM (
    'created',
    'status_changed',
    'priority_changed',
    'bumped',
    'completed',
    'updated'
);

CREATE TABLE task_history (
    id BIGSERIAL PRIMARY KEY,
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    event_type history_event_type NOT NULL,

    -- Event data
    old_value JSONB, -- Previous state (for changes)
    new_value JSONB, -- New state (for changes)

    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,

    -- Optional: why was this changed?
    reason TEXT
);

-- Indexes
CREATE INDEX idx_task_history_task_id ON task_history(task_id, created_at DESC);
CREATE INDEX idx_task_history_user_id ON task_history(user_id, created_at DESC);
CREATE INDEX idx_task_history_event_type ON task_history(event_type);
CREATE INDEX idx_task_history_created_at ON task_history(created_at DESC);
```

**Fields:**
- `id`: Unique history entry ID
- `task_id`: Reference to task
- `user_id`: User who made the change
- `event_type`: Type of change (created, status_changed, priority_changed, bumped, completed, updated)
- `old_value`: Previous state as JSON (for changes)
- `new_value`: New state as JSON (for changes)
- `created_at`: When event occurred
- `reason`: Optional explanation (e.g., "Waiting for code review")

**Example JSONB values:**

```json
// Bump event
{
  "old_value": null,
  "new_value": {
    "bump_count": 3,
    "previous_priority": 65.5,
    "new_priority": 75.5
  }
}

// Status change
{
  "old_value": {"status": "todo"},
  "new_value": {"status": "done"}
}

// Priority recalculation
{
  "old_value": {"priority_score": 50.0},
  "new_value": {
    "priority_score": 72.5,
    "factors": {
      "user_priority": 50,
      "time_decay": 40,
      "deadline_urgency": 80,
      "bump_penalty": 30
    }
  }
}
```

---

### 4. Categories Table

User-defined categories for task organization.

```sql
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    name VARCHAR(50) NOT NULL,
    color VARCHAR(7), -- Hex color code (e.g., '#FF5733')

    -- Statistics (cached for performance)
    task_count INT DEFAULT 0,
    avg_completion_time_days DECIMAL(5,2),
    avg_bump_count DECIMAL(3,2),

    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,

    UNIQUE(user_id, name)
);

-- Indexes
CREATE INDEX idx_categories_user_id ON categories(user_id);
CREATE INDEX idx_categories_name ON categories(name);
```

**Fields:**
- `id`: Category ID
- `user_id`: Owner
- `name`: Category name (unique per user)
- `color`: Display color
- `task_count`: Cached count of active tasks in this category
- `avg_completion_time_days`: Average time to complete tasks in this category
- `avg_bump_count`: Average bumps for tasks in this category
- `created_at`, `updated_at`: Timestamps

---

## Materialized Views for Analytics

### Daily Task Metrics

Pre-computed daily aggregates for faster analytics queries.

```sql
CREATE MATERIALIZED VIEW daily_task_metrics AS
SELECT
    user_id,
    DATE(created_at) as metric_date,
    COUNT(*) as tasks_created,
    COUNT(*) FILTER (WHERE status = 'done') as tasks_completed,
    COUNT(*) FILTER (WHERE bump_count >= 3) as tasks_at_risk,
    AVG(bump_count) as avg_bump_count,
    AVG(CASE
        WHEN status = 'done' AND completed_at IS NOT NULL
        THEN EXTRACT(EPOCH FROM (completed_at - created_at)) / 86400.0
    END) as avg_completion_days
FROM tasks
WHERE deleted_at IS NULL
GROUP BY user_id, DATE(created_at);

CREATE UNIQUE INDEX idx_daily_metrics_user_date
ON daily_task_metrics(user_id, metric_date DESC);

-- Refresh daily
-- Could be run as cron job or background worker
-- REFRESH MATERIALIZED VIEW daily_task_metrics;
```

---

### Category Statistics

```sql
CREATE MATERIALIZED VIEW category_stats AS
SELECT
    user_id,
    category,
    COUNT(*) as total_tasks,
    COUNT(*) FILTER (WHERE status = 'done') as completed_tasks,
    ROUND(
        (COUNT(*) FILTER (WHERE status = 'done')::DECIMAL / COUNT(*)) * 100,
        2
    ) as completion_rate,
    AVG(bump_count) as avg_bump_count,
    AVG(CASE
        WHEN status = 'done' AND completed_at IS NOT NULL
        THEN EXTRACT(EPOCH FROM (completed_at - created_at)) / 86400.0
    END) as avg_days_to_complete
FROM tasks
WHERE deleted_at IS NULL AND category IS NOT NULL
GROUP BY user_id, category;

CREATE INDEX idx_category_stats_user ON category_stats(user_id);
```

---

## Common Queries

### Get User's Priority-Sorted Tasks

```sql
SELECT
    id,
    title,
    status,
    priority_score,
    due_date,
    bump_count,
    category,
    created_at
FROM tasks
WHERE user_id = $1
  AND deleted_at IS NULL
  AND status != 'done'
ORDER BY priority_score DESC, created_at DESC
LIMIT 50;
```

---

### Get At-Risk Tasks

```sql
SELECT
    id,
    title,
    bump_count,
    priority_score,
    due_date,
    category
FROM tasks
WHERE user_id = $1
  AND deleted_at IS NULL
  AND (
    bump_count >= 3
    OR (due_date IS NOT NULL AND due_date < NOW() + INTERVAL '3 days')
  )
ORDER BY priority_score DESC;
```

---

### Delay Analysis (Most Bumped Tasks)

```sql
SELECT
    category,
    COUNT(*) as task_count,
    AVG(bump_count) as avg_bumps,
    MAX(bump_count) as max_bumps
FROM tasks
WHERE user_id = $1
  AND deleted_at IS NULL
  AND bump_count > 0
GROUP BY category
ORDER BY avg_bumps DESC;
```

---

### Estimation Accuracy

```sql
SELECT
    estimated_effort,
    COUNT(*) as task_count,
    AVG(time_to_complete_minutes) as avg_actual_minutes,
    CASE estimated_effort
        WHEN 'small' THEN 30  -- 0.5 hours
        WHEN 'medium' THEN 150 -- 2.5 hours
        WHEN 'large' THEN 360  -- 6 hours
        WHEN 'xlarge' THEN 600 -- 10 hours
    END as estimated_minutes
FROM tasks
WHERE user_id = $1
  AND status = 'done'
  AND estimated_effort IS NOT NULL
  AND time_to_complete_minutes IS NOT NULL
GROUP BY estimated_effort;
```

---

## Indexes Summary

| Index | Purpose | Impact |
|-------|---------|--------|
| `idx_tasks_user_priority` | Main task list query | High |
| `idx_tasks_priority_score` | Sorting by priority | High |
| `idx_tasks_search` | Full-text search | Medium |
| `idx_tasks_due_date` | Deadline queries | Medium |
| `idx_task_history_task_id` | History lookups | Medium |
| `idx_daily_metrics_user_date` | Analytics dashboard | High |

---

## Data Retention & Archival

### Soft Delete Strategy

All tables use `deleted_at` for soft deletes:
- Queries filter `WHERE deleted_at IS NULL`
- Allows data recovery
- Maintains referential integrity

### Hard Delete (Future)

After 90 days, permanently delete soft-deleted records:

```sql
DELETE FROM tasks
WHERE deleted_at < NOW() - INTERVAL '90 days';
```

### Archival (Future)

For completed tasks older than 1 year, move to archive table:

```sql
CREATE TABLE tasks_archive (LIKE tasks INCLUDING ALL);

INSERT INTO tasks_archive
SELECT * FROM tasks
WHERE status = 'done'
  AND completed_at < NOW() - INTERVAL '1 year';
```

---

## Migration Strategy

### Phase 1: Core Tables

1. Create `users` table
2. Create `tasks` table with basic fields
3. Create `task_history` table

### Phase 2: Analytics

4. Add materialized views
5. Create `categories` table
6. Add full-text search

### Phase 3: Optimizations

7. Add composite indexes
8. Add partitioning for `task_history` (if needed for performance)

---

## Database Size Estimates

**Assumptions:**
- 1,000 users
- Average 500 tasks per user
- 10 history events per task

| Table | Rows | Size per Row | Total Size |
|-------|------|--------------|------------|
| **users** | 1,000 | ~500 bytes | ~500 KB |
| **tasks** | 500,000 | ~1 KB | ~500 MB |
| **task_history** | 5,000,000 | ~300 bytes | ~1.5 GB |
| **categories** | 10,000 | ~200 bytes | ~2 MB |
| **Total** | | | **~2 GB** |

With indexes: **~3-4 GB total**

---

## Performance Considerations

### Query Optimization

1. **Use composite indexes** for multi-column WHERE + ORDER BY
2. **Limit result sets** with LIMIT (pagination)
3. **Materialized views** for expensive analytics
4. **Connection pooling** (pgxpool with 25-50 connections)

### Write Performance

1. **Batch inserts** for history events
2. **Async priority recalculation** (background job)
3. **Debounce updates** to avoid excessive history entries

---

## Schema Evolution

### Adding New Fields

Use `ALTER TABLE ADD COLUMN` with defaults:

```sql
ALTER TABLE tasks
ADD COLUMN tags TEXT[] DEFAULT '{}';
```

### Changing Types

Create new column, migrate data, drop old:

```sql
ALTER TABLE tasks ADD COLUMN new_priority INT;
UPDATE tasks SET new_priority = CAST(priority_score AS INT);
ALTER TABLE tasks DROP COLUMN priority_score;
ALTER TABLE tasks RENAME COLUMN new_priority TO priority_score;
```

---

## Related Documents

- `PRD.md` - Product requirements
- `priority-algorithm.md` - Priority calculation logic
- `phase-2-weeks-3-4.md` - Implementation guide

---

**Document Status:** Ready for Implementation
**Last Review:** 2025-01-15
