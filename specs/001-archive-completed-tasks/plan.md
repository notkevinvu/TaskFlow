# Implementation Plan: Archive Completed Tasks

**Branch**: `001-archive-completed-tasks` | **Date**: 2025-12-02 | **Spec**: [spec.md](../../.specify/specs/001-archive-completed-tasks/spec.md)
**Input**: Feature specification from `/specs/001-archive-completed-tasks/spec.md`

## Summary

Hide completed tasks from the main dashboard by default, providing two ways to access them: (1) a tab toggle on the dashboard for quick switching between Active/Completed views, and (2) a dedicated /archive page with full search, filter, pagination, and bulk action capabilities. The implementation is primarily frontend-focused, leveraging existing status filtering APIs, with two new backend endpoints for bulk operations.

## Technical Context

**Language/Version**: Go 1.24 (backend), TypeScript/React 19 (frontend)
**Primary Dependencies**: Gin framework, React Query, shadcn/ui (Tabs, Table, Dialog)
**Storage**: PostgreSQL via existing task table (no schema changes)
**Testing**: Go table-driven tests, React Testing Library
**Target Platform**: Web application (Next.js 16)
**Project Type**: Web (frontend + backend)
**Performance Goals**: Dashboard <2s load, bulk actions <5s for 50 tasks, archive pagination <500ms
**Constraints**: Existing API patterns, Clean Architecture, shadcn/ui design system
**Scale/Scope**: 1000+ completed tasks per user

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| I. Clean Architecture | ✅ PASS | New endpoints follow handler→service→repository pattern |
| II. Type Safety & SQL Security | ✅ PASS | Parameterized queries, typed interfaces |
| III. Test-Driven Quality | ✅ PASS | Tests required for bulk handlers, frontend components |
| IV. Design System Consistency | ✅ PASS | Using shadcn/ui Tabs, Table, existing design tokens |
| V. Simplicity & YAGNI | ✅ PASS | Minimal changes, no over-engineering |

**Code Quality Gates**:
- [ ] All tests pass
- [ ] No TypeScript errors
- [ ] Go builds without warnings
- [ ] SQL uses parameterized queries only
- [ ] UI follows design system tokens
- [ ] Changes documented in design-system.md (if new patterns)

## Project Structure

### Documentation (this feature)

```text
specs/001-archive-completed-tasks/
├── plan.md              # This file
├── research.md          # Phase 0 output - technical research
├── data-model.md        # Phase 1 output - client-side state models
├── quickstart.md        # Phase 1 output - dev setup guide
├── contracts/           # Phase 1 output - new API contracts
│   └── bulk-operations.yaml
└── tasks.md             # Phase 2 output (created by /tasks command)
```

### Source Code (repository root)

```text
backend/
├── internal/
│   ├── handler/
│   │   └── task_handler.go      # MODIFY: Add BulkDelete, BulkRestore handlers
│   ├── service/
│   │   └── task_service.go      # MODIFY: Add BulkDelete, BulkRestore methods
│   ├── repository/
│   │   └── task_repository.go   # MODIFY: Add BulkDelete, BulkRestore queries
│   └── domain/
│       └── task.go              # MODIFY: Add BulkOperationDTO
├── cmd/server/
│   └── main.go                  # MODIFY: Register new routes
└── tests/
    └── handler/
        └── task_handler_test.go # MODIFY: Add bulk operation tests

frontend/
├── app/(dashboard)/
│   ├── dashboard/
│   │   └── page.tsx             # MODIFY: Add Tabs for Active/Completed toggle
│   ├── archive/
│   │   └── page.tsx             # CREATE: Dedicated archive page
│   └── layout.tsx               # MODIFY: Add Archive link to sidebar
├── components/
│   ├── archive/
│   │   ├── ArchiveTable.tsx     # CREATE: Table with bulk selection
│   │   ├── ArchiveFilters.tsx   # CREATE: Search, category, date filters
│   │   └── BulkActionsBar.tsx   # CREATE: Delete/Restore bulk actions
│   └── ui/
│       └── (existing shadcn)    # USE: Tabs, Table, Checkbox, Dialog
├── hooks/
│   └── useTasks.ts              # MODIFY: Add useArchivedTasks, bulk mutations
└── lib/
    └── api.ts                   # MODIFY: Add bulk API methods
```

