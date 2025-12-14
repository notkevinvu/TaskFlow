# Module 08: Technical Decisions

## Learning Objectives

By the end of this module, you will:
- Understand the "we did X because Y" decision framework
- Learn to evaluate tradeoffs systematically
- See intentional deferrals documented
- Apply decision tables to your own projects

---

## Decision Framework

Every technical decision follows this pattern:

```
┌─────────────────────────────────────────────────────────────┐
│                    DECISION FRAMEWORK                        │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  1. CONTEXT: What problem are we solving?                   │
│                                                              │
│  2. OPTIONS: What are the alternatives?                     │
│                                                              │
│  3. CRITERIA: What matters most?                            │
│     • Performance                                            │
│     • Developer experience                                   │
│     • Scalability                                            │
│     • Cost                                                   │
│     • Time to implement                                      │
│                                                              │
│  4. DECISION: What did we choose and why?                   │
│                                                              │
│  5. CONSEQUENCES: What tradeoffs did we accept?             │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Decision 1: JWT over Sessions

### Context

TaskFlow needs user authentication. Should we use stateless JWT tokens or server-side sessions?

### Options Compared

| Aspect | JWT | Sessions |
|--------|-----|----------|
| **Server State** | None (stateless) | Required (session store) |
| **Scalability** | Easy (any server can validate) | Harder (need shared store) |
| **Revocation** | Hard (wait for expiry) | Easy (delete session) |
| **Token Size** | Larger (contains claims) | Smaller (just session ID) |
| **Mobile Support** | Native | Requires cookies |
| **Microservices** | Easy (pass token between services) | Complex (need central auth) |

### Decision

**Chose JWT because:**
1. **Stateless** - No session store needed
2. **Scalable** - Any server can validate without shared state
3. **Mobile-ready** - Works with native apps (no cookies)
4. **Microservice-ready** - Services can validate independently

### Consequences

**Accepted tradeoffs:**
- Cannot instantly revoke tokens (must wait for expiry)
- Mitigation: Short expiry times (24h), refresh token pattern

### Implementation

```go
// backend/internal/middleware/auth.go

func AuthRequired(secret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenString := extractToken(c.GetHeader("Authorization"))

        token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
            // CRITICAL: Validate signing algorithm
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
            }
            return []byte(secret), nil
        })

        if err != nil || !token.Valid {
            c.AbortWithStatusJSON(401, gin.H{"error": "unauthorized"})
            return
        }

        claims := token.Claims.(jwt.MapClaims)
        c.Set(UserIDKey, claims["user_id"])
        c.Set(UserEmailKey, claims["email"])
        c.Next()
    }
}
```

---

## Decision 2: PostgreSQL over MongoDB

### Context

TaskFlow needs a database. Should we use relational (PostgreSQL) or document (MongoDB)?

### Options Compared

| Aspect | PostgreSQL | MongoDB |
|--------|------------|---------|
| **Data Model** | Relational (tables, joins) | Document (JSON) |
| **Relationships** | Native (foreign keys) | Manual (references) |
| **ACID** | Full support | Limited (multi-document) |
| **Schema** | Strict (migrations) | Flexible (schema-less) |
| **Analytics** | Excellent (window functions) | Limited |
| **Full-Text Search** | Built-in (tsvector) | Built-in (text index) |

### Decision

**Chose PostgreSQL because:**

1. **Complex Relationships** - Tasks have subtasks, dependencies, series
```sql
-- This is natural in SQL
SELECT t.*, COUNT(s.id) as subtask_count
FROM tasks t
LEFT JOIN tasks s ON s.parent_task_id = t.id
WHERE t.user_id = $1
GROUP BY t.id;
```

2. **Analytics Queries** - Priority distribution, completion trends
```sql
-- Window functions for ranking
SELECT title, priority_score,
       RANK() OVER (PARTITION BY category ORDER BY priority_score DESC) as category_rank
