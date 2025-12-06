-- Migration: Add subtask support with task_type discrimination
-- This enables parent-child task relationships (single-level nesting only)

-- Create task_type enum to distinguish task relationship types
CREATE TYPE task_type AS ENUM ('regular', 'recurring', 'subtask');

-- Add task_type column to tasks table
ALTER TABLE tasks ADD COLUMN task_type task_type;

-- Backfill existing data based on series_id presence
-- Tasks with series_id are recurring, all others are regular
UPDATE tasks SET task_type =
    CASE
        WHEN series_id IS NOT NULL THEN 'recurring'::task_type
        ELSE 'regular'::task_type
    END;

-- Make task_type NOT NULL with default after backfill
ALTER TABLE tasks ALTER COLUMN task_type SET NOT NULL;
ALTER TABLE tasks ALTER COLUMN task_type SET DEFAULT 'regular';

-- Create composite index for efficient subtask queries
-- This index optimizes: "find all subtasks for parent X"
CREATE INDEX idx_tasks_subtask_lookup ON tasks(parent_task_id, task_type)
    WHERE task_type = 'subtask';

-- Create function to count incomplete subtasks for a parent task
-- Used for blocking parent completion and showing progress
CREATE OR REPLACE FUNCTION count_incomplete_subtasks(parent_id UUID)
RETURNS INTEGER AS $$
    SELECT COUNT(*)::INTEGER
    FROM tasks
    WHERE parent_task_id = $1
        AND task_type = 'subtask'
        AND status != 'done';
$$ LANGUAGE SQL STABLE;

-- Add comment documenting the dual-use of parent_task_id
COMMENT ON COLUMN tasks.parent_task_id IS
    'Parent task reference. Interpretation depends on task_type:
     - task_type=recurring: links to previous occurrence in series (temporal chain)
     - task_type=subtask: links to parent task (hierarchical relationship)
     - task_type=regular: should be NULL';

COMMENT ON COLUMN tasks.task_type IS
    'Discriminator for task relationships:
     - regular: standalone task (default)
     - recurring: part of a recurring series (has series_id)
     - subtask: child of another task (has parent_task_id, no series_id)';
