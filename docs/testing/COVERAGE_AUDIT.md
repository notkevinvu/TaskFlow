# Test Coverage Audit Report

**Date:** 2025-11-29
**Overall Coverage:** 8.0%
**Target Coverage:** >80%

---

## Executive Summary

Current test coverage is **critically low at 8.0%**. While some foundational tests exist, most business logic, handlers, and services are completely untested. This represents a significant risk for production deployment.

### Key Findings

✅ **Good Coverage (>80%)**
- `config` - 81.2%
- `validation` - 84.4%
- `logger` - 100%

⚠️ **Needs Attention**
- `priority/calculator` - 100% but tests are **FAILING** (scale mismatch in test data)
- `repository` - 1.2% (only integration test stubs exist)

❌ **Critical Gaps (0% coverage)**
- `cmd/server` - Main application entry point
- `domain` - Domain models and types
- `handler` - **All API handlers** (auth, task, category)
- `middleware` - CORS, rate limiting, auth middleware
- `service` - **All business logic** (auth, task, priority services)
- `sqlc` - Generated SQL queries (expected, but should be covered by integration tests)
- `ratelimit` - Rate limiting implementation

---

## Detailed Coverage by Package

### ✅ Well-Tested (>70%)

#### `internal/config` - 81.2%
- Configuration loading and validation
- Environment variable parsing
- **Action:** Increase to 100% by testing edge cases

#### `internal/validation` - 84.4%
- Text sanitization
- Email/password validation
- Priority/category validation
- **Gap:** `ValidateOptionalText` at 0%
- **Action:** Add tests for optional text validation

#### `internal/logger` - 100%
- Structured logging configuration
- **Status:** Complete

### ⚠️ Partially Tested

#### `internal/domain/priority/calculator` - 100% (but failing)
**Issue:** Tests use incorrect scale for `UserPriority`
- Tests expect 0-100 scale
- Implementation expects 1-10 scale (correct per domain model)

**Failing Tests:**
- ❌ TestCalculate (all 6 sub-tests failing)
- ❌ TestCalculateDeadlineUrgency (1 sub-test failing)
- ✅ TestCalculateTimeDecay (passing)
- ✅ TestCalculateBumpPenalty (passing)
- ✅ TestGetEffortBoost (passing)
- ✅ TestIsAtRisk (passing)

**Action:** Fix test data to use 1-10 scale instead of 0-100

#### `internal/repository` - 1.2%
**Existing Tests:**
- `task_repository_test.go` - Integration test stubs
- `user_repository_test.go` - Integration test stubs
- `task_history_repository_test.go` - Integration test stubs

**Gap:** Tests exist but provide minimal coverage (11 tests passing)
**Action:** Expand integration tests with testcontainers to cover all repository methods

---

### ❌ Critical Gaps (0% coverage)

#### `internal/handler` - 0%
**Missing Tests:**
- `auth_handler.go` - Login, Register, GetCurrentUser
- `task_handler.go` - CRUD operations, priority updates, bump tracking
- `category_handler.go` - Rename, delete, validation
- `analytics_handler.go` - Metrics endpoints

**Risk:** High - handlers are the API surface, untested means no validation of request/response contracts

**Priority:** **CRITICAL**

#### `internal/service` - 0%
**Missing Tests:**
- `auth_service.go` - Authentication logic, token generation
- `task_service.go` - Business logic, priority calculation integration
- `priority_service.go` - Priority recalculation
- `analytics_service.go` - Metrics aggregation

**Risk:** High - untested business logic is a recipe for bugs

**Priority:** **CRITICAL**

#### `internal/middleware` - 0%
**Missing Tests:**
- `auth_middleware.go` - JWT validation
- `cors.go` - CORS configuration
- `rate_limit.go` - Rate limiting

**Risk:** Medium - security-critical code should be tested

**Priority:** **HIGH**

#### `internal/ratelimit` - 0%
**Missing Tests:**
- Redis-backed rate limiter
- In-memory fallback

**Risk:** Medium - scalability feature needs verification

**Priority:** **MEDIUM**

#### `cmd/server` - 0%
**Note:** Main functions are typically hard to test
**Action:** Extract testable initialization logic into separate functions

---

## Testing Infrastructure Gaps

### Backend

**Missing Dependencies:**
- ❌ `github.com/stretchr/testify` - Assertions and mocks
- ❌ `github.com/testcontainers/testcontainers-go` - Integration tests
- ❌ `github.com/testcontainers/testcontainers-go/modules/postgres` - PostgreSQL test containers

**Missing Interfaces:**
- ❌ Service interfaces for mocking (handlers depend on concrete types)
- ❌ Repository interfaces for mocking (services depend on concrete types)

