-- Optimize partial index for main task list queries
-- Previous index (idx_tasks_user_active) only filtered by status != 'done'
-- This new index also excludes deleted tasks and subtasks for better performance

-- Drop old partial index
DROP INDEX IF EXISTS idx_tasks_user_active;

-- Create comprehensive partial index for main task list
-- Matches the WHERE clause used in task list queries
CREATE INDEX idx_tasks_user_active_main
ON tasks(user_id, priority_score DESC, created_at DESC)
WHERE status != 'done'
  AND deleted_at IS NULL
  AND (task_type IS NULL OR task_type != 'subtask');

-- Create separate index for analytics queries (includes all statuses)
CREATE INDEX idx_tasks_user_analytics
ON tasks(user_id, bump_count)
WHERE deleted_at IS NULL;
