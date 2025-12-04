# Tasks: Design Tokens Migration

**Input**: Design documents from `/specs/003-design-tokens-migration/`
**Prerequisites**: plan.md (required), spec.md (required), data-model.md, research.md

**Tests**: Not required - this feature uses visual verification and TypeScript compilation checks per spec.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

- **Frontend**: `frontend/app/`, `frontend/lib/`, `frontend/components/`
- **Docs**: `docs/`

---

## Phase 1: Setup

**Purpose**: Verify prerequisites and prepare for token system expansion

- [ ] T001 Verify Feature 002 tokens exist at `frontend/app/tokens.css`
- [ ] T002 Verify existing token exports at `frontend/lib/tokens/index.ts`

---

## Phase 2: User Story 1 - Token Abstraction Layer (Priority: P1) 🎯 MVP

**Goal**: Define complete token interface that abstracts over shadcn/ui CSS variables (60 tokens)

**Independent Test**: Check browser DevTools → Elements → Computed styles for `--token-*` variables on `:root`. Verify tokens resolve correctly.

### Implementation for User Story 1

#### Text & Surface Tokens
- [ ] T003 [P] [US1] Add text tokens (default, secondary, tertiary, disabled, inverse) to `frontend/app/tokens.css`
- [ ] T004 [P] [US1] Add surface tokens (default, muted, elevated, overlay) to `frontend/app/tokens.css`

#### Utility Tokens
- [ ] T005 [P] [US1] Add highlight tokens (default, muted, strong) to `frontend/app/tokens.css`
- [ ] T006 [P] [US1] Add shadow tokens (light, default, heavy) with dark mode to `frontend/app/tokens.css`
- [ ] T007 [P] [US1] Add gradient tokens (primary, surface, accent) with dark mode to `frontend/app/tokens.css`

#### Status Tokens (update existing)
- [ ] T008 [US1] Update existing status tokens to follow `-default`/`-muted`/`-foreground` pattern in `frontend/app/tokens.css`

#### Accent Color Tokens
- [ ] T009 [P] [US1] Add blue accent tokens (default, muted, strong) with dark mode to `frontend/app/tokens.css`
- [ ] T010 [P] [US1] Add purple accent tokens (default, muted, strong) with dark mode to `frontend/app/tokens.css`
- [ ] T011 [P] [US1] Add pink accent tokens (default, muted, strong) with dark mode to `frontend/app/tokens.css`
- [ ] T012 [P] [US1] Add orange accent tokens (default, muted, strong) with dark mode to `frontend/app/tokens.css`
- [ ] T013 [P] [US1] Add yellow accent tokens (default, muted, strong) with dark mode to `frontend/app/tokens.css`
- [ ] T014 [P] [US1] Add green accent tokens (default, muted, strong) with dark mode to `frontend/app/tokens.css`
- [ ] T015 [P] [US1] Add cyan accent tokens (default, muted, strong) with dark mode to `frontend/app/tokens.css`
- [ ] T016 [P] [US1] Add teal accent tokens (default, muted, strong) with dark mode to `frontend/app/tokens.css`

#### Intensity Scale
- [ ] T017 [US1] Add intensity scale tokens (0-5) with dark mode to `frontend/app/tokens.css`

#### TypeScript Exports
- [ ] T018 [US1] Create unified token exports in `frontend/lib/tokens/tokens.ts` with all token categories
- [ ] T019 [US1] Update `frontend/lib/tokens/index.ts` to re-export new token structure
- [ ] T020 [US1] Verify TypeScript compilation passes (`npm run build`)

**Checkpoint**: Token interface complete. All 60 tokens visible in DevTools. TypeScript exports available.

---

## Phase 3: User Story 2 - Chart Components Use Accent Tokens (Priority: P1)

**Goal**: Migrate all 5 chart components to use design tokens

**Independent Test**: Toggle dark mode while viewing Analytics page. All chart colors should adapt appropriately.

### Implementation for User Story 2

- [ ] T021 [P] [US2] Migrate `frontend/components/charts/CategoryTrendsChart.tsx` - Replace 8 hardcoded HSL with accent tokens
- [ ] T022 [P] [US2] Migrate `frontend/components/charts/CategoryChart.tsx` - Replace `#8884d8` with accent token
- [ ] T023 [P] [US2] Migrate `frontend/components/charts/ProductivityHeatmap.tsx` - Replace opacity classes with intensity tokens
- [ ] T024 [P] [US2] Migrate `frontend/components/charts/BumpChart.tsx` - Replace `text-red-600` with error token
- [ ] T025 [P] [US2] Migrate `frontend/components/charts/CompletionChart.tsx` - Update to use status tokens
- [ ] T026 [US2] Verify all chart components render correctly in light mode
- [ ] T027 [US2] Verify all chart components adapt in dark mode

**Checkpoint**: All charts use design tokens. Dark mode works for all charts.

---

