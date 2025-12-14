# Module 06: The Development Journey

## Learning Objectives

By the end of this module, you will:
- Understand the phased development approach used in TaskFlow
- See how features evolved over 28 days (237 commits)
- Learn from the development timeline and decision points
- Identify which features depended on each other

---

## Overview: 0 to 1 in 28 Days

TaskFlow was built from an empty repository to a feature-complete v1.0 in 28 days. This module traces that journey, showing not just what was built, but **when** and **why** in that order.

### Development Timeline

```
Nov 15        Nov 22        Dec 1         Dec 7         Dec 13
   │             │             │             │             │
   ▼             ▼             ▼             ▼             ▼
┌──────┐    ┌──────┐    ┌──────────┐   ┌──────────┐   ┌──────┐
│Phase │    │Phase │    │ Phase 3  │   │ Phase 5  │   │Audit │
│  1   │───▶│  2   │───▶│Code Qual.│──▶│ Features │──▶│Phase │
│Docs +│    │Backend│    │ Testing │   │ Advanced │   │ v1.0 │
│ UI   │    │Integr.│    │ Indexes │   │ Gamific. │   │      │
└──────┘    └──────┘    └──────────┘   └──────────┘   └──────┘
```

### Key Statistics

| Metric | Value |
|--------|-------|
| Total commits | 237 |
| Total PRs | 84 |
| Development days | 28 |
| Database migrations | 12 |
| Backend lines (Go) | ~6,000 |
| Frontend lines (TS) | ~11,000 |
| Test count | 277+ backend, 29 frontend |

---

## Phase 1: Documentation + Frontend (Nov 15)

**Duration:** 1 day
**Commits:** 2
**Lines of code:** 9,269 (frontend)

### What Was Built

```
Phase 1 Deliverables:
├── Documentation (14 files)
│   ├── PRD.md - Product requirements
│   ├── data-model.md - Database schema
│   ├── priority-algorithm.md - Algorithm specification
│   └── phase-1-4 implementation plans
│
└── Frontend Shell
    ├── Next.js 16 + React 19 + TypeScript
    ├── Authentication pages (login/register)
    ├── Dashboard layout with sidebar
    ├── Task display with priority colors
    ├── Analytics page skeleton
    └── React Query setup (ready for backend)
```

### Key Decision: Documentation First

Before writing a single line of application code, the entire feature roadmap was planned:

```
docs/
├── product/
│   ├── PRD.md                    # What we're building
│   ├── data-model.md             # Database schema
│   └── priority-algorithm.md     # Business rules
└── implementation/
    ├── phase-1-weeks-1-2.md      # Frontend plan
    ├── phase-2-weeks-3-4.md      # Backend plan
    ├── phase-3-weeks-5-6.md      # Production readiness
    └── phase-4-month-2-plus.md   # Future features
```

**Why documentation first?**

1. **Clarity of vision** - Forces you to think through the entire system before building
2. **Enables parallel work** - Multiple developers could work from the same spec
3. **Reduces rework** - Catches design issues before they're coded
4. **Reference material** - Documentation becomes living reference

### Frontend-First Approach

The frontend was built completely before the backend existed:

```typescript
// frontend/lib/api.ts (Phase 1)
export const taskAPI = {
  list: async () => {
    // TODO: Replace with real API call
    return mockTasks;
  },
  create: async (data: CreateTaskDTO) => {
    // TODO: Replace with real API call
    return { ...data, id: 'mock-id' };
  },
};
```

**Why frontend first?**

1. **Visual prototype** - Stakeholders can see and interact with the product
2. **API contract defined** - Frontend defines what data it needs
3. **Parallel development** - Backend can be built to match frontend expectations
4. **UX iteration** - Easier to iterate on UI without backend changes

---

## Phase 2: Backend Integration (Nov 22)

**Duration:** 7 days
**Commits:** 50+
**Lines of code:** ~3,000 (backend)

### What Was Built

```
Phase 2 Deliverables:
├── Go Backend
│   ├── Clean Architecture structure
│   ├── Gin HTTP framework
│   ├── JWT authentication
│   ├── bcrypt password hashing
│   └── Priority calculation algorithm
│
├── Database
│   ├── PostgreSQL on Supabase
│   ├── 3 initial tables (users, tasks, task_history)
│   ├── Full-text search (tsvector + GIN)
│   └── Composite indexes
│
└── Integration
    ├── 12 API endpoints
    ├── Frontend connected to real backend
    ├── Rate limiting (100 req/min)
    └── CORS configuration
```

