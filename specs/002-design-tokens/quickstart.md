# Quickstart: Using Design Tokens

**Feature**: 002-design-tokens
**Date**: 2025-12-03

## Overview

TaskFlow's design tokens provide a consistent, centralized way to use colors, spacing, and typography values across the application.

---

## Using CSS Tokens

### In Tailwind/CSS

```css
/* Use tokens in custom CSS */
.custom-success-message {
  background-color: var(--token-success);
  color: var(--token-success-foreground);
  padding: var(--token-space-4);
}
```

### In Tailwind's Arbitrary Values

```tsx
// Use tokens in Tailwind arbitrary value syntax
<div className="bg-[var(--token-success)] text-[var(--token-success-foreground)]">
  Success message
</div>
```

---

## Using TypeScript Tokens

### Importing Tokens

```typescript
// Import specific category
import { colors } from '@/lib/tokens';

// Or import everything
import { colors, spacing, typography } from '@/lib/tokens';
```

### In React Components

```tsx
import { colors } from '@/lib/tokens';

// For Recharts or other libraries needing actual color values
const CHART_COLORS = {
  success: colors.success,
  warning: colors.warning,
  error: colors.error,
};

// In component
<Bar fill={colors.success} />
```

### Available Token Objects

```typescript
// colors.ts
colors.success           // 'oklch(0.72 0.19 142.5)'
colors.successForeground // 'oklch(0.985 0 0)'
colors.warning           // 'oklch(0.80 0.18 84.4)'
colors.warningForeground // 'oklch(0.145 0 0)'
colors.error             // 'oklch(0.63 0.24 25.3)'
colors.errorForeground   // 'oklch(0.985 0 0)'
colors.info              // 'oklch(0.62 0.21 250.1)'
colors.infoForeground    // 'oklch(0.985 0 0)'

// Chart-specific colors (for libraries like Recharts)
colors.chart.critical    // 'oklch(0.63 0.24 25.3)'
colors.chart.high        // 'oklch(0.77 0.19 70.1)'
colors.chart.medium      // 'oklch(0.62 0.21 250.1)'
colors.chart.low         // 'oklch(0.70 0 0)'

// spacing.ts
spacing.space0_5  // '0.125rem'
spacing.space1    // '0.25rem'
spacing.space2    // '0.5rem'
spacing.space4    // '1rem'
// ... etc

// typography.ts
typography.fontSize.xs     // '0.75rem'
typography.fontSize.base   // '1rem'
typography.lineHeight.normal  // '1.5'
typography.fontWeight.bold    // '700'
```

---

## When to Use Tokens vs. Tailwind Classes

### Use Tokens When:

1. **Passing colors to third-party libraries** (Recharts, D3, etc.)
   ```tsx
   // Recharts can't use CSS variables
   <Bar fill={colors.success} />  // ✅ Use token
   ```

2. **Programmatic color selection**
   ```tsx
   const getStatusColor = (status: string) => {
     switch (status) {
       case 'success': return colors.success;
       case 'error': return colors.error;
       default: return colors.info;
     }
   };
   ```

3. **Custom CSS with semantic colors**
   ```css
   .status-indicator {
     background: var(--token-success);
   }
   ```

### Use Tailwind Classes When:

1. **Standard component styling**
   ```tsx
   // Tailwind classes are more readable for common cases
   <div className="bg-green-600 text-white">Success</div>
   ```

2. **Using shadcn/ui component variants**
   ```tsx
   // shadcn/ui already uses semantic tokens internally
   <Badge variant="destructive">Error</Badge>
   ```

3. **Layout and spacing**
   ```tsx
   // Tailwind spacing classes are more ergonomic
   <div className="p-4 gap-2">Content</div>
   ```

---

## Dark Mode Support

CSS tokens automatically switch values in dark mode:

```css
/* tokens.css */
:root {
  --token-success: oklch(0.72 0.19 142.5);  /* Light mode */
}

.dark {
  --token-success: oklch(0.80 0.16 142.5);  /* Dark mode - slightly brighter */
}
```

**Note**: TypeScript tokens export light mode values. For dark mode support in JS contexts, continue using CSS variables where possible, or implement a theme context wrapper.

---

## Common Patterns

### Status Badge with Token Colors

```tsx
import { colors } from '@/lib/tokens';

const STATUS_COLORS = {
  completed: { bg: colors.success, text: colors.successForeground },
  pending: { bg: colors.warning, text: colors.warningForeground },
  failed: { bg: colors.error, text: colors.errorForeground },
};

function StatusBadge({ status }: { status: keyof typeof STATUS_COLORS }) {
  const { bg, text } = STATUS_COLORS[status];
  return (
    <span style={{ backgroundColor: bg, color: text }}>
      {status}
    </span>
  );
}
```

### Chart with Priority Colors

```tsx
import { colors } from '@/lib/tokens';

const PRIORITY_COLORS: Record<string, string> = {
  'Critical (90-100)': colors.chart.critical,
  'High (75-89)': colors.chart.high,
  'Medium (50-74)': colors.chart.medium,
  'Low (0-49)': colors.chart.low,
};
```

---

## Troubleshooting

### Token Not Showing Correct Color

1. Ensure `tokens.css` is imported in `globals.css`
2. Check browser DevTools → Elements → Computed for the actual CSS variable value
3. Verify the variable name matches (e.g., `--token-success` not `--success`)

### TypeScript Import Errors

1. Ensure the path alias `@/lib/tokens` is configured in `tsconfig.json`
2. Check that `index.ts` exports the token you need
3. Restart TypeScript server if needed (`Ctrl+Shift+P` → "TypeScript: Restart TS Server")

---

## Reference

- **CSS Tokens File**: `frontend/app/tokens.css`
- **TypeScript Tokens**: `frontend/lib/tokens/`
- **Design System Docs**: `docs/design-system.md`
- **Full Token Definitions**: `specs/002-design-tokens/data-model.md`
