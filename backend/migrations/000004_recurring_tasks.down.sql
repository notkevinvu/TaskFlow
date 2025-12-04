-- Remove foreign key constraint first
ALTER TABLE task_series DROP CONSTRAINT IF EXISTS fk_task_series_original_task;

-- Remove triggers
DROP TRIGGER IF EXISTS update_category_preferences_updated_at ON category_preferences;
DROP TRIGGER IF EXISTS update_user_preferences_updated_at ON user_preferences;
DROP TRIGGER IF EXISTS update_task_series_updated_at ON task_series;

-- Remove indexes from tasks
DROP INDEX IF EXISTS idx_tasks_parent_task_id;
DROP INDEX IF EXISTS idx_tasks_series_id;

-- Remove columns from tasks
ALTER TABLE tasks DROP COLUMN IF EXISTS parent_task_id;
ALTER TABLE tasks DROP COLUMN IF EXISTS series_id;

-- Drop category_preferences table
DROP TABLE IF EXISTS category_preferences;

-- Drop user_preferences table
DROP TABLE IF EXISTS user_preferences;

-- Drop task_series table and indexes
DROP TABLE IF EXISTS task_series;

-- Drop enums
DROP TYPE IF EXISTS due_date_calculation;
DROP TYPE IF EXISTS recurrence_pattern;
