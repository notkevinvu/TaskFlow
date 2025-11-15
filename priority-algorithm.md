# Priority Algorithm Specification
## Intelligent Task Prioritization System

**Version:** 1.0
**Last Updated:** 2025-01-15

---

## Overview

The priority algorithm calculates a composite score (0-100) for each task, combining multiple factors to surface the most important work at the right time.

**Design Goals:**
1. **Respect user intent:** User-set priority carries the most weight
2. **Time awareness:** Old tasks shouldn't languish forever
3. **Deadline sensitivity:** Approaching deadlines create urgency
4. **Penalize procrastination:** Repeatedly bumped tasks escalate
5. **Favor quick wins:** Small tasks get slight boost

---

## Priority Score Formula

```
PriorityScore = (
    UserPriority × 0.4 +
    TimeDecay × 0.3 +
    DeadlineUrgency × 0.2 +
    BumpPenalty × 0.1
) × EffortBoost
```

**Output Range:** 0-100 (capped)
**Recalculation Frequency:** Every 6 hours (background job) or on-demand

---

## Component Calculations

### 1. UserPriority (0-100, Weight: 0.4)

**Source:** User-set importance level

**Mapping:**

| User Level | Value | Description |
|------------|-------|-------------|
| **Low** | 25 | "Nice to have" |
| **Medium** | 50 | Default priority |
| **High** | 75 | Important work |
| **Critical** | 100 | Urgent, must do soon |

**Calculation:**

```go
func calculateUserPriority(task *Task) float64 {
    return float64(task.UserPriority)
}
```

**Example:**
```
Task with High priority:
UserPriority = 75
```

---

### 2. TimeDecay (0-100, Weight: 0.3)

**Source:** Task age (time since creation)

**Purpose:** Prevent old tasks from being forgotten

**Formula:**

```
TimeDecay = min(100, (DaysSinceCreation / 30) × 100)
```

**Decay Curve:**
- 0 days → 0 points
- 7 days → 23 points
- 15 days → 50 points
- 30 days → 100 points (max)
- 60 days → 100 points (capped)

**Implementation:**

```go
func calculateTimeDecay(task *Task) float64 {
    daysSinceCreation := time.Since(task.CreatedAt).Hours() / 24.0

    decay := (daysSinceCreation / 30.0) * 100.0

    // Cap at 100
    if decay > 100 {
        return 100
    }

    return decay
}
```

**Example:**
```
Task created 10 days ago:
TimeDecay = (10 / 30) × 100 = 33.33
```

**Rationale:** 30-day window ensures tasks become urgent within a month if not addressed.

---

### 3. DeadlineUrgency (0-100, Weight: 0.2)

**Source:** Days until due date

**Purpose:** Create exponential urgency as deadline approaches

**Formula:**

```
If no due_date:
    DeadlineUrgency = 0

If overdue:
    DeadlineUrgency = 100

If due within 30 days:
    DaysUntilDue = (due_date - now) in days
    DeadlineUrgency = 100 × (1 - (DaysUntilDue / 30)^2)
```

**Urgency Curve (Quadratic):**

| Days Until Due | Urgency |
|----------------|---------|
| 30+ days | 0 |
| 21 days | 19 |
| 14 days | 36 |
| 7 days | 64 |
| 3 days | 81 |
| 1 day | 96 |
| Overdue | 100 |

**Implementation:**

```go
func calculateDeadlineUrgency(task *Task) float64 {
    // No deadline = no urgency
    if task.DueDate == nil {
        return 0
    }

    daysUntil := task.DueDate.Sub(time.Now()).Hours() / 24.0

    // Overdue
    if daysUntil <= 0 {
        return 100
    }

    // Due beyond 30 days = low urgency
    if daysUntil > 30 {
        return 0
    }

    // Quadratic increase as deadline approaches
    urgency := 100.0 * (1.0 - math.Pow(daysUntil/30.0, 2.0))

    return urgency
}
```

**Example:**
```
Task due in 5 days:
DeadlineUrgency = 100 × (1 - (5 / 30)^2)
                = 100 × (1 - 0.0278)
                = 97.22
```

**Rationale:** Quadratic curve creates sharp urgency in final days while keeping distant deadlines low-priority.

---

