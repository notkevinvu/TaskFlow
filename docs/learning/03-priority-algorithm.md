# Module 03: Priority Algorithm

## Learning Objectives

By the end of this module, you will:
- Understand multi-factor priority scoring systems
- Learn how to design explainable algorithms (not black boxes)
- See how business rules translate to code
- Calculate priority scores by hand

---

## The Priority Problem

Most task managers let users set priority manually (High/Medium/Low). This has problems:

1. **Everything becomes "High"** - No differentiation
2. **No time awareness** - Old tasks get buried
3. **Deadline blindness** - Urgency isn't automatic
4. **Procrastination hides** - Repeated delays are invisible

TaskFlow solves this with an **automatic, multi-factor algorithm**.

---

## The Formula

```
Score = (UserPriority × 0.4 + TimeDecay × 0.3 + DeadlineUrgency × 0.2 + BumpPenalty × 0.1) × EffortBoost
```

### Factor Breakdown

| Factor | Weight | Range | Purpose |
|--------|--------|-------|---------|
| **User Priority** | 40% | 0-100 | Respects user's explicit importance rating |
| **Time Decay** | 30% | 0-100 | Prevents old tasks from being forgotten |
| **Deadline Urgency** | 20% | 0-100 | Creates urgency as deadlines approach |
| **Bump Penalty** | 10% | 0-50 | Exposes procrastinated tasks |
| **Effort Boost** | 1.0-1.3x | Multiplier | Encourages completing small tasks |

---

## Factor 1: User Priority (40%)

The user rates importance from 1-10. We scale to 0-100:

```go
// User rates 1-10, we scale to 0-100
userPriority := float64(task.UserPriority * 10)
// 1 → 10, 5 → 50, 10 → 100
```

**Why 40%?** User intent is the strongest signal. They know their context better than any algorithm.

### Example Calculations

| User Rating | Scaled Value | Weighted (×0.4) |
|-------------|--------------|-----------------|
| 1 (Low) | 10 | 4 |
| 5 (Medium) | 50 | 20 |
| 10 (High) | 100 | 40 |

---

## Factor 2: Time Decay (30%)

Tasks get more urgent as they age. Linear growth over 30 days:

```go
// backend/internal/domain/priority/calculator.go

func (calc *Calculator) calculateTimeDecay(createdAt time.Time) float64 {
    age := time.Since(createdAt)
    days := age.Hours() / 24

    // Linear growth over 30 days
    decay := (days / 30.0) * 100

    // Cap at 100
    return math.Min(100, decay)
}
```

### Time Decay Curve

```
Score
100 ─────────────────────────────────■■■■■■■■■
 90 │                           ■■■■
 80 │                       ■■■■
 70 │                   ■■■■
 60 │               ■■■■
 50 │           ■■■■
 40 │       ■■■■
 30 │   ■■■■
 20 ■■■■
 10 │
  0 ├───┬───┬───┬───┬───┬───┬───┬───► Days
    0   5   10  15  20  25  30  35
```

### Example Calculations

| Task Age | Time Decay | Weighted (×0.3) |
|----------|------------|-----------------|
| New (0 days) | 0 | 0 |
| 1 week | 23.3 | 7 |
| 2 weeks | 46.7 | 14 |
| 30 days | 100 | 30 |
| 60 days | 100 (capped) | 30 |

**Why linear, not exponential?** Linear is predictable. Users can understand "old tasks rise steadily."

---

## Factor 3: Deadline Urgency (20%)

**Quadratic** increase in the final 7 days. This matches human psychology - we feel more pressure as deadlines approach.

```go
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
```

### Urgency Curve (Quadratic)

```
Urgency
100 ─────────────────────────────────■
 90 │                               ■
 80 │                            ■■
 70 │                          ■■
 60 │                        ■■
 50 │                     ■■■
 40 │                  ■■■
 30 │              ■■■■
 20 │         ■■■■■
 10 │    ■■■■■
  0 ■■■■■────┬───┬───┬───┬───┬───┬───► Days Until Due
    14   7   6   5   4   3   2   1   0
```

### Example Calculations

| Days Until Due | Urgency Score | Weighted (×0.2) |
|----------------|---------------|-----------------|
| 14 days | 0 | 0 |
| 7 days | 0 | 0 |
| 5 days | 49 | 9.8 |
| 3 days | 82 | 16.4 |
| 1 day | 98 | 19.6 |
| Overdue | 100 | 20 |

**Why quadratic, not linear?** Quadratic creates the "sharp final push" that matches how we actually experience deadlines. Day 7 to day 3 is gradual, but day 2 to day 1 is dramatic.

---

