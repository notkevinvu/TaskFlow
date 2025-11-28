# Phase 3: Production Readiness & Infrastructure (Weeks 5-6)

**Goal:** Harden the application for production deployment with comprehensive testing, observability, and scalability improvements.

**Prerequisites:** Complete Phase 2.5 (Quick Wins + Core Features Sprint)

**By the end of this phase, you will have:**
- ✅ Comprehensive test coverage (unit + integration)
- ✅ Structured logging with slog
- ✅ Production-grade error handling
- ✅ Multi-stage Docker builds for efficient deployment
- ✅ Interface-based dependency injection (testable architecture)
- ✅ Redis-backed rate limiting (scalable)
- ✅ Health check endpoints
- ✅ Environment-based configuration
- ✅ CI/CD pipeline (GitHub Actions)

**What you're building:** A production-grade backend that can scale to multiple instances, with comprehensive testing and observability.

---

## Current Status Checklist

### ✅ Completed (Phase 1 & 2)
- [x] Next.js + React frontend
- [x] Go backend with Clean Architecture
- [x] PostgreSQL database with Supabase
- [x] JWT authentication
- [x] Priority calculation algorithm
- [x] Bump tracking system
- [x] Dark mode support
- [x] Calendar widget

### ✅ Completed (Phase 2.5)
- [ ] JWT_SECRET security fix
- [ ] Calendar popover positioning
- [ ] Categories dropdown
- [ ] Search & filtering
- [ ] Analytics dashboard with charts
- [ ] Design system documentation

---

## Implementation Roadmap

### Priority 1: Critical Backend Fixes (From Architecture Analysis)
*Reference: `docs/architecture/backend-analysis-report.md`*

- [x] **sqlc Migration for Type-Safe SQL** (code quality + safety improvement)
  - [x] Install and configure sqlc
  - [x] Create `queries/` directory with SQL query files
  - [x] Define schema.sql for sqlc code generation
  - [x] Migrate TaskRepository methods to sqlc (16 methods)
  - [x] Migrate UserRepository methods to sqlc
  - [x] Migrate TaskHistoryRepository methods to sqlc
  - [x] Created integration tests (11 tests, all passing)
  - [x] Verified compile-time safety and auto-generated scanning
  - [x] Benefits achieved: Compile-time safety, auto-generated scanning, easier maintenance
- [ ] **Interface-based Dependency Injection** (testability improvement)
- [ ] **Redis Rate Limiting Migration** (scalability fix)
- [ ] **Structured Logging with slog** (observability)
- [ ] **Custom Error Types & Error Handling** (production-grade error management)
- [ ] **Input Validation Improvements** (special characters, sanitization)

### Priority 2: Testing Infrastructure
- [ ] Audit existing code coverage and identify gaps
- [ ] Backend unit tests (services, priority calculator, handlers)
- [ ] Backend integration tests (repositories with testcontainers)
- [ ] Category handler tests (rename, delete, validation)
- [ ] **Search & filter tests** (Phase 2.5 follow-up):
  - [ ] Backend: Filter combination tests (search + category + priority)
  - [ ] Backend: Edge case tests (min_priority > max_priority, invalid dates)
  - [ ] Frontend: Debounce behavior tests
  - [ ] Frontend: Filter chip removal tests
  - [ ] E2E: Search → filter → clear flow
- [ ] Frontend component tests
- [ ] Coverage reporting and enforcement (target >80%)

### Priority 3: Production Infrastructure & Database Optimization
- [ ] **Database indexes for search performance** (Phase 2.5 follow-up):
  - [ ] `CREATE INDEX idx_tasks_priority_score ON tasks(priority_score)`
  - [ ] `CREATE INDEX idx_tasks_due_date ON tasks(due_date)`
  - [ ] `CREATE INDEX idx_tasks_category ON tasks(category)`
  - [ ] Verify query performance with EXPLAIN ANALYZE
