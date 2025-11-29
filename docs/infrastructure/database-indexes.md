# Database Indexes for TaskFlow

## Overview

TaskFlow uses strategic database indexing to optimize query performance for search, filtering, and sorting operations.

## Existing Indexes (Migration 000001)

### Single-Column Indexes

**Users Table:**
- `idx_users_email` - Email lookups for authentication

**Tasks Table:**
- `idx_tasks_user_id` - Filter tasks by user
- `idx_tasks_status` - Filter by task status (todo, in_progress, done)
- `idx_tasks_priority_score` - Sort by priority (DESC)
- `idx_tasks_due_date` - Filter/sort by due date
- `idx_tasks_category` - Filter by category
- `idx_tasks_created_at` - Sort by creation date

**Full-Text Search:**
- `idx_tasks_search_vector` (GIN index) - Fast full-text search on title, description, context

**Task History Table:**
- `idx_task_history_task_id` - Look up history for a task
- `idx_task_history_user_id` - Look up history by user
- `idx_task_history_event_type` - Filter by event type
- `idx_task_history_created_at` - Sort by timestamp (DESC)

## Composite Indexes (Migration 000003)

Composite indexes optimize queries that filter on multiple columns. PostgreSQL can use these for queries that match the leftmost prefix of the index.

###1. User + Status + Priority
```sql
idx_tasks_user_status_priority ON tasks(user_id, status, priority_score DESC)
```

**Optimizes:**
- Get all "todo" tasks for a user, sorted by priority
- Get all "in_progress" tasks for a user
- Dashboard views filtering by status

**Example Query:**
```sql
SELECT * FROM tasks
WHERE user_id = '...' AND status = 'todo'
ORDER BY priority_score DESC;
```

### 2. User + Category + Priority
```sql
idx_tasks_user_category_priority ON tasks(user_id, category, priority_score DESC)
WHERE category IS NOT NULL
```

**Optimizes:**
- Get all tasks in a category for a user, sorted by priority
- Category-filtered views
- Partial index (excludes NULL categories) for smaller index size

**Example Query:**
```sql
SELECT * FROM tasks
WHERE user_id = '...' AND category = 'Work'
ORDER BY priority_score DESC;
```

### 3. User + Status + Due Date
```sql
idx_tasks_user_status_due_date ON tasks(user_id, status, due_date)
```

**Optimizes:**
- Upcoming incomplete tasks
- Calendar views with status filtering
- Due date range queries

**Example Query:**
```sql
SELECT * FROM tasks
WHERE user_id = '...' AND status != 'done'
  AND due_date BETWEEN NOW() AND NOW() + INTERVAL '7 days';
```

### 4. User + Active Tasks (Partial Index)
```sql
idx_tasks_user_active ON tasks(user_id, status, bump_count)
WHERE status != 'done'
```

**Optimizes:**
- Analytics queries on active tasks
- Bump count aggregations
- At-risk task detection
- Partial index (excludes completed tasks) for smaller size

**Example Query:**
```sql
SELECT AVG(bump_count) FROM tasks
WHERE user_id = '...' AND status != 'done';
```

## Index Selection Strategy

PostgreSQL automatically chooses the most efficient index based on query patterns. Our strategy:

1. **Single-column indexes** for simple filters and sorts
2. **Composite indexes** for common multi-column filter patterns
3. **Partial indexes** (with WHERE clause) for frequently filtered subsets
4. **GIN indexes** for full-text search
5. **DESC indexes** for reverse-order sorts (priority)

## Testing Index Performance

### 1. Check Query Plan

Use `EXPLAIN ANALYZE` to see which indexes are being used:

```sql
EXPLAIN ANALYZE
SELECT * FROM tasks
WHERE user_id = '...' AND status = 'todo'
ORDER BY priority_score DESC
LIMIT 20;
```

**What to look for:**
- `Index Scan` or `Bitmap Index Scan` (good - using index)
- `Seq Scan` (bad - full table scan, might need index)
- `Rows Removed by Filter` (bad - index not selective enough)
- `Execution Time` (lower is better)

### 2. List All Indexes

```sql
SELECT
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE schemaname = 'public'
  AND tablename = 'tasks'
ORDER BY indexname;
```

### 3. Check Index Usage Statistics

```sql
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan as index_scans,
    idx_tup_read as tuples_read,
    idx_tup_fetch as tuples_fetched
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
  AND tablename = 'tasks'
ORDER BY idx_scan DESC;
```

**What to look for:**
- High `idx_scan` values indicate frequently used indexes
- Zero `idx_scan` might indicate an unused index (consider removing)

### 4. Check Index Bloat

```sql
SELECT
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

## Performance Benchmarks

### Before Composite Indexes (Baseline)
```sql
EXPLAIN ANALYZE
SELECT * FROM tasks
WHERE user_id = '...' AND status = 'todo'
ORDER BY priority_score DESC
LIMIT 20;
```
Expected: Sequential scan or bitmap heap scan combining multiple indexes

### After Composite Indexes
Same query should use `idx_tasks_user_status_priority` directly with lower execution time.

## Maintenance

### Automatic Maintenance
PostgreSQL autovacuum handles index maintenance automatically.

### Manual Reindex (Rare)
If you suspect index corruption or bloat:

```sql
REINDEX INDEX idx_tasks_user_status_priority;
-- or reindex all indexes on a table:
REINDEX TABLE tasks;
```

### Monitoring Index Size
```sql
SELECT
    indexname,
    pg_size_pretty(pg_relation_size(indexname::regclass)) as size
FROM pg_indexes
WHERE tablename = 'tasks';
```

## Migration Notes

- **Migration 000001**: Created base indexes (single-column + full-text)
- **Migration 000002**: Updated user priority scale (no index changes)
- **Migration 000003**: Added composite indexes for query optimization

## Future Optimizations

Consider adding if query patterns change:

1. **Composite index on user_id + search_vector** - If full-text search is frequently scoped to a user
2. **Index on estimated_effort** - If filtering/sorting by effort becomes common
3. **GIN index on related_people array** - If querying by related people is added
4. **Materialized views** - For complex analytics dashboards if they become slow

## Best Practices

1. **Don't over-index** - Each index has storage and write overhead
2. **Monitor index usage** - Remove unused indexes
3. **Test with production data volumes** - Index benefits vary with table size
4. **Use EXPLAIN ANALYZE** - Verify assumptions about query performance
5. **Update statistics** - Run `ANALYZE tasks;` after bulk data changes
