# Analytics Gaps Analysis for Phase 5 Features

**Created:** 2025-12-04
**Purpose:** Identify analytics tracking requirements for Phase 5 features

---

## Current Analytics Coverage Summary

### What We Track Today

| Category | Metrics | Coverage |
|----------|---------|----------|
| **Task Lifecycle** | Created, completed, deleted, bumped | ✅ Excellent |
| **Priority** | Distribution by range, calculated scores | ✅ Excellent |
| **Categories** | Breakdown, trends, bump stats per category | ✅ Excellent |
| **Time Patterns** | Completion by day/hour (heatmap), velocity | ✅ Excellent |
| **Bump Behavior** | Count distribution, at-risk detection, avoidance patterns | ✅ Excellent |
| **Effort** | Task creation by effort size | ✅ Good |
| **Insights** | 6 rule-based insights, time estimation | ✅ Good |
| **Operations** | HTTP requests, latency, in-flight (Prometheus) | ✅ Excellent |

### What We DON'T Track

| Gap | Impact | Priority |
|-----|--------|----------|
| User engagement (session time, feature usage) | Can't measure UX improvements | High |
| Task completion time (duration from create→done) | Limited time estimation accuracy | Medium |
| User streaks/consistency | Can't enable gamification | High |
| Subtask progress | Can't track granular completion | Medium |
| Focus/work sessions | No Pomodoro support | Low (future) |
| NLP parsing success rate | Can't measure NLP quality | Low (future) |

---

## Analytics Requirements by Feature

### Phase 5A: Quick Wins

#### 1. Recurring Tasks
**New Analytics Needed:**

| Metric | Type | Purpose |
|--------|------|---------|
| `recurrence_pattern` | Task field | Track daily/weekly/monthly patterns |
| `recurrence_completed_count` | Counter | Times this recurring task was completed |
| `recurrence_skip_count` | Counter | Times user skipped a recurrence |
| `taskflow_recurring_tasks_active` | Prometheus gauge | Currently active recurring tasks |
| `recurring_completion_rate` | Computed | % of recurrences completed on time |

**Insight Opportunities:**
- "You complete 'Weekly Review' 80% of the time"
- "Monday tasks have lowest recurrence completion rate"

**Backend Changes:**
- Add `recurrence_rule` column to tasks table
- Add `parent_recurring_task_id` for generated instances
- New SQL queries for recurrence analytics

---

#### 2. Priority Explanation Panel
**New Analytics Needed:**

| Metric | Type | Purpose |
|--------|------|---------|
| `priority_panel_views` | Counter | Times users viewed priority breakdown |
| `priority_factor_breakdown` | Already exists | Expose existing calculation data |

**No Backend Changes Required** - Priority factors already computed, just need to expose them in API response.

**API Enhancement:**
```json
// Current task response
{ "priority_score": 75.5 }

// Enhanced response
{
  "priority_score": 75.5,
  "priority_breakdown": {
    "user_priority": 30.0,
    "deadline_urgency": 25.0,
    "time_decay": 10.5,
    "bump_penalty": 10.0,
    "effort_boost": 0.0
  }
}
```

---

#### 3. Quick Add (Cmd+K)
**New Analytics Needed:**

| Metric | Type | Purpose |
|--------|------|---------|
| `quick_add_usage_count` | Counter | Times quick add was used |
| `task_creation_method` | Dimension | "quick_add" vs "dialog" vs "calendar" |
| `quick_add_abandonment` | Counter | Opened but didn't create task |

**Frontend Only** - Track via existing task creation endpoint with new `source` parameter.

---

#### 4. Keyboard Navigation
**New Analytics Needed:**

| Metric | Type | Purpose |
|--------|------|---------|
| `keyboard_shortcut_usage` | Counter by shortcut | Which shortcuts are used most |
| `navigation_method` | Dimension | "keyboard" vs "mouse" |

**Frontend Only** - Analytics event tracking, no backend changes.

---

### Phase 5B: Core Enhancements

#### 5. Subtasks/Checklists
**New Analytics Needed:**

| Metric | Type | Purpose |
|--------|------|---------|
| `subtask_count` | Task field | Number of subtasks per parent |
| `subtask_completion_rate` | Computed | % of subtasks completed |
| `tasks_with_subtasks` | Counter | Tasks that use subtasks feature |
| `avg_subtasks_per_task` | Computed | Subtask usage patterns |
| `subtask_completion_order` | Analytics | Do users complete in order or randomly? |

**Backend Changes:**
- New `subtasks` table (id, parent_task_id, title, completed, order)
- Parent task progress = completed_subtasks / total_subtasks
- New SQL queries for subtask analytics

**Insight Opportunities:**
- "Tasks with subtasks have 40% higher completion rate"
- "You tend to complete easy subtasks first"

---

#### 6. Gamification (Streaks & Achievements)
**New Analytics Needed:**

| Metric | Type | Purpose |
|--------|------|---------|
| `current_streak` | User field | Consecutive days with completions |
| `longest_streak` | User field | All-time best streak |
| `total_points` | User field | Accumulated gamification points |
| `achievements_unlocked` | User field | JSON array of achievement IDs |
| `daily_completion_count` | Analytics | Tasks completed today |
| `streak_broken_count` | Counter | Times streak was broken |
| `achievement_unlocked_events` | Events | When achievements are earned |

**Backend Changes:**
- New `user_stats` table (user_id, current_streak, longest_streak, total_points, etc.)
- New `achievements` table (id, name, description, criteria, points)
- New `user_achievements` junction table
- Daily job to update streaks
- Achievement trigger logic

**Insight Opportunities:**
- "You're on a 7-day streak! Keep it up!"
- "3 more tasks to unlock 'Centurion' achievement"

