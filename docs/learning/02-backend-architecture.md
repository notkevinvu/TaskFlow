# Module 02: Backend Architecture

## Learning Objectives

By the end of this module, you will:
- Understand Clean Architecture layers and dependency inversion
- Learn how interfaces enable testability and flexibility
- Trace a request from HTTP handler to database and back
- Implement dependency injection patterns

---

## Clean Architecture Overview

TaskFlow's backend follows **Clean Architecture** (also known as Hexagonal Architecture or Ports & Adapters). The key principle is that **dependencies point inward**:

```
┌─────────────────────────────────────────────────────────────┐
│                         Handlers                             │
│                     (HTTP/REST Layer)                        │
│  Parse requests, validate input, format responses            │
├─────────────────────────────────────────────────────────────┤
│                         Services                             │
│                     (Business Logic)                         │
│  Orchestration, validation, priority calculation             │
├─────────────────────────────────────────────────────────────┤
│                          Domain                              │
│                    (Entities & Rules)                        │
│  Pure business entities, no external dependencies            │
├─────────────────────────────────────────────────────────────┤
│                       Repositories                           │
│                      (Data Access)                           │
│  SQL queries via sqlc, database operations                   │
└─────────────────────────────────────────────────────────────┘
                    Dependencies flow UP
              (Inner layers know nothing about outer layers)
```

---

## Layer 1: Domain (The Core)

**Location:** `backend/internal/domain/`

The domain layer contains **pure business entities** with no external dependencies. This is the heart of your application.

### Task Entity

```go
// backend/internal/domain/task.go

package domain

import "time"

// TaskStatus represents the status of a task
type TaskStatus string

const (
    TaskStatusTodo       TaskStatus = "todo"
    TaskStatusInProgress TaskStatus = "in_progress"
    TaskStatusDone       TaskStatus = "done"
    TaskStatusOnHold     TaskStatus = "on_hold"
    TaskStatusBlocked    TaskStatus = "blocked"
)

// TaskEffort represents estimated effort
type TaskEffort string

const (
    TaskEffortSmall  TaskEffort = "small"
    TaskEffortMedium TaskEffort = "medium"
    TaskEffortLarge  TaskEffort = "large"
    TaskEffortXLarge TaskEffort = "xlarge"
)

// Task represents a task entity
type Task struct {
    ID              string      `json:"id"`
    UserID          string      `json:"user_id"`
    Title           string      `json:"title"`
    Description     *string     `json:"description,omitempty"`
    Status          TaskStatus  `json:"status"`
    UserPriority    int         `json:"user_priority"`    // 1-10 user rating
    PriorityScore   int         `json:"priority_score"`   // 0-100 calculated
    EstimatedEffort *TaskEffort `json:"estimated_effort,omitempty"`
    DueDate         *time.Time  `json:"due_date,omitempty"`
    Category        *string     `json:"category,omitempty"`
    BumpCount       int         `json:"bump_count"`
    TaskType        TaskType    `json:"task_type"`
    ParentTaskID    *string     `json:"parent_task_id,omitempty"`
    SeriesID        *string     `json:"series_id,omitempty"`
    CreatedAt       time.Time   `json:"created_at"`
    UpdatedAt       time.Time   `json:"updated_at"`
    CompletedAt     *time.Time  `json:"completed_at,omitempty"`
    DeletedAt       *time.Time  `json:"deleted_at,omitempty"`
}

// CanHaveSubtasks returns true if this task type can have subtasks
func (t *Task) CanHaveSubtasks() bool {
    return t.TaskType == TaskTypeRegular
}

// IsCompleted returns true if the task is done
func (t *Task) IsCompleted() bool {
    return t.Status == TaskStatusDone
}
```

### DTOs (Data Transfer Objects)

DTOs define the **contract** between layers:

```go
// backend/internal/domain/dto.go

// CreateTaskDTO represents the data needed to create a task
type CreateTaskDTO struct {
    Title           string      `json:"title" binding:"required,min=1,max=255"`
    Description     *string     `json:"description,omitempty" binding:"omitempty,max=2000"`
    UserPriority    int         `json:"user_priority" binding:"min=1,max=10"`
    EstimatedEffort *TaskEffort `json:"estimated_effort,omitempty"`
    DueDate         *time.Time  `json:"due_date,omitempty"`
    Category        *string     `json:"category,omitempty" binding:"omitempty,max=50"`
    ParentTaskID    *string     `json:"parent_task_id,omitempty"`
}

// UpdateTaskDTO represents the data for updating a task
type UpdateTaskDTO struct {
    Title           *string     `json:"title,omitempty" binding:"omitempty,min=1,max=255"`
    Description     *string     `json:"description,omitempty" binding:"omitempty,max=2000"`
    Status          *TaskStatus `json:"status,omitempty"`
    UserPriority    *int        `json:"user_priority,omitempty" binding:"omitempty,min=1,max=10"`
    EstimatedEffort *TaskEffort `json:"estimated_effort,omitempty"`
    DueDate         *time.Time  `json:"due_date,omitempty"`
    Category        *string     `json:"category,omitempty" binding:"omitempty,max=50"`
}
```

**Key Insight:** DTOs use pointer types for optional fields. This allows distinguishing between "not provided" (nil) and "set to empty" ("").

---

## Layer 2: Ports (Interfaces)

**Location:** `backend/internal/ports/`

Ports define **contracts** that outer layers must implement. This is the **dependency inversion** in action.

### Repository Interfaces

```go
// backend/internal/ports/repositories.go

package ports

import (
    "context"
    "github.com/notkevinvu/taskflow/backend/internal/domain"
)

// TaskRepository defines the interface for task data access
type TaskRepository interface {
    Create(ctx context.Context, task *domain.Task) error
    FindByID(ctx context.Context, id string) (*domain.Task, error)
    List(ctx context.Context, userID string, filter *domain.TaskListFilter) ([]*domain.Task, error)
    Update(ctx context.Context, task *domain.Task) error
    Delete(ctx context.Context, id, userID string) error
    IncrementBumpCount(ctx context.Context, id, userID string) error

    // Subtask operations
    GetSubtasks(ctx context.Context, parentTaskID string) ([]*domain.Task, error)
    GetSubtasksBatch(ctx context.Context, parentTaskIDs []string) (map[string][]*domain.Task, error)
}

// UserRepository defines the interface for user data access
type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    FindByEmail(ctx context.Context, email string) (*domain.User, error)
    FindByID(ctx context.Context, id string) (*domain.User, error)
    EmailExists(ctx context.Context, email string) (bool, error)
}
```

### Service Interfaces

```go
// backend/internal/ports/services.go

package ports

import (
    "context"
    "github.com/notkevinvu/taskflow/backend/internal/domain"
)

// TaskService defines the interface for task business logic
type TaskService interface {
    Create(ctx context.Context, userID string, dto *domain.CreateTaskDTO) (*domain.Task, error)
    Get(ctx context.Context, userID, taskID string) (*domain.Task, error)
    List(ctx context.Context, userID string, filter *domain.TaskListFilter) ([]*domain.Task, error)
    Update(ctx context.Context, userID, taskID string, dto *domain.UpdateTaskDTO) (*domain.Task, error)
    Delete(ctx context.Context, userID, taskID string) error
    Complete(ctx context.Context, userID, taskID string) (*domain.Task, error)
    Bump(ctx context.Context, userID, taskID string) (*domain.Task, error)
}

// AuthService defines the interface for authentication
type AuthService interface {
    Register(ctx context.Context, dto *domain.CreateUserDTO) (*domain.User, error)
    Login(ctx context.Context, dto *domain.LoginDTO) (string, error)  // Returns JWT token
    GetCurrentUser(ctx context.Context, userID string) (*domain.User, error)
}
```

**Why interfaces matter:**

```go
// Handler depends on INTERFACE, not concrete implementation
type TaskHandler struct {
    taskService ports.TaskService  // Interface!
}

// This enables:
// 1. Testing with mocks
// 2. Swapping implementations (e.g., cache layer)
// 3. Compile-time contract verification
```

---

## Layer 3: Services (Business Logic)

**Location:** `backend/internal/service/`

Services contain **business logic** and orchestrate between repositories.

### Task Service

