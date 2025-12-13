# Bug Backlog Fix Plan

**Created:** 2025-12-07
**Status:** Investigation Complete, Ready for Implementation
**Branch:** TBD (suggest: `fix/bug-backlog-batch-1`)

---

## Executive Summary

Investigation of 6 bugs from `PROJECT_STATUS.md` bug backlog. All bugs have been analyzed with root causes identified and fixes documented.

| Bug | Severity | Status | Complexity |
|-----|----------|--------|------------|
| #1 Task Completion Latency | 🔴 High | Root cause found | Medium |
| #2 Completed Tasks on Calendar | 🟡 Medium | Root cause found | Low |
| #3 Keyboard Shortcuts Scope | 🟡 Medium | Root cause found | Low |
| #4 Missing cursor:pointer | 🟢 Low | Root cause found | Low |
| #5 Template Field Inheritance | 🔴 High | Needs verification | Unknown |
| #6 Dialog Overflow Issues | 🟡 Medium | Root cause found | Low |

---

## Bug #1: Task Completion Latency (2+ sec delay)

### Severity: 🔴 High

### Symptoms
- Noticeable 2+ second delay when completing tasks
- UI feels unresponsive after clicking "Complete"

### Root Cause Analysis

The task completion flow triggers an excessive number of database operations:

#### Backend Operations (per completion)

**File:** `backend/internal/service/task_service.go:331-413`

```go
// Core operations (5-7 DB calls):
1. taskRepo.FindByID()                    // 1 DB call
2. subtaskService.ValidateParentCompletion() // 1+ DB calls (if subtasks exist)
3. dependencyService.ValidateCompletion()    // 1+ DB calls (if dependencies exist)
4. taskRepo.Update()                      // 1 DB call
5. logHistorySimple()                     // 1 DB call
6. recurrenceService.GenerateNextTask()  // 1+ DB calls (if recurring)
7. gamificationService.ProcessTaskCompletion() // See below
```

**File:** `backend/internal/service/gamification_service.go:80-131`

```go
// Gamification processing (8-15+ DB calls):
func ProcessTaskCompletion():
  1. GetStats()                           // 1 DB call
  2. IncrementCategoryMastery()           // 1 DB call (if category exists)

  // ComputeStats() internally calls:
  3. GetUserTimezone()                    // 1 DB call
  4. GetTotalCompletedTasks()             // 1 DB call
  5. GetCompletionsByDate() (365 days)    // 1 DB call
  6. GetCompletionStats()                 // 1 DB call
  7. GetOnTimeCompletionRate()            // 1 DB call
  8. GetEffortDistribution()              // 1 DB call

  // CheckAndAwardAchievements() internally calls:
  9-N. HasAchievement() per milestone     // N DB calls (checking each threshold)
  N+1. GetCategoryMastery()               // 1 DB call
  N+2. GetSpeedCompletions()              // 1 DB call
  N+3. GetUserTimezone() (again!)         // 1 DB call (duplicate)
  N+4. GetWeeklyCompletionDays()          // 1 DB call

  10. UpsertStats()                       // 1 DB call
```

**Total: 15-25+ database operations per task completion**

#### Frontend Operations

**File:** `frontend/hooks/useTasks.ts:92-117`

```typescript
export function useCompleteTask() {
  return useMutation({
    mutationFn: (id: string) => taskAPI.complete(id),
    onSuccess: (response) => {
      // Problem 1: Invalidates ALL task queries (no staleTime)
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      // Problem 2: Invalidates gamification queries
      queryClient.invalidateQueries({ queryKey: gamificationKeys.all });
      // ... achievement toast logic
    },
  });
}
```

**File:** `frontend/hooks/useTasks.ts:20-32`

```typescript
export function useTasks(filters?: TaskFilters) {
  return useQuery({
    queryKey: ['tasks', filters],
    queryFn: async () => { ... },
    // Problem: No staleTime - refetches immediately on invalidation
  });
}
```

### Recommended Fix

#### Phase 1: Frontend Optimistic Updates (Immediate improvement)

