package domain

import "time"

// InsightType represents the type of insight/suggestion
type InsightType string

const (
	InsightAvoidancePattern   InsightType = "avoidance_pattern"
	InsightPeakPerformance    InsightType = "peak_performance"
	InsightQuickWins          InsightType = "quick_wins"
	InsightDeadlineClustering InsightType = "deadline_clustering"
	InsightAtRiskAlert        InsightType = "at_risk_alert"
	InsightCategoryOverload   InsightType = "category_overload"
)

// InsightPriority represents the importance level of an insight (1-5, higher = more important)
type InsightPriority int

const (
	InsightPriorityLow      InsightPriority = 1
	InsightPriorityMedium   InsightPriority = 2
	InsightPriorityHigh     InsightPriority = 3
	InsightPriorityUrgent   InsightPriority = 4
	InsightPriorityCritical InsightPriority = 5
)

// Insight represents a smart suggestion or observation about user's task patterns
type Insight struct {
	Type        InsightType            `json:"type"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Priority    InsightPriority        `json:"priority"`
	ActionURL   *string                `json:"action_url,omitempty"`
	Data        map[string]interface{} `json:"data,omitempty"`
	GeneratedAt time.Time              `json:"generated_at"`
}

// InsightResponse contains a list of insights for a user
type InsightResponse struct {
	Insights []Insight `json:"insights"`
	CachedAt time.Time `json:"cached_at"`
}

// TimeEstimate represents an estimated completion time for a task
type TimeEstimate struct {
	EstimatedDays   float64           `json:"estimated_days"`
	ConfidenceLevel string            `json:"confidence_level"` // "low", "medium", "high"
	BasedOn         int               `json:"based_on"`         // Number of similar tasks used for estimation
	Factors         TimeEstimateFactor `json:"factors"`
}

// TimeEstimateFactor contains the individual factors used in time estimation
type TimeEstimateFactor struct {
	BaseEstimate   float64 `json:"base_estimate"`
	CategoryFactor float64 `json:"category_factor"`
	EffortFactor   float64 `json:"effort_factor"`
	BumpFactor     float64 `json:"bump_factor"`
}

// CategorySuggestion represents a suggested category based on task content
type CategorySuggestion struct {
	Category   string  `json:"category"`
	Confidence float64 `json:"confidence"` // 0.0 to 1.0
	MatchedKeywords []string `json:"matched_keywords,omitempty"`
}

// CategorySuggestionRequest is the input for category suggestion
type CategorySuggestionRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

// CategorySuggestionResponse contains category suggestions
type CategorySuggestionResponse struct {
	Suggestions []CategorySuggestion `json:"suggestions"`
}

// Analytics query result types for insights

// CategoryBumpStats contains bump statistics for a category
type CategoryBumpStats struct {
	Category   string  `json:"category"`
	AvgBumps   float64 `json:"avg_bumps"`
	TaskCount  int     `json:"task_count"`
}

// DayOfWeekStats contains completion statistics for a day of the week
type DayOfWeekStats struct {
	DayOfWeek      int `json:"day_of_week"` // 0 = Sunday, 6 = Saturday
	CompletedCount int `json:"completed_count"`
}

// DeadlineCluster represents a group of tasks due on the same date
type DeadlineCluster struct {
	DueDate   time.Time `json:"due_date"`
	TaskCount int       `json:"task_count"`
	Titles    []string  `json:"titles"`
}

// CompletionTimeStats contains statistics for task completion times
type CompletionTimeStats struct {
	SampleSize     int     `json:"sample_size"`
	MedianDays     float64 `json:"median_days"`
	AvgDays        float64 `json:"avg_days"`
	CategoryMedian float64 `json:"category_median"`
}

// CategoryDistribution contains the distribution of pending tasks by category
type CategoryDistribution struct {
	Category   string  `json:"category"`
	TaskCount  int     `json:"task_count"`
	Percentage float64 `json:"percentage"`
}

// HeatmapCell represents a single cell in the productivity heatmap
type HeatmapCell struct {
	DayOfWeek int `json:"day_of_week"` // 0 = Sunday, 6 = Saturday
	Hour      int `json:"hour"`        // 0-23
	Count     int `json:"count"`       // Number of completions
}

// ProductivityHeatmap contains heatmap data for productivity visualization
type ProductivityHeatmap struct {
	Cells    []HeatmapCell `json:"cells"`
	MaxCount int           `json:"max_count"` // For color scaling
}

// CategoryTrendPoint represents a single point in category trends over time
type CategoryTrendPoint struct {
	WeekStart  string         `json:"week_start"`  // ISO date string (YYYY-MM-DD)
	Categories map[string]int `json:"categories"`  // Category -> count mapping
}

// CategoryTrends contains weekly category breakdown for trend visualization
type CategoryTrends struct {
	Weeks      []CategoryTrendPoint `json:"weeks"`
	Categories []string             `json:"categories"` // All unique categories for legend
}
