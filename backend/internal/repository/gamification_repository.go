package repository

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
)

// GamificationRepository handles gamification data access
type GamificationRepository struct {
	pool *pgxpool.Pool
}

// NewGamificationRepository creates a new gamification repository
func NewGamificationRepository(pool *pgxpool.Pool) *GamificationRepository {
	return &GamificationRepository{pool: pool}
}

// GetStats retrieves gamification stats for a user
func (r *GamificationRepository) GetStats(ctx context.Context, userID string) (*domain.GamificationStats, error) {
	query := `
		SELECT user_id, current_streak, longest_streak, last_completion_date,
			   total_completed, productivity_score, completion_rate, streak_score,
			   on_time_percentage, effort_mix_score, last_computed_at, created_at, updated_at
		FROM gamification_stats
		WHERE user_id = $1
	`

	var stats domain.GamificationStats
	var lastCompletionDate *time.Time

	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&stats.UserID,
		&stats.CurrentStreak,
		&stats.LongestStreak,
		&lastCompletionDate,
		&stats.TotalCompleted,
		&stats.ProductivityScore,
		&stats.CompletionRate,
		&stats.StreakScore,
		&stats.OnTimePercentage,
		&stats.EffortMixScore,
		&stats.LastComputedAt,
		&stats.CreatedAt,
		&stats.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrGamificationStatsNotFound
		}
		return nil, err
	}

	if lastCompletionDate != nil {
		formatted := lastCompletionDate.Format("2006-01-02")
		stats.LastCompletionDate = &formatted
	}

	return &stats, nil
}

// UpsertStats updates or creates gamification stats
func (r *GamificationRepository) UpsertStats(ctx context.Context, stats *domain.GamificationStats) error {
	var lastCompletionDate *time.Time
	if stats.LastCompletionDate != nil {
		t, err := time.Parse("2006-01-02", *stats.LastCompletionDate)
		if err != nil {
			slog.Warn("Failed to parse last_completion_date, storing as null",
				"value", *stats.LastCompletionDate, "error", err)
		} else {
			lastCompletionDate = &t
		}
	}

	query := `
		INSERT INTO gamification_stats (
			user_id, current_streak, longest_streak, last_completion_date,
			total_completed, productivity_score, completion_rate, streak_score,
			on_time_percentage, effort_mix_score, last_computed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (user_id) DO UPDATE SET
			current_streak = EXCLUDED.current_streak,
			longest_streak = EXCLUDED.longest_streak,
			last_completion_date = EXCLUDED.last_completion_date,
			total_completed = EXCLUDED.total_completed,
			productivity_score = EXCLUDED.productivity_score,
			completion_rate = EXCLUDED.completion_rate,
			streak_score = EXCLUDED.streak_score,
			on_time_percentage = EXCLUDED.on_time_percentage,
			effort_mix_score = EXCLUDED.effort_mix_score,
			last_computed_at = EXCLUDED.last_computed_at,
			updated_at = NOW()
	`

	_, err := r.pool.Exec(ctx, query,
		stats.UserID,
		stats.CurrentStreak,
		stats.LongestStreak,
		lastCompletionDate,
		stats.TotalCompleted,
		stats.ProductivityScore,
		stats.CompletionRate,
		stats.StreakScore,
		stats.OnTimePercentage,
		stats.EffortMixScore,
		time.Now(),
	)

	return err
}

// CreateAchievement stores a new achievement (idempotent via UNIQUE constraint)
func (r *GamificationRepository) CreateAchievement(ctx context.Context, achievement *domain.UserAchievement) error {
	var metadataJSON []byte
	var err error

	if achievement.Metadata != nil {
		metadataJSON, err = json.Marshal(achievement.Metadata)
		if err != nil {
			return err
		}
	}

	if achievement.ID == "" {
		achievement.ID = uuid.New().String()
	}

	// Note: The unique_achievement_per_user index is a functional index using COALESCE,
	// so we can't use ON CONFLICT ON CONSTRAINT. The HasAchievement check in the service
	// layer prevents duplicates. If a rare race condition occurs, the unique index
	// will raise an error which is caught and handled gracefully.
	query := `
		INSERT INTO user_achievements (id, user_id, achievement_type, earned_at, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`

	_, err = r.pool.Exec(ctx, query,
		achievement.ID,
		achievement.UserID,
		achievement.AchievementType,
		achievement.EarnedAt,
		metadataJSON,
	)

	return err
}