### Architecture Decision: Clean Architecture

The backend was structured with clear layer separation from day one:

```
backend/internal/
├── domain/         # Business entities (no dependencies)
├── repository/     # Data access (depends on domain)
├── service/        # Business logic (depends on domain, repository)
└── handler/        # HTTP layer (depends on service)
```

**Why Clean Architecture from the start?**

1. **Testability** - Each layer can be tested in isolation
2. **Flexibility** - Database or framework can be swapped
3. **Maintainability** - Clear ownership of responsibilities
4. **Onboarding** - New developers know where to look

### Priority Algorithm Implementation

The priority algorithm was one of the first backend features:

```go
// backend/internal/domain/priority/calculator.go
func (calc *Calculator) CalculateWithBreakdown(task *domain.Task) (int, *domain.PriorityBreakdown) {
    userPriority := float64(task.UserPriority * 10)       // 40%
    timeDecay := calc.calculateTimeDecay(task.CreatedAt)  // 30%
    deadlineUrgency := calc.calculateDeadlineUrgency(...)  // 20%
    bumpPenalty := calc.calculateBumpPenalty(...)          // 10%
    effortBoost := calc.getEffortBoost(...)               // Multiplier

    score := userPriority*0.4 + timeDecay*0.3 +
             deadlineUrgency*0.2 + bumpPenalty*0.1
    score = score * effortBoost

    return int(math.Min(100, math.Max(0, score))), breakdown
}
```

**Key insight:** The algorithm was documented in `docs/product/priority-algorithm.md` BEFORE implementation. The code matches the spec exactly.

---

## Phase 3: Code Quality (Nov 27 - Dec 1)

**Duration:** 5 days
**Commits:** 15+
**PRs:** #10-#27

### What Was Built

```
Phase 3 Deliverables:
├── Code Quality
│   ├── sqlc migration (type-safe SQL)
│   ├── Interface-based DI
│   ├── Custom error types
│   ├── Input validation
│   └── Structured logging (slog)
│
├── Infrastructure
│   ├── Redis rate limiting (with fallback)
│   ├── Database composite indexes
│   └── GitHub Actions CI/CD
│
└── Testing
    ├── testify + testcontainers setup
    ├── AuthHandler tests
    ├── TaskHandler tests
    ├── Service layer tests (55 tests)
    └── Frontend Vitest tests (29 tests)
```

### Major Refactor: Raw SQL → sqlc

This was the first major refactoring effort:

**Before (Phase 2):**
```go
// Manual SQL with error-prone scanning
query := `SELECT id, title, status FROM tasks WHERE user_id = $1`
rows, err := r.db.Query(ctx, query, userID)
defer rows.Close()

var tasks []Task
for rows.Next() {
    var id, title, status string
    if err := rows.Scan(&id, &title, &status); err != nil {
        return nil, err
    }
    tasks = append(tasks, Task{ID: id, Title: title, Status: status})
}
```

**After (Phase 3):**
```go
// Type-safe, generated code
tasks, err := r.queries.GetTasksByUserID(ctx, userID)
// That's it! No manual scanning, compile-time type checking
```

**Impact:**
- 382 lines of boilerplate eliminated
- Compile-time SQL validation
- Better IDE support (autocomplete, go-to-definition)

### Interface-Based Dependency Injection

Another Phase 3 refactor enabled proper testing:

**Before:**
```go
type TaskHandler struct {
    taskService *TaskService  // Concrete type = hard to test
}
```

**After:**
```go
type TaskHandler struct {
    taskService ports.TaskService  // Interface = can mock
}
```

**Why this mattered:**
```go
// Now we can write tests with mocks
func TestTaskHandler_Create(t *testing.T) {
    mockService := &MockTaskService{}
    handler := NewTaskHandler(mockService)
    // Test handler without real database
}
```

---

## Phase 4: Analytics & Polish (Dec 1-5)

**Duration:** 4 days
**PRs:** #28-#41

### What Was Built

```
Phase 4 Deliverables:
├── Analytics
│   ├── Date range picker
│   ├── Filter presets (High Priority, Due This Week)
│   ├── URL persistence for filters
│   ├── CategoryTrendsChart
│   ├── ProductivityHeatmap
│   └── Prometheus metrics
│
├── UI Polish
│   ├── Design tokens (60 CSS variables)
│   ├── Archive feature
│   └── Calendar view with task badges
│
└── Smart Features
    └── Insights service (rule-based suggestions)
```

