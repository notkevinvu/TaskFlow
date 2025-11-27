package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/notkevinvu/taskflow/backend/internal/middleware"
	"github.com/notkevinvu/taskflow/backend/internal/repository"
)

// AnalyticsHandler handles HTTP requests for analytics
type AnalyticsHandler struct {
	taskRepo *repository.TaskRepository
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(taskRepo *repository.TaskRepository) *AnalyticsHandler {
	return &AnalyticsHandler{taskRepo: taskRepo}
}

// GetSummary returns a comprehensive analytics summary
// GET /api/v1/analytics/summary?days=30
func (h *AnalyticsHandler) GetSummary(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse days parameter (default to 30 days)
	daysBack := 30
	if daysStr := c.Query("days"); daysStr != "" {
		if days, err := strconv.Atoi(daysStr); err == nil && days > 0 && days <= 365 {
			daysBack = days
		}
	}

	// Get completion stats
	completionStats, err := h.taskRepo.GetCompletionStats(c.Request.Context(), userID, daysBack)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch completion stats", "details": err.Error()})
		return
	}

	// Get bump analytics
	bumpAnalyticsRaw, err := h.taskRepo.GetBumpAnalytics(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch bump analytics", "details": err.Error()})
		return
	}

	// Transform bump analytics to match frontend expectations
	bumpDistribution := make(map[string]int)
	bumpDistribution["0 bumps"] = 0
	bumpDistribution["1-2 bumps"] = 0
	bumpDistribution["3-5 bumps"] = 0
	bumpDistribution["6+ bumps"] = 0

	for bumpCount, taskCount := range bumpAnalyticsRaw.TasksByBumpCount {
		if bumpCount == 0 {
			bumpDistribution["0 bumps"] += taskCount
		} else if bumpCount >= 1 && bumpCount <= 2 {
			bumpDistribution["1-2 bumps"] += taskCount
		} else if bumpCount >= 3 && bumpCount <= 5 {
			bumpDistribution["3-5 bumps"] += taskCount
		} else {
			bumpDistribution["6+ bumps"] += taskCount
		}
	}

	bumpAnalytics := gin.H{
		"average_bump_count": bumpAnalyticsRaw.AverageBumpCount,
		"at_risk_count":      bumpAnalyticsRaw.AtRiskCount,
		"bump_distribution":  bumpDistribution,
	}

	// Get category breakdown
	categoryStats, err := h.taskRepo.GetCategoryBreakdown(c.Request.Context(), userID, daysBack)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch category stats", "details": err.Error()})
		return
	}

	// Get priority distribution
	priorityDist, err := h.taskRepo.GetPriorityDistribution(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch priority distribution", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"period_days":          daysBack,
		"completion_stats":     completionStats,
		"bump_analytics":       bumpAnalytics,
		"category_breakdown":   categoryStats,
		"priority_distribution": priorityDist,
	})
}

// GetTrends returns task completion trends over time
// GET /api/v1/analytics/trends?days=30
func (h *AnalyticsHandler) GetTrends(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Parse days parameter (default to 30 days)
	daysBack := 30
	if daysStr := c.Query("days"); daysStr != "" {
		if days, err := strconv.Atoi(daysStr); err == nil && days > 0 && days <= 365 {
			daysBack = days
		}
	}

	// Get velocity metrics
	velocityMetrics, err := h.taskRepo.GetVelocityMetrics(c.Request.Context(), userID, daysBack)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch velocity metrics", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"period_days":      daysBack,
		"velocity_metrics": velocityMetrics,
	})
}