### 4. BumpPenalty (0-50, Weight: 0.1)

**Source:** Number of times task was bumped (delayed)

**Purpose:** Escalate tasks that are chronically avoided

**Formula:**

```
BumpPenalty = min(50, BumpCount × 10)
```

**Penalty Curve:**

| Bump Count | Penalty |
|------------|---------|
| 0 | 0 |
| 1 | 10 |
| 2 | 20 |
| 3 | 30 (at-risk) |
| 5+ | 50 (max) |

**Implementation:**

```go
func calculateBumpPenalty(task *Task) float64 {
    penalty := float64(task.BumpCount) * 10.0

    // Cap at 50
    if penalty > 50 {
        return 50
    }

    return penalty
}
```

**Example:**
```
Task bumped 3 times:
BumpPenalty = 3 × 10 = 30
```

**Rationale:** 3 bumps triggers "at-risk" flag. After 5 bumps, task is maximally penalized.

---

### 5. EffortBoost (1.0-1.3 multiplier)

**Source:** Estimated effort

**Purpose:** Give slight boost to small tasks (quick wins)

**Formula:**

```
EffortBoost = {
    1.3   if effort == "small"   (< 1 hour)
    1.15  if effort == "medium"  (1-4 hours)
    1.0   if effort == "large"   (4-8 hours)
    1.0   if effort == "xlarge"  (> 8 hours)
}
```

**Implementation:**

```go
func calculateEffortBoost(task *Task) float64 {
    switch task.EstimatedEffort {
    case "small":
        return 1.3
    case "medium":
        return 1.15
    case "large":
        return 1.0
    case "xlarge":
        return 1.0
    default:
        return 1.0 // No estimate
    }
}
```

**Example:**
```
Small task (< 1 hour):
EffortBoost = 1.3

Large task (4-8 hours):
EffortBoost = 1.0
```

**Rationale:** Small boost encourages completing quick tasks, but doesn't overpower other factors.

---

## Complete Calculation Example

### Example 1: New High-Priority Task

**Task Attributes:**
- User Priority: High (75)
- Created: 2 days ago
- Due Date: In 10 days
- Bump Count: 0
- Estimated Effort: Medium

**Step-by-Step:**

```
UserPriority = 75

TimeDecay = (2 / 30) × 100 = 6.67

DeadlineUrgency = 100 × (1 - (10 / 30)^2)
                = 100 × (1 - 0.111)
                = 88.9

BumpPenalty = 0 × 10 = 0

EffortBoost = 1.15 (medium)

PriorityScore = (75 × 0.4 + 6.67 × 0.3 + 88.9 × 0.2 + 0 × 0.1) × 1.15
              = (30 + 2.0 + 17.78 + 0) × 1.15
              = 49.78 × 1.15
              = 57.25
```

**Final Score: 57.25**

---

### Example 2: Old Low-Priority Task (Bumped Multiple Times)

**Task Attributes:**
- User Priority: Low (25)
- Created: 45 days ago
- Due Date: None
- Bump Count: 4
- Estimated Effort: Small

**Step-by-Step:**

```
UserPriority = 25

TimeDecay = (45 / 30) × 100 = 150 → capped at 100

DeadlineUrgency = 0 (no deadline)

BumpPenalty = 4 × 10 = 40

EffortBoost = 1.3 (small)

PriorityScore = (25 × 0.4 + 100 × 0.3 + 0 × 0.2 + 40 × 0.1) × 1.3
              = (10 + 30 + 0 + 4) × 1.3
              = 44 × 1.3
              = 57.2
```

**Final Score: 57.2**

**Insight:** Despite low user priority, age + bumps + small effort bring it to similar priority as Example 1.

---

### Example 3: Critical Task Due Tomorrow

**Task Attributes:**
- User Priority: Critical (100)
- Created: 1 day ago
- Due Date: Tomorrow (1 day away)
- Bump Count: 0
- Estimated Effort: Large

**Step-by-Step:**

```
UserPriority = 100

TimeDecay = (1 / 30) × 100 = 3.33

DeadlineUrgency = 100 × (1 - (1 / 30)^2)
                = 100 × (1 - 0.00111)
                = 99.89

BumpPenalty = 0

EffortBoost = 1.0 (large)

PriorityScore = (100 × 0.4 + 3.33 × 0.3 + 99.89 × 0.2 + 0 × 0.1) × 1.0
              = (40 + 1.0 + 19.98 + 0) × 1.0
              = 60.98
```

