# Research: Design Tokens Migration

**Feature**: 003-design-tokens-migration
**Date**: 2025-12-03
**Updated**: 2025-12-03
**Status**: Complete

## Overview

This research documents architectural decisions for building a comprehensive design token system that abstracts over shadcn/ui CSS variables.

---

## Research Topic 1: Token Architecture Pattern

**Context**: Should tokens be component-specific or semantic/abstract?

### Decision
Use a **semantic/abstract naming convention** with an **interface/implementation pattern** that abstracts over shadcn/ui variables.

### Rationale
- Component-specific names (`--token-chart-category-1`) create coupling and duplication
- Semantic names (`--token-accent-blue-default`) are reusable across different component types
- Interface pattern allows swapping the underlying implementation without touching components

### Architecture
```
Components → Token Interface (tokens.css) → shadcn Implementation (globals.css)
```

### Alternatives Considered
1. **Component-specific tokens**: Rejected - creates coupling, not reusable
2. **Numbered tokens** (`--token-accent-1`): Rejected - arbitrary, not descriptive
3. **Direct shadcn usage**: Rejected - no abstraction layer for future flexibility

---

## Research Topic 2: Token Naming Convention

**Context**: How should tokens be named for clarity and consistency?

### Decision
Use explicit `-default` suffix on all base tokens and named color variants.

### Pattern
```
--token-{category}-{name}-{variant}

Examples:
--token-text-default          (not --token-text)
--token-accent-blue-default   (not --token-accent-blue)
--token-accent-blue-muted
--token-accent-blue-strong
```

### Rationale
- Explicit is better than implicit - no ambiguity about which variant
- Consistent pattern across all token categories
- Easy to discover via IDE autocomplete

### Alternatives Considered
1. **Implicit default**: Rejected - ambiguous, inconsistent
2. **Numeric variants**: Rejected - doesn't convey semantic meaning

---

## Research Topic 3: Accent Color Naming

**Context**: How should the multi-color palette be named?

### Decision
Use **named colors** (blue, purple, pink, etc.) instead of numbers.

### Color Palette
- Blue (hue ~230)
- Purple (hue ~275)
- Pink (hue ~320)
- Orange (hue ~55)
- Yellow (hue ~90)
- Green (hue ~145)
- Cyan (hue ~195)
- Teal (hue ~175)

### Rationale
- Named colors are intuitive and self-documenting
- Easy to communicate in design discussions ("use the blue accent")
- Each color has 3 variants: default, muted, strong

### Alternatives Considered
1. **Numbered colors** (`--token-accent-1`): Rejected - arbitrary, not intuitive
2. **Single color with opacity**: Rejected - doesn't provide enough variety

---

## Research Topic 4: Interface over shadcn Variables

**Context**: How should tokens reference shadcn/ui CSS variables?

### Decision
Create a token interface layer that references shadcn variables where applicable, with custom values for new tokens.

### Implementation
```css
/* Interface over shadcn */
--token-text-default: var(--foreground);
--token-surface-default: var(--background);

/* Custom values (not from shadcn) */
--token-accent-blue-default: oklch(0.65 0.20 230);
--token-shadow-default: oklch(0 0 0 / 0.10);
```

### Rationale
- Leverages shadcn's dark mode handling for base tokens
- Allows custom values where shadcn doesn't provide what we need
- Single abstraction layer for all styling decisions
- Easy to swap underlying implementation if needed

### Alternatives Considered
1. **Duplicate shadcn values**: Rejected - maintenance burden, out of sync risk
2. **Only custom values**: Rejected - reinventing what shadcn already provides
3. **Direct shadcn usage**: Rejected - no abstraction, harder to change later

---

## Research Topic 5: Intensity Scale Design

**Context**: How should the heatmap intensity scale be structured?

### Decision
Use **numeric scale (0-5)** with oklch values varying in lightness.

### Scale
```
0: Near-white (0.95 lightness) → None/Minimal
1: Light (0.85 lightness)
2: Medium-light (0.75 lightness)
3: Medium (0.65 lightness)
4: Dark (0.55 lightness)
5: Darkest (0.45 lightness) → Maximum
```

### Rationale
- Numbers make sense for a scale (0 = none, 5 = maximum)
- Lightness-based scale works well for both light and dark modes
- 6 levels provide sufficient granularity for productivity heatmaps

### Dark Mode Adjustment
In dark mode, the scale is inverted (0 is darkest, 5 is lightest) to maintain visual meaning.

### Alternatives Considered
1. **Qualitative names** (weak, medium, strong): Could work but numeric is more precise for scales
2. **Opacity-based**: Current approach - doesn't adapt well to dark mode

---

## Research Topic 6: Token Categories

**Context**: What categories of tokens should be defined?

### Decision
Define 8 token categories covering all styling needs.

### Categories

| Category | Purpose | Count |
|----------|---------|-------|
| Text | Text color hierarchy | 5 |
| Surface | Background colors | 4 |
| Highlight | Selection, focus states | 3 |
| Shadow | Elevation, depth | 3 |
| Gradient | Background effects | 3 |
| Status | Semantic feedback | 12 |
| Accent | Multi-color palette | 24 |
| Intensity | Density visualization | 6 |

**Total: 60 tokens**

### Rationale
- Covers all current component needs
- Aligned with common design system patterns
- Extensible for future needs

---

## Research Topic 7: Browser Compatibility

**Context**: What CSS features are required and what is their support level?

### Features Used
1. **CSS Custom Properties**: 97%+ support ✅
2. **oklch color format**: 93%+ support ✅
3. **color-mix()**: 95%+ support ✅
4. **linear-gradient()**: 99%+ support ✅

### Decision
All features have acceptable browser support. No fallbacks needed.

### Rationale
- TaskFlow targets modern browsers
- All features have 93%+ global support
- oklch provides perceptual uniformity benefits that outweigh edge case compatibility

---

## Summary

| Topic | Decision | Confidence |
|-------|----------|------------|
| Architecture | Interface/Implementation pattern | High |
| Naming | Explicit `-default` suffix | High |
| Accent Colors | Named colors (blue, purple, etc.) | High |
| shadcn Integration | Reference via var() | High |
| Intensity Scale | Numeric 0-5 | High |
| Categories | 8 categories, 60 tokens | High |
| Browser Compat | No fallbacks needed | High |

All research complete. No NEEDS CLARIFICATION items remain.
