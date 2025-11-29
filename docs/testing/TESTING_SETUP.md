# Testing Infrastructure Setup Guide

**Status:** In Progress
**Last Updated:** 2025-11-29
**Phase:** Phase 3 - Production Readiness

---

## Overview

This document tracks the setup and configuration of the testing infrastructure for TaskFlow, including backend (Go) and frontend (TypeScript/React) testing.

---

## Backend Testing Setup

### ✅ Completed Steps

#### 1. Dependencies Installed

```bash
# Testing framework and assertions
go get github.com/stretchr/testify@v1.11.1

# Integration testing with Docker containers
go get github.com/testcontainers/testcontainers-go@v0.40.0
go get github.com/testcontainers/testcontainers-go/modules/postgres@v0.40.0
```

**Testify** provides:
- `assert` package - Simple assertions (e.g., `assert.Equal`, `assert.NoError`)
- `require` package - Assertions that stop test execution on failure
- `mock` package - Mock generation for interfaces
- `suite` package - Test suite runner with setup/teardown

**Testcontainers** provides:
- Real PostgreSQL instances for integration tests
- Automatic container lifecycle management
- Isolated test environments
- Parallel test execution support

#### 2. Existing Test Coverage

| Package | Coverage | Status | Notes |
|---------|----------|--------|-------|
| `config` | 81.2% | ✅ Good | Configuration parsing and validation |
| `logger` | 100% | ✅ Excellent | Structured logging setup |
| `validation` | 84.4% | ✅ Good | Input validation utilities |
| `priority/calculator` | 100% | ✅ Excellent | **Fixed** - Priority algorithm tests |
| `repository` | 1.2% | ⚠️ Needs Work | Integration test stubs exist |
| `handler` | 0% | ❌ Missing | **HIGH PRIORITY** |
| `service` | 0% | ❌ Missing | **HIGH PRIORITY** |
| `middleware` | 0% | ❌ Missing | Security-critical code |

**Overall Coverage:** 8.0% → **Target: 80%+**

#### 3. Fixed Failing Tests

**Issue:** Priority calculator tests were failing due to scale mismatch
- Tests used 0-100 scale for `UserPriority`
- Implementation expects 1-10 scale (per domain model)

**Resolution:** Updated test fixtures to use correct 1-10 scale

**Results:** ✅ All 30 priority calculator tests now passing

---

### 🔄 In Progress

#### 4. Interface-Based Dependency Injection

**Goal:** Enable proper unit testing with mocks by depending on interfaces instead of concrete types.

**Current Architecture (Problematic):**
```go
// Handler depends on concrete service type
func NewTaskHandler(service *TaskService) *TaskHandler {
    return &TaskHandler{service: service}
}

// Cannot mock service for handler tests!
```

**Target Architecture (Testable):**
```go
// 1. Define interface in ports package
type TaskService interface {
    CreateTask(ctx context.Context, req *CreateTaskRequest) (*Task, error)
    GetTask(ctx context.Context, id uuid.UUID) (*Task, error)
    // ... other methods
}

// 2. Handler depends on interface
func NewTaskHandler(service ports.TaskService) *TaskHandler {
    return &TaskHandler{service: service}
}

// 3. Can now mock service in tests!
type MockTaskService struct {
    mock.Mock
}

func (m *MockTaskService) CreateTask(ctx context.Context, req *CreateTaskRequest) (*Task, error) {
    args := m.Called(ctx, req)
    return args.Get(0).(*Task), args.Error(1)
}
```

**Files to Update:**
- [ ] `internal/ports/services.go` - Define service interfaces
- [ ] `internal/ports/repositories.go` - Define repository interfaces
- [ ] `internal/handler/*.go` - Update constructors to accept interfaces
- [ ] `internal/service/*.go` - Update constructors to accept repository interfaces
- [ ] `cmd/server/main.go` - Wire dependencies (concrete implementations)

**Benefits:**
- ✅ Handlers can be unit tested without real services
- ✅ Services can be unit tested without real repositories
- ✅ Fast test execution (no database needed for unit tests)
- ✅ Better separation of concerns

---

### ⏳ Next Steps

#### 5. Handler Tests

**Priority:** CRITICAL

**Target Coverage:** >80%