- [ ] Multi-stage Docker builds
- [ ] Docker Compose production config
- [ ] Health check endpoints
- [ ] CI/CD pipeline (GitHub Actions)

### Priority 4: Search & Filter Enhancements (Phase 2.5 follow-up)
- [ ] **Date range picker UI** for due date filtering
  - [ ] Install date picker library (e.g., `react-day-picker` already installed)
  - [ ] Create DateRangePicker component
  - [ ] Add to TaskFilters component
  - [ ] Wire up to backend `due_date_start` and `due_date_end` params
- [ ] **Filter presets** for common queries
  - [ ] "High Priority" preset (priority >= 75)
  - [ ] "Due This Week" preset (due_date_start = today, due_date_end = +7 days)
  - [ ] "At Risk" preset (bump_count >= 3)
  - [ ] "Quick Wins" preset (effort = small, bump_count = 0)
  - [ ] Add preset dropdown to TaskFilters
- [ ] **Filter URL persistence** for shareable links
  - [ ] Sync filters to URL query params
  - [ ] Parse URL params on page load
  - [ ] Update URL without page reload (Next.js router)
- [ ] **Saved searches** (optional)
  - [ ] Save custom filter combinations
  - [ ] Backend endpoint to store user searches
  - [ ] Quick access dropdown


---

## Week 5: Backend Hardening & Testing

### Day 1-2: Interface-Based Dependency Injection

**Goal:** Refactor constructors to accept interfaces instead of concrete types, enabling proper unit testing with mocks.

**Current Issue:**
```go
// handlers currently depend on concrete service types
func NewTaskHandler(s *TaskService) *TaskHandler
// Makes unit testing handlers impossible without real services
```

**Solution:**
```go
// handlers depend on service interfaces
type TaskService interface {
    Create(ctx context.Context, userID uuid.UUID, req *CreateTaskRequest) (*Task, error)
    // ... other methods
}
func NewTaskHandler(s TaskService) *TaskHandler
```

**Tasks:**
- [ ] Create service interfaces in `internal/ports/services.go` (may already exist)
- [ ] Update handler constructors to accept interfaces
- [ ] Update service constructors to accept repository interfaces
- [ ] Update `cmd/server/main.go` to wire dependencies
- [ ] Verify app still compiles and runs

**Checklist:**
- [ ] `internal/handler/task_handler.go` uses `ports.TaskService` interface
- [ ] `internal/handler/auth_handler.go` uses `ports.AuthService` interface
- [ ] `internal/service/task_service.go` uses `ports.TaskRepository` interface
- [ ] `internal/service/auth_service.go` uses `ports.UserRepository` interface
- [ ] All tests pass (if any exist)

---

### Day 3-4: Backend Testing Infrastructure

**Day 3 Morning: Code Coverage Audit**

**Tasks:**
- [ ] Run coverage analysis: `go test ./... -coverprofile=coverage.out`
- [ ] Generate coverage report: `go tool cover -html=coverage.out`
- [ ] Identify untested packages and functions
- [ ] Create testing roadmap for missing coverage
- [ ] Document coverage gaps in testing plan

**Target Areas for Testing:**
- [ ] `internal/handler/category_handler.go` (new code, needs tests)
- [ ] `internal/handler/task_handler.go` (validate all endpoints)
- [ ] `internal/handler/auth_handler.go` (security-critical)
- [ ] `internal/service/task_service.go` (business logic)
- [ ] `internal/service/auth_service.go` (authentication logic)
- [ ] `internal/domain/priority/calculator.go` (algorithm correctness)
- [ ] `internal/repository/*` (data integrity)

**Checklist:**
- [ ] Coverage baseline documented
- [ ] Critical gaps identified (handlers, services, domain logic)
- [ ] Testing plan prioritized by risk and impact

---

**Day 3 Afternoon - Day 4: Implement Tests**

**Install Testing Dependencies:**

