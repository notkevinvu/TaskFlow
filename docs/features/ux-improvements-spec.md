# UX Improvements: Button Loading States, Toast Undo, Task Status Management

**Date:** 2025-12-08
**Status:** Spec Complete
**Type:** Product Specification

---

## Overview

This feature improves user experience through three interconnected enhancements:
1. Visual feedback when buttons are processing (loading states)
2. Ability to undo task completion and deletion via toast action buttons
3. Backend support for reversing task operations with proper gamification handling

---

## Problem Statement

### Current Pain Points

1. **No Loading Feedback on Quick Actions**
   - When users click Complete, Bump, or Delete buttons, they only see the button become disabled
   - No visual indication that the action is processing
   - Users may click multiple times or feel uncertain if their action registered

2. **No Undo for Accidental Actions**
   - If a user accidentally completes or deletes a task, there's no easy way to reverse it
   - BulkRestore exists but is designed for bulk operations, not quick undo
   - Deleting a task is permanent (hard delete)

3. **Gamification Integrity Issues**
   - If a task is manually "uncompleted" via status change, gamification stats remain inflated
   - Achievements earned from the completion are not revoked
   - Category mastery counts don't decrease

---

## User Stories

### Loading States

**US-1:** As a user, when I click a button that triggers an API call, I want to see a loading spinner so I know my action is being processed.

**US-2:** As a user, I want buttons to be disabled during loading so I don't accidentally trigger duplicate actions.

### Undo Complete

**US-3:** As a user, when I complete a task, I want to see an "Undo" option in the success toast so I can quickly reverse an accidental completion.

**US-4:** As a user, when I undo a task completion, I expect my gamification stats to be corrected (completion count decremented, achievements revoked if no longer qualified).

### Undo Delete

**US-5:** As a user, when I delete a task, I want to see an "Undo" option in the toast so I can recover an accidentally deleted task.

**US-6:** As a user, I expect the undo option to be available for a reasonable time (5 seconds) before the toast auto-dismisses.

---

## Requirements

### Functional Requirements

#### FR-1: Button Loading States
- [ ] Buttons show a spinner icon when their action is processing
- [ ] Buttons are disabled during loading state
- [ ] Original button text remains visible alongside spinner
- [ ] Loading state applies to: Complete, Bump, Delete, Save, Create buttons

#### FR-2: Toast Undo for Completion
- [ ] Task completion toast includes an "Undo" action button
- [ ] Toast remains visible for 5 seconds
- [ ] Clicking "Undo" restores task to `todo` status
- [ ] Clicking "Undo" clears the `completed_at` timestamp
- [ ] Clicking "Undo" triggers gamification reversal
- [ ] Success feedback shown after undo completes

#### FR-3: Toast Undo for Deletion
- [ ] Task deletion toast includes an "Undo" action button
- [ ] Toast remains visible for 5 seconds
- [ ] Clicking "Undo" restores the deleted task
- [ ] Task reappears in appropriate lists after restoration

#### FR-4: Gamification Reversal
- [ ] Category mastery count decremented when task uncompleted
- [ ] Stats recomputed (total_completed, completion_rate, streaks)
- [ ] Achievements revoked if user no longer qualifies
- [ ] Streak may be broken if uncompleted task was the only completion that day

#### FR-5: Soft Delete
- [ ] Task deletion sets `deleted_at` timestamp instead of hard delete
- [ ] Soft-deleted tasks excluded from all normal queries
- [ ] Soft-deleted tasks can be restored via undelete endpoint

### Non-Functional Requirements

#### NFR-1: Performance
- Loading spinner should appear within 50ms of button click
- Undo action should complete within 500ms

#### NFR-2: Accessibility
- Loading state should be announced to screen readers
- Undo button in toast must be keyboard accessible

#### NFR-3: Reliability
- If undo fails, show error toast with clear message
- Gamification reversal failures should be logged but not block undo

---

## UX Decisions

### Toast Duration: 5 seconds
**Rationale:** Based on UX research, 5 seconds provides enough time to read and decide without being intrusive. For destructive actions like deletion, some apps use longer (30-60s), but since we have soft delete, 5 seconds is sufficient.

### Keep Button Text During Loading
**Rationale:** Nielsen Norman Group recommends keeping the label visible so users know which action is processing. The spinner appears alongside the text, not replacing it.

### Gamification Reversal: Full Reversal
**Decision:** When undoing a completion:
- Decrement completion count
- Recompute streaks (may break current streak)
- Revoke achievements if no longer qualified

**Rationale:** This maintains data integrity and prevents gaming the system. Users who accidentally complete a task and immediately undo shouldn't have inflated stats.

### Single Undo Level
**Decision:** Only the most recent action can be undone (within toast duration). No undo history.

**Rationale:** Keeps implementation simple and aligns with user expectations. Complex undo stacks add cognitive overhead.

---

## Acceptance Criteria

### AC-1: Button Loading States
```gherkin
Given I am on the task list
When I click the "Complete" button on a task
Then the button should show a spinner
And the button should be disabled
And the original button text should remain visible
And after completion, the spinner should disappear
```

### AC-2: Undo Task Completion
```gherkin
Given I have just completed a task
When the success toast appears
Then it should contain an "Undo" button
And the toast should remain visible for 5 seconds

When I click "Undo" within 5 seconds
Then the task should return to "todo" status
And the task should reappear in my active task list
And my gamification stats should be updated
And I should see "Task restored to active" confirmation
```

### AC-3: Undo Task Deletion
```gherkin
Given I have just deleted a task
When the success toast appears
Then it should contain an "Undo" button

When I click "Undo" within 5 seconds
Then the task should be restored
And it should reappear in my task list
And I should see "Task restored" confirmation
```

### AC-4: Gamification Reversal
```gherkin
Given I have completed a task that earned me an achievement
When I undo the completion
Then my total completed count should decrease by 1
And if I no longer qualify for the achievement, it should be revoked
And my category mastery for that task's category should decrease
```

---

## Out of Scope

- Undo for bulk operations (BulkDelete, BulkRestore)
- Undo for task updates/edits
- Undo history (multi-level undo)
- Permanent delete cleanup job (can be added later)
- Toast customization settings (duration preferences)

---

## Dependencies

- None (self-contained feature)

---

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Achievement revocation may frustrate users | Low | Clear messaging; only revoke if truly unqualified |
| Soft delete increases storage over time | Low | Add cleanup job later for tasks deleted >30 days |
| Race condition if user undoes during gamification processing | Low | Async processing already handles this gracefully |

---

## Success Metrics

- Reduced support requests about accidental task completion/deletion
- Improved user satisfaction with task management flow
- No increase in "gaming" behavior (repeated complete/undo)

---

## References

- [Button Loading States UX - UX Movement](https://uxmovement.com/buttons/when-you-need-to-show-a-buttons-loading-state/)
- [Toast Notifications Best Practices - LogRocket](https://blog.logrocket.com/ux-design/toast-notifications/)
- [Undo Duration Research - UX Stack Exchange](https://ux.stackexchange.com/questions/116634/how-long-should-a-toast-message-with-undo-appear)
