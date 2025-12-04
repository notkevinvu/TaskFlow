# Feature Specification: Design Tokens Migration

**Feature Branch**: `003-design-tokens-migration`
**Created**: 2025-12-03
**Updated**: 2025-12-03
**Status**: Draft
**Input**: User description: "Migrate all existing UI components to use the design token system established in Feature 002"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Token Abstraction Layer (Priority: P1)

Developers use a consistent token interface (`--token-*`) that abstracts over the underlying implementation (shadcn/ui variables). This allows the design system to evolve independently of the component library.

**Why this priority**: This is the foundation. All other migrations depend on having the token interface defined first.

**Independent Test**: Import any token and verify it resolves correctly. Change the underlying shadcn value and confirm the token updates automatically.

**Acceptance Scenarios**:

1. **Given** a component using `var(--token-text-default)`, **When** shadcn's `--foreground` changes, **Then** the component automatically reflects the new color
2. **Given** a developer needs a text color, **When** they check the token system, **Then** they find clearly named options (`-default`, `-secondary`, `-tertiary`)
3. **Given** dark mode is enabled, **When** tokens referencing shadcn vars are used, **Then** they automatically adapt via shadcn's dark mode handling

---

### User Story 2 - Chart Components Use Accent Tokens (Priority: P1)

Chart components (CategoryTrendsChart, CategoryChart, etc.) use named accent color tokens (`--token-accent-blue-default`, `--token-accent-purple-default`, etc.) instead of hardcoded color values.

**Why this priority**: Charts have the most hardcoded colors and are highly visible. Named accent tokens are reusable beyond charts.

**Independent Test**: Toggle dark mode while viewing the Analytics page. All chart colors should adapt appropriately.

**Acceptance Scenarios**:

1. **Given** CategoryTrendsChart with 8 series, **When** rendered, **Then** each series uses a named accent token (`blue`, `purple`, `pink`, `orange`, `yellow`, `green`, `cyan`, `teal`)
2. **Given** a chart using accent tokens, **When** dark mode is toggled, **Then** colors adjust for visibility on dark backgrounds
3. **Given** the ProductivityHeatmap, **When** displaying intensity levels, **Then** it uses the numeric intensity scale (`--token-intensity-0` through `-5`)

---

### User Story 3 - Dashboard Components Use Semantic Tokens (Priority: P2)

Dashboard components (TaskDetailsSidebar, InsightCard, InsightsList) use semantic tokens for text, surfaces, and status colors.

**Why this priority**: Establishes consistent visual hierarchy using the new text/surface token structure.

**Independent Test**: View dashboard with various task states. Verify text hierarchy (primary, secondary, tertiary) is visually distinct and status colors are consistent.

**Acceptance Scenarios**:

1. **Given** InsightCard displaying different insight types, **When** rendered, **Then** each type maps to appropriate semantic status tokens
2. **Given** TaskDetailsSidebar, **When** showing a warning state, **Then** it uses `--token-warning-default` consistently
3. **Given** secondary text in any component, **When** rendered, **Then** it uses `--token-text-secondary`

---

### User Story 4 - Page Components Use Token-Based Styling (Priority: P3)

Authentication pages (Login, Register) and the Analytics page use design tokens for error states, surfaces, and text hierarchy.

**Why this priority**: Completes the migration with page-level consistency.

**Independent Test**: Trigger an error on the login page. Verify styling uses token variables.

**Acceptance Scenarios**:

1. **Given** an error on the login page, **When** displayed, **Then** it uses `--token-error-default` and appropriate surface tokens
2. **Given** the auth page background, **When** rendered, **Then** it uses `--token-gradient-surface` or `--token-surface-*` tokens

---

### Edge Cases

- What happens when a component references a token that doesn't exist? TypeScript compilation should fail, preventing runtime errors.
- How does the system handle shadcn variable changes? Tokens reference shadcn vars, so changes propagate automatically.
- What if we need a color not in the accent palette? Extend the palette or use semantic tokens if appropriate.

## Requirements *(mandatory)*

### Functional Requirements

**Token Interface Layer:**
- **FR-001**: System MUST define a token interface that abstracts over shadcn/ui CSS variables
- **FR-002**: All base tokens MUST use the `-default` suffix pattern explicitly (e.g., `--token-text-default`, not `--token-text`)

**Text Tokens:**
- **FR-003**: System MUST provide text color tokens: `-default`, `-secondary`, `-tertiary`, `-disabled`, `-inverse`

**Surface Tokens:**
- **FR-004**: System MUST provide surface tokens: `-default`, `-muted`, `-elevated`, `-overlay`

**Accent Tokens:**
- **FR-005**: System MUST provide named accent colors with variants: `blue`, `purple`, `pink`, `orange`, `yellow`, `green`, `cyan`, `teal`
- **FR-006**: Each accent color MUST have `-default`, `-muted`, and `-strong` variants

**Utility Tokens:**
- **FR-007**: System MUST provide highlight tokens: `-default`, `-muted`, `-strong`
- **FR-008**: System MUST provide shadow tokens: `-light`, `-default`, `-heavy`
- **FR-009**: System MUST provide gradient tokens: `-primary`, `-surface`, `-accent`
- **FR-010**: System MUST provide numeric intensity scale: `--token-intensity-0` through `--token-intensity-5`

**Semantic Tokens:**
- **FR-011**: Existing semantic tokens (success, warning, error, info) MUST follow `-default`/`-muted` pattern

**Migration:**
- **FR-012**: All chart components MUST be migrated to use accent and intensity tokens
- **FR-013**: All dashboard components MUST be migrated to use text, surface, and semantic tokens
- **FR-014**: All page error states MUST be migrated to use error tokens
- **FR-015**: TypeScript exports MUST be updated to include all new tokens

**Documentation:**
- **FR-016**: Documentation MUST be updated to reflect the complete token system

### Key Entities

- **Token Interface (tokens.css)**: CSS custom properties defining the public API, referencing shadcn vars where applicable
- **Token Exports (lib/tokens/)**: TypeScript constants for programmatic access
- **shadcn Variables**: The underlying implementation (--foreground, --background, etc.)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of components use token interface with zero direct shadcn variable references in application code
- **SC-002**: All tokens follow the explicit `-default` naming convention
- **SC-003**: All migrated components maintain visual parity in light mode
- **SC-004**: Dark mode toggle results in appropriate color changes for all components
- **SC-005**: Zero TypeScript compilation errors related to token imports
- **SC-006**: Token documentation is complete and developers can identify correct token within 30 seconds

## Assumptions

- shadcn/ui CSS variables remain stable and follow current naming conventions
- The token interface pattern provides sufficient abstraction for future flexibility
- 8 accent colors are sufficient for current needs (expandable later)
- Intensity scale of 0-5 provides adequate granularity for heatmaps

## Out of Scope

- Migrating shadcn/ui component internals (they use their own variables)
- Adding new color themes beyond light/dark mode
- Performance optimization of CSS variable resolution
- Creating a visual token documentation site (using markdown docs)
