# Phase 3: Testing, Observability & Production Ready (Weeks 5-6)

**Goal:** Add testing, logging, monitoring, and production-ready features.

**By the end of this phase:**
- ✅ Unit and integration tests
- ✅ Structured logging with slog
- ✅ Error handling patterns
- ✅ Docker multi-stage builds
- ✅ Environment-based configuration
- ✅ Basic observability setup

---

## Week 5: Testing

### Backend Testing

**Install Testing Dependencies:**

```bash
cd backend
go get github.com/stretchr/testify
go get github.com/testcontainers/testcontainers-go
go get github.com/testcontainers/testcontainers-go/modules/postgres
```

**Unit Test Example (`internal/services/auth_service_test.go`):**

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

## Week 6: Observability & Production

### Structured Logging

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

**Use in Handlers:**

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

## Summary

**You've added:**
- ✅ Unit and integration tests
- ✅ Structured logging
- ✅ Error handling
- ✅ Production Docker builds
- ✅ Health checks

**Next:** Advanced features and scaling (Phase 4)!
