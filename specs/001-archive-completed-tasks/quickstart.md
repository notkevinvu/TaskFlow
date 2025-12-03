# Quickstart: Archive Completed Tasks

**Date**: 2025-12-02
**Feature**: 001-archive-completed-tasks
**Branch**: `001-archive-completed-tasks`

## Prerequisites

Ensure you have the development environment set up:

```bash
# Backend
cd backend
go mod download
# Ensure .env is configured with DATABASE_URL and JWT_SECRET

# Frontend
cd frontend
npm install
# Ensure .env has NEXT_PUBLIC_API_URL=http://localhost:8080
```

## Development Workflow

### 1. Start the Development Servers

```bash
# From repo root - starts both backend and frontend
scripts/start.bat
# OR manually:
# Terminal 1: cd backend && go run cmd/server/main.go
# Terminal 2: cd frontend && npm run dev
```

### 2. Verify Existing Functionality

Before making changes, verify the current task listing works:

```bash
# List tasks with status filter (should work)
curl -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/tasks?status=done"

# Expected: JSON array of completed tasks
```

### 3. Install shadcn/ui Checkbox (if needed)

```bash
cd frontend
npx shadcn@latest add checkbox
```

## Implementation Order

### Phase 1: Backend Bulk Operations

1. **Add DTOs** (`backend/internal/domain/task.go`):
   ```go
   type BulkOperationRequest struct {
       TaskIDs []string `json:"task_ids" binding:"required,min=1,max=100"`
   }
   ```

2. **Add repository methods** (`backend/internal/repository/task_repository.go`):
   - `BulkDelete(ctx, userID, taskIDs) (int, []string, error)`
   - `BulkUpdateStatus(ctx, userID, taskIDs, newStatus) (int, []string, error)`

3. **Add service methods** (`backend/internal/service/task_service.go`)

4. **Add handlers** (`backend/internal/handler/task_handler.go`)

5. **Register routes** (`backend/cmd/server/main.go`):
   ```go
   tasks.POST("/bulk-delete", taskHandler.BulkDelete)
   tasks.POST("/bulk-restore", taskHandler.BulkRestore)
   ```

6. **Test endpoints**:
   ```bash
   # Bulk delete
   curl -X POST -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"task_ids": ["uuid1", "uuid2"]}' \
     http://localhost:8080/api/v1/tasks/bulk-delete

   # Bulk restore
   curl -X POST -H "Authorization: Bearer $TOKEN" \
     -H "Content-Type: application/json" \
     -d '{"task_ids": ["uuid1", "uuid2"]}' \
     http://localhost:8080/api/v1/tasks/bulk-restore
   ```

### Phase 2: Dashboard Toggle

1. **Modify dashboard page** (`frontend/app/(dashboard)/dashboard/page.tsx`):
   - Import and wrap content in shadcn Tabs
   - Add Active/Completed TabsTrigger
   - Filter tasks based on active tab

2. **Test in browser**:
   - Navigate to http://localhost:3000/dashboard
   - Click between Active and Completed tabs
   - Verify task lists update correctly

### Phase 3: Archive Page

1. **Create archive page** (`frontend/app/(dashboard)/archive/page.tsx`)

2. **Create components**:
   - `frontend/components/archive/ArchiveTable.tsx`
   - `frontend/components/archive/ArchiveFilters.tsx`
   - `frontend/components/archive/BulkActionsBar.tsx`

3. **Add API methods** (`frontend/lib/api.ts`):
   ```typescript
   bulkDelete: (taskIds: string[]) =>
     api.post('/api/v1/tasks/bulk-delete', { task_ids: taskIds }),
   bulkRestore: (taskIds: string[]) =>
     api.post('/api/v1/tasks/bulk-restore', { task_ids: taskIds }),
   ```

4. **Add hooks** (`frontend/hooks/useTasks.ts`):
   - `useBulkDelete()`
   - `useBulkRestore()`

5. **Test in browser**:
   - Navigate to http://localhost:3000/archive
   - Verify table displays completed tasks
   - Test search and filters
   - Test bulk selection and actions

### Phase 4: Navigation

1. **Add sidebar link** (`frontend/app/(dashboard)/layout.tsx`):
   ```tsx
   <Link href="/archive" className={...}>
     <Archive className="w-4 h-4" />
     Archive
   </Link>
   ```

2. **Test navigation**:
   - Verify Archive link appears in sidebar
   - Click link, verify navigation works
   - Verify active state highlighting

## Testing Commands

```bash
# Backend tests
cd backend
go test ./internal/handler/... -v -run TestBulk

# Frontend type check
cd frontend
npm run type-check

# Full build verification
cd frontend
npm run build
```

## Key Files to Modify

| Phase | File | Action |
|-------|------|--------|
| 1 | `backend/internal/domain/task.go` | Add DTOs |
| 1 | `backend/internal/repository/task_repository.go` | Add bulk queries |
| 1 | `backend/internal/service/task_service.go` | Add bulk methods |
| 1 | `backend/internal/handler/task_handler.go` | Add handlers |
| 1 | `backend/cmd/server/main.go` | Register routes |
| 2 | `frontend/app/(dashboard)/dashboard/page.tsx` | Add tabs |
| 3 | `frontend/app/(dashboard)/archive/page.tsx` | Create page |
| 3 | `frontend/components/archive/*.tsx` | Create components |
| 3 | `frontend/hooks/useTasks.ts` | Add hooks |
| 3 | `frontend/lib/api.ts` | Add API methods |
| 4 | `frontend/app/(dashboard)/layout.tsx` | Add nav link |

## Troubleshooting

### "Unauthorized" error on API calls
- Check JWT token is valid and not expired
- Verify `Authorization: Bearer <token>` header format

### Tasks not filtering correctly
- Check browser Network tab for actual API request
- Verify `status=done` parameter is being sent

### Bulk operations failing
- Check task IDs are valid UUIDs
- Verify tasks belong to current user
- Check server logs for detailed errors

### Checkbox component not found
- Run `npx shadcn@latest add checkbox` in frontend directory
