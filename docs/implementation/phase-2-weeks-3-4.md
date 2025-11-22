# Phase 2: Backend Setup with Go + Gin + sqlc (Weeks 3-4)
## Intelligent Task Prioritization System - Backend

**Goal:** Build a Go REST API with task management, priority calculation algorithm, and authentication.

**By the end of this phase, you will have:**
- ✅ Go project with Clean Architecture structure
- ✅ REST API with Gin framework
- ✅ Task database migrations with priority tracking
- ✅ Type-safe database access with sqlc
- ✅ **Priority calculation algorithm** implemented
- ✅ **Bump tracking** for delayed tasks
- ✅ Task CRUD operations with auto-prioritization
- ✅ JWT authentication implemented
- ✅ Frontend connected to live task data

**What you're building:** The backend that powers intelligent task prioritization, automatically calculating priority scores based on user priority, time decay, deadlines, and bump penalties.

---

## Prerequisites

### Install Go

```bash
# macOS
brew install go

# Ubuntu/Linux
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# Windows
# Download installer from https://go.dev/dl/

# Verify installation
go version  # Should show 1.23 or higher
```

### Install Additional Tools

```bash
# golang-migrate
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# sqlc
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# Air (hot reload)
go install github.com/air-verse/air@latest

# Verify installations
migrate -version
sqlc version
air -v
```

---

## Week 3: Backend Foundation

### Day 15: Initialize Go Project

**Step 1: Create Backend Directory**

```bash
# From project root (web-app)
mkdir backend
cd backend

# Initialize Go module
go mod init github.com/yourusername/webapp
```

**Step 2: Create Project Structure**

```bash
mkdir -p cmd/api
mkdir -p internal/{domain,ports,adapters/{handlers,repositories/postgres},services}
mkdir -p db/{migrations,queries,sqlc}
mkdir -p pkg/{logger,middleware}
mkdir -p config
```

**Step 3: Install Dependencies**

```bash
go get github.com/gin-gonic/gin
go get github.com/jackc/pgx/v5/pgxpool
go get github.com/golang-jwt/jwt/v5
go get github.com/joho/godotenv
go get golang.org/x/crypto/bcrypt
go get github.com/go-playground/validator/v10
```

---

### Day 16-17: Database Migrations & sqlc

**Step 1: Create Migration Files**

```bash
# Create users table migration
migrate create -ext sql -dir db/migrations -seq create_users_table

# Create tasks table migration
migrate create -ext sql -dir db/migrations -seq create_tasks_table
```

This creates:
- `db/migrations/000001_create_users_table.up.sql`
- `db/migrations/000001_create_users_table.down.sql`

**Edit `000001_create_users_table.up.sql`:**

```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
```

**Edit `000001_create_users_table.down.sql`:**

```sql
DROP TABLE IF EXISTS users;
```

**Edit `000002_create_tasks_table.up.sql`:**

```sql
CREATE TYPE task_status AS ENUM ('todo', 'in_progress', 'done');
CREATE TYPE task_effort AS ENUM ('small', 'medium', 'large', 'xlarge');

CREATE TABLE tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,

    -- Core task data
    title VARCHAR(200) NOT NULL,
    description TEXT,
    status task_status DEFAULT 'todo' NOT NULL,

    -- Priority inputs
    user_priority INT DEFAULT 50 CHECK (user_priority BETWEEN 0 AND 100),
    due_date TIMESTAMPTZ,
    estimated_effort task_effort,
    category VARCHAR(50),

    -- Context (optional)
    context TEXT,
    related_people TEXT[],

    -- Computed priority
    priority_score DECIMAL(5,2) DEFAULT 50.00 CHECK (priority_score BETWEEN 0 AND 100),

    -- Tracking
    bump_count INT DEFAULT 0,

    -- Metadata
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    completed_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ
);

-- Indexes for performance
CREATE INDEX idx_tasks_user_id ON tasks(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_priority_score ON tasks(priority_score DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_tasks_status ON tasks(status) WHERE deleted_at IS NULL;

-- Composite index for main query (user's priority-sorted tasks)
CREATE INDEX idx_tasks_user_priority
ON tasks(user_id, priority_score DESC, created_at DESC)
WHERE deleted_at IS NULL AND status != 'done';
```

**Edit `000002_create_tasks_table.down.sql`:**

```sql
DROP TABLE IF EXISTS tasks;
DROP TYPE IF EXISTS task_status;
DROP TYPE IF EXISTS task_effort;
```

**Step 2: Run Migrations**

```bash
migrate -path db/migrations -database "postgresql://postgres:postgres@localhost:5432/webapp_dev?sslmode=disable" up
```

**Step 3: Create sqlc Configuration**

Create `sqlc.yaml` in backend root:

```yaml
version: "2"
sql:
  - engine: "postgresql"
    queries: "db/queries"
    schema: "db/migrations"
    gen:
      go:
        package: "sqlc"
        out: "db/sqlc"
        emit_json_tags: true
        emit_prepared_queries: false
        emit_interface: true
        emit_exact_table_names: false
```

**Step 4: Write SQL Queries**

Create `db/queries/users.sql`:

```sql
-- name: GetUser :one
SELECT * FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 LIMIT 1;

-- name: CreateUser :one
INSERT INTO users (
    email, name, password_hash
) VALUES (
    $1, $2, $3
)
RETURNING *;
```

Create `db/queries/tasks.sql`:

