# Phase 5: Product Enhancements

**Goal:** Transform TaskFlow from a task manager into an intelligent productivity platform with unique differentiators.

**Timeline:** December 2025 onwards

---

## Phase 5A: Quick Wins ✅ COMPLETE

High-impact features with low-medium effort that improve daily usability.

### 5A.1 Recurring Tasks (PR #45)

**Status:** ✅ Complete

**What was built:**
- `task_series` table for recurrence rules (daily/weekly/monthly patterns)
- Series management: create, update, stop recurrence
- Automatic next task generation on completion
- User preferences for recurrence behavior (per-category defaults)
- RecurrenceCompletionDialog with options:
  - Calculate next due from original vs completion date
  - Skip next occurrence
  - Stop recurrence entirely

**Key Files:**
- `backend/migrations/000004_recurring_tasks.up.sql`
- `backend/internal/domain/recurrence.go`
- `backend/internal/service/recurrence_service.go`
- `frontend/components/RecurrenceSelector.tsx`
- `frontend/components/RecurrenceCompletionDialog.tsx`

---

### 5A.2 Priority Explanation Panel (PR #44)

**Status:** ✅ Complete

**What was built:**
- Visual donut chart showing priority factor contributions
- Detailed breakdown table: Raw → Weight → Points
- Formula explanation inline
- Integration in TaskDetailsSidebar

**Key Files:**
- `frontend/components/PriorityBreakdownPanel.tsx`
- `backend/internal/domain/priority/calculator.go` (CalculateWithBreakdown)

---

### 5A.3 Quick Add - Cmd+K (PR #46)

**Status:** ✅ Complete

**What was built:**
- Global `Cmd/Ctrl+K` keyboard shortcut
- Opens CreateTaskDialog instantly
- Works from anywhere in the app
- `?` shortcut to show help modal

**Key Files:**
- `frontend/hooks/useGlobalKeyboardShortcuts.ts`
- `frontend/contexts/KeyboardShortcutsContext.tsx`

---

### 5A.4 Keyboard Navigation (PR #46)

**Status:** ✅ Complete

**What was built:**
- `j/k` for task list navigation (vim-style)
- `Enter` to open task details
- `e` to edit selected task
- `c` to complete selected task
- `d` to delete selected task
- `Esc` to close dialogs
- Visual ring highlight on selected task

**Key Files:**
- `frontend/hooks/useTaskNavigation.ts`
- `frontend/components/KeyboardShortcutsHelp.tsx`

---

## Phase 5B: Core Enhancements (In Progress)

Medium-effort features that add significant depth to task management.

### 5B.1 Subtasks / Parent-Child Tasks 🔄 IN PROGRESS

**Status:** In Development

**Goal:** Allow tasks to have hierarchical relationships with subtasks.

**Scope:**
- Parent-child task relationships (hierarchical nesting)
- Independent subtask completion (not auto-completing parent)
- Prompt user to close parent when last subtask is completed
- Priority inheritance from parent to children

**NOT in scope (Phase 5B.2):**
- "Blocked by" dependencies
- Dependency graph visualization

**Database Changes:**
```sql
ALTER TABLE tasks ADD COLUMN parent_id UUID REFERENCES tasks(id) ON DELETE CASCADE;
ALTER TABLE tasks ADD COLUMN is_subtask BOOLEAN DEFAULT false;
CREATE INDEX idx_tasks_parent_id ON tasks(parent_id);
```

**Backend Changes:**
- Add `parent_id` and `is_subtask` fields to Task domain
- Create subtask endpoints (POST /api/v1/tasks/:id/subtasks)
- List subtasks endpoint (GET /api/v1/tasks/:id/subtasks)
- Modify priority calculator to support parent priority inheritance
- Handle subtask completion → check if parent should be prompted

**Frontend Changes:**
- Subtask display under parent tasks (expandable)
- "Add subtask" button on task cards/details
- Completion prompt dialog when last subtask done
- Visual hierarchy in task list (indentation/nesting)