```bash
cd backend
go get github.com/stretchr/testify
go get github.com/testcontainers/testcontainers-go
go get github.com/testcontainers/testcontainers-go/modules/postgres
```

**Tasks:**
- [ ] Create mock implementations of interfaces using testify/mock
- [ ] Write unit tests for `CategoryHandler` (rename, delete, validation edge cases)
- [ ] Write unit tests for `TaskHandler` (all CRUD operations)
- [ ] Write unit tests for `AuthHandler` (login, register, token validation)
- [ ] Write unit tests for `PriorityService` (algorithm testing)
- [ ] Write unit tests for `TaskService` (business logic, including category operations)
- [ ] Write unit tests for `AuthService` (authentication logic)
- [ ] Write integration tests for repositories using testcontainers
- [ ] Configure test coverage reporting

**Checklist:**
- [ ] `internal/handler/category_handler_test.go` created (validation, edge cases)
- [ ] `internal/handler/task_handler_test.go` created with comprehensive tests
- [ ] `internal/handler/auth_handler_test.go` created (security validation)
- [ ] `internal/service/priority_service_test.go` created with test cases
- [ ] `internal/service/task_service_test.go` created with mocks
- [ ] `internal/service/auth_service_test.go` created with mocks
- [ ] `internal/repository/task_repository_test.go` integration tests
- [ ] Test coverage > 80% on handlers and services
- [ ] All tests pass: `go test ./internal/...`

**Unit Test Example (`internal/service/auth_service_test.go`):**

```go
package services

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/yourusername/webapp/internal/domain"
)

type MockUserRepository struct {
    mock.Mock
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
    args := m.Called(ctx, email)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
    args := m.Called(ctx, user)
    return args.Error(0)
}

func TestAuthService_Register(t *testing.T) {
    mockRepo := new(MockUserRepository)
    authService := NewAuthService(mockRepo, "test-secret")

    mockRepo.On("GetByEmail", mock.Anything, "test@example.com").Return(nil, nil)
    mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

    req := &domain.CreateUserRequest{
        Email:    "test@example.com",
        Name:     "Test User",
        Password: "password123",
    }

    response, err := authService.Register(context.Background(), req)

    assert.NoError(t, err)
    assert.NotNil(t, response)
    assert.Equal(t, "test@example.com", response.User.Email)
    mockRepo.AssertExpectations(t)
}
```

**Integration Test with Testcontainers:**

```go
//go:build integration

package postgres_test

import (
    "context"
    "testing"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/stretchr/testify/assert"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestUserRepository_Create(t *testing.T) {
    ctx := context.Background()

    // Start PostgreSQL container
    postgresContainer, err := postgres.RunContainer(ctx,
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("postgres"),
        postgres.WithPassword("postgres"),
    )
    assert.NoError(t, err)
    defer postgresContainer.Terminate(ctx)

    // Get connection string
    connStr, err := postgresContainer.ConnectionString(ctx)
    assert.NoError(t, err)

    // Connect
    pool, err := pgxpool.New(ctx, connStr)
    assert.NoError(t, err)
    defer pool.Close()

    // Run migrations
    // ... (use golang-migrate)

    // Test repository
    repo := NewUserRepository(pool)
    user := &domain.User{
        Email: "test@example.com",
        Name:  "Test User",
    }

    err = repo.Create(ctx, user)
    assert.NoError(t, err)
    assert.NotNil(t, user.ID)
}
```

**Run Tests:**

```bash
# Unit tests only
go test ./internal/... -v

# Integration tests
go test -tags=integration ./... -v

# With coverage
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Frontend Testing

**Install Dependencies:**

```bash
cd frontend
npm install -D vitest @testing-library/react @testing-library/jest-dom
npm install -D @testing-library/user-event happy-dom
```

**Configure Vitest (`vitest.config.ts`):**

```typescript
import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'happy-dom',
    setupFiles: ['./test/setup.ts'],
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './'),
    },
  },
});
```

**Component Test Example:**

```typescript
import { render, screen, fireEvent } from '@testing-library/react';
import { Button } from '@/components/ui/button';
import { describe, it, expect, vi } from 'vitest';

