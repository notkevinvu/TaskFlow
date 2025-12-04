# Quickstart: Design Tokens Migration

**Feature**: 003-design-tokens-migration
**Date**: 2025-12-03

## Overview

This guide provides quick examples for migrating components to use design tokens.

---

## 1. Using Category Colors in Charts

### Before (Hardcoded)
```tsx
const CATEGORY_COLORS = [
  'hsl(var(--primary))',
  'hsl(210, 80%, 55%)',
  'hsl(150, 60%, 45%)',
  // ... more hardcoded values
];

<Line stroke={CATEGORY_COLORS[index]} />
```

### After (Design Tokens)
```tsx
// No import needed - use CSS variables directly
const CATEGORY_COLORS = [
  'var(--token-chart-category-1)',
  'var(--token-chart-category-2)',
  'var(--token-chart-category-3)',
  'var(--token-chart-category-4)',
  'var(--token-chart-category-5)',
  'var(--token-chart-category-6)',
  'var(--token-chart-category-7)',
  'var(--token-chart-category-8)',
];

<Line stroke={CATEGORY_COLORS[index % 8]} />
```

---

## 2. Using Heatmap Intensity

### Before (Tailwind Opacity)
```tsx
const getIntensityClass = (value: number) => {
  if (value === 0) return 'bg-muted/30';
  if (value <= 2) return 'bg-primary/20';
  if (value <= 4) return 'bg-primary/40';
  if (value <= 6) return 'bg-primary/60';
  if (value <= 8) return 'bg-primary/80';
  return 'bg-primary';
};

<div className={getIntensityClass(count)} />
```

### After (Design Tokens)
```tsx
const getIntensityStyle = (value: number): React.CSSProperties => {
  const level = Math.min(5, Math.floor(value / 2));
  return {
    backgroundColor: `var(--token-heatmap-intensity-${level})`,
  };
};

<div style={getIntensityStyle(count)} />
```

Or with CSS classes:
```css
/* In your CSS */
.heatmap-0 { background-color: var(--token-heatmap-intensity-0); }
.heatmap-1 { background-color: var(--token-heatmap-intensity-1); }
/* ... etc */
```

---

## 3. Using Semantic Colors for Status

### Before (Hardcoded Tailwind)
```tsx
// InsightCard with hardcoded colors
const insightConfig = {
  avoidance_pattern: {
    color: 'text-amber-600',
    bgColor: 'bg-amber-50 dark:bg-amber-950/30',
  },
  at_risk_alert: {
    color: 'text-red-600',
    bgColor: 'bg-red-50 dark:bg-red-950/30',
  },
};
```

### After (Design Tokens)
```tsx
const INSIGHT_TOKENS: Record<InsightType, { text: string; bg: string }> = {
  avoidance_pattern: {
    text: 'text-[var(--token-warning)]',
    bg: 'bg-[var(--token-warning)]/10',
  },
  at_risk_alert: {
    text: 'text-[var(--token-error)]',
    bg: 'bg-[var(--token-error)]/10',
  },
  peak_performance: {
    text: 'text-[var(--token-success)]',
    bg: 'bg-[var(--token-success)]/10',
  },
  quick_wins: {
    text: 'text-[var(--token-info)]',
    bg: 'bg-[var(--token-info)]/10',
  },
  deadline_clustering: {
    text: 'text-[var(--token-info)]',
    bg: 'bg-[var(--token-info)]/10',
  },
  category_overload: {
    text: 'text-[var(--token-warning)]',
    bg: 'bg-[var(--token-warning)]/10',
  },
};
```

---

## 4. Using Error Tokens in Pages

### Before (Hardcoded)
```tsx
// Login page error display
{error && (
  <div className="text-red-600 bg-red-50 border border-red-200 p-3 rounded">
    {error}
  </div>
)}
```

### After (Design Tokens)
```tsx
{error && (
  <div className="text-[var(--token-error)] bg-[var(--token-error)]/10 border border-[var(--token-error)]/30 p-3 rounded">
    {error}
  </div>
)}
```

Or create a reusable component:
```tsx
// components/ui/error-alert.tsx
export function ErrorAlert({ children }: { children: React.ReactNode }) {
  return (
    <div
      className="p-3 rounded border"
      style={{
        color: 'var(--token-error)',
        backgroundColor: 'color-mix(in oklch, var(--token-error) 10%, transparent)',
        borderColor: 'color-mix(in oklch, var(--token-error) 30%, transparent)',
      }}
    >
      {children}
    </div>
  );
}
```

---

## 5. Replacing Hardcoded Hex Values

### Before
```tsx
// CategoryChart with hardcoded hex
<Bar fill="#8884d8" />
```

### After
```tsx
<Bar fill="var(--token-chart-category-1)" />
```

---

## 6. Dark Mode Verification

After migrating a component, verify dark mode works:

1. Open the application
2. Navigate to the component
3. Toggle dark mode (usually in header/settings)
4. Verify colors adapt appropriately:
   - Light backgrounds → Dark backgrounds
   - Dark text → Light text
   - Colors remain visible and semantic

### Quick DevTools Check
```javascript
// In browser console
document.documentElement.classList.toggle('dark');
```

---

## Migration Checklist per Component

- [ ] Identify all hardcoded colors (hex, hsl, rgb, Tailwind color classes)
- [ ] Map each to appropriate design token
- [ ] Replace hardcoded values with `var(--token-*)` syntax
- [ ] Test light mode (visual parity)
- [ ] Test dark mode (proper adaptation)
- [ ] Verify TypeScript compilation passes
- [ ] Remove any unused imports

---

## Token Reference

| Use Case | Token |
|----------|-------|
| Success/positive | `var(--token-success)` |
| Warning/caution | `var(--token-warning)` |
| Error/danger | `var(--token-error)` |
| Info/neutral | `var(--token-info)` |
| Priority critical | `var(--token-chart-critical)` |
| Priority high | `var(--token-chart-high)` |
| Priority medium | `var(--token-chart-medium)` |
| Priority low | `var(--token-chart-low)` |
| Category color N | `var(--token-chart-category-N)` (1-8) |
| Heatmap level N | `var(--token-heatmap-intensity-N)` (0-5) |
