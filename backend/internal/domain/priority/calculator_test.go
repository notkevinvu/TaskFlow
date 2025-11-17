package priority

import (
	"testing"
	"time"

	"github.com/notkevinvu/taskflow/backend/internal/domain"
)

func TestCalculate(t *testing.T) {
	calc := NewCalculator()

	tests := []struct {
		name     string
		task     *domain.Task
		expected int
		desc     string
	}{
		{
			name: "High priority, no decay, no deadline, no bumps",
			task: &domain.Task{
				UserPriority: 100,
				CreatedAt:    time.Now(),
				BumpCount:    0,
			},
			expected: 40, // 100 * 0.4 + 0 + 0 + 0 = 40
			desc:     "Only user priority contributes",
		},
		{
			name: "Medium priority, 30 days old, no deadline, no bumps",
			task: &domain.Task{
				UserPriority: 50,
				CreatedAt:    time.Now().AddDate(0, 0, -30),
				BumpCount:    0,
			},
			expected: 50, // 50*0.4 + 100*0.3 + 0 + 0 = 20 + 30 = 50
			desc:     "User priority + time decay",
		},
		{
			name: "Low priority, new, due tomorrow, no bumps",
			task: &domain.Task{
				UserPriority: 25,
				CreatedAt:    time.Now(),
				DueDate:      timePtr(time.Now().AddDate(0, 0, 1)),
				BumpCount:    0,
			},
			expected: 28, // 25*0.4 + 0 + ~80*0.2 + 0 ≈ 10 + 16 = 26-28
			desc:     "User priority + deadline urgency",
		},
		{
			name: "Medium priority, new, no deadline, 5 bumps",
			task: &domain.Task{
				UserPriority: 50,
				CreatedAt:    time.Now(),
				BumpCount:    5,
			},
			expected: 25, // 50*0.4 + 0 + 0 + 50*0.1 = 20 + 5 = 25
			desc:     "User priority + bump penalty",
		},
		{
			name: "High priority, small task, no other factors",
			task: &domain.Task{
				UserPriority:    75,
				CreatedAt:       time.Now(),
				EstimatedEffort: effortPtr(domain.TaskEffortSmall),
				BumpCount:       0,
			},
			expected: 39, // (75*0.4 + 0 + 0 + 0) * 1.3 = 30 * 1.3 = 39
			desc:     "User priority with small task boost",
		},
		{
			name: "Overdue task",
			task: &domain.Task{
				UserPriority: 50,
				CreatedAt:    time.Now().AddDate(0, 0, -10),
				DueDate:      timePtr(time.Now().AddDate(0, 0, -5)),
				BumpCount:    0,
			},
			expected: 53, // 50*0.4 + ~33*0.3 + 100*0.2 + 0 = 20 + 10 + 20 = 50+
			desc:     "Overdue tasks get max deadline urgency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.Calculate(tt.task)
			// Allow ±2 points tolerance due to time-based calculations
			if abs(result-tt.expected) > 2 {
				t.Errorf("%s: expected ~%d, got %d", tt.desc, tt.expected, result)
			}
		})
	}
}

func TestCalculateTimeDecay(t *testing.T) {
	calc := NewCalculator()

	tests := []struct {
		name      string
		createdAt time.Time
		expected  float64
	}{
		{
			name:      "Brand new task",
			createdAt: time.Now(),
			expected:  0,
		},
		{
			name:      "15 days old",
			createdAt: time.Now().AddDate(0, 0, -15),
			expected:  50,
		},
		{
			name:      "30 days old",
			createdAt: time.Now().AddDate(0, 0, -30),
			expected:  100,
		},
		{
			name:      "60 days old (capped)",
			createdAt: time.Now().AddDate(0, 0, -60),
			expected:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.calculateTimeDecay(tt.createdAt)
			if abs(int(result)-int(tt.expected)) > 2 {
				t.Errorf("expected ~%.0f, got %.0f", tt.expected, result)
			}
		})
	}
}

