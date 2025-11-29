-- Remove composite indexes

DROP INDEX IF EXISTS idx_tasks_user_active;
DROP INDEX IF EXISTS idx_tasks_user_status_due_date;
DROP INDEX IF EXISTS idx_tasks_user_category_priority;
DROP INDEX IF EXISTS idx_tasks_user_status_priority;
