# Phase 2.5: Quick Wins + Core Features Sprint
## TaskFlow - Production Polish & Essential Features

**Goal:** Implement high-value UX improvements, core features, and critical production readiness fixes before Phase 3's comprehensive production hardening.

**Timeline:** 7-10 days

**By the end of this phase, you will have:**
- ✅ Critical security fixes (JWT_SECRET enforcement)
- ✅ Polished UX (calendar positioning, categories dropdown)
- ✅ Core search & filtering functionality
- ✅ Complete analytics dashboard with visualizations
- ✅ Documented design system patterns

**What you're building:** Essential features that make TaskFlow production-ready for real users while establishing design consistency.

---

## Current Status: What's Been Completed

### ✅ Phase 1 (Completed)
- [x] Next.js 16 + React 19 + TypeScript frontend
- [x] Supabase PostgreSQL database
- [x] shadcn/ui component library
- [x] Authentication UI (login/register)
- [x] Task dashboard with priority visualization
- [x] Analytics dashboard skeleton
- [x] Dark mode support

### ✅ Phase 2 (Completed)
- [x] Go 1.23 backend with Gin framework
- [x] Clean Architecture implementation
- [x] Priority calculation algorithm (multi-factor scoring)
- [x] Bump tracking system
- [x] JWT authentication
- [x] Task CRUD API endpoints
- [x] Frontend-backend integration
- [x] Calendar widget (mini-calendar in sidebar)