```sql
-- name: GetTask :one
SELECT * FROM tasks
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
LIMIT 1;

-- name: ListTasks :many
SELECT * FROM tasks
WHERE user_id = $1 AND deleted_at IS NULL AND status != 'done'
ORDER BY priority_score DESC, created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListCompletedTasks :many
SELECT * FROM tasks
WHERE user_id = $1 AND deleted_at IS NULL AND status = 'done'
ORDER BY completed_at DESC
LIMIT $2 OFFSET $3;

-- name: CreateTask :one
INSERT INTO tasks (
    user_id, title, description, user_priority, due_date,
    estimated_effort, category, context, related_people, priority_score
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: UpdateTask :one
UPDATE tasks
SET
    title = $2,
    description = $3,
    user_priority = $4,
    due_date = $5,
    estimated_effort = $6,
    category = $7,
    context = $8,
    related_people = $9,
    updated_at = NOW()
WHERE id = $1 AND user_id = $10 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateTaskPriority :exec
UPDATE tasks
SET priority_score = $2, updated_at = NOW()
WHERE id = $1;

-- name: BumpTask :one
UPDATE tasks
SET bump_count = bump_count + 1, updated_at = NOW()
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
RETURNING *;

-- name: CompleteTask :one
UPDATE tasks
SET
    status = 'done',
    completed_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteTask :exec
UPDATE tasks
SET deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND user_id = $2;

-- name: GetAtRiskTasks :many
SELECT * FROM tasks
WHERE user_id = $1 AND deleted_at IS NULL AND status != 'done'
  AND bump_count >= 3
ORDER BY priority_score DESC;
```

**Step 5: Generate Go Code**

```bash
sqlc generate
```

This generates `db/sqlc/` directory with type-safe Go code!

---

### Day 18-19: Implement Clean Architecture Layers

**Step 1: Domain Layer**

Create `internal/domain/user.go`:

```go
package domain

import (
    "time"

    "github.com/google/uuid"
)

type User struct {
    ID           uuid.UUID `json:"id"`
    Email        string    `json:"email"`
    Name         string    `json:"name"`
    PasswordHash string    `json:"-"` // Never expose in JSON
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

type CreateUserRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Name     string `json:"name" binding:"required"`
    Password string `json:"password" binding:"required,min=8"`
}

type LoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
    User        *User  `json:"user"`
    AccessToken string `json:"access_token"`
}
```

Create `internal/domain/task.go`:

```go
package domain

import (
    "time"

    "github.com/google/uuid"
)

type TaskStatus string
type TaskEffort string

const (
    TaskStatusTodo       TaskStatus = "todo"
    TaskStatusInProgress TaskStatus = "in_progress"
    TaskStatusDone       TaskStatus = "done"
)

const (
    TaskEffortSmall  TaskEffort = "small"  // < 1 hour
    TaskEffortMedium TaskEffort = "medium" // 1-4 hours
    TaskEffortLarge  TaskEffort = "large"  // 4-8 hours
    TaskEffortXLarge TaskEffort = "xlarge" // > 8 hours
)

type Task struct {
    ID            uuid.UUID   `json:"id"`
    UserID        uuid.UUID   `json:"user_id"`
    Title         string      `json:"title"`
    Description   *string     `json:"description,omitempty"`
    Status        TaskStatus  `json:"status"`
    UserPriority  int32       `json:"user_priority"`
    DueDate       *time.Time  `json:"due_date,omitempty"`
    EstimatedEffort *TaskEffort `json:"estimated_effort,omitempty"`
    Category      *string     `json:"category,omitempty"`
    Context       *string     `json:"context,omitempty"`
    RelatedPeople []string    `json:"related_people,omitempty"`
    PriorityScore float64     `json:"priority_score"`
    BumpCount     int32       `json:"bump_count"`
    CreatedAt     time.Time   `json:"created_at"`
    UpdatedAt     time.Time   `json:"updated_at"`
    CompletedAt   *time.Time  `json:"completed_at,omitempty"`
}

type CreateTaskRequest struct {
    Title           string      `json:"title" binding:"required,max=200"`
    Description     *string     `json:"description,omitempty"`
    UserPriority    *int32      `json:"user_priority,omitempty"`
    DueDate         *time.Time  `json:"due_date,omitempty"`
    EstimatedEffort *TaskEffort `json:"estimated_effort,omitempty"`
    Category        *string     `json:"category,omitempty"`
    Context         *string     `json:"context,omitempty"`
    RelatedPeople   []string    `json:"related_people,omitempty"`
}

type UpdateTaskRequest struct {
    Title           *string     `json:"title,omitempty"`
    Description     *string     `json:"description,omitempty"`
    UserPriority    *int32      `json:"user_priority,omitempty"`
    DueDate         *time.Time  `json:"due_date,omitempty"`
    EstimatedEffort *TaskEffort `json:"estimated_effort,omitempty"`
    Category        *string     `json:"category,omitempty"`
    Context         *string     `json:"context,omitempty"`
    RelatedPeople   []string    `json:"related_people,omitempty"`
}

type BumpTaskRequest struct {
    Reason *string `json:"reason,omitempty"` // Optional: why was this delayed?
}

type TaskListResponse struct {
    Tasks      []*Task `json:"tasks"`
    TotalCount int     `json:"total_count"`
}
```

**Step 2: Ports (Interfaces)**

Create `internal/ports/repositories.go`:

```go
package ports

import (
    "context"

    "github.com/google/uuid"
    "github.com/yourusername/webapp/internal/domain"
)

type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
    GetByEmail(ctx context.Context, email string) (*domain.User, error)
    List(ctx context.Context) ([]*domain.User, error)
    Update(ctx context.Context, user *domain.User) error
    Delete(ctx context.Context, id uuid.UUID) error
}

type TaskRepository interface {
    Create(ctx context.Context, task *domain.Task) error
    GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Task, error)
    List(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*domain.Task, error)
    ListCompleted(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*domain.Task, error)
    Update(ctx context.Context, task *domain.Task) error
    UpdatePriority(ctx context.Context, id uuid.UUID, priorityScore float64) error
    Bump(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Task, error)
    Complete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Task, error)
    Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
    GetAtRisk(ctx context.Context, userID uuid.UUID) ([]*domain.Task, error)
}
```

Create `internal/ports/services.go`:

```go
package ports

import (
    "context"

    "github.com/google/uuid"
    "github.com/yourusername/webapp/internal/domain"
)

type AuthService interface {
    Register(ctx context.Context, req *domain.CreateUserRequest) (*domain.AuthResponse, error)
    Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error)
    ValidateToken(token string) (*domain.User, error)
}

type UserService interface {
    GetByID(ctx context.Context, id string) (*domain.User, error)
    List(ctx context.Context) ([]*domain.User, error)
}

type TaskService interface {
    Create(ctx context.Context, userID uuid.UUID, req *domain.CreateTaskRequest) (*domain.Task, error)
    GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Task, error)
    List(ctx context.Context, userID uuid.UUID, limit, offset int32) (*domain.TaskListResponse, error)
    Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *domain.UpdateTaskRequest) (*domain.Task, error)
    Bump(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *domain.BumpTaskRequest) (*domain.Task, error)
    Complete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Task, error)
    Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
    GetAtRisk(ctx context.Context, userID uuid.UUID) ([]*domain.Task, error)
    RecalculatePriority(ctx context.Context, task *domain.Task) (float64, error)
}

type PriorityService interface {
    CalculatePriority(task *domain.Task) float64
}
```

