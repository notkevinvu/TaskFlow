# N+1 Query Elimination: Deep Dive

## What is the N+1 Query Problem?

The N+1 query problem is one of the most common performance anti-patterns in database-driven applications. It occurs when code fetches a list of N items, then makes an additional query for each item to get related data.

**The "N+1" name comes from:**
- **1** query to get the initial list
- **N** additional queries (one per item in the list)
- Total: **N+1** queries

---

## The Specific Pattern in TaskFlow

### Scenario: Completing a Blocker Task

In TaskFlow, tasks can have dependencies. Task B might be "blocked by" Task A, meaning B can't be completed until A is done. When A completes, the system needs to check if B (and any other tasks blocked by A) are now fully unblocked.

**The complication:** A task might be blocked by *multiple* tasks. So when Task A completes, we can't just mark Task B as unblocked - we need to check if B has any *other* incomplete blockers.

### The Original Code (N+1 Pattern)

```go
// File: backend/internal/service/dependency_service.go (lines 183-199)

func (s *DependencyService) GetBlockerCompletionInfo(ctx context.Context, blockerTaskID string) (*domain.BlockerCompletionInfo, error) {
    // Query 1: Get all tasks blocked by the completing task
    blockedTaskIDs, err := s.dependencyRepo.GetTasksBlockedBy(ctx, blockerTaskID)
    // Returns: ["task-B", "task-C", "task-D", "task-E", "task-F"] (5 tasks)

    var unblockedIDs []string

    // N queries: One query PER blocked task
    for _, taskID := range blockedTaskIDs {
        // Query 2: Check task-B's incomplete blockers
        // Query 3: Check task-C's incomplete blockers
        // Query 4: Check task-D's incomplete blockers
        // Query 5: Check task-E's incomplete blockers
        // Query 6: Check task-F's incomplete blockers
        incompleteCount, err := s.dependencyRepo.CountIncompleteBlockers(ctx, taskID)
        if incompleteCount == 0 {
            unblockedIDs = append(unblockedIDs, taskID)
        }
    }

    return &domain.BlockerCompletionInfo{...}, nil
}
```

**Database Activity (5 blocked tasks):**
```
Query 1: SELECT task_id FROM task_dependencies WHERE blocked_by_id = 'task-A'
         → Returns 5 task IDs

Query 2: SELECT COUNT(*) FROM task_dependencies td
         INNER JOIN tasks t ON t.id = td.blocked_by_id
         WHERE td.task_id = 'task-B' AND t.status != 'done'

Query 3: SELECT COUNT(*) FROM task_dependencies td
         INNER JOIN tasks t ON t.id = td.blocked_by_id
         WHERE td.task_id = 'task-C' AND t.status != 'done'

Query 4: SELECT COUNT(*) FROM task_dependencies td
         INNER JOIN tasks t ON t.id = td.blocked_by_id
         WHERE td.task_id = 'task-D' AND t.status != 'done'

Query 5: SELECT COUNT(*) FROM task_dependencies td
         INNER JOIN tasks t ON t.id = td.blocked_by_id
         WHERE td.task_id = 'task-E' AND t.status != 'done'

Query 6: SELECT COUNT(*) FROM task_dependencies td
         INNER JOIN tasks t ON t.id = td.blocked_by_id
         WHERE td.task_id = 'task-F' AND t.status != 'done'

Total: 6 queries (1 + 5)
```

---

## The Fix: Batch Query

### The New Code

```go
// File: backend/internal/service/dependency_service.go (lines 183-202)

func (s *DependencyService) GetBlockerCompletionInfo(ctx context.Context, blockerTaskID string) (*domain.BlockerCompletionInfo, error) {
    // Query 1: Get all tasks blocked by the completing task (unchanged)
    blockedTaskIDs, err := s.dependencyRepo.GetTasksBlockedBy(ctx, blockerTaskID)
    // Returns: ["task-B", "task-C", "task-D", "task-E", "task-F"]

    var unblockedIDs []string

    if len(blockedTaskIDs) > 0 {
        // Query 2: ONE query for ALL blocked tasks
        blockerCounts, err := s.dependencyRepo.CountIncompleteBlockersBatch(ctx, blockedTaskIDs)
        // Returns: map[string]int{
        //   "task-B": 0,  // No other blockers → UNBLOCKED!
        //   "task-C": 2,  // Still has 2 other incomplete blockers
        //   "task-D": 0,  // No other blockers → UNBLOCKED!
        //   "task-E": 1,  // Still has 1 other incomplete blocker
        //   "task-F": 0,  // No other blockers → UNBLOCKED!
        // }

        for _, taskID := range blockedTaskIDs {
            if blockerCounts[taskID] == 0 {
                unblockedIDs = append(unblockedIDs, taskID)
            }
        }
    }

    return &domain.BlockerCompletionInfo{...}, nil
}
```

### The Batch Query SQL

```go
// File: backend/internal/repository/dependency_repository.go (lines 267-305)

func (r *DependencyRepository) CountIncompleteBlockersBatch(ctx context.Context, taskIDs []string) (map[string]int, error) {
    result := make(map[string]int)

    // Pre-populate with zeros (tasks not in result have 0 incomplete blockers)
    for _, id := range taskIDs {
        result[id] = 0
    }

    // SINGLE query with GROUP BY
    rows, err := r.db.Query(ctx, `
        SELECT td.task_id, COUNT(*)::int as incomplete_count
        FROM task_dependencies td
        INNER JOIN tasks t ON t.id = td.blocked_by_id
        WHERE td.task_id = ANY($1)           -- PostgreSQL array parameter
          AND t.status != 'done'
        GROUP BY td.task_id                  -- Aggregate by task
    `, taskIDs)

    // ... scan results into map ...

    return result, nil
}
```

