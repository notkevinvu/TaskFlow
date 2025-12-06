package domain

import (
	"errors"
	"time"
)

// Gamification errors
var (
	ErrGamificationStatsNotFound = errors.New("gamification stats not found")
	ErrInvalidTimezone           = errors.New("invalid timezone")
	ErrAchievementAlreadyEarned  = errors.New("achievement already earned")
)

// AchievementType represents different achievement categories
type AchievementType string

const (
	// Milestone achievements (task completion counts)
	AchievementFirstTask    AchievementType = "first_task"
	AchievementMilestone10  AchievementType = "milestone_10"
	AchievementMilestone50  AchievementType = "milestone_50"
	AchievementMilestone100 AchievementType = "milestone_100"

	// Streak achievements (consecutive days)
	AchievementStreak3  AchievementType = "streak_3"
	AchievementStreak7  AchievementType = "streak_7"
	AchievementStreak14 AchievementType = "streak_14"
	AchievementStreak30 AchievementType = "streak_30"

	// Category mastery (10 tasks in same category)
	AchievementCategoryMaster AchievementType = "category_master"

	// Speed achievements
	AchievementSpeedDemon AchievementType = "speed_demon"

	// Consistency achievements
	AchievementConsistencyKing AchievementType = "consistency_king"
)

// UserAchievement represents an earned achievement badge
type UserAchievement struct {
	ID              string                 `json:"id"`
	UserID          string                 `json:"user_id"`
	AchievementType AchievementType        `json:"achievement_type"`
	EarnedAt        time.Time              `json:"earned_at"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// AchievementDefinition contains display info for achievements
type AchievementDefinition struct {
	Type        AchievementType `json:"type"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Icon        string          `json:"icon"`     // Emoji for badge display
	Category    string          `json:"category"` // "milestone", "streak", "mastery", "speed", "consistency"
}

// GetAchievementDefinitions returns all available achievement definitions
func GetAchievementDefinitions() []AchievementDefinition {
	return []AchievementDefinition{
		// Milestone achievements
		{Type: AchievementFirstTask, Title: "First Step", Description: "Complete your first task", Icon: "🎯", Category: "milestone"},
		{Type: AchievementMilestone10, Title: "Getting Started", Description: "Complete 10 tasks", Icon: "🌟", Category: "milestone"},
		{Type: AchievementMilestone50, Title: "Task Master", Description: "Complete 50 tasks", Icon: "💪", Category: "milestone"},
		{Type: AchievementMilestone100, Title: "Centurion", Description: "Complete 100 tasks", Icon: "🏆", Category: "milestone"},

		// Streak achievements
		{Type: AchievementStreak3, Title: "On a Roll", Description: "Maintain a 3-day streak", Icon: "🔥", Category: "streak"},
		{Type: AchievementStreak7, Title: "Week Warrior", Description: "Maintain a 7-day streak", Icon: "⚡", Category: "streak"},
		{Type: AchievementStreak14, Title: "Fortnight Champion", Description: "Maintain a 14-day streak", Icon: "🚀", Category: "streak"},
		{Type: AchievementStreak30, Title: "Monthly Master", Description: "Maintain a 30-day streak", Icon: "👑", Category: "streak"},

		// Category mastery
		{Type: AchievementCategoryMaster, Title: "Category Expert", Description: "Complete 10 tasks in a single category", Icon: "⭐", Category: "mastery"},

		// Speed
		{Type: AchievementSpeedDemon, Title: "Speed Demon", Description: "Complete 5 tasks within 24 hours of creation", Icon: "💨", Category: "speed"},

		// Consistency
		{Type: AchievementConsistencyKing, Title: "Consistency King", Description: "Complete tasks on 5+ days in a week", Icon: "📅", Category: "consistency"},
	}
}

// GetAchievementDefinition returns the definition for a specific achievement type
func GetAchievementDefinition(achievementType AchievementType) *AchievementDefinition {
	for _, def := range GetAchievementDefinitions() {
		if def.Type == achievementType {
			return &def
		}
	}
	return nil
}

