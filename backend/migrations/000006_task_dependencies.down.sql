-- Rollback: Remove task dependencies support

DROP FUNCTION IF EXISTS get_blocker_ids(UUID);
DROP FUNCTION IF EXISTS count_incomplete_blockers(UUID);
DROP TABLE IF EXISTS task_dependencies;
