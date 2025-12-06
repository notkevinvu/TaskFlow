-- Rollback: Remove task templates

DROP TRIGGER IF EXISTS update_task_templates_updated_at ON task_templates;
DROP INDEX IF EXISTS idx_task_templates_user_category;
DROP INDEX IF EXISTS idx_task_templates_user_id;
DROP TABLE IF EXISTS task_templates;
