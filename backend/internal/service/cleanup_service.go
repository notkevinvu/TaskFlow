package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/notkevinvu/taskflow/backend/internal/ports"
)

// CleanupService handles periodic cleanup of expired anonymous users
type CleanupService struct {
	userRepo ports.UserRepository
}

// NewCleanupService creates a new cleanup service
func NewCleanupService(userRepo ports.UserRepository) *CleanupService {
	return &CleanupService{
		userRepo: userRepo,
	}
}

// CleanupResult contains the result of a cleanup operation
type CleanupResult struct {
	DeletedCount int
	FailedCount  int
	TotalTasks   int
	Duration     time.Duration
	Errors       []string
}

// CleanupExpiredAnonymousUsers finds and deletes all expired anonymous users
// along with their associated data (tasks, etc. via cascade delete).
// It logs each deletion for audit purposes.
func (s *CleanupService) CleanupExpiredAnonymousUsers(ctx context.Context) (*CleanupResult, error) {
	startTime := time.Now()
	result := &CleanupResult{
		Errors: []string{},
	}

	slog.Info("[Cleanup] Starting anonymous user cleanup")

	// Check if context is already cancelled before starting
	select {
	case <-ctx.Done():
		slog.Info("[Cleanup] Context cancelled before start")
		result.Duration = time.Since(startTime)
		return result, ctx.Err()
	default:
	}

	// Find all expired anonymous users
	expiredUsers, err := s.userRepo.FindExpiredAnonymous(ctx)
	if err != nil {
		slog.Error("[Cleanup] Failed to find expired users", "error", err)
		return nil, err
	}

	if len(expiredUsers) == 0 {
		slog.Info("[Cleanup] No expired anonymous users found")
		result.Duration = time.Since(startTime)
		return result, nil
	}

	slog.Info("[Cleanup] Found expired anonymous users", "count", len(expiredUsers))

	// Process each expired user
	for _, user := range expiredUsers {
		// Count tasks before deletion (for audit)
		taskCount, err := s.userRepo.CountTasksByUserID(ctx, user.ID)
		if err != nil {
			slog.Error("[Cleanup] Failed to count tasks for user - skipping to preserve audit integrity",
				"user_id", user.ID,
				"error", err,
			)
			result.FailedCount++
			result.Errors = append(result.Errors, "task count failed for user "+user.ID+": "+err.Error())
			continue // Skip deletion - cannot create accurate audit record
		}
		result.TotalTasks += taskCount

		// Log the cleanup for audit purposes - REQUIRED before deletion
		if err := s.userRepo.LogAnonymousCleanup(ctx, user.ID, taskCount, user.CreatedAt); err != nil {
			slog.Error("[Cleanup] Failed to log cleanup - skipping deletion to preserve audit trail",
				"user_id", user.ID,
				"task_count", taskCount,
				"error", err,
			)
			result.FailedCount++
			result.Errors = append(result.Errors, "audit log failed for user "+user.ID+": "+err.Error())
			continue // Skip deletion - audit is required for compliance
		}

		// Delete the user (cascades to tasks via FK)
		if err := s.userRepo.Delete(ctx, user.ID); err != nil {
			slog.Error("[Cleanup] Failed to delete user",
				"user_id", user.ID,
				"error", err,
			)
			result.FailedCount++
			result.Errors = append(result.Errors, err.Error())
			continue
		}

		result.DeletedCount++
		slog.Info("[Cleanup] Deleted expired anonymous user",
			"user_id", user.ID,
			"task_count", taskCount,
			"created_at", user.CreatedAt,
			"expired_at", user.ExpiresAt,
		)
	}

	result.Duration = time.Since(startTime)
	slog.Info("[Cleanup] Cleanup completed",
		"deleted", result.DeletedCount,
		"failed", result.FailedCount,
		"total_tasks", result.TotalTasks,
		"duration", result.Duration,
	)

	return result, nil
}

// RunCleanupLoop starts a background loop that runs cleanup at the specified interval.
// It blocks until the context is cancelled.
func (s *CleanupService) RunCleanupLoop(ctx context.Context, interval time.Duration) {
	slog.Info("[Cleanup] Starting cleanup loop", "interval", interval)

	// Run immediately on start
	if _, err := s.CleanupExpiredAnonymousUsers(ctx); err != nil {
		slog.Error("[Cleanup] Initial cleanup failed", "error", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			slog.Info("[Cleanup] Cleanup loop stopped")
			return
		case <-ticker.C:
			if _, err := s.CleanupExpiredAnonymousUsers(ctx); err != nil {
				slog.Error("[Cleanup] Scheduled cleanup failed", "error", err)
			}
		}
	}
}