```typescript
// frontend/hooks/useTasks.ts
export function useCompleteTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => taskAPI.complete(id),
    // Add optimistic update for instant feedback
    onMutate: async (taskId) => {
      await queryClient.cancelQueries({ queryKey: ['tasks'] });
      const previousTasks = queryClient.getQueryData(['tasks']);

      // Optimistically update task status
      queryClient.setQueryData(['tasks'], (old: any) => ({
        ...old,
        tasks: old.tasks.map((t: any) =>
          t.id === taskId ? { ...t, status: 'done' } : t
        ),
      }));

      return { previousTasks };
    },
    onError: (err, taskId, context) => {
      // Rollback on error
      queryClient.setQueryData(['tasks'], context?.previousTasks);
    },
    onSettled: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      queryClient.invalidateQueries({ queryKey: gamificationKeys.all });
    },
  });
}
```

#### Phase 2: Add staleTime to reduce refetches

```typescript
// frontend/hooks/useTasks.ts
export function useTasks(filters?: TaskFilters) {
  return useQuery({
    queryKey: ['tasks', filters],
    queryFn: async () => { ... },
    staleTime: 5000, // Don't refetch within 5 seconds
  });
}
```

#### Phase 3: Backend optimization (Future)

- Batch gamification queries into fewer database calls
- Cache frequently accessed data (streaks, stats) in memory/Redis
- Consider async gamification processing (message queue)

### Files to Modify

1. `frontend/hooks/useTasks.ts` - Add optimistic updates + staleTime

### Testing

1. Complete a task and verify instant UI update
2. Verify gamification stats still update correctly
3. Verify error rollback works if completion fails

---

## Bug #2: Completed Tasks Showing on Calendar

### Severity: 🟡 Medium

### Symptoms
- Calendar shows completed tasks alongside active tasks
- Confusing UX as users see "done" tasks on future dates

### Root Cause Analysis

**File:** `frontend/components/Calendar.tsx:28-31`

```typescript
const { data: calendarData, isLoading, error } = useCalendarTasks({
  start_date: format(calendarStart, 'yyyy-MM-dd'),
  end_date: format(calendarEnd, 'yyyy-MM-dd'),
  // ❌ Missing status filter!
});
```

The calendar does not pass a `status` parameter, so the backend returns ALL tasks.

**File:** `backend/internal/repository/task_repository.go:514-527`

```go
// Optional status filter - only applied when filter.Status has values
if len(filter.Status) > 0 {
  statusConditions := ""
  for i, status := range filter.Status {
    // ... builds WHERE clause for status
  }
  query += fmt.Sprintf(" AND (%s)", statusConditions)
}
```

### Recommended Fix

**File:** `frontend/components/Calendar.tsx`

```typescript
// Line 28-31: Add status filter
const { data: calendarData, isLoading, error } = useCalendarTasks({
  start_date: format(calendarStart, 'yyyy-MM-dd'),
  end_date: format(calendarEnd, 'yyyy-MM-dd'),
  status: 'todo,in_progress', // Only show active tasks
});
```

### Files to Modify

1. `frontend/components/Calendar.tsx` - Add status filter

### Testing

1. Complete a task with a due date
2. Verify it no longer appears on the calendar
3. Verify active tasks still appear correctly

---

## Bug #3: Keyboard Shortcuts Only Work on Dashboard

### Severity: 🟡 Medium

### Symptoms
- Cmd/Ctrl+K, ?, and other shortcuts only work when on /dashboard
- Don't work on /settings, /archive, or other pages

### Root Cause Analysis

**File:** `frontend/app/(dashboard)/dashboard/page.tsx:119`

```typescript
// Keyboard shortcuts hook is only called here
useGlobalKeyboardShortcuts();
```

The hook is called in the dashboard page component, not in a shared layout. Other pages don't invoke this hook.

### Recommended Fix

Move the hook to the dashboard layout (which wraps all authenticated pages).

**File:** `frontend/app/(dashboard)/layout.tsx`