**Step 3: Repository Implementation**

Create `internal/adapters/repositories/postgres/user_repo.go`:

```go
package postgres

import (
    "context"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/yourusername/webapp/db/sqlc"
    "github.com/yourusername/webapp/internal/domain"
)

type userRepository struct {
    queries *sqlc.Queries
}

func NewUserRepository(pool *pgxpool.Pool) *userRepository {
    return &userRepository{
        queries: sqlc.New(pool),
    }
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
    params := sqlc.CreateUserParams{
        Email:        user.Email,
        Name:         user.Name,
        PasswordHash: user.PasswordHash,
    }

    result, err := r.queries.CreateUser(ctx, params)
    if err != nil {
        return err
    }

    user.ID = result.ID
    user.CreatedAt = result.CreatedAt
    user.UpdatedAt = result.UpdatedAt
    return nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
    result, err := r.queries.GetUserByEmail(ctx, email)
    if err != nil {
        return nil, err
    }

    return &domain.User{
        ID:           result.ID,
        Email:        result.Email,
        Name:         result.Name,
        PasswordHash: result.PasswordHash,
        CreatedAt:    result.CreatedAt,
        UpdatedAt:    result.UpdatedAt,
    }, nil
}

// Implement other methods...
```

Create `internal/adapters/repositories/postgres/task_repo.go`:

```go
package postgres

import (
    "context"
    "database/sql"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/yourusername/webapp/db/sqlc"
    "github.com/yourusername/webapp/internal/domain"
)

type taskRepository struct {
    queries *sqlc.Queries
}

func NewTaskRepository(pool *pgxpool.Pool) *taskRepository {
    return &taskRepository{
        queries: sqlc.New(pool),
    }
}

func (r *taskRepository) Create(ctx context.Context, task *domain.Task) error {
    params := sqlc.CreateTaskParams{
        UserID:          task.UserID,
        Title:           task.Title,
        Description:     toNullString(task.Description),
        UserPriority:    sql.NullInt32{Int32: task.UserPriority, Valid: true},
        DueDate:         toNullTime(task.DueDate),
        EstimatedEffort: toNullTaskEffort(task.EstimatedEffort),
        Category:        toNullString(task.Category),
        Context:         toNullString(task.Context),
        RelatedPeople:   task.RelatedPeople,
        PriorityScore:   sql.NullString{String: fmt.Sprintf("%.2f", task.PriorityScore), Valid: true},
    }

    result, err := r.queries.CreateTask(ctx, params)
    if err != nil {
        return err
    }

    task.ID = result.ID
    task.CreatedAt = result.CreatedAt
    task.UpdatedAt = result.UpdatedAt
    return nil
}

func (r *taskRepository) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Task, error) {
    result, err := r.queries.GetTask(ctx, sqlc.GetTaskParams{
        ID:     id,
        UserID: userID,
    })
    if err != nil {
        return nil, err
    }

    return toTaskDomain(result), nil
}

func (r *taskRepository) List(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*domain.Task, error) {
    results, err := r.queries.ListTasks(ctx, sqlc.ListTasksParams{
        UserID: userID,
        Limit:  limit,
        Offset: offset,
    })
    if err != nil {
        return nil, err
    }

    tasks := make([]*domain.Task, len(results))
    for i, result := range results {
        tasks[i] = toTaskDomain(result)
    }
    return tasks, nil
}

func (r *taskRepository) Bump(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Task, error) {
    result, err := r.queries.BumpTask(ctx, sqlc.BumpTaskParams{
        ID:     id,
        UserID: userID,
    })
    if err != nil {
        return nil, err
    }

    return toTaskDomain(result), nil
}

func (r *taskRepository) Complete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Task, error) {
    result, err := r.queries.CompleteTask(ctx, sqlc.CompleteTaskParams{
        ID:     id,
        UserID: userID,
    })
    if err != nil {
        return nil, err
    }

    return toTaskDomain(result), nil
}

func (r *taskRepository) GetAtRisk(ctx context.Context, userID uuid.UUID) ([]*domain.Task, error) {
    results, err := r.queries.GetAtRiskTasks(ctx, userID)
    if err != nil {
        return nil, err
    }

    tasks := make([]*domain.Task, len(results))
    for i, result := range results {
        tasks[i] = toTaskDomain(result)
    }
    return tasks, nil
}

// Helper conversion functions
func toTaskDomain(t sqlc.Task) *domain.Task {
    task := &domain.Task{
        ID:           t.ID,
        UserID:       t.UserID,
        Title:        t.Title,
        Status:       domain.TaskStatus(t.Status),
        UserPriority: t.UserPriority.Int32,
        BumpCount:    t.BumpCount,
        CreatedAt:    t.CreatedAt,
        UpdatedAt:    t.UpdatedAt,
    }

    if t.Description.Valid {
        task.Description = &t.Description.String
    }
    if t.DueDate.Valid {
        task.DueDate = &t.DueDate.Time
    }
    if t.EstimatedEffort.Valid {
        effort := domain.TaskEffort(t.EstimatedEffort.TaskEffort)
        task.EstimatedEffort = &effort
    }
    if t.Category.Valid {
        task.Category = &t.Category.String
    }
    if t.Context.Valid {
        task.Context = &t.Context.String
    }
    if len(t.RelatedPeople) > 0 {
        task.RelatedPeople = t.RelatedPeople
    }
    if t.PriorityScore.Valid {
        score, _ := strconv.ParseFloat(t.PriorityScore.String, 64)
        task.PriorityScore = score
    }
    if t.CompletedAt.Valid {
        task.CompletedAt = &t.CompletedAt.Time
    }

    return task
}

func toNullString(s *string) sql.NullString {
    if s == nil {
        return sql.NullString{Valid: false}
    }
    return sql.NullString{String: *s, Valid: true}
}

func toNullTime(t *time.Time) sql.NullTime {
    if t == nil {
        return sql.NullTime{Valid: false}
    }
    return sql.NullTime{Time: *t, Valid: true}
}

func toNullTaskEffort(e *domain.TaskEffort) sqlc.NullTaskEffort {
    if e == nil {
        return sqlc.NullTaskEffort{Valid: false}
    }
    return sqlc.NullTaskEffort{TaskEffort: sqlc.TaskEffort(*e), Valid: true}
}

// Implement other methods (Update, Delete, etc.)...
```

