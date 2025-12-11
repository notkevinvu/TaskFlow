-- Revert to original index
DROP INDEX IF EXISTS idx_tasks_user_active_main;
DROP INDEX IF EXISTS idx_tasks_user_analytics;

-- Restore original index
CREATE INDEX IF NOT EXISTS idx_tasks_user_active
ON tasks(user_id, bump_count)
WHERE status != 'done';