```go
// backend/internal/service/task_service.go

package service

import (
    "context"
    "github.com/notkevinvu/taskflow/backend/internal/domain"
    "github.com/notkevinvu/taskflow/backend/internal/domain/priority"
    "github.com/notkevinvu/taskflow/backend/internal/ports"
)

type TaskService struct {
    taskRepo          ports.TaskRepository
    taskHistoryRepo   ports.TaskHistoryRepository
    priorityCalc      *priority.Calculator

    // Optional services (setter injection)
    recurrenceService   ports.RecurrenceService
    subtaskService      ports.SubtaskService
    dependencyService   ports.DependencyService
    gamificationService ports.GamificationService
}

func NewTaskService(
    taskRepo ports.TaskRepository,
    taskHistoryRepo ports.TaskHistoryRepository,
) *TaskService {
    return &TaskService{
        taskRepo:        taskRepo,
        taskHistoryRepo: taskHistoryRepo,
        priorityCalc:    priority.NewCalculator(),
    }
}

// Setter injection for optional dependencies
func (s *TaskService) SetGamificationService(gs ports.GamificationService) {
    s.gamificationService = gs
}

func (s *TaskService) Create(ctx context.Context, userID string, dto *domain.CreateTaskDTO) (*domain.Task, error) {
    // 1. Create task entity
    task := &domain.Task{
        ID:              generateUUID(),
        UserID:          userID,
        Title:           dto.Title,
        Description:     dto.Description,
        UserPriority:    dto.UserPriority,
        Status:          domain.TaskStatusTodo,
        EstimatedEffort: dto.EstimatedEffort,
        DueDate:         dto.DueDate,
        Category:        dto.Category,
        TaskType:        domain.TaskTypeRegular,
        CreatedAt:       time.Now(),
        UpdatedAt:       time.Now(),
    }

    // 2. Calculate priority score
    task.PriorityScore = s.priorityCalc.Calculate(task)

    // 3. Validate (business rules)
    if dto.ParentTaskID != nil {
        // Validate parent exists and can have subtasks
        parent, err := s.taskRepo.FindByID(ctx, *dto.ParentTaskID)
        if err != nil {
            return nil, domain.NewNotFoundError("parent task", *dto.ParentTaskID)
        }
        if !parent.CanHaveSubtasks() {
            return nil, domain.NewValidationError("parent_task_id", "parent cannot have subtasks")
        }
        task.ParentTaskID = dto.ParentTaskID
        task.TaskType = domain.TaskTypeSubtask
    }

    // 4. Persist
    if err := s.taskRepo.Create(ctx, task); err != nil {
        return nil, err
    }

    // 5. Create audit log
    s.taskHistoryRepo.Create(ctx, &domain.TaskHistory{
        TaskID:    task.ID,
        EventType: domain.EventTypeCreated,
        NewValue:  task.Title,
    })

    return task, nil
}

func (s *TaskService) Complete(ctx context.Context, userID, taskID string) (*domain.Task, error) {
    // 1. Get task
    task, err := s.taskRepo.FindByID(ctx, taskID)
    if err != nil {
        return nil, err
    }

    // 2. Validate ownership
    if task.UserID != userID {
        return nil, domain.NewUnauthorizedError("cannot complete other user's task")
    }

    // 3. Check dependencies (if dependency service exists)
    if s.dependencyService != nil {
        blocked, err := s.dependencyService.IsBlocked(ctx, taskID)
        if err != nil {
            return nil, err
        }
        if blocked {
            return nil, domain.NewValidationError("status", "task is blocked by incomplete dependencies")
        }
    }

    // 4. Update task
    task.Status = domain.TaskStatusDone
    now := time.Now()
    task.CompletedAt = &now
    task.UpdatedAt = now

    if err := s.taskRepo.Update(ctx, task); err != nil {
        return nil, err
    }

    // 5. Process gamification (async, non-blocking)
    if s.gamificationService != nil {
        s.gamificationService.ProcessTaskCompletionAsync(userID, task)
    }

    // 6. Handle recurring tasks (if applicable)
    if s.recurrenceService != nil && task.SeriesID != nil {
        go s.recurrenceService.CreateNextOccurrence(context.Background(), task)
    }

    return task, nil
}
```

### Key Patterns