**Test Structure:**
```go
// internal/handler/auth_handler_test.go
package handler

import (
    "testing"
    "net/http"
    "net/http/httptest"

    "github.com/gin-gonic/gin"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

type MockAuthService struct {
    mock.Mock
}

func (m *MockAuthService) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
    args := m.Called(ctx, email, password)
    if args.Get(0) == nil {
        return nil, args.Error(1)
    }
    return args.Get(0).(*LoginResponse), args.Error(1)
}

func TestAuthHandler_Login_Success(t *testing.T) {
    // Setup
    gin.SetMode(gin.TestMode)
    mockService := new(MockAuthService)
    handler := NewAuthHandler(mockService, nil) // nil for logger in tests

    // Mock expectations
    expectedResponse := &LoginResponse{Token: "test-token", User: &User{ID: "123"}}
    mockService.On("Login", mock.Anything, "test@example.com", "password123").
        Return(expectedResponse, nil)

    // Create request
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("POST", "/api/v1/auth/login",
        strings.NewReader(`{"email":"test@example.com","password":"password123"}`))
    c.Request.Header.Set("Content-Type", "application/json")

    // Execute
    handler.Login(c)

    // Assert
    assert.Equal(t, http.StatusOK, w.Code)
    mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
    // Test error case
    mockService := new(MockAuthService)
    handler := NewAuthHandler(mockService, nil)

    mockService.On("Login", mock.Anything, "test@example.com", "wrong").
        Return(nil, domain.ErrUnauthorized)

    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)
    c.Request = httptest.NewRequest("POST", "/api/v1/auth/login",
        strings.NewReader(`{"email":"test@example.com","password":"wrong"}`))

    handler.Login(c)

    assert.Equal(t, http.StatusUnauthorized, w.Code)
}
```

**Handlers to Test:**
- `auth_handler.go` - Login, Register, GetCurrentUser
- `task_handler.go` - Create, Update, Delete, GetByID, List, Bump, Complete
- `category_handler.go` - Rename, Delete, GetCategories
- `analytics_handler.go` - GetMetrics, GetVelocity

**Test Coverage Goals:**
- Happy path (200 OK responses)
- Validation errors (400 Bad Request)
- Not found errors (404)
- Unauthorized errors (401)
- Server errors (500)
- Edge cases (empty lists, nil pointers, etc.)

#### 6. Service Tests

**Priority:** CRITICAL

**Target Coverage:** >90% (business logic should be thoroughly tested)

**Test Structure:**
```go
// internal/service/task_service_test.go
package service

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/google/uuid"
)

type MockTaskRepository struct {
    mock.Mock
}

func (m *MockTaskRepository) Create(ctx context.Context, task *domain.Task) error {
    args := m.Called(ctx, task)
    return args.Error(0)
}

func TestTaskService_CreateTask_Success(t *testing.T) {
    // Setup
    mockRepo := new(MockTaskRepository)
    mockPriorityCalc := priority.NewCalculator()
    service := NewTaskService(mockRepo, mockPriorityCalc, nil)

    // Mock expectations
    mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Task")).
        Return(nil)

    // Execute
    req := &CreateTaskRequest{
        Title:        "Test task",
        UserPriority: 5,
    }
    userID := uuid.New()

    task, err := service.CreateTask(context.Background(), userID, req)

    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, task)
    assert.Equal(t, "Test task", task.Title)
    assert.Greater(t, task.PriorityScore, 0) // Priority calculated
    mockRepo.AssertExpectations(t)
}

func TestTaskService_CreateTask_ValidationError(t *testing.T) {
    // Test validation (empty title, etc.)
    service := NewTaskService(nil, nil, nil)

    req := &CreateTaskRequest{
        Title: "", // Invalid - empty title
    }

    task, err := service.CreateTask(context.Background(), uuid.New(), req)

    assert.Error(t, err)
    assert.Nil(t, task)
    assert.IsType(t, &domain.ValidationError{}, err)
}
```

**Services to Test:**
- `auth_service.go` - Register, Login, ValidateToken
- `task_service.go` - CRUD operations, priority updates, bump logic
- `priority_service.go` - Recalculation logic
- `analytics_service.go` - Metrics aggregation

#### 7. Integration Tests

**Priority:** HIGH

**Target:** Test repositories with real PostgreSQL via testcontainers

**Test Structure:**
```go
//go:build integration
// +build integration

package repository_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/testcontainers/testcontainers-go/modules/postgres"
)

func TestTaskRepository_Create_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration test")
    }

    ctx := context.Background()

    // Start PostgreSQL container
    container, err := postgres.Run(ctx,
        "postgres:16-alpine",
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        postgres.WithInitScripts("../../migrations"),
    )
    assert.NoError(t, err)
    defer container.Terminate(ctx)

    // Get connection string
    connStr, err := container.ConnectionString(ctx, "sslmode=disable")
    assert.NoError(t, err)

    // Create repository
    pool, err := pgxpool.New(ctx, connStr)
    assert.NoError(t, err)
    defer pool.Close()

    repo := repository.NewTaskRepository(pool)

    // Test
    task := &domain.Task{
        UserID:       uuid.New(),
        Title:        "Test task",
        UserPriority: 5,
    }

    err = repo.Create(ctx, task)
    assert.NoError(t, err)
    assert.NotEqual(t, uuid.Nil, task.ID)

    // Verify in database
    retrieved, err := repo.GetByID(ctx, task.ID)
    assert.NoError(t, err)
    assert.Equal(t, task.Title, retrieved.Title)
}
```