func TestCalculateDeadlineUrgency(t *testing.T) {
	calc := NewCalculator()

	tests := []struct {
		name     string
		dueDate  *time.Time
		expected float64
	}{
		{
			name:     "No deadline",
			dueDate:  nil,
			expected: 0,
		},
		{
			name:     "Due in 10 days",
			dueDate:  timePtr(time.Now().AddDate(0, 0, 10)),
			expected: 0,
		},
		{
			name:     "Due in 7 days",
			dueDate:  timePtr(time.Now().AddDate(0, 0, 7)),
			expected: 0,
		},
		{
			name:     "Due in 3 days",
			dueDate:  timePtr(time.Now().AddDate(0, 0, 3)),
			expected: 63, // ~63%
		},
		{
			name:     "Due in 1 day",
			dueDate:  timePtr(time.Now().AddDate(0, 0, 1)),
			expected: 98, // ~98%
		},
		{
			name:     "Overdue",
			dueDate:  timePtr(time.Now().AddDate(0, 0, -1)),
			expected: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.calculateDeadlineUrgency(tt.dueDate)
			if abs(int(result)-int(tt.expected)) > 3 {
				t.Errorf("expected ~%.0f, got %.0f", tt.expected, result)
			}
		})
	}
}

func TestCalculateBumpPenalty(t *testing.T) {
	calc := NewCalculator()

	tests := []struct {
		name      string
		bumpCount int
		expected  float64
	}{
		{
			name:      "No bumps",
			bumpCount: 0,
			expected:  0,
		},
		{
			name:      "1 bump",
			bumpCount: 1,
			expected:  10,
		},
		{
			name:      "3 bumps",
			bumpCount: 3,
			expected:  30,
		},
		{
			name:      "5 bumps (capped)",
			bumpCount: 5,
			expected:  50,
		},
		{
			name:      "10 bumps (capped)",
			bumpCount: 10,
			expected:  50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.calculateBumpPenalty(tt.bumpCount)
			if result != tt.expected {
				t.Errorf("expected %.0f, got %.0f", tt.expected, result)
			}
		})
	}
}

func TestGetEffortBoost(t *testing.T) {
	calc := NewCalculator()

	tests := []struct {
		name     string
		effort   *domain.TaskEffort
		expected float64
	}{
		{
			name:     "No estimate",
			effort:   nil,
			expected: 1.0,
		},
		{
			name:     "Small task",
			effort:   effortPtr(domain.TaskEffortSmall),
			expected: 1.3,
		},
		{
			name:     "Medium task",
			effort:   effortPtr(domain.TaskEffortMedium),
			expected: 1.15,
		},
		{
			name:     "Large task",
			effort:   effortPtr(domain.TaskEffortLarge),
			expected: 1.0,
		},
		{
			name:     "XLarge task",
			effort:   effortPtr(domain.TaskEffortXLarge),
			expected: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.getEffortBoost(tt.effort)
			if result != tt.expected {
				t.Errorf("expected %.2f, got %.2f", tt.expected, result)
			}
		})
	}
}

func TestIsAtRisk(t *testing.T) {
	calc := NewCalculator()

	tests := []struct {
		name     string
		task     *domain.Task
		expected bool
	}{
		{
			name: "No risk",
			task: &domain.Task{
				BumpCount: 0,
			},
			expected: false,
		},
		{
			name: "At risk due to bumps",
			task: &domain.Task{
				BumpCount: 3,
			},
			expected: true,
		},
		{
			name: "At risk due to overdue",
			task: &domain.Task{
				BumpCount: 0,
				DueDate:   timePtr(time.Now().AddDate(0, 0, -3)),
			},
			expected: true,
		},
		{
			name: "Not at risk - only 2 bumps",
			task: &domain.Task{
				BumpCount: 2,
			},
			expected: false,
		},
		{
			name: "Not at risk - only 2 days overdue",
			task: &domain.Task{
				BumpCount: 0,
				DueDate:   timePtr(time.Now().AddDate(0, 0, -2)),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calc.IsAtRisk(tt.task)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

// Helper functions
func timePtr(t time.Time) *time.Time {
	return &t
}

func effortPtr(e domain.TaskEffort) *domain.TaskEffort {
	return &e
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
