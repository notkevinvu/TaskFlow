# Quickstart: Design Tokens Migration

**Feature**: 003-design-tokens-migration
**Date**: 2025-12-03

## Overview

This guide provides quick examples for migrating components to use design tokens.

---

## 1. Using Accent Colors in Charts

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
import { tokens, ACCENT_COLORS } from '@/lib/tokens';

// Use the ACCENT_COLORS array with token values
const CATEGORY_COLORS = ACCENT_COLORS.map(
  (color) => tokens.accent[color].default
);

<Line stroke={CATEGORY_COLORS[index % CATEGORY_COLORS.length]} />
```

**Available accent colors**: blue, purple, pink, orange, yellow, green, cyan, teal

---

## 2. Using Intensity Tokens for Heatmaps

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
import { tokens, IntensityLevel } from '@/lib/tokens';

function getIntensityLevel(count: number, maxCount: number): IntensityLevel {
  if (count === 0 || maxCount === 0) return 0;
  const ratio = count / maxCount;
  if (ratio <= 0.2) return 1;
  if (ratio <= 0.4) return 2;
  if (ratio <= 0.6) return 3;
  if (ratio <= 0.8) return 4;
  return 5;
}

const level = getIntensityLevel(count, maxCount);
<div style={{ backgroundColor: tokens.intensity[level] }} />
```

**Available intensity levels**: 0 (none), 1 (low), 2, 3, 4, 5 (high)

---

## 3. Using Status Colors

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
import { tokens } from '@/lib/tokens';

const insightConfig = {
  avoidance_pattern: {
    colorToken: tokens.status.warning.default,
    bgToken: tokens.status.warning.muted,
  },
  at_risk_alert: {
    colorToken: tokens.status.error.default,
    bgToken: tokens.status.error.muted,
  },
  peak_performance: {
    colorToken: tokens.status.success.default,
    bgToken: tokens.status.success.muted,
  },
  quick_wins: {
    colorToken: tokens.status.info.default,
    bgToken: tokens.status.info.muted,
  },
};

// Usage
<div
  style={{
    color: config.colorToken,
    backgroundColor: config.bgToken,
  }}
>
  {content}
</div>
```

**Available status types**: success, warning, error, info
**Available variants**: default, muted, foreground

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
import { tokens } from '@/lib/tokens';

{error && (
  <div
    className="p-3 text-sm rounded border"
    style={{
      color: tokens.status.error.default,
      backgroundColor: tokens.status.error.muted,
      borderColor: tokens.status.error.default,
    }}
  >
    {error}
  </div>
)}
```

---

## 5. Using Surface Tokens for Backgrounds

### Before (Tailwind)
```tsx
<div className="bg-white dark:bg-gray-900">
```

### After (Design Tokens)
```tsx
import { tokens } from '@/lib/tokens';

<div style={{ backgroundColor: tokens.surface.elevated }}>
```

**Available surface tokens**: default, muted, elevated, overlay

---

## 6. Using Gradient Tokens

### Before (Tailwind)
```tsx
<div className="bg-gradient-to-br from-gray-50 to-white dark:from-gray-900 dark:to-gray-800">
```

### After (Design Tokens)
```tsx
import { tokens } from '@/lib/tokens';

<div style={{ background: tokens.gradient.surface }}>
```

**Available gradients**: primary, surface, accent

---

## 7. Replacing Hardcoded Hex Values

### Before
```tsx
// CategoryChart with hardcoded hex
<Bar fill="#8884d8" />
```

### After
```tsx
import { tokens } from '@/lib/tokens';

<Bar fill={tokens.accent.purple.default} />
```

---

## 8. Dark Mode Verification

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
- [ ] Import tokens: `import { tokens } from '@/lib/tokens';`
- [ ] Replace hardcoded values with token references
- [ ] Test light mode (visual parity)
- [ ] Test dark mode (proper adaptation)
- [ ] Verify TypeScript compilation passes
- [ ] Remove any unused imports

---

## Token Reference

### Status Colors
| Use Case | Token |
|----------|-------|
| Success/positive | `tokens.status.success.default` |
| Warning/caution | `tokens.status.warning.default` |
| Error/danger | `tokens.status.error.default` |
| Info/neutral | `tokens.status.info.default` |

### Accent Colors (for charts/visualization)
| Color | Token |
|-------|-------|
| Blue | `tokens.accent.blue.default` |
| Purple | `tokens.accent.purple.default` |
| Pink | `tokens.accent.pink.default` |
| Orange | `tokens.accent.orange.default` |
| Yellow | `tokens.accent.yellow.default` |
| Green | `tokens.accent.green.default` |
| Cyan | `tokens.accent.cyan.default` |
| Teal | `tokens.accent.teal.default` |

### Intensity Scale (for heatmaps)
| Level | Token |
|-------|-------|
| 0 (none) | `tokens.intensity[0]` |
| 1-5 | `tokens.intensity[1]` ... `tokens.intensity[5]` |

### Surface Colors
| Use Case | Token |
|----------|-------|
| Page background | `tokens.surface.default` |
| Muted sections | `tokens.surface.muted` |
| Cards/elevated | `tokens.surface.elevated` |
| Modals/overlays | `tokens.surface.overlay` |

### Helper Functions
```tsx
import { getStatusColor, getAccentColor, getIntensityColor } from '@/lib/tokens';

// Type-safe status color access
getStatusColor('warning', 'default') // 'var(--token-status-warning-default)'

// Type-safe accent color access
getAccentColor('blue', 'muted') // 'var(--token-accent-blue-muted)'

// Type-safe intensity access
getIntensityColor(3) // 'var(--token-intensity-3)'
```
