package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/ports"
	"golang.org/x/sync/errgroup"
)

// GamificationService provides gamification business logic
type GamificationService struct {
	gamificationRepo ports.GamificationRepository
	taskRepo         ports.TaskRepository
}

// NewGamificationService creates a new gamification service
func NewGamificationService(
	gamificationRepo ports.GamificationRepository,
	taskRepo ports.TaskRepository,
) *GamificationService {
	return &GamificationService{
		gamificationRepo: gamificationRepo,
		taskRepo:         taskRepo,
	}
}

// GetDashboard returns all gamification data for the dashboard
func (s *GamificationService) GetDashboard(ctx context.Context, userID string) (*domain.GamificationDashboard, error) {
	// Get or compute stats
	stats, err := s.gamificationRepo.GetStats(ctx, userID)
	if err != nil {
		if errors.Is(err, domain.ErrGamificationStatsNotFound) {
			// Compute stats for first time
			stats, err = s.ComputeStats(ctx, userID)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	// Get all achievements
	allAchievements, err := s.gamificationRepo.GetAchievements(ctx, userID)
	if err != nil {
		slog.Warn("Failed to get achievements", "user_id", userID, "error", err)
		allAchievements = []*domain.UserAchievement{}
	}

	// Get recent achievements (last 5)
	recentAchievements, err := s.gamificationRepo.GetRecentAchievements(ctx, userID, 5)
	if err != nil {
		slog.Warn("Failed to get recent achievements", "user_id", userID, "error", err)
		recentAchievements = []*domain.UserAchievement{}
	}

	// Get category progress
	categoryProgress, err := s.gamificationRepo.GetAllCategoryMastery(ctx, userID)
	if err != nil {
		slog.Warn("Failed to get category mastery", "user_id", userID, "error", err)
		categoryProgress = []*domain.CategoryMastery{}
	}

	return &domain.GamificationDashboard{
		Stats:                 stats,
		RecentAchievements:    recentAchievements,
		AllAchievements:       allAchievements,
		AvailableAchievements: domain.GetAchievementDefinitions(),
		CategoryProgress:      categoryProgress,
		UnviewedCount:         0, // TODO: Track viewed status if needed
	}, nil
}

// ProcessTaskCompletion handles gamification logic when a task is completed
func (s *GamificationService) ProcessTaskCompletion(
	ctx context.Context,
	userID string,
	task *domain.Task,
) (*domain.TaskCompletionGamificationResult, error) {
	// Get previous stats for streak comparison
	previousStats, err := s.gamificationRepo.GetStats(ctx, userID)
	if err != nil && !errors.Is(err, domain.ErrGamificationStatsNotFound) {
		slog.Warn("Failed to get previous stats for streak comparison",
			"user_id", userID, "error", err)
	}
	previousStreak := 0
	if previousStats != nil {
		previousStreak = previousStats.CurrentStreak
	}

	// Update category mastery if task has category
	if task.Category != nil && *task.Category != "" {
		_, err := s.gamificationRepo.IncrementCategoryMastery(ctx, userID, *task.Category)
		if err != nil {
			slog.Warn("Failed to increment category mastery",
				"user_id", userID, "category", *task.Category, "error", err)
		}
	}

	// Compute updated stats
	stats, err := s.ComputeStats(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check for new achievements
	newAchievements, err := s.CheckAndAwardAchievements(ctx, userID, task, stats)
	if err != nil {
		slog.Warn("Failed to check achievements", "user_id", userID, "error", err)
		newAchievements = []*domain.AchievementEarnedEvent{}
	}

	// Persist updated stats
	if err := s.gamificationRepo.UpsertStats(ctx, stats); err != nil {
		slog.Error("Failed to persist gamification stats - progress may be lost",
			"user_id", userID, "error", err)
	}

	return &domain.TaskCompletionGamificationResult{
		UpdatedStats:    stats,
		NewAchievements: newAchievements,
		StreakExtended:  stats.CurrentStreak > previousStreak,
		PreviousStreak:  previousStreak,
	}, nil
}

// ProcessTaskCompletionAsync processes gamification in a background goroutine.
// This allows the task completion API to return immediately while gamification
// (which involves multiple DB queries) happens asynchronously.
// Uses context.Background() since the HTTP request context may be cancelled.
func (s *GamificationService) ProcessTaskCompletionAsync(userID string, task *domain.Task) {
	go func() {
		// Use a background context with a reasonable timeout
		// The original request context is likely to be cancelled after HTTP response
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		result, err := s.ProcessTaskCompletion(ctx, userID, task)
		if err != nil {
			slog.Error("Async gamification processing failed",
				"user_id", userID,
				"task_id", task.ID,
				"error", err)
			return
		}

		// Log achievements earned (useful for monitoring)
		if len(result.NewAchievements) > 0 {
			achievementTypes := make([]string, len(result.NewAchievements))
			for i, a := range result.NewAchievements {
				achievementTypes[i] = string(a.Achievement.AchievementType)
			}
			slog.Info("User earned new achievements",
				"user_id", userID,
				"task_id", task.ID,
				"achievements", achievementTypes)
		}

		if result.StreakExtended {
			slog.Debug("User streak extended",
				"user_id", userID,
				"previous_streak", result.PreviousStreak,
				"new_streak", result.UpdatedStats.CurrentStreak)
		}
	}()
}

// ComputeStats calculates all gamification stats from scratch.
// Uses parallel queries via errgroup for improved performance.
func (s *GamificationService) ComputeStats(ctx context.Context, userID string) (*domain.GamificationStats, error) {
	// Get timezone first (needed for streak computation)
	timezone, err := s.GetUserTimezone(ctx, userID)
	if err != nil {
		slog.Warn("Failed to get user timezone, using UTC",
			"user_id", userID, "error", err)
		timezone = "UTC"
	}

	// Run all independent queries in parallel using errgroup
	g, gCtx := errgroup.WithContext(ctx)

	var (
		totalCompleted   int
		streakResult     domain.StreakCalculationResult
		completionRate   float64
		onTimePercentage float64
		effortMixScore   float64
		mu               sync.Mutex // Protect shared writes
	)

	// Query 1: Total completed tasks
	g.Go(func() error {
		count, err := s.gamificationRepo.GetTotalCompletedTasks(gCtx, userID)
		if err != nil {
			return fmt.Errorf("failed to get total completed: %w", err)
		}
		mu.Lock()
		totalCompleted = count
		mu.Unlock()
		return nil
	})

	// Query 2: Compute streaks (involves GetCompletionsByDate)
	g.Go(func() error {
		result := s.computeStreaks(gCtx, userID, timezone)
		mu.Lock()
		streakResult = result
		mu.Unlock()
		return nil
	})

	// Query 3: Completion rate (involves GetCompletionStats)
	g.Go(func() error {
		rate := s.computeCompletionRate(gCtx, userID, 30)
		mu.Lock()
		completionRate = rate
		mu.Unlock()
		return nil
	})

	// Query 4: On-time completion percentage
	g.Go(func() error {
		percentage, err := s.gamificationRepo.GetOnTimeCompletionRate(gCtx, userID, 30)
		if err != nil {
			slog.Warn("Failed to get on-time completion rate, using 0%",
				"user_id", userID, "error", err)
		}
		mu.Lock()
		onTimePercentage = percentage
		mu.Unlock()
		return nil
	})

	// Query 5: Effort mix score (involves GetEffortDistribution)
	g.Go(func() error {
		score := s.computeEffortMixScore(gCtx, userID, 30)
		mu.Lock()
		effortMixScore = score
		mu.Unlock()
		return nil
	})

	// Wait for all queries to complete
	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Compute productivity score from parallel results
	streakScore := s.computeStreakScore(streakResult.CurrentStreak)

	// Weighted productivity score formula
	productivityScore := (completionRate * domain.ProductivityScoreWeights.CompletionRate) +
		(streakScore * domain.ProductivityScoreWeights.Streak) +
		(onTimePercentage * domain.ProductivityScoreWeights.OnTime) +
		(effortMixScore * domain.ProductivityScoreWeights.EffortMix)

	// Cap at 100
	productivityScore = math.Min(productivityScore, 100)

	return &domain.GamificationStats{
		UserID:             userID,
		CurrentStreak:      streakResult.CurrentStreak,
		LongestStreak:      streakResult.LongestStreak,
		LastCompletionDate: streakResult.LastCompletionDate,
		TotalCompleted:     totalCompleted,
		ProductivityScore:  productivityScore,
		CompletionRate:     completionRate,
		StreakScore:        streakScore,
		OnTimePercentage:   onTimePercentage,
		EffortMixScore:     effortMixScore,
		LastComputedAt:     time.Now(),
	}, nil
}

// computeStreaks calculates current and longest streaks
func (s *GamificationService) computeStreaks(ctx context.Context, userID, timezone string) domain.StreakCalculationResult {
	result := domain.StreakCalculationResult{}

	// Get completions by date for last 365 days
	completionsByDate, err := s.gamificationRepo.GetCompletionsByDate(ctx, userID, timezone, 365)
	if err != nil || len(completionsByDate) == 0 {
		return result
	}

	// Load user timezone
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		slog.Warn("Failed to load timezone location, using UTC",
			"timezone", timezone, "error", err)
		loc = time.UTC
	}

	now := time.Now().In(loc)
	today := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

	// Find last completion date
	var dates []string
	for date := range completionsByDate {
		dates = append(dates, date)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(dates)))

	if len(dates) > 0 {
		result.LastCompletionDate = &dates[0]
	}

	// Compute current streak (working backwards from today/yesterday)
	currentStreak := 0
	checkDate := today

	// If no completion today, start from yesterday
	if _, hasToday := completionsByDate[today]; !hasToday {
		checkDate = yesterday
		// If also no completion yesterday, streak is 0
		if _, hasYesterday := completionsByDate[yesterday]; !hasYesterday {
			// Check if streak was extended today
			result.CurrentStreak = 0
			result.LongestStreak = s.computeLongestStreak(dates)
			return result
		}
	}

	// Count consecutive days backwards
	for {
		if _, exists := completionsByDate[checkDate]; exists {
			currentStreak++
			t, err := time.Parse("2006-01-02", checkDate)
			if err != nil {
				break
			}
			checkDate = t.AddDate(0, 0, -1).Format("2006-01-02")
		} else {
			break
		}
	}

	result.CurrentStreak = currentStreak
	result.LongestStreak = s.computeLongestStreak(dates)

	// Update longest if current is longer
	if result.CurrentStreak > result.LongestStreak {
		result.LongestStreak = result.CurrentStreak
	}

	return result
}

// computeLongestStreak finds the longest consecutive sequence of dates
func (s *GamificationService) computeLongestStreak(dates []string) int {
	if len(dates) == 0 {
		return 0
	}

	// Sort dates ascending
	sortedDates := make([]string, len(dates))
	copy(sortedDates, dates)
	sort.Strings(sortedDates)

	longest := 1
	current := 1

	for i := 1; i < len(sortedDates); i++ {
		prevDate, _ := time.Parse("2006-01-02", sortedDates[i-1])
		currDate, _ := time.Parse("2006-01-02", sortedDates[i])

		expectedNext := prevDate.AddDate(0, 0, 1)
		if currDate.Equal(expectedNext) {
			current++
			if current > longest {
				longest = current
			}
		} else {
			current = 1
		}
	}

	return longest
}

// computeCompletionRate calculates the task completion rate
func (s *GamificationService) computeCompletionRate(ctx context.Context, userID string, daysBack int) float64 {
	stats, err := s.taskRepo.GetCompletionStats(ctx, userID, daysBack)
	if err != nil {
		slog.Warn("Failed to get completion stats, using 0%",
			"user_id", userID, "days_back", daysBack, "error", err)
		return 0
	}
	if stats.TotalTasks == 0 {
		return 0 // Legitimate case - no tasks in period
	}

	return float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
}

// computeStreakScore converts streak days to a 0-100 score
func (s *GamificationService) computeStreakScore(currentStreak int) float64 {
	// Scale: 30-day streak = 100 points
	// Linear up to 30 days, capped at 100
	score := float64(currentStreak) / 30.0 * 100
	return math.Min(score, 100)
}

// computeEffortMixScore calculates how balanced the effort distribution is
func (s *GamificationService) computeEffortMixScore(ctx context.Context, userID string, daysBack int) float64 {
	distribution, err := s.gamificationRepo.GetEffortDistribution(ctx, userID, daysBack)
	if err != nil {
		slog.Warn("Failed to get effort distribution, using default score",
			"user_id", userID, "days_back", daysBack, "error", err)
		return 50
	}
	if len(distribution) == 0 {
		return 50 // Legitimate case - no effort data in period
	}

	// Ideal distribution: 30% small, 40% medium, 20% large, 10% xlarge
	idealRatios := map[domain.TaskEffort]float64{
		domain.TaskEffortSmall:  0.30,
		domain.TaskEffortMedium: 0.40,
		domain.TaskEffortLarge:  0.20,
		domain.TaskEffortXLarge: 0.10,
	}

	// Calculate total tasks
	total := 0
	for _, count := range distribution {
		total += count
	}

	if total == 0 {
		return 50
	}

	// Calculate deviation from ideal
	totalDeviation := 0.0
	for effort, idealRatio := range idealRatios {
		actual := float64(distribution[effort]) / float64(total)
		deviation := math.Abs(actual - idealRatio)
		totalDeviation += deviation
	}

	// Convert deviation to score (0 deviation = 100, max deviation = 0)
	// Max possible deviation is 2.0 (0% everywhere vs 100% in one category)
	score := (1 - totalDeviation/2.0) * 100
	return math.Max(score, 0)
}

// CheckAndAwardAchievements checks if user earned new achievements
// Uses parallel queries for database lookups to improve performance
func (s *GamificationService) CheckAndAwardAchievements(
	ctx context.Context,
	userID string,
	task *domain.Task,
	stats *domain.GamificationStats,
) ([]*domain.AchievementEarnedEvent, error) {
	var newAchievements []*domain.AchievementEarnedEvent
	var mu sync.Mutex // Protect parallel query result variables

	// Check milestone achievements (threshold check uses stats param; awarding makes DB calls)
	for achievementType, threshold := range domain.MilestoneThresholds {
		if stats.TotalCompleted >= threshold {
			if earned := s.awardAchievementIfNew(ctx, userID, achievementType, nil); earned != nil {
				newAchievements = append(newAchievements, earned)
			}
		}
	}

	// Check streak achievements (threshold check uses stats param; awarding makes DB calls)
	for achievementType, threshold := range domain.StreakThresholds {
		if stats.CurrentStreak >= threshold {
			if earned := s.awardAchievementIfNew(ctx, userID, achievementType, nil); earned != nil {
				newAchievements = append(newAchievements, earned)
			}
		}
	}

	// Get timezone first (needed for consistency check)
	timezone, tzErr := s.GetUserTimezone(ctx, userID)
	if tzErr != nil {
		slog.Warn("Failed to get timezone for consistency check, using UTC",
			"user_id", userID, "error", tzErr)
		timezone = "UTC"
	}
	loc, locErr := time.LoadLocation(timezone)
	if locErr != nil {
		slog.Warn("Failed to load timezone for consistency check, using UTC",
			"timezone", timezone, "error", locErr)
		loc = time.UTC
	}
	now := time.Now().In(loc)
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, loc)
	weekStartStr := weekStart.Format("2006-01-02")

	// Run three independent DB queries in parallel
	g, gCtx := errgroup.WithContext(ctx)

	var (
		mastery              *domain.CategoryMastery
		speedCount           int
		daysWithCompletions  int
		categoryToCheck      string
		hasCategoryToCheck   bool
	)

	// Determine if we need to check category mastery
	if task.Category != nil && *task.Category != "" {
		categoryToCheck = *task.Category
		hasCategoryToCheck = true
	}

	// Query 1: Category mastery (if applicable)
	if hasCategoryToCheck {
		g.Go(func() error {
			m, err := s.gamificationRepo.GetCategoryMastery(gCtx, userID, categoryToCheck)
			if err != nil {
				slog.Warn("Failed to get category mastery for achievement check",
					"user_id", userID, "category", categoryToCheck, "error", err)
				return nil // Don't fail the whole check
			}
			mu.Lock()
			mastery = m
			mu.Unlock()
			return nil
		})
	}

	// Query 2: Speed completions
	g.Go(func() error {
		count, err := s.gamificationRepo.GetSpeedCompletions(gCtx, userID)
		if err != nil {
			slog.Warn("Failed to get speed completions for achievement check",
				"user_id", userID, "error", err)
			return nil // Don't fail the whole check
		}
		mu.Lock()
		speedCount = count
		mu.Unlock()
		return nil
	})

	// Query 3: Weekly completion days
	g.Go(func() error {
		days, err := s.gamificationRepo.GetWeeklyCompletionDays(gCtx, userID, weekStartStr, timezone)
		if err != nil {
			slog.Warn("Failed to get weekly completion days for consistency check",
				"user_id", userID, "week_start", weekStartStr, "error", err)
			return nil // Don't fail the whole check
		}
		mu.Lock()
		daysWithCompletions = days
		mu.Unlock()
		return nil
	})

	// Wait for all parallel queries - individual goroutines log their own errors
	if err := g.Wait(); err != nil {
		slog.Warn("Unexpected error from achievement check errgroup",
			"user_id", userID, "error", err)
	}

	// Now check thresholds and award achievements based on parallel query results

	// Category mastery (10 tasks in same category)
	if hasCategoryToCheck && mastery != nil && mastery.CompletedCount >= domain.CategoryMasteryThreshold {
		if earned := s.awardAchievementIfNew(ctx, userID, domain.AchievementCategoryMaster, &categoryToCheck); earned != nil {
			newAchievements = append(newAchievements, earned)
		}
	}

	// Speed demon (5+ tasks completed within 24h of creation)
	if speedCount >= domain.SpeedDemonThreshold {
		if earned := s.awardAchievementIfNew(ctx, userID, domain.AchievementSpeedDemon, nil); earned != nil {
			newAchievements = append(newAchievements, earned)
		}
	}

	// Consistency king (5+ days in current week)
	if daysWithCompletions >= domain.ConsistencyKingThreshold {
		if earned := s.awardAchievementIfNew(ctx, userID, domain.AchievementConsistencyKing, nil); earned != nil {
			newAchievements = append(newAchievements, earned)
		}
	}

	return newAchievements, nil
}

