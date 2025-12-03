# Tasks: Archive Completed Tasks

**Input**: Design documents from `/specs/001-archive-completed-tasks/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/bulk-operations.yaml

**Tests**: Tests included for backend bulk operations (handler tests following existing patterns).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Backend**: `backend/internal/` (handler, service, repository, domain)
- **Frontend**: `frontend/app/`, `frontend/components/`, `frontend/hooks/`, `frontend/lib/`
- Following existing TaskFlow web application structure from plan.md

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Install any missing dependencies and prepare for implementation

- [ ] T001 Verify shadcn/ui Checkbox component exists, install if needed via `npx shadcn@latest add checkbox` in frontend/
- [ ] T002 [P] Create archive components directory at frontend/components/archive/

---

## Phase 2: Foundational (Backend Bulk Operations)

**Purpose**: Backend endpoints that MUST be complete before archive page bulk actions work

**⚠️ CRITICAL**: Archive page bulk actions (User Story 4) depend on these endpoints

### Backend DTOs

- [ ] T003 [P] Add BulkOperationRequest and BulkOperationResponse DTOs in backend/internal/domain/task.go

### Backend Repository

- [ ] T004 [P] Add BulkDelete method in backend/internal/repository/task_repository.go
- [ ] T005 [P] Add BulkUpdateStatus method in backend/internal/repository/task_repository.go

### Backend Service

- [ ] T006 Add BulkDelete method in backend/internal/service/task_service.go (depends on T004)
- [ ] T007 Add BulkRestore method in backend/internal/service/task_service.go (depends on T005)

### Backend Handlers

- [ ] T008 Add BulkDelete handler in backend/internal/handler/task_handler.go (depends on T006)
- [ ] T009 Add BulkRestore handler in backend/internal/handler/task_handler.go (depends on T007)

### Backend Routes

- [ ] T010 Register POST /api/v1/tasks/bulk-delete route in backend/cmd/server/main.go (depends on T008)
- [ ] T011 Register POST /api/v1/tasks/bulk-restore route in backend/cmd/server/main.go (depends on T009)

### Backend Tests

- [ ] T012 [P] Add TestBulkDelete handler test in backend/internal/handler/task_handler_test.go
- [ ] T013 [P] Add TestBulkRestore handler test in backend/internal/handler/task_handler_test.go

### Frontend API Client

- [ ] T014 [P] Add taskAPI.bulkDelete method in frontend/lib/api.ts
- [ ] T015 [P] Add taskAPI.bulkRestore method in frontend/lib/api.ts

### Frontend Hooks

- [ ] T016 Add useBulkDelete mutation hook in frontend/hooks/useTasks.ts (depends on T014)
- [ ] T017 Add useBulkRestore mutation hook in frontend/hooks/useTasks.ts (depends on T015)

**Checkpoint**: Backend bulk operations ready - bulk actions in archive page can now work

---

## Phase 3: User Story 1 - View Active Tasks Only (Priority: P1) 🎯 MVP

**Goal**: Dashboard shows only active tasks (todo, in_progress) by default, hiding completed tasks

**Independent Test**: Log in, view dashboard - should see only active tasks, no completed tasks visible

### Implementation for User Story 1

- [ ] T018 [US1] Modify dashboard page to filter tasks to status todo/in_progress by default in frontend/app/(dashboard)/dashboard/page.tsx
- [ ] T019 [US1] Add empty state message for "no active tasks" when all tasks are completed in frontend/app/(dashboard)/dashboard/page.tsx

**Checkpoint**: Dashboard now shows only active tasks by default - core value delivered

---

## Phase 4: User Story 2 - Toggle Completed on Dashboard (Priority: P2)

**Goal**: Users can toggle between Active and Completed task views on the dashboard

**Independent Test**: On dashboard, click Completed tab - should show completed tasks sorted by most recent

### Implementation for User Story 2

- [ ] T020 [US2] Import and add shadcn Tabs component wrapper to dashboard in frontend/app/(dashboard)/dashboard/page.tsx
- [ ] T021 [US2] Create Active tab content showing filtered active tasks in frontend/app/(dashboard)/dashboard/page.tsx
- [ ] T022 [US2] Create Completed tab content showing tasks with status=done sorted by updated_at DESC in frontend/app/(dashboard)/dashboard/page.tsx
- [ ] T023 [US2] Add useCompletedTasks hook for fetching completed tasks in frontend/hooks/useTasks.ts
- [ ] T024 [US2] Add empty state for Completed tab when no completed tasks exist in frontend/app/(dashboard)/dashboard/page.tsx

**Checkpoint**: Dashboard toggle fully working - users can switch between active and completed views

---

## Phase 5: User Story 3 - Browse Archive Page (Priority: P3)

**Goal**: Dedicated /archive page with table, search, filters, and pagination

**Independent Test**: Navigate to /archive - should see table of completed tasks with working search and filters

### Implementation for User Story 3

- [ ] T025 [P] [US3] Create ArchiveFilters component with search, category, date range inputs in frontend/components/archive/ArchiveFilters.tsx
- [ ] T026 [P] [US3] Create ArchiveTable component with sortable columns in frontend/components/archive/ArchiveTable.tsx
- [ ] T027 [US3] Create archive page using ArchiveFilters and ArchiveTable in frontend/app/(dashboard)/archive/page.tsx
- [ ] T028 [US3] Add useArchivedTasks hook with filter parameters in frontend/hooks/useTasks.ts
- [ ] T029 [US3] Implement pagination controls in archive page in frontend/app/(dashboard)/archive/page.tsx
- [ ] T030 [US3] Add empty state and "no results found" states to archive page in frontend/app/(dashboard)/archive/page.tsx

**Checkpoint**: Archive page browsing complete - users can search, filter, and paginate completed tasks

---

## Phase 6: User Story 4 - Bulk Actions on Archive (Priority: P4)

**Goal**: Users can select multiple tasks and perform bulk delete or restore actions

**Independent Test**: On archive page, select 3 tasks, click "Delete Selected" - should delete after confirmation

### Implementation for User Story 4

- [ ] T031 [P] [US4] Create BulkActionsBar component with Delete/Restore buttons in frontend/components/archive/BulkActionsBar.tsx
- [ ] T032 [US4] Add checkbox column and selection state to ArchiveTable in frontend/components/archive/ArchiveTable.tsx
- [ ] T033 [US4] Add select all checkbox in table header in frontend/components/archive/ArchiveTable.tsx
- [ ] T034 [US4] Integrate BulkActionsBar into archive page with selection state in frontend/app/(dashboard)/archive/page.tsx
- [ ] T035 [US4] Add confirmation AlertDialog for bulk delete action in frontend/app/(dashboard)/archive/page.tsx
- [ ] T036 [US4] Wire up bulk delete action with useBulkDelete hook in frontend/app/(dashboard)/archive/page.tsx
- [ ] T037 [US4] Wire up bulk restore action with useBulkRestore hook in frontend/app/(dashboard)/archive/page.tsx
- [ ] T038 [US4] Add toast notifications for bulk action success/failure in frontend/app/(dashboard)/archive/page.tsx

**Checkpoint**: Bulk actions complete - users can efficiently manage large archives

---

## Phase 7: User Story 5 - Navigate to Archive (Priority: P5)

**Goal**: Archive page accessible from sidebar navigation

**Independent Test**: On any page, look at sidebar - should see Archive link, click to navigate

### Implementation for User Story 5

- [ ] T039 [US5] Add Archive navigation link with Archive icon to sidebar in frontend/app/(dashboard)/layout.tsx
- [ ] T040 [US5] Add active state highlighting for Archive link when on /archive route in frontend/app/(dashboard)/layout.tsx

**Checkpoint**: Navigation complete - users can easily find and access the archive

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Final polish and documentation updates

- [ ] T041 [P] Run go build and go test to verify backend changes in backend/
- [ ] T042 [P] Run npm run build to verify frontend compiles without errors in frontend/
- [ ] T043 Update docs/design-system.md with any new UI patterns used (if applicable)
- [ ] T044 Run quickstart.md validation steps to verify feature works end-to-end

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies - can start immediately
- **Phase 2 (Foundational)**: Depends on Setup - BLOCKS User Story 4 bulk actions
- **Phase 3 (US1)**: Can start after Setup (no backend dependency)
- **Phase 4 (US2)**: Can start after Setup (no backend dependency)
- **Phase 5 (US3)**: Can start after Setup (no backend dependency)
- **Phase 6 (US4)**: Depends on Phase 2 (Foundational) completion for bulk operations
- **Phase 7 (US5)**: Can start after Setup (no dependency on other stories)
- **Phase 8 (Polish)**: Depends on all user stories being complete

### User Story Dependencies

| Story | Depends On | Can Start After |
|-------|------------|-----------------|
| US1 (P1) | Setup only | Phase 1 |
| US2 (P2) | US1 (builds on dashboard) | Phase 3 |
| US3 (P3) | Setup only | Phase 1 |
| US4 (P4) | Foundational + US3 | Phase 2 + Phase 5 |
| US5 (P5) | Setup only | Phase 1 |

### Within Each User Story

- Frontend components before page integration
- API client methods before hooks
- Hooks before page usage
- Core implementation before polish

### Parallel Opportunities

**Phase 2 (Backend)**:
```
Parallel: T003, T004, T005 (different files)
Parallel: T012, T013 (test files)
Parallel: T014, T015 (API methods)
```

**Phase 5 (Archive Page)**:
```
Parallel: T025, T026 (different components)
```

**Phase 6 (Bulk Actions)**:
```
Parallel: T031 (BulkActionsBar is independent initially)
```

---

## Parallel Example: Foundational Backend

```bash
# Launch backend repository methods together:
Task: "Add BulkDelete method in backend/internal/repository/task_repository.go"
Task: "Add BulkUpdateStatus method in backend/internal/repository/task_repository.go"

