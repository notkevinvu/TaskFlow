-- Add composite indexes for improved query performance on filtered lists

-- Composite index for user + status queries (common pattern: get all todos for a user)
-- This helps queries like: WHERE user_id = ? AND status = ? ORDER BY priority_score DESC
CREATE INDEX IF NOT EXISTS idx_tasks_user_status_priority
ON tasks(user_id, status, priority_score DESC);

-- Composite index for user + category queries (common pattern: get tasks by category)
-- This helps queries like: WHERE user_id = ? AND category = ? ORDER BY priority_score DESC
CREATE INDEX IF NOT EXISTS idx_tasks_user_category_priority
ON tasks(user_id, category, priority_score DESC)
WHERE category IS NOT NULL;

-- Composite index for user + status + due_date (common pattern: upcoming incomplete tasks)
-- This helps queries like: WHERE user_id = ? AND status != 'done' AND due_date BETWEEN ? AND ?
CREATE INDEX IF NOT EXISTS idx_tasks_user_status_due_date
ON tasks(user_id, status, due_date);

-- Composite index for analytics queries on non-done tasks
-- This helps queries like: WHERE user_id = ? AND status != 'done' for aggregations
CREATE INDEX IF NOT EXISTS idx_tasks_user_active
ON tasks(user_id, status, bump_count)
WHERE status != 'done';
