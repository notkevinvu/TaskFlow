# Data Model: Design Tokens Migration

**Feature**: 003-design-tokens-migration
**Date**: 2025-12-03
**Updated**: 2025-12-03

## Overview

This document defines the complete token system architecture, including the interface/implementation pattern that abstracts over shadcn/ui variables.

---

## Architecture: Interface/Implementation Pattern

```
┌─────────────────────────────────────────────────────────┐
│                    COMPONENT LAYER                       │
│   Components use: var(--token-text-default)              │
└─────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────┐
│                   TOKEN INTERFACE                        │
│   --token-text-default: var(--foreground);               │
│   (Our abstraction layer - tokens.css)                   │
└─────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────┐
│                 SHADCN IMPLEMENTATION                    │
│   --foreground: oklch(0.145 0 0);                        │
│   (shadcn/ui globals.css)                                │
└─────────────────────────────────────────────────────────┘
```

---

## Complete Token Definitions (CSS)

### Text Tokens

```css
:root {
  /* Text Colors - Interface over shadcn foreground vars */
  --token-text-default: var(--foreground);
  --token-text-secondary: var(--muted-foreground);
  --token-text-tertiary: color-mix(in oklch, var(--muted-foreground) 70%, transparent);
  --token-text-disabled: color-mix(in oklch, var(--foreground) 40%, transparent);
  --token-text-inverse: var(--background);
}

/* Dark mode: shadcn handles --foreground/--muted-foreground automatically */
```

### Surface Tokens

```css
:root {
  /* Surface Colors - Interface over shadcn background vars */
  --token-surface-default: var(--background);
  --token-surface-muted: var(--muted);
  --token-surface-elevated: var(--card);
  --token-surface-overlay: var(--popover);
}

/* Dark mode: shadcn handles automatically */
```

### Highlight Tokens

```css
:root {
  /* Highlight Colors - For selection, focus, emphasis */
  --token-highlight-default: var(--ring);
  --token-highlight-muted: color-mix(in oklch, var(--ring) 30%, transparent);
  --token-highlight-strong: var(--primary);
}
```

### Shadow Tokens

```css
:root {
  /* Shadow Colors - Custom values (not from shadcn) */
  --token-shadow-light: oklch(0 0 0 / 0.05);
  --token-shadow-default: oklch(0 0 0 / 0.10);
  --token-shadow-heavy: oklch(0 0 0 / 0.25);
}

.dark {
  --token-shadow-light: oklch(0 0 0 / 0.20);
  --token-shadow-default: oklch(0 0 0 / 0.40);
  --token-shadow-heavy: oklch(0 0 0 / 0.60);
}
```

### Gradient Tokens

```css
:root {
  /* Gradient Tokens */
  --token-gradient-primary: linear-gradient(135deg,
    oklch(0.65 0.20 250) 0%,
    oklch(0.55 0.25 280) 100%);
  --token-gradient-surface: linear-gradient(180deg,
    var(--token-surface-default) 0%,
    var(--token-surface-muted) 100%);
  --token-gradient-accent: linear-gradient(135deg,
    var(--token-accent-blue-default) 0%,
    var(--token-accent-purple-default) 100%);
}

.dark {
  --token-gradient-primary: linear-gradient(135deg,
    oklch(0.55 0.18 250) 0%,
    oklch(0.45 0.22 280) 100%);
}
```

### Semantic Status Tokens

