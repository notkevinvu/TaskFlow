# TaskFlow Constitution

## Core Principles

### I. Clean Architecture

All backend code follows Clean Architecture with strict layer separation:
- **Domain:** Business logic and entities - no external dependencies
- **Repository:** Data access interfaces and implementations - only database concerns
- **Service:** Business orchestration - coordinates domain and repository
- **Handler:** HTTP concerns only - request/response handling, validation

Layer dependencies flow inward only. Handlers depend on services, services depend on repositories and domain. Domain depends on nothing.

### II. Type Safety & SQL Security

Type safety is mandatory across the entire stack:
- **Backend (Go):** Use pgx with parameterized queries. NEVER use string concatenation for SQL
- **Frontend (TypeScript):** Strict TypeScript with no `any` types in business logic
- **API Contracts:** All API endpoints must have typed request/response structures

SQL injection vulnerabilities are unacceptable. All database queries must use parameterized statements.

### III. Test-Driven Quality

Testing is required for all business-critical code:
- **Backend:** Table-driven tests for handlers, services, and domain logic
- **Frontend:** React Testing Library for components with user interactions
- **Coverage:** Aim for high coverage on priority calculation, task management, and auth flows

Tests should be independent, repeatable, and fast. Mock external dependencies appropriately.

### IV. Design System Consistency

UI development follows established patterns:
- **Components:** Build on shadcn/ui as the foundation
- **Styling:** Tailwind CSS with design tokens from `globals.css`
- **Documentation:** All new UI patterns must be documented in `docs/design-system.md`
- **Tokens:** Use semantic color tokens (`--primary`, `--destructive`, etc.) - no hardcoded colors

Visual consistency and accessibility are non-negotiable.

### V. Simplicity & YAGNI

Keep implementations minimal and focused:
- **No over-engineering:** Only build what's explicitly required
- **No premature abstraction:** Three similar lines are better than an unused utility
- **No feature creep:** Defer nice-to-haves to future iterations
- **Clear boundaries:** Each feature should be independently testable and deployable

## Development Standards

### Code Quality Gates

Before merging any feature:
- [ ] All tests pass
- [ ] No TypeScript errors
- [ ] Go builds without warnings
- [ ] SQL uses parameterized queries only
- [ ] UI follows design system tokens
- [ ] Changes are documented where appropriate

### Performance Requirements

- Page load: < 1 second
- API response: < 500ms for standard operations
- Priority recalculation: < 100ms per task
- Support 1,000+ tasks per user without degradation

### Security Requirements

- JWT-based authentication with secure password hashing (bcrypt)
- Users can only access their own data
- All user inputs validated at handler layer
- No secrets in code or logs

## Workflow Standards

### Feature Development Process

1. **Specify:** Define requirements using `/specify` command
2. **Plan:** Create implementation plan using `/plan` command
3. **Task:** Break down into actionable items using `/tasks` command
4. **Implement:** Execute with incremental delivery
5. **Review:** Verify compliance with constitution principles

### API Development Pattern

1. Add handler method in `backend/internal/handler/`
2. Register route in `backend/cmd/server/main.go`
3. Add API client method in `frontend/lib/api.ts`
4. Create React Query hook in `frontend/hooks/`
5. Write tests for handler and integration

### UI Development Pattern

1. Create component in `frontend/components/`
2. Use shadcn/ui components as base
3. Apply Tailwind classes with design tokens
4. Document new patterns in `docs/design-system.md`
5. Write component tests for user interactions

## Governance

This constitution supersedes all other development practices. All code changes must:
- Comply with the five core principles
- Pass the code quality gates
- Follow the established workflow standards

Amendments to this constitution require:
- Documentation of the change rationale
- Review and approval
- Migration plan for existing code if needed

Use `.claude/CLAUDE.md` for runtime development guidance and detailed conventions.

**Version**: 1.0.0 | **Ratified**: 2025-12-01 | **Last Amended**: 2025-12-01