```typescript
// Add import
import { useGlobalKeyboardShortcuts } from '@/hooks/useGlobalKeyboardShortcuts';

// Inside the component, before return:
export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  // ... existing state and hooks

  // Add this line
  useGlobalKeyboardShortcuts();

  // ... rest of component
}
```

Then remove from `dashboard/page.tsx`:
```typescript
// Remove this line from dashboard/page.tsx
// useGlobalKeyboardShortcuts();
```

### Files to Modify

1. `frontend/app/(dashboard)/layout.tsx` - Add hook call
2. `frontend/app/(dashboard)/dashboard/page.tsx` - Remove hook call

### Testing

1. Navigate to /settings and press Cmd/Ctrl+K
2. Verify Quick Add dialog opens
3. Press ? anywhere and verify help dialog opens
4. Test on /archive and other pages

---

## Bug #4: Missing cursor:pointer on Sidebar Buttons

### Severity: 🟢 Low

### Symptoms
- Sidebar template buttons don't show pointer cursor on hover
- Inconsistent with other interactive elements

### Root Cause Analysis

**File:** `frontend/components/ui/button.tsx:7-8`

```typescript
const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap ...",
  // ❌ Missing: cursor-pointer
```

The shadcn Button component doesn't include `cursor-pointer` in its base styles. While native `<button>` elements typically show pointer cursor, Tailwind's CSS reset normalizes this behavior.

**File:** `frontend/app/(dashboard)/layout.tsx:190-205`

```typescript
// These buttons don't have explicit cursor-pointer
<Button
  variant="ghost"
  className="w-full justify-start gap-2 h-9 ..."
  onClick={() => setTemplatePickerOpen(true)}
>
```

### Recommended Fix

Add `cursor-pointer` to the base Button component:

**File:** `frontend/components/ui/button.tsx`

```typescript
const buttonVariants = cva(
  "cursor-pointer inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-all disabled:pointer-events-none disabled:opacity-50 ...",
```

### Files to Modify

1. `frontend/components/ui/button.tsx` - Add cursor-pointer to base styles

### Testing

1. Hover over sidebar template buttons
2. Verify pointer cursor appears
3. Check other buttons still work correctly

---

## Bug #5: Template Not Inheriting Fields

### Severity: 🔴 High (Needs Manual Verification)

### Symptoms (Reported)
- Creating task from template doesn't pre-fill description, category, context

### Investigation Results

Code review shows the data flow appears correct:

**File:** `frontend/hooks/useTemplates.ts:192-211`

```typescript
export function templateToFormValues(template: TaskTemplate): Partial<CreateTaskDTO> {
  return {
    title: template.title,
    description: template.description,      // ✓ Included
    category: template.category,            // ✓ Included
    estimated_effort: template.estimated_effort,
    user_priority: template.user_priority !== 5 ? template.user_priority : undefined,
    context: template.context,              // ✓ Included
    related_people: template.related_people,
    due_date: dueDate,
  };
}
```

**File:** `frontend/components/CreateTaskDialog.tsx:88-102`

```typescript
if (initialValues) {
  setFormData({
    title: initialValues.title || '',
    description: initialValues.description || '',     // ✓ Uses value
    category: initialValues.category || '',           // ✓ Uses value
    estimated_effort: initialValues.estimated_effort || 'medium',
    user_priority: initialValues.user_priority || 5,
    due_date: formatDateForInput(initialValues.due_date),
    context: initialValues.context || '',             // ✓ Uses value
    recurrence: null,
  });
}
```

Backend also correctly saves and retrieves all fields.

### Recommended Action

1. **Manual verification required** - Test the template flow:
   - Create a template with description, category, context filled
   - Use "Create from Template"
   - Verify if fields pre-fill

2. If bug confirmed, possible causes:
   - Race condition in state updates
   - Template data not being fetched correctly
   - Browser caching stale data

### Files to Check

1. `frontend/hooks/useTemplates.ts`
2. `frontend/components/TemplatePickerDialog.tsx`
3. `frontend/components/CreateTaskDialog.tsx`
4. `frontend/app/(dashboard)/layout.tsx` (state management)

---

## Bug #6: Dialog Overflow/Scroll Issues