describe('Button', () => {
  it('renders correctly', () => {
    render(<Button>Click me</Button>);
    expect(screen.getByText('Click me')).toBeInTheDocument();
  });

  it('handles click events', () => {
    const handleClick = vi.fn();
    render(<Button onClick={handleClick}>Click me</Button>);

    fireEvent.click(screen.getByText('Click me'));
    expect(handleClick).toHaveBeenCalledOnce();
  });
});
```

---

## Week 6: Observability & Production Infrastructure

### Day 1: Custom Error Types & Input Validation

**Goal:** Implement production-grade error handling with custom error types and comprehensive input validation.

**Tasks:**

**Custom Error Types:**
- [ ] Create `internal/domain/errors.go` with custom error types
- [ ] Add error types for validation, authentication, authorization, not found, conflict
- [ ] Update handlers to return specific error types instead of generic errors
- [ ] Create error middleware to map error types to HTTP status codes
- [ ] Add structured error logging with error context

**Input Validation Improvements:**
- [ ] Create `internal/validation/validator.go` utility package
- [ ] Add text sanitization functions (trim whitespace, normalize Unicode)
- [ ] Add validation for special characters in text fields (optional allowlist/blocklist)
- [ ] Validate all user-input text fields consistently:
  - Title (max 200 chars, no control characters)
  - Description (max 2000 chars)
  - Category (max 50 chars) ✅ Already implemented
  - Context (max 500 chars)
  - Related people names (reasonable length)
- [ ] Add validation for email format (RFC 5322 compliance)
- [ ] Add validation for password strength (optional: min length, complexity)
- [ ] Update all handlers to use validation utilities

**Special Character Handling Strategy:**
- Allow most printable Unicode characters for international support
- Trim leading/trailing whitespace automatically
- Reject control characters (0x00-0x1F, 0x7F-0x9F)
- Optional: Reject zero-width characters and combining marks if they cause issues
- Validate length in runes/characters, not bytes (Unicode support)

**Checklist:**
- [ ] `internal/domain/errors.go` created with error types
- [ ] Error middleware maps errors to HTTP status codes
- [ ] `internal/validation/validator.go` created with utilities
- [ ] All text input fields validated consistently
- [ ] Control characters rejected
- [ ] Unicode support maintained (international users)
- [ ] Error responses include helpful messages (not just "invalid input")

**Example Custom Errors (`internal/domain/errors.go`):**
```go
package domain

import "fmt"

type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