**Final Score: 60.98**

**Note:** Even critical tasks don't hit 100 unless all factors align.

---

## Edge Cases

### Edge Case 1: All Max Values

**Task Attributes:**
- User Priority: Critical (100)
- Created: 60 days ago
- Due Date: Overdue
- Bump Count: 10
- Estimated Effort: Small

```
PriorityScore = (100 × 0.4 + 100 × 0.3 + 100 × 0.2 + 50 × 0.1) × 1.3
              = (40 + 30 + 20 + 5) × 1.3
              = 95 × 1.3
              = 123.5 → capped at 100
```

**Final Score: 100 (capped)**

**Handling:** Cap all scores at 100 to maintain consistent range.

---

### Edge Case 2: Minimal Task

**Task Attributes:**
- User Priority: Low (25)
- Created: Today
- Due Date: None
- Bump Count: 0
- Estimated Effort: None

```
PriorityScore = (25 × 0.4 + 0 × 0.3 + 0 × 0.2 + 0 × 0.1) × 1.0
              = 10 × 1.0
              = 10
```

**Final Score: 10**

---

### Edge Case 3: Negative Days Until Due (Overdue)

**Handling:**
```go
if daysUntil <= 0 {
    return 100  // Max urgency
}
```

---

## Implementation

### Go Service Implementation

```go
package services

import (
    "math"
    "time"

    "github.com/yourusername/webapp/internal/domain"
)

type PriorityCalculator struct{}

func NewPriorityCalculator() *PriorityCalculator {
    return &PriorityCalculator{}
}

func (pc *PriorityCalculator) CalculatePriority(task *domain.Task) float64 {
    userPriority := float64(task.UserPriority)

    timeDecay := pc.calculateTimeDecay(task)

    deadlineUrgency := pc.calculateDeadlineUrgency(task)

    bumpPenalty := float64(task.BumpCount) * 10.0
    if bumpPenalty > 50 {
        bumpPenalty = 50
    }

    effortBoost := pc.calculateEffortBoost(task)

    // Composite score
    score := (
        userPriority*0.4 +
        timeDecay*0.3 +
        deadlineUrgency*0.2 +
        bumpPenalty*0.1,
    ) * effortBoost

    // Cap at 100
    if score > 100 {
        return 100
    }

    return score
}

func (pc *PriorityCalculator) calculateTimeDecay(task *domain.Task) float64 {
    daysSinceCreation := time.Since(task.CreatedAt).Hours() / 24.0
    decay := (daysSinceCreation / 30.0) * 100.0

    if decay > 100 {
        return 100
    }

    return decay
}

func (pc *PriorityCalculator) calculateDeadlineUrgency(task *domain.Task) float64 {
    if task.DueDate == nil {
        return 0
    }

    daysUntil := task.DueDate.Sub(time.Now()).Hours() / 24.0

    if daysUntil <= 0 {
        return 100
    }

    if daysUntil > 30 {
        return 0
    }

    return 100.0 * (1.0 - math.Pow(daysUntil/30.0, 2.0))
}

func (pc *PriorityCalculator) calculateEffortBoost(task *domain.Task) float64 {
    switch task.EstimatedEffort {
    case "small":
        return 1.3
    case "medium":
        return 1.15
    case "large", "xlarge":
        return 1.0
    default:
        return 1.0
    }
}
```

---

### Background Job for Recalculation

