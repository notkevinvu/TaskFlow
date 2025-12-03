/**
 * Typography Tokens
 *
 * Typography values for programmatic use in TypeScript/JavaScript.
 * Matches Tailwind's typography scale for consistency.
 *
 * For CSS usage, prefer var(--token-font-size-base) etc. from tokens.css
 *
 * @see specs/002-design-tokens/data-model.md for token definitions
 */

/**
 * Font size scale matching Tailwind's text-* utilities.
 */
export const fontSize = {
  /** 0.75rem = 12px */
  xs: '0.75rem',
  /** 0.875rem = 14px */
  sm: '0.875rem',
  /** 1rem = 16px */
  base: '1rem',
  /** 1.125rem = 18px */
  lg: '1.125rem',
  /** 1.25rem = 20px */
  xl: '1.25rem',
  /** 1.5rem = 24px */
  '2xl': '1.5rem',
  /** 1.875rem = 30px */
  '3xl': '1.875rem',
  /** 2.25rem = 36px */
  '4xl': '2.25rem',
} as const;

/**
 * Line height scale matching Tailwind's leading-* utilities.
 */
export const lineHeight = {
  /** 1.25 - compact text */
  tight: '1.25',
  /** 1.5 - default body text */
  normal: '1.5',
  /** 1.75 - spacious text */
  relaxed: '1.75',
} as const;

/**
 * Font weight scale matching Tailwind's font-* utilities.
 */
export const fontWeight = {
  /** 400 - regular text */
  normal: '400',
  /** 500 - slightly emphasized */
  medium: '500',
  /** 600 - headings, labels */
  semibold: '600',
  /** 700 - strong emphasis */
  bold: '700',
} as const;

/**
 * Combined typography tokens for convenience.
 */
export const typography = {
  fontSize,
  lineHeight,
  fontWeight,
} as const;

/** Type for font size token names */
export type FontSizeToken = keyof typeof fontSize;

/** Type for line height token names */
export type LineHeightToken = keyof typeof lineHeight;

/** Type for font weight token names */
export type FontWeightToken = keyof typeof fontWeight;