type NotFoundError struct {
    Resource string
    ID       string
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

type ConflictError struct {
    Message string
}

func (e *ConflictError) Error() string {
    return fmt.Sprintf("conflict: %s", e.Message)
}

type UnauthorizedError struct {
    Message string
}

func (e *UnauthorizedError) Error() string {
    return e.Message
}
```

**Example Validation Utilities (`internal/validation/validator.go`):**
```go
package validation

import (
    "strings"
    "unicode"
)

// SanitizeText trims whitespace and rejects control characters
func SanitizeText(text string, maxLength int) (string, error) {
    text = strings.TrimSpace(text)

    if len([]rune(text)) > maxLength {
        return "", fmt.Errorf("text exceeds maximum length of %d characters", maxLength)
    }

    // Check for control characters
    for _, r := range text {
        if unicode.IsControl(r) {
            return "", fmt.Errorf("text contains invalid control characters")
        }
    }

    return text, nil
}

// ValidateEmail checks if email format is valid
func ValidateEmail(email string) error {
    // Use regex or library for RFC 5322 validation
    // ...
    return nil
}
```

---

### Day 2: Redis Rate Limiting Migration

**Current Issue:** In-memory rate limiter (`pkg/middleware/rate_limit.go`) uses a Go map that doesn't scale across multiple server instances.

**Goal:** Migrate to Redis-backed rate limiting for horizontal scalability.

**Tasks:**
- [ ] Add Redis to `docker-compose.yml`
- [ ] Install Redis client: `go get github.com/redis/go-redis/v9`
- [ ] Create `pkg/ratelimit/redis_limiter.go`
- [ ] Update rate limit middleware to use Redis
- [ ] Test with multiple requests
- [ ] Update configuration for Redis URL

**Checklist:**
- [ ] Redis container running in docker-compose
- [ ] Rate limiter uses Redis instead of memory
- [ ] Rate limits persist across server restarts
- [ ] Multiple instances share same rate limit counters
- [ ] Graceful fallback if Redis is unavailable (optional)

**Add to `docker-compose.yml`:**
```yaml
redis:
  image: redis:7-alpine
  ports:
    - "6379:6379"
  volumes:
    - redis_data:/data
  healthcheck:
    test: ["CMD", "redis-cli", "ping"]
    interval: 5s
    timeout: 3s
    retries: 5

volumes:
  redis_data:
```

**Example Redis Rate Limiter:**
```go
package ratelimit

import (
    "context"
    "fmt"
    "time"
    "github.com/redis/go-redis/v9"
)

type RedisLimiter struct {
    client *redis.Client
}

func NewRedisLimiter(redisURL string) *RedisLimiter {
    client := redis.NewClient(&redis.Options{
        Addr: redisURL,
    })
    return &RedisLimiter{client: client}
}

func (l *RedisLimiter) Allow(key string, limit int, window time.Duration) (bool, error) {
    ctx := context.Background()

    // Increment counter
    count, err := l.client.Incr(ctx, key).Result()
    if err != nil {
        return false, err
    }

    // Set expiry on first request
    if count == 1 {
        l.client.Expire(ctx, key, window)
    }

    return count <= int64(limit), nil
}
```

---

### Day 3: Structured Logging with slog

**Goal:** Replace `log.Println` with structured logging using Go's built-in `slog` package for better observability.

**Tasks:**
- [ ] Create `pkg/logger/logger.go` with slog configuration
- [ ] Add logger to handlers (inject via constructor)
- [ ] Replace all `log.Println` calls with `logger.Info/Error/Debug`
- [ ] Add request logging middleware
- [ ] Configure JSON logging for production
- [ ] Test log output locally and verify JSON format

**Checklist:**
- [ ] `pkg/logger/logger.go` created
- [ ] All handlers accept logger in constructor
- [ ] No more `log.Println` in codebase
- [ ] Request logging middleware logs all requests
- [ ] JSON logs in production, text logs in development
- [ ] Logs include structured fields (user_id, request_id, etc.)

**Update Backend Logger (`pkg/logger/logger.go`):**

```go
package logger

import (
    "log/slog"
    "os"
)

func New(env string) *slog.Logger {
    var handler slog.Handler

    if env == "production" {
        handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelInfo,
        })
    } else {
        handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
            Level: slog.LevelDebug,
        })
    }

    return slog.New(handler)
}
```

**Update Handlers to Accept Logger:**

```go
func (h *AuthHandler) Login(c *gin.Context) {
    var req domain.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        h.logger.Error("invalid request",
            "error", err,
            "path", c.Request.URL.Path,
        )
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    h.logger.Info("login attempt",
        "email", req.Email,
        "ip", c.ClientIP(),
    )

    // ... rest of handler
}
```

### Error Handling Middleware

**Create `pkg/middleware/error.go`:**

```go
package middleware

import (
    "log/slog"
    "net/http"

    "github.com/gin-gonic/gin"
)

func ErrorHandler(logger *slog.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        if len(c.Errors) > 0 {
            err := c.Errors.Last()

            logger.Error("request error",
                "error", err.Error(),
                "path", c.Request.URL.Path,
                "method", c.Request.Method,
            )

            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "internal server error",
            })
        }
    }
}
```

### Docker Production Build

**Backend Dockerfile (`backend/docker/Dockerfile.prod`):**

```dockerfile
# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/api

