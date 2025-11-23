# Calendar Component Implementation Plan

**Feature:** Calendar view with task badges and interactive day selection

**Created:** 2025-11-23

---

## Overview

Add a calendar component to the dashboard that shows task due dates with color-coded badges. Users can click on days to view tasks or create new ones.

---

## Architecture

### Layout Strategy (Responsive)

**Wide screens (≥1280px):**
```
┌─────────────┬──────────────┬─────────────┐
│   Left      │   Calendar   │   Details   │
│   Sidebar   │   Component  │   Sidebar   │
│             │              │             │
│             │   Task List  │             │
└─────────────┴──────────────┴─────────────┘
```

**Narrow screens (<1280px):**
```
┌─────────────┬──────────────┐
│   Left      │   Calendar   │
│   Sidebar   │              │
│             │   Task List  │
│             │              │
└─────────────┴──────────────┘
```

### Badge Color Coding

Based on highest priority task on that day:
- **Red badge** (bg-red-500): Priority ≥ 70
- **Yellow badge** (bg-yellow-500): Priority 40-69
- **Blue badge** (bg-blue-500): Priority < 40

---

## Backend Implementation

### API Endpoint

**Route:** `GET /api/v1/tasks/calendar`

**Query Parameters:**
- `start_date` (required) - Start of date range (ISO 8601: YYYY-MM-DD)
- `end_date` (required) - End of date range (ISO 8601: YYYY-MM-DD)
- `status` (optional) - Filter by status (comma-separated: `pending,in_progress`)

**Response Format:**
```json
{
  "dates": {
    "2025-11-23": {
      "count": 3,
      "badge_color": "red",
      "tasks": [
        {
          "id": "uuid",
          "title": "Task name",
          "user_priority": 8,
          "calculated_priority": 72.5,
          "category": "work",
          "effort": "small",
          "status": "pending",
          "due_date": "2025-11-23T00:00:00Z"
        }
      ]
    },
    "2025-11-25": {
      "count": 1,
      "badge_color": "blue",
      "tasks": [...]
    }
  }
}
```

**Badge Color Logic:**
```go
func calculateBadgeColor(tasks []Task) string {
    maxPriority := 0.0
    for _, task := range tasks {
        if task.CalculatedPriority > maxPriority {
            maxPriority = task.CalculatedPriority
        }
    }

    if maxPriority >= 70 { return "red" }
    if maxPriority >= 40 { return "yellow" }
    return "blue"
}
```

---

## Frontend Implementation

### Dependencies

```bash
npm install react-day-picker date-fns
```

### Component Structure

```
frontend/components/
├── Calendar.tsx                    # Main wrapper with month state
├── CalendarDay.tsx                 # Day cell + badge rendering
├── CalendarTaskPopover.tsx         # Popover with mini task cards
└── EmptyDayDialog.tsx              # "Create task for [date]?" dialog

frontend/hooks/
└── useTasks.ts                     # Add useCalendarTasks hook
```

### Calendar Component Features

- Month navigation (← November 2025 →)
- Highlight today's date
- Color-coded badges on days with tasks
- Click handlers for all days (with tasks or empty)
- Loading skeleton while fetching data
- Responsive layout integration

### Task Popover Design (Mini Cards)

Each card shows:
- Task title (truncated if long)
- Priority badge (X/10 with color)
- Category badge
- Status indicator dot
- Effort indicator
- Click anywhere on card to open TaskDetailsSidebar

### Empty Day Dialog

When clicking a day with no tasks:
- "No tasks due on November 23, 2025. Create one?"
- Pre-fills due_date with selected date
- Opens CreateTaskDialog with date context

---

## Implementation Checklist

### Backend Tasks

- [ ] Create calendar endpoint handler in `backend/internal/handler/task_handler.go`
  - [ ] Add `GetTasksCalendar(c *gin.Context)` method
  - [ ] Parse start_date and end_date query params
  - [ ] Validate date range (max 90 days)
  - [ ] Call service layer to fetch tasks
- [ ] Implement service layer method in `backend/internal/service/task_service.go`
  - [ ] Add `GetCalendarTasks(ctx, userID, startDate, endDate, status)` method
  - [ ] Call repository to fetch tasks in date range
  - [ ] Group tasks by due_date
  - [ ] Calculate badge color for each date
  - [ ] Format response structure
- [ ] Add repository method in `backend/internal/repository/task_repository.go`
  - [ ] Add `GetTasksByDateRange(ctx, userID, startDate, endDate)` query
  - [ ] Use date range filter: `due_date >= $1 AND due_date <= $2`
  - [ ] Order by due_date, calculated_priority DESC