**1. Constructor Injection (Required Dependencies):**
```go
func NewTaskService(
    taskRepo ports.TaskRepository,      // Required
    taskHistoryRepo ports.TaskHistoryRepository,  // Required
) *TaskService
```

**2. Setter Injection (Optional Dependencies):**
```go
func (s *TaskService) SetGamificationService(gs ports.GamificationService) {
    s.gamificationService = gs
}

// Usage is conditional
if s.gamificationService != nil {
    s.gamificationService.ProcessTaskCompletionAsync(userID, task)
}
```

**Why two patterns?**
- Constructor injection: Service won't work without these
- Setter injection: Service works fine without these (graceful degradation)

---

## Layer 4: Handlers (HTTP Layer)

**Location:** `backend/internal/handler/`

Handlers are the **adapters** that connect HTTP to business logic.

### Task Handler

```go
// backend/internal/handler/task_handler.go

package handler

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/notkevinvu/taskflow/backend/internal/domain"
    "github.com/notkevinvu/taskflow/backend/internal/middleware"
    "github.com/notkevinvu/taskflow/backend/internal/ports"
)

const MaxLimit = 100  // Pagination limit enforcement

type TaskHandler struct {
    taskService ports.TaskService  // Interface, not concrete!
}

func NewTaskHandler(taskService ports.TaskService) *TaskHandler {
    return &TaskHandler{taskService: taskService}
}

func (h *TaskHandler) Create(c *gin.Context) {
    // 1. Get user ID from context (set by auth middleware)
    userID := c.GetString(middleware.UserIDKey)
    if userID == "" {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    // 2. Parse request body
    var dto domain.CreateTaskDTO
    if err := c.ShouldBindJSON(&dto); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 3. Call service
    task, err := h.taskService.Create(c.Request.Context(), userID, &dto)
    if err != nil {
        handleError(c, err)
        return
    }

    // 4. Return response
    c.JSON(http.StatusCreated, task)
}

func (h *TaskHandler) List(c *gin.Context) {
    userID := c.GetString(middleware.UserIDKey)

    // Parse query parameters with defaults
    filter := &domain.TaskListFilter{
        Limit:  parseIntOrDefault(c.Query("limit"), 50),
        Offset: parseIntOrDefault(c.Query("offset"), 0),
    }

    // Enforce pagination limit (security!)
    if filter.Limit > MaxLimit {
        filter.Limit = MaxLimit
    }

    // Parse optional filters
    if status := c.Query("status"); status != "" {
        s := domain.TaskStatus(status)
        filter.Status = &s
    }
    if category := c.Query("category"); category != "" {
        filter.Category = &category
    }
    if search := c.Query("search"); search != "" {
        filter.Search = &search
    }

    tasks, err := h.taskService.List(c.Request.Context(), userID, filter)
    if err != nil {
        handleError(c, err)
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "tasks":       tasks,
        "total_count": len(tasks),
    })
}

// handleError converts domain errors to HTTP responses
func handleError(c *gin.Context, err error) {
    switch e := err.(type) {
    case *domain.ValidationError:
        c.JSON(http.StatusBadRequest, gin.H{"error": e.Message, "field": e.Field})
    case *domain.NotFoundError:
        c.JSON(http.StatusNotFound, gin.H{"error": e.Error()})
    case *domain.UnauthorizedError:
        c.JSON(http.StatusForbidden, gin.H{"error": e.Error()})
    default:
        c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
    }
}
```

---

## Layer 5: Repositories (Data Access)

**Location:** `backend/internal/repository/`

Repositories implement the `ports.*Repository` interfaces using sqlc.

### Task Repository

