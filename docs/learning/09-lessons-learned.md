# Module 09: Lessons Learned

## Learning Objectives

By the end of this module, you will:
- Synthesize key lessons from the TaskFlow development journey
- Extract reusable patterns for your own projects
- Understand what worked well and what could improve
- Have a checklist for future full-stack projects

---

## What Worked Well

### 1. Documentation-First Development

**Pattern:** Write specs before code.

**Evidence:**
- 14 documentation files before first code commit
- PRD, data model, priority algorithm all specified upfront
- Phase plans with clear exit criteria

**Benefits:**
- Caught design issues before implementation
- Enabled parallel work (frontend could mock backend)
- Created living reference documentation
- Reduced scope creep ("it's not in the spec")

**Reusable Practice:**
```
Before coding a feature:
1. Write a 1-page spec (problem, solution, acceptance criteria)
2. Define API contract (request/response shapes)
3. Draw data model (entities, relationships)
4. List edge cases and error scenarios
```

---

### 2. Type Safety Everywhere

**Pattern:** TypeScript (frontend) + Go (backend) + sqlc (database).

**Evidence:**
- Zero runtime type errors in production
- Compile-time SQL validation
- Refactoring with confidence

**Benefits:**
- Bugs caught at build time, not runtime
- IDE autocomplete and go-to-definition
- Safe schema changes (sqlc regeneration fails fast)

**Reusable Practice:**
```
Type Safety Checklist:
□ TypeScript strict mode enabled
□ No 'any' types (use 'unknown' if needed)
□ Database types generated from schema
□ API types shared between frontend/backend
```

---

### 3. Clean Architecture with Interfaces

**Pattern:** Depend on abstractions, not implementations.

**Evidence:**
- 277+ backend tests with mocks
- Swappable implementations (Redis vs. in-memory rate limiting)
- Easy to test handlers without database

**Code Example:**
```go
// Handler depends on interface
type TaskHandler struct {
    taskService ports.TaskService  // Interface!
}

// Can inject mock for testing
func TestTaskHandler(t *testing.T) {
    mock := &MockTaskService{}
    handler := NewTaskHandler(mock)
    // Test without database!
}
```

**Reusable Practice:**
```
For each service:
1. Define interface in ports/
2. Implement in service/
3. Inject via constructor
4. Optional services use setter injection
```

---

### 4. Incremental Migrations

**Pattern:** One feature per migration, never modify existing migrations.

**Evidence:**
- 12 migrations, each focused on one feature
- Zero data loss incidents
- Easy rollback (each has .down.sql)

**Benefits:**
- Version control for database schema
- Safe production deployments
- Clear history of schema evolution

**Reusable Practice:**
```
Migration Rules:
1. NEVER modify applied migrations
2. Each migration = one logical change
3. Always write both up and down
4. Test rollback before deploying
```

---

### 5. Optimistic Updates for UX

**Pattern:** Update UI immediately, rollback on error.

**Evidence:**
- Task completion feels instant
- No loading spinners for common operations
- Graceful error handling with rollback

**Code Example:**
```typescript
onMutate: async (taskId) => {
  // 1. Snapshot for rollback
  const previous = queryClient.getQueryData(['tasks']);

  // 2. Optimistically update
  queryClient.setQueryData(['tasks'], (old) => ({
    ...old,
    tasks: old.tasks.map(t =>
      t.id === taskId ? { ...t, status: 'done' } : t
    ),
  }));

  return { previous };  // For rollback
},
onError: (err, taskId, context) => {
  // Rollback on error
  queryClient.setQueryData(['tasks'], context.previous);
},
```

---

### 6. Security Audit Mindset

**Pattern:** Review code specifically for security issues.

**Evidence:**
- PR #74 fixed JWT algorithm confusion vulnerability
- PR #75 configured connection pool limits
- PR #76 enforced pagination limits

**Findings Prevented:**
- Token forgery via algorithm confusion
- Database connection exhaustion
- Memory exhaustion via unlimited queries

