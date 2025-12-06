-- Rollback: Remove subtask support

-- Remove comments
COMMENT ON COLUMN tasks.parent_task_id IS NULL;
COMMENT ON COLUMN tasks.task_type IS NULL;

-- Drop function
DROP FUNCTION IF EXISTS count_incomplete_subtasks(UUID);

-- Drop index
DROP INDEX IF EXISTS idx_tasks_subtask_lookup;

-- Remove task_type column
ALTER TABLE tasks DROP COLUMN IF EXISTS task_type;

-- Drop enum type
DROP TYPE IF EXISTS task_type;
