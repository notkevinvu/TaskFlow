# Research: Design System Tokens

**Feature**: 002-design-tokens
**Date**: 2025-12-03

## Research Summary

This document consolidates research findings for implementing design tokens in TaskFlow.

---

## 1. Token Naming Convention

**Decision**: Use `--token-` prefix for semantic tokens to distinguish from shadcn's `--` variables

**Rationale**:
- Avoids collision with existing shadcn/ui CSS variables (e.g., `--primary`, `--destructive`)
- Makes it clear which variables are from our token system vs. the component library
- Allows both systems to coexist without conflicts

**Alternatives Considered**:
- Using same names as shadcn (rejected: collision risk, harder to identify source)
- Using `--tf-` prefix (rejected: less intuitive than `--token-`)
- Using CSS Layers (rejected: over-engineering for current needs)

---

## 2. Token Categories

**Decision**: Organize tokens into three categories matching the spec requirements

### Colors (Semantic)
```
--token-success, --token-success-foreground
--token-warning, --token-warning-foreground
--token-error, --token-error-foreground
--token-info, --token-info-foreground
```

These supplement shadcn's existing color tokens (`--primary`, `--secondary`, `--destructive`, etc.)

### Spacing
```
--token-space-0.5 through --token-space-96
```

Mirrors Tailwind's spacing scale for programmatic access.

### Typography
```
--token-font-size-xs through --token-font-size-4xl
--token-line-height-tight, --token-line-height-normal, --token-line-height-relaxed
--token-font-weight-normal, --token-font-weight-medium, --token-font-weight-semibold, --token-font-weight-bold
```

**Rationale**: These three categories cover the most common use cases identified in the codebase (status colors in charts/badges, spacing in layouts, typography in headings).

---

## 3. TypeScript Token Structure

**Decision**: Export static values (not CSS variable references) in TypeScript files

**Rationale**:
- Recharts and other libraries cannot process `var(--token-x)` strings
- Static values work everywhere, including server-side rendering contexts
- Values can be kept in sync with CSS via code comments referencing tokens.css

**Structure**:
```typescript
// frontend/lib/tokens/colors.ts
export const colors = {
  success: 'oklch(0.72 0.19 142.5)',
  successForeground: 'oklch(0.985 0 0)',
  // ...
} as const;

// frontend/lib/tokens/index.ts - re-exports all token categories
export { colors } from './colors';
export { spacing } from './spacing';
export { typography } from './typography';
```

**Alternatives Considered**:
- CSS variable references in TS (rejected: incompatible with many libraries)
- JSON source of truth with CSS/TS generation (rejected: over-engineering)
- Single tokens.ts file (rejected: harder to tree-shake)

---

## 4. CSS Integration Approach

**Decision**: Import tokens.css at the top of globals.css

**Rationale**:
- Tokens defined before they're used
- Clear separation of concerns (tokens vs. base styles)
- Easy to find token definitions

**Import Order**:
```css
@import "./tokens.css";  /* New - token definitions */
@import "tailwindcss";
@import "tw-animate-css";
/* ... rest of globals.css */
```

---

## 5. Component Migration Strategy

**Decision**: Migrate PriorityChart as proof-of-concept

**Rationale**:
- Already uses CSS variables (`hsl(var(--destructive))`)
- Has hardcoded color mapping that benefits from tokens
- Demonstrates both CSS and TypeScript token usage
- Self-contained component with clear boundaries

**Pattern to Establish**:
1. Replace hardcoded `hsl(var(--x))` with imported token constants
2. Import from `@/lib/tokens`
3. Keep existing visual appearance unchanged

---

## 6. Dark Mode Support

**Decision**: Define dark mode values in `.dark` selector within tokens.css

**Rationale**:
- Consistent with existing shadcn/ui pattern in globals.css
- Works with next-themes provider already in place
- No additional configuration needed

**Structure**:
```css
:root {
  --token-success: oklch(0.72 0.19 142.5);
}

.dark {
  --token-success: oklch(0.80 0.16 142.5);
}
```

---

## 7. File Location

**Decision**: Place token files in `frontend/lib/tokens/`

**Rationale**:
- Consistent with existing `frontend/lib/` pattern (api.ts, utils.ts)
- Close to where they're consumed
- Natural import path: `@/lib/tokens`

**File Structure**:
```
frontend/
├── lib/
│   ├── tokens/
│   │   ├── index.ts      # Re-exports
│   │   ├── colors.ts     # Color tokens
│   │   ├── spacing.ts    # Spacing tokens
│   │   └── typography.ts # Typography tokens
│   ├── api.ts
│   └── utils.ts
├── app/
│   ├── globals.css       # Modified to import tokens
│   └── tokens.css        # New - CSS token definitions
```

---

## Resolved Clarifications

All technical decisions have been made. No NEEDS CLARIFICATION items remain.