**Step 4: Service Implementation**

Create `internal/services/auth_service.go`:

```go
package services

import (
    "context"
    "errors"
    "time"

    "github.com/golang-jwt/jwt/v5"
    "github.com/yourusername/webapp/internal/domain"
    "github.com/yourusername/webapp/internal/ports"
    "golang.org/x/crypto/bcrypt"
)

type authService struct {
    userRepo  ports.UserRepository
    jwtSecret []byte
}

func NewAuthService(userRepo ports.UserRepository, jwtSecret string) ports.AuthService {
    return &authService{
        userRepo:  userRepo,
        jwtSecret: []byte(jwtSecret),
    }
}

func (s *authService) Register(ctx context.Context, req *domain.CreateUserRequest) (*domain.AuthResponse, error) {
    // Check if user exists
    existing, _ := s.userRepo.GetByEmail(ctx, req.Email)
    if existing != nil {
        return nil, errors.New("email already registered")
    }

    // Hash password
    hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }

    // Create user
    user := &domain.User{
        Email:        req.Email,
        Name:         req.Name,
        PasswordHash: string(hash),
    }

    if err := s.userRepo.Create(ctx, user); err != nil {
        return nil, err
    }

    // Generate token
    token, err := s.generateToken(user)
    if err != nil {
        return nil, err
    }

    return &domain.AuthResponse{
        User:        user,
        AccessToken: token,
    }, nil
}

func (s *authService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.AuthResponse, error) {
    // Get user
    user, err := s.userRepo.GetByEmail(ctx, req.Email)
    if err != nil {
        return nil, errors.New("invalid credentials")
    }

    // Verify password
    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
        return nil, errors.New("invalid credentials")
    }

    // Generate token
    token, err := s.generateToken(user)
    if err != nil {
        return nil, err
    }

    return &domain.AuthResponse{
        User:        user,
        AccessToken: token,
    }, nil
}

func (s *authService) generateToken(user *domain.User) (string, error) {
    claims := jwt.MapClaims{
        "user_id": user.ID.String(),
        "email":   user.Email,
        "exp":     time.Now().Add(24 * time.Hour).Unix(),
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(s.jwtSecret)
}

func (s *authService) ValidateToken(tokenString string) (*domain.User, error) {
    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        return s.jwtSecret, nil
    })

    if err != nil || !token.Valid {
        return nil, errors.New("invalid token")
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        return nil, errors.New("invalid token claims")
    }

    userID := claims["user_id"].(string)
    email := claims["email"].(string)

    return &domain.User{
        ID:    uuid.MustParse(userID),
        Email: email,
    }, nil
}
```

Create `internal/services/priority_service.go`:

```go
package services

import (
    "math"
    "time"

    "github.com/yourusername/webapp/internal/domain"
    "github.com/yourusername/webapp/internal/ports"
)

type priorityService struct{}

func NewPriorityService() ports.PriorityService {
    return &priorityService{}
}

// CalculatePriority computes task priority based on multiple factors
// Formula: (UserPriority × 0.4 + TimeDecay × 0.3 + DeadlineUrgency × 0.2 + BumpPenalty × 0.1) × EffortBoost
func (s *priorityService) CalculatePriority(task *domain.Task) float64 {
    userPriority := float64(task.UserPriority)
    timeDecay := s.calculateTimeDecay(task)
    deadlineUrgency := s.calculateDeadlineUrgency(task)
    bumpPenalty := s.calculateBumpPenalty(task)
    effortBoost := s.calculateEffortBoost(task)

    score := (
        userPriority*0.4 +
        timeDecay*0.3 +
        deadlineUrgency*0.2 +
        bumpPenalty*0.1,
    ) * effortBoost

    // Clamp to 0-100
    if score > 100 {
        return 100
    }
    if score < 0 {
        return 0
    }

    return math.Round(score*100) / 100 // Round to 2 decimal places
}

// calculateTimeDecay: Age-based urgency (0-100)
// Increases linearly over 30 days
func (s *priorityService) calculateTimeDecay(task *domain.Task) float64 {
    age := time.Since(task.CreatedAt)
    daysSinceCreation := age.Hours() / 24.0

    // Linear increase over 30 days
    decay := (daysSinceCreation / 30.0) * 100.0

    if decay > 100 {
        return 100
    }
    return decay
}

// calculateDeadlineUrgency: Proximity to due date (0-100)
// Exponential increase in final 3 days
func (s *priorityService) calculateDeadlineUrgency(task *domain.Task) float64 {
    if task.DueDate == nil {
        return 0 // No deadline = no urgency from deadline
    }

    timeUntilDue := time.Until(*task.DueDate)
    hoursUntilDue := timeUntilDue.Hours()
    daysUntilDue := hoursUntilDue / 24.0

    // Overdue tasks get maximum urgency
    if daysUntilDue < 0 {
        return 100
    }

    // Exponential urgency in final 3 days
    if daysUntilDue <= 3 {
        // Maps: 3 days → 50, 1 day → 80, 0 days → 100
        urgency := 100 - (daysUntilDue/3.0)*50
        return urgency
    }

    // Linear urgency from 7 days to 3 days (0 → 50)
    if daysUntilDue <= 7 {
        urgency := ((7 - daysUntilDue) / 4.0) * 50
        return urgency
    }

    // Low urgency for tasks > 7 days away
    return 0
}

// calculateBumpPenalty: Punishment for delays (0-50)
// +10 points per bump, capped at 50
func (s *priorityService) calculateBumpPenalty(task *domain.Task) float64 {
    penalty := float64(task.BumpCount) * 10.0
    if penalty > 50 {
        return 50
    }
    return penalty
}

// calculateEffortBoost: Multiplier for small tasks (1.0-1.3)
// Small tasks get 1.3x, Large tasks get 1.0x
func (s *priorityService) calculateEffortBoost(task *domain.Task) float64 {
    if task.EstimatedEffort == nil {
        return 1.0 // No estimate = no boost
    }

    switch *task.EstimatedEffort {
    case domain.TaskEffortSmall:
        return 1.3 // < 1 hour: 30% boost
    case domain.TaskEffortMedium:
        return 1.15 // 1-4 hours: 15% boost
    case domain.TaskEffortLarge:
        return 1.0 // 4-8 hours: no boost
    case domain.TaskEffortXLarge:
        return 0.95 // > 8 hours: slight penalty
    default:
        return 1.0
    }
}
```