// awardAchievementIfNew creates achievement if not already earned
func (s *GamificationService) awardAchievementIfNew(
	ctx context.Context,
	userID string,
	achievementType domain.AchievementType,
	category *string,
) *domain.AchievementEarnedEvent {
	// Check if already earned
	has, err := s.gamificationRepo.HasAchievement(ctx, userID, achievementType, category)
	if err != nil {
		slog.Warn("Failed to check if achievement exists, attempting to create",
			"user_id", userID, "type", achievementType, "error", err)
		// Fall through to attempt create - database constraint will handle duplicate
	}
	if has {
		return nil
	}

	// Create achievement
	achievement := &domain.UserAchievement{
		UserID:          userID,
		AchievementType: achievementType,
		EarnedAt:        time.Now(),
	}

	if category != nil {
		achievement.Metadata = map[string]interface{}{
			"category": *category,
		}
	}

	if err := s.gamificationRepo.CreateAchievement(ctx, achievement); err != nil {
		slog.Warn("Failed to create achievement",
			"user_id", userID, "type", achievementType, "error", err)
		return nil
	}

	def := domain.GetAchievementDefinition(achievementType)
	if def == nil {
		return nil
	}

	// Customize title for category master
	customDef := *def
	if achievementType == domain.AchievementCategoryMaster && category != nil {
		customDef.Title = fmt.Sprintf("%s Expert", *category)
		customDef.Description = fmt.Sprintf("Complete 10 tasks in %s", *category)
	}

	return &domain.AchievementEarnedEvent{
		Achievement: achievement,
		Definition:  customDef,
	}
}