# Launch backend tests together (after handlers):
Task: "Add TestBulkDelete handler test in backend/internal/handler/task_handler_test.go"
Task: "Add TestBulkRestore handler test in backend/internal/handler/task_handler_test.go"

# Launch frontend API methods together:
Task: "Add taskAPI.bulkDelete method in frontend/lib/api.ts"
Task: "Add taskAPI.bulkRestore method in frontend/lib/api.ts"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 3: User Story 1 (dashboard shows active only)
3. **STOP and VALIDATE**: Test that completed tasks are hidden
4. Deploy/demo if ready - core value delivered!

### Incremental Delivery

1. Setup → Done
2. US1 (active only dashboard) → Test → **MVP deployed**
3. US2 (dashboard toggle) → Test → Enhanced dashboard
4. Backend Foundational → Ready for bulk ops
5. US3 (archive page) → Test → Full archive browsing
6. US4 (bulk actions) → Test → Power user features
7. US5 (navigation) → Test → Complete feature
8. Polish → Production ready

### Recommended Order (Single Developer)

1. T001-T002 (Setup)
2. T018-T019 (US1 - MVP)
3. T020-T024 (US2 - Dashboard toggle)
4. T003-T017 (Backend bulk ops)
5. T025-T030 (US3 - Archive page)
6. T031-T038 (US4 - Bulk actions)
7. T039-T040 (US5 - Navigation)
8. T041-T044 (Polish)

---

## Summary

| Phase | Story | Task Count | Parallel Tasks |
|-------|-------|------------|----------------|
| 1 | Setup | 2 | 1 |
| 2 | Foundational | 15 | 7 |
| 3 | US1 (P1) | 2 | 0 |
| 4 | US2 (P2) | 5 | 0 |
| 5 | US3 (P3) | 6 | 2 |
| 6 | US4 (P4) | 8 | 1 |
| 7 | US5 (P5) | 2 | 0 |
| 8 | Polish | 4 | 2 |
| **Total** | | **44** | **13** |

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story is independently testable after completion
- Backend bulk operations are foundational but only required for US4
- US1-US3 and US5 can proceed without waiting for backend bulk ops
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
