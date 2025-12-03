/**
 * Design System Tokens
 *
 * Centralized exports for all design token modules.
 * Import from '@/lib/tokens' for easy access to all tokens.
 *
 * @example
 * import { colors, spacing, typography } from '@/lib/tokens';
 *
 * // Use in Recharts
 * <Bar fill={colors.chart.critical} />
 *
 * // Use for programmatic styling
 * const style = { padding: spacing.space4 };
 *
 * @see docs/design-system.md for usage guidelines
 * @see specs/002-design-tokens/quickstart.md for examples
 */

// Color tokens
export { colors } from './colors';
export type { ColorToken, ChartColorToken } from './colors';

// Spacing tokens
export { spacing } from './spacing';
export type { SpacingToken } from './spacing';

// Typography tokens
export { fontSize, lineHeight, fontWeight, typography } from './typography';
export type { FontSizeToken, LineHeightToken, FontWeightToken } from './typography';
