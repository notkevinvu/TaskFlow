# Data Model: Archive Completed Tasks

**Date**: 2025-12-02
**Feature**: 001-archive-completed-tasks

## Overview

This feature requires no database schema changes. All data models are either existing (Task) or client-side state management (filter state, selection state).

---

## Existing Entities (No Changes)

### Task

The existing Task entity is sufficient. Key fields for archive feature:

| Field | Type | Usage in Archive |
|-------|------|------------------|
| `id` | UUID | Selection, bulk operations |
| `status` | enum(todo, in_progress, done) | Filter completed (`done`) |
| `title` | string | Display, search |
| `description` | string | Display, search |
| `category` | string (nullable) | Filter by category |
| `updated_at` | timestamp | Sort by completion time |
| `user_id` | UUID | Authorization |

---

## New Backend DTOs

### BulkOperationRequest

Request body for bulk delete and bulk restore operations.

```go
// domain/task.go

// BulkOperationRequest represents a request to perform bulk operations on tasks
type BulkOperationRequest struct {
    TaskIDs []string `json:"task_ids" binding:"required,min=1,max=100,dive,uuid"`
}
```

**Validation Rules**:
- `task_ids`: Required, array of 1-100 UUIDs
- Max 100 tasks per bulk operation (prevents abuse)
- Each ID must be valid UUID format

### BulkOperationResponse

Response for bulk operations.

```go
// domain/task.go

// BulkOperationResponse represents the result of a bulk operation
type BulkOperationResponse struct {
    SuccessCount int      `json:"success_count"`
    FailedIDs    []string `json:"failed_ids,omitempty"`
    Message      string   `json:"message"`
}
```

**Response Fields**:
- `success_count`: Number of tasks successfully processed
- `failed_ids`: IDs of tasks that failed (not found, not owned, already in target state)
- `message`: Human-readable summary

---

## Frontend State Models

### ArchiveFilterState

Client-side state for archive page filtering.

```typescript
// types/archive.ts

interface ArchiveFilterState {
  search: string;           // Text search query
  category: string | null;  // Selected category filter
  dateStart: string | null; // YYYY-MM-DD completion date start
  dateEnd: string | null;   // YYYY-MM-DD completion date end
  page: number;             // Current page (1-indexed for UI)
  pageSize: number;         // Items per page (default: 20)
}

const defaultArchiveFilters: ArchiveFilterState = {
  search: '',
  category: null,
  dateStart: null,
  dateEnd: null,
  page: 1,
  pageSize: 20,
};
```

### SelectionState

Client-side state for bulk selection.

```typescript
// types/archive.ts

interface SelectionState {
  selectedIds: Set<string>;   // Currently selected task IDs
  isAllSelected: boolean;     // Header checkbox state (current page)
}

// Helper functions
function toggleSelection(state: SelectionState, taskId: string): SelectionState;
function selectAll(state: SelectionState, taskIds: string[]): SelectionState;
function clearSelection(): SelectionState;
```

### DashboardViewState

Client-side state for dashboard Active/Completed toggle.

```typescript
// Already managed via URL query params or local state

type DashboardTab = 'active' | 'completed';

// URL-based: ?tab=active or ?tab=completed
// OR React state: const [tab, setTab] = useState<DashboardTab>('active');
```

---

## API Request/Response Types

### Frontend TypeScript Types

```typescript
// lib/api.ts additions

interface BulkOperationRequest {
  task_ids: string[];
}

interface BulkOperationResponse {
  success_count: number;
  failed_ids?: string[];
  message: string;
}

// API methods
taskAPI.bulkDelete(taskIds: string[]): Promise<BulkOperationResponse>;
taskAPI.bulkRestore(taskIds: string[]): Promise<BulkOperationResponse>;
```

---

## State Transitions

### Task Status Transitions (Existing)

```
┌──────────┐     complete()     ┌──────────┐
│   todo   │ ─────────────────► │   done   │
└──────────┘                    └──────────┘
      │                              │
      │ update(status)               │ bulkRestore()
      ▼                              ▼
┌──────────────┐                ┌──────────┐
│ in_progress  │ ──────────────►│   todo   │
└──────────────┘   complete()   └──────────┘
```

**New Transition** (Bulk Restore):
- `done` → `todo`: Restores completed task to active state
- Does NOT restore to `in_progress` (simplest approach)

---

## Validation Rules Summary

| Operation | Validation |
|-----------|------------|
| Bulk Delete | 1-100 task IDs, all must be user's tasks |
| Bulk Restore | 1-100 task IDs, all must be user's tasks with status=done |
| Archive Filter | Search max 200 chars, page >= 1, pageSize 10-100 |
| Selection | Max tasks selected = pageSize (current page only) |

---

## Indexes (Existing, No Changes)

The following existing indexes support archive queries:

- `idx_tasks_user_id` - Fast user filtering
- `idx_tasks_status` - Fast status filtering
- `idx_tasks_updated_at` - Fast sorting by completion time
- `idx_tasks_category` - Fast category filtering
- Full-text search index on title, description, context