FROM tasks
WHERE user_id = $1;
```

3. **ACID Transactions** - Task completion with gamification
```sql
BEGIN;
UPDATE tasks SET status = 'done' WHERE id = $1;
INSERT INTO task_history (task_id, event_type) VALUES ($1, 'completed');
UPDATE gamification_stats SET total_xp = total_xp + 10 WHERE user_id = $2;
COMMIT;
```

### Consequences

**Accepted tradeoffs:**
- Schema migrations required for changes
- Less flexibility for unstructured data

---

## Decision 3: sqlc over ORMs

### Context

How should Go code interact with PostgreSQL? ORM (GORM), query builder (sqlx), or code generation (sqlc)?

### Options Compared

| Aspect | GORM (ORM) | sqlx (Query Builder) | sqlc (Code Gen) |
|--------|------------|---------------------|-----------------|
| **Learning Curve** | Medium | Low | Low |
| **SQL Knowledge** | Abstracted | Required | Required |
| **Type Safety** | Runtime | Runtime | Compile-time |
| **Performance** | Overhead | Direct | Direct |
| **Complex Queries** | Limited | Full SQL | Full SQL |
| **Migration** | Built-in | Manual | Manual |

### Decision

**Chose sqlc because:**

1. **Learn SQL properly** - No magic abstraction layer
2. **Compile-time type safety** - SQL errors caught during build
3. **Zero overhead** - Generated code is direct pgx calls
4. **Full SQL power** - Window functions, CTEs, complex joins

### Example

```sql
-- queries/tasks.sql
-- name: GetTasksWithSubtaskCount :many
SELECT t.*,
       (SELECT COUNT(*) FROM tasks s
        WHERE s.parent_task_id = t.id AND s.status != 'done') as incomplete_subtasks
FROM tasks t
WHERE t.user_id = $1 AND t.deleted_at IS NULL
ORDER BY t.priority_score DESC;
```

Generated Go code is type-safe and efficient.

---

## Decision 4: React Query over Redux

### Context

How should the frontend manage state? Global store (Redux) or purpose-built tools?

### Options Compared

| Aspect | Redux | React Query + Zustand |
|--------|-------|----------------------|
| **Boilerplate** | High (actions, reducers) | Low |
| **Server State** | Manual caching | Automatic |
| **Optimistic Updates** | Manual | Built-in |
| **Background Refetch** | Manual | Automatic |
| **Bundle Size** | ~15KB | ~12KB total |
| **Learning Curve** | Steep | Moderate |

### Decision

**Chose React Query + Zustand because:**

1. **Separation of concerns** - Server state ≠ client state
2. **Less boilerplate** - No actions/reducers for API calls
3. **Built-in optimistic updates** - Snappy UI with minimal code
4. **Automatic caching** - Stale-while-revalidate out of the box

```typescript
// This is ALL the code needed for task CRUD
export function useCreateTask() {
  return useMutation({
    mutationFn: taskAPI.create,
    onSuccess: () => queryClient.invalidateQueries(['tasks']),
  });
}

