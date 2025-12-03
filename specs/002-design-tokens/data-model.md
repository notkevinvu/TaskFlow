# Data Model: Design System Tokens

**Feature**: 002-design-tokens
**Date**: 2025-12-03

## Overview

Design tokens are not stored in a database - they are static configuration files. This document defines the structure and schema of token definitions.

---

## Token Entity

A **Token** represents a named design value with semantic meaning.

| Property | Type | Description |
|----------|------|-------------|
| name | string | Semantic identifier (e.g., "success", "space-4") |
| cssVariable | string | CSS custom property name (e.g., "--token-success") |
| lightValue | string | Value in light mode (oklch, px, rem, etc.) |
| darkValue | string | Value in dark mode (same format as lightValue) |
| category | enum | "color" \| "spacing" \| "typography" |

---

## Color Tokens

### Semantic Status Colors

| Token Name | CSS Variable | Light Value | Dark Value | Usage |
|------------|--------------|-------------|------------|-------|
| success | --token-success | oklch(0.72 0.19 142.5) | oklch(0.80 0.16 142.5) | Success states, positive feedback |
| success-foreground | --token-success-foreground | oklch(0.985 0 0) | oklch(0.145 0 0) | Text on success background |
| warning | --token-warning | oklch(0.80 0.18 84.4) | oklch(0.85 0.15 84.4) | Warning states, caution |
| warning-foreground | --token-warning-foreground | oklch(0.145 0 0) | oklch(0.145 0 0) | Text on warning background |
| error | --token-error | oklch(0.63 0.24 25.3) | oklch(0.70 0.20 25.3) | Error states, destructive |
| error-foreground | --token-error-foreground | oklch(0.985 0 0) | oklch(0.985 0 0) | Text on error background |
| info | --token-info | oklch(0.62 0.21 250.1) | oklch(0.70 0.18 250.1) | Informational states |
| info-foreground | --token-info-foreground | oklch(0.985 0 0) | oklch(0.985 0 0) | Text on info background |

### Chart Colors (TypeScript Only)

These are for Recharts and other libraries that need actual color values:

| Token Name | Value | Usage |
|------------|-------|-------|
| chart.critical | oklch(0.63 0.24 25.3) | Critical priority in charts |
| chart.high | oklch(0.77 0.19 70.1) | High priority in charts |
| chart.medium | oklch(0.62 0.21 250.1) | Medium priority in charts |
| chart.low | oklch(0.70 0 0) | Low priority in charts |

---

## Spacing Tokens

Following Tailwind's spacing scale for consistency:

| Token Name | CSS Variable | Value | Tailwind Equivalent |
|------------|--------------|-------|---------------------|
| space-0.5 | --token-space-0-5 | 0.125rem | space-0.5 (2px) |
| space-1 | --token-space-1 | 0.25rem | space-1 (4px) |
| space-2 | --token-space-2 | 0.5rem | space-2 (8px) |
| space-3 | --token-space-3 | 0.75rem | space-3 (12px) |
| space-4 | --token-space-4 | 1rem | space-4 (16px) |
| space-5 | --token-space-5 | 1.25rem | space-5 (20px) |
| space-6 | --token-space-6 | 1.5rem | space-6 (24px) |
| space-8 | --token-space-8 | 2rem | space-8 (32px) |
| space-10 | --token-space-10 | 2.5rem | space-10 (40px) |
| space-12 | --token-space-12 | 3rem | space-12 (48px) |
| space-16 | --token-space-16 | 4rem | space-16 (64px) |

---

## Typography Tokens

### Font Sizes

| Token Name | CSS Variable | Value | Tailwind Equivalent |
|------------|--------------|-------|---------------------|
| font-size-xs | --token-font-size-xs | 0.75rem | text-xs |
| font-size-sm | --token-font-size-sm | 0.875rem | text-sm |
| font-size-base | --token-font-size-base | 1rem | text-base |
| font-size-lg | --token-font-size-lg | 1.125rem | text-lg |
| font-size-xl | --token-font-size-xl | 1.25rem | text-xl |
| font-size-2xl | --token-font-size-2xl | 1.5rem | text-2xl |
| font-size-3xl | --token-font-size-3xl | 1.875rem | text-3xl |
| font-size-4xl | --token-font-size-4xl | 2.25rem | text-4xl |

### Line Heights

| Token Name | CSS Variable | Value | Tailwind Equivalent |
|------------|--------------|-------|---------------------|
| line-height-tight | --token-line-height-tight | 1.25 | leading-tight |
| line-height-normal | --token-line-height-normal | 1.5 | leading-normal |
| line-height-relaxed | --token-line-height-relaxed | 1.75 | leading-relaxed |

### Font Weights

| Token Name | CSS Variable | Value | Tailwind Equivalent |
|------------|--------------|-------|---------------------|
| font-weight-normal | --token-font-weight-normal | 400 | font-normal |
| font-weight-medium | --token-font-weight-medium | 500 | font-medium |
| font-weight-semibold | --token-font-weight-semibold | 600 | font-semibold |
| font-weight-bold | --token-font-weight-bold | 700 | font-bold |

---

## TypeScript Type Definitions

```typescript
// Token category types
export type ColorToken =
  | 'success' | 'successForeground'
  | 'warning' | 'warningForeground'
  | 'error' | 'errorForeground'
  | 'info' | 'infoForeground';

export type ChartColorToken =
  | 'critical' | 'high' | 'medium' | 'low';

export type SpacingToken =
  | 'space0_5' | 'space1' | 'space2' | 'space3' | 'space4'
  | 'space5' | 'space6' | 'space8' | 'space10' | 'space12' | 'space16';

export type FontSizeToken =
  | 'xs' | 'sm' | 'base' | 'lg' | 'xl' | '2xl' | '3xl' | '4xl';

export type LineHeightToken = 'tight' | 'normal' | 'relaxed';

export type FontWeightToken = 'normal' | 'medium' | 'semibold' | 'bold';
```

---

## Relationships

```
Token System
├── CSS Tokens (tokens.css)
│   ├── :root (light mode values)
│   └── .dark (dark mode values)
│
└── TypeScript Tokens (lib/tokens/)
    ├── colors.ts (matches CSS colors + chart-specific)
    ├── spacing.ts (matches CSS spacing)
    └── typography.ts (matches CSS typography)
```

Both CSS and TypeScript tokens represent the same design values. CSS tokens are used in stylesheets and Tailwind utilities. TypeScript tokens are used in JavaScript/React code where CSS variables aren't supported.
