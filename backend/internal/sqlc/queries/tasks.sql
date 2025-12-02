-- Task queries for sqlc code generation

-- name: CreateTask :exec
INSERT INTO tasks (
    id, user_id, title, description, status, user_priority,
    due_date, estimated_effort, category, context, related_people,
    priority_score, bump_count, created_at, updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15);

-- name: GetTaskByID :one
SELECT id, user_id, title, description, status, user_priority,
       due_date, estimated_effort, category, context, related_people,
       priority_score, bump_count, created_at, updated_at, completed_at
FROM tasks
WHERE id = $1;

-- name: UpdateTask :exec
UPDATE tasks
SET title = $1, description = $2, status = $3, user_priority = $4,
    due_date = $5, estimated_effort = $6, category = $7, context = $8,
    related_people = $9, priority_score = $10, bump_count = $11,
    updated_at = $12, completed_at = $13
WHERE id = $14 AND user_id = $15;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE id = $1 AND user_id = $2;

-- name: IncrementBumpCount :exec
UPDATE tasks
SET bump_count = bump_count + 1, updated_at = NOW()
WHERE id = $1 AND user_id = $2;

-- name: GetAtRiskTasks :many
SELECT id, user_id, title, description, status, user_priority,
       due_date, estimated_effort, category, context, related_people,
       priority_score, bump_count, created_at, updated_at, completed_at
FROM tasks
WHERE user_id = $1
  AND (
      bump_count >= 3
      OR (due_date IS NOT NULL AND due_date < NOW() - INTERVAL '3 days')
  )
  AND status != 'done'
ORDER BY priority_score DESC;

-- name: GetCategories :many
SELECT DISTINCT category
FROM tasks
WHERE user_id = $1 AND category IS NOT NULL
ORDER BY category;

-- name: RenameCategoryForUser :exec
UPDATE tasks
SET category = $1, updated_at = NOW()
WHERE user_id = $2 AND category = $3 AND status != 'done';

-- name: DeleteCategoryForUser :exec
UPDATE tasks
SET category = NULL, updated_at = NOW()
WHERE user_id = $1 AND category = $2 AND status != 'done';

-- Analytics queries

-- name: GetCompletionStats :one
SELECT
    COUNT(*) as total_tasks,
    COUNT(*) FILTER (WHERE status = 'done') as completed_tasks,
    COUNT(*) FILTER (WHERE status != 'done') as pending_tasks
FROM tasks
WHERE user_id = $1
  AND created_at >= NOW() - INTERVAL '1 day' * $2;

-- name: GetAverageBumpCount :one
SELECT COALESCE(AVG(bump_count), 0) as avg_bump_count
FROM tasks
WHERE user_id = $1 AND status != 'done';

-- name: GetTasksByBumpCount :many
SELECT bump_count, COUNT(*) as task_count
FROM tasks
WHERE user_id = $1 AND status != 'done'
GROUP BY bump_count
ORDER BY bump_count;

-- name: GetCategoryBreakdown :many
SELECT
    COALESCE(category, 'Uncategorized') as category,
    COUNT(*) as total_count,
    COUNT(*) FILTER (WHERE status = 'done') as completed_count
FROM tasks
WHERE user_id = $1
  AND created_at >= NOW() - INTERVAL '1 day' * $2
GROUP BY category
ORDER BY total_count DESC;

-- name: GetVelocityMetrics :many
SELECT
    DATE(completed_at) as completion_date,
    COUNT(*) as completed_count
FROM tasks
WHERE user_id = $1
  AND status = 'done'
  AND completed_at >= NOW() - INTERVAL '1 day' * $2
GROUP BY DATE(completed_at)
ORDER BY completion_date;

-- name: GetPriorityDistribution :many
SELECT
    CASE
        WHEN priority_score >= 90 THEN 'Critical (90-100)'
        WHEN priority_score >= 75 THEN 'High (75-89)'
        WHEN priority_score >= 50 THEN 'Medium (50-74)'
        ELSE 'Low (0-49)'
    END as priority_range,
    COUNT(*) as task_count
FROM tasks
WHERE user_id = $1 AND status != 'done'
GROUP BY
    CASE
        WHEN priority_score >= 90 THEN 'Critical (90-100)'
        WHEN priority_score >= 75 THEN 'High (75-89)'
        WHEN priority_score >= 50 THEN 'Medium (50-74)'
        ELSE 'Low (0-49)'
    END
ORDER BY
    MIN(
        CASE
            WHEN priority_score >= 90 THEN 1
            WHEN priority_score >= 75 THEN 2
            WHEN priority_score >= 50 THEN 3
            ELSE 4
        END
    );

-- Insights analytics queries

-- name: GetCategoryBumpStats :many
-- Gets average bump count per category for detecting avoidance patterns
SELECT
    COALESCE(category, 'Uncategorized') as category,
    AVG(bump_count)::float as avg_bumps,
    COUNT(*)::int as task_count
FROM tasks
WHERE user_id = $1 AND status != 'done'
GROUP BY category
HAVING AVG(bump_count) > 1
ORDER BY avg_bumps DESC;

-- name: GetCompletionByDayOfWeek :many
-- Gets task completions grouped by day of week for peak performance detection
SELECT
    EXTRACT(DOW FROM completed_at)::int as day_of_week,
    COUNT(*)::int as completed_count
FROM tasks
WHERE user_id = $1
  AND status = 'done'
  AND completed_at >= NOW() - INTERVAL '1 day' * $2
GROUP BY day_of_week
ORDER BY completed_count DESC;

-- name: GetAgingQuickWins :many
-- Gets small effort tasks that are aging (quick wins)
SELECT id, user_id, title, description, status, user_priority,
       due_date, estimated_effort, category, context, related_people,
       priority_score, bump_count, created_at, updated_at, completed_at
FROM tasks
WHERE user_id = $1
  AND estimated_effort = 'small'
  AND status != 'done'
  AND created_at <= NOW() - INTERVAL '1 day' * $2
ORDER BY created_at ASC
LIMIT $3;

-- name: GetDeadlineClusters :many
-- Gets dates with multiple tasks due (deadline clustering)
SELECT
    DATE(due_date) as due_date,
    COUNT(*)::int as task_count,
    array_agg(title) as titles
FROM tasks
WHERE user_id = $1
  AND status != 'done'
  AND due_date IS NOT NULL
  AND due_date >= NOW()
  AND due_date <= NOW() + INTERVAL '1 day' * $2
GROUP BY DATE(due_date)
HAVING COUNT(*) >= 3
ORDER BY due_date;

-- name: GetCompletionTimeStats :one
-- Gets completion time statistics for time estimation
SELECT
    COUNT(*)::int as sample_size,
    COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY
        EXTRACT(EPOCH FROM (completed_at - created_at)) / 86400
    ), 0)::float as median_days,
    COALESCE(AVG(EXTRACT(EPOCH FROM (completed_at - created_at)) / 86400), 0)::float as avg_days
FROM tasks
WHERE user_id = $1
  AND status = 'done'
  AND completed_at IS NOT NULL
  AND ($2::text IS NULL OR category = $2)
  AND ($3::task_effort IS NULL OR estimated_effort = $3);

-- name: GetCategoryDistribution :many
-- Gets distribution of pending tasks by category for overload detection
SELECT
    COALESCE(category, 'Uncategorized') as category,
    COUNT(*)::int as task_count,
    (COUNT(*)::float / NULLIF(SUM(COUNT(*)) OVER (), 0) * 100)::float as percentage
FROM tasks
WHERE user_id = $1 AND status != 'done'
GROUP BY category
ORDER BY task_count DESC;