// SetUserTimezone updates the user's timezone setting
func (s *GamificationService) SetUserTimezone(ctx context.Context, userID, timezone string) error {
	return s.gamificationRepo.SetUserTimezone(ctx, userID, timezone)
}

// GetUserTimezone retrieves the user's timezone setting
func (s *GamificationService) GetUserTimezone(ctx context.Context, userID string) (string, error) {
	return s.gamificationRepo.GetUserTimezone(ctx, userID)
}

// ProcessTaskUncompletion reverses gamification effects when a task is uncompleted
func (s *GamificationService) ProcessTaskUncompletion(
	ctx context.Context,
	userID string,
	task *domain.Task,
) error {
	// 1. Decrement category mastery if task had category
	if task.Category != nil && *task.Category != "" {
		if err := s.gamificationRepo.DecrementCategoryMastery(ctx, userID, *task.Category); err != nil {
			slog.Warn("Failed to decrement category mastery",
				"user_id", userID,
				"category", *task.Category,
				"error", err,
			)
		}
	}

	// 2. Recompute stats from source of truth (completed tasks)
	stats, err := s.ComputeStats(ctx, userID)
	if err != nil {
		slog.Error("Failed to recompute stats during uncompletion",
			"user_id", userID,
			"error", err,
		)
		return err
	}

	// 3. Check and revoke achievements user no longer qualifies for
	if err := s.revokeInvalidAchievements(ctx, userID, stats); err != nil {
		slog.Warn("Failed to revoke invalid achievements",
			"user_id", userID,
			"error", err,
		)
	}

	// 4. Persist updated stats
	if err := s.gamificationRepo.UpsertStats(ctx, stats); err != nil {
		slog.Error("Failed to persist updated gamification stats",
			"user_id", userID,
			"error", err,
		)
		return err
	}

	return nil
}