## Phase 4: User Story 3 - Dashboard Components Use Semantic Tokens (Priority: P2)

**Goal**: Migrate dashboard components to use text, surface, and status tokens

**Independent Test**: View dashboard with various task states. Verify status indicators use consistent colors.

### Implementation for User Story 3

- [ ] T028 [P] [US3] Migrate `frontend/components/insights/InsightCard.tsx` - Replace 6 insight type colors with status tokens
- [ ] T029 [P] [US3] Migrate `frontend/components/insights/InsightsList.tsx` - Replace icon color with warning token
- [ ] T030 [US3] Migrate `frontend/components/TaskDetailsSidebar.tsx` - Replace backgrounds and warning colors with surface/status tokens
- [ ] T031 [US3] Verify dashboard components use consistent semantic colors
- [ ] T032 [US3] Verify dark mode works for all dashboard components

**Checkpoint**: Dashboard components use semantic tokens. Status indicators are consistent.

---

## Phase 5: User Story 4 - Page Components Use Token-Based Styling (Priority: P3)

**Goal**: Migrate page-level error states and styling to tokens

**Independent Test**: Trigger an error on login page. Verify error styling uses token colors.

### Implementation for User Story 4

- [ ] T033 [P] [US4] Migrate `frontend/app/(auth)/login/page.tsx` - Replace error colors and gradient
- [ ] T034 [P] [US4] Migrate `frontend/app/(auth)/register/page.tsx` - Replace error colors and gradient
- [ ] T035 [US4] Migrate `frontend/app/(dashboard)/analytics/page.tsx` - Replace error state colors
- [ ] T036 [US4] Verify error messages display correctly on all pages
- [ ] T037 [US4] Verify dark mode works for all page components

**Checkpoint**: All pages use token-based styling. Error states are consistent.

---

## Phase 6: Polish & Documentation

**Purpose**: Update documentation and final verification

- [ ] T038 [P] Update `docs/design-system.md` with complete token reference (all 60 tokens)
- [ ] T039 [P] Update `docs/design-system.md` with usage examples for new token categories
- [ ] T040 Verify frontend builds without TypeScript errors (`npm run build`)
- [ ] T041 Final dark mode verification across entire application
- [ ] T042 Remove deprecated token code from Feature 002 if any

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **US1 - Token Layer (Phase 2)**: Depends on Setup - BLOCKS all other user stories
- **US2 - Charts (Phase 3)**: Depends on US1 (needs tokens to exist)
- **US3 - Dashboard (Phase 4)**: Depends on US1 (needs tokens to exist)
- **US4 - Pages (Phase 5)**: Depends on US1 (needs tokens to exist)
- **Polish (Phase 6)**: Depends on all user stories complete

### User Story Dependencies

```
Phase 1: Setup
    ↓
Phase 2: US1 (Token Layer) ← BLOCKING - must complete first
    ↓
┌───┴───┬───────┐
↓       ↓       ↓
US2     US3     US4  ← Can run in parallel (different files)
(Charts) (Dashboard) (Pages)
    ↓
 Polish
```

### Parallel Opportunities

**Within Phase 2 (US1 - Token Layer):**
```
# These CSS token tasks can run in parallel (same file but different sections):
T003-T007: Text, surface, highlight, shadow, gradient tokens
T009-T016: All 8 accent color tokens
```

**Phases 3, 4, 5 can run in parallel:**
```
# Different component files, no dependencies between them:
Phase 3 (US2): All chart migrations (T021-T025)
Phase 4 (US3): All dashboard migrations (T028-T030)
Phase 5 (US4): All page migrations (T033-T035)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: US1 - Token Abstraction Layer
3. **STOP and VALIDATE**: Verify 60 tokens in DevTools
4. Token system is usable for any future component work

### Full Implementation

1. Complete Setup + US1 (Token Layer) → Tokens ready
2. Complete US2 (Charts) → Charts use tokens, dark mode works
3. Complete US3 (Dashboard) → Dashboard uses tokens
4. Complete US4 (Pages) → Pages use tokens
5. Complete Polish → Documentation complete, verified

### Parallel Execution (2-3 Developers)

```
Developer A: US2 (Charts) - 5 component migrations
Developer B: US3 (Dashboard) - 3 component migrations
Developer C: US4 (Pages) - 3 page migrations
    ↓ (sync point)
All: Polish & Verification
```

---

## Summary

| Metric | Count |
|--------|-------|
| Total Tasks | 42 |
| US1 Tasks (Token Layer) | 18 |
| US2 Tasks (Charts) | 7 |
| US3 Tasks (Dashboard) | 5 |
| US4 Tasks (Pages) | 5 |
| Setup Tasks | 2 |
| Polish Tasks | 5 |
| Parallel Opportunities | 25 tasks marked [P] |

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- No unit tests needed - visual verification and TypeScript compilation per spec
- All token values are defined in data-model.md
- Commit after each phase completion
- Stop at any checkpoint to validate independently
