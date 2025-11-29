# Product Requirements Document (PRD)
## Intelligent Task Prioritization System

**Version:** 1.0
**Last Updated:** 2025-01-15
**Status:** Draft
**Author:** Product Team

---

## Executive Summary

An intelligent task management system designed to capture informal commitments from discussions and meetings, then automatically prioritize them using multi-factor algorithms. The system learns from user behavior, tracks task delays, and provides analytics to help users understand their work patterns and improve task completion rates.

**Core Value Proposition:** Never lose track of commitments, automatically surface the right tasks at the right time, and gain insights into what tasks you consistently avoid or underestimate.

---

## Problem Statement

### Current Pain Points

1. **Commitment Tracking:** Tasks mentioned in casual conversations, meetings, or email threads get lost
2. **Priority Overload:** Everything feels urgent; hard to decide what to work on first
3. **Time Blindness:** No visibility into which tasks consistently get postponed
4. **Estimation Failure:** Chronically underestimate or overestimate certain types of work
5. **Context Loss:** Weeks later, can't remember why a task was created or who requested it

### User Quote
> "I'll be in a meeting and agree to do something small, but by the time I get back to my desk, I've forgotten about it. Then it becomes urgent weeks later when someone asks about it."

---

## Goals and Objectives

### Primary Goals

1. **Capture:** Make it frictionless to capture tasks from any context
2. **Prioritize:** Automatically surface the most important tasks based on multiple factors
3. **Learn:** Track patterns to help users understand their work habits
4. **Adapt:** Evolve task priorities as circumstances change

### Success Metrics

| Metric | Target (3 months) | Target (6 months) |
|--------|------------------|-------------------|
| **Daily Active Usage** | 5+ days/week | 7 days/week |
| **Task Completion Rate** | 60% | 75% |
| **Average Time to Complete** | < 7 days | < 5 days |
| **Tasks Created per Week** | 10+ | 15+ |
| **Delay Detection** | 30% tasks flagged | 20% tasks flagged |

---

## User Personas

### Primary Persona: "Busy Professional Sarah"

**Demographics:**
- Role: Individual contributor (engineer, designer, analyst)
- Experience: 3-7 years in career
- Tech savvy: High

**Behaviors:**
- Attends 5-10 meetings per week
- Gets pulled into ad-hoc discussions
- Juggles multiple projects simultaneously
- Prefers quick, lightweight tools

**Pain Points:**
- Too many commitments, no central tracking
- Everything feels urgent
- Loses context on older tasks
- Wants to improve time estimation

**Goals:**
- Never drop a commitment
- Know what to work on next without thinking
- Understand personal productivity patterns
- Reduce stress from task overload

---

## Use Cases

### Use Case 1: Capture Task from Meeting

**Actor:** User
**Precondition:** User is in a meeting or discussion
**Trigger:** Someone asks user to do something

**Main Flow:**
1. User quickly opens app (mobile or desktop)
2. Enters task title: "Review design doc for auth flow"
3. Optionally adds:
   - Context: "From sync with Alice - she needs feedback by Friday"
   - Related people: "Alice"
   - Category: "Code Review"
   - Effort: "Small (< 1 hour)"
   - Due date: Friday
4. System saves task and assigns initial priority based on due date and user's default settings
5. User returns to meeting

**Postcondition:** Task is saved and will appear in priority-sorted list

---

### Use Case 2: Daily Task Review

**Actor:** User
**Precondition:** User opens app in morning
**Trigger:** Start of workday

**Main Flow:**
1. User sees dashboard with priority-sorted tasks
2. Top tasks show:
   - High priority items (user-set + deadline proximity + time decay)
   - Tasks "at risk" (bumped multiple times)
   - Quick wins (small effort, aging)
3. User selects top task and starts working
4. Marks task complete or bumps to later
5. If bumped, system increments bump counter and adjusts future priority

**Postcondition:** User knows exactly what to work on

---

### Use Case 3: Analyzing Delay Patterns

**Actor:** User
**Precondition:** User has been using system for 2+ weeks
**Trigger:** User navigates to Analytics dashboard

**Main Flow:**
1. User sees "Delay Analysis" section showing:
   - Tasks bumped most frequently
   - Categories most often delayed
   - Time of day/week tasks get created vs completed