```css
:root {
  /* Success */
  --token-success-default: oklch(0.72 0.19 142.5);
  --token-success-muted: color-mix(in oklch, var(--token-success-default) 15%, transparent);
  --token-success-foreground: oklch(0.985 0 0);

  /* Warning */
  --token-warning-default: oklch(0.80 0.18 84.4);
  --token-warning-muted: color-mix(in oklch, var(--token-warning-default) 15%, transparent);
  --token-warning-foreground: oklch(0.145 0 0);

  /* Error */
  --token-error-default: oklch(0.63 0.24 25.3);
  --token-error-muted: color-mix(in oklch, var(--token-error-default) 15%, transparent);
  --token-error-foreground: oklch(0.985 0 0);

  /* Info */
  --token-info-default: oklch(0.62 0.21 250.1);
  --token-info-muted: color-mix(in oklch, var(--token-info-default) 15%, transparent);
  --token-info-foreground: oklch(0.985 0 0);
}

.dark {
  --token-success-default: oklch(0.80 0.16 142.5);
  --token-warning-default: oklch(0.85 0.15 84.4);
  --token-error-default: oklch(0.70 0.20 25.3);
  --token-info-default: oklch(0.70 0.18 250.1);
}
```

### Accent Color Tokens

```css
:root {
  /* Blue */
  --token-accent-blue-default: oklch(0.65 0.20 230);
  --token-accent-blue-muted: oklch(0.85 0.08 230);
  --token-accent-blue-strong: oklch(0.50 0.24 230);

  /* Purple */
  --token-accent-purple-default: oklch(0.65 0.20 275);
  --token-accent-purple-muted: oklch(0.85 0.08 275);
  --token-accent-purple-strong: oklch(0.50 0.24 275);

  /* Pink */
  --token-accent-pink-default: oklch(0.65 0.20 320);
  --token-accent-pink-muted: oklch(0.85 0.08 320);
  --token-accent-pink-strong: oklch(0.50 0.24 320);

  /* Orange */
  --token-accent-orange-default: oklch(0.70 0.20 55);
  --token-accent-orange-muted: oklch(0.88 0.08 55);
  --token-accent-orange-strong: oklch(0.55 0.24 55);

  /* Yellow */
  --token-accent-yellow-default: oklch(0.80 0.18 90);
  --token-accent-yellow-muted: oklch(0.92 0.06 90);
  --token-accent-yellow-strong: oklch(0.65 0.22 90);

  /* Green */
  --token-accent-green-default: oklch(0.65 0.20 145);
  --token-accent-green-muted: oklch(0.85 0.08 145);
  --token-accent-green-strong: oklch(0.50 0.24 145);

  /* Cyan */
  --token-accent-cyan-default: oklch(0.70 0.15 195);
  --token-accent-cyan-muted: oklch(0.88 0.06 195);
  --token-accent-cyan-strong: oklch(0.55 0.20 195);

  /* Teal */
  --token-accent-teal-default: oklch(0.65 0.15 175);
  --token-accent-teal-muted: oklch(0.85 0.06 175);
  --token-accent-teal-strong: oklch(0.50 0.20 175);
}

.dark {
  /* Blue */
  --token-accent-blue-default: oklch(0.75 0.18 230);
  --token-accent-blue-muted: oklch(0.30 0.08 230);
  --token-accent-blue-strong: oklch(0.85 0.20 230);

  /* Purple */
  --token-accent-purple-default: oklch(0.75 0.18 275);
  --token-accent-purple-muted: oklch(0.30 0.08 275);
  --token-accent-purple-strong: oklch(0.85 0.20 275);

  /* Pink */
  --token-accent-pink-default: oklch(0.75 0.18 320);
  --token-accent-pink-muted: oklch(0.30 0.08 320);
  --token-accent-pink-strong: oklch(0.85 0.20 320);

  /* Orange */
  --token-accent-orange-default: oklch(0.78 0.18 55);
  --token-accent-orange-muted: oklch(0.32 0.08 55);
  --token-accent-orange-strong: oklch(0.88 0.20 55);

  /* Yellow */
  --token-accent-yellow-default: oklch(0.85 0.16 90);
  --token-accent-yellow-muted: oklch(0.35 0.06 90);
  --token-accent-yellow-strong: oklch(0.92 0.18 90);

  /* Green */
  --token-accent-green-default: oklch(0.75 0.18 145);
  --token-accent-green-muted: oklch(0.30 0.08 145);
  --token-accent-green-strong: oklch(0.85 0.20 145);

  /* Cyan */
  --token-accent-cyan-default: oklch(0.78 0.14 195);
  --token-accent-cyan-muted: oklch(0.32 0.06 195);
  --token-accent-cyan-strong: oklch(0.88 0.16 195);

  /* Teal */
  --token-accent-teal-default: oklch(0.75 0.14 175);
  --token-accent-teal-muted: oklch(0.30 0.06 175);
  --token-accent-teal-strong: oklch(0.85 0.16 175);
}
```

