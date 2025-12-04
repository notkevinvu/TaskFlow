-- name: CreateTaskSeries :exec
INSERT INTO task_series (
    id, user_id, original_task_id, pattern, interval_value,
    end_date, due_date_calculation, is_active, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10);

-- name: GetTaskSeriesByID :one
SELECT id, user_id, original_task_id, pattern, interval_value,
       end_date, due_date_calculation, is_active, created_at, updated_at
FROM task_series
WHERE id = $1;

-- name: GetTaskSeriesByUserID :many
SELECT id, user_id, original_task_id, pattern, interval_value,
       end_date, due_date_calculation, is_active, created_at, updated_at
FROM task_series
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetActiveTaskSeriesByUserID :many
SELECT id, user_id, original_task_id, pattern, interval_value,
       end_date, due_date_calculation, is_active, created_at, updated_at
FROM task_series
WHERE user_id = $1 AND is_active = true
ORDER BY created_at DESC;

-- name: UpdateTaskSeries :exec
UPDATE task_series
SET pattern = COALESCE($3, pattern),
    interval_value = COALESCE($4, interval_value),
    end_date = $5,
    due_date_calculation = COALESCE($6, due_date_calculation),
    is_active = COALESCE($7, is_active),
    updated_at = NOW()
WHERE id = $1 AND user_id = $2;

-- name: DeactivateTaskSeries :exec
UPDATE task_series
SET is_active = false, updated_at = NOW()
WHERE id = $1 AND user_id = $2;

-- name: DeleteTaskSeries :exec
DELETE FROM task_series
WHERE id = $1 AND user_id = $2;

-- name: GetTasksBySeriesID :many
SELECT id, user_id, title, description, status, user_priority,
       due_date, estimated_effort, category, context, related_people,
       priority_score, bump_count, created_at, updated_at, completed_at,
       series_id, parent_task_id
FROM tasks
WHERE series_id = $1
ORDER BY created_at ASC;

-- name: GetTaskSeriesWithLatestTask :one
SELECT
    ts.id, ts.user_id, ts.original_task_id, ts.pattern, ts.interval_value,
    ts.end_date, ts.due_date_calculation, ts.is_active, ts.created_at, ts.updated_at,
    t.id AS latest_task_id, t.title AS latest_task_title, t.status AS latest_task_status,
    t.due_date AS latest_task_due_date
FROM task_series ts
LEFT JOIN tasks t ON t.series_id = ts.id
WHERE ts.id = $1
ORDER BY t.created_at DESC
LIMIT 1;

-- name: CountTasksInSeries :one
SELECT COUNT(*) FROM tasks WHERE series_id = $1;