```go
// backend/internal/repository/task_repository.go

package repository

import (
    "context"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/notkevinvu/taskflow/backend/internal/domain"
    "github.com/notkevinvu/taskflow/backend/internal/sqlc"
)

type TaskRepository struct {
    db      *pgxpool.Pool
    queries *sqlc.Queries
}

func NewTaskRepository(db *pgxpool.Pool) *TaskRepository {
    return &TaskRepository{
        db:      db,
        queries: sqlc.New(db),
    }
}

func (r *TaskRepository) Create(ctx context.Context, task *domain.Task) error {
    return r.queries.CreateTask(ctx, sqlc.CreateTaskParams{
        ID:              stringToPgtypeUUID(task.ID),
        UserID:          stringToPgtypeUUID(task.UserID),
        Title:           task.Title,
        Description:     stringPtrToPgtypeText(task.Description),
        Status:          sqlc.TaskStatus(task.Status),
        UserPriority:    int32(task.UserPriority),
        PriorityScore:   int32(task.PriorityScore),
        // ... other fields
    })
}

func (r *TaskRepository) FindByID(ctx context.Context, id string) (*domain.Task, error) {
    row, err := r.queries.GetTaskByID(ctx, stringToPgtypeUUID(id))
    if err != nil {
        return nil, convertError(err)
    }
    return rowToTask(row), nil
}

// Batch query - solves N+1 problem
func (r *TaskRepository) GetSubtasksBatch(ctx context.Context, parentTaskIDs []string) (map[string][]*domain.Task, error) {
    result := make(map[string][]*domain.Task)

    // Initialize all keys (even if empty)
    for _, id := range parentTaskIDs {
        result[id] = []*domain.Task{}
    }

    if len(parentTaskIDs) == 0 {
        return result, nil
    }

    // Convert to pgtype UUIDs
    uuids := make([]pgtype.UUID, len(parentTaskIDs))
    for i, id := range parentTaskIDs {
        uuids[i] = stringToPgtypeUUID(id)
    }

    // Single batch query
    rows, err := r.queries.GetSubtasksBatch(ctx, uuids)
    if err != nil {
        return nil, err
    }

    // Group by parent
    for _, row := range rows {
        parentID := pgtypeUUIDToString(row.ParentTaskID)
        task := rowToTask(row)
        result[parentID] = append(result[parentID], task)
    }

    return result, nil
}
```

---

## Wiring It All Together

**Location:** `backend/cmd/server/main.go`

The entry point wires all layers together:

```go
// backend/cmd/server/main.go

package main

import (
    "context"
    "log/slog"
    "os"
    "os/signal"
    "syscall"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5/pgxpool"
)

func main() {
    // 1. Load configuration
    cfg := config.Load()

    // 2. Setup database connection pool
    poolConfig, _ := pgxpool.ParseConfig(cfg.DatabaseURL)
    poolConfig.MaxConns = 25
    poolConfig.MinConns = 5
    dbPool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
    if err != nil {
        slog.Error("Failed to connect to database", "error", err)
        os.Exit(1)
    }
    defer dbPool.Close()

    // 3. Create repositories (data access layer)
    userRepo := repository.NewUserRepository(dbPool)
    taskRepo := repository.NewTaskRepository(dbPool)
    taskHistoryRepo := repository.NewTaskHistoryRepository(dbPool)
    dependencyRepo := repository.NewDependencyRepository(dbPool)
    gamificationRepo := repository.NewGamificationRepository(dbPool)

    // 4. Create services (business logic layer)
    authService := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTExpiry)
    taskService := service.NewTaskService(taskRepo, taskHistoryRepo)
    dependencyService := service.NewDependencyService(dependencyRepo, taskRepo)
    gamificationService := service.NewGamificationService(gamificationRepo)

    // 5. Wire optional dependencies (setter injection)
    taskService.SetDependencyService(dependencyService)
    taskService.SetGamificationService(gamificationService)

    // 6. Create handlers (HTTP layer)
    authHandler := handler.NewAuthHandler(authService)
    taskHandler := handler.NewTaskHandler(taskService)
    dependencyHandler := handler.NewDependencyHandler(dependencyService)

    // 7. Setup router with middleware
    router := gin.New()
    router.Use(gin.Recovery())
    router.Use(middleware.RequestLogger())
    router.Use(middleware.CORS(cfg.AllowedOrigins))
    router.Use(gzip.Gzip(gzip.DefaultCompression))

    // Rate limiting with Redis fallback
    rateLimiterConfig, rateLimiterMiddleware := middleware.RateLimiterWithContext(
        context.Background(), redisClient, cfg.RateLimitRPM,
    )
    router.Use(rateLimiterMiddleware)

    // 8. Register routes
    api := router.Group("/api/v1")
    {
        // Public routes
        api.POST("/auth/register", authHandler.Register)
        api.POST("/auth/login", authHandler.Login)
        api.POST("/auth/guest", authHandler.GuestLogin)

        // Protected routes
        protected := api.Group("")
        protected.Use(middleware.AuthRequired(cfg.JWTSecret))
        {
            protected.GET("/auth/me", authHandler.Me)
            protected.GET("/tasks", taskHandler.List)
            protected.POST("/tasks", taskHandler.Create)
            protected.GET("/tasks/:id", taskHandler.Get)
            protected.PUT("/tasks/:id", taskHandler.Update)
            protected.DELETE("/tasks/:id", taskHandler.Delete)
            protected.POST("/tasks/:id/complete", taskHandler.Complete)
            protected.POST("/tasks/:id/bump", taskHandler.Bump)
        }
    }

    // 9. Graceful shutdown
    server := &http.Server{
        Addr:         ":8080",
        Handler:      router,
        ReadTimeout:  15 * time.Second,
        WriteTimeout: 15 * time.Second,
    }

    go func() {
        if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            slog.Error("Server failed", "error", err)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    // Graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    if rateLimiterConfig != nil {
        rateLimiterConfig.Stop()
    }

    server.Shutdown(ctx)
    slog.Info("Server stopped gracefully")
}
```