**Reusable Practice:**
```
Security Checklist:
□ JWT algorithm explicitly validated
□ Input validation on all endpoints
□ Parameterized queries (no SQL injection)
□ Rate limiting configured
□ Pagination enforced
□ CORS properly configured
```

---

## What Could Improve

### 1. Earlier N+1 Detection

**Issue:** N+1 queries discovered late (audit phase).

**What Happened:**
- Dependencies feature shipped with N+1 pattern
- Discovered during code audit
- Fixed in PR #77

**What Would Help:**
```
Prevention Strategies:
□ Enable query logging in development
□ Set query count alerts (> 10 queries per request)
□ Review all loops that call repositories
□ Consider batch methods from the start
```

---

### 2. Feature Flags from Start

**Issue:** Anonymous users required conditional logic everywhere.

**What Happened:**
- PR #57 added anonymous user support
- Required adding `if user.IsAnonymous` checks throughout
- Some features gated, some not (inconsistent)

**What Would Help:**
```go
// Instead of scattered conditionals:
if user.IsAnonymous {
    return nil, ErrFeatureNotAvailable
}

// Use centralized feature flags:
if !features.IsEnabled("gamification", user) {
    return nil, ErrFeatureNotAvailable
}
```

---

### 3. More Comprehensive Error Types

**Issue:** Some errors lack context for debugging.

**What Happened:**
- Generic "internal server error" for some cases
- Log correlation missing in some paths
- Frontend can't always show specific messages

**What Would Help:**
```go
// Instead of:
return nil, err

// Return wrapped error with context:
return nil, fmt.Errorf("failed to create task for user %s: %w", userID, err)
```

---

### 4. Test Coverage Gaps

**Issue:** Handler and repository layers under 80%.

**What Happened:**
- Focus on service layer tests
- Generated sqlc code (0%) drags down averages
- Some edge cases not covered

**What Would Help:**
```
Test Priority:
1. Business logic (services) - 80%+ ✓
2. Domain validation - 80%+ ✓
3. HTTP handlers - 80%+ (gap)
4. Repository integration - 70%+ (gap)
```

---

## Reusable Patterns Checklist

### Query Key Factory Pattern

```typescript
export const taskKeys = {
  all: ['tasks'] as const,
  lists: () => [...taskKeys.all, 'list'] as const,
  list: (filters?: Filters) => [...taskKeys.lists(), filters] as const,
  detail: (id: string) => [...taskKeys.all, 'detail', id] as const,
};

// Usage:
queryClient.invalidateQueries({ queryKey: taskKeys.lists() });
```

### Optimistic Update with Rollback

```typescript
useMutation({
  mutationFn: api.update,
  onMutate: async (data) => {
    await queryClient.cancelQueries({ queryKey: ['items'] });
    const previous = queryClient.getQueryData(['items']);
    queryClient.setQueryData(['items'], /* optimistic update */);
    return { previous };
  },
  onError: (err, data, context) => {
    queryClient.setQueryData(['items'], context.previous);
  },
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['items'] });
  },
});
```

### Setter Injection for Optional Dependencies

```go
type TaskService struct {
    taskRepo ports.TaskRepository  // Required
    gamification ports.GamificationService  // Optional
}

func NewTaskService(taskRepo ports.TaskRepository) *TaskService {
    return &TaskService{taskRepo: taskRepo}
}

func (s *TaskService) SetGamificationService(gs ports.GamificationService) {
    s.gamification = gs
}

func (s *TaskService) Complete(...) {
    // ... complete task ...
    if s.gamification != nil {
        s.gamification.ProcessAsync(...)
    }
}
```

### Batch Query Pattern

```go
// Instead of:
for _, id := range ids {
    result, _ := repo.GetByID(ctx, id)  // N queries!
}

// Use:
results, _ := repo.GetByIDBatch(ctx, ids)  // 1 query

// SQL:
// SELECT * FROM items WHERE id = ANY($1::uuid[])
```

