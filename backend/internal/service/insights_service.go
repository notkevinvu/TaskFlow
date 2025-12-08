package service

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/ports"
	"golang.org/x/sync/errgroup"
)

// InsightsService provides smart suggestions based on user behavior patterns
type InsightsService struct {
	taskRepo ports.TaskRepository
}

// NewInsightsService creates a new insights service
func NewInsightsService(taskRepo ports.TaskRepository) *InsightsService {
	return &InsightsService{taskRepo: taskRepo}
}

// GetInsights generates all applicable insights for a user.
// All 6 insight checks run in parallel for better performance.
func (s *InsightsService) GetInsights(ctx context.Context, userID string) (*domain.InsightResponse, error) {
	var insights []domain.Insight
	var mu sync.Mutex

	// Use errgroup to run all checks in parallel
	g, ctx := errgroup.WithContext(ctx)

	// Define all insight check functions
	checks := []func(context.Context, string) *domain.Insight{
		s.checkAvoidancePattern,
		s.checkPeakPerformance,
		s.checkQuickWins,
		s.checkDeadlineClustering,
		s.checkAtRiskTasks,
		s.checkCategoryOverload,
	}

	// Launch all checks in parallel
	for _, check := range checks {
		check := check // Capture loop variable
		g.Go(func() error {
			if insight := check(ctx, userID); insight != nil {
				mu.Lock()
				insights = append(insights, *insight)
				mu.Unlock()
			}
			return nil // Don't fail the group if one check fails - they handle errors internally
		})
	}

	// Wait for all checks to complete
	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Sort by priority (higher first)
	sort.Slice(insights, func(i, j int) bool {
		return insights[i].Priority > insights[j].Priority
	})

	return &domain.InsightResponse{
		Insights: insights,
		CachedAt: time.Now(),
	}, nil
}

// checkAvoidancePattern detects categories with high average bump counts
func (s *InsightsService) checkAvoidancePattern(ctx context.Context, userID string) *domain.Insight {
	stats, err := s.taskRepo.GetCategoryBumpStats(ctx, userID)
	if err != nil {
		slog.Warn("Failed to get category bump stats for avoidance pattern insight",
			"user_id", userID, "error", err)
		return nil
	}
	if len(stats) == 0 {
		return nil
	}

	// Find category with highest avg bumps (threshold: 2.5)
	for _, stat := range stats {
		if stat.AvgBumps >= 2.5 && stat.TaskCount >= 2 {
			return &domain.Insight{
				Type:     domain.InsightAvoidancePattern,
				Title:    "Avoidance Pattern Detected",
				Message:  fmt.Sprintf("You tend to delay '%s' tasks (avg %.1f bumps). Consider breaking them down or scheduling dedicated time.", stat.Category, stat.AvgBumps),
				Priority: domain.InsightPriorityHigh,
				Data: map[string]interface{}{
					"category":   stat.Category,
					"avg_bumps":  stat.AvgBumps,
					"task_count": stat.TaskCount,
				},
				GeneratedAt: time.Now(),
			}
		}
	}

	return nil
}

// checkPeakPerformance identifies the day when user completes most tasks
func (s *InsightsService) checkPeakPerformance(ctx context.Context, userID string) *domain.Insight {
	stats, err := s.taskRepo.GetCompletionByDayOfWeek(ctx, userID, 90) // Last 90 days
	if err != nil {
		slog.Warn("Failed to get completion stats by day of week for peak performance insight",
			"user_id", userID, "error", err)
		return nil
	}
	if len(stats) == 0 {
		return nil
	}

	// Need at least 5 completions to be meaningful
	if stats[0].CompletedCount < 5 {
		return nil
	}

	// Bounds check for DayOfWeek (0=Sunday, 6=Saturday)
	if stats[0].DayOfWeek < 0 || stats[0].DayOfWeek > 6 {
		slog.Warn("Invalid day of week value from database",
			"user_id", userID, "day_of_week", stats[0].DayOfWeek)
		return nil
	}

	dayNames := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	bestDay := dayNames[stats[0].DayOfWeek]

	return &domain.Insight{
		Type:     domain.InsightPeakPerformance,
		Title:    "Peak Performance Day",
		Message:  fmt.Sprintf("You complete most tasks on %s (%d tasks in last 90 days). Schedule important work then!", bestDay, stats[0].CompletedCount),
		Priority: domain.InsightPriorityMedium,
		Data: map[string]interface{}{
			"day_of_week":     stats[0].DayOfWeek,
			"day_name":        bestDay,
			"completed_count": stats[0].CompletedCount,
		},
		GeneratedAt: time.Now(),
	}
}