### Intensity Scale Tokens

```css
:root {
  /* Intensity Scale - For heatmaps and density visualization */
  --token-intensity-0: oklch(0.95 0.02 250);
  --token-intensity-1: oklch(0.85 0.08 250);
  --token-intensity-2: oklch(0.75 0.12 250);
  --token-intensity-3: oklch(0.65 0.16 250);
  --token-intensity-4: oklch(0.55 0.20 250);
  --token-intensity-5: oklch(0.45 0.22 250);
}

.dark {
  --token-intensity-0: oklch(0.20 0.02 250);
  --token-intensity-1: oklch(0.35 0.08 250);
  --token-intensity-2: oklch(0.45 0.12 250);
  --token-intensity-3: oklch(0.55 0.16 250);
  --token-intensity-4: oklch(0.65 0.18 250);
  --token-intensity-5: oklch(0.75 0.20 250);
}
```

---

## TypeScript Token Exports

### Structure (lib/tokens/colors.ts)

```typescript
/**
 * Design Token Colors
 *
 * These are LIGHT MODE values for programmatic use.
 * For dark mode support, use CSS variables: var(--token-*)
 */

export const tokens = {
  text: {
    default: 'var(--token-text-default)',
    secondary: 'var(--token-text-secondary)',
    tertiary: 'var(--token-text-tertiary)',
    disabled: 'var(--token-text-disabled)',
    inverse: 'var(--token-text-inverse)',
  },

  surface: {
    default: 'var(--token-surface-default)',
    muted: 'var(--token-surface-muted)',
    elevated: 'var(--token-surface-elevated)',
    overlay: 'var(--token-surface-overlay)',
  },

  highlight: {
    default: 'var(--token-highlight-default)',
    muted: 'var(--token-highlight-muted)',
    strong: 'var(--token-highlight-strong)',
  },

  shadow: {
    light: 'var(--token-shadow-light)',
    default: 'var(--token-shadow-default)',
    heavy: 'var(--token-shadow-heavy)',
  },

  accent: {
    blue: {
      default: 'var(--token-accent-blue-default)',
      muted: 'var(--token-accent-blue-muted)',
      strong: 'var(--token-accent-blue-strong)',
    },
    purple: {
      default: 'var(--token-accent-purple-default)',
      muted: 'var(--token-accent-purple-muted)',
      strong: 'var(--token-accent-purple-strong)',
    },
    pink: {
      default: 'var(--token-accent-pink-default)',
      muted: 'var(--token-accent-pink-muted)',
      strong: 'var(--token-accent-pink-strong)',
    },
    orange: {
      default: 'var(--token-accent-orange-default)',
      muted: 'var(--token-accent-orange-muted)',
      strong: 'var(--token-accent-orange-strong)',
    },
    yellow: {
      default: 'var(--token-accent-yellow-default)',
      muted: 'var(--token-accent-yellow-muted)',
      strong: 'var(--token-accent-yellow-strong)',
    },
    green: {
      default: 'var(--token-accent-green-default)',
      muted: 'var(--token-accent-green-muted)',
      strong: 'var(--token-accent-green-strong)',
    },
    cyan: {
      default: 'var(--token-accent-cyan-default)',
      muted: 'var(--token-accent-cyan-muted)',
      strong: 'var(--token-accent-cyan-strong)',
    },
    teal: {
      default: 'var(--token-accent-teal-default)',
      muted: 'var(--token-accent-teal-muted)',
      strong: 'var(--token-accent-teal-strong)',
    },
  },

  status: {
    success: {
      default: 'var(--token-success-default)',
      muted: 'var(--token-success-muted)',
      foreground: 'var(--token-success-foreground)',
    },
    warning: {
      default: 'var(--token-warning-default)',
      muted: 'var(--token-warning-muted)',
      foreground: 'var(--token-warning-foreground)',
    },
    error: {
      default: 'var(--token-error-default)',
      muted: 'var(--token-error-muted)',
      foreground: 'var(--token-error-foreground)',
    },
    info: {
      default: 'var(--token-info-default)',
      muted: 'var(--token-info-muted)',
      foreground: 'var(--token-info-foreground)',
    },
  },

  intensity: {
    0: 'var(--token-intensity-0)',
    1: 'var(--token-intensity-1)',
    2: 'var(--token-intensity-2)',
    3: 'var(--token-intensity-3)',
    4: 'var(--token-intensity-4)',
    5: 'var(--token-intensity-5)',
  },

  gradient: {
    primary: 'var(--token-gradient-primary)',
    surface: 'var(--token-gradient-surface)',
    accent: 'var(--token-gradient-accent)',
  },
} as const;

// Type exports
export type TextToken = keyof typeof tokens.text;
export type SurfaceToken = keyof typeof tokens.surface;
export type HighlightToken = keyof typeof tokens.highlight;
export type ShadowToken = keyof typeof tokens.shadow;
export type AccentColor = keyof typeof tokens.accent;
export type AccentVariant = 'default' | 'muted' | 'strong';
export type StatusType = keyof typeof tokens.status;
export type StatusVariant = 'default' | 'muted' | 'foreground';
export type IntensityLevel = keyof typeof tokens.intensity;
export type GradientToken = keyof typeof tokens.gradient;
```