- [ ] Register route in `backend/cmd/server/main.go`
  - [ ] Add route: `tasksGroup.GET("/calendar", taskHandler.GetTasksCalendar)`
- [ ] Test endpoint with curl/Postman

### Frontend Tasks

- [ ] Install dependencies
  - [ ] Run `npm install react-day-picker date-fns`
- [ ] Create API client method in `frontend/lib/api.ts`
  - [ ] Add `getCalendarTasks(startDate, endDate, status?)` function
  - [ ] Add TypeScript types for response
- [ ] Add React Query hook in `frontend/hooks/useTasks.ts`
  - [ ] Add `useCalendarTasks(startDate, endDate)` hook
  - [ ] Configure staleTime and cacheTime
  - [ ] Integrate with existing invalidation on mutations
- [ ] Create Calendar component (`frontend/components/Calendar.tsx`)
  - [ ] Implement month state management
  - [ ] Integrate react-day-picker
  - [ ] Add month navigation buttons
  - [ ] Render day badges with color coding
  - [ ] Handle loading and error states
  - [ ] Add click handlers for days
- [ ] Create CalendarDay component (`frontend/components/CalendarDay.tsx`)
  - [ ] Render day number
  - [ ] Render badge with count and color
  - [ ] Apply today highlight styling
  - [ ] Add hover effects
- [ ] Create CalendarTaskPopover component (`frontend/components/CalendarTaskPopover.tsx`)
  - [ ] Use shadcn/ui Popover component
  - [ ] Render mini task cards
  - [ ] Show priority, category, status, effort
  - [ ] Add click handler to open TaskDetailsSidebar
- [ ] Create EmptyDayDialog component (`frontend/components/EmptyDayDialog.tsx`)
  - [ ] Use shadcn/ui Dialog component
  - [ ] Show "Create task for [date]?" message
  - [ ] Integrate with existing CreateTaskDialog
  - [ ] Pre-fill due_date field
- [ ] Update dashboard layout (`frontend/app/(dashboard)/dashboard/page.tsx`)
  - [ ] Add responsive grid for calendar placement
  - [ ] Wide screens: Calendar on left of task list
  - [ ] Narrow screens: Calendar above task list
  - [ ] Ensure proper spacing and scrolling
- [ ] Add loading states, error handling, and empty states
  - [ ] Loading skeleton for calendar
  - [ ] Error message if API fails
  - [ ] Empty state for months with no tasks
- [ ] Test responsive behavior
  - [ ] Test on wide screen (≥1280px)
  - [ ] Test on narrow screen (<1280px)
  - [ ] Test on mobile view

### Documentation & Polish

- [ ] Update design system docs (`docs/design-system.md`)
  - [ ] Document calendar badge patterns
  - [ ] Document mini card design
  - [ ] Document color coding system
  - [ ] Document responsive layout pattern
- [ ] Add calendar feature documentation
  - [ ] Usage guide for users
  - [ ] Technical architecture notes
- [ ] Update main README.md with calendar feature

---

## Technical Decisions

### Library Choice
- **react-day-picker**: Lightweight, customizable, built for React. Good for custom styling with Tailwind.

### Responsive Strategy
- Use Tailwind breakpoints: `xl:grid-cols-[280px,1fr,400px]`
- Calendar + task list stack vertically on narrow screens
- Side-by-side on wide screens

### Data Fetching
- Fetch full month at once (better UX, fewer requests)
- React Query caching with 5-minute staleTime
- Automatic refetch after task mutations (existing invalidation logic)

### Performance Considerations
- Backend limits date range to 90 days max
- Frontend debounces month navigation
- Badges calculated server-side (avoids client computation)

---

## Future Enhancements

- [ ] Drag-and-drop to reschedule tasks between days
- [ ] Week view option
- [ ] Filter calendar by category
- [ ] Show task dots in addition to count badge
- [ ] Multi-select days to bulk create/edit tasks
- [ ] Export calendar view as image/PDF
- [ ] Recurring task support with calendar visualization

---

## Notes

- Badge colors align with existing priority thresholds in the system
- Calendar respects existing task filters (if implemented)
- Month data cached separately by start/end date range
- Pre-filling due_date in CreateTaskDialog already exists, just need to pass date prop

---

## Related Files

- `backend/internal/handler/task_handler.go` - API handlers
- `backend/internal/service/task_service.go` - Business logic
- `backend/internal/repository/task_repository.go` - Database queries
- `frontend/app/(dashboard)/dashboard/page.tsx` - Dashboard layout
- `frontend/components/CreateTaskDialog.tsx` - Task creation modal
- `frontend/hooks/useTasks.ts` - React Query hooks
- `docs/design-system.md` - UI/UX patterns
