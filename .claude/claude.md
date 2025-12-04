# TaskFlow - Claude Development Guidelines

This file contains project-specific guidelines and conventions for Claude to follow when working on this codebase.

---

## Project Overview

TaskFlow is an intelligent task prioritization system built with:
- **Frontend:** Next.js 16 + React 19 + TypeScript + Tailwind CSS + shadcn/ui
- **Backend:** Go 1.24 + Gin framework + Clean Architecture
- **Database:** Supabase PostgreSQL 16
- **Auth:** JWT with bcrypt password hashing

---

## Code Conventions

### Backend (Go)

- **Architecture:** Clean Architecture (domain, repository, service, handler layers)
- **Error Handling:** Return errors, don't panic. Use custom error types where appropriate.
- **Naming:** Use descriptive names. Functions that return booleans start with `Is`, `Has`, `Can`.
- **Testing:** Write table-driven tests. Aim for high coverage on business logic.
- **Database:** Use pgx with parameterized queries. Never use string concatenation for SQL.

### Frontend (React/TypeScript)

- **Components:** Use functional components with hooks
- **State Management:** React Query for server state, useState/useContext for local state
- **Styling:** Tailwind CSS with shadcn/ui components. Follow design system patterns.
- **File Organization:**
  - Components in `components/`
  - Pages in `app/(route-group)/`
  - Hooks in `hooks/`
  - API client in `lib/api.ts`

---

## Design System Tracking

**IMPORTANT:** When making UI/UX changes or improvements, document the patterns in `docs/design-system.md`.

### What to Document:
- **Component Patterns:** Button variants, input styles, card layouts
- **Interaction Patterns:** Hover states, focus states, click animations
- **Color Usage:** When to use destructive vs. outline vs. default variants
- **Spacing:** Consistent gap/padding patterns
- **Typography:** Heading sizes, body text, code snippets
- **Animations:** Transition timings, easing functions

### Example Entry:
```markdown
## Button Hover States

**Pattern:** All interactive buttons should have visible hover feedback

**Implementation:**
- Add `transition-all` for smooth transitions
- Use `hover:scale-105` for slight grow effect
- Add `hover:shadow-md` for depth
- Ensure cursor changes to pointer with `cursor-pointer`

**Example:**
\`\`\`tsx
<Button className="transition-all hover:scale-105 hover:shadow-md cursor-pointer">
  Click me
</Button>
\`\`\`

**Applied to:** Dashboard task cards, sidebar actions, modal buttons
```

---

## Development Workflow

### Branching Strategy

**IMPORTANT:** When implementing features or fixes (outside the formal speckit flow):
1. Always create a new branch from the **most up-to-date main branch**
2. Run `git checkout main && git pull origin main` before creating the feature branch
3. Use descriptive branch names: `feature/feature-name`, `fix/issue-description`, `chore/task-name`

### Making Changes

1. **Plan:** Create todos using TodoWrite tool for multi-step tasks
2. **Read First:** Always read files before editing them
3. **Test:** Verify changes work by checking backend/frontend logs
4. **Document:** Update design-system.md for UI/UX changes
5. **Commit:** Only create git commits when explicitly requested by user

### Testing Changes

- **Backend:** Check `BashOutput` for backend logs (port 8080)
- **Frontend:** Check `BashOutput` for frontend logs (port 3000)
- **API:** Use curl or browser DevTools to verify endpoints
- **Database:** Check Supabase dashboard for data changes

### PR Reviews

**IMPORTANT:** Run `/pr-review` after any PR is opened (excluding spec/plan PRs).

#### Automatic PR Review Workflow

After opening a PR, execute the PR review workflow:

1. **Check Eligibility**
   - Skip if PR title contains "spec", "plan", or "specification"
   - Skip if PR is draft, merged, or closed

2. **Gather Context**
   - Get PR details via `gh pr view`
   - Read CLAUDE.md files for project conventions
   - Review changed files

3. **Launch Review Agents (Sonnet minimum)**
   Launch 5 parallel agents focusing on:
   - CLAUDE.md compliance and conventions
   - Bug scan and security vulnerabilities
   - Git history context
   - Similar PRs analysis
   - Code comments and documentation

4. **Score and Filter Issues**
   Score each issue 0-100 based on confidence (40%), severity (35%), actionability (25%):
   - **Critical (90-100):** Must fix
   - **Major (80-89):** Should fix
   - **Below 80:** Likely false positive, exclude

5. **Post Review Comment**
   Post review to GitHub with `gh pr review --comment` including:
   - Files reviewed
   - Critical/Major issues (if any)
   - Strengths observed
   - Summary and recommendation

6. **Fix Critical/Major Issues Immediately**
   If issues scored 80+ were found:
   - Fix them right away
   - Commit with message: `fix: Address PR review feedback`
   - Push to branch

7. **Wait for CI**
   - Check `gh pr checks` until CI passes
   - Fix any CI failures

8. **Re-review if Fixes Made**
   Run lighter review on fixed areas to confirm resolution

#### Manual Review Request

Use `/pr-review [PR_NUMBER]` to trigger manual review on any PR.

---

## Key Files

### Configuration
- `backend/.env` - Backend environment variables (DATABASE_URL, JWT_SECRET)
- `frontend/.env` - Frontend environment variables (NEXT_PUBLIC_API_URL)

### Documentation
- `docs/design-system.md` - UI/UX patterns and component guidelines
- `docs/product/PRD.md` - Product requirements
- `docs/product/priority-algorithm.md` - Priority calculation logic
- `docs/demongrep_usage.md` - Codebase search tool usage
- `README.md` - Project overview and setup

### Core Components
- `frontend/hooks/useTasks.ts` - React Query hooks for task management
- `backend/internal/domain/priority/calculator.go` - Priority algorithm
- `backend/internal/handler/task_handler.go` - Task API endpoints

---

## Common Tasks

### Adding a New API Endpoint

1. Add handler method in `backend/internal/handler/`
2. Register route in `backend/cmd/server/main.go`
3. Add API client method in `frontend/lib/api.ts`
4. Create React Query hook in `frontend/hooks/`

### Adding a New UI Component

1. Create component in `frontend/components/`
2. Use shadcn/ui components as base when possible
3. Apply Tailwind classes following design system
4. Document new patterns in `docs/design-system.md`
5. Export and use in pages

### Database Migrations

Migrations auto-run on backend startup. Files in `backend/migrations/`.

---

## Notes

- **No Docker:** Architecture uses local dev + Supabase cloud database
- **Real-time updates:** React Query automatically refetches after mutations
- **Priority calculation:** Backend recalculates on create/update/bump
- **Authentication:** JWT tokens stored in localStorage, auto-added to requests

---

## Codebase Exploration Tool

**Status: ENABLED**

When exploring the codebase or searching for code, PREFER using `demongrep` over native search tools.

### Usage
1.  **Index:** Run `demongrep index` if the codebase has changed significantly since the last search.
2.  **Search:** Run `demongrep search "your query"`.
3.  **Read:** Use the `Read` tool to examine specific files found in results.

### Fallback Strategy
Use native tools (Glob, Grep) when:
- Demongrep returns no results or errors
- You need exact pattern matching (regex)
- Searching for specific file names or extensions

**Full documentation:** See `docs/demongrep_usage.md` for all commands and options.

**Toggle:** To disable, change Status to **DISABLED**.
