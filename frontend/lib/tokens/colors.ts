/**
 * Color Tokens
 *
 * Semantic color values for programmatic use in TypeScript/JavaScript.
 * Use these when CSS variables aren't supported (e.g., Recharts, canvas).
 *
 * For CSS usage, prefer var(--token-success) etc. from tokens.css
 *
 * @see specs/002-design-tokens/data-model.md for token definitions
 */

/**
 * Semantic status colors for UI feedback states.
 * Values are in oklch format for consistent color representation.
 */
export const colors = {
  /** Success state - positive feedback, completed actions */
  success: 'oklch(0.72 0.19 142.5)',
  /** Text color for use on success backgrounds */
  successForeground: 'oklch(0.985 0 0)',

  /** Warning state - caution, attention needed */
  warning: 'oklch(0.80 0.18 84.4)',
  /** Text color for use on warning backgrounds */
  warningForeground: 'oklch(0.145 0 0)',

  /** Error state - destructive actions, failures */
  error: 'oklch(0.63 0.24 25.3)',
  /** Text color for use on error backgrounds */
  errorForeground: 'oklch(0.985 0 0)',

  /** Info state - informational, neutral feedback */
  info: 'oklch(0.62 0.21 250.1)',
  /** Text color for use on info backgrounds */
  infoForeground: 'oklch(0.985 0 0)',

  /**
   * Chart-specific colors for Recharts and similar libraries.
   * These match the priority levels used in TaskFlow.
   */
  chart: {
    /** Critical priority (90-100) - red/destructive */
    critical: 'oklch(0.63 0.24 25.3)',
    /** High priority (75-89) - orange/warning */
    high: 'oklch(0.77 0.19 70.1)',
    /** Medium priority (50-74) - blue/info */
    medium: 'oklch(0.62 0.21 250.1)',
    /** Low priority (0-49) - gray/muted */
    low: 'oklch(0.70 0 0)',
  },
} as const;

/** Type for semantic color token names */
export type ColorToken = keyof Omit<typeof colors, 'chart'>;

/** Type for chart color token names */
export type ChartColorToken = keyof typeof colors.chart;