// GetAchievements retrieves all achievements for a user
func (r *GamificationRepository) GetAchievements(ctx context.Context, userID string) ([]*domain.UserAchievement, error) {
	query := `
		SELECT id, user_id, achievement_type, earned_at, metadata
		FROM user_achievements
		WHERE user_id = $1
		ORDER BY earned_at DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var achievements []*domain.UserAchievement
	for rows.Next() {
		var a domain.UserAchievement
		var metadataJSON []byte

		if err := rows.Scan(&a.ID, &a.UserID, &a.AchievementType, &a.EarnedAt, &metadataJSON); err != nil {
			return nil, err
		}

		if metadataJSON != nil {
			if err := json.Unmarshal(metadataJSON, &a.Metadata); err != nil {
				slog.Warn("Failed to unmarshal achievement metadata", "id", a.ID, "error", err)
			}
		}

		achievements = append(achievements, &a)
	}

	return achievements, rows.Err()
}

// GetRecentAchievements retrieves the most recent achievements for a user
func (r *GamificationRepository) GetRecentAchievements(ctx context.Context, userID string, limit int) ([]*domain.UserAchievement, error) {
	// Validate limit parameter
	if limit <= 0 {
		limit = 5 // Default
	} else if limit > 100 {
		limit = 100 // Max cap
	}

	query := `
		SELECT id, user_id, achievement_type, earned_at, metadata
		FROM user_achievements
		WHERE user_id = $1
		ORDER BY earned_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var achievements []*domain.UserAchievement
	for rows.Next() {
		var a domain.UserAchievement
		var metadataJSON []byte

		if err := rows.Scan(&a.ID, &a.UserID, &a.AchievementType, &a.EarnedAt, &metadataJSON); err != nil {
			return nil, err
		}

		if metadataJSON != nil {
			if err := json.Unmarshal(metadataJSON, &a.Metadata); err != nil {
				slog.Warn("Failed to unmarshal achievement metadata", "id", a.ID, "error", err)
			}
		}

		achievements = append(achievements, &a)
	}

	return achievements, rows.Err()
}

// HasAchievement checks if a user has earned a specific achievement
func (r *GamificationRepository) HasAchievement(ctx context.Context, userID string, achievementType domain.AchievementType, category *string) (bool, error) {
	var query string
	var args []interface{}

	if category != nil {
		// For category-specific achievements (like category_master)
		query = `
			SELECT EXISTS(
				SELECT 1 FROM user_achievements
				WHERE user_id = $1 AND achievement_type = $2
				AND metadata->>'category' = $3
			)
		`
		args = []interface{}{userID, achievementType, *category}
	} else {
		// For non-category achievements
		query = `
			SELECT EXISTS(
				SELECT 1 FROM user_achievements
				WHERE user_id = $1 AND achievement_type = $2
			)
		`
		args = []interface{}{userID, achievementType}
	}

	var exists bool
	err := r.pool.QueryRow(ctx, query, args...).Scan(&exists)
	return exists, err
}

