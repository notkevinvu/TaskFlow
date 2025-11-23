# TaskFlow Design System

**Version:** 1.0
**Last Updated:** 2025-01-22
**Status:** Living Document

This document tracks UI/UX patterns, component guidelines, and interaction standards for TaskFlow.

---

## Table of Contents

- [Colors](#colors)
- [Typography](#typography)
- [Spacing](#spacing)
- [Components](#components)
- [Interactions](#interactions)
- [Animations](#animations)
- [Responsive Design](#responsive-design)

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

**Pattern:** Show loading state during async operations.

**Implementation:**
```tsx
<Button disabled={isPending}>
  {isPending ? 'Loading...' : 'Submit'}
</Button>
```

**Applied to:**
- ✅ Create task button
- ✅ Bump/Complete/Delete buttons

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

## Changelog

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
