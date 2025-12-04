/**
 * Design System Tokens
 *
 * Centralized exports for all design token modules.
 * Import from '@/lib/tokens' for easy access to all tokens.
 *
 * @example
 * // New unified API (recommended)
 * import { tokens, getAccentColor, ACCENT_COLORS } from '@/lib/tokens';
 * <Bar fill={tokens.accent.blue.default} />
 * <div style={{ color: tokens.text.secondary }} />
 *
 * // Legacy API (still supported for backwards compatibility)
 * import { colors, spacing, typography } from '@/lib/tokens';
 * <Bar fill={colors.chart.critical} />
 *
 * @see docs/design-system.md for usage guidelines
 * @see specs/003-design-tokens-migration/data-model.md for token definitions
 */

// ============================================================================
// New Unified Token API (Feature 003)
// ============================================================================

export {
  tokens,
  getAccentColor,
  getStatusColor,
  getIntensityColor,
  ACCENT_COLORS,
} from './tokens';

export type {
  TextToken,
  SurfaceToken,
  HighlightToken,
  ShadowToken,
  AccentColor,
  AccentVariant,
  StatusType,
  StatusVariant,
  IntensityLevel,
  GradientToken,
  ChartToken,
} from './tokens';

// ============================================================================
// Legacy API (Feature 002 - maintained for backwards compatibility)
// ============================================================================

// Color tokens
export { colors } from './colors';
export type { ColorToken, ChartColorToken } from './colors';

// Spacing tokens
export { spacing } from './spacing';
export type { SpacingToken } from './spacing';

// Typography tokens
export { fontSize, lineHeight, fontWeight, typography } from './typography';
export type { FontSizeToken, LineHeightToken, FontWeightToken } from './typography';
