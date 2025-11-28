-- Task History queries for sqlc code generation

-- name: CreateTaskHistory :exec
INSERT INTO task_history (id, user_id, task_id, event_type, old_value, new_value, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetTaskHistoryByTaskID :many
SELECT id, user_id, task_id, event_type, old_value, new_value, created_at
FROM task_history
WHERE task_id = $1
ORDER BY created_at DESC;
