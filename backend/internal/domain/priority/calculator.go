package priority

import (
	"math"
	"time"

	"github.com/notkevinvu/taskflow/backend/internal/domain"
)

// Calculator handles task priority calculations
type Calculator struct{}

// NewCalculator creates a new priority calculator
func NewCalculator() *Calculator {
	return &Calculator{}
}

// Calculate computes the priority score for a task
// Formula: (UserPriority × 0.4 + TimeDecay × 0.3 + DeadlineUrgency × 0.2 + BumpPenalty × 0.1) × EffortBoost
func (calc *Calculator) Calculate(task *domain.Task) int {
	userPriority := float64(task.UserPriority)
	timeDecay := calc.calculateTimeDecay(task.CreatedAt)
	deadlineUrgency := calc.calculateDeadlineUrgency(task.DueDate)
	bumpPenalty := calc.calculateBumpPenalty(task.BumpCount)
	effortBoost := calc.getEffortBoost(task.EstimatedEffort)

	// Weighted sum
	score := (userPriority * 0.4) + (timeDecay * 0.3) + (deadlineUrgency * 0.2) + (bumpPenalty * 0.1)

	// Apply effort multiplier
	score = score * effortBoost

	// Clamp to 0-100 range
	return int(math.Min(100, math.Max(0, score)))
}

// calculateTimeDecay returns 0-100 based on task age
// Linear increase over 30 days: 0 days = 0, 30 days = 100
func (calc *Calculator) calculateTimeDecay(createdAt time.Time) float64 {
	age := time.Since(createdAt)
	days := age.Hours() / 24

	// Linear growth over 30 days
	decay := (days / 30.0) * 100

	// Cap at 100
	return math.Min(100, decay)
}

// calculateDeadlineUrgency returns 0-100 based on proximity to due date
// Quadratic increase in final 3 days
func (calc *Calculator) calculateDeadlineUrgency(dueDate *time.Time) float64 {
	if dueDate == nil {
		return 0 // No deadline = no urgency
	}

	now := time.Now()
	if now.After(*dueDate) {
		return 100 // Overdue = maximum urgency
	}

	daysRemaining := dueDate.Sub(now).Hours() / 24

	// No urgency if more than 7 days away
	if daysRemaining > 7 {
		return 0
	}

	// Quadratic urgency in final 7 days
	// Formula: 100 * (1 - (days/7)^2)
	urgency := 100 * (1 - math.Pow(daysRemaining/7, 2))

	return math.Max(0, urgency)
}

// calculateBumpPenalty returns 0-50 based on bump count
// +10 points per bump, capped at 50
func (calc *Calculator) calculateBumpPenalty(bumpCount int) float64 {
	penalty := float64(bumpCount) * 10
	return math.Min(50, penalty)
}

// getEffortBoost returns multiplier based on estimated effort
// Small tasks get 1.3x, large tasks get 1.0x
func (calc *Calculator) getEffortBoost(effort *domain.TaskEffort) float64 {
	if effort == nil {
		return 1.0 // No estimate = no boost
	}

	return effort.GetEffortMultiplier()
}

// IsAtRisk determines if a task is at risk
// Criteria: Bump count >= 3 OR overdue by >= 3 days
func (calc *Calculator) IsAtRisk(task *domain.Task) bool {
	// Check bump count
	if task.BumpCount >= 3 {
		return true
	}

	// Check if overdue by 3+ days
	if task.DueDate != nil {
		overdueDays := time.Since(*task.DueDate).Hours() / 24
		if overdueDays >= 3 {
			return true
		}
	}

	return false
}