Create `internal/services/task_service.go`:

```go
package services

import (
    "context"
    "errors"

    "github.com/google/uuid"
    "github.com/yourusername/webapp/internal/domain"
    "github.com/yourusername/webapp/internal/ports"
)

type taskService struct {
    taskRepo        ports.TaskRepository
    priorityService ports.PriorityService
}

func NewTaskService(taskRepo ports.TaskRepository, priorityService ports.PriorityService) ports.TaskService {
    return &taskService{
        taskRepo:        taskRepo,
        priorityService: priorityService,
    }
}

func (s *taskService) Create(ctx context.Context, userID uuid.UUID, req *domain.CreateTaskRequest) (*domain.Task, error) {
    // Set defaults
    userPriority := int32(50)
    if req.UserPriority != nil {
        userPriority = *req.UserPriority
    }

    task := &domain.Task{
        UserID:          userID,
        Title:           req.Title,
        Description:     req.Description,
        Status:          domain.TaskStatusTodo,
        UserPriority:    userPriority,
        DueDate:         req.DueDate,
        EstimatedEffort: req.EstimatedEffort,
        Category:        req.Category,
        Context:         req.Context,
        RelatedPeople:   req.RelatedPeople,
        BumpCount:       0,
    }

    // Calculate initial priority
    task.PriorityScore = s.priorityService.CalculatePriority(task)

    // Save to database
    if err := s.taskRepo.Create(ctx, task); err != nil {
        return nil, err
    }

    return task, nil
}

func (s *taskService) GetByID(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Task, error) {
    return s.taskRepo.GetByID(ctx, id, userID)
}

func (s *taskService) List(ctx context.Context, userID uuid.UUID, limit, offset int32) (*domain.TaskListResponse, error) {
    tasks, err := s.taskRepo.List(ctx, userID, limit, offset)
    if err != nil {
        return nil, err
    }

    return &domain.TaskListResponse{
        Tasks:      tasks,
        TotalCount: len(tasks), // TODO: Add count query
    }, nil
}

func (s *taskService) Bump(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *domain.BumpTaskRequest) (*domain.Task, error) {
    // Increment bump count
    task, err := s.taskRepo.Bump(ctx, id, userID)
    if err != nil {
        return nil, err
    }

    // Recalculate priority with new bump count
    newPriority := s.priorityService.CalculatePriority(task)
    if err := s.taskRepo.UpdatePriority(ctx, id, newPriority); err != nil {
        return nil, err
    }

    task.PriorityScore = newPriority
    return task, nil
}

func (s *taskService) Complete(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.Task, error) {
    return s.taskRepo.Complete(ctx, id, userID)
}

func (s *taskService) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
    return s.taskRepo.Delete(ctx, id, userID)
}

func (s *taskService) GetAtRisk(ctx context.Context, userID uuid.UUID) ([]*domain.Task, error) {
    return s.taskRepo.GetAtRisk(ctx, userID)
}

func (s *taskService) RecalculatePriority(ctx context.Context, task *domain.Task) (float64, error) {
    newPriority := s.priorityService.CalculatePriority(task)
    if err := s.taskRepo.UpdatePriority(ctx, task.ID, newPriority); err != nil {
        return 0, err
    }
    return newPriority, nil
}

// Update implementation
func (s *taskService) Update(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *domain.UpdateTaskRequest) (*domain.Task, error) {
    // Get existing task
    task, err := s.taskRepo.GetByID(ctx, id, userID)
    if err != nil {
        return nil, err
    }

    // Update fields
    if req.Title != nil {
        task.Title = *req.Title
    }
    if req.Description != nil {
        task.Description = req.Description
    }
    if req.UserPriority != nil {
        task.UserPriority = *req.UserPriority
    }
    if req.DueDate != nil {
        task.DueDate = req.DueDate
    }
    if req.EstimatedEffort != nil {
        task.EstimatedEffort = req.EstimatedEffort
    }
    if req.Category != nil {
        task.Category = req.Category
    }
    if req.Context != nil {
        task.Context = req.Context
    }
    if req.RelatedPeople != nil {
        task.RelatedPeople = req.RelatedPeople
    }

    // Recalculate priority after updates
    task.PriorityScore = s.priorityService.CalculatePriority(task)

    // Save updates
    if err := s.taskRepo.Update(ctx, task); err != nil {
        return nil, err
    }

    return task, nil
}
```

---

### Day 20-21: HTTP Handlers with Gin

**Step 1: Create Auth Middleware**

Create `pkg/middleware/auth.go`:

```go
package middleware

import (
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"
    "github.com/yourusername/webapp/internal/ports"
)

func Auth(authService ports.AuthService) gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
            c.Abort()
            return
        }

        parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
            c.Abort()
            return
        }

        user, err := authService.ValidateToken(parts[1])
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            c.Abort()
            return
        }

        c.Set("user", user)
        c.Next()
    }
}
```

**Step 2: Create Handlers**

Create `internal/adapters/handlers/auth_handler.go`:

```go
package handlers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/yourusername/webapp/internal/domain"
    "github.com/yourusername/webapp/internal/ports"
)

type AuthHandler struct {
    authService ports.AuthService
}

func NewAuthHandler(authService ports.AuthService) *AuthHandler {
    return &AuthHandler{authService: authService}
}

func (h *AuthHandler) Register(c *gin.Context) {
    var req domain.CreateUserRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    response, err := h.authService.Register(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, response)
}

func (h *AuthHandler) Login(c *gin.Context) {
    var req domain.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    response, err := h.authService.Login(c.Request.Context(), &req)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, response)
}

func (h *AuthHandler) Me(c *gin.Context) {
    user, exists := c.Get("user")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
        return
    }

    c.JSON(http.StatusOK, user)
}
```

Create `internal/adapters/handlers/task_handler.go`:

```go
package handlers

import (
    "net/http"
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/yourusername/webapp/internal/domain"
    "github.com/yourusername/webapp/internal/ports"
)

type TaskHandler struct {
    taskService ports.TaskService
}

func NewTaskHandler(taskService ports.TaskService) *TaskHandler {
    return &TaskHandler{taskService: taskService}
}

// getUserID extracts user ID from context (set by auth middleware)
func getUserID(c *gin.Context) (uuid.UUID, error) {
    userInterface, exists := c.Get("user")
    if !exists {
        return uuid.Nil, errors.New("user not found in context")
    }

    user, ok := userInterface.(*domain.User)
    if !ok {
        return uuid.Nil, errors.New("invalid user type in context")
    }

    return user.ID, nil
}

// CreateTask creates a new task with priority calculation
func (h *TaskHandler) CreateTask(c *gin.Context) {
    userID, err := getUserID(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    var req domain.CreateTaskRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    task, err := h.taskService.Create(c.Request.Context(), userID, &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusCreated, task)
}

// ListTasks returns priority-sorted tasks for the user
func (h *TaskHandler) ListTasks(c *gin.Context) {
    userID, err := getUserID(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    // Parse pagination params
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
    offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

    response, err := h.taskService.List(c.Request.Context(), userID, int32(limit), int32(offset))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, response)
}

// GetTask returns a single task by ID
func (h *TaskHandler) GetTask(c *gin.Context) {
    userID, err := getUserID(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    taskID, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
        return
    }

    task, err := h.taskService.GetByID(c.Request.Context(), taskID, userID)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
        return
    }

    c.JSON(http.StatusOK, task)
}

// UpdateTask updates a task and recalculates priority
func (h *TaskHandler) UpdateTask(c *gin.Context) {
    userID, err := getUserID(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    taskID, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
        return
    }

    var req domain.UpdateTaskRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    task, err := h.taskService.Update(c.Request.Context(), taskID, userID, &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, task)
}

// BumpTask delays a task and adjusts priority
func (h *TaskHandler) BumpTask(c *gin.Context) {
    userID, err := getUserID(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    taskID, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
        return
    }

    var req domain.BumpTaskRequest
    // Reason is optional, so ignore bind errors
    _ = c.ShouldBindJSON(&req)

    task, err := h.taskService.Bump(c.Request.Context(), taskID, userID, &req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "task bumped",
        "task":    task,
    })
}

// CompleteTask marks a task as done
func (h *TaskHandler) CompleteTask(c *gin.Context) {
    userID, err := getUserID(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    taskID, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
        return
    }

    task, err := h.taskService.Complete(c.Request.Context(), taskID, userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, task)
}

// DeleteTask soft-deletes a task
func (h *TaskHandler) DeleteTask(c *gin.Context) {
    userID, err := getUserID(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    taskID, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task ID"})
        return
    }

    if err := h.taskService.Delete(c.Request.Context(), taskID, userID); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "task deleted"})
}

// GetAtRiskTasks returns tasks with high bump counts
func (h *TaskHandler) GetAtRiskTasks(c *gin.Context) {
    userID, err := getUserID(c)
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
        return
    }

    tasks, err := h.taskService.GetAtRisk(c.Request.Context(), userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "tasks": tasks,
        "count": len(tasks),
    })
}
```

**Step 3: Wire Everything Together**

Create `cmd/api/main.go`:

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/joho/godotenv"
    "github.com/yourusername/webapp/internal/adapters/handlers"
    "github.com/yourusername/webapp/internal/adapters/repositories/postgres"
    "github.com/yourusername/webapp/internal/services"
    "github.com/yourusername/webapp/pkg/middleware"
)

func main() {
    // Load environment variables
    godotenv.Load()

    // Connect to database
    dbURL := os.Getenv("DATABASE_URL")
    pool, err := pgxpool.New(context.Background(), dbURL)
    if err != nil {
        log.Fatal("Unable to connect to database:", err)
    }
    defer pool.Close()

    // Initialize repositories
    userRepo := postgres.NewUserRepository(pool)
    taskRepo := postgres.NewTaskRepository(pool)

    // Initialize services
    jwtSecret := os.Getenv("JWT_SECRET")
    authService := services.NewAuthService(userRepo, jwtSecret)
    priorityService := services.NewPriorityService()
    taskService := services.NewTaskService(taskRepo, priorityService)

    // Initialize handlers
    authHandler := handlers.NewAuthHandler(authService)
    taskHandler := handlers.NewTaskHandler(taskService)

    // Setup router
    router := gin.Default()

    // CORS middleware
    router.Use(func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(204)
            return
        }

        c.Next()
    })

    // Public routes
    router.POST("/api/v1/auth/register", authHandler.Register)
    router.POST("/api/v1/auth/login", authHandler.Login)

    // Protected routes
    protected := router.Group("/api/v1")
    protected.Use(middleware.Auth(authService))
    {
        // Auth routes
        protected.GET("/auth/me", authHandler.Me)

        // Task routes
        protected.POST("/tasks", taskHandler.CreateTask)
        protected.GET("/tasks", taskHandler.ListTasks)
        protected.GET("/tasks/at-risk", taskHandler.GetAtRiskTasks)
        protected.GET("/tasks/:id", taskHandler.GetTask)
        protected.PUT("/tasks/:id", taskHandler.UpdateTask)
        protected.POST("/tasks/:id/bump", taskHandler.BumpTask)
        protected.POST("/tasks/:id/complete", taskHandler.CompleteTask)
        protected.DELETE("/tasks/:id", taskHandler.DeleteTask)
    }

    // Health check
    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    // Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    log.Printf("Server starting on port %s", port)
    if err := router.Run(":" + port); err != nil {
        log.Fatal("Server failed to start:", err)
    }
}
```

**Step 4: Create .env File**

Create `backend/.env`:

```bash
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/webapp_dev?sslmode=disable
JWT_SECRET=your-secret-key-change-this-in-production
PORT=8080
```

---

## Week 4: Connect Frontend to Backend

### Day 22-23: Integrate Authentication

**Update Frontend API Client**

Update `frontend/lib/api.ts`:

```typescript
import axios from 'axios';