```go
package worker

import (
    "context"
    "log"
    "time"

    "github.com/yourusername/webapp/internal/ports"
    "github.com/yourusername/webapp/internal/services"
)

type PriorityRecalculationWorker struct {
    taskRepo    ports.TaskRepository
    calculator  *services.PriorityCalculator
}

func (w *PriorityRecalculationWorker) Run(ctx context.Context) {
    ticker := time.NewTicker(6 * time.Hour)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            if err := w.recalculateAllPriorities(ctx); err != nil {
                log.Printf("Priority recalculation failed: %v", err)
            }
        case <-ctx.Done():
            return
        }
    }
}

func (w *PriorityRecalculationWorker) recalculateAllPriorities(ctx context.Context) error {
    // Get all active tasks (status != done, deleted_at IS NULL)
    tasks, err := w.taskRepo.GetActiveTasks(ctx)
    if err != nil {
        return err
    }

    for _, task := range tasks {
        oldPriority := task.PriorityScore
        newPriority := w.calculator.CalculatePriority(task)

        // Only update if priority changed significantly (> 5 points)
        if math.Abs(newPriority - oldPriority) > 5 {
            task.PriorityScore = newPriority
            task.PriorityLastCalculatedAt = time.Now()

            if err := w.taskRepo.UpdatePriority(ctx, task); err != nil {
                log.Printf("Failed to update priority for task %s: %v", task.ID, err)
                continue
            }

            // Log history
            // ... record priority change event
        }
    }

    log.Printf("Recalculated priorities for %d tasks", len(tasks))
    return nil
}
```

---

## Tuning & Customization

### Weight Adjustments

Current weights can be tuned based on user feedback:

| Component | Current Weight | Rationale |
|-----------|---------------|-----------|
| UserPriority | 0.4 | Strongest signal - user knows best |
| TimeDecay | 0.3 | Important to prevent forgotten tasks |
| DeadlineUrgency | 0.2 | Deadlines matter but not everything |
| BumpPenalty | 0.1 | Nudge, not force |

**Potential Adjustments:**
- Increase TimeDecay to 0.4 if users report old tasks languishing
- Increase BumpPenalty to 0.15 if procrastination is common

### User Preferences

Future enhancement: Allow users to customize weights in settings.

```go
type UserPreferences struct {
    UserPriorityWeight    float64 `json:"user_priority_weight"`
    TimeDecayWeight       float64 `json:"time_decay_weight"`
    DeadlineUrgencyWeight float64 `json:"deadline_urgency_weight"`
    BumpPenaltyWeight     float64 `json:"bump_penalty_weight"`
}
```

---

## Testing Strategy

### Unit Tests

Test each component in isolation:

```go
func TestCalculateTimeDecay(t *testing.T) {
    tests := []struct {
        name string
        age  time.Duration
        want float64
    }{
        {"New task", 0, 0},
        {"1 week old", 7 * 24 * time.Hour, 23.33},
        {"30 days old", 30 * 24 * time.Hour, 100},
        {"60 days old (capped)", 60 * 24 * time.Hour, 100},
    }

    pc := NewPriorityCalculator()

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            task := &domain.Task{
                CreatedAt: time.Now().Add(-tt.age),
            }

            got := pc.calculateTimeDecay(task)

            if math.Abs(got-tt.want) > 0.5 {
                t.Errorf("got %f, want %f", got, tt.want)
            }
        })
    }
}
```

### Integration Tests

Test complete priority calculation:

```go
func TestCalculatePriority_HighPriorityDueTomorrow(t *testing.T) {
    pc := NewPriorityCalculator()

    tomorrow := time.Now().Add(24 * time.Hour)
    task := &domain.Task{
        UserPriority:    75, // High
        CreatedAt:       time.Now().Add(-48 * time.Hour), // 2 days ago
        DueDate:         &tomorrow,
        BumpCount:       0,
        EstimatedEffort: "medium",
    }

    score := pc.CalculatePriority(task)

    // Expected: ~58-65 range
    if score < 55 || score > 70 {
        t.Errorf("Unexpected priority score: %f", score)
    }
}
```

---

## Future Enhancements

### Machine Learning

After collecting 6+ months of data:
- Train ML model to predict completion probability
- Adjust weights based on user behavior
- Personalized priority algorithms per user

### Context-Aware Priority

- Time of day (morning vs evening)
- Day of week (Monday vs Friday)
- Current workload (many high-priority tasks → increase threshold)

### Smart Bump Detection

Instead of simple counter, analyze bump patterns:
- Bumped at same time every day → suggest scheduling
- Bumped with same category → suggest delegation

---

## Related Documents

- `PRD.md` - Product requirements
- `data-model.md` - Database schema
- `phase-2-weeks-3-4.md` - Implementation guide

---

**Document Status:** Ready for Implementation
**Last Review:** 2025-01-15