// ProcessTaskUncompletionAsync runs uncompletion processing in background goroutine
func (s *GamificationService) ProcessTaskUncompletionAsync(userID string, task *domain.Task) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := s.ProcessTaskUncompletion(ctx, userID, task); err != nil {
			slog.Error("Async gamification uncompletion failed",
				"user_id", userID,
				"task_id", task.ID,
				"error", err,
			)
		}
	}()
}

// revokeInvalidAchievements checks all earned achievements and revokes those the user no longer qualifies for
func (s *GamificationService) revokeInvalidAchievements(
	ctx context.Context,
	userID string,
	stats *domain.GamificationStats,
) error {
	achievements, err := s.gamificationRepo.GetAchievements(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get achievements: %w", err)
	}

	// Get category mastery data for checking category master achievements
	categoryMastery, err := s.gamificationRepo.GetAllCategoryMastery(ctx, userID)
	if err != nil {
		slog.Warn("Failed to get category mastery for achievement check",
			"user_id", userID, "error", err)
		categoryMastery = []*domain.CategoryMastery{}
	}

	for _, achievement := range achievements {
		stillQualifies := s.checkAchievementQualification(achievement.AchievementType, stats, categoryMastery, achievement)

		if !stillQualifies {
			slog.Info("Revoking achievement - user no longer qualifies",
				"user_id", userID,
				"achievement_type", achievement.AchievementType,
			)
			if err := s.gamificationRepo.RevokeAchievement(ctx, userID, achievement.AchievementType); err != nil {
				slog.Warn("Failed to revoke achievement",
					"achievement_type", achievement.AchievementType,
					"error", err,
				)
			}
		}
	}

	return nil
}

