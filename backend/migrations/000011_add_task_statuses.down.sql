-- Down migration for 000011_add_task_statuses
-- WARNING: PostgreSQL does not support removing values from enums
--
-- To roll back this migration:
-- 1. Update all tasks with on_hold/blocked status to 'todo'
-- 2. Recreate the enum type without these values
-- 3. This requires dropping/recreating all columns that use the enum
--
-- Manual steps required:
-- UPDATE tasks SET status = 'todo' WHERE status IN ('on_hold', 'blocked');
-- Then recreate the enum (complex operation requiring table recreation)

-- Placeholder: No automatic rollback available for enum value additions
SELECT 1;