2. User notices "Documentation" tasks get bumped 3x more than "Code" tasks
3. Insight: User realizes they avoid documentation work
4. User can adjust strategy: schedule documentation time, delegate, or accept pattern

**Postcondition:** User gains self-awareness about work habits

---

## Functional Requirements

### Core Features (MVP)

#### 1. Task Management

**1.1 Create Task**
- **Required fields:** Title
- **Optional fields:** Description, due date, category, effort estimate, initial priority, context, related people
- **Default values:** Status = "To Do", Created date = now, Priority calculated from inputs
- **Validation:** Title max 200 chars, description max 2000 chars

**1.2 View Tasks**
- **List view:** All tasks sorted by computed priority (descending)
- **Filters:** By status (To Do, In Progress, Done), category, due date range
- **Search:** By title, description, context, related people
- **Visual indicators:** Overdue (red), at risk (yellow), on track (green)

**1.3 Update Task**
- **Editable fields:** All fields except created_at, bump_count
- **Status transitions:** To Do → In Progress → Done, or To Do → Done
- **Priority override:** User can manually boost/lower priority (tracked separately)

**1.4 Complete Task**
- **Action:** Mark as Done
- **Capture:** Actual completion time if effort was estimated
- **Logging:** Record final bump count, time from creation to completion

**1.5 Bump/Delay Task**
- **Action:** "Do this later" button
- **Effect:** Increment bump_count, log bump event with timestamp and reason
- **Priority adjustment:** Penalize priority calculation for chronic bumping

---

#### 2. Smart Prioritization

**2.1 Priority Calculation**

The system calculates a composite priority score (0-100) using:

```
Priority Score = (
    UserPriority × 0.4 +
    TimeDecay × 0.3 +
    DeadlineUrgency × 0.2 +
    BumpPenalty × 0.1
) × EffortBoost
```

**Components:**

| Factor | Range | Description |
|--------|-------|-------------|
| **UserPriority** | 0-100 | User-set importance: Low=25, Medium=50, High=75, Critical=100 |
| **TimeDecay** | 0-100 | Age-based urgency: increases linearly over 30 days |
| **DeadlineUrgency** | 0-100 | Proximity to due date: exponential increase in final 3 days |
| **BumpPenalty** | 0-50 | Punishment for delays: +10 points per bump |
| **EffortBoost** | 1.0-1.3 | Small tasks (< 1 hour) get 1.3x, Large tasks (> 4 hours) get 1.0x |

**Detailed formulas in:** `priority-algorithm.md`

**2.2 Automatic Reprioritization**
- **Frequency:** Every 6 hours (background job)
- **Triggers:** Due date approaching, task aging, multiple bumps
- **Notification:** Alert user if task jumps into top 3

**2.3 At-Risk Detection**
- **Criteria:** Bump count ≥ 3 OR overdue by ≥ 3 days
- **Visual:** Yellow/red badge on task card
- **Suggestion:** "This task has been delayed X times. Consider delegating or breaking it down."

---

#### 3. Analytics Dashboard

**3.1 Delay/Bump Analysis**

**Metrics:**
- Tasks bumped by count (histogram: 0, 1-2, 3-5, 6+)
- Categories with highest avg bump count
- Tasks currently "at risk"
- Pattern detection: "You tend to bump 'Documentation' tasks on Fridays"

**Visualization:** Bar chart, heatmap

**3.2 Time Estimation Accuracy**

**Metrics:**
- Estimated vs actual time (scatter plot)
- Estimation error by category (bar chart)
- Accuracy trend over time (line chart)

**Insights:**
- "Your 'Bug Fix' estimates are 2x too low on average"
- "You're most accurate with 'Code Review' tasks"

**3.3 Source/Category Breakdown**

**Metrics:**
- Task count by category (pie chart)
- Completion rate by category (bar chart)
- Time spent by category (treemap)

**Insights:**
- "40% of tasks come from meetings"
- "Email-sourced tasks have 30% lower completion rate"

**3.4 Velocity & Completion Patterns**

**Metrics:**
- Tasks completed per week (line chart)
- Completion rate trend (line chart with target)
- Best/worst days for task completion (heatmap)
- Time from creation to completion (histogram)

**Insights:**
- "You complete 2x more tasks on Tuesday than Friday"
- "Average completion time: 5.2 days (target: 3 days)"

