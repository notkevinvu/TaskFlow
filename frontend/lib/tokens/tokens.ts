/**
 * Design System Tokens - Unified Export
 *
 * Complete token system for programmatic access in TypeScript/JavaScript.
 * All values reference CSS custom properties for automatic dark mode support.
 *
 * Usage:
 *   import { tokens } from '@/lib/tokens';
 *   <div style={{ color: tokens.text.default }} />
 *
 * For Recharts and similar libraries that need string values:
 *   <Bar fill={tokens.accent.blue.default} />
 *
 * @see docs/design-system.md for usage guidelines
 * @see specs/003-design-tokens-migration/data-model.md for token definitions
 */

/**
 * Complete design token system with all 60 tokens organized by category.
 */
export const tokens = {
  /**
   * Text colors for typography hierarchy.
   * Interface over shadcn's --foreground variables.
   */
  text: {
    /** Primary text color - use for headings and body text */
    default: 'var(--token-text-default)',
    /** Secondary text color - use for supporting text */
    secondary: 'var(--token-text-secondary)',
    /** Tertiary text color - use for subtle labels */
    tertiary: 'var(--token-text-tertiary)',
    /** Disabled text color - use for inactive states */
    disabled: 'var(--token-text-disabled)',
    /** Inverse text color - use on dark backgrounds */
    inverse: 'var(--token-text-inverse)',
  },

  /**
   * Surface colors for backgrounds.
   * Interface over shadcn's --background variables.
   */
  surface: {
    /** Default background - page background */
    default: 'var(--token-surface-default)',
    /** Muted background - subtle sections */
    muted: 'var(--token-surface-muted)',
    /** Elevated background - cards, raised elements */
    elevated: 'var(--token-surface-elevated)',
    /** Overlay background - modals, popovers */
    overlay: 'var(--token-surface-overlay)',
  },

  /**
   * Highlight colors for emphasis and interaction states.
   * Interface over shadcn's --ring and --primary variables.
   */
  highlight: {
    /** Default highlight - focus rings, selection */
    default: 'var(--token-highlight-default)',
    /** Muted highlight - subtle emphasis */
    muted: 'var(--token-highlight-muted)',
    /** Strong highlight - primary actions, CTAs */
    strong: 'var(--token-highlight-strong)',
  },

  /**
   * Shadow colors for elevation and depth.
   * Custom values (not from shadcn).
   */
  shadow: {
    /** Light shadow - subtle elevation */
    light: 'var(--token-shadow-light)',
    /** Default shadow - standard elevation */
    default: 'var(--token-shadow-default)',
    /** Heavy shadow - pronounced elevation */
    heavy: 'var(--token-shadow-heavy)',
  },

  /**
   * Named accent colors for data visualization and categorization.
   * Each color has default, muted, and strong variants.
   */
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

  /**
   * Semantic status colors for feedback states.
   * Each status has default, muted, and foreground variants.
   */
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

  /**
   * Intensity scale for heatmaps and density visualization.
   * 0 = minimum intensity, 5 = maximum intensity.
   */
  intensity: {
    0: 'var(--token-intensity-0)',
    1: 'var(--token-intensity-1)',
    2: 'var(--token-intensity-2)',
    3: 'var(--token-intensity-3)',
    4: 'var(--token-intensity-4)',
    5: 'var(--token-intensity-5)',
  },

  /**
   * Gradient tokens for background effects.
   */
  gradient: {
    /** Primary gradient - hero sections, branding */
    primary: 'var(--token-gradient-primary)',
    /** Surface gradient - subtle background transitions */
    surface: 'var(--token-gradient-surface)',
    /** Accent gradient - decorative elements */
    accent: 'var(--token-gradient-accent)',
  },

  /**
   * Chart-specific colors for priority visualization.
   * @deprecated Use tokens.status or tokens.accent for new components.
   */
  chart: {
    critical: 'var(--token-chart-critical)',
    high: 'var(--token-chart-high)',
    medium: 'var(--token-chart-medium)',
    low: 'var(--token-chart-low)',
  },
} as const;

// ============================================================================
// Type Exports
// ============================================================================

/** Text token variant names */
export type TextToken = keyof typeof tokens.text;

/** Surface token variant names */
export type SurfaceToken = keyof typeof tokens.surface;

/** Highlight token variant names */
export type HighlightToken = keyof typeof tokens.highlight;

/** Shadow token variant names */
export type ShadowToken = keyof typeof tokens.shadow;

/** Accent color names */
export type AccentColor = keyof typeof tokens.accent;

/** Accent color variant names */
export type AccentVariant = 'default' | 'muted' | 'strong';

/** Status type names */
export type StatusType = keyof typeof tokens.status;

/** Status variant names */
export type StatusVariant = 'default' | 'muted' | 'foreground';

/** Intensity level (0-5) */
export type IntensityLevel = keyof typeof tokens.intensity;

/** Gradient token names */
export type GradientToken = keyof typeof tokens.gradient;

/** Chart color names (deprecated) */
export type ChartToken = keyof typeof tokens.chart;

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Get an accent color by name and variant.
 * Useful for dynamic color selection in charts.
 *
 * @example
 * const color = getAccentColor('blue', 'default');
 * // Returns: 'var(--token-accent-blue-default)'
 */
export function getAccentColor(
  color: AccentColor,
  variant: AccentVariant = 'default'
): string {
  return tokens.accent[color][variant];
}

/**
 * Get a status color by type and variant.
 *
 * @example
 * const color = getStatusColor('success', 'muted');
 * // Returns: 'var(--token-success-muted)'
 */
export function getStatusColor(
  status: StatusType,
  variant: StatusVariant = 'default'
): string {
  return tokens.status[status][variant];
}

/**
 * Get an intensity color by level (0-5).
 *
 * @example
 * const color = getIntensityColor(3);
 * // Returns: 'var(--token-intensity-3)'
 */
export function getIntensityColor(level: IntensityLevel): string {
  return tokens.intensity[level];
}

/**
 * Array of all accent color names for iteration.
 * Useful for chart series color assignment.
 *
 * @example
 * ACCENT_COLORS.map((color, i) => (
 *   <Line key={i} stroke={tokens.accent[color].default} />
 * ))
 */
export const ACCENT_COLORS: AccentColor[] = [
  'blue',
  'purple',
  'pink',
  'orange',
  'yellow',
  'green',
  'cyan',
  'teal',
];