// GamificationStats contains cached productivity stats
type GamificationStats struct {
	UserID             string     `json:"user_id"`
	CurrentStreak      int        `json:"current_streak"`
	LongestStreak      int        `json:"longest_streak"`
	LastCompletionDate *string    `json:"last_completion_date,omitempty"` // YYYY-MM-DD in user timezone
	TotalCompleted     int        `json:"total_completed"`
	ProductivityScore  float64    `json:"productivity_score"` // 0-100
	CompletionRate     float64    `json:"completion_rate"`    // 0-100
	StreakScore        float64    `json:"streak_score"`       // 0-100
	OnTimePercentage   float64    `json:"on_time_percentage"` // 0-100
	EffortMixScore     float64    `json:"effort_mix_score"`   // 0-100
	LastComputedAt     time.Time  `json:"last_computed_at"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// CategoryMastery tracks progress towards category achievements
type CategoryMastery struct {
	ID              string     `json:"id"`
	UserID          string     `json:"user_id"`
	Category        string     `json:"category"`
	CompletedCount  int        `json:"completed_count"`
	LastCompletedAt *time.Time `json:"last_completed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// GamificationDashboard is the combined view for frontend
type GamificationDashboard struct {
	Stats                 *GamificationStats      `json:"stats"`
	RecentAchievements    []*UserAchievement      `json:"recent_achievements"`    // Last 5 earned
	AllAchievements       []*UserAchievement      `json:"all_achievements"`       // All earned
	AvailableAchievements []AchievementDefinition `json:"available_achievements"` // All possible
	CategoryProgress      []*CategoryMastery      `json:"category_progress"`      // Progress per category
	UnviewedCount         int                     `json:"unviewed_count"`         // New achievements since last view
}

// AchievementEarnedEvent represents a newly earned achievement for notifications
type AchievementEarnedEvent struct {
	Achievement *UserAchievement      `json:"achievement"`
	Definition  AchievementDefinition `json:"definition"`
}

// TaskCompletionGamificationResult contains achievements earned from completing a task
type TaskCompletionGamificationResult struct {
	UpdatedStats    *GamificationStats        `json:"updated_stats"`
	NewAchievements []*AchievementEarnedEvent `json:"new_achievements"`
	StreakExtended  bool                      `json:"streak_extended"`
	PreviousStreak  int                       `json:"previous_streak"`
}

// StreakCalculationResult holds the computed streak data
type StreakCalculationResult struct {
	CurrentStreak      int
	LongestStreak      int
	LastCompletionDate *string // YYYY-MM-DD format
	StreakExtended     bool    // True if current completion extended the streak
}

// ProductivityScoreWeights defines the weight factors for score calculation
var ProductivityScoreWeights = struct {
	CompletionRate float64
	Streak         float64
	OnTime         float64
	EffortMix      float64
}{
	CompletionRate: 0.30, // 30% weight
	Streak:         0.25, // 25% weight
	OnTime:         0.25, // 25% weight
	EffortMix:      0.20, // 20% weight
}

// MilestoneThresholds defines task counts for milestone achievements
var MilestoneThresholds = map[AchievementType]int{
	AchievementFirstTask:    1,
	AchievementMilestone10:  10,
	AchievementMilestone50:  50,
	AchievementMilestone100: 100,
}

// StreakThresholds defines day counts for streak achievements
var StreakThresholds = map[AchievementType]int{
	AchievementStreak3:  3,
	AchievementStreak7:  7,
	AchievementStreak14: 14,
	AchievementStreak30: 30,
}

// CategoryMasteryThreshold is the number of tasks needed for category mastery
const CategoryMasteryThreshold = 10

// SpeedDemonThreshold is the number of quick completions needed
const SpeedDemonThreshold = 5

// SpeedDemonDuration is the time limit for a "quick" completion
const SpeedDemonDuration = 24 * time.Hour

// ConsistencyKingThreshold is the number of active days per week needed
const ConsistencyKingThreshold = 5

// UpdateTimezoneDTO is used for updating user timezone
type UpdateTimezoneDTO struct {
	Timezone string `json:"timezone" binding:"required"`
}