---

### User Interface Requirements

#### Dashboard Layout

```
┌──────────────────────────────────────────────────────────┐
│  Header: "Today's Priorities" | [+ New Task]             │
├──────────────────────────────────────────────────────────┤
│  ┌────────────────────┐  ┌────────────────────┐         │
│  │ At Risk (3)        │  │ Quick Wins (5)     │         │
│  │ ⚠️ Tasks bumped 3x │  │ ⚡ Small tasks     │         │
│  └────────────────────┘  └────────────────────┘         │
├──────────────────────────────────────────────────────────┤
│  Priority Tasks (sorted)                                 │
│  ┌──────────────────────────────────────────────────┐   │
│  │ 🔴 [95] Review design doc          Due: Tomorrow │   │
│  │ Context: From Alice - auth flow                  │   │
│  │ [Complete] [Bump]                               │   │
│  └──────────────────────────────────────────────────┘   │
│  ┌──────────────────────────────────────────────────┐   │
│  │ 🟡 [78] Update README             Bumped: 2x    │   │
│  │ Context: Tech debt, small task                   │   │
│  │ [Complete] [Bump]                               │   │
│  └──────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────┘
```

#### Task Detail Modal

- Full description
- All metadata (dates, effort, category, people)
- Priority score breakdown (show the math)
- Bump history timeline
- Edit inline

---

## Non-Functional Requirements

### Performance

- **Page load:** < 1 second
- **Task creation:** < 500ms
- **Priority recalculation:** < 100ms per task
- **Analytics query:** < 2 seconds

### Scalability

- **MVP:** Support up to 1,000 tasks per user
- **Future:** Support 10,000+ tasks with pagination

### Usability

- **Mobile-first:** Works on phone for quick capture
- **Keyboard shortcuts:** Fast task creation without mouse
- **Offline support:** (Future) Queue tasks when offline

### Security

- **Authentication:** JWT-based, secure password hashing
- **Authorization:** Users can only see their own tasks (for now)
- **Data privacy:** No sharing of task data

### Reliability

- **Uptime:** 99% (MVP), 99.9% (future)
- **Data backup:** Daily backups
- **Error handling:** Graceful degradation, clear error messages

---

## Data Model (Summary)

See `data-model.md` for full specification.

**Core Entities:**
1. **Users** - Authentication and profile
2. **Tasks** - Main task data
3. **TaskHistory** - Priority changes, bumps, status transitions
4. **Categories** - User-defined categorization
5. **Analytics** - Pre-computed metrics for dashboards

---

## User Stories

### Epic 1: Task Capture
- ✅ **US-001:** As a user, I can quickly create a task with just a title so I don't lose track of commitments
- ✅ **US-002:** As a user, I can add context to tasks so I remember why I created them later
- ✅ **US-003:** As a user, I can categorize tasks so I can filter and analyze by type

### Epic 2: Prioritization
- ✅ **US-004:** As a user, I see my tasks sorted by computed priority so I know what to work on first
- ✅ **US-005:** As a user, I can manually set task priority so urgent items stay on top
- ✅ **US-006:** As a user, I'm notified when tasks become at-risk so I can address delays proactively

### Epic 3: Analytics
- ✅ **US-007:** As a user, I can see which tasks I delay most often so I can address avoidance patterns
- ✅ **US-008:** As a user, I can compare estimated vs actual time so I improve estimation skills
- ✅ **US-009:** As a user, I can see my task completion velocity so I understand my capacity

---

## Technical Architecture

**Stack:**
- **Frontend:** Next.js 16, TypeScript, Shadcn/UI, React Query, Zustand
- **Backend:** Go 1.23, Gin framework, Clean Architecture
- **Database:** PostgreSQL 16 (Supabase), sqlc for type-safe queries
- **Caching:** Redis (Phase 3+)
- **Auth:** JWT tokens
- **Deployment:** Local development + Supabase cloud (Docker deferred)

See `architecture-overview.md` for details.

---

## Implementation Phases

### Phase 1 (Weeks 1-2): Foundation
- Frontend setup with task list UI
- Database schema
- Basic CRUD operations

### Phase 2 (Weeks 3-4): Core Features
- Smart prioritization algorithm
- Bump tracking
- Task filtering and search

