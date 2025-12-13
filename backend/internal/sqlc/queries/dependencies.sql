-- Dependency queries for sqlc code generation

-- name: VerifyTasksExistForUser :one
-- Verify both tasks exist and belong to the user (returns count, should be 2)
SELECT COUNT(*)::int FROM tasks
WHERE id IN ($1, $2) AND user_id = $3;

-- name: AddDependency :exec
-- Add a new dependency relationship
INSERT INTO task_dependencies (task_id, blocked_by_id, created_at)
VALUES ($1, $2, $3);

-- name: RemoveDependency :execrows
-- Remove a dependency (verifies user owns the task)
DELETE FROM task_dependencies td
USING tasks t
WHERE td.task_id = $1
  AND td.blocked_by_id = $2
  AND t.id = td.task_id
  AND t.user_id = $3;

-- name: CheckDependencyExists :one
-- Check if a specific dependency exists
SELECT EXISTS(
    SELECT 1 FROM task_dependencies
    WHERE task_id = $1 AND blocked_by_id = $2
)::bool;

-- name: GetBlockerTasks :many
-- Get all tasks that block the given task (with task details)
SELECT t.id, t.title, t.status, td.created_at
FROM task_dependencies td
INNER JOIN tasks t ON t.id = td.blocked_by_id
WHERE td.task_id = $1
ORDER BY td.created_at ASC;

-- name: GetBlockingTasks :many
-- Get all tasks that are blocked by the given task (with task details)
SELECT t.id, t.title, t.status, td.created_at
FROM task_dependencies td
INNER JOIN tasks t ON t.id = td.task_id
WHERE td.blocked_by_id = $1
ORDER BY td.created_at ASC;

-- name: GetBlockerIDs :many
-- Get just the blocker task IDs (for cycle detection)
SELECT blocked_by_id FROM task_dependencies
WHERE task_id = $1;

-- name: GetDependencyGraph :many
-- Get all dependency relationships for a user's tasks
SELECT td.task_id, td.blocked_by_id
FROM task_dependencies td
INNER JOIN tasks t ON t.id = td.task_id
WHERE t.user_id = $1;

-- name: CountIncompleteBlockers :one
-- Count incomplete blockers for a single task
SELECT COUNT(*)::int
FROM task_dependencies td
INNER JOIN tasks t ON t.id = td.blocked_by_id
WHERE td.task_id = $1
  AND t.status != 'done';

-- name: GetTasksBlockedByTask :many
-- Get task IDs that are blocked by the given task
SELECT task_id FROM task_dependencies
WHERE blocked_by_id = $1;

-- name: CountIncompleteBlockersBatch :many
-- Batch query: count incomplete blockers for multiple tasks at once
-- Key optimization from PR #77 - eliminates N+1 query pattern
SELECT td.task_id, COUNT(*)::int as incomplete_count
FROM task_dependencies td
INNER JOIN tasks t ON t.id = td.blocked_by_id
WHERE td.task_id = ANY($1::uuid[])
  AND t.status != 'done'
GROUP BY td.task_id;