// checkAchievementQualification checks if user still qualifies for a specific achievement
func (s *GamificationService) checkAchievementQualification(
	achievementType domain.AchievementType,
	stats *domain.GamificationStats,
	categoryMastery []*domain.CategoryMastery,
	achievement *domain.UserAchievement,
) bool {
	// Check milestone achievements
	if threshold, ok := domain.MilestoneThresholds[achievementType]; ok {
		return stats.TotalCompleted >= threshold
	}

	// Check streak achievements
	if threshold, ok := domain.StreakThresholds[achievementType]; ok {
		// For streaks, we check against LongestStreak since that's historical
		// If they achieved a streak once, it stays (even if current streak is broken)
		return stats.LongestStreak >= threshold
	}

	// Check category master achievement
	if achievementType == domain.AchievementCategoryMaster {
		// Extract category from achievement metadata
		if achievement.Metadata != nil {
			if category, ok := achievement.Metadata["category"].(string); ok {
				// Find mastery for this category
				for _, mastery := range categoryMastery {
					if mastery.Category == category {
						// Category master threshold is 10 tasks
						return mastery.CompletedCount >= 10
					}
				}
				// Category not found in mastery list - no longer qualifies
				return false
			}
		}
		// No metadata - can't verify, be conservative and keep
		return true
	}

	// For unknown achievements, keep them (conservative)
	return true
}