---

## Testing with Mocks

The interface-based architecture enables easy testing:

```go
// backend/internal/service/mocks_test.go

type MockTaskRepository struct {
    tasks map[string]*domain.Task
    err   error
}

func (m *MockTaskRepository) Create(ctx context.Context, task *domain.Task) error {
    if m.err != nil {
        return m.err
    }
    m.tasks[task.ID] = task
    return nil
}

func (m *MockTaskRepository) FindByID(ctx context.Context, id string) (*domain.Task, error) {
    if m.err != nil {
        return nil, m.err
    }
    task, ok := m.tasks[id]
    if !ok {
        return nil, domain.NewNotFoundError("task", id)
    }
    return task, nil
}

// Test using mock
func TestTaskService_Create(t *testing.T) {
    mockRepo := &MockTaskRepository{tasks: make(map[string]*domain.Task)}
    mockHistory := &MockTaskHistoryRepository{}

    service := NewTaskService(mockRepo, mockHistory)

    dto := &domain.CreateTaskDTO{
        Title:        "Test Task",
        UserPriority: 5,
    }

    task, err := service.Create(context.Background(), "user-123", dto)

    require.NoError(t, err)
    assert.Equal(t, "Test Task", task.Title)
    assert.NotEmpty(t, task.ID)
    assert.Greater(t, task.PriorityScore, 0)
}
```

---

## Exercises

### 🔰 Beginner: Trace a Request

1. Open `backend/cmd/server/main.go`
2. Find where `TaskHandler` is created
3. Trace which interface it receives
4. Find the concrete implementation in `repository/`

### 🎯 Intermediate: Add a New Endpoint

Design (don't implement) a new endpoint:
- `GET /api/v1/tasks/overdue` - Returns all overdue tasks

1. What method would you add to `ports.TaskService`?
2. What method would you add to `ports.TaskRepository`?
3. What SQL query would you write?

### 🚀 Advanced: Implement a Cache Layer

1. Create a `CachedTaskRepository` that wraps `TaskRepository`
2. Implement `FindByID` with in-memory caching
3. How would you invalidate the cache on updates?

---

## Reflection Questions

1. **Why separate ports and implementation?** What happens if you skip the interface layer?

2. **Why setter injection for optional services?** Could you use constructor injection instead?

3. **Why is the handler layer thin?** What would happen if business logic leaked into handlers?

4. **How does this architecture support testing?** Could you test the service layer without a database?

---

## Key Takeaways

1. **Dependencies point inward.** Handlers depend on services, services depend on domain. Never the reverse.

2. **Interfaces enable flexibility.** You can swap implementations without changing business logic.

3. **Constructor injection for required, setter for optional.** This makes dependencies explicit.

4. **The handler layer is thin.** Parse input, call service, format output. No business logic.

5. **Repositories hide database details.** Services don't know if you're using PostgreSQL or MongoDB.

---

## Next Module

Continue to **[Module 03: Priority Algorithm](./03-priority-algorithm.md)** to see how business logic is implemented in the domain layer.