### Phase 3 (Weeks 5-6): Analytics
- Delay analysis dashboard
- Completion velocity charts
- Category breakdown

### Phase 4 (Month 2+): Advanced Features
- Background job for auto-reprioritization
- Advanced analytics (estimation accuracy)
- Mobile optimization
- Export/reporting

---

## Future Roadmap (Post-MVP)

### Team Features (v2.0)
- **Shared tasks:** Assign to team members
- **Team analytics:** Team velocity, bottleneck detection
- **Comments:** Discuss tasks inline
- **Notifications:** Task assignments, mentions

### Intelligence (v2.5)
- **ML predictions:** Predict completion time based on history
- **Smart suggestions:** "Based on past behavior, consider delegating this"
- **Pattern recognition:** Auto-categorize tasks from title/context

### Integrations (v3.0)
- **Calendar sync:** Tasks appear in Google/Outlook calendar
- **Email integration:** Create tasks from emails
- **Slack bot:** Quick task creation from Slack
- **GitHub/Jira:** Import issues as tasks

---

## Out of Scope (MVP)

- ❌ Subtasks or nested tasks
- ❌ Time tracking (actual time spent)
- ❌ Recurring tasks
- ❌ File attachments
- ❌ Task dependencies
- ❌ Kanban board view (only list view for MVP)
- ❌ Mobile native apps (PWA only)
- ❌ Real-time collaboration

---

## Open Questions

1. **Bump reasons:** Should we ask users WHY they bumped a task?
   - Pros: Better analytics
   - Cons: Adds friction

2. **Effort estimation:** Required or optional field?
   - Current: Optional
   - Consider: Small nudge to estimate for better analytics

3. **Notifications:** How aggressive should at-risk notifications be?
   - Current: Once when task becomes at-risk
   - Consider: Daily digest of at-risk tasks

---

## Acceptance Criteria

### MVP Launch Criteria

- [ ] User can create, read, update, delete tasks
- [ ] Tasks are sorted by computed priority
- [ ] Bump tracking works and increments counter
- [ ] At-risk detection flags tasks with 3+ bumps
- [ ] Analytics dashboard shows:
  - [ ] Delay/bump analysis
  - [ ] Category breakdown
  - [ ] Completion velocity
- [ ] Mobile-responsive UI
- [ ] Authentication works (register, login, protected routes)
- [ ] No critical bugs
- [ ] Page load < 1 second

---

## Risks and Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|------------|------------|
| **Priority algorithm too complex** | High | Medium | Start simple, iterate based on user feedback |
| **Users don't track enough tasks** | High | Medium | Make capture frictionless, add quick-add shortcuts |
| **Analytics not actionable** | Medium | Medium | Focus on 2-3 key insights, not data overload |
| **Performance with 1000+ tasks** | Medium | Low | Use pagination, indexes, caching early |

---

## Glossary

- **Bump:** User action to delay/postpone a task
- **At-risk:** Task with 3+ bumps or 3+ days overdue
- **Time decay:** Priority increase due to task age
- **Effort boost:** Priority multiplier for small tasks
- **Velocity:** Task completion rate (tasks/week)
- **Quick win:** Small effort task that's been aging

---

## Design System Improvements (Future Work)

### Current State

The application uses a **solid foundation** with shadcn/ui components and CSS variables:
- OKLCH color space for modern, perceptually uniform colors
- Semantic tokens (`--primary`, `--destructive`, etc.)
- Light/dark mode ready
- 9 shadcn/ui components using design tokens correctly

### Issues Identified

**Inconsistent Token Usage (15 instances)**
- Hardcoded colors in application code (`text-red-600`, `bg-gray-50`, etc.)
- Missing semantic tokens: `success`, `warning`, `info`
- No standardized gray scale tokens

**Affected Files:**
- `app/(dashboard)/dashboard/page.tsx` (3 hardcoded colors)
- `app/(dashboard)/analytics/page.tsx` (4 hardcoded colors)
- `app/(dashboard)/layout.tsx` (5 hardcoded colors)
- `components/TaskDetailsSidebar.tsx` (2 hardcoded colors)
- `app/(auth)/login/page.tsx` (3 hardcoded colors)
- `app/(auth)/register/page.tsx` (3 hardcoded colors)

### Recommended Improvements

#### Option A: Quick Token Cleanup (3-4 hours)
**Scope:** Minimal design system refinement before Figma integration