export const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
  headers: {
    'Content-Type': 'application/json',
  },
});

api.interceptors.request.use((config) => {
  if (typeof window !== 'undefined') {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  }
  return config;
});

export const authAPI = {
  register: (data: { email: string; name: string; password: string }) =>
    api.post('/api/v1/auth/register', data),

  login: (data: { email: string; password: string }) =>
    api.post('/api/v1/auth/login', data),

  me: () => api.get('/api/v1/auth/me'),
};

export interface CreateTaskDTO {
  title: string;
  description?: string;
  user_priority?: number;
  due_date?: string;
  estimated_effort?: 'small' | 'medium' | 'large' | 'xlarge';
  category?: string;
  context?: string;
  related_people?: string[];
}

export interface Task {
  id: string;
  user_id: string;
  title: string;
  description?: string;
  status: 'todo' | 'in_progress' | 'done';
  user_priority: number;
  due_date?: string;
  estimated_effort?: 'small' | 'medium' | 'large' | 'xlarge';
  category?: string;
  context?: string;
  related_people?: string[];
  priority_score: number;
  bump_count: number;
  created_at: string;
  updated_at: string;
  completed_at?: string;
}

export const taskAPI = {
  // Create a new task
  create: (data: CreateTaskDTO) =>
    api.post<Task>('/api/v1/tasks', data),

  // Get all tasks (priority-sorted)
  list: (params?: { limit?: number; offset?: number }) =>
    api.get<{ tasks: Task[]; total_count: number }>('/api/v1/tasks', { params }),

  // Get single task by ID
  getById: (id: string) =>
    api.get<Task>(`/api/v1/tasks/${id}`),

  // Update task
  update: (id: string, data: Partial<CreateTaskDTO>) =>
    api.put<Task>(`/api/v1/tasks/${id}`, data),

  // Bump task (delay it)
  bump: (id: string, reason?: string) =>
    api.post<{ message: string; task: Task }>(`/api/v1/tasks/${id}/bump`, { reason }),

  // Complete task
  complete: (id: string) =>
    api.post<Task>(`/api/v1/tasks/${id}/complete`),

  // Delete task
  delete: (id: string) =>
    api.delete(`/api/v1/tasks/${id}`),

  // Get at-risk tasks (bumped 3+ times)
  getAtRisk: () =>
    api.get<{ tasks: Task[]; count: number }>('/api/v1/tasks/at-risk'),
};
```

**Create Auth Hook**

Create `frontend/hooks/useAuth.ts`:

```typescript
'use client';

import { create } from 'zustand';
import { authAPI } from '@/lib/api';

interface User {
  id: string;
  email: string;
  name: string;
}

interface AuthStore {
  user: User | null;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (email: string, name: string, password: string) => Promise<void>;
  logout: () => void;
}

export const useAuth = create<AuthStore>((set) => ({
  user: null,
  isLoading: false,

  login: async (email, password) => {
    set({ isLoading: true });
    try {
      const response = await authAPI.login({ email, password });
      localStorage.setItem('token', response.data.access_token);
      set({ user: response.data.user, isLoading: false });
    } catch (error) {
      set({ isLoading: false });
      throw error;
    }
  },

  register: async (email, name, password) => {
    set({ isLoading: true });
    try {
      const response = await authAPI.register({ email, name, password });
      localStorage.setItem('token', response.data.access_token);
      set({ user: response.data.user, isLoading: false });
    } catch (error) {
      set({ isLoading: false });
      throw error;
    }
  },

  logout: () => {
    localStorage.removeItem('token');
    set({ user: null });
  },
}));
```

**Update Login Page**

Update `frontend/app/(auth)/login/page.tsx`:

```typescript
'use client';

import { useState } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import Link from "next/link";
import { useAuth } from '@/hooks/useAuth';

export default function LoginPage() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const { login, isLoading } = useAuth();
  const router = useRouter();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    try {
      await login(email, password);
      router.push('/dashboard');
    } catch (err: any) {
      setError(err.response?.data?.error || 'Login failed');
    }
  };

  return (
    <div className="flex min-h-screen items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle>Sign In</CardTitle>
          <CardDescription>
            Enter your email and password to access your account
          </CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <div className="p-3 text-sm text-red-500 bg-red-50 rounded">
                {error}
              </div>
            )}

            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="you@example.com"
                required
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="password">Password</Label>
              <Input
                id="password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                required
              />
            </div>

            <Button type="submit" className="w-full" disabled={isLoading}>
              {isLoading ? 'Signing in...' : 'Sign In'}
            </Button>

            <div className="text-center text-sm">
              Don't have an account?{" "}
              <Link href="/register" className="underline">
                Sign up
              </Link>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
```

---

**Create Task Hooks with React Query**

Create `frontend/hooks/useTasks.ts`:

```typescript
'use client';

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { taskAPI, CreateTaskDTO, Task } from '@/lib/api';
import { toast } from 'sonner';

export function useTasks() {
  return useQuery({
    queryKey: ['tasks'],
    queryFn: async () => {
      const response = await taskAPI.list({ limit: 100, offset: 0 });
      return response.data;
    },
  });
}

export function useCreateTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (data: CreateTaskDTO) => taskAPI.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      toast.success('Task created successfully!');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to create task');
    },
  });
}

export function useBumpTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, reason }: { id: string; reason?: string }) =>
      taskAPI.bump(id, reason),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      toast.info('Task delayed');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to bump task');
    },
  });
}