### Design Tokens System

Created a systematic approach to styling:

```css
/* frontend/app/tokens.css - 60 design tokens */
:root {
  --color-primary: 222.2 84% 4.9%;
  --color-secondary: 210 40% 96.1%;
  --spacing-xs: 0.25rem;
  --spacing-sm: 0.5rem;
  --radius-sm: 0.25rem;
  --shadow-card: 0 1px 3px rgba(0,0,0,0.12);
  /* ... 54 more tokens */
}
```

**Why design tokens?**
- Consistency across components
- Easy theme changes (dark mode)
- Design-to-code translation

---

## Phase 5: Advanced Features (Dec 5-8)

**Duration:** 3 days
**PRs:** #44-#57

### What Was Built

This was the most feature-dense phase:

```
Phase 5 Deliverables:
├── 5A: Quick Wins
│   ├── Recurring tasks (daily/weekly/monthly)
│   ├── Priority explanation panel
│   ├── Quick add (Cmd+K)
│   └── Keyboard navigation (j/k/e/c/d)
│
├── 5B: Core Enhancements
│   ├── Subtasks (parent-child hierarchy)
│   ├── Blocked-by dependencies
│   └── Gamification (streaks, achievements)
│
├── 5C: Advanced Features
│   ├── Task templates
│   └── Pomodoro timer
│
└── Anonymous User Support
    ├── Guest mode
    ├── Feature gating
    └── Account conversion
```

### Feature Complexity: Gamification

The gamification system was the largest single feature (2,674 lines, 22 files):

**Database additions:**
```sql
CREATE TABLE user_achievements (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id),
    achievement_id TEXT NOT NULL,
    earned_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE gamification_stats (
    user_id UUID PRIMARY KEY REFERENCES users(id),
    total_xp INT DEFAULT 0,
    current_streak INT DEFAULT 0,
    best_streak INT DEFAULT 0,
    level INT DEFAULT 1
);
```

**Service implementation:**
```go
type GamificationService struct {
    repo ports.GamificationRepository
}

func (s *GamificationService) ProcessTaskCompletion(ctx context.Context, userID string, task *domain.Task) (*domain.GamificationResult, error) {
    // Calculate XP based on task properties
    // Check for new achievements
    // Update streak
    // Return results for UI toast
}
```

### Performance Issue Discovered

After shipping gamification, completion was slow (~500ms):

```
Complete Task Request Timeline (before optimization):
├── Mark task complete: 20ms
├── Update priority: 10ms
├── Create history: 15ms
├── Calculate XP: 50ms
├── Check achievements: 100ms   ← Problem!
├── Update streak: 50ms
├── Get stats: 100ms            ← Problem!
├── Get category mastery: 50ms
└── Response: 500ms total
```

This led to the performance optimization in the audit phase.

---

## Audit Phase: Security & Performance (Dec 9-13)

**Duration:** 4 days
**PRs:** #74-#84

### The Audit

A comprehensive code review identified 5 critical/high priority issues:

| Issue | Severity | PR |
|-------|----------|-----|
| JWT algorithm not validated | Critical | #74 |
| Database pool defaults (max 4) | High | #75 |
| No pagination enforcement | High | #76 |
| N+1 query in dependencies | High | #77 |
| No response compression | Medium | #78 |

### Fixes Applied

**1. JWT Algorithm Validation (Security)**
```go
// BEFORE: Accepts any algorithm!
token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    return []byte(secret), nil
})

// AFTER: Validates HMAC only
token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
        return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
    }
    return []byte(secret), nil
})
```

**2. Connection Pool Tuning (Performance)**
```go
// BEFORE: pgxpool defaults (max 4 connections)
dbPool, err := pgxpool.New(ctx, databaseURL)

// AFTER: Configured for production
config.MaxConns = 25
config.MinConns = 5
config.MaxConnIdleTime = 30 * time.Minute
dbPool, err := pgxpool.NewWithConfig(ctx, config)
```

**3. N+1 Query Elimination (Performance)**
```go
// BEFORE: N queries for N tasks
for _, taskID := range blockedTaskIDs {
    count, _ := repo.CountIncompleteBlockers(ctx, taskID)
    // ...
}

// AFTER: 1 query for all tasks
counts, _ := repo.CountIncompleteBlockersBatch(ctx, blockedTaskIDs)
```

**4. Async Gamification (Performance)**
```go
// BEFORE: Blocking (500ms added to response)
result, _ := gamificationService.ProcessTaskCompletion(ctx, userID, task)

// AFTER: Fire and forget (0ms added to response)
go func() {
    gamificationService.ProcessTaskCompletionAsync(userID, task)
}()
```