1. **Add Missing Color Tokens** (30 minutes)
   ```css
   /* Add to globals.css */
   --success: oklch(0.7 0.15 145);        /* Green */
   --success-foreground: oklch(0.98 0 0);
   --warning: oklch(0.75 0.15 85);        /* Amber */
   --warning-foreground: oklch(0.2 0 0);
   --info: oklch(0.65 0.15 250);          /* Blue */
   --info-foreground: oklch(0.98 0 0);
   --surface: oklch(0.98 0 0);            /* Replaces bg-gray-50 */
   --surface-variant: oklch(0.96 0 0);    /* Replaces bg-gray-100 */
   ```

2. **Replace Hardcoded Colors** (1 hour)
   - Update 6 files listed above
   - Use semantic tokens instead of direct colors

3. **Create Token Documentation** (1 hour)
   - Color palette reference
   - Component usage guidelines
   - When to use which variant

4. **Add ESLint Rule** (30 minutes)
   - Prevent future hardcoded colors
   - Enforce design token usage

5. **Cross-mode Testing** (1 hour)
   - Verify light/dark mode consistency
   - Test all color combinations

#### Option B: Full Design System with Figma Integration (12-20 hours)
**Scope:** Comprehensive design system aligned with Figma designs

1. **Token Extraction & Mapping** (3-4 hours)
   - Export design tokens from Figma (using Figma API or plugins like Figma Tokens)
   - Map to CSS variables
   - Create comprehensive token system:
     - **Colors:** Full semantic palette + all variants
     - **Typography:** Font families, sizes, weights, line heights
     - **Spacing:** 4px/8px grid system with semantic names
     - **Shadows:** Elevation system (sm, md, lg, xl)
     - **Borders:** Radius and width tokens
     - **Animations:** Duration and easing functions

2. **Component Updates** (4-6 hours)
   - Update all 6 files with hardcoded values
   - Add new shadcn/ui components if Figma has different patterns
   - Create custom components for Figma-specific designs
   - Ensure all components consume tokens

3. **Design System Documentation** (2-3 hours)
   - Set up Storybook for component library
   - Token reference documentation
   - Component usage guidelines
   - Accessibility guidelines
   - Design principles

4. **Figma Sync Tooling** (3-5 hours)
   - Set up Figma token sync (Style Dictionary or Figma Tokens plugin)
   - Create build pipeline: Figma tokens → CSS variables
   - Document workflow for design updates
   - Automate token extraction where possible

5. **Testing & Refinement** (2-3 hours)
   - Visual regression testing setup
   - Cross-browser testing
   - Responsive testing at all breakpoints
   - Accessibility audit

### Image-Based Design Changes (if Figma not available)

**Effort Estimates:**
- **Simple style updates** (colors/spacing): 2-4 hours
- **Component redesigns**: 1-2 hours per component
- **Full page redesigns**: 3-5 hours per page
- **New patterns/components**: 2-4 hours each

**Note:** Static images require manual translation to code. Figma provides:
- Exact measurements via inspect mode
- CSS export for some properties
- Token extraction plugins
- Easier iteration and collaboration

### Decision Matrix

| Approach | Effort | Best For | When to Use |
|----------|--------|----------|-------------|
| **Option A** | 3-4 hours | Quick consistency fix | Before Phase 2, minimal design changes expected |
| **Option B** | 12-20 hours | Full design system | When Figma files available, major redesign planned |
| **Images Only** | Variable | One-off changes | Ad-hoc design updates, no comprehensive redesign |

### Recommendation

**Phase 1 Completion:** Defer design system work (Option A or B) until Figma files or design direction is provided.

**Future Phase:** Implement Option B (Full Design System) when:
- Figma designs are finalized
- Design direction is clear
- Ready for comprehensive UI polish

---

## Appendix

### Related Documents
- `data-model.md` - Full database schema
- `priority-algorithm.md` - Detailed prioritization logic
- `architecture-overview.md` - Technical architecture
- `phase-1-weeks-1-2.md` - Implementation guide

### References
- Eisenhower Matrix (urgent/important framework)
- Getting Things Done (GTD) methodology
- Jira task management patterns
- Todoist priority system

---

**Document Status:** Ready for Development
**Next Review:** After Phase 1 completion