# Production stage
FROM gcr.io/distroless/static-debian11

WORKDIR /

COPY --from=builder /app/main /main
COPY --from=builder /app/.env /.env

EXPOSE 8080

CMD ["/main"]
```

**Frontend Dockerfile (`frontend/docker/Dockerfile.prod`):**

```dockerfile
# Dependencies
FROM node:20-alpine AS deps
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

# Builder
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Runner
FROM node:20-alpine AS runner
WORKDIR /app

ENV NODE_ENV production

COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static

EXPOSE 3000

CMD ["node", "server.js"]
```

**Production docker-compose:**

```yaml
version: '3.8'

services:
  backend:
    build:
      context: ./backend
      dockerfile: docker/Dockerfile.prod
    ports:
      - "8080:8080"
    environment:
      - DATABASE_URL=postgresql://postgres:postgres@postgres:5432/webapp_prod
      - JWT_SECRET=${JWT_SECRET}
    depends_on:
      - postgres
    restart: unless-stopped

  frontend:
    build:
      context: ./frontend
      dockerfile: docker/Dockerfile.prod
    ports:
      - "3000:3000"
    environment:
      - NEXT_PUBLIC_API_URL=http://backend:8080
    depends_on:
      - backend
    restart: unless-stopped

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: webapp_prod
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

volumes:
  postgres_data:
```

### Health Checks

**Add to Backend:**

```go
router.GET("/health", func(c *gin.Context) {
    // Check database
    if err := pool.Ping(c.Request.Context()); err != nil {
        c.JSON(503, gin.H{
            "status": "unhealthy",
            "database": "down",
        })
        return
    }

    c.JSON(200, gin.H{
        "status": "healthy",
        "database": "up",
        "version": "1.0.0",
    })
})
```

---

### Day 4-6: Testing, CI/CD Pipeline & Final Polish

**Goal:** Automate testing and deployment with GitHub Actions.

**Tasks:**
- [ ] Create `.github/workflows/test.yml` for automated testing
- [ ] Create `.github/workflows/deploy.yml` for deployment
- [ ] Configure test coverage reporting
- [ ] Add build status badge to README
- [ ] Final testing pass of all features

**Checklist:**
- [ ] GitHub Actions workflow runs tests on PR
- [ ] Tests must pass before merge
- [ ] Coverage report generated and uploaded
- [ ] All Phase 3 features working end-to-end

**Example GitHub Actions Workflow (`.github/workflows/test.yml`):**
```yaml
name: Test

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main, develop]

jobs:
  backend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - name: Run tests
        working-directory: ./backend
        run: |
          go test ./... -v -coverprofile=coverage.out
          go tool cover -func=coverage.out

  frontend-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - name: Install dependencies
        working-directory: ./frontend
        run: npm ci
      - name: Run tests
        working-directory: ./frontend
        run: npm test