**API Design:**
```
POST   /api/v1/tasks/:id/subtasks     - Create subtask for parent
GET    /api/v1/tasks/:id/subtasks     - List subtasks
GET    /api/v1/tasks/:id              - Include subtask_count, open_subtask_count
DELETE /api/v1/tasks/:id              - Cascade delete subtasks
```

---

### 5B.2 Blocked-By Dependencies (Planned)

**Status:** [ ] Planned (after 5B.1)

**Goal:** Allow tasks to have "blocked by" relationships.

**Scope:**
- Task A can be blocked by Task B
- Blocked tasks show warning indicator
- Completing a blocking task updates blocked tasks
- Optional: Dependency graph visualization

**Database Changes:**
```sql
CREATE TABLE task_dependencies (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    blocked_by_task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(task_id, blocked_by_task_id)
);
```

---

### 5B.3 Gamification (Planned)

**Status:** [ ] Planned

**Goal:** Add streaks, achievements, and productivity scores.

**Features:**
- Daily completion streaks
- Achievement badges (first task, 10 tasks, streak milestones)
- Productivity score based on completion patterns
- Leaderboard (optional, for teams)

---

### 5B.4 Procrastination Detection (Planned)

**Status:** [ ] Planned

**Goal:** AI-powered insights from bump patterns.

**Features:**
- Pattern detection: "You often bump tasks on Mondays"
- Category insights: "Exercise tasks sit 5 days before action"
- Actionable suggestions
- Integration with Smart Insights service

---

### 5B.5 Natural Language Input (Planned)

**Status:** [ ] Planned

**Goal:** Parse natural language into task fields.

**Examples:**
- "Buy groceries tomorrow high priority" → title, due date, priority
- "Call mom every Sunday" → title, recurrence
- "Finish report by Friday #work" → title, due date, category

**Implementation:**
- Claude API integration for parsing
- Inline preview of parsed fields
- Edit before save

---

## Phase 5C: Advanced Features (Future)

High-effort features for power users and enterprise needs.

### 5C.1 Pomodoro Timer (Planned)

Built-in focus timer tied to tasks.

### 5C.2 AI Daily Briefing (Planned)

Claude-powered morning productivity summary.

### 5C.3 Smart Scheduling (Planned)

Calendar integration with auto time-blocking.

### 5C.4 Mobile PWA (Planned)

Progressive Web App for mobile access.

---

## Exit Criteria

### Phase 5A ✅ Complete
- [x] Recurring Tasks implemented
- [x] Priority Explanation Panel implemented
- [x] Quick Add (Cmd+K) implemented
- [x] Keyboard Navigation implemented

### Phase 5B Exit Criteria
- [ ] Subtasks (5B.1) complete
- [ ] Blocked-By Dependencies (5B.2) complete
- [ ] At least 1 additional 5B feature complete
- [ ] Analytics expanded for new features

### Phase 5C Exit Criteria
- [ ] At least 1 Phase 5C feature complete
- [ ] User engagement metrics show improvement

---

## Analytics Tracking

See `docs/analytics-gaps-phase5.md` for detailed analytics requirements per feature.

### Key Metrics to Add:
- Subtask usage (tasks with subtasks, avg subtasks per task)
- Subtask completion rate
- Parent completion triggered by subtask prompt
- Dependency graph usage

---

## Technical Considerations

### Database Migrations
- Use sequential migration files (000005, 000006, etc.)
- Always provide up and down migrations
- Test migrations on development data first

### Priority Inheritance
- Parent priority should boost child priority by ~15%
- Recalculate child priorities when parent changes
- Consider cascade performance for deep hierarchies

### UI/UX Patterns
- Use existing expansion patterns from TaskFilters
- Follow keyboard shortcut patterns from Phase 5A
- Maintain visual hierarchy with indentation
- Keep parent/child relationship obvious but not cluttered