**Structure Decision**: Web application structure. Frontend receives the majority of changes (new archive page, dashboard modifications), with targeted backend additions for bulk operations only.

## Complexity Tracking

> No constitution violations. All changes follow existing patterns.

| Area | Complexity | Justification |
|------|------------|---------------|
| Backend | Low | Two new endpoints following existing patterns |
| Frontend | Medium | New page + dashboard modification + new components |
| Testing | Low | Standard handler tests + component tests |

## Implementation Phases

### Phase 1: Backend Bulk Operations (Foundation)

**Goal**: Add bulk delete and restore endpoints

**Files to modify**:
1. `backend/internal/domain/task.go` - Add `BulkOperationRequest` DTO
2. `backend/internal/repository/task_repository.go` - Add `BulkDelete`, `BulkUpdateStatus` methods
3. `backend/internal/service/task_service.go` - Add `BulkDelete`, `BulkRestore` methods
4. `backend/internal/handler/task_handler.go` - Add `BulkDelete`, `BulkRestore` handlers
5. `backend/cmd/server/main.go` - Register new routes
6. `backend/internal/handler/task_handler_test.go` - Add tests

**API Contracts**:
- `POST /api/v1/tasks/bulk-delete` - Delete multiple tasks by IDs
- `POST /api/v1/tasks/bulk-restore` - Restore multiple tasks to "todo" status

### Phase 2: Frontend Dashboard Toggle (P1+P2 User Stories)

**Goal**: Filter dashboard to active tasks, add Completed tab

**Files to modify**:
1. `frontend/app/(dashboard)/dashboard/page.tsx` - Wrap in Tabs, filter by status
2. `frontend/hooks/useTasks.ts` - Add `useCompletedTasks` hook

**Key changes**:
- Default dashboard shows `status=todo,in_progress` only
- Add shadcn Tabs: "Active" (default) | "Completed"
- Completed tab shows tasks with `status=done` sorted by `updated_at` DESC

### Phase 3: Archive Page (P3+P4 User Stories)

**Goal**: Create dedicated archive page with full capabilities

**Files to create**:
1. `frontend/app/(dashboard)/archive/page.tsx` - Archive page
2. `frontend/components/archive/ArchiveTable.tsx` - Table with checkboxes
3. `frontend/components/archive/ArchiveFilters.tsx` - Search, category, date filters
4. `frontend/components/archive/BulkActionsBar.tsx` - Bulk action buttons

**Files to modify**:
1. `frontend/hooks/useTasks.ts` - Add `useBulkDelete`, `useBulkRestore` mutations
2. `frontend/lib/api.ts` - Add `taskAPI.bulkDelete`, `taskAPI.bulkRestore`

### Phase 4: Navigation & Polish (P5 User Story)

**Goal**: Add sidebar link, empty states, final polish

**Files to modify**:
1. `frontend/app/(dashboard)/layout.tsx` - Add Archive link to sidebar nav
2. Archive components - Add empty state handling
3. Dashboard - Add empty state for "no active tasks"

## Dependencies

| Dependency | Type | Notes |
|------------|------|-------|
| shadcn/ui Tabs | Existing | Already available in components/ui |
| shadcn/ui Table | Existing | Already available in components/ui |
| shadcn/ui Checkbox | May need install | Check if available, add if not |
| React Query | Existing | Already used for all data fetching |

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Bulk delete data loss | Confirmation dialog required before delete |
| Performance with 1000+ tasks | Server-side pagination, limit 20 per page |
| Race conditions on restore | Server validates task ownership and current status |
| Breaking existing dashboard | Feature flag or gradual rollout (optional) |

## Testing Strategy

**Backend**:
- Unit tests for BulkDelete, BulkRestore handlers
- Verify authorization (user can only bulk delete own tasks)
- Verify validation (task IDs must be valid UUIDs)
- Test partial failures (some tasks don't exist)

**Frontend**:
- Component tests for ArchiveTable bulk selection
- Integration test for dashboard tab toggle
- Test empty states
- Test bulk action confirmation flow

## Success Metrics

From spec SC-001 to SC-006:
- [ ] Users find completed tasks within 30 seconds
- [ ] Dashboard loads under 2 seconds with tabs
- [ ] 90% toggle success rate on first attempt
- [ ] Bulk actions process 50 tasks in under 5 seconds
- [ ] Archive handles 1000+ tasks without degradation