// checkQuickWins finds small effort tasks that are aging
func (s *InsightsService) checkQuickWins(ctx context.Context, userID string) *domain.Insight {
	tasks, err := s.taskRepo.GetAgingQuickWins(ctx, userID, 5, 10) // 5+ days old, limit 10
	if err != nil {
		slog.Warn("Failed to get aging quick wins for insight",
			"user_id", userID, "error", err)
		return nil
	}
	if len(tasks) == 0 {
		return nil
	}

	// Need at least 2 quick wins to show insight
	if len(tasks) < 2 {
		return nil
	}

	taskTitles := make([]string, 0, 3)
	for i, t := range tasks {
		if i >= 3 {
			break
		}
		taskTitles = append(taskTitles, t.Title)
	}

	actionURL := "/dashboard?effort=small&status=todo"
	return &domain.Insight{
		Type:     domain.InsightQuickWins,
		Title:    "Quick Wins Available",
		Message:  fmt.Sprintf("You have %d quick wins aging for 5+ days. Knock them out to build momentum!", len(tasks)),
		Priority: domain.InsightPriorityMedium,
		ActionURL: &actionURL,
		Data: map[string]interface{}{
			"count":       len(tasks),
			"task_titles": taskTitles,
		},
		GeneratedAt: time.Now(),
	}
}

// checkDeadlineClustering finds dates with multiple tasks due
func (s *InsightsService) checkDeadlineClustering(ctx context.Context, userID string) *domain.Insight {
	clusters, err := s.taskRepo.GetDeadlineClusters(ctx, userID, 14) // Next 14 days
	if err != nil {
		slog.Warn("Failed to get deadline clusters for insight",
			"user_id", userID, "error", err)
		return nil
	}
	if len(clusters) == 0 {
		return nil
	}

	// Get the most clustered date
	cluster := clusters[0]
	dateStr := cluster.DueDate.Format("Jan 2")

	return &domain.Insight{
		Type:     domain.InsightDeadlineClustering,
		Title:    "Deadline Cluster Alert",
		Message:  fmt.Sprintf("You have %d tasks due on %s. Consider spreading them out or starting early.", cluster.TaskCount, dateStr),
		Priority: domain.InsightPriorityHigh,
		Data: map[string]interface{}{
			"due_date":   cluster.DueDate.Format("2006-01-02"),
			"task_count": cluster.TaskCount,
			"titles":     cluster.Titles,
		},
		GeneratedAt: time.Now(),
	}
}

// checkAtRiskTasks finds tasks with high bump counts
func (s *InsightsService) checkAtRiskTasks(ctx context.Context, userID string) *domain.Insight {
	tasks, err := s.taskRepo.FindAtRiskTasks(ctx, userID)
	if err != nil {
		slog.Warn("Failed to get at-risk tasks for insight",
			"user_id", userID, "error", err)
		return nil
	}
	if len(tasks) == 0 {
		return nil
	}

	// Find task with highest bump count
	var worstTask *domain.Task
	maxBumps := 0
	for _, t := range tasks {
		if t.BumpCount > maxBumps {
			maxBumps = t.BumpCount
			worstTask = t
		}
	}

	if worstTask == nil || maxBumps < 3 {
		return nil
	}

	actionURL := fmt.Sprintf("/dashboard?taskId=%s", worstTask.ID)
	return &domain.Insight{
		Type:     domain.InsightAtRiskAlert,
		Title:    "At-Risk Task",
		Message:  fmt.Sprintf("'%s' has been delayed %d times. Time to commit, delegate, or delete?", worstTask.Title, maxBumps),
		Priority: domain.InsightPriorityUrgent,
		ActionURL: &actionURL,
		Data: map[string]interface{}{
			"task_id":    worstTask.ID,
			"task_title": worstTask.Title,
			"bump_count": maxBumps,
			"total_at_risk": len(tasks),
		},
		GeneratedAt: time.Now(),
	}
}

// checkCategoryOverload finds if a single category dominates the backlog
func (s *InsightsService) checkCategoryOverload(ctx context.Context, userID string) *domain.Insight {
	dist, err := s.taskRepo.GetCategoryDistribution(ctx, userID)
	if err != nil {
		slog.Warn("Failed to get category distribution for insight",
			"user_id", userID, "error", err)
		return nil
	}
	if len(dist) == 0 {
		return nil
	}

	// Check if top category is > 40% of backlog and has at least 5 tasks
	topCategory := dist[0]
	if topCategory.Percentage < 40 || topCategory.TaskCount < 5 {
		return nil
	}

	return &domain.Insight{
		Type:     domain.InsightCategoryOverload,
		Title:    "Category Overload",
		Message:  fmt.Sprintf("'%s' dominates your backlog (%.0f%%). Consider batch-processing or delegating these tasks.", topCategory.Category, topCategory.Percentage),
		Priority: domain.InsightPriorityMedium,
		Data: map[string]interface{}{
			"category":   topCategory.Category,
			"task_count": topCategory.TaskCount,
			"percentage": topCategory.Percentage,
		},
		GeneratedAt: time.Now(),
	}
}

