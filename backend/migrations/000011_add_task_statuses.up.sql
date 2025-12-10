-- Migration 011: Add on_hold and blocked task statuses
-- These provide JIRA-like workflow states for better task management

-- Add new status values to the task_status enum
-- Note: PostgreSQL allows adding values to enums, but not removing them
ALTER TYPE task_status ADD VALUE IF NOT EXISTS 'on_hold';
ALTER TYPE task_status ADD VALUE IF NOT EXISTS 'blocked';

-- Add comment explaining the new statuses
COMMENT ON TYPE task_status IS 'Task status: todo, in_progress, done, on_hold (paused), blocked (manually marked as blocked)';