**Database Activity (5 blocked tasks):**
```
Query 1: SELECT task_id FROM task_dependencies WHERE blocked_by_id = 'task-A'
         → Returns 5 task IDs

Query 2: SELECT td.task_id, COUNT(*)::int as incomplete_count
         FROM task_dependencies td
         INNER JOIN tasks t ON t.id = td.blocked_by_id
         WHERE td.task_id = ANY(ARRAY['task-B','task-C','task-D','task-E','task-F'])
           AND t.status != 'done'
         GROUP BY td.task_id
         → Returns: [('task-C', 2), ('task-E', 1)]
         (Only tasks WITH incomplete blockers appear in result)

Total: 2 queries (always 2, regardless of N)
```

---

## Query Reduction Math

| Blocked Tasks (N) | Before (N+1) | After (Batch) | Queries Saved | Reduction % |
|-------------------|--------------|---------------|---------------|-------------|
| 1 | 2 | 2 | 0 | 0% |
| 5 | 6 | 2 | 4 | 67% |
| 10 | 11 | 2 | 9 | **82%** |
| 25 | 26 | 2 | 24 | 92% |
| 50 | 51 | 2 | 49 | **96%** |
| 100 | 101 | 2 | 99 | 98% |

**Formula:**
- Before: `N + 1` queries
- After: `2` queries (constant)
- Reduction: `(N + 1 - 2) / (N + 1)` = `(N - 1) / (N + 1)`

As N grows, reduction approaches **100%**.

---

## Latency Impact

Each database query has overhead:
- **Network round-trip**: ~1-5ms (local) to ~10-50ms (cloud/remote)
- **Query parsing & planning**: ~0.5-2ms
- **Index lookup**: ~0.1-1ms
- **Result serialization**: ~0.1-0.5ms

**Typical per-query latency: 5-10ms**

| Scenario | Before | After | Time Saved |
|----------|--------|-------|------------|
| 10 tasks @ 5ms/query | 55ms | 10ms | **45ms** |
| 10 tasks @ 10ms/query | 110ms | 20ms | **90ms** |
| 50 tasks @ 5ms/query | 255ms | 10ms | **245ms** |
| 50 tasks @ 10ms/query | 510ms | 20ms | **490ms** |

---

## Why This Pattern Happens

N+1 queries often appear because:

1. **Intuitive coding**: The loop pattern is how humans naturally think about the problem
2. **Works fine in development**: With small test data (N=1 or 2), performance is acceptable
3. **Hidden in abstractions**: ORMs and repository patterns can hide the actual query count
4. **Late discovery**: Performance degrades gradually as data grows

```go
// This LOOKS efficient - just a simple loop!
for _, task := range tasks {
    data := fetchRelatedData(task.ID)  // Hidden: This is a DB query!
}
```

---

## Visual Comparison

**Before (N+1):**
```
Application                    Database
    │                              │
    ├──── Query 1: Get list ──────►│
    │◄─── [task-B, C, D, E, F] ────┤
    │                              │
    ├──── Query 2: Count for B ───►│
    │◄─── 0 ───────────────────────┤
    │                              │
    ├──── Query 3: Count for C ───►│
    │◄─── 2 ───────────────────────┤
    │                              │
    ├──── Query 4: Count for D ───►│
    │◄─── 0 ───────────────────────┤
    │                              │
    ├──── Query 5: Count for E ───►│
    │◄─── 1 ───────────────────────┤
    │                              │
    ├──── Query 6: Count for F ───►│
    │◄─── 0 ───────────────────────┤
    │                              │
    ▼                              ▼
   Done                          Done

Total round-trips: 6
```

**After (Batch):**
```
Application                    Database
    │                              │
    ├──── Query 1: Get list ──────►│
    │◄─── [task-B, C, D, E, F] ────┤
    │                              │
    ├──── Query 2: Batch count ───►│
    │     (all 5 IDs at once)      │
    │◄─── {B:0, C:2, D:0, E:1, F:0}┤
    │                              │
    ▼                              ▼
   Done                          Done

Total round-trips: 2
```

---

## PostgreSQL Optimization

The batch query uses `ANY($1)` which PostgreSQL optimizes efficiently:

```sql
-- The ANY operator with an array
WHERE td.task_id = ANY($1)

-- Is equivalent to (but more efficient than)
WHERE td.task_id IN ('task-B', 'task-C', 'task-D', 'task-E', 'task-F')
```

PostgreSQL uses the existing index `idx_task_dependencies_task_id` to perform an **Index Scan** rather than a sequential table scan, making the batch query nearly as fast as a single lookup.

---

## Key Takeaway

**The N+1 problem scales linearly with data size**, making it a "time bomb" in codebases. A feature that works fine with 5 tasks becomes unusably slow with 500 tasks. The batch query pattern converts **O(N) database round-trips** into **O(1)**, providing consistent performance regardless of data volume.

This fix ensures that completing a blocker task takes ~10-20ms whether it unblocks 5 tasks or 500 tasks.

---

## References

- PR: https://github.com/notkevinvu/TaskFlow/pull/77
- Audit Finding: `docs/audits/findings-optimization.md`