## Factor 4: Bump Penalty (10%)

When users delay a task (bump it), the priority increases. This surfaces procrastinated work.

```go
func (calc *Calculator) calculateBumpPenalty(bumpCount int) float64 {
    penalty := float64(bumpCount) * 10
    return math.Min(50, penalty)  // Cap at 50
}
```

### Bump Penalty Table

| Bump Count | Penalty | Weighted (×0.1) | Status |
|------------|---------|-----------------|--------|
| 0 | 0 | 0 | Normal |
| 1 | 10 | 1 | Delayed once |
| 2 | 20 | 2 | Getting concerning |
| 3+ | 30-50 | 3-5 | **At-risk task!** |
| 5+ | 50 (cap) | 5 | Maximum penalty |

**At-Risk Detection:** Tasks with 3+ bumps are flagged as "at-risk" in the UI.

---

## Factor 5: Effort Boost (Multiplier)

Small tasks get a boost to encourage "quick wins" - completing small tasks builds momentum.

```go
func (calc *Calculator) getEffortBoost(effort *domain.TaskEffort) float64 {
    if effort == nil {
        return 1.0 // No estimate = no boost
    }
    return effort.GetEffortMultiplier()
}

// In domain/task.go
func (e TaskEffort) GetEffortMultiplier() float64 {
    switch e {
    case TaskEffortSmall:
        return 1.3  // 30% boost
    case TaskEffortMedium:
        return 1.15 // 15% boost
    case TaskEffortLarge:
        return 1.0  // No boost
    case TaskEffortXLarge:
        return 0.9  // 10% reduction (don't let big tasks dominate)
    default:
        return 1.0
    }
}
```

### Effort Multipliers

| Effort | Multiplier | Effect |
|--------|------------|--------|
| Small | 1.3x | +30% priority |
| Medium | 1.15x | +15% priority |
| Large | 1.0x | No change |
| X-Large | 0.9x | -10% priority |

**Why boost small tasks?** Quick wins build momentum and prevent the backlog from being dominated by large projects.

---

## Complete Calculation Example

**Task:** "Review PR feedback"
- User Priority: 7/10
- Created: 5 days ago
- Due: 2 days from now
- Bump Count: 1
- Effort: Small

**Step-by-step:**

```
1. User Priority: 7 × 10 = 70
   Weighted: 70 × 0.4 = 28

2. Time Decay: (5/30) × 100 = 16.67
   Weighted: 16.67 × 0.3 = 5

3. Deadline Urgency: 100 × (1 - (2/7)²) = 91.8
   Weighted: 91.8 × 0.2 = 18.4

4. Bump Penalty: 1 × 10 = 10
   Weighted: 10 × 0.1 = 1

5. Base Score: 28 + 5 + 18.4 + 1 = 52.4

6. Effort Boost: 52.4 × 1.3 = 68.1

Final Score: 68 (rounded)
```

---

## Priority Breakdown (Transparency Feature)

Users can see exactly why a task is prioritized:

```go
// backend/internal/domain/priority/calculator.go

func (calc *Calculator) CalculateWithBreakdown(task *domain.Task) (int, *domain.PriorityBreakdown) {
    userPriority := float64(task.UserPriority * 10)
    timeDecay := calc.calculateTimeDecay(task.CreatedAt)
    deadlineUrgency := calc.calculateDeadlineUrgency(task.DueDate)
    bumpPenalty := calc.calculateBumpPenalty(task.BumpCount)
    effortBoost := calc.getEffortBoost(task.EstimatedEffort)

    // Calculate weighted contributions
    userPriorityWeighted := userPriority * 0.4
    timeDecayWeighted := timeDecay * 0.3
    deadlineUrgencyWeighted := deadlineUrgency * 0.2
    bumpPenaltyWeighted := bumpPenalty * 0.1

    score := userPriorityWeighted + timeDecayWeighted +
             deadlineUrgencyWeighted + bumpPenaltyWeighted
    score = score * effortBoost

    // Build breakdown for UI
    breakdown := &domain.PriorityBreakdown{
        UserPriority:            userPriority,
        TimeDecay:               timeDecay,
        DeadlineUrgency:         deadlineUrgency,
        BumpPenalty:             bumpPenalty,
        EffortBoost:             effortBoost,
        UserPriorityWeighted:    userPriorityWeighted * effortBoost,
        TimeDecayWeighted:       timeDecayWeighted * effortBoost,
        DeadlineUrgencyWeighted: deadlineUrgencyWeighted * effortBoost,
        BumpPenaltyWeighted:     bumpPenaltyWeighted * effortBoost,
    }

    return int(math.Min(100, math.Max(0, score))), breakdown
}
```

