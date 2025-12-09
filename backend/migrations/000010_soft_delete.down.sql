-- Rollback Migration 010: Remove soft delete support
-- WARNING: This will permanently delete all soft-deleted tasks

-- Remove the index first
DROP INDEX IF EXISTS idx_tasks_not_deleted;

-- Remove the deleted_at column (this will lose soft-deleted tasks!)
ALTER TABLE tasks DROP COLUMN IF EXISTS deleted_at;