**Missing Test Utilities:**
- ❌ Test fixtures for common domain objects
- ❌ Mock HTTP server for handler tests
- ❌ Test database setup/teardown helpers

### Frontend

**Missing Dependencies:**
- ❌ `vitest` - Test runner
- ❌ `@testing-library/react` - Component testing
- ❌ `@testing-library/jest-dom` - DOM assertions
- ❌ `@testing-library/user-event` - User interaction simulation
- ❌ `happy-dom` - DOM environment

**Missing Tests:**
- ❌ Component tests (0 files)
- ❌ Hook tests (0 files)
- ❌ Integration tests (0 files)

**Coverage:** 0%

---

## Priority Recommendations

### Phase 1: Fix Existing Tests (1-2 hours)
1. **Fix calculator tests** - Update test data to use 1-10 scale for UserPriority
2. **Verify all existing tests pass** - Ensure green baseline before adding more

### Phase 2: Install Testing Infrastructure (2-3 hours)
1. **Backend:**
   - Install testify, testcontainers
   - Create mock interfaces (refactor handlers/services to accept interfaces)
   - Create test fixtures and helpers

2. **Frontend:**
   - Install vitest, @testing-library/react
   - Configure test environment
   - Create test utilities

### Phase 3: Critical Coverage (2-3 days)
**Target:** Cover 80% of business logic

1. **Handler Tests** (1 day)
   - AuthHandler (login, register, JWT validation)
   - TaskHandler (CRUD, priority updates)
   - CategoryHandler (rename, delete)

2. **Service Tests** (1 day)
   - AuthService with mocked repository
   - TaskService with mocked repository
   - PriorityService (algorithm integration)

3. **Integration Tests** (1 day)
   - Repository tests with testcontainers
   - End-to-end API tests (optional)

### Phase 4: Frontend Testing (1-2 days)
1. Component tests for critical UI (dashboard, task forms)
2. Hook tests (useTasks, useAuth)
3. Integration tests (user flows)

### Phase 5: Coverage Enforcement (1 day)
1. Configure coverage reporting in CI/CD
2. Set minimum coverage threshold (80%)
3. Block PRs that reduce coverage

---

## Testing Strategy

### Unit Tests
**Target:** Business logic, calculators, validators
- **Mock external dependencies** (database, HTTP clients)
- **Fast execution** (<100ms per test)
- **High coverage** (>90% of business logic)

### Integration Tests
**Target:** Repositories, database queries, API contracts
- **Use testcontainers** for real PostgreSQL
- **Test actual SQL queries** (not mocks)
- **Moderate speed** (<5s per test)

### Component Tests (Frontend)
**Target:** React components, hooks
- **Test user interactions** (clicks, input, navigation)
- **Test rendering logic** (conditional UI, data display)
- **Mock API calls**

### End-to-End Tests (Optional - Phase 4+)
**Target:** Full user flows
- **Test critical paths** (register → create task → complete task)
- **Run against real stack** (backend + frontend + database)
- **Slow execution** (>10s per test)
- **Tool:** Playwright or Cypress

---

## Success Criteria

### Phase 3 Exit Criteria (Production Readiness)
- [x] Overall coverage > 80%
- [x] All handlers tested
- [x] All services tested
- [x] Integration tests for repositories
- [x] All tests passing in CI/CD
- [x] Coverage enforcement enabled

### Code Quality Metrics
- **Test execution time:** <30s for unit tests, <2min for integration tests
- **Test reliability:** 0 flaky tests
- **Code coverage:** >80% overall, >90% for business logic
- **Test documentation:** Clear test names, descriptive assertions

---

## Estimated Effort

| Phase | Effort | Priority |
|-------|--------|----------|
| Fix existing tests | 2 hours | CRITICAL |
| Install infrastructure | 3 hours | CRITICAL |
| Handler tests | 1 day | CRITICAL |
| Service tests | 1 day | CRITICAL |
| Integration tests | 1 day | HIGH |
| Frontend tests | 2 days | MEDIUM |
| Coverage enforcement | 0.5 day | HIGH |
| **Total** | **5-6 days** | |

---

## Next Steps

1. ✅ Complete this audit
2. ⬜ Fix calculator tests (1-10 scale)
3. ⬜ Install backend testing dependencies
4. ⬜ Create service/repository interfaces for DI
5. ⬜ Write handler tests
6. ⬜ Write service tests
7. ⬜ Expand integration tests
8. ⬜ Install frontend testing dependencies
9. ⬜ Write component tests
10. ⬜ Configure CI/CD coverage reporting

---

**Report Status:** Complete
**Next Review:** After Phase 3 completion
