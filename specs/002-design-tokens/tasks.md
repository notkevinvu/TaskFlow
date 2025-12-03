# Tasks: Design System Tokens

**Input**: Design documents from `/specs/002-design-tokens/`
**Prerequisites**: plan.md (required), spec.md (required), data-model.md, research.md, quickstart.md

**Tests**: Not required - this feature uses visual verification and TypeScript compilation checks.

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

**Purpose**: Create directory structure for token files

- [x] T001 Create tokens directory at `frontend/lib/tokens/`

---

## Phase 2: User Story 1 - CSS Tokens for Styling (Priority: P1) 🎯 MVP

**Goal**: Developers can use semantic CSS custom properties instead of raw color values or Tailwind classes

**Independent Test**: View browser DevTools → Elements → Computed styles and verify `--token-*` variables are visible on `:root`. Toggle dark mode and verify values change.

### Implementation for User Story 1

- [x] T002 [US1] Create `frontend/app/tokens.css` with color tokens (success, warning, error, info) for `:root` and `.dark` selectors using values from data-model.md
- [x] T003 [US1] Add spacing tokens to `frontend/app/tokens.css` (--token-space-0-5 through --token-space-16)
- [x] T004 [US1] Add typography tokens to `frontend/app/tokens.css` (font-size, line-height, font-weight)
- [x] T005 [US1] Update `frontend/app/globals.css` to import tokens.css at the top before other imports

**Checkpoint**: CSS tokens visible in browser DevTools on `:root` and `.dark` selector

---

## Phase 3: User Story 2 - TypeScript Token Access (Priority: P2)

**Goal**: Developers can import token constants in TypeScript for programmatic use (charts, conditional styling)

**Independent Test**: Import `{ colors }` from `@/lib/tokens` in any TypeScript file and verify IDE autocomplete shows all token options.

### Implementation for User Story 2

- [x] T006 [P] [US2] Create `frontend/lib/tokens/colors.ts` with color object including semantic colors and chart.* subset, using values from data-model.md
- [x] T007 [P] [US2] Create `frontend/lib/tokens/spacing.ts` with spacing scale (space0_5 through space16)
- [x] T008 [P] [US2] Create `frontend/lib/tokens/typography.ts` with fontSize, lineHeight, and fontWeight objects
- [x] T009 [US2] Create `frontend/lib/tokens/index.ts` to re-export all token modules

**Checkpoint**: TypeScript imports resolve without errors; IDE autocomplete shows token options

---

## Phase 4: User Story 3 - Component Migration (Priority: P3)

**Goal**: PriorityChart demonstrates the token usage pattern as proof-of-concept

**Independent Test**: Compare PriorityChart visual appearance before and after migration - should be identical

### Implementation for User Story 3

- [x] T010 [US3] Update `frontend/components/charts/PriorityChart.tsx` to import colors from `@/lib/tokens`
- [x] T011 [US3] Replace PRIORITY_COLORS hardcoded values with token constants (colors.chart.critical, colors.chart.high, etc.)

**Checkpoint**: PriorityChart renders identically; no visual regression in Analytics page

---

## Phase 5: User Story 4 - Documentation (Priority: P3)

**Goal**: Design system documentation includes token usage guidelines

**Independent Test**: Developer can read docs and successfully use tokens without additional guidance

### Implementation for User Story 4

- [x] T012 [US4] Add "Design Tokens" section to `docs/design-system.md` with overview and file locations
- [x] T013 [US4] Document CSS token usage examples in design-system.md
- [x] T014 [US4] Document TypeScript token usage examples in design-system.md
- [x] T015 [US4] Document when to use tokens vs. Tailwind classes in design-system.md

**Checkpoint**: Design system docs include complete tokens reference section

---

## Phase 6: Polish & Verification

**Purpose**: Final verification and cleanup

- [x] T016 [P] Verify frontend builds without TypeScript errors (`npm run build`)
- [ ] T017 [P] Verify CSS tokens visible in browser DevTools on `:root` (manual verification needed)
- [ ] T018 [P] Verify dark mode tokens apply when `.dark` class is active (manual verification needed)
- [ ] T019 Verify PriorityChart renders correctly in Analytics page (manual verification needed)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **US1 - CSS Tokens (Phase 2)**: Depends on Setup
- **US2 - TypeScript Tokens (Phase 3)**: Depends on Setup (can run parallel to US1)
- **US3 - Component Migration (Phase 4)**: Depends on US2 (needs TypeScript tokens)
- **US4 - Documentation (Phase 5)**: Depends on US1 and US2 (documents completed system)
- **Polish (Phase 6)**: Depends on all user stories

### User Story Dependencies

```
Phase 1: Setup
    ↓
┌───┴───┐
↓       ↓
US1     US2  ← Can run in parallel (different files)
(CSS)   (TS)
    ↓
   US3  ← Depends on US2 only
    ↓
   US4  ← Depends on US1 + US2
    ↓
 Polish
```

### Parallel Opportunities

**Within Phase 3 (US2 - TypeScript Tokens):**
```bash
# These can run in parallel (different files):
T006: Create colors.ts
T007: Create spacing.ts
T008: Create typography.ts
```

**Within Phase 6 (Polish):**
```bash
# These verification tasks can run in parallel:
T016: Verify build
T017: Verify DevTools
T018: Verify dark mode
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: US1 - CSS Tokens
3. **STOP and VALIDATE**: Verify tokens in DevTools
4. Feature is usable for CSS-based styling

### Full Implementation

1. Complete Setup + US1 (CSS Tokens) → CSS tokens ready
2. Complete US2 (TypeScript Tokens) → Programmatic access ready
3. Complete US3 (Component Migration) → Pattern demonstrated
4. Complete US4 (Documentation) → Team can adopt
5. Complete Polish → Verified and ready for PR

### Parallel Execution (2 Developers)

```
Developer A: US1 (CSS Tokens)
Developer B: US2 (TypeScript Tokens)
    ↓ (sync point)
Developer A: US3 (Component Migration)
Developer B: US4 (Documentation)
    ↓
Both: Polish & Verification
```

---

## Summary

| Metric | Count |
|--------|-------|
| Total Tasks | 19 |
| US1 Tasks | 4 |
| US2 Tasks | 4 |
| US3 Tasks | 2 |
| US4 Tasks | 4 |
| Setup Tasks | 1 |
| Polish Tasks | 4 |
| Parallel Opportunities | 7 tasks marked [P] |

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- No unit tests needed - visual verification and TypeScript compilation
- All token values are defined in data-model.md
- Commit after each phase completion
- Stop at any checkpoint to validate independently
