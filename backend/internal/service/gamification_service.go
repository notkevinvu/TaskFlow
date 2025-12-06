package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"time"

	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/ports"
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
	previousStats, _ := s.gamificationRepo.GetStats(ctx, userID)
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
		slog.Warn("Failed to persist gamification stats", "user_id", userID, "error", err)
	}

	return &domain.TaskCompletionGamificationResult{
		UpdatedStats:    stats,
		NewAchievements: newAchievements,
		StreakExtended:  stats.CurrentStreak > previousStreak,
		PreviousStreak:  previousStreak,
	}, nil
}

// ComputeStats calculates all gamification stats from scratch
func (s *GamificationService) ComputeStats(ctx context.Context, userID string) (*domain.GamificationStats, error) {
	timezone, err := s.GetUserTimezone(ctx, userID)
	if err != nil {
		timezone = "UTC"
	}

	// Get total completed tasks
	totalCompleted, err := s.gamificationRepo.GetTotalCompletedTasks(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get total completed: %w", err)
	}

	// Compute streaks (check last 365 days of completions)
	streakResult := s.computeStreaks(ctx, userID, timezone)

	// Compute productivity score components (last 30 days)
	completionRate := s.computeCompletionRate(ctx, userID, 30)
	streakScore := s.computeStreakScore(streakResult.CurrentStreak)
	onTimePercentage, _ := s.gamificationRepo.GetOnTimeCompletionRate(ctx, userID, 30)
	effortMixScore := s.computeEffortMixScore(ctx, userID, 30)

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
	if err != nil || stats.TotalTasks == 0 {
		return 0
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
	if err != nil || len(distribution) == 0 {
		return 50 // Default score if no effort data
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
func (s *GamificationService) CheckAndAwardAchievements(
	ctx context.Context,
	userID string,
	task *domain.Task,
	stats *domain.GamificationStats,
) ([]*domain.AchievementEarnedEvent, error) {
	var newAchievements []*domain.AchievementEarnedEvent

	// Check milestone achievements
	for achievementType, threshold := range domain.MilestoneThresholds {
		if stats.TotalCompleted >= threshold {
			if earned := s.awardAchievementIfNew(ctx, userID, achievementType, nil); earned != nil {
				newAchievements = append(newAchievements, earned)
			}
		}
	}

	// Check streak achievements
	for achievementType, threshold := range domain.StreakThresholds {
		if stats.CurrentStreak >= threshold {
			if earned := s.awardAchievementIfNew(ctx, userID, achievementType, nil); earned != nil {
				newAchievements = append(newAchievements, earned)
			}
		}
	}

	// Check category mastery (10 tasks in same category)
	if task.Category != nil && *task.Category != "" {
		mastery, _ := s.gamificationRepo.GetCategoryMastery(ctx, userID, *task.Category)
		if mastery != nil && mastery.CompletedCount >= domain.CategoryMasteryThreshold {
			category := *task.Category
			if earned := s.awardAchievementIfNew(ctx, userID, domain.AchievementCategoryMaster, &category); earned != nil {
				newAchievements = append(newAchievements, earned)
			}
		}
	}

	// Check speed demon (5+ tasks completed within 24h of creation)
	speedCount, _ := s.gamificationRepo.GetSpeedCompletions(ctx, userID)
	if speedCount >= domain.SpeedDemonThreshold {
		if earned := s.awardAchievementIfNew(ctx, userID, domain.AchievementSpeedDemon, nil); earned != nil {
			newAchievements = append(newAchievements, earned)
		}
	}

	// Check consistency king (5+ days in current week)
	timezone, _ := s.GetUserTimezone(ctx, userID)
	loc, _ := time.LoadLocation(timezone)
	if loc == nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)

	// Get start of week (Sunday)
	weekStart := now.AddDate(0, 0, -int(now.Weekday()))
	weekStart = time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, loc)
	weekStartStr := weekStart.Format("2006-01-02")

	daysWithCompletions, _ := s.gamificationRepo.GetWeeklyCompletionDays(ctx, userID, weekStartStr, timezone)
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
	has, _ := s.gamificationRepo.HasAchievement(ctx, userID, achievementType, category)
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