---

## Component Migration Mappings

### Chart Components

| Component | Current | Target Token |
|-----------|---------|--------------|
| CategoryTrendsChart | 8 hardcoded HSL | `--token-accent-{color}-default` |
| CategoryChart | `#8884d8` | `--token-accent-blue-default` |
| ProductivityHeatmap | `bg-primary/20..100` | `--token-intensity-0..5` |
| BumpChart | `text-red-600` | `--token-error-default` |
| CompletionChart | `hsl(var(--primary))` | `--token-success-default` |

### Dashboard Components

| Component | Current | Target Token |
|-----------|---------|--------------|
| TaskDetailsSidebar | `dark:bg-gray-900` | `--token-surface-elevated` |
| TaskDetailsSidebar | `text-yellow-600` | `--token-warning-default` |
| InsightCard backgrounds | `bg-amber-50` etc | `--token-{status}-muted` |
| InsightCard text | `text-amber-600` etc | `--token-{status}-default` |
| InsightsList icon | `text-yellow-500` | `--token-warning-default` |

### Page Components

| Component | Current | Target Token |
|-----------|---------|--------------|
| Login/Register error | `text-red-600 bg-red-50` | `--token-error-default`, `--token-error-muted` |
| Analytics error | `text-red-600` | `--token-error-default` |
| Auth gradient | `from-blue-50 to-indigo-100` | `--token-gradient-surface` |

---

## Token Summary

| Category | Count | Naming Pattern |
|----------|-------|----------------|
| Text | 5 | `--token-text-{variant}` |
| Surface | 4 | `--token-surface-{variant}` |
| Highlight | 3 | `--token-highlight-{variant}` |
| Shadow | 3 | `--token-shadow-{variant}` |
| Gradient | 3 | `--token-gradient-{type}` |
| Status | 12 | `--token-{status}-{variant}` |
| Accent | 24 | `--token-accent-{color}-{variant}` |
| Intensity | 6 | `--token-intensity-{0-5}` |
| **Total** | **60** | |