```

---

## Phase 3 Completion Checklist

### Backend Architecture
- [ ] Interface-based DI implemented
- [ ] All constructors accept interfaces
- [ ] Codebase compiles and runs with new architecture

### Testing
- [ ] Priority service unit tests written
- [ ] Task service unit tests with mocks
- [ ] Auth service unit tests with mocks
- [ ] Repository integration tests with testcontainers
- [ ] Frontend component tests configured
- [ ] Test coverage > 70% on backend services
- [ ] All tests passing

### Observability
- [ ] Structured logging with slog implemented
- [ ] All `log.Println` replaced with structured logs
- [ ] Request logging middleware added
- [ ] JSON logs in production mode
- [ ] Error handling middleware added

### Scalability
- [ ] Redis rate limiting implemented
- [ ] In-memory rate limiter removed
- [ ] Rate limits work across multiple instances

### Production Infrastructure
- [ ] Multi-stage Docker builds created
- [ ] Production docker-compose configured
- [ ] Health check endpoints implemented
- [ ] Redis integrated for rate limiting
- [ ] Environment variables properly configured

### CI/CD
- [ ] GitHub Actions workflow for testing
- [ ] Tests run automatically on PR
- [ ] Build status visible in README

---

## Summary

**Phase 3 Achievements:**

1. **Testability:** Refactored to interface-based DI, enabling proper unit testing
2. **Test Coverage:** >70% coverage on business logic with unit and integration tests
3. **Observability:** Structured logging with slog, request/error logging middleware
4. **Scalability:** Redis rate limiting for horizontal scaling
5. **Production Ready:** Docker builds, health checks, proper configuration
6. **Automation:** CI/CD pipeline for continuous testing

**Production Readiness Score:** 8/10
- ✅ Horizontal scalability (Redis)
- ✅ Observability (structured logs)
- ✅ Testing (automated tests)
- ✅ Security (JWT enforcement from Phase 2.5)
- ✅ Containerization (Docker)
- ⚠️  Monitoring/Alerting (defer to Phase 4)
- ⚠️  Distributed tracing (defer to Phase 4)

**Next Phase:** Advanced features and optimization (Phase 4)
- Background jobs & workers
- Advanced analytics
- Performance optimization
- Kubernetes deployment
- Monitoring & alerting (Prometheus, Grafana)

---

## Deferred Feature: Anonymous User Support

**Added:** 2025-11-25
**Priority:** Medium (Phase 3.5 or Phase 4)
**Effort:** ~3-5 days

### Goal
Allow users to try TaskFlow without registering, then optionally create an account to save their data.

### Research Summary

**Recommended Approach:** Cookie-based anonymous sessions with database records

**Architecture:**
```
1. User visits site → Backend creates anonymous user + secure cookie
2. Tasks are created normally, belong to anonymous user
3. User registers → Migrate all anonymous tasks to new registered account
4. Cleanup job deletes old anonymous users after 30 days
```

**Database Changes:**
```sql
-- Migration: Add anonymous user support
ALTER TABLE users ADD COLUMN is_anonymous BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN anonymous_expires_at TIMESTAMPTZ;
CREATE INDEX idx_users_anonymous ON users(is_anonymous, anonymous_expires_at);
```

**Backend Implementation:**
- [ ] Add anonymous session middleware (create cookie if not present)
- [ ] Create anonymous user endpoint (auto-called by middleware)
- [ ] Add `POST /auth/convert-anonymous` endpoint (migrate data on registration)
- [ ] Add cleanup job to delete expired anonymous users (30 days)
- [ ] Cookie configuration: httpOnly, Secure, SameSite=Strict

**Frontend Implementation:**
- [ ] Show "Create account to save your tasks" banner for anonymous users
- [ ] Handle anonymous → registered conversion flow
- [ ] No other changes needed (cookies handled automatically)

**Security Considerations:**
- [ ] Cookie must be httpOnly (prevent XSS)
- [ ] Cookie must be Secure in production (HTTPS only)
- [ ] Rate limit anonymous user creation (prevent abuse)
- [ ] Set cookie expiry (30 days)
- [ ] Clean up old anonymous users regularly

**Testing:**
- [ ] Test anonymous user creation
- [ ] Test task creation as anonymous user
- [ ] Test registration + data migration
- [ ] Test cookie persistence across sessions
- [ ] Test cleanup job

**Notes:**
- Defer until Phase 3.5 or Phase 4
- Requires cookie middleware (not currently implemented)
- Consider privacy implications (GDPR, data retention)
- May need "Export data" feature for anonymous users

---

**Updated:** Phase ordering adjusted to include anonymous support

**Next Phase:** Advanced features and optimization (Phase 4)
- Anonymous user support (Phase 3.5 or early Phase 4)
- Background jobs & workers
- Advanced analytics
- Performance optimization
- Kubernetes deployment
- Monitoring & alerting (Prometheus, Grafana)
