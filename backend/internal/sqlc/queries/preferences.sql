-- name: GetUserPreferences :one
SELECT user_id, default_due_date_calculation, created_at, updated_at
FROM user_preferences
WHERE user_id = $1;

-- name: UpsertUserPreferences :exec
INSERT INTO user_preferences (user_id, default_due_date_calculation, created_at, updated_at)
VALUES ($1, $2, NOW(), NOW())
ON CONFLICT (user_id) DO UPDATE
SET default_due_date_calculation = EXCLUDED.default_due_date_calculation,
    updated_at = NOW();

-- name: DeleteUserPreferences :exec
DELETE FROM user_preferences WHERE user_id = $1;

-- name: GetCategoryPreference :one
SELECT id, user_id, category, due_date_calculation, created_at, updated_at
FROM category_preferences
WHERE user_id = $1 AND category = $2;

-- name: GetCategoryPreferencesByUserID :many
SELECT id, user_id, category, due_date_calculation, created_at, updated_at
FROM category_preferences
WHERE user_id = $1
ORDER BY category ASC;

-- name: UpsertCategoryPreference :exec
INSERT INTO category_preferences (id, user_id, category, due_date_calculation, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
ON CONFLICT (user_id, category) DO UPDATE
SET due_date_calculation = EXCLUDED.due_date_calculation,
    updated_at = NOW();

-- name: DeleteCategoryPreference :exec
DELETE FROM category_preferences
WHERE user_id = $1 AND category = $2;

-- name: DeleteAllCategoryPreferences :exec
DELETE FROM category_preferences WHERE user_id = $1;