### Severity: 🟡 Medium

### Symptoms
- Dialogs with many fields overflow on smaller screens
- Cannot scroll to see all form fields
- Footer buttons may be cut off

### Root Cause Analysis

**File:** `frontend/components/ui/dialog.tsx:62-64`

```typescript
<DialogPrimitive.Content
  className={cn(
    "bg-background ... fixed top-[50%] left-[50%] z-50 grid w-full max-w-[calc(100%-2rem)] translate-x-[-50%] translate-y-[-50%] gap-4 rounded-lg border p-6 shadow-lg duration-200 sm:max-w-lg",
    // ❌ Missing: max-h-[90vh] overflow-y-auto
    className
  )}
```

The dialog has no height constraint (`max-h`) or overflow handling (`overflow-y-auto`).

### Recommended Fix

**File:** `frontend/components/ui/dialog.tsx`

```typescript
<DialogPrimitive.Content
  className={cn(
    "bg-background data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0 data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95 fixed top-[50%] left-[50%] z-50 grid w-full max-w-[calc(100%-2rem)] max-h-[90vh] overflow-y-auto translate-x-[-50%] translate-y-[-50%] gap-4 rounded-lg border p-6 shadow-lg duration-200 sm:max-w-lg",
    className
  )}
```

Key additions:
- `max-h-[90vh]` - Constrain height to 90% of viewport
- `overflow-y-auto` - Enable vertical scrolling when needed

### Files to Modify

1. `frontend/components/ui/dialog.tsx` - Add height constraint and overflow

### Testing

1. Open CreateTaskDialog on a small viewport (resize browser)
2. Verify dialog doesn't overflow screen
3. Verify scrolling works to access all fields
4. Verify footer buttons remain accessible

---

## Implementation Order

Recommended order based on impact and complexity:

### Batch 1 (Quick Wins - Low Complexity)
1. **Bug #4** - cursor:pointer (1 line change)
2. **Bug #6** - Dialog overflow (1 line change)
3. **Bug #2** - Calendar filter (1 line change)

### Batch 2 (Medium Complexity)
4. **Bug #3** - Keyboard shortcuts (move hook, 2 files)

### Batch 3 (Higher Complexity)
5. **Bug #1** - Task completion latency (optimistic updates)

### Batch 4 (Verification Needed)
6. **Bug #5** - Manual testing first, then fix if confirmed

---

## Commit Strategy

Following CLAUDE.md guidelines for granular commits:

```
fix(ui): Add cursor-pointer to base Button component

Fixes missing pointer cursor on sidebar template buttons and
ensures consistent hover behavior across all buttons.

Bug #4 from bug backlog.
```

```
fix(ui): Add max-height and overflow to DialogContent

Prevents dialog overflow on small viewports by constraining
height to 90vh and enabling vertical scrolling.

Bug #6 from bug backlog.
```

```
fix(calendar): Filter out completed tasks from calendar view

Adds status filter to calendar API call to only show
todo and in_progress tasks.

Bug #2 from bug backlog.
```

```
fix(keyboard): Move global shortcuts hook to layout

Enables keyboard shortcuts (Cmd+K, ?, P) across all
authenticated pages, not just dashboard.

Bug #3 from bug backlog.
```

```
perf(tasks): Add optimistic updates for task completion

Provides instant UI feedback when completing tasks by
optimistically updating local state before server response.
Reduces perceived latency from 2+ seconds to near-instant.

Bug #1 from bug backlog.
```

---

## Related Files Quick Reference

| Bug | Primary Files |
|-----|---------------|
| #1 | `frontend/hooks/useTasks.ts` |
| #2 | `frontend/components/Calendar.tsx` |
| #3 | `frontend/app/(dashboard)/layout.tsx`, `frontend/app/(dashboard)/dashboard/page.tsx` |
| #4 | `frontend/components/ui/button.tsx` |
| #5 | `frontend/hooks/useTemplates.ts`, `frontend/components/CreateTaskDialog.tsx` |
| #6 | `frontend/components/ui/dialog.tsx` |