export function useCompleteTask() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: (id: string) => taskAPI.complete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tasks'] });
      toast.success('Task completed!');
    },
    onError: (error: any) => {
      toast.error(error.response?.data?.error || 'Failed to complete task');
    },
  });
}

export function useAtRiskTasks() {
  return useQuery({
    queryKey: ['tasks', 'at-risk'],
    queryFn: async () => {
      const response = await taskAPI.getAtRisk();
      return response.data;
    },
  });
}
```

**Update Dashboard to Use Real Data**

Update `frontend/app/(dashboard)/dashboard/page.tsx`:

```typescript
'use client';

import { useTasks, useBumpTask, useCompleteTask, useAtRiskTasks } from '@/hooks/useTasks';
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";

export default function DashboardPage() {
  const { data: tasksData, isLoading } = useTasks();
  const { data: atRiskData } = useAtRiskTasks();
  const bumpTask = useBumpTask();
  const completeTask = useCompleteTask();

  if (isLoading) {
    return <div>Loading tasks...</div>;
  }

  const tasks = tasksData?.tasks || [];
  const atRiskCount = atRiskData?.count || 0;
  const quickWins = tasks.filter(
    (t) => t.estimated_effort === 'small' && t.bump_count === 0
  ).length;

  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-3xl font-bold">Today's Priorities</h2>
        <p className="text-muted-foreground">
          Tasks sorted by intelligent priority algorithm
        </p>
      </div>

      {/* Stats */}
      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">At Risk</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{atRiskCount}</div>
            <p className="text-xs text-muted-foreground">
              Tasks bumped 3+ times
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="pb-2">
            <CardTitle className="text-sm font-medium">Quick Wins</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="text-2xl font-bold">{quickWins}</div>
            <p className="text-xs text-muted-foreground">
              Small tasks ready to complete
            </p>
          </CardContent>
        </Card>
      </div>

      {/* Task List */}
      <div className="space-y-4">
        <h3 className="text-lg font-semibold">Your Tasks</h3>
        {tasks.map((task) => (
          <Card key={task.id}>
            <CardContent className="pt-6">
              <div className="flex items-start justify-between">
                <div className="flex-1">
                  <div className="flex items-center gap-2">
                    <h4 className="font-semibold">{task.title}</h4>
                    <Badge
                      variant={
                        task.priority_score >= 90
                          ? "destructive"
                          : task.priority_score >= 75
                          ? "default"
                          : "secondary"
                      }
                    >
                      {Math.round(task.priority_score)}
                    </Badge>
                    {task.bump_count > 0 && (
                      <Badge variant="outline">Bumped {task.bump_count}x</Badge>
                    )}
                  </div>
                  {task.context && (
                    <p className="text-sm text-muted-foreground mt-1">
                      {task.context}
                    </p>
                  )}
                  <div className="flex gap-4 mt-2 text-sm text-muted-foreground">
                    {task.category && <span>Category: {task.category}</span>}
                    {task.due_date && (
                      <span>Due: {new Date(task.due_date).toLocaleDateString()}</span>
                    )}
                  </div>
                </div>
                <div className="flex gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => bumpTask.mutate({ id: task.id })}
                    disabled={bumpTask.isPending}
                  >
                    Bump
                  </Button>
                  <Button
                    size="sm"
                    onClick={() => completeTask.mutate(task.id)}
                    disabled={completeTask.isPending}
                  >
                    Complete
                  </Button>
                </div>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  );
}
```

---

### Day 24-28: Test and Polish

**Test the Full Flow:**

```bash
# Terminal 1: Start database
docker-compose up -d

# Terminal 2: Start backend
cd backend
go run cmd/api/main.go

# Terminal 3: Start frontend
cd frontend
npm run dev
```

**Test Checklist:**
- [ ] Register new user at http://localhost:3000/register
- [ ] Login at http://localhost:3000/login
- [ ] Redirects to dashboard after login
- [ ] **Create a new task** with title, priority, and due date
- [ ] **Tasks appear sorted by priority score**
- [ ] **Bump a task** and see priority increase
- [ ] **Complete a task** and see it move to completed
- [ ] **At-risk indicator** shows tasks with 3+ bumps
- [ ] Priority scores update correctly based on age, deadline, and bumps
- [ ] Token persists in localStorage
- [ ] Logout works correctly

---

## Summary

**You've built:**
- ✅ Go backend with Gin framework
- ✅ Clean Architecture implementation
- ✅ PostgreSQL integration with sqlc
- ✅ **Task management with priority calculation algorithm**
- ✅ **Multi-factor priority scoring** (user priority, time decay, deadline urgency, bump penalty, effort boost)
- ✅ **Bump tracking** for delayed tasks
- ✅ **At-risk task detection** (3+ bumps)
- ✅ JWT authentication
- ✅ REST API endpoints for tasks (create, list, update, bump, complete, delete)
- ✅ Frontend connected to backend with React Query
- ✅ Real-time task prioritization in UI
- ✅ Full authentication flow

**Key Features Implemented:**
1. **Intelligent Priority Algorithm**: Tasks automatically scored based on multiple factors
2. **Bump Tracking**: Tasks that get delayed are penalized in priority calculation
3. **At-Risk Detection**: Tasks bumped 3+ times are flagged
4. **Priority-Sorted List**: Tasks always displayed in order of computed priority
5. **Context Tracking**: Store why tasks were created and who's involved
6. **Effort Estimation**: Small tasks get priority boost

**API Endpoints Available:**
- `POST /api/v1/tasks` - Create task (auto-calculates priority)
- `GET /api/v1/tasks` - List tasks (sorted by priority)
- `GET /api/v1/tasks/:id` - Get single task
- `PUT /api/v1/tasks/:id` - Update task (recalculates priority)
- `POST /api/v1/tasks/:id/bump` - Delay task (increments bump count, adjusts priority)
- `POST /api/v1/tasks/:id/complete` - Mark task as done
- `DELETE /api/v1/tasks/:id` - Soft delete task
- `GET /api/v1/tasks/at-risk` - Get tasks with 3+ bumps

**Next Phase:** Analytics dashboard, background priority recalculation job, and advanced features!

See `phase-3-weeks-5-6.md` for next steps.