// Compare to Redux: actions, reducers, selectors, thunks...
```

---

## Decision 5: Priority Algorithm (Rule-Based over ML)

### Context

How should task priority be calculated? Machine learning or explicit rules?

### Options Compared

| Aspect | Machine Learning | Rule-Based |
|--------|-----------------|------------|
| **Explainability** | Low (black box) | High (visible formula) |
| **Training Data** | Required | Not needed |
| **Cold Start** | Problem | No problem |
| **Debugging** | Hard | Easy |
| **User Trust** | Lower | Higher |
| **Customization** | Hard | Easy |

### Decision

**Chose Rule-Based because:**

1. **Explainability** - Users can see exactly why tasks are ranked
2. **No training data** - Works from day one
3. **Debuggable** - Can verify calculation is correct
4. **Trust** - Users understand and trust transparent systems

### The Formula

```
Score = (UserPriority × 0.4 + TimeDecay × 0.3 + DeadlineUrgency × 0.2 + BumpPenalty × 0.1) × EffortBoost
```

**Future consideration:** Collect 6+ months of data, then train ML model if rule-based proves insufficient.

---

## Decision 6: Supabase over Self-Hosted

### Context

Where should PostgreSQL run? Self-hosted, AWS RDS, or Supabase?

### Options Compared

| Aspect | Self-Hosted | AWS RDS | Supabase |
|--------|-------------|---------|----------|
| **Setup Time** | Hours | 30 min | 5 min |
| **Cost** | VPS cost | $15+/month | Free tier |
| **Management** | Full | Partial | None |
| **Backups** | Manual | Automatic | Automatic |
| **Scaling** | Manual | Assisted | Automatic |
| **Dashboard** | None | AWS Console | Supabase UI |

### Decision

**Chose Supabase because:**

1. **Free tier** - 500MB database, 2GB bandwidth
2. **Zero management** - No server maintenance
3. **Great DX** - Web dashboard for database management
4. **Easy migration** - Can export to RDS later if needed

---

## Intentional Deferrals

Some features were intentionally NOT built:

| Feature | Status | Reason |
|---------|--------|--------|
| **WebSockets** | Deferred | No clear use case identified |
| **Kubernetes** | Deferred | Not needed until production scale |
| **Grafana Alerting** | Deferred | Prometheus metrics added, alerting later |
| **Background Workers** | Deferred | Async goroutines sufficient for now |
| **ML Priority** | Deferred | Need 6+ months of data first |
| **Natural Language Input** | Optional | Marked as nice-to-have |

### Why Document Deferrals?

1. **Prevents "why didn't we..." questions**
2. **Shows intentional decision vs. oversight**
3. **Provides context for future work**
4. **Reduces scope creep**

---

## Security Decisions

### JWT Algorithm Validation (PR #74)

**Context:** JWT libraries accept multiple algorithms. An attacker could forge tokens using the "none" algorithm.

**Decision:** Explicitly validate HMAC-SHA256:

```go
if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
    return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
}
```

### Rate Limiting with Fallback (PR #15)

**Context:** Need to prevent abuse. Redis is ideal but adds infrastructure.

**Decision:** Redis-based rate limiting with in-memory fallback:

```go
redisLimiter, err := ratelimit.NewRedisLimiter(cfg.RedisURL)
if err != nil {
    slog.Warn("Redis unavailable, using in-memory rate limiting")
    redisLimiter = nil  // Falls back to in-memory
}
```

**Tradeoff:** In-memory rate limiting doesn't work across multiple server instances, but acceptable for MVP.

---

## Decision Table Template

Use this template for your own decisions:

```markdown
## Decision: [Title]

### Context
[What problem are we solving?]

### Options
| Aspect | Option A | Option B | Option C |
|--------|----------|----------|----------|
| [Criteria 1] | | | |
| [Criteria 2] | | | |
| [Criteria 3] | | | |

### Decision
**Chose [Option] because:**
1. [Reason 1]
2. [Reason 2]

### Consequences
**Accepted tradeoffs:**
- [Tradeoff 1]
- [Tradeoff 2]

**Mitigations:**
- [How we handle the tradeoffs]
```

---

## Exercises

### 🔰 Beginner: Document a Decision

Document a technology decision from your own project using the template above.

### 🎯 Intermediate: Challenge a Decision

Pick one of TaskFlow's decisions and argue for the alternative. What would you gain? What would you lose?

### 🚀 Advanced: Design a New Feature

Design a "task sharing" feature. Document:
- Context (what problem)
- Options (at least 3)
- Decision (with criteria)
- Consequences (tradeoffs)

---

## Reflection Questions

1. **When should you NOT document decisions?** What decisions are too trivial?

2. **How do you handle changing requirements?** When do past decisions need revisiting?

3. **Who should be involved in technical decisions?** When do you need buy-in?

4. **How do you evaluate decisions in hindsight?** What signals indicate a bad decision?

---

## Key Takeaways

1. **Document the "why", not just the "what".** Future you will thank you.

2. **Evaluate tradeoffs explicitly.** Every decision has costs.

3. **Defer intentionally.** "Not now" is different from "forgot".

4. **Decisions are reversible.** Most can be changed later.

5. **Context matters.** The right decision for an MVP may differ from a mature product.

---

## Next Module

Continue to **[Module 09: Lessons Learned](./09-lessons-learned.md)** for a synthesis of best practices and reusable patterns.