### 🔄 Phase 2 Enhancements (Completed via PRs)
- [x] Calendar widget redesign (#4)
- [x] Dark mode improvements (#3)
- [x] Priority scale refinement (#3)
- [x] UX improvements (#3)

---

## Phase 2.5 Roadmap

### 🎯 Day 1: Critical Fixes & Quick Wins

#### Morning: Security & Positioning
**Tasks:**
- [x] **Security Fix:** Remove JWT_SECRET defaults in `backend/internal/config/config.go`
  - [x] Make JWT_SECRET required (panic if missing)
  - [x] Update `.env.example` with clear documentation
  - [x] Test that backend crashes on missing secret
  - [x] Also made DATABASE_URL required
  - [x] Added comprehensive unit tests

- [x] **Calendar Popover Fix:** Improve positioning
  - [x] Found component: `frontend/components/CalendarTaskPopover.tsx`
  - [x] Add collision detection with `collisionPadding={16}`
  - [x] Explicitly enable `avoidCollisions={true}` for automatic repositioning
  - [x] Popover will now flip to avoid viewport overflow
  - [ ] Test on various screen sizes (manual testing needed)
  - [ ] Document in design system

**Acceptance Criteria:**
- ✅ Backend refuses to start without JWT_SECRET
- ✅ Calendar popover has collision detection configured
- ⏳ Visual testing needed (requires running dev server)

---

#### Afternoon: Categories Dropdown
**Tasks:**
- [ ] **Backend:** Verify category field exists in Task model (already done)
- [ ] **Frontend:** Create categories select component
  - [ ] Add to `frontend/components/tasks/TaskForm.tsx`
  - [ ] Use shadcn Select component
  - [ ] Define default categories: `["Work", "Personal", "Meeting", "Code Review", "Bug Fix", "Documentation", "Planning", "Other"]`
  - [ ] Allow custom category input
  - [ ] Add category to create/edit task forms

- [ ] **UI Polish:**
  - [ ] Add category badges to task cards
  - [ ] Color-code categories (optional)
  - [ ] Update design system docs

**Acceptance Criteria:**
- ✅ Can select category when creating/editing task
- ✅ Category displays on task cards
- ✅ Category saves to database correctly

---

### 🔍 Days 2-3: Search & Filtering

#### Day 2: Backend Search API
**Tasks:**
- [ ] **Add search/filter queries** to `backend/internal/repository/task_repository.go`
  - [ ] Add `SearchTasks()` method with filters:
    - Search query (title/description full-text)
    - Category filter
    - Status filter (todo/in_progress/done)
    - Priority range filter
    - Due date range filter
  - [ ] Use PostgreSQL `ILIKE` for simple search or `tsvector` for full-text
  - [ ] Maintain priority sorting with filters

- [ ] **Add handler** in `backend/internal/handler/task_handler.go`
  - [ ] New endpoint: `GET /api/v1/tasks/search`
  - [ ] Parse query parameters
  - [ ] Return filtered, sorted results

**Acceptance Criteria:**
- ✅ Search endpoint returns filtered tasks
- ✅ Multiple filters can be combined
- ✅ Results maintain priority sorting

---

#### Day 3: Frontend Search UI
**Tasks:**
- [ ] **Create search components**
  - [ ] `frontend/components/tasks/TaskSearch.tsx` - Search input with debounce
  - [ ] `frontend/components/tasks/TaskFilters.tsx` - Filter chips/dropdowns
  - [ ] `frontend/hooks/useTaskSearch.ts` - React Query hook for search

- [ ] **Integrate into dashboard**
  - [ ] Add search bar above task list
  - [ ] Add filter panel (collapsible)
  - [ ] Show active filters as removable chips
  - [ ] Add "Clear all filters" button

- [ ] **UX Polish**
  - [ ] Debounce search input (300ms)
  - [ ] Show loading state during search
  - [ ] Show "no results" state
  - [ ] Preserve filters in URL query params (optional)

**Acceptance Criteria:**
- ✅ Can search tasks by text
- ✅ Can filter by category, status, priority
- ✅ Filters update results immediately
- ✅ Search is performant (debounced)

---

### 📊 Days 4-6: Analytics Dashboard

#### Day 4: Backend Analytics Queries
**Tasks:**
- [ ] **Create analytics queries** in `backend/internal/repository/task_repository.go`
  - [ ] `GetCompletionStats()` - Completed vs total tasks by time period
  - [ ] `GetBumpAnalytics()` - Average bump count, tasks by bump count
  - [ ] `GetCategoryBreakdown()` - Task count and completion rate by category
  - [ ] `GetVelocityMetrics()` - Tasks completed per day/week
  - [ ] `GetPriorityDistribution()` - Count of tasks by priority ranges

- [ ] **Add analytics handler**
  - [ ] New endpoint: `GET /api/v1/analytics/summary`
  - [ ] New endpoint: `GET /api/v1/analytics/trends`
  - [ ] Return aggregated metrics

**Acceptance Criteria:**
- ✅ Analytics endpoints return correct aggregated data
- ✅ Queries are performant (use indexes)
- ✅ Data covers last 7, 30, 90 days

---

#### Day 5: Frontend Analytics Components
**Tasks:**
- [ ] **Install chart library** (if not already installed)
  - [ ] `npm install recharts`

- [ ] **Create chart components**
  - [ ] `frontend/components/analytics/CompletionChart.tsx` - Line/bar chart
  - [ ] `frontend/components/analytics/CategoryPieChart.tsx` - Category breakdown
  - [ ] `frontend/components/analytics/BumpHeatmap.tsx` - Bump frequency
  - [ ] `frontend/components/analytics/VelocityChart.tsx` - Tasks completed over time

- [ ] **Create analytics hook**
  - [ ] `frontend/hooks/useAnalytics.ts` - Fetch analytics data
  - [ ] Handle time period selection (7d/30d/90d)

**Acceptance Criteria:**
- ✅ Charts display real data from backend
- ✅ Charts are responsive and styled consistently
- ✅ Can switch between time periods

---

#### Day 6: Analytics Dashboard Integration
**Tasks:**
- [ ] **Update analytics page** `frontend/app/(dashboard)/dashboard/analytics/page.tsx`
  - [ ] Remove mock data
  - [ ] Integrate real chart components
  - [ ] Add stat cards (total tasks, completion rate, avg priority, etc.)
  - [ ] Add time period selector
  - [ ] Add export data button (optional)

- [ ] **Polish & UX**
  - [ ] Loading states for charts
  - [ ] Empty states ("No data yet")
  - [ ] Tooltips on hover
  - [ ] Consistent color scheme

**Acceptance Criteria:**
- ✅ Analytics page shows real, live data
- ✅ All charts render correctly
- ✅ Page is visually polished and consistent

---

### 📐 Day 7: Design System Documentation

#### Tasks:
- [ ] **Audit existing components**
  - [ ] List all custom components used
  - [ ] Document shadcn/ui components used
  - [ ] Note custom styling patterns

- [ ] **Update** `docs/design-system.md`
  - [ ] Document color usage (primary, destructive, etc.)
  - [ ] Document spacing patterns (gaps, padding)
  - [ ] Document typography scale
  - [ ] Document button variants and usage
  - [ ] Document card layouts
  - [ ] Document form patterns
  - [ ] Document badge/chip usage
  - [ ] Document animation patterns

- [ ] **Create component examples**
  - [ ] Add code snippets for each pattern
  - [ ] Add "when to use" guidelines
  - [ ] Add accessibility notes

- [ ] **Document new components from this sprint**
  - [ ] Category dropdown pattern
  - [ ] Search/filter UI pattern
  - [ ] Chart component patterns
  - [ ] Calendar popover pattern

**Acceptance Criteria:**
- ✅ Design system docs are comprehensive
- ✅ All patterns used in app are documented
- ✅ Includes code examples and guidelines
- ✅ Team can reference for future features

---

## Testing Checklist

### End-to-End Testing
After completing all features, verify:

- [ ] **Security**
  - [ ] Backend refuses to start without JWT_SECRET
  - [ ] No default secrets in code

- [ ] **UX Polish**
  - [ ] Calendar popover positions correctly
  - [ ] Category dropdown works in create/edit forms
  - [ ] Category badges display on task cards

- [ ] **Search & Filtering**
  - [ ] Text search finds relevant tasks
  - [ ] Category filter works
  - [ ] Status filter works
  - [ ] Priority filter works
  - [ ] Multiple filters combine correctly
  - [ ] "Clear filters" resets to all tasks

- [ ] **Analytics**
  - [ ] All charts display data
  - [ ] Time period selector updates charts
  - [ ] Stat cards show correct numbers
  - [ ] Charts are responsive on mobile

- [ ] **Design System**
  - [ ] Documentation is complete and accurate
  - [ ] Examples match actual implementation

---

## Known Issues & Future Work

### Deferred to Phase 3
- [ ] Comprehensive unit testing
- [ ] Integration testing with testcontainers
- [ ] Structured logging (slog)
- [ ] Interface-based dependency injection
- [ ] Redis rate limiting migration
- [ ] Production Docker builds
- [ ] CI/CD pipeline

### Deferred to Phase 4+
- [ ] General site redesign
- [ ] iOS SwiftUI companion app
- [ ] Advanced analytics (ML predictions, recommendations)
- [ ] Real-time collaboration features
- [ ] API documentation (Swagger/OpenAPI)

---

## Success Metrics

By the end of Phase 2.5, you should have:

1. **Zero critical security issues** (JWT secrets enforced)
2. **Core features complete** (search, filtering, analytics, categories)
3. **Polished UX** (calendar positioning, consistent design)
4. **Documented patterns** (design system established)
5. **Production-ready codebase** (ready for Phase 3 hardening)

---

## Next Steps

After completing Phase 2.5, proceed to:
- **Phase 3:** Production hardening (testing, logging, monitoring, Docker)
- **Phase 4:** Advanced features and scaling (Redis, background jobs, optimization)

---

## Notes

- **Focus:** This phase prioritizes user-facing value and critical fixes
- **Scope:** Keep features simple and ship quickly
- **Quality:** Maintain code quality, document as you go
- **Testing:** Manual testing is acceptable for now (automated testing in Phase 3)

**Remember:** The goal is production readiness for real users, not perfection. Ship, iterate, improve.
