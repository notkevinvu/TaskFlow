/**
 * Spacing Tokens
 *
 * Spacing scale values for programmatic use in TypeScript/JavaScript.
 * Matches Tailwind's spacing scale for consistency.
 *
 * For CSS usage, prefer var(--token-space-4) etc. from tokens.css
 *
 * @see specs/002-design-tokens/data-model.md for token definitions
 */

/**
 * Spacing scale matching Tailwind's default spacing values.
 * Use these for consistent spacing in programmatic contexts.
 */
export const spacing = {
  /** 0.125rem = 2px */
  space0_5: '0.125rem',
  /** 0.25rem = 4px */
  space1: '0.25rem',
  /** 0.5rem = 8px */
  space2: '0.5rem',
  /** 0.75rem = 12px */
  space3: '0.75rem',
  /** 1rem = 16px */
  space4: '1rem',
  /** 1.25rem = 20px */
  space5: '1.25rem',
  /** 1.5rem = 24px */
  space6: '1.5rem',
  /** 2rem = 32px */
  space8: '2rem',
  /** 2.5rem = 40px */
  space10: '2.5rem',
  /** 3rem = 48px */
  space12: '3rem',
  /** 4rem = 64px */
  space16: '4rem',
} as const;

/** Type for spacing token names */
export type SpacingToken = keyof typeof spacing;
