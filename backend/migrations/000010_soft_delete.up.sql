-- Migration 010: Add soft delete support for tasks
-- This enables undo functionality for task deletion

-- Add deleted_at column for soft delete
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ DEFAULT NULL;

-- Create index for efficient filtering of non-deleted tasks
-- Partial index only includes non-deleted tasks for better performance
CREATE INDEX IF NOT EXISTS idx_tasks_not_deleted ON tasks (user_id, status)
WHERE deleted_at IS NULL;

-- Add comment explaining soft delete behavior
COMMENT ON COLUMN tasks.deleted_at IS 'Soft delete timestamp - NULL means not deleted, set value means deleted at that time';