### Async Processing Pattern

```go
func (s *Service) ProcessAsync(data Data) {
    go func() {
        ctx := context.Background()  // Independent context

        if err := s.process(ctx, data); err != nil {
            slog.Error("Async processing failed",
                "error", err,
                "data_id", data.ID,
            )
        }
    }()
}
```

---

## Full-Stack Project Checklist

### Phase 1: Foundation

```
□ Write product requirements document
□ Design database schema
□ Specify API endpoints
□ Create technology decision log
□ Setup project structure
□ Configure TypeScript strict mode
□ Setup Go with proper package structure
```

### Phase 2: Core Features

```
□ Implement authentication (JWT + bcrypt)
□ Build CRUD operations
□ Add input validation
□ Setup error handling
□ Implement business logic
□ Connect frontend to backend
□ Add loading/error states
```

### Phase 3: Production Readiness

```
□ Add structured logging
□ Configure rate limiting
□ Create database indexes
□ Setup CI/CD pipeline
□ Write unit tests (80%+ on business logic)
□ Add integration tests
□ Document API endpoints
```

### Phase 4: Polish

```
□ Performance profiling
□ N+1 query audit
□ Security audit
□ Add monitoring/metrics
□ Optimize bundle size
□ Add design system tokens
□ User testing feedback
```

---

## Final Reflection Questions

1. **What would you do differently?** Looking back at the journey, what decisions would you change?

2. **What patterns will you reuse?** Which patterns from TaskFlow will you apply to your next project?

3. **What's missing?** What features or practices would you add?

4. **How would you scale?** If TaskFlow had 10x users, what would break first?

---

## Key Takeaways

1. **Documentation-first prevents rework.** Invest time in specs before code.

2. **Type safety compounds.** TypeScript + Go + sqlc catches bugs across the stack.

3. **Clean architecture enables testing.** Interfaces make mocking trivial.

4. **Optimistic updates feel fast.** User experience trumps implementation simplicity.

5. **Build → Use → Optimize.** Don't over-engineer upfront.

6. **Security audits find real bugs.** Review code specifically for vulnerabilities.

7. **Document decisions and deferrals.** Future you will appreciate the context.

8. **Patterns are reusable.** Query key factories, batch queries, and async processing apply everywhere.

---

## Congratulations!

You've completed the TaskFlow Learning Curriculum! 🎓

You now understand:
- Full-stack architecture from frontend to database
- Clean Architecture with dependency injection
- React Query patterns for server state
- Database design and migration strategies
- Real-world refactoring techniques
- Technical decision-making frameworks

**Next Steps:**

1. **Build something.** Apply these patterns to your own project.
2. **Explore the code.** Clone TaskFlow and trace through the implementations.
3. **Experiment.** Try alternative approaches and compare results.
4. **Share.** Teach someone else what you learned.

Happy building! 🚀

---

## Quick Reference

### Key Files
| Purpose | Path |
|---------|------|
| Backend entry | `backend/cmd/server/main.go` |
| Task service | `backend/internal/service/task_service.go` |
| Priority algorithm | `backend/internal/domain/priority/calculator.go` |
| Task hooks | `frontend/hooks/useTasks.ts` |
| API client | `frontend/lib/api.ts` |

### Key Documentation
| Purpose | Path |
|---------|------|
| Project status | `docs/PROJECT_STATUS.md` |
| Priority spec | `docs/product/priority-algorithm.md` |
| N+1 fix | `docs/optimizations/n1-query-fix-explained.md` |
| Security audit | `docs/audits/findings-security.md` |

### Key Patterns
| Pattern | Use Case |
|---------|----------|
| Query key factory | Cache invalidation |
| Optimistic updates | Fast UI feedback |
| Setter injection | Optional dependencies |
| Batch queries | N+1 elimination |
| Async processing | Non-critical, slow operations |
