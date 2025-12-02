package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/middleware"
	"github.com/notkevinvu/taskflow/backend/internal/ports"
)

// AnalyticsHandler handles HTTP requests for analytics
type AnalyticsHandler struct {
	taskRepo ports.TaskRepository
}

// NewAnalyticsHandler creates a new analytics handler
func NewAnalyticsHandler(taskRepo ports.TaskRepository) *AnalyticsHandler {
	return &AnalyticsHandler{taskRepo: taskRepo}
}

// GetSummary returns a comprehensive analytics summary
// GET /api/v1/analytics/summary?days=30
func (h *AnalyticsHandler) GetSummary(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
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
		middleware.AbortWithError(c, domain.NewInternalError("failed to fetch completion stats", err))
		return
	}

	// Get bump analytics
	bumpAnalyticsRaw, err := h.taskRepo.GetBumpAnalytics(c.Request.Context(), userID)
	if err != nil {
		middleware.AbortWithError(c, domain.NewInternalError("failed to fetch bump analytics", err))
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
		middleware.AbortWithError(c, domain.NewInternalError("failed to fetch category stats", err))
		return
	}

	// Get priority distribution
	priorityDist, err := h.taskRepo.GetPriorityDistribution(c.Request.Context(), userID)
	if err != nil {
		middleware.AbortWithError(c, domain.NewInternalError("failed to fetch priority distribution", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"period_days":           daysBack,
		"completion_stats":      completionStats,
		"bump_analytics":        bumpAnalytics,
		"category_breakdown":    categoryStats,
		"priority_distribution": priorityDist,
	})
}

// GetTrends returns task completion trends over time
// GET /api/v1/analytics/trends?days=30
func (h *AnalyticsHandler) GetTrends(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
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
		middleware.AbortWithError(c, domain.NewInternalError("failed to fetch velocity metrics", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"period_days":      daysBack,
		"velocity_metrics": velocityMetrics,
	})
}

// GetProductivityHeatmap returns completion counts by day of week and hour
// GET /api/v1/analytics/heatmap?days=90
func (h *AnalyticsHandler) GetProductivityHeatmap(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	// Parse days parameter (default to 90 days for meaningful heatmap data)
	daysBack := 90
	if daysStr := c.Query("days"); daysStr != "" {
		if days, err := strconv.Atoi(daysStr); err == nil && days > 0 && days <= 365 {
			daysBack = days
		}
	}

	heatmap, err := h.taskRepo.GetProductivityHeatmap(c.Request.Context(), userID, daysBack)
	if err != nil {
		middleware.AbortWithError(c, domain.NewInternalError("failed to fetch productivity heatmap", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"period_days": daysBack,
		"heatmap":     heatmap,
	})
}

// GetCategoryTrends returns weekly category breakdown for trend visualization
// GET /api/v1/analytics/category-trends?days=90
func (h *AnalyticsHandler) GetCategoryTrends(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		middleware.AbortWithError(c, domain.NewUnauthorizedError("user not authenticated"))
		return
	}

	// Parse days parameter (default to 90 days for trend visualization)
	daysBack := 90
	if daysStr := c.Query("days"); daysStr != "" {
		if days, err := strconv.Atoi(daysStr); err == nil && days > 0 && days <= 365 {
			daysBack = days
		}
	}

	trends, err := h.taskRepo.GetCategoryTrends(c.Request.Context(), userID, daysBack)
	if err != nil {
		middleware.AbortWithError(c, domain.NewInternalError("failed to fetch category trends", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"period_days": daysBack,
		"trends":      trends,
	})
}
