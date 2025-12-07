# UI Polish Improvements Plan

**Created:** 2025-12-07
**Branch:** `feature/phase-5c-task-templates`
**Status:** Complete

---

## Overview

This plan addresses three UI polish issues identified during user testing:
1. Missing cursor-pointer on interactive elements
2. Insufficient padding in collapsible sidebar sections
3. Poor tooltip contrast on analytics charts

---

## Issue 1: Missing `cursor-pointer` on Interactive Elements

### Problem
Several interactive elements have hover effects and click handlers but don't change the cursor to indicate clickability.

### Affected Elements

| File | Element | Line(s) |
|------|---------|---------|
| `frontend/app/(dashboard)/layout.tsx` | CollapsibleTrigger (Templates) | 174 |
| `frontend/app/(dashboard)/layout.tsx` | CollapsibleTrigger (Pomodoro) | 212 |
| `frontend/app/(dashboard)/layout.tsx` | CollapsibleTrigger (Progress) | 233 |
| `frontend/components/PomodoroWidget.tsx` | Mode toggle buttons (Focus/Short/Long) | ~105-115 |
| `frontend/components/SubtaskList.tsx` | Expand button | ~216-222 |
| `frontend/components/DependencySection.tsx` | Expand button | ~228-233 |

### Elements Already Correct
- Calendar navigation buttons (lines 81, 93) - already have `cursor-pointer`
- All shadcn Button components - inherit cursor-pointer
- Task cards in dashboard - already have cursor-pointer

### Solution
Add `cursor-pointer` class to all identified elements.

---

## Issue 2: Collapsible Section Padding

### Problem
When hovering over the Pomodoro and Progress section headers, the hover highlight appears immediately adjacent to the content below with no visual separation.

### Root Cause
`CollapsibleContent` has `className="px-4 pb-2"` which provides:
- `px-4`: Horizontal padding (16px)
- `pb-2`: Bottom padding (8px)
- **Missing**: Top padding between trigger and content

### Solution
Add `pt-2` (top padding 8px) to all `CollapsibleContent` elements:

```tsx
// Before
<CollapsibleContent className="px-4 pb-2">

// After
<CollapsibleContent className="px-4 pt-2 pb-2">
```

This creates visual breathing room between the collapsible header and its content.

---

## Issue 3: Chart Tooltip Contrast

### Problem
Recharts tooltips on the analytics page have poor text contrast, especially in dark mode. The tooltip background adapts to the theme but the text color doesn't.

### Root Cause
Current tooltip styling only sets background and border:

```tsx
<Tooltip
  contentStyle={{
    backgroundColor: 'hsl(var(--card))',
    border: '1px solid hsl(var(--border))',
    borderRadius: '6px',
  }}
/>
```

Text color defaults to Recharts' built-in dark color which doesn't adapt to dark mode.

### Affected Charts

| File | Chart Type |
|------|------------|
| `CompletionChart.tsx` | LineChart |
| `CategoryChart.tsx` | PieChart |
| `BumpChart.tsx` | BarChart |
| `PriorityChart.tsx` | BarChart |
| `CategoryTrendsChart.tsx` | AreaChart |

**Note:** `ProductivityHeatmap.tsx` uses shadcn's Tooltip component which already handles dark mode correctly.

### Solution
Add text color to contentStyle that uses the card-foreground CSS variable:

```tsx
<Tooltip
  contentStyle={{
    backgroundColor: 'hsl(var(--card))',
    border: '1px solid hsl(var(--border))',
    borderRadius: '6px',
    color: 'hsl(var(--card-foreground))',
  }}
/>
```

This ensures tooltip text adapts to both light and dark modes.

---

## Implementation Tasks

### Task 1: Add cursor-pointer to interactive elements
- [x] Update 3 CollapsibleTrigger elements in layout.tsx
- [x] Update mode toggle buttons in PomodoroWidget.tsx
- [x] Update expand button in SubtaskList.tsx
- [x] Update expand button in DependencySection.tsx

### Task 2: Add padding to collapsible sections
- [x] Update CollapsibleContent for Templates section
- [x] Update CollapsibleContent for Pomodoro section
- [x] Update CollapsibleContent for Progress section

### Task 3: Fix chart tooltip contrast
- [x] Update CompletionChart.tsx
- [x] Update CategoryChart.tsx
- [x] Update BumpChart.tsx
- [x] Update PriorityChart.tsx
- [x] Update CategoryTrendsChart.tsx

### Task 4: Documentation
- [x] Update design-system.md with cursor-pointer guidelines
- [x] Update PROJECT_STATUS.md

---

## Testing Checklist

- [x] Verify cursor changes to pointer on all collapsible headers
- [x] Verify cursor changes on Pomodoro mode buttons
- [x] Verify padding appears between collapsible headers and content
- [ ] Verify tooltip text is readable in light mode
- [ ] Verify tooltip text is readable in dark mode
- [ ] Test on different screen sizes

---

## Estimate

**Effort:** Small (~1 hour)
**Risk:** Low (CSS-only changes)
**Files Modified:** 9 files
