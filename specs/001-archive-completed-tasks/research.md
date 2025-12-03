# Research: Archive Completed Tasks

**Date**: 2025-12-02
**Feature**: 001-archive-completed-tasks

## Research Questions & Findings

### 1. Does the existing API support status filtering for completed tasks?

**Decision**: Yes, existing API is sufficient for fetching completed tasks

**Rationale**:
- `GET /api/v1/tasks?status=done` already works
- TaskStatus enum includes: `todo`, `in_progress`, `done`
- Pagination via `limit` and `offset` parameters already supported
- Full-text search via `search` parameter already supported
- Category filtering via `category` parameter already supported

**Evidence**:
- `backend/internal/handler/task_handler.go` lines 75-78 parse status query param
- `backend/internal/domain/task.go` defines TaskStatus with valid values
- `backend/internal/repository/task_repository.go` applies status filter in SQL

**Alternatives Considered**:
- Creating new "archived" status → Rejected: Adds complexity, existing "done" status is sufficient
- Creating separate archive table → Rejected: Over-engineering, status filtering is simpler

---

### 2. Do bulk operation endpoints exist?

**Decision**: No, new endpoints required for bulk delete and bulk restore

**Rationale**:
- Only single-task operations exist: `DELETE /tasks/:id`, `PUT /tasks/:id`
- Existing bulk operations are category-focused only (rename, delete category)
- Need to add `POST /api/v1/tasks/bulk-delete` and `POST /api/v1/tasks/bulk-restore`

**Evidence**:
- `backend/cmd/server/main.go` route registration shows no bulk task endpoints
- `backend/internal/handler/task_handler.go` has no bulk methods
- Category handler has bulk operations pattern that can be followed

**Alternatives Considered**:
- Multiple individual API calls from frontend → Rejected: Poor UX, N+1 requests
- Batch endpoint with generic operations → Rejected: Over-engineering for two simple operations

---

### 3. What shadcn/ui components are available for the archive table?

**Decision**: Use existing Tabs, Table, Checkbox, Dialog components

**Rationale**:
- `Tabs` component exists at `frontend/components/ui/tabs.tsx`
- `Table` component exists at `frontend/components/ui/table.tsx`
- `Checkbox` needs to be verified/installed
- `Dialog` and `AlertDialog` exist for confirmation modals

**Evidence**:
- Glob search found 15+ shadcn components in `frontend/components/ui/`
- Table component already defined with TableHeader, TableBody, TableRow, TableCell
- Tabs component already used in other parts of the app

**Alternatives Considered**:
- Build custom table with checkboxes → Rejected: Reinventing wheel
- Use third-party data table (TanStack) → Rejected: Adds dependency, shadcn Table sufficient

---

### 4. How should completed tasks be sorted?

**Decision**: Sort by `updated_at` DESC (most recently completed first)

**Rationale**:
- The `updated_at` field is automatically updated when status changes to "done"
- No separate `completed_at` field exists in current schema
- Most recent completions are most relevant for users

**Evidence**:
- Task entity has `UpdatedAt time.Time` field
- PostgreSQL trigger updates `updated_at` on any row modification
- Schema shows `updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()`

**Alternatives Considered**:
- Add dedicated `completed_at` column → Rejected: Schema change, YAGNI for now
- Sort by priority_score → Rejected: Irrelevant for completed tasks

---

### 5. How to handle the dashboard filtering logic?

**Decision**: Fetch with status filter, not client-side filtering

**Rationale**:
- Server-side filtering is more efficient for large task lists
- Already supported by existing API
- Active tasks: `?status=todo` OR `?status=in_progress` (need two calls or API enhancement)

**Implementation Note**:
- Currently API only supports single status filter
- For dashboard active view, may need to fetch all and filter client-side
- OR make two API calls (one for todo, one for in_progress)
- Simplest: Fetch all tasks, filter client-side for dashboard (limited scope)

**Alternatives Considered**:
- Modify API to support multiple status values → Could be future enhancement
- Always fetch all tasks → Current approach, works for reasonable task counts

---

### 6. Sidebar navigation structure

**Decision**: Add "Archive" link below existing navigation items

**Rationale**:
- Sidebar exists at `frontend/app/(dashboard)/layout.tsx`
- Navigation items use consistent pattern with icons
- Archive fits naturally after Dashboard and Analytics

**Evidence**:
- Layout file shows navigation with Link components
- Active state highlighting based on pathname
- lucide-react icons used throughout

**Implementation**:
```tsx
<Link href="/archive" className={...}>
  <Archive className="w-4 h-4" />
  Archive
</Link>
```

---

## Summary of Technical Decisions

| Area | Decision | Impact |
|------|----------|--------|
| Status filtering | Use existing `?status=done` API | No backend changes for listing |
| Bulk operations | Add two new endpoints | ~100 lines backend code |
| UI components | shadcn Tabs, Table, Checkbox | No new dependencies |
| Sorting | Use `updated_at` DESC | No schema changes |
| Dashboard filter | Client-side for active toggle | Simple, performant for typical usage |
| Navigation | Add Archive link to sidebar | ~10 lines frontend code |

## Open Questions (Resolved)

- ~~Does Checkbox component exist?~~ → Verify during implementation, add via shadcn CLI if needed
- ~~Multiple status filter support?~~ → Use client-side filtering for dashboard (acceptable tradeoff)