---

#### 7. Procrastination Detection
**New Analytics Needed:**

| Metric | Type | Purpose |
|--------|------|---------|
| `bump_pattern_by_day` | Analytics | Which days see most bumps |
| `bump_pattern_by_category` | Already exists | Enhance with more detail |
| `task_age_at_bump` | Analytics | How old are tasks when bumped |
| `bump_to_completion_ratio` | Computed | Bumps before eventual completion |
| `procrastination_score` | Computed | Per-task or per-category score |

**Mostly Covered** - We already track bump patterns. Enhancements:
- Add `created_at` to bump tracking for age analysis
- New insight: "Tasks in 'Exercise' category sit 5 days before first action"
- Pattern detection: "You bump tasks on Monday mornings"

---

#### 8. Natural Language Input
**New Analytics Needed:**

| Metric | Type | Purpose |
|--------|------|---------|
| `nlp_parse_attempts` | Counter | Times NLP parsing was attempted |
| `nlp_parse_success` | Counter | Successful parses |
| `nlp_parse_corrections` | Counter | User corrected parsed result |
| `nlp_fields_extracted` | Breakdown | Which fields (date, priority, category) |
| `nlp_confidence_score` | Per-parse | How confident was the parse |

**Backend Changes:**
- Claude API integration for parsing
- Parse result logging
- Feedback loop for corrections

---

### Phase 5C: Advanced Features

#### 9. Pomodoro Timer
**New Analytics Needed:**

| Metric | Type | Purpose |
|--------|------|---------|
| `focus_sessions_started` | Counter | Pomodoro sessions begun |
| `focus_sessions_completed` | Counter | Full 25-min sessions |
| `focus_sessions_interrupted` | Counter | Sessions cut short |
| `focus_time_total` | Duration | Total focused time |
| `focus_time_per_task` | Per-task | Time spent on specific task |
| `avg_session_length` | Computed | Average focus duration |
| `peak_focus_hours` | Analytics | When user focuses best |

**Backend Changes:**
- New `focus_sessions` table (id, user_id, task_id, started_at, ended_at, completed)
- Real-time session tracking
- Integration with task completion

**Insight Opportunities:**
- "You focus best between 9-11am"
- "Average 3 Pomodoros per task completion"

---

#### 10. AI Daily Briefing
**New Analytics Needed:**

| Metric | Type | Purpose |
|--------|------|---------|
| `briefing_generated_count` | Counter | Daily briefings created |
| `briefing_viewed_count` | Counter | Users who read briefing |
| `briefing_action_taken` | Counter | Tasks actioned from briefing |
| `briefing_feedback` | Rating | Helpful/not helpful |

**Backend Changes:**
- Claude API integration
- Briefing generation job (morning)
- Briefing content storage/caching

---

#### 11. Smart Scheduling
**New Analytics Needed:**

| Metric | Type | Purpose |
|--------|------|---------|
| `calendar_sync_enabled` | User field | Users with calendar connected |
| `time_blocks_created` | Counter | Auto-scheduled blocks |
| `time_blocks_respected` | Counter | User kept the scheduled time |
| `time_blocks_moved` | Counter | User rescheduled |
| `scheduling_accuracy` | Computed | How often predictions are right |

**Backend Changes:**
- Google Calendar OAuth integration
- Time block suggestion algorithm
- Calendar event creation/sync

---

#### 12. Mobile PWA
**New Analytics Needed:**

| Metric | Type | Purpose |
|--------|------|---------|
| `platform` | Dimension | "web", "pwa-ios", "pwa-android" |
| `device_type` | Dimension | "desktop", "tablet", "mobile" |
| `offline_actions` | Counter | Actions taken while offline |
| `pwa_install_count` | Counter | PWA installations |

**Frontend Changes:**
- Service worker for offline
- Platform detection
- Install prompt tracking

---

## Priority Matrix: What to Build First

### High Priority (Required for Phase 5A)

| Analytics | Reason | Effort |
|-----------|--------|--------|
| Recurrence tracking | Core feature requirement | Medium |
| Priority breakdown exposure | Already computed, just expose | Low |
| Task creation source | Enables quick add analytics | Low |
| Keyboard usage events | Frontend-only tracking | Low |

### Medium Priority (Required for Phase 5B)

| Analytics | Reason | Effort |
|-----------|--------|--------|
| Subtask schema & tracking | Core feature requirement | Medium |
| Streak/achievement system | Gamification foundation | High |
| Enhanced bump patterns | Procrastination detection | Low |

### Lower Priority (Phase 5C)

| Analytics | Reason | Effort |
|-----------|--------|--------|
| Focus session tracking | Pomodoro feature | Medium |
| Calendar integration | Smart scheduling | High |
| Platform analytics | PWA feature | Low |

---

## Recommended Implementation Order

1. **Immediate (with Phase 5A features):**
   - Expose priority breakdown in task API
   - Add `source` field to task creation
   - Frontend keyboard/mouse analytics events

2. **Next Sprint (Phase 5B prep):**
   - Design subtasks schema
   - Design user_stats/achievements schema
   - Enhance bump pattern queries

3. **Future (Phase 5C prep):**
   - Focus session schema
   - Calendar OAuth flow
   - Platform detection

---

## Summary

**Current State:** Excellent task lifecycle and behavioral analytics. Strong foundation.

**Key Gaps:**
1. No user engagement tracking (session time, feature usage)
2. No streak/consistency tracking (blocks gamification)
3. No subtask granularity (blocks checklist feature)
4. Priority factors computed but not exposed (quick fix)

**Recommendation:** Phase 5A features require minimal analytics expansion. Phase 5B needs schema changes. Plan for these in parallel with feature development.