// EstimateCompletionTime estimates how long a task will take based on historical data
func (s *InsightsService) EstimateCompletionTime(ctx context.Context, userID string, task *domain.Task) (*domain.TimeEstimate, error) {
	// Get base statistics
	baseStats, err := s.taskRepo.GetCompletionTimeStats(ctx, userID, nil, nil)
	if err != nil {
		return nil, err
	}

	baseEstimate := baseStats.MedianDays
	if baseEstimate == 0 {
		baseEstimate = 3.0 // Default fallback
	}

	// Get category-specific statistics
	categoryFactor := 1.0
	if task.Category != nil {
		catStats, err := s.taskRepo.GetCompletionTimeStats(ctx, userID, task.Category, nil)
		if err == nil && catStats.SampleSize >= 3 && catStats.MedianDays > 0 {
			categoryFactor = catStats.MedianDays / baseEstimate
		}
	}

	// Effort factor based on task effort level
	effortFactors := map[domain.TaskEffort]float64{
		domain.TaskEffortSmall:  0.5,
		domain.TaskEffortMedium: 1.0,
		domain.TaskEffortLarge:  2.0,
		domain.TaskEffortXLarge: 3.5,
	}
	effortFactor := 1.0
	if task.EstimatedEffort != nil {
		if factor, ok := effortFactors[*task.EstimatedEffort]; ok {
			effortFactor = factor
		}
	}

	// Bump factor: each bump adds 20% to estimate
	bumpFactor := 1.0 + (0.2 * float64(task.BumpCount))

	estimatedDays := baseEstimate * categoryFactor * effortFactor * bumpFactor

	// Determine confidence level
	confidence := "medium"
	if baseStats.SampleSize < 5 {
		confidence = "low"
	} else if baseStats.SampleSize >= 20 {
		confidence = "high"
	}

	return &domain.TimeEstimate{
		EstimatedDays:   estimatedDays,
		ConfidenceLevel: confidence,
		BasedOn:         baseStats.SampleSize,
		Factors: domain.TimeEstimateFactor{
			BaseEstimate:   baseEstimate,
			CategoryFactor: categoryFactor,
			EffortFactor:   effortFactor,
			BumpFactor:     bumpFactor,
		},
	}, nil
}

// Category patterns for auto-categorization hints
var categoryPatterns = map[string][]string{
	"Code Review":    {"review", "pr", "pull request", "code review", "cr"},
	"Bug Fix":        {"bug", "fix", "issue", "error", "crash", "broken"},
	"Documentation":  {"doc", "readme", "wiki", "document", "write up"},
	"Meeting":        {"meeting", "sync", "standup", "call", "discuss", "1:1", "one on one"},
	"Research":       {"research", "investigate", "explore", "poc", "spike", "prototype"},
	"Testing":        {"test", "qa", "verify", "validation", "coverage"},
	"Design":         {"design", "mockup", "wireframe", "ui", "ux", "figma"},
	"Refactor":       {"refactor", "cleanup", "tech debt", "improve", "optimize"},
	"DevOps":         {"deploy", "ci", "cd", "pipeline", "infrastructure", "docker", "kubernetes"},
	"Communication":  {"email", "slack", "respond", "reply", "follow up"},
}

// SuggestCategory suggests categories based on task title and description
func (s *InsightsService) SuggestCategory(ctx context.Context, userID string, req *domain.CategorySuggestionRequest) (*domain.CategorySuggestionResponse, error) {
	text := strings.ToLower(req.Title + " " + req.Description)
	suggestions := []domain.CategorySuggestion{}

	// Check against known patterns
	for category, keywords := range categoryPatterns {
		matchCount := 0
		matchedKeywords := []string{}
		for _, keyword := range keywords {
			if strings.Contains(text, keyword) {
				matchCount++
				matchedKeywords = append(matchedKeywords, keyword)
			}
		}
		if matchCount > 0 {
			confidence := float64(matchCount) / float64(len(keywords))
			if confidence > 1.0 {
				confidence = 1.0
			}
			suggestions = append(suggestions, domain.CategorySuggestion{
				Category:        category,
				Confidence:      confidence,
				MatchedKeywords: matchedKeywords,
			})
		}
	}

	// Also check user's existing categories
	existingCategories, err := s.taskRepo.GetCategories(ctx, userID)
	if err != nil {
		slog.Warn("Failed to get existing categories for suggestions",
			"user_id", userID, "error", err)
		// Continue with pattern-based suggestions only
	} else {
		for _, existingCat := range existingCategories {
			catLower := strings.ToLower(existingCat)
			if strings.Contains(text, catLower) {
				// Check if already in suggestions
				found := false
				for i, s := range suggestions {
					if strings.EqualFold(s.Category, existingCat) {
						// Boost existing category matches, clamped to max 1.0
						suggestions[i].Confidence = math.Min(suggestions[i].Confidence+0.2, 1.0)
						found = true
						break
					}
				}
				if !found {
					suggestions = append(suggestions, domain.CategorySuggestion{
						Category:        existingCat,
						Confidence:      0.5,
						MatchedKeywords: []string{catLower},
					})
				}
			}
		}
	}

	// Sort by confidence
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].Confidence > suggestions[j].Confidence
	})

	// Limit to top 3
	if len(suggestions) > 3 {
		suggestions = suggestions[:3]
	}

	return &domain.CategorySuggestionResponse{
		Suggestions: suggestions,
	}, nil
}
