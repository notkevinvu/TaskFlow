-- Migration: Add task dependencies (blocked-by relationships)
-- Enables peer-to-peer task blocking with cycle prevention

-- Create junction table for task dependencies
-- A row means: task_id is blocked by blocked_by_id
CREATE TABLE task_dependencies (
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    blocked_by_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (task_id, blocked_by_id),
    -- Prevent self-dependencies at DB level
    CONSTRAINT no_self_dependency CHECK (task_id != blocked_by_id)
);

-- Index for efficient "what blocks this task?" queries
CREATE INDEX idx_task_dependencies_task ON task_dependencies(task_id);

-- Index for efficient "what does this task block?" queries
CREATE INDEX idx_task_dependencies_blocker ON task_dependencies(blocked_by_id);

-- Function to count incomplete blockers for a task
-- Used for completion validation
CREATE OR REPLACE FUNCTION count_incomplete_blockers(p_task_id UUID)
RETURNS INTEGER AS $$
    SELECT COUNT(*)::INTEGER
    FROM task_dependencies td
    INNER JOIN tasks t ON t.id = td.blocked_by_id
    WHERE td.task_id = p_task_id
      AND t.status != 'done';
$$ LANGUAGE SQL STABLE;

-- Function to get all blocker task IDs for cycle detection
-- Returns array of task IDs that block the given task (directly)
CREATE OR REPLACE FUNCTION get_blocker_ids(p_task_id UUID)
RETURNS UUID[] AS $$
    SELECT COALESCE(array_agg(blocked_by_id), ARRAY[]::UUID[])
    FROM task_dependencies
    WHERE task_id = p_task_id;
$$ LANGUAGE SQL STABLE;

COMMENT ON TABLE task_dependencies IS
    'Junction table for blocked-by task relationships.
     task_id is blocked by blocked_by_id.
     Only regular tasks can have/be dependencies (enforced at application layer).
     Cycles are prevented at application layer via graph traversal.';

COMMENT ON COLUMN task_dependencies.task_id IS
    'The task that is blocked (cannot be completed until blocker is done)';

COMMENT ON COLUMN task_dependencies.blocked_by_id IS
    'The task that is blocking (must be completed first)';