### UI Display

The frontend shows a donut chart with each factor's contribution:

```
┌─────────────────────────────────────┐
│   Priority: 68                       │
│                                      │
│      ╭───────╮                       │
│     ╱  User   ╲  User Priority: 28  │
│    │  28pts   │  Time Decay: 5      │
│    │ ╭─────╮  │  Deadline: 18       │
│     ╲│ 68 │╱   Bumps: 1            │
│      ╰─────╯   ──────               │
│                Effort: ×1.3         │
└─────────────────────────────────────┘
```

---

## At-Risk Detection

Separate from priority, tasks are flagged as "at-risk":

```go
// backend/internal/domain/priority/calculator.go

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
```

**At-Risk Criteria:**
- 3+ bumps (repeated delays)
- 3+ days overdue

---

## Testing the Algorithm

The priority calculator has 100% test coverage:

```go
// backend/internal/domain/priority/calculator_test.go

func TestCalculate_UserPriorityDominates(t *testing.T) {
    pc := NewCalculator()

    // High priority task (10/10)
    highPriority := &domain.Task{
        UserPriority: 10,
        CreatedAt:    time.Now(),
    }

    // Low priority task (1/10)
    lowPriority := &domain.Task{
        UserPriority: 1,
        CreatedAt:    time.Now(),
    }

    highScore := pc.Calculate(highPriority)
    lowScore := pc.Calculate(lowPriority)

    assert.Greater(t, highScore, lowScore)
    assert.GreaterOrEqual(t, highScore, 40) // User priority alone is 40
}

func TestCalculateDeadlineUrgency_QuadraticGrowth(t *testing.T) {
    pc := NewCalculator()

    tests := []struct {
        daysUntilDue int
        wantMin      float64
        wantMax      float64
    }{
        {7, 0, 1},       // Just at threshold
        {5, 48, 52},     // Mid-urgency
        {3, 80, 84},     // High urgency
        {1, 96, 100},    // Very high urgency
    }

    for _, tt := range tests {
        dueDate := time.Now().Add(time.Duration(tt.daysUntilDue) * 24 * time.Hour)
        urgency := pc.calculateDeadlineUrgency(&dueDate)

        assert.GreaterOrEqual(t, urgency, tt.wantMin)
        assert.LessOrEqual(t, urgency, tt.wantMax)
    }
}
```

---

## Design Decisions

### Why Not Machine Learning?

| ML Approach | Rule-Based Approach |
|-------------|---------------------|
| Black box (hard to explain) | Transparent (users see factors) |
| Needs training data | Works immediately |
| Can drift over time | Consistent behavior |
| Complex to debug | Easy to debug |

**Decision:** Explainability > sophistication for an MVP.

### Why These Specific Weights?

The weights (40/30/20/10) were chosen based on:

1. **User intent is paramount** (40%) - The user knows their context
2. **Time awareness matters** (30%) - Old tasks shouldn't be forgotten
3. **Deadlines create focus** (20%) - But shouldn't override everything
4. **Bumps are signals** (10%) - Important but not dominant

These could be made configurable per-user in a future version.

---

## Exercises

### 🔰 Beginner: Calculate by Hand

Calculate the priority for this task:
- User Priority: 5/10
- Created: 10 days ago
- Due: tomorrow
- Bump Count: 2
- Effort: Medium

### 🎯 Intermediate: Add a Factor

Design a new factor: **Collaboration Priority**
- Tasks with collaborators should get a boost
- How would you weight it?
- What data would you need?

### 🚀 Advanced: A/B Test Design

Design an A/B test to compare:
- Current weights (40/30/20/10)
- Alternative weights (50/25/15/10)

What metrics would you measure? How long would you run it?

---

## Reflection Questions

1. **Why quadratic for deadlines but linear for time decay?** What would change if both were linear?

2. **Why cap bump penalty at 50?** What could happen without a cap?

3. **Why is the algorithm deterministic?** What would the tradeoffs be for adding randomness?

4. **How would you add task dependencies?** Should blocked tasks have lower or higher priority?

---

## Key Takeaways

1. **Explainability builds trust.** Users can see exactly why tasks are ranked.

2. **Multiple factors prevent gaming.** You can't just set everything to "high priority."

3. **Curves matter.** Quadratic deadline urgency matches human psychology.

4. **Test thoroughly.** 100% coverage prevents algorithm bugs from affecting users.

5. **Start simple.** Rule-based beats ML when explainability matters.

---

## Next Module

Continue to **[Module 04: Frontend Architecture](./04-frontend-architecture.md)** to see how the priority data flows to the UI.