**Run Integration Tests:**
```bash
# Skip integration tests (fast unit tests only)
go test ./... -short

# Run integration tests only
go test ./... -tags=integration -v

# Run all tests
go test ./...
```

---

## Frontend Testing Setup

### ⏳ Pending

#### 1. Install Dependencies

```bash
cd frontend
npm install -D vitest @testing-library/react @testing-library/jest-dom
npm install -D @testing-library/user-event happy-dom
```

#### 2. Configure Vitest

Create `frontend/vitest.config.ts`:
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

Create `frontend/test/setup.ts`:
```typescript
import '@testing-library/jest-dom';
```

#### 3. Component Tests

**Target:** Critical UI components

Example test for `TaskCard` component:
```typescript
// components/tasks/__tests__/TaskCard.test.tsx
import { render, screen, fireEvent } from '@testing-library/react';
import { TaskCard } from '../TaskCard';
import { describe, it, expect, vi } from 'vitest';

describe('TaskCard', () => {
  const mockTask = {
    id: '123',
    title: 'Test Task',
    user_priority: 5,
    priority_score: 50,
    status: 'todo',
  };

  it('renders task title', () => {
    render(<TaskCard task={mockTask} />);
    expect(screen.getByText('Test Task')).toBeInTheDocument();
  });

  it('calls onComplete when complete button clicked', () => {
    const onComplete = vi.fn();
    render(<TaskCard task={mockTask} onComplete={onComplete} />);

    fireEvent.click(screen.getByText('Complete'));
    expect(onComplete).toHaveBeenCalledWith('123');
  });
});
```

#### 4. Hook Tests

**Target:** Custom React hooks

Example test for `useTasks`:
```typescript
// hooks/__tests__/useTasks.test.ts
import { renderHook, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useTasks } from '../useTasks';
import { vi } from 'vitest';

const createWrapper = () => {
  const queryClient = new QueryClient();
  return ({ children }) => (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
};

describe('useTasks', () => {
  it('fetches tasks successfully', async () => {
    const { result } = renderHook(() => useTasks(), {
      wrapper: createWrapper(),
    });

    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toBeDefined();
  });
});
```

---

## Test Execution

### Backend

```bash
# Run all tests
cd backend && go test ./...

# Run with coverage
go test ./... -cover

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# Run specific package
go test ./internal/handler -v

# Run integration tests only
go test -tags=integration ./... -v

# Skip integration tests (fast)
go test -short ./...
```

### Frontend

```bash
# Run all tests
cd frontend && npm test

# Run in watch mode
npm test -- --watch

# Generate coverage
npm test -- --coverage

# Run specific test file
npm test TaskCard.test.tsx
```

---

## CI/CD Integration

### GitHub Actions Workflow

Create `.github/workflows/test.yml`:

```yaml
name: Tests

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
          go-version: '1.24'

      - name: Run tests
        working-directory: ./backend
        run: |
          go test ./... -v -coverprofile=coverage.out
          go tool cover -func=coverage.out

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: ./backend/coverage.out
          flags: backend

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
        run: npm test -- --coverage

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: ./frontend/coverage/coverage-final.json
          flags: frontend
```

---

## Coverage Goals

### Phase 3 Exit Criteria

- [ ] Overall backend coverage > 80%
- [ ] Handler coverage > 80%
- [ ] Service coverage > 90% (business logic)
- [ ] Repository integration tests cover all CRUD operations
- [ ] Frontend component coverage > 70%
- [ ] All tests passing in CI/CD
- [ ] Coverage reports generated automatically

### Current Progress

| Area | Current | Target | Status |
|------|---------|--------|--------|
| Backend Overall | 8% | 80% | 🔴 |
| Handlers | 0% | 80% | 🔴 |
| Services | 0% | 90% | 🔴 |
| Domain Logic | 100% | 90% | ✅ |
| Repositories | 1% | 70% | 🔴 |
| Frontend | 0% | 70% | 🔴 |

---

## Next Actions

1. ✅ Install backend testing dependencies
2. ⏳ Create service/repository interfaces for DI
3. ⏳ Write handler tests
4. ⏳ Write service tests
5. ⏳ Expand integration tests
6. ⏳ Install frontend testing dependencies
7. ⏳ Configure frontend test environment
8. ⏳ Write component tests
9. ⏳ Configure CI/CD coverage reporting

---

**Document Status:** In Progress
**Next Review:** After completing interface-based DI
