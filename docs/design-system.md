# TaskFlow Design System

**Version:** 2.1
**Last Updated:** 2025-12-04
**Status:** Living Document

This document tracks UI/UX patterns, component guidelines, and interaction standards for TaskFlow.

---

## Table of Contents

- [Design Tokens](#design-tokens)
- [Colors](#colors)
- [Typography](#typography)
- [Spacing](#spacing)
- [Components](#components)
- [Interactions](#interactions)
- [Animations](#animations)
- [Responsive Design](#responsive-design)

---

## Design Tokens

Design tokens provide a centralized, consistent way to use colors, spacing, and typography values across the application. The system includes **60 semantic tokens** organized into 8 categories.

### Token Files

| File | Purpose | Usage |
|------|---------|-------|
| `frontend/app/tokens.css` | CSS custom properties (60 tokens) | Use in CSS: `var(--token-status-success-default)` |
| `frontend/lib/tokens/tokens.ts` | Unified TypeScript API | Use in JS/TS: `tokens.status.success.default` |
| `frontend/lib/tokens/index.ts` | Re-exports + legacy compatibility | Import from: `@/lib/tokens` |

### Quick Start

```tsx
import { tokens, ACCENT_COLORS } from '@/lib/tokens';

// Status colors for alerts/feedback
<div style={{ color: tokens.status.error.default }}>Error!</div>

// Accent colors for charts/visualization
<Bar fill={tokens.accent.blue.default} />

// Surface colors for backgrounds
<div style={{ backgroundColor: tokens.surface.elevated }}>Card</div>

// Intensity scale for heatmaps
<div style={{ backgroundColor: tokens.intensity[3] }} />
```

### Token Categories (60 Total)

#### 1. Text Tokens (5 tokens)
Interface over shadcn's `--foreground` variables for typography hierarchy.

| Token | CSS Variable | Usage |
|-------|--------------|-------|
| `tokens.text.default` | `--token-text-default` | Primary text - headings, body |
| `tokens.text.secondary` | `--token-text-secondary` | Supporting text, descriptions |
| `tokens.text.tertiary` | `--token-text-tertiary` | Subtle labels, hints |
| `tokens.text.disabled` | `--token-text-disabled` | Disabled/inactive states |
| `tokens.text.inverse` | `--token-text-inverse` | Text on dark backgrounds |

#### 2. Surface Tokens (4 tokens)
Interface over shadcn's `--background` variables for backgrounds.

| Token | CSS Variable | Usage |
|-------|--------------|-------|
| `tokens.surface.default` | `--token-surface-default` | Page background |
| `tokens.surface.muted` | `--token-surface-muted` | Subtle sections |
| `tokens.surface.elevated` | `--token-surface-elevated` | Cards, raised elements |
| `tokens.surface.overlay` | `--token-surface-overlay` | Modals, popovers |

#### 3. Highlight Tokens (3 tokens)
Interface over shadcn's `--ring` and `--primary` for emphasis.

| Token | CSS Variable | Usage |
|-------|--------------|-------|
| `tokens.highlight.default` | `--token-highlight-default` | Focus rings, selection |
| `tokens.highlight.muted` | `--token-highlight-muted` | Subtle emphasis |
| `tokens.highlight.strong` | `--token-highlight-strong` | Primary actions, CTAs |

#### 4. Shadow Tokens (3 tokens)
Custom shadow values for elevation and depth.

| Token | CSS Variable | Usage |
|-------|--------------|-------|
| `tokens.shadow.light` | `--token-shadow-light` | Subtle elevation |
| `tokens.shadow.default` | `--token-shadow-default` | Standard elevation |
| `tokens.shadow.heavy` | `--token-shadow-heavy` | Pronounced elevation |

#### 5. Status Tokens (12 tokens)
Semantic colors for feedback states (success, warning, error, info).

| Status | Default | Muted | Foreground |
|--------|---------|-------|------------|
| Success | `tokens.status.success.default` | `tokens.status.success.muted` | `tokens.status.success.foreground` |
| Warning | `tokens.status.warning.default` | `tokens.status.warning.muted` | `tokens.status.warning.foreground` |
| Error | `tokens.status.error.default` | `tokens.status.error.muted` | `tokens.status.error.foreground` |
| Info | `tokens.status.info.default` | `tokens.status.info.muted` | `tokens.status.info.foreground` |

**Example - Error Alert:**
```tsx
<div
  style={{
    color: tokens.status.error.default,
    backgroundColor: tokens.status.error.muted,
    borderColor: tokens.status.error.default,
  }}
>
  Error message
</div>
```

#### 6. Accent Tokens (24 tokens)
Named colors for data visualization and categorization (8 colors × 3 variants).

| Color | Default | Muted | Strong |
|-------|---------|-------|--------|
| Blue | `tokens.accent.blue.default` | `.muted` | `.strong` |
| Purple | `tokens.accent.purple.default` | `.muted` | `.strong` |
| Pink | `tokens.accent.pink.default` | `.muted` | `.strong` |
| Orange | `tokens.accent.orange.default` | `.muted` | `.strong` |
| Yellow | `tokens.accent.yellow.default` | `.muted` | `.strong` |
| Green | `tokens.accent.green.default` | `.muted` | `.strong` |
| Cyan | `tokens.accent.cyan.default` | `.muted` | `.strong` |
| Teal | `tokens.accent.teal.default` | `.muted` | `.strong` |

**Example - Chart Colors:**
```tsx
import { tokens, ACCENT_COLORS } from '@/lib/tokens';

// Use ACCENT_COLORS array for iterating
const CHART_COLORS = ACCENT_COLORS.map(
  (color) => tokens.accent[color].default
);

<Bar fill={CHART_COLORS[index % CHART_COLORS.length]} />
```

#### 7. Intensity Tokens (6 tokens)
Scale from 0-5 for heatmaps and data density visualization.

| Level | Token | Usage |
|-------|-------|-------|
| 0 | `tokens.intensity[0]` | No activity (lightest) |
| 1 | `tokens.intensity[1]` | Low activity |
| 2 | `tokens.intensity[2]` | Moderate-low |
| 3 | `tokens.intensity[3]` | Moderate-high |
| 4 | `tokens.intensity[4]` | High activity |
| 5 | `tokens.intensity[5]` | Peak activity (darkest) |

**Example - Productivity Heatmap:**
```tsx
function getIntensityLevel(count: number, maxCount: number): IntensityLevel {
  if (count === 0 || maxCount === 0) return 0;
  const ratio = count / maxCount;
  if (ratio <= 0.2) return 1;
  if (ratio <= 0.4) return 2;
  if (ratio <= 0.6) return 3;
  if (ratio <= 0.8) return 4;
  return 5;
}

const level = getIntensityLevel(taskCount, maxTasks);
<div style={{ backgroundColor: tokens.intensity[level] }} />
```

#### 8. Gradient Tokens (3 tokens)
Pre-defined gradients for backgrounds and decorative elements.

| Token | CSS Variable | Usage |
|-------|--------------|-------|
| `tokens.gradient.primary` | `--token-gradient-primary` | Primary accent gradient |
| `tokens.gradient.surface` | `--token-gradient-surface` | Page/section backgrounds |
| `tokens.gradient.accent` | `--token-gradient-accent` | Decorative accents |

**Example - Login Page Background:**
```tsx
<div style={{ background: tokens.gradient.surface }}>
  <LoginCard />
</div>
```

### Helper Functions

Type-safe accessor functions for programmatic color selection:

```tsx
import { getStatusColor, getAccentColor, getIntensityColor } from '@/lib/tokens';

// Get status color by type and variant
getStatusColor('error', 'default')   // 'var(--token-status-error-default)'
getStatusColor('warning', 'muted')   // 'var(--token-status-warning-muted)'

// Get accent color by name and variant
getAccentColor('blue', 'default')    // 'var(--token-accent-blue-default)'
getAccentColor('purple', 'strong')   // 'var(--token-accent-purple-strong)'

// Get intensity by level
getIntensityColor(3)                 // 'var(--token-intensity-3)'
```

### When to Use Tokens vs. Tailwind

| Scenario | Use | Example |
|----------|-----|---------|
| Standard layout/spacing | Tailwind classes | `className="p-4 gap-2"` |
| shadcn/ui component variants | Component props | `<Badge variant="destructive">` |
| Charts/SVG fills | Design tokens | `fill={tokens.accent.blue.default}` |
| Status indicators | Design tokens | `color: tokens.status.error.default` |
| Heatmaps/intensity | Design tokens | `backgroundColor: tokens.intensity[level]` |
| Error/success messages | Design tokens | `tokens.status.{type}.{variant}` |
| Page backgrounds | Design tokens | `background: tokens.gradient.surface` |
| Custom semantic styling | Design tokens | Dynamic token selection |

### Dark Mode

All CSS tokens automatically adapt to dark mode via the `.dark` selector in `tokens.css`. The TypeScript API returns CSS variable references (e.g., `'var(--token-status-success-default)'`), which resolve at runtime to the appropriate light/dark value.

**No special handling needed** - just use the tokens and dark mode works automatically.

---

## Colors

### Color Palette

Using Tailwind CSS default palette + shadcn/ui theme system.

**Primary Actions:**
- `primary` - Main actions (buttons, links)
- `secondary` - Supporting actions
- `destructive` - Delete, cancel, dangerous actions (red)

**Status Colors:**
- Success: `green-600`
- Warning: `yellow-600`
- Error: `red-600`
- Info: `blue-600`

**Priority Badges:**
- High Priority (≥90): `destructive` variant (red)
- Medium Priority (75-89): `default` variant (dark)
- Low Priority (<75): `secondary` variant (gray)

**Background:**
- Light mode: `background` (white)
- Dark mode: `background` (dark gray)
- Cards: `card` with `border`

---

## Typography

### Heading Hierarchy

```tsx
// Page Title
<h2 className="text-3xl font-bold">Today's Priorities</h2>

// Section Title
<h3 className="text-lg font-semibold">Your Tasks</h3>

// Card Title
<h4 className="font-semibold">Task Title</h4>

// Stat Title
<CardTitle className="text-sm font-medium text-muted-foreground">
  Total Tasks
</CardTitle>
```

### Body Text

```tsx
// Description
<p className="text-sm text-muted-foreground">
  Task description here...
</p>

// Metadata
<span className="text-sm text-muted-foreground">
  📁 Category
</span>

// Context (italic)
<p className="text-sm italic text-muted-foreground">
  "Context from meeting"
</p>
```

---

## Spacing

### Consistent Gaps

- **Layout sections:** `gap-6` or `space-y-6`
- **Card grid:** `gap-4`
- **Form fields:** `gap-4` or `gap-2`
- **Button groups:** `gap-2`

### Padding

- **Card content:** `pt-6` (or use `CardContent`)
- **Dialog content:** `py-4`
- **Page container:** Default (no extra padding with route groups)

---

## Components

### Buttons

#### Button Variants

```tsx
// Primary action
<Button>Complete</Button>

// Secondary action
<Button variant="outline">Bump</Button>

// Destructive action
<Button variant="destructive">
  <Trash2 className="h-4 w-4" />
</Button>
```

#### Button Sizes

```tsx
<Button size="sm">Small</Button>    // Task card actions
<Button size="default">Default</Button>  // Main actions
```

#### Icon Buttons

```tsx
// Icon with text
<Button>
  <Plus className="mr-2 h-4 w-4" />
  Quick Add
</Button>

// Icon only
<Button variant="destructive" size="sm">
  <Trash2 className="h-4 w-4" />
</Button>
```

### Cards

```tsx
<Card className="hover:shadow-md transition-shadow cursor-pointer">
  <CardContent className="pt-6">
    {/* Content */}
  </CardContent>
</Card>
```

**Guidelines:**
- Use `hover:shadow-md` for interactive cards
- Add `cursor-pointer` for clickable cards
- Use `transition-shadow` for smooth hover effect

### Badges

```tsx
// Priority badge
<Badge variant={priorityVariant}>
  {Math.round(task.priority_score)}
</Badge>

// Status badge
<Badge variant="outline" className="text-yellow-600 border-yellow-600">
  ⚠️ Bumped {count}x
</Badge>

// Effort badge
<Badge variant="outline" className="capitalize">
  {task.estimated_effort}
</Badge>
```

### Dialogs

```tsx
<Dialog>
  <DialogContent className="sm:max-w-[525px]">
    <DialogHeader>
      <DialogTitle>Modal Title</DialogTitle>
      <DialogDescription>Description text</DialogDescription>
    </DialogHeader>

    <div className="grid gap-4 py-4">
      {/* Form content */}
    </div>

    <DialogFooter>
      <Button variant="outline">Cancel</Button>
      <Button>Confirm</Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
```

### Category Management

**Pattern:** Dropdown with create-new functionality + management dialog for bulk operations.

#### CategorySelect Component

**Usage:** Task creation/editing forms

```tsx
<CategorySelect
  value={category}
  onChange={(value) => setCategory(value)}
  userCategories={['Work', 'Personal', 'Bug Fix']} // From user's existing tasks
/>
```

**Features:**
- Shows existing categories from user's tasks
- Includes common category suggestions
- "Create custom category..." option reveals text input
- Auto-focuses input when creating new category
- Keyboard support (Enter to confirm, Escape to cancel)

**Applied to:**
- ✅ CreateTaskDialog
- ✅ EditTaskDialog

#### ManageCategoriesDialog Component

**Usage:** Bulk category operations

```tsx
<ManageCategoriesDialog
  open={isOpen}
  onOpenChange={setIsOpen}
/>
```

**Features:**
- Lists all categories with task counts
- Inline rename with Enter/Escape keyboard support
- Delete category with confirmation dialog
- Backend sync (PUT /api/v1/categories/rename, DELETE /api/v1/categories/:name)
- Optimistic updates with React Query

**Applied to:**
- ✅ Dashboard page ("Manage Categories" button)

**Design Notes:**
- Category badges use blue styling: `bg-blue-50 dark:bg-blue-950 text-blue-700 dark:text-blue-300 border-blue-200 dark:border-blue-800`
- Delete operation removes category from all tasks (not just deletes the label)
- Rename operation updates all tasks with that category

---

### Recurrence Components

**Pattern:** Task recurrence management with pattern selection, completion options, and visual indicators.

#### RecurrenceSelector Component

**Usage:** Task creation forms for setting up recurring tasks

```tsx
<RecurrenceSelector
  recurrence={recurrenceState}
  onChange={handleRecurrenceChange}
/>
```

**Data Format:**
```tsx
interface RecurrenceState {
  pattern: 'none' | 'daily' | 'weekly' | 'monthly';
  intervalValue: number; // e.g., 2 for "every 2 weeks"
  endDate: string | null; // ISO date string or null
  dueDateCalculation: 'from_original' | 'from_completion';
}
```

**Features:**
- Pattern dropdown with icons (`Repeat` from lucide-react)
- Interval input with inline label ("every X days/weeks/months")
- Optional end date with shadcn Calendar picker
- Due date calculation mode selector
- Collapses to single select when pattern is "none"

**Applied to:**
- ✅ CreateTaskDialog

**Design Notes:**
- Uses `grid grid-cols-2` layout for pattern and interval
- End date uses `Popover` with `Calendar` component
- Shows due date calculation only when recurrence is active
- Default values: pattern="none", intervalValue=1, dueDateCalculation="from_original"

#### RecurrenceCompletionDialog Component

**Usage:** Modal shown when completing a recurring task

```tsx
<RecurrenceCompletionDialog
  open={isOpen}
  onOpenChange={setIsOpen}
  taskId={taskId}
  category={task.category}
/>
```

**Features:**
- RadioGroup with custom-styled options (bordered with hover states)
- Icon-enhanced labels (`CalendarClock`, `CalendarCheck2` from lucide-react)
- Hierarchical checkbox interactions:
  - "Save as my default" checkbox
  - "Save for this category" checkbox (mutually exclusive behaviors)
  - "Skip next occurrence" checkbox
  - "Stop recurring" checkbox (disables other options)
- Automatic preference saving to backend

**Applied to:**
- ✅ Task completion flow (when task.series_id is set)

**Implementation Details:**
```tsx
// Custom RadioGroup styling with bordered options
<div className="flex items-center space-x-2 rounded-lg border p-3 hover:bg-accent/50 transition-colors">
  <RadioGroupItem value="from_original" id="from_original" />
  <Label htmlFor="from_original" className="flex items-center gap-2 cursor-pointer flex-1">
    <CalendarClock className="h-4 w-4 text-muted-foreground" />
    <span>From original due date</span>
  </Label>
</div>
```

**Design Notes:**
- Purple accent color (`text-purple-600`, `border-purple-300`) for recurring task branding
- Uses `Dialog` with `DialogContent` for modal presentation
- Checkbox states are mutually exclusive in some cases (handled in onChange logic)
- On confirm, calls `completeWithOptions` mutation with selected preferences

#### Recurring Task Badge

**Usage:** Visual indicator for tasks that belong to a recurring series

```tsx
{task.series_id && (
  <Badge
    variant="outline"
    className="text-purple-600 dark:text-purple-400 border-purple-300 dark:border-purple-600 bg-purple-50 dark:bg-purple-950"
  >
    <Repeat className="h-3 w-3 mr-1" />
    Recurring
  </Badge>
)}
```

**Features:**
- Purple color scheme (distinct from other badge types)
- `Repeat` icon from lucide-react
- Outline variant with custom purple colors
- Dark mode support with appropriate color adjustments

**Applied to:**
- ✅ Dashboard task cards
- ✅ TaskDetailsSidebar

**Design Notes:**
- Consistent purple branding across all recurrence-related UI
- Icon size `h-3 w-3` to fit within badge
- Uses custom class overrides on `variant="outline"` for purple theme

---

### Search & Filtering

**Pattern:** Instant search with debouncing + collapsible filter panel with active filter chips.

#### TaskSearch Component

**Usage:** Search input with debounced API calls

```tsx
<TaskSearch
  value={searchQuery}
  onChange={handleSearchChange}
  placeholder="Search tasks..."
  debounceMs={300}
/>
```

**Features:**
- 300ms debounce to prevent excessive API calls
- Clear button (X icon) appears when text entered
- Search icon visual indicator
- Controlled component pattern

**Implementation Details:**
```tsx
// Debounce logic
const [localValue, setLocalValue] = useState(value);

useEffect(() => {
  const timer = setTimeout(() => {
    onChange(localValue);
  }, debounceMs);
  return () => clearTimeout(timer);
}, [localValue, debounceMs, onChange]);
```

**Applied to:**
- ✅ Dashboard page

#### TaskFilters Component

**Usage:** Multi-criteria filtering with visual feedback

```tsx
<TaskFilters
  filters={filters}
  onChange={handleFiltersChange}
  onClear={handleClearFilters}
  availableCategories={categories} // Passed from parent to avoid duplicate fetch
/>
```

**Features:**
- Collapsible filter panel (expand/collapse button)
- Active filter count badge
- Filter chips showing active filters (removable with X button)
- "Clear all" button when filters active
- Three filter types:
  - **Status:** todo, in_progress, done
  - **Category:** User's existing categories
  - **Priority Range:** Critical (90-100), High (75-89), Medium (50-74), Low (0-49)

**Applied to:**
- ✅ Dashboard page

**Design Notes:**
- Filters collapse by default to save space
- Active filters always visible as chips (even when panel collapsed)
- Filter panel uses 3-column grid on desktop (`md:grid-cols-3`)
- Categories passed from parent to prevent duplicate task fetch

**Backend Integration:**
- Query params: `?search=text&status=todo&category=Work&min_priority=75&max_priority=100`
- Filters combine with AND logic (all must match)
- Results maintain priority sorting

#### Date Range Picker

**Usage:** Filter tasks by due date range

**Implementation:**
```tsx
<Popover>
  <PopoverTrigger asChild>
    <Button variant="outline" className="w-full justify-start text-left font-normal">
      <CalendarIcon className="mr-2 h-4 w-4" />
      {dateRange ? `${format(from, 'MMM d')} - ${format(to, 'MMM d')}` : 'Pick a date range'}
    </Button>
  </PopoverTrigger>
  <PopoverContent className="w-auto p-0" align="start">
    <Calendar
      mode="range"
      selected={dateRange}
      onSelect={handleDateRangeChange}
      numberOfMonths={2}
    />
  </PopoverContent>
</Popover>
```

**Features:**
- Two-month calendar view for easy range selection
- Displays selected range in button text
- "Clear dates" button inside popover
- Uses local timezone parsing (avoids UTC shift issues)
- Safe date formatting with fallback for invalid dates

**Applied to:**
- ✅ TaskFilters component

**Design Notes:**
- Uses shadcn/ui Calendar component with `mode="range"`
- Date strings stored as `YYYY-MM-DD` format for API compatibility
- `parseLocalDate()` helper ensures dates render in user's local timezone

---

#### Filter Presets (Quick Filters)

**Usage:** One-click filter application for common scenarios

**Implementation:**
```tsx
<div className="flex flex-wrap gap-2">
  {presets.map((preset) => (
    <Button
      key={preset.id}
      variant="outline"
      size="sm"
      onClick={() => applyPreset(preset)}
      className="h-7 text-xs"
    >
      {preset.label}
    </Button>
  ))}
</div>
```

**Available Presets:**
| Preset | Filters Applied |
|--------|-----------------|
| High Priority | `minPriority: 75, maxPriority: 100` |
| Critical Only | `minPriority: 90, maxPriority: 100` |
| Due This Week | `dueDateStart: today, dueDateEnd: endOfWeek` |
| Overdue | `dueDateStart: 2000-01-01, dueDateEnd: yesterday` |
| In Progress | `status: 'in_progress'` |

**Features:**
- Presets **replace** existing filters (don't merge) for predictable behavior
- Date-based presets use `getFilters()` function for fresh dates on each click
- Small button size (`h-7 text-xs`) to fit multiple presets in row
- Zap icon indicates "quick" actions

**Applied to:**
- ✅ TaskFilters component

**Design Notes:**
- `getFilterPresets()` returns fresh preset array on each render
- Each preset has `getFilters()` method to compute dates at click time (not module load)
- This prevents stale dates if user keeps page open across midnight

---

#### Filter URL Persistence

**Usage:** Shareable/bookmarkable filter links

**Implementation:**
```tsx
// Parse filters from URL on page load
function parseFiltersFromURL(searchParams: URLSearchParams): TaskFilterState {
  // Validate all inputs before using
  if (status && VALID_STATUSES.includes(status)) {
    filters.status = status;
  }
  // parseInt with NaN check
  if (minPriority) {
    const parsed = parseInt(minPriority, 10);
    if (!isNaN(parsed) && parsed >= 0 && parsed <= 100) {
      filters.minPriority = parsed;
    }
  }
  // Date validation (YYYY-MM-DD format)
  if (dueDateStart && isValidDateString(dueDateStart)) {
    filters.dueDateStart = dueDateStart;
  }
}

// Sync filters to URL on change
useEffect(() => {
  const newUrl = pathname + serializeFiltersToURL(filters, searchQuery, selectedTaskId);
  router.replace(newUrl, { scroll: false });
}, [filters, searchQuery, selectedTaskId]);
```

**Features:**
- All filters persist to URL query params
- URL updates use `router.replace()` (no browser history spam)
- Invalid URL params are silently ignored (graceful degradation)
- Status validated against known values
- Priority values validated (0-100 range, not NaN)
- Date strings validated (YYYY-MM-DD format)

**Applied to:**
- ✅ Dashboard page

**URL Format:**
```
/dashboard?status=todo&category=Work&minPriority=75&maxPriority=100&dueDateStart=2024-12-01&dueDateEnd=2024-12-07
```

**Design Notes:**
- Uses `{ scroll: false }` to prevent page scroll on URL updates
- Defensive parsing prevents crashes from malformed URLs
- Search query also persisted (`?search=keyword`)
- Task ID persisted for sidebar deep links

---

### Form Components

#### Select Dropdowns

**Pattern:** shadcn/ui Select with consistent styling

```tsx
<Select value={value} onValueChange={setValue}>
  <SelectTrigger>
    <SelectValue placeholder="Select option" />
  </SelectTrigger>
  <SelectContent>
    <SelectItem value="option1">Option 1</SelectItem>
    <SelectItem value="option2">Option 2</SelectItem>
  </SelectContent>
</Select>
```

**Applied to:**
- ✅ TaskFilters (status, category, priority range)
- ✅ CategorySelect (category selection)
- ✅ Analytics time period selector
- ✅ CreateTaskDialog (estimated effort)

**Guidelines:**
- Use `SelectValue` placeholder for empty state
- Group related options with `SelectGroup` (optional)
- Use `SelectSeparator` between option groups
- Add "All X" option for filters (e.g., "All statuses")

#### Text Inputs

**Pattern:** shadcn/ui Input with consistent sizing

```tsx
<Input
  type="text"
  placeholder="Task title"
  value={value}
  onChange={(e) => setValue(e.target.value)}
  className="pl-10" // For icon padding if needed
/>
```

**Applied to:**
- ✅ CreateTaskDialog (title, description)
- ✅ EditTaskDialog (title, description)
- ✅ TaskSearch (search input)
- ✅ CategorySelect (custom category input)
- ✅ ManageCategoriesDialog (rename input)

**Guidelines:**
- Always include placeholder text
- Use `type="text"` for general input
- Use `type="email"` for email fields
- Add icons with `absolute` positioning and `pl-10` padding

---

### Chart Components

**Pattern:** Recharts library with consistent styling and responsive design.

**Common Features Across All Charts:**
- Responsive container (`<ResponsiveContainer width="100%" height={300}>`)
- Consistent color scheme using CSS variables (`hsl(var(--primary))`)
- Card wrapper with title and description
- Empty state handling ("No data available")
- Tooltips with dark mode support
- Loading skeleton states

#### CompletionChart (Line Chart)

**Usage:** Daily task completion trends

```tsx
<CompletionChart data={velocityMetrics} />
```

**Data Format:**
```tsx
interface VelocityMetric {
  date: string; // "2025-01-22"
  completed_count: number;
}
```

**Features:**
- Line chart with monotone curve
- CartesianGrid with dashed lines
- X-axis shows dates (MMM DD format)
- Y-axis shows task count (integers only)
- Primary color line with dots
- Tooltip shows date and count

**Applied to:**
- ✅ Analytics page (full-width span)

#### CategoryChart (Pie Chart)

**Usage:** Category distribution with completion rates

```tsx
<CategoryChart data={categoryBreakdown} />
```

**Data Format:**
```tsx
interface CategoryStat {
  category: string;
  task_count: number;
  completion_rate: number; // 0-100
}
```

**Features:**
- Pie chart with labels showing percentages
- Colors from `--chart-1` through `--chart-5` CSS variables
- Tooltip shows task count and completion rate
- Legend below chart
- Handles "Uncategorized" category

**Applied to:**
- ✅ Analytics page

#### PriorityChart (Bar Chart)

**Usage:** Task distribution across priority ranges

```tsx
<PriorityChart data={priorityDistribution} />
```

**Data Format:**
```tsx
interface PriorityDistribution {
  priority_range: string; // "Critical (90-100)"
  task_count: number;
}
```

**Features:**
- Color-coded bars:
  - Critical: `hsl(var(--destructive))` (red)
  - High: `hsl(var(--chart-5))` (orange)
  - Medium: `hsl(var(--primary))` (blue)
  - Low: `hsl(var(--muted))` (gray)
- Rounded bar tops (`radius={[8, 8, 0, 0]}`)
- X-axis shows abbreviated labels ("Critical", "High", etc.)
- Tooltip shows full range and count

**Applied to:**
- ✅ Analytics page

#### PriorityBreakdownPanel (Donut + Table)

**Usage:** Visual breakdown of task priority score calculation

```tsx
<PriorityBreakdownPanel
  breakdown={task.priority_breakdown}
  finalScore={task.priority_score}
/>
```

**Data Format:**
```tsx
interface PriorityBreakdown {
  // Raw values (0-100 scale, except effort_boost 1.0-1.3)
  user_priority: number;
  time_decay: number;
  deadline_urgency: number;
  bump_penalty: number;
  effort_boost: number;
  // Weighted contributions (after weights and effort boost)
  user_priority_weighted: number;    // × 0.4 × effort_boost
  time_decay_weighted: number;       // × 0.3 × effort_boost
  deadline_urgency_weighted: number; // × 0.2 × effort_boost
  bump_penalty_weighted: number;     // × 0.1 × effort_boost
}
```

**Features:**
- **Donut chart** showing contribution proportions (filtered to non-zero factors)
- **Detailed table** with raw values, weights, and final points
- **Color-coded** factors using accent tokens:
  - User Priority: `tokens.accent.blue`
  - Time Decay: `tokens.accent.purple`
  - Deadline Urgency: `tokens.accent.orange`
  - Bump Penalty: `tokens.accent.pink`
  - Effort Boost: `tokens.accent.green`
- **Effort boost section** only shown when multiplier ≠ 1.0
- **Final score** colored by priority level (red/yellow/default)
- **Formula explanation** at bottom for educational context
- **Analytics tracking** logs view events for panel usage metrics

**Applied to:**
- ✅ TaskDetailsSidebar (replaces basic priority text explanation)

**Implementation Notes:**
- Backend calculates breakdown in `task_service.Get()` method
- Breakdown only populated for single task fetch (not list views)
- Uses Recharts PieChart with inner radius for donut style
- Tooltips match existing chart styling patterns

---

#### BumpChart (Bar Chart with Stats)

**Usage:** Bump distribution analysis

```tsx
<BumpChart data={bumpAnalytics} />
```

**Data Format:**
```tsx
interface BumpAnalytics {
  average_bump_count: number;
  at_risk_count: number;
  bump_distribution: Record<string, number>; // {"0 bumps": 5, "1-2 bumps": 3}
}
```

**Features:**
- Stat cards above chart (Average Bumps, At Risk count)
- Bar chart showing distribution across bump count ranges
- Consistent primary color
- Sorted by bump count (0 bumps → 6+ bumps)

**Applied to:**
- ✅ Analytics page (full-width span)

**Chart Design Guidelines:**
- Always wrap in Card component
- Include CardHeader with CardTitle
- Set explicit height (e.g., `h-80`, `h-96`)
- Use CSS variables for colors (supports dark mode)
- Provide empty state message
- Match loading skeleton to final chart size
- Use grid layout: `lg:grid-cols-2` for side-by-side, `lg:col-span-2` for full-width

---

### Calendar Components

#### Calendar (Mini-Calendar Widget)

**Usage:** Sidebar calendar with task count indicators

```tsx
<Calendar
  onTaskClick={(taskId) => router.push(`/dashboard?taskId=${taskId}`)}
  onCreateTask={(dueDate) => setCreateDialogOpen(true)}
/>
```

**Features:**
- Mini calendar format (compact month view)
- Task count badges on dates (red/yellow/blue based on count)
- Click date to show tasks in popover
- Click "Add Task" in popover to create task with pre-filled due date
- Fetches tasks for current month from backend

**Applied to:**
- ✅ Dashboard layout sidebar

**Design Notes:**
- Badge colors: 3+ tasks (red), 1-2 tasks (yellow), upcoming (blue)
- Calendar updates when tasks are created/completed/deleted
- Uses React Query for automatic refresh

#### CalendarTaskPopover

**Usage:** Popover showing tasks for selected date

```tsx
<CalendarTaskPopover
  date={selectedDate}
  tasks={tasksForDate}
  onTaskClick={handleTaskClick}
  onCreateTask={handleCreateTask}
/>
```

**Features:**
- Lists tasks due on selected date
- Shows priority score badges
- "Add Task" button with pre-filled due date
- Collision detection to avoid viewport overflow
- Responsive positioning

**Applied to:**
- ✅ Calendar component (triggered by date click)

**Implementation Details:**
```tsx
<PopoverContent
  align="start"
  side="right"
  collisionPadding={16}
  avoidCollisions={true}
>
```

---

## Interactions

### Hover States

*Pattern established:* All interactive elements should provide visual feedback on hover.

#### Buttons

**Pattern:** Buttons should have visible hover feedback with smooth transitions.

**Implementation:**
```tsx
<Button className="transition-all hover:scale-105 hover:shadow-md cursor-pointer">
  Click me
</Button>

// Destructive buttons get enhanced shadow
<Button variant="destructive" className="transition-all hover:scale-105 hover:shadow-lg cursor-pointer">
  Delete
</Button>
```

**Applied to:**
- ✅ Dashboard task action buttons (Edit, Bump, Complete, Delete)
- ✅ Sidebar action buttons (Edit, Bump, Complete, Delete)
- ✅ Quick Add button
- ✅ Sign Out button
- ✅ Theme toggle button

**Details:**
- `transition-all` - Smooth transitions for all properties
- `hover:scale-105` - Slight grow effect (5% scale)
- `hover:shadow-md` - Medium drop shadow for depth (`hover:shadow-lg` for destructive)
- `cursor-pointer` - Change cursor to indicate clickability

---

#### Interactive Cards

**Pattern:** Cards that are clickable should indicate interactivity.

**Implementation:**
```tsx
<Card className="hover:shadow-md transition-shadow cursor-pointer">
  {/* Card content */}
</Card>
```

**Applied to:**
- ✅ Task cards on dashboard
- ⏳ Category cards (future)

---

### Focus States

**Pattern:** All interactive elements should have keyboard-accessible focus states.

**Implementation:**
- Default shadcn/ui components include focus rings
- Use `focus-visible:ring-2 focus-visible:ring-offset-2`

---

### Loading States

**Pattern:** Provide immediate visual feedback for all async operations with progressive loading states.

#### Page-Level Loading (Navigation)

**Pattern:** Use Next.js `loading.tsx` files for instant navigation feedback while page bundles load.

**Implementation:**
```tsx
// app/(dashboard)/analytics/loading.tsx
import { Skeleton } from "@/components/ui/skeleton";

export default function AnalyticsLoading() {
  return (
    <div className="space-y-6">
      <Skeleton className="h-10 w-64" />
      <div className="grid gap-4 md:grid-cols-4">
        {[...Array(4)].map((_, i) => (
          <Skeleton key={i} className="h-28" />
        ))}
      </div>
      {/* More skeleton components matching page layout */}
    </div>
  );
}
```

**Applied to:**
- ✅ Dashboard page (`/dashboard/loading.tsx`)
- ✅ Analytics page (`/analytics/loading.tsx`)

**Benefits:**
- Instant navigation (no delay when clicking nav links)
- Shows skeleton UI while JavaScript bundle loads
- Consistent with Next.js App Router best practices

---

#### Data Fetching Loading

**Pattern:** Show skeleton states while React Query fetches data, with graceful degradation.

**Implementation:**
```tsx
const { data, isLoading, isFetching } = useQuery(...);

// Full skeleton on initial load
if (isLoading && !data) {
  return <Skeleton className="h-64" />;
}

// Subtle loading indicator for refetches
{isFetching && <Loader2 className="h-4 w-4 animate-spin" />}

// Fade content during background updates
<div className={`transition-opacity ${isFetching ? 'opacity-50' : 'opacity-100'}`}>
  {/* Content */}
</div>
```

**Applied to:**
- ✅ Dashboard task list (shows skeleton initially, opacity fade on refetch)
- ✅ Analytics charts (shows skeleton while fetching)
- ✅ Task details sidebar

**Key Principles:**
- Use `isLoading` for initial load (show full skeleton)
- Use `isFetching` for background updates (show subtle indicator)
- Keep existing content visible during refetch with opacity fade
- Only reload specific sections, not entire page

---

#### Button/Action Loading

**Pattern:** Disable and show loading state during button actions.

**Implementation:**
```tsx
<Button disabled={isPending}>
  {isPending ? 'Loading...' : 'Submit'}
</Button>
```

**Applied to:**
- ✅ Create task button
- ✅ Bump/Complete/Delete buttons
- ✅ Category management buttons

---

### Disabled States

**Pattern:** Disabled buttons should be visually distinct.

**Implementation:**
- Use `disabled` prop on Button component
- Default styles: reduced opacity, no cursor change

---

## Animations

### Transition Timings

**Standard transitions:**
- `transition-all` - Default Tailwind timing (150ms)
- `transition-shadow` - For hover shadow effects
- Custom: `duration-[180ms]` - Sidebar slide animation

### Slide Animations

**Sidebar slide:**
```tsx
// Container adjusts margin when sidebar opens
<div className={`transition-all duration-[180ms] ${
  selectedTaskId ? 'lg:pr-96' : ''
}`}>
```

**Sidebar component:**
```tsx
<aside className="fixed right-0 top-0 h-full w-96
  transform transition-transform duration-[180ms]
  translate-x-0 shadow-2xl">
```

---

### Scale Animations

**Hover grow:**
- `hover:scale-105` - Subtle 5% scale increase
- Combine with `transition-all` for smoothness

---

## Responsive Design

### Breakpoints

Follow Tailwind defaults:
- `sm`: 640px
- `md`: 768px
- `lg`: 1024px
- `xl`: 1280px
- `2xl`: 1536px

### Layout Patterns

**Stats Grid:**
```tsx
<div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
  {/* Stat cards */}
</div>
```

**Task List:**
- Full width on mobile
- Adjusts with sidebar on desktop (`lg:pr-96` when sidebar open)

**Dialog Widths:**
```tsx
<DialogContent className="sm:max-w-[525px]">
```

---

## Form Patterns

### Input Fields

```tsx
<div className="grid gap-2">
  <Label htmlFor="field">Field Label</Label>
  <Input
    id="field"
    placeholder="Placeholder text"
    required
  />
</div>
```

### Select Dropdowns

#### Standard Dropdown

```tsx
<Select value={value} onValueChange={setValue}>
  <SelectTrigger>
    <SelectValue />
  </SelectTrigger>
  <SelectContent>
    <SelectItem value="option1">Option 1</SelectItem>
    <SelectItem value="option2">Option 2</SelectItem>
  </SelectContent>
</Select>
```

#### Priority Dropdown (1-10 Scale)

**Pattern:** User priority uses a 1-10 dropdown with labeled extremes.

```tsx
<Select
  value={formData.user_priority.toString()}
  onValueChange={(value) => setFormData({ ...formData, user_priority: parseInt(value) })}
>
  <SelectTrigger id="priority">
    <SelectValue />
  </SelectTrigger>
  <SelectContent>
    <SelectItem value="1">1 - Lowest</SelectItem>
    <SelectItem value="2">2</SelectItem>
    <SelectItem value="3">3</SelectItem>
    <SelectItem value="4">4</SelectItem>
    <SelectItem value="5">5 - Medium</SelectItem>
    <SelectItem value="6">6</SelectItem>
    <SelectItem value="7">7</SelectItem>
    <SelectItem value="8">8</SelectItem>
    <SelectItem value="9">9</SelectItem>
    <SelectItem value="10">10 - Highest</SelectItem>
  </SelectContent>
</Select>
```

**Applied to:**
- ✅ CreateTaskDialog (default: 5)
- ✅ EditTaskDialog
- ✅ TaskDetailsSidebar displays as "X/10"

### Textarea

```tsx
<Textarea
  placeholder="Add more details..."
  rows={3}
/>
```

---

## Accessibility

### Keyboard Navigation

- All interactive elements must be keyboard accessible
- Use semantic HTML (`<button>`, `<a>`, `<form>`)
- Focus states must be visible

### Screen Readers

- Use proper ARIA labels where needed
- Ensure proper heading hierarchy
- Form inputs must have associated labels

### Color Contrast

- Ensure sufficient contrast for text
- Don't rely on color alone for information
- Test with dark mode enabled

---

## Icons

### Icon Library

Using **Lucide React** for icons.

**Common icons:**
- `Plus` - Add/create actions
- `Trash2` - Delete actions
- `Edit` - Edit actions (future)
- `Calendar` - Date-related features
- `Users` - People/collaboration

**Icon Sizing:**
- Small (button icons): `h-4 w-4`
- Medium (inline): `h-5 w-5`
- Large (standalone): `h-6 w-6`

**Icon Placement:**
```tsx
// Icon before text
<Button>
  <Plus className="mr-2 h-4 w-4" />
  Add Task
</Button>

// Icon only
<Button size="sm">
  <Trash2 className="h-4 w-4" />
</Button>
```

---

## Dark Mode

**Status:** ✅ Implemented

**Implementation:**

### Theme Provider

Using `next-themes` for seamless theme switching with system preference support.

**Setup:**
```tsx
// app/providers.tsx
import { ThemeProvider } from '@/components/ThemeProvider';

<ThemeProvider
  attribute="class"
  defaultTheme="system"
  enableSystem
  disableTransitionOnChange
>
  {children}
</ThemeProvider>
```

### Theme Toggle

**Component:** `components/ThemeToggle.tsx`

**Features:**
- Dropdown menu with 3 options: Light, Dark, System
- Icon animates between sun (light) and moon (dark)
- Smooth transitions using Tailwind `dark:` classes
- Includes hover effects matching design system

**Location:** Bottom of sidebar, next to Sign Out button

**Usage:**
```tsx
import { ThemeToggle } from '@/components/ThemeToggle';

<ThemeToggle />
```

### Dark Mode Support

**Requirements for components:**
- Use `dark:` prefix for dark mode classes
- Leverage CSS variables from globals.css
- Test in all three modes (light, dark, system)

**Important:** While shadcn/ui components automatically support dark mode through CSS variables, layout backgrounds require explicit `dark:` classes.

**Background Colors:**
```tsx
// Main layout container
<div className="bg-gray-50 dark:bg-gray-950">

// Sidebar/panel backgrounds
<div className="bg-white dark:bg-gray-900">

// Borders
<div className="border-gray-200 dark:border-gray-800">
```

**Applied to:**
- ✅ Dashboard layout (main container and left sidebar)
- ✅ TaskDetailsSidebar (right slide-out panel)
- ✅ All shadcn/ui components (automatic via CSS variables)

**Example:**
```tsx
<div className="bg-white dark:bg-gray-900 text-gray-900 dark:text-gray-100 border-gray-200 dark:border-gray-800">
  Content
</div>
```

---

## Future Considerations

### Patterns to Define Later

- ⏳ Calendar widget styling
- ⏳ Category color system
- ⏳ Toast notification styling
- ⏳ Empty state illustrations
- ⏳ Error state patterns
- ⏳ Skeleton loading patterns (currently using shadcn/ui Skeleton)

---

## Accessibility

**Priority:** All components must be keyboard-accessible and screen-reader friendly.

### Keyboard Navigation

#### General Principles
- **Tab order:** Follows logical visual order (top to bottom, left to right)
- **Focus indicators:** All interactive elements have visible focus rings (`focus-visible:ring-2`)
- **Escape key:** Closes dialogs, popovers, and dropdowns
- **Enter key:** Confirms actions, submits forms
- **Arrow keys:** Navigate through lists, calendar dates, select options

#### Component-Specific Keyboard Support

**Dialogs (Modals):**
- `Escape` - Closes dialog
- `Tab` - Cycles through focusable elements within dialog
- Focus trapped within dialog when open
- First focusable element auto-focused on open
- Focus returned to trigger element on close

**Applied to:** CreateTaskDialog, EditTaskDialog, ManageCategoriesDialog

**Select Dropdowns:**
- `Space` or `Enter` - Opens dropdown
- `Arrow Up/Down` - Navigate options
- `Enter` - Selects option
- `Escape` - Closes dropdown
- Type-ahead search (typing filters options)

**Applied to:** CategorySelect, TaskFilters, Analytics time period selector

**Calendar:**
- `Arrow keys` - Navigate dates
- `Enter` - Selects date/opens popover
- `Escape` - Closes popover
- `Tab` - Moves to next interactive element

**Applied to:** Calendar widget, CalendarTaskPopover

**Search & Filters:**
- `Enter` - Executes search immediately (bypasses debounce)
- `Escape` - Clears search input
- `Tab` - Moves between filter controls
- Filter chips removable with `Enter` or `Space` when focused

**Applied to:** TaskSearch, TaskFilters

**Category Management:**
- `Enter` - Confirms rename, saves new category
- `Escape` - Cancels rename, discards input
- Inline editing auto-focuses input field

**Applied to:** CategorySelect, ManageCategoriesDialog

---

### Screen Reader Support

#### ARIA Labels

**Interactive Elements:**
```tsx
// Button with icon only
<Button aria-label="Delete task">
  <Trash2 className="h-4 w-4" />
</Button>

// Search input
<Input
  type="text"
  placeholder="Search tasks..."
  aria-label="Search tasks"
/>

// Select dropdown
<Select aria-label="Filter by status">
  <SelectTrigger>
    <SelectValue placeholder="All statuses" />
  </SelectTrigger>
</Select>
```

**Applied to:**
- ✅ All icon-only buttons
- ✅ Search inputs
- ✅ Filter dropdowns
- ✅ Theme toggle

**Status Messages:**
```tsx
// Loading state
<div role="status" aria-live="polite">
  <Loader2 className="animate-spin" />
  <span className="sr-only">Loading analytics data...</span>
</div>

// Error message
<div role="alert" aria-live="assertive">
  Unable to load data. Please try again.
</div>
```

**Applied to:**
- ✅ Loading skeletons (page-level)
- ✅ Error messages in analytics page
- ✅ Form validation errors

**Live Regions:**
```tsx
// Task count updates
<div aria-live="polite" aria-atomic="true">
  Showing {tasks.length} tasks
</div>

// Filter updates
<div aria-live="polite">
  {activeFilterCount} filters active
</div>
```

**Applied to:**
- ✅ Task list count
- ✅ Filter count badge
- ✅ Search results

---

### Focus Management

#### Focus Visible

All interactive elements use `focus-visible:ring-2` for keyboard focus indicators (mouse clicks don't show ring, keyboard navigation does).

```tsx
// Already applied by shadcn/ui default styles
outline-ring/50
focus-visible:ring-2
focus-visible:ring-offset-2
```

**Applied to:**
- ✅ All buttons
- ✅ All inputs
- ✅ All select dropdowns
- ✅ All links

#### Auto-Focus

**When to use:**
- First input in dialogs (CreateTaskDialog title input)
- Custom category input when "Create custom..." selected
- Rename input in ManageCategoriesDialog
- Search input on page load (optional, can be intrusive)

**When not to use:**
- Auto-focusing on page load (disruptive to keyboard users)
- Auto-focusing elements outside viewport
- Auto-focusing destructive actions

**Applied to:**
- ✅ CreateTaskDialog (title input auto-focused)
- ✅ CategorySelect (custom input auto-focused when revealed)
- ✅ ManageCategoriesDialog (rename input auto-focused)

---

### Color Contrast

**WCAG AA Compliance** (4.5:1 for normal text, 3:1 for large text)

**Text Colors:**
- Primary text: `foreground` - High contrast with background
- Muted text: `muted-foreground` - Passes AA for large text (14px+ bold, 18px+ regular)
- Links: `primary` - Sufficient contrast, underline on focus/hover

**Button Colors:**
- Default button: Sufficient contrast in both light and dark modes
- Destructive button: Red background with white text (passes AAA)
- Outline button: Border + text meet AA standards

**Status Colors:**
- Success (green): `green-600` in light, `green-400` in dark
- Warning (yellow): `yellow-600` in light, `yellow-400` in dark
- Error (red): `red-600` in light, `red-400` in dark

**Charts:**
- All chart colors tested for sufficient contrast with backgrounds
- Tooltips have high contrast dark backgrounds
- Priority color-coding includes both color and position for color-blind users

**Applied to:**
- ✅ All text elements
- ✅ All buttons
- ✅ All badges
- ✅ All charts
- ✅ Dark mode variants

---

### Semantic HTML

**Use semantic elements** for better screen reader comprehension:

```tsx
// Page structure
<main>         // Main content area
  <header>     // Page/section header
  <nav>        // Navigation (sidebar)
  <section>    // Content sections
  <article>    // Independent content (task cards)
</main>

// Forms
<form>
  <label htmlFor="title">Title</label>
  <input id="title" />
  <button type="submit">Save</button>
</form>
```

**Applied to:**
- ✅ Dashboard layout (nav, main, header structure)
- ✅ All forms (proper label associations)
- ✅ Task cards (article elements)

---

### Touch Targets

**Minimum size:** 44×44px (WCAG AA mobile guidelines)

**Implementation:**
- All buttons use `size="sm"` (minimum 40px) or `size="default"` (44px+)
- Interactive cards have generous padding for large touch areas
- Calendar dates have minimum 40px tap area
- Filter chips have adequate spacing between them

**Applied to:**
- ✅ All buttons
- ✅ All interactive cards
- ✅ Calendar dates
- ✅ Filter chips
- ✅ Select dropdowns

---

### Testing Checklist

**Keyboard Testing:**
- [ ] Tab through entire page without mouse
- [ ] All interactive elements reachable via keyboard
- [ ] Focus indicators visible on all elements
- [ ] Dialogs trap focus correctly
- [ ] Escape key works for all dismissible components

**Screen Reader Testing:**
- [ ] Test with NVDA (Windows) or VoiceOver (Mac)
- [ ] All buttons announce their purpose
- [ ] Form fields have associated labels
- [ ] Error messages are announced
- [ ] Loading states are announced
- [ ] Dynamic content updates are announced

**Color Contrast Testing:**
- [ ] Use browser DevTools contrast checker
- [ ] Test all text colors against backgrounds
- [ ] Test button colors in both states
- [ ] Test chart colors for differentiation
- [ ] Verify dark mode contrast ratios

**Future Improvements:**
- [ ] Add skip navigation link ("Skip to main content")
- [ ] Add keyboard shortcuts documentation
- [ ] Add reduced motion support (`prefers-reduced-motion`)
- [ ] Add high contrast mode support
- [ ] Comprehensive automated accessibility testing (axe-core, Lighthouse)

---

## Changelog

### 2025-12-04 (Priority Explanation Panel - v2.1)
- ✅ **New Feature:** PriorityBreakdownPanel component (Phase 5A)
- ✅ Added donut chart visualization for priority factor contributions
- ✅ Added detailed breakdown table with raw values, weights, and points
- ✅ Color-coded factors using accent tokens (blue, purple, orange, pink, green)
- ✅ Effort boost section conditionally shown when multiplier ≠ 1.0
- ✅ Final score colored by priority level (red/yellow/default)
- ✅ Formula explanation for educational context
- ✅ Analytics tracking for panel view events
- ✅ Backend: Added `PriorityBreakdown` struct to domain/task.go
- ✅ Backend: Added `CalculateWithBreakdown()` to priority/calculator.go
- ✅ Backend: Task service populates breakdown on single task fetch
- ✅ Frontend: Added `PriorityBreakdown` interface to api.ts
- ✅ Frontend: Created PriorityBreakdownPanel.tsx component
- ✅ Frontend: TaskDetailsSidebar now uses breakdown panel when available
- ✅ Version bumped to 2.1

### 2025-12-03 (Design Token Migration - v2.0)
- ✅ **Major Update:** Complete 60-token design system (Feature 003)
- ✅ Expanded `frontend/app/tokens.css` to include all 60 tokens:
  - **Text tokens** (5): default, secondary, tertiary, disabled, inverse
  - **Surface tokens** (4): default, muted, elevated, overlay
  - **Highlight tokens** (3): default, muted, strong
  - **Shadow tokens** (3): light, default, heavy
  - **Status tokens** (12): success/warning/error/info × default/muted/foreground
  - **Accent tokens** (24): 8 colors × 3 variants (default/muted/strong)
  - **Intensity tokens** (6): 0-5 scale for heatmaps
  - **Gradient tokens** (3): primary, surface, accent
- ✅ Created unified TypeScript API in `frontend/lib/tokens/tokens.ts`
  - Single `tokens` object with all 60 tokens
  - Helper functions: `getStatusColor()`, `getAccentColor()`, `getIntensityColor()`
  - `ACCENT_COLORS` array for chart iteration
  - Full TypeScript type exports
- ✅ Migrated 11 components to use design tokens:
  - **Chart components** (5): CategoryTrendsChart, CategoryChart, ProductivityHeatmap, BumpChart, CompletionChart
  - **Dashboard components** (3): InsightCard, InsightsList, TaskDetailsSidebar
  - **Page components** (3): Login, Register, Analytics
- ✅ All hardcoded HSL/hex colors replaced with semantic tokens
- ✅ Dark mode works automatically via CSS variable resolution
- ✅ Version bumped to 2.0

### 2025-12-03 (Design Tokens - v1.3)
- ✅ **New Section:** Design Tokens system for semantic values
- ✅ Created `frontend/app/tokens.css` with CSS custom properties
  - Color tokens: success, warning, error, info (with foreground variants)
  - Spacing tokens: matching Tailwind scale (space-0.5 through space-16)
  - Typography tokens: font-size, line-height, font-weight
  - Dark mode support via `.dark` selector
- ✅ Created TypeScript token modules in `frontend/lib/tokens/`
  - `colors.ts` - Semantic colors + chart-specific colors
  - `spacing.ts` - Spacing scale
  - `typography.ts` - Font size, line height, font weight
  - `index.ts` - Re-exports for easy imports
- ✅ Migrated PriorityChart to use design tokens (proof-of-concept)
- ✅ Added comprehensive token documentation with usage examples
- ✅ Documented when to use tokens vs. Tailwind classes
- ✅ Version bumped to 1.3

### 2025-12-02 (Filter Enhancements - v1.2)
- ✅ Added Date Range Picker pattern (shadcn Calendar with `mode="range"`)
- ✅ Added Filter Presets pattern (Quick Filters with replace behavior)
- ✅ Added Filter URL Persistence pattern (shareable/bookmarkable links)
- ✅ Documented `parseLocalDate()` for timezone-safe date handling
- ✅ Documented `safeFormatDate()` for graceful error handling
- ✅ Updated TaskFilters documentation with new features
- ✅ Version bumped to 1.2

### 2025-01-27 (Phase 2.5 Completion - v1.1)
- ✅ **Major Update:** Comprehensive Phase 2.5 component documentation
- ✅ Added Category Management pattern (CategorySelect, ManageCategoriesDialog)
- ✅ Added Search & Filtering pattern (TaskSearch, TaskFilters with collapsible panel)
- ✅ Added Form Components section (Select dropdowns, Text inputs with guidelines)
- ✅ Added Chart Components pattern (CompletionChart, CategoryChart, PriorityChart, BumpChart)
- ✅ Added Calendar Components pattern (Calendar widget, CalendarTaskPopover)
- ✅ **New Section:** Comprehensive Accessibility guidelines
  - Keyboard navigation for all component types
  - Screen reader support (ARIA labels, status messages, live regions)
  - Focus management (focus-visible, auto-focus guidelines)
  - Color contrast (WCAG AA compliance)
  - Semantic HTML patterns
  - Touch targets (44×44px minimum)
  - Accessibility testing checklist
- ✅ Updated Loading States section with page-level, data-fetching, and button loading patterns
- ✅ Documented all 15 custom components from Phase 2.5
- ✅ Added implementation details, code examples, and applied-to lists for all patterns
- ✅ Version bumped to 1.1

### 2025-01-22 (Update 4)
- ✅ Changed user priority from 0-100 input to 1-10 dropdown
- ✅ Added priority dropdown pattern to design system
- ✅ Updated CreateTaskDialog with priority dropdown (default: 5)
- ✅ Updated EditTaskDialog with priority dropdown
- ✅ Updated TaskDetailsSidebar to display priority as X/10
- ✅ Backend scaled to convert 1-10 to 0-100 for calculations

### 2025-01-22 (Update 3)
- ✅ Fixed dark mode backgrounds for layout sections
- ✅ Added `dark:bg-gray-950` to main container
- ✅ Added `dark:bg-gray-900` to sidebars (left nav and TaskDetailsSidebar)
- ✅ Added `dark:border-gray-800` to all borders
- ✅ Added `dark:text-gray-300` to navigation links
- ✅ Documented dark mode background pattern

### 2025-01-22 (Update 2)
- ✅ Implemented hover states for all buttons (scale + shadow effects)
- ✅ Added dark mode support with next-themes
- ✅ Created ThemeToggle component with dropdown menu
- ✅ Added action buttons to TaskDetailsSidebar
- ✅ Documented all new interaction patterns

### 2025-01-22 (Initial)
- Initial design system document created
- Documented existing button, card, and badge patterns
- Defined hover state standards
- Documented current animation timings