// GetCategoryMastery retrieves mastery progress for a specific category
func (r *GamificationRepository) GetCategoryMastery(ctx context.Context, userID, category string) (*domain.CategoryMastery, error) {
	query := `
		SELECT id, user_id, category, completed_count, last_completed_at, created_at, updated_at
		FROM category_mastery
		WHERE user_id = $1 AND category = $2
	`

	var m domain.CategoryMastery
	err := r.pool.QueryRow(ctx, query, userID, category).Scan(
		&m.ID, &m.UserID, &m.Category, &m.CompletedCount, &m.LastCompletedAt, &m.CreatedAt, &m.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &m, nil
}

// GetAllCategoryMastery retrieves all category mastery progress for a user
func (r *GamificationRepository) GetAllCategoryMastery(ctx context.Context, userID string) ([]*domain.CategoryMastery, error) {
	query := `
		SELECT id, user_id, category, completed_count, last_completed_at, created_at, updated_at
		FROM category_mastery
		WHERE user_id = $1
		ORDER BY completed_count DESC
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var masteries []*domain.CategoryMastery
	for rows.Next() {
		var m domain.CategoryMastery
		if err := rows.Scan(&m.ID, &m.UserID, &m.Category, &m.CompletedCount, &m.LastCompletedAt, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		masteries = append(masteries, &m)
	}

	return masteries, rows.Err()
}

// IncrementCategoryMastery increments the completion count for a category
func (r *GamificationRepository) IncrementCategoryMastery(ctx context.Context, userID, category string) (*domain.CategoryMastery, error) {
	query := `
		INSERT INTO category_mastery (id, user_id, category, completed_count, last_completed_at)
		VALUES ($1, $2, $3, 1, NOW())
		ON CONFLICT (user_id, category) DO UPDATE SET
			completed_count = category_mastery.completed_count + 1,
			last_completed_at = NOW(),
			updated_at = NOW()
		RETURNING id, user_id, category, completed_count, last_completed_at, created_at, updated_at
	`

	var m domain.CategoryMastery
	err := r.pool.QueryRow(ctx, query, uuid.New().String(), userID, category).Scan(
		&m.ID, &m.UserID, &m.Category, &m.CompletedCount, &m.LastCompletedAt, &m.CreatedAt, &m.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &m, nil
}

// GetTotalCompletedTasks returns the total number of completed tasks for a user
func (r *GamificationRepository) GetTotalCompletedTasks(ctx context.Context, userID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM tasks
		WHERE user_id = $1 AND status = 'done'
	`

	var count int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

// GetCompletionsByDate returns task completions grouped by date for streak calculation
func (r *GamificationRepository) GetCompletionsByDate(ctx context.Context, userID string, timezone string, daysBack int) (map[string]int, error) {
	query := `
		SELECT DATE(completed_at AT TIME ZONE $2) as completion_date, COUNT(*) as count
		FROM tasks
		WHERE user_id = $1
		  AND status = 'done'
		  AND completed_at IS NOT NULL
		  AND completed_at >= NOW() - INTERVAL '1 day' * $3
		GROUP BY completion_date
		ORDER BY completion_date DESC
	`

	rows, err := r.pool.Query(ctx, query, userID, timezone, daysBack)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	completions := make(map[string]int)
	for rows.Next() {
		var date time.Time
		var count int
		if err := rows.Scan(&date, &count); err != nil {
			return nil, err
		}
		completions[date.Format("2006-01-02")] = count
	}

	return completions, rows.Err()
}

// GetOnTimeCompletionRate returns the percentage of tasks completed on or before due date
func (r *GamificationRepository) GetOnTimeCompletionRate(ctx context.Context, userID string, daysBack int) (float64, error) {
	query := `
		SELECT
			COUNT(*) FILTER (WHERE completed_at <= due_date) as on_time,
			COUNT(*) as total
		FROM tasks
		WHERE user_id = $1
		  AND status = 'done'
		  AND due_date IS NOT NULL
		  AND completed_at IS NOT NULL
		  AND completed_at >= NOW() - INTERVAL '1 day' * $2
	`

	var onTime, total int
	err := r.pool.QueryRow(ctx, query, userID, daysBack).Scan(&onTime, &total)
	if err != nil {
		return 0, err
	}

	if total == 0 {
		return 100.0, nil // No tasks with due dates = 100% on time
	}

	return float64(onTime) / float64(total) * 100, nil
}

// GetEffortDistribution returns the distribution of tasks by effort level
func (r *GamificationRepository) GetEffortDistribution(ctx context.Context, userID string, daysBack int) (map[domain.TaskEffort]int, error) {
	query := `
		SELECT estimated_effort, COUNT(*) as count
		FROM tasks
		WHERE user_id = $1
		  AND status = 'done'
		  AND estimated_effort IS NOT NULL
		  AND completed_at >= NOW() - INTERVAL '1 day' * $2
		GROUP BY estimated_effort
	`

	rows, err := r.pool.Query(ctx, query, userID, daysBack)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	distribution := make(map[domain.TaskEffort]int)
	for rows.Next() {
		var effort domain.TaskEffort
		var count int
		if err := rows.Scan(&effort, &count); err != nil {
			return nil, err
		}
		distribution[effort] = count
	}

	return distribution, rows.Err()
}

// GetSpeedCompletions returns the count of tasks completed within 24 hours of creation
func (r *GamificationRepository) GetSpeedCompletions(ctx context.Context, userID string) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM tasks
		WHERE user_id = $1
		  AND status = 'done'
		  AND completed_at IS NOT NULL
		  AND (completed_at - created_at) <= INTERVAL '24 hours'
	`

	var count int
	err := r.pool.QueryRow(ctx, query, userID).Scan(&count)
	return count, err
}

// GetWeeklyCompletionDays returns the number of unique days with completions in a given week
func (r *GamificationRepository) GetWeeklyCompletionDays(ctx context.Context, userID string, weekStart string, timezone string) (int, error) {
	query := `
		SELECT COUNT(DISTINCT DATE(completed_at AT TIME ZONE $3))
		FROM tasks
		WHERE user_id = $1
		  AND status = 'done'
		  AND completed_at IS NOT NULL
		  AND DATE(completed_at AT TIME ZONE $3) >= $2::date
		  AND DATE(completed_at AT TIME ZONE $3) < $2::date + INTERVAL '7 days'
	`

	var count int
	err := r.pool.QueryRow(ctx, query, userID, weekStart, timezone).Scan(&count)
	return count, err
}

// GetUserTimezone retrieves the user's timezone setting
func (r *GamificationRepository) GetUserTimezone(ctx context.Context, userID string) (string, error) {
	query := `
		SELECT COALESCE(timezone, 'UTC')
		FROM user_preferences
		WHERE user_id = $1
	`

	var timezone string
	err := r.pool.QueryRow(ctx, query, userID).Scan(&timezone)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "UTC", nil
		}
		return "", err
	}

	return timezone, nil
}

// SetUserTimezone updates the user's timezone setting
func (r *GamificationRepository) SetUserTimezone(ctx context.Context, userID, timezone string) error {
	// Validate timezone
	_, err := time.LoadLocation(timezone)
	if err != nil {
		return domain.ErrInvalidTimezone
	}

	query := `
		INSERT INTO user_preferences (user_id, timezone)
		VALUES ($1, $2)
		ON CONFLICT (user_id) DO UPDATE SET
			timezone = EXCLUDED.timezone,
			updated_at = NOW()
	`

	_, err = r.pool.Exec(ctx, query, userID, timezone)
	return err
}

// DecrementCategoryMastery decreases the completion count for a category
func (r *GamificationRepository) DecrementCategoryMastery(ctx context.Context, userID, category string) error {
	query := `
		UPDATE category_mastery
		SET completed_count = GREATEST(completed_count - 1, 0),
			updated_at = NOW()
		WHERE user_id = $1 AND category = $2
	`
	_, err := r.pool.Exec(ctx, query, userID, category)
	return err
}

// RevokeAchievement removes an earned achievement from a user
func (r *GamificationRepository) RevokeAchievement(ctx context.Context, userID string, achievementType domain.AchievementType) error {
	query := `
		DELETE FROM user_achievements
		WHERE user_id = $1 AND achievement_type = $2
	`
	_, err := r.pool.Exec(ctx, query, userID, achievementType)
	return err
}
