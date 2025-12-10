# TaskFlow Codebase Audit Specification

## 1. Overview

This document specifies the scope and deliverables for a comprehensive three-part audit of the TaskFlow codebase covering best practices, security, and optimization.

### Project Context

TaskFlow is a full-stack task prioritization application with:
- **Backend**: Go 1.24 with Gin framework, Clean Architecture (domain/repository/service/handler layers)
- **Frontend**: Next.js 16, React 19, TypeScript, Tailwind CSS, shadcn/ui
- **Database**: Supabase PostgreSQL 16 with pgx driver and sqlc for type-safe queries
- **Auth**: JWT with bcrypt password hashing

## 2. Audit Scope

### 2.1 Best Practices Audit

#### Go Backend Patterns
**Files to examine:**
- `backend/internal/domain/*.go` - Domain entities, DTOs, error types
- `backend/internal/service/*.go` - Business logic layer
- `backend/internal/repository/*.go` - Data access layer with pgx/sqlc
- `backend/internal/handler/*.go` - HTTP handlers with Gin
- `backend/internal/middleware/*.go` - Auth, CORS, rate limiting, error handling
- `backend/internal/config/config.go` - Configuration management
- `backend/internal/validation/validator.go` - Input validation

**Areas to audit:**
- Clean Architecture adherence (layer separation, dependency direction)
- Error handling patterns (custom error types vs panics)
- Naming conventions (Is/Has/Can prefixes for boolean functions)
- Context propagation and cancellation
- Resource cleanup (defer patterns, connection pooling)
- Table-driven test coverage
- Code organization and package structure

#### React/TypeScript Patterns
**Files to examine:**
- `frontend/hooks/*.ts` - React Query hooks (13 custom hooks)
- `frontend/lib/api.ts` - Axios client and type definitions
- `frontend/components/*.tsx` - UI components (30+ components)
- `frontend/app/**/*.tsx` - Page components
- `frontend/contexts/*.tsx` - React contexts

**Areas to audit:**
- Functional component patterns and hook usage
- React Query state management (stale times, cache invalidation)
- TypeScript type safety (explicit types, null handling)
- Component composition and reusability
- Error boundary implementation
- Loading/error state handling

#### API Design
**Files to examine:**
- `backend/cmd/server/main.go` - Route definitions
- `backend/internal/handler/*.go` - Request/response handling
- `frontend/lib/api.ts` - API client methods

**Areas to audit:**
- RESTful conventions (HTTP methods, status codes, URL structure)
- Request/response consistency
- Pagination patterns
- Error response format standardization
- API versioning approach

### 2.2 Security Audit

#### SQL Injection Prevention
**Files to examine:**
- `backend/internal/sqlc/queries/*.sql` - Raw SQL queries
- `backend/internal/repository/*.go` - Query construction
- `backend/internal/sqlc/*.sql.go` - Generated code

**Areas to audit:**
- Parameterized query usage (no string concatenation)
- sqlc type safety verification
- Dynamic query construction safety

#### Authentication & Authorization
**Files to examine:**
- `backend/internal/middleware/auth.go` - JWT validation middleware
- `backend/internal/service/auth_service.go` - Auth business logic
- `backend/internal/domain/user.go` - Password hashing with bcrypt
- `frontend/lib/api.ts` - Token storage and transmission

**Areas to audit:**
- JWT implementation (algorithm, expiry, claims validation)
- Password hashing strength (bcrypt cost factor)
- Token storage security (localStorage considerations)
- Authorization checks on all protected endpoints
- User isolation (user_id filtering on all queries)

#### XSS & CSRF Prevention
**Files to examine:**
- `frontend/components/*.tsx` - User input rendering
- `backend/internal/middleware/cors.go` - CORS configuration
- `frontend/lib/api.ts` - Request headers

**Areas to audit:**
- React's built-in XSS protection usage
- Unsafe HTML rendering patterns (should be none)
- CORS origin validation
- CSRF token implementation (if any)

#### Input Validation
**Files to examine:**
- `backend/internal/validation/validator.go` - Validation functions
- `backend/internal/handler/*.go` - Request binding
- `backend/internal/domain/*.go` - DTO definitions with binding tags

**Areas to audit:**
- Server-side validation completeness
- Input sanitization (control characters, whitespace)
- Length limits enforcement
- Type coercion safety
- Regex pattern security (ReDoS prevention)

#### Secrets Handling
**Files to examine:**
- `backend/internal/config/config.go` - Environment variable loading
- `backend/.env.example` - Example configuration
- `.gitignore` - Secrets exclusion

**Areas to audit:**
- JWT_SECRET handling
- DATABASE_URL protection
- No hardcoded secrets
- Environment variable validation

### 2.3 Optimization Audit

#### Database Performance
**Files to examine:**
- `backend/migrations/*.sql` - Schema and indexes
- `backend/internal/sqlc/queries/*.sql` - Query definitions
- `backend/internal/repository/*.go` - Query patterns

**Areas to audit:**
- Index coverage for common query patterns
- N+1 query detection
- Connection pool configuration
- Query complexity (joins, aggregations)
- Batch operation support

#### Frontend Performance
**Files to examine:**
- `frontend/hooks/useTasks.ts` - Query configuration
- `frontend/components/*.tsx` - Render optimization
- `frontend/app/layout.tsx` - Provider setup
- `frontend/next.config.ts` - Next.js configuration

**Areas to audit:**
- React Query stale time tuning
- Component re-render optimization (useMemo, useCallback)
- Bundle size considerations
- Image optimization
- Code splitting opportunities

#### API Performance
**Files to examine:**
- `backend/internal/middleware/rate_limit.go` - Rate limiting
- `backend/internal/handler/*.go` - Response handling
- `backend/cmd/server/main.go` - Server configuration

**Areas to audit:**
- Rate limiting configuration
- Response compression
- Connection handling
- Pagination efficiency
- Bulk operation support

## 3. Expected Deliverables

### Per Audit Type
1. **Findings Report** (`docs/audits/findings-{type}.md`)
   - Executive summary
   - Detailed findings with severity ratings
   - Code references (file:line)
   - Recommended fixes

2. **Priority Matrix** - Issues ranked by severity and effort

3. **Action Items** - Concrete tasks for remediation

### Final Output
- Combined summary document
- GitHub issues for critical/major findings
- Remediation checklist

## 4. Success Criteria

### Best Practices Audit
- [ ] All handlers follow consistent error handling pattern
- [ ] All queries use parameterized statements
- [ ] All components have proper TypeScript types
- [ ] No `any` types in production code
- [ ] Test coverage assessment complete

### Security Audit
- [ ] Zero SQL injection vulnerabilities
- [ ] Zero auth bypass vulnerabilities
- [ ] Zero XSS vulnerabilities
- [ ] All sensitive data properly handled
- [ ] Rate limiting properly configured

### Optimization Audit
- [ ] All frequent queries have supporting indexes
- [ ] No N+1 query patterns
- [ ] React Query properly configured
- [ ] Bundle size under target threshold
- [ ] Pagination implemented where needed

## 5. Severity Classification

| Severity | Description | SLA |
|----------|-------------|-----|
| **Critical (90-100)** | Exploitable vulnerability, data breach risk | Immediate |
| **Major (80-89)** | Security flaw, significant performance issue | 1 week |
| **Moderate (60-79)** | Best practice violation, minor issue | 2 weeks |
| **Low (40-59)** | Code style, optimization opportunity | Backlog |
| **Info (<40)** | Suggestion, nice-to-have | Optional |