---

## Development Patterns Observed

### Pattern 1: Build → Use → Optimize

Features were built quickly, used in practice, then optimized:

```
Phase 2: Build raw SQL queries (works but verbose)
    ↓
Phase 3: Use them, notice boilerplate pain
    ↓
Phase 3: Migrate to sqlc (optimized)
```

```
Phase 5: Build gamification (works but slow)
    ↓
Audit: Use it, notice 500ms latency
    ↓
Audit: Add async processing (optimized)
```

### Pattern 2: Documentation → Code → Tests

Each feature followed this order:

```
1. Document the feature (spec or plan)
2. Implement the database schema
3. Implement the domain entities
4. Implement the repository
5. Implement the service
6. Implement the handler
7. Implement the frontend
8. Add tests
```

### Pattern 3: Commit Granularity

Features were split into digestible commits:

```
feat(gamification): Add database schema and domain types
feat(gamification): Add repository layer
feat(gamification): Add service layer
feat(gamification): Add handler layer
feat(gamification): Add frontend hooks
feat(gamification): Add UI components
test(gamification): Add service tests
```

**Why this matters:**
- Easy to review (smaller diffs)
- Easy to bisect (find which commit broke things)
- Easy to revert (undo just one piece)

---

## Timeline Visualization

```
Week 1 (Nov 15-22)
├── Day 1: Documentation + Frontend shell
├── Days 2-7: Backend implementation
│   ├── Clean Architecture setup
│   ├── Authentication (JWT + bcrypt)
│   ├── Task CRUD
│   ├── Priority algorithm
│   └── Full-text search
└── End of Week 1: Working MVP ✓

Week 2 (Nov 22-29)
├── Phase 2 enhancements
│   ├── Calendar widget
│   ├── Dark mode
│   └── Category management
├── Phase 3 begins
│   ├── sqlc migration
│   ├── Interface-based DI
│   └── Testing infrastructure
└── End of Week 2: Production-ready foundation ✓

Week 3 (Dec 1-7)
├── Phase 4 analytics
│   ├── Charts and heatmaps
│   ├── Design tokens
│   └── Prometheus metrics
├── Phase 5 features
│   ├── Recurring tasks
│   ├── Subtasks
│   ├── Dependencies
│   └── Gamification
└── End of Week 3: Feature-complete ✓

Week 4 (Dec 8-13)
├── Anonymous user support
├── Performance optimization
├── Security & performance audit
├── Bug fixes
└── v1.0 Release ✓
```

---

## Exercises

### 🔰 Beginner: Explore the Git History

```bash
cd TaskFlow
git log --oneline -50
```

1. Find the first commit that mentions "sqlc"
2. Find the first commit that mentions "gamification"
3. Count how many commits mention "fix"

### 🎯 Intermediate: Trace a Feature

1. Find all commits related to "dependencies" or "blocked-by"
   ```bash
   git log --oneline --all --grep="depend"
   git log --oneline --all --grep="blocked"
   ```
2. List the order in which layers were implemented
3. Identify which PR introduced the feature

### 🚀 Advanced: Plan Your Own Feature

1. Choose a hypothetical feature (e.g., "task labels" or "task comments")
2. Write a mini-spec following TaskFlow's pattern
3. List the commits you would make in order
4. Estimate time based on TaskFlow's velocity (~8 commits/day)

---

## Reflection Questions

1. **Why documentation first?** What problems does it prevent? What problems might it cause?

2. **Why was gamification slow initially?** Could this have been predicted? How?

3. **When did refactoring happen?** Was it too early? Too late? Just right?

4. **What was the riskiest change?** (Hint: think about the audit findings)

5. **If you had 50% more time, what would you add?** What would you skip?

---

## Key Takeaways

1. **Documentation-driven development** catches design issues early and aligns the team.

2. **Frontend-first** creates a visual prototype and defines the API contract.

3. **Build → Use → Optimize** is a natural pattern - don't over-engineer upfront.

4. **Audits find real issues.** The JWT vulnerability could have been a security incident.

5. **Commit granularity matters.** Small, focused commits enable bisecting and reverting.

6. **Performance is measurable.** The 500ms → 50ms improvement was discovered through usage.

---

## Next Module

Continue to **[Module 02: Backend Architecture](./02-backend-architecture.md)** to understand the Clean Architecture layers in detail.
