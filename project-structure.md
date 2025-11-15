# Project Structure Guide

This guide explains how to organize your full-stack application for maintainability, scalability, and developer experience.

---

## Table of Contents

- [Overview](#overview)
- [Full Project Structure](#full-project-structure)
- [Backend Structure (Go)](#backend-structure-go)
- [Frontend Structure (Next.js)](#frontend-structure-nextjs)
- [File Naming Conventions](#file-naming-conventions)
- [Module Organization](#module-organization)

---

## Overview

```
web-app/                      # Project root
├── backend/                  # Go API server
├── frontend/                 # Next.js application
├── docker-compose.yml        # Local development environment
├── docker-compose.prod.yml   # Production configuration
├── .gitignore
└── README.md
```

**Why separate backend/frontend directories?**
- Independent deployment (can deploy separately)
- Different languages/toolchains
- Clear boundaries
- Different teams can work independently

---

## Full Project Structure

```
web-app/
│
├── backend/
│   ├── cmd/
│   │   └── api/
│   │       └── main.go                 # Application entry point
│   │
│   ├── internal/                        # Private application code
│   │   ├── domain/                      # Business entities
│   │   │   ├── user.go
│   │   │   ├── analytics.go
│   │   │   └── errors.go
│   │   │
│   │   ├── ports/                       # Interfaces (contracts)
│   │   │   ├── repositories.go
│   │   │   └── services.go
│   │   │
│   │   ├── adapters/                    # Implementations
│   │   │   ├── handlers/               # HTTP handlers
│   │   │   │   ├── user_handler.go
│   │   │   │   ├── analytics_handler.go
│   │   │   │   └── middleware.go
│   │   │   │
│   │   │   └── repositories/           # Data access implementations
│   │   │       └── postgres/
│   │   │           ├── user_repo.go
│   │   │           └── analytics_repo.go
│   │   │
│   │   └── services/                    # Business logic
│   │       ├── user_service.go
│   │       └── analytics_service.go
│   │
│   ├── pkg/                             # Public reusable code (optional)
│   │   ├── logger/
│   │   │   └── logger.go
│   │   ├── validator/
│   │   │   └── validator.go
│   │   └── middleware/
│   │       ├── auth.go
│   │       └── cors.go
│   │
│   ├── db/                              # Database code
│   │   ├── migrations/                  # SQL migrations
│   │   │   ├── 000001_create_users.up.sql
│   │   │   ├── 000001_create_users.down.sql
│   │   │   ├── 000002_create_analytics.up.sql
│   │   │   └── 000002_create_analytics.down.sql
│   │   │
│   │   ├── queries/                     # SQL queries for sqlc
│   │   │   ├── users.sql
│   │   │   └── analytics.sql
│   │   │
│   │   └── sqlc/                        # Generated code (by sqlc)
│   │       ├── db.go
│   │       ├── models.go
│   │       └── queries.sql.go
│   │
│   ├── config/                          # Configuration
│   │   ├── config.go
│   │   └── config.yaml
│   │
│   ├── docker/
│   │   ├── Dockerfile.dev
│   │   └── Dockerfile.prod
│   │
│   ├── scripts/                         # Utility scripts
│   │   ├── seed.go                      # Database seeding
│   │   └── migrate.sh
│   │
│   ├── .air.toml                        # Hot reload configuration
│   ├── sqlc.yaml                        # sqlc configuration
│   ├── go.mod
│   ├── go.sum
│   └── README.md
│
├── frontend/
│   ├── app/                             # Next.js App Router
│   │   ├── (auth)/                      # Route group (doesn't affect URL)
│   │   │   ├── login/
│   │   │   │   └── page.tsx            # /login
│   │   │   └── register/
│   │   │       └── page.tsx            # /register
│   │   │
│   │   ├── dashboard/
│   │   │   ├── layout.tsx               # Dashboard layout
│   │   │   ├── page.tsx                 # /dashboard
│   │   │   ├── analytics/
│   │   │   │   └── page.tsx            # /dashboard/analytics
│   │   │   └── settings/
│   │   │       └── page.tsx            # /dashboard/settings
│   │   │
│   │   ├── api/                         # API routes (optional BFF)
│   │   │   └── auth/
│   │   │       └── route.ts            # /api/auth
│   │   │
│   │   ├── layout.tsx                   # Root layout
│   │   ├── page.tsx                     # / (home page)
│   │   ├── loading.tsx                  # Loading UI
│   │   └── error.tsx                    # Error UI
│   │
│   ├── components/                      # React components
│   │   ├── ui/                          # Shadcn components
│   │   │   ├── button.tsx
│   │   │   ├── card.tsx
│   │   │   ├── dialog.tsx
│   │   │   └── table.tsx
│   │   │
│   │   ├── features/                    # Feature-specific components
│   │   │   ├── auth/
│   │   │   │   ├── LoginForm.tsx
│   │   │   │   └── RegisterForm.tsx
│   │   │   ├── dashboard/
│   │   │   │   ├── MetricsCard.tsx
│   │   │   │   └── RecentActivity.tsx
│   │   │   └── analytics/
│   │   │       ├── Chart.tsx
│   │   │       └── DataTable.tsx
│   │   │
│   │   └── layout/                      # Layout components
│   │       ├── Header.tsx
│   │       ├── Sidebar.tsx
│   │       └── Footer.tsx
│   │
│   ├── lib/                             # Utilities and configurations
│   │   ├── api.ts                       # API client
│   │   ├── utils.ts                     # Utility functions
│   │   ├── constants.ts                 # Constants
│   │   └── cn.ts                        # Tailwind utility
│   │
│   ├── hooks/                           # Custom React hooks
│   │   ├── useAuth.ts
│   │   ├── useAnalytics.ts
│   │   └── useDebounce.ts
│   │
│   ├── types/                           # TypeScript types
│   │   ├── user.ts
│   │   ├── analytics.ts
│   │   └── api.ts
│   │
│   ├── store/                           # Zustand stores
│   │   ├── authStore.ts
│   │   └── uiStore.ts
│   │
│   ├── styles/                          # Global styles
│   │   └── globals.css
│   │
│   ├── public/                          # Static assets
│   │   ├── images/
│   │   └── icons/
│   │
│   ├── docker/
│   │   ├── Dockerfile.dev
│   │   └── Dockerfile.prod
│   │
│   ├── .env.local                       # Environment variables (gitignored)
│   ├── .env.example                     # Example env file
│   ├── next.config.js                   # Next.js configuration
│   ├── tailwind.config.ts               # Tailwind configuration
│   ├── tsconfig.json                    # TypeScript configuration
│   ├── package.json
│   └── README.md
│
├── docker-compose.yml                   # Development environment
├── docker-compose.prod.yml              # Production environment
├── .gitignore
└── README.md
```

---

## Backend Structure (Go)

### `/cmd` - Application Entry Points

```go
// cmd/api/main.go
package main

import (
    "log"
    "github.com/yourusername/project/internal/adapters/handlers"
    "github.com/yourusername/project/config"
)

func main() {
    // Load configuration
    cfg := config.Load()

    // Initialize database
    db := setupDatabase(cfg)

    // Wire dependencies (dependency injection)
    repos := setupRepositories(db)
    services := setupServices(repos)
    handlers := setupHandlers(services)

    // Start server
    router := setupRouter(handlers)
    log.Fatal(router.Run(":8080"))
}
```

**Why `/cmd`?**
- Multiple entry points possible: `cmd/api`, `cmd/worker`, `cmd/migrate`
- Keep main.go small (just wiring)
- Easy to see what binaries are built

---

### `/internal` - Private Application Code

**Why `/internal`?**
- Go enforces: other projects can't import from `/internal`
- Prevents accidental reuse of internal code
- Clear boundary between public and private

#### `internal/domain` - Business Entities

```go
// internal/domain/user.go
package domain

import "time"

// User represents a user in the system
// Pure data structure, no dependencies
type User struct {
    ID        int64
    Email     string
    Name      string
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Business validation
func (u *User) IsValid() error {
    if u.Email == "" {
        return ErrInvalidEmail
    }
    if u.Name == "" {
        return ErrInvalidName
    }
    return nil
}
```

**Characteristics:**
- Pure Go structs
- Business validation
- No external dependencies (no database, no HTTP)
- Core of your application

---

#### `internal/ports` - Interfaces

```go
// internal/ports/repositories.go
package ports

import (
    "context"
    "github.com/yourusername/project/internal/domain"
)

// UserRepository defines data access operations
type UserRepository interface {
    Create(ctx context.Context, user *domain.User) error
    GetByID(ctx context.Context, id int64) (*domain.User, error)
    GetByEmail(ctx context.Context, email string) (*domain.User, error)
    Update(ctx context.Context, user *domain.User) error
    Delete(ctx context.Context, id int64) error
}

// AnalyticsRepository defines analytics data operations
type AnalyticsRepository interface {
    RecordEvent(ctx context.Context, event *domain.Event) error
    GetMetrics(ctx context.Context, params domain.MetricsParams) (*domain.Metrics, error)
}
```

**Why interfaces?**
- Define contracts (what, not how)
- Enable dependency injection
- Easy to mock for testing
- Swap implementations (Postgres → MySQL)

---

#### `internal/adapters` - Implementations

**Handlers (HTTP Layer):**

```go
// internal/adapters/handlers/user_handler.go
package handlers

import (
    "github.com/gin-gonic/gin"
    "github.com/yourusername/project/internal/ports"
)

type UserHandler struct {
    userService ports.UserService
}

func NewUserHandler(userService ports.UserService) *UserHandler {
    return &UserHandler{userService: userService}
}

func (h *UserHandler) Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, ErrorResponse{Error: err.Error()})
        return
    }

    user, err := h.userService.Register(c.Request.Context(), req.Email, req.Password)
    if err != nil {
        c.JSON(400, ErrorResponse{Error: err.Error()})
        return
    }

    c.JSON(201, toUserResponse(user))
}
```

**Repositories (Database Layer):**

```go
// internal/adapters/repositories/postgres/user_repo.go
package postgres

import (
    "context"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/yourusername/project/internal/domain"
    "github.com/yourusername/project/db/sqlc"
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
        Email: user.Email,
        Name:  user.Name,
    }

    result, err := r.queries.CreateUser(ctx, params)
    if err != nil {
        return err
    }

    user.ID = result.ID
    user.CreatedAt = result.CreatedAt
    return nil
}
```

---

#### `internal/services` - Business Logic

```go
// internal/services/user_service.go
package services

import (
    "context"
    "errors"
    "github.com/yourusername/project/internal/domain"
    "github.com/yourusername/project/internal/ports"
    "golang.org/x/crypto/bcrypt"
)

type userService struct {
    userRepo ports.UserRepository
}

func NewUserService(userRepo ports.UserRepository) ports.UserService {
    return &userService{userRepo: userRepo}
}

func (s *userService) Register(ctx context.Context, email, password string) (*domain.User, error) {
    // Business validation
    if !isValidEmail(email) {
        return nil, errors.New("invalid email format")
    }

    if len(password) < 8 {
        return nil, errors.New("password must be at least 8 characters")
    }

    // Check if user exists
    existing, _ := s.userRepo.GetByEmail(ctx, email)
    if existing != nil {
        return nil, errors.New("email already registered")
    }

    // Hash password
    hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    if err != nil {
        return nil, err
    }

    // Create user
    user := &domain.User{
        Email:        email,
        Name:         extractNameFromEmail(email),
        PasswordHash: string(hash),
    }

    err = s.userRepo.Create(ctx, user)
    return user, err
}
```

---

### `/pkg` - Public Reusable Code

```go
// pkg/logger/logger.go
package logger

import "log/slog"

func New() *slog.Logger {
    return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
```

**When to use `/pkg`:**
- Code you want other projects to import
- Utilities that could be open-sourced
- Generic, reusable packages

**When NOT to use `/pkg`:**
- Business logic (use `/internal/services`)
- Application-specific code (use `/internal`)

---

### `/db` - Database Code

```
db/
├── migrations/              # Version-controlled schema changes
│   ├── 000001_init.up.sql
│   └── 000001_init.down.sql
│
├── queries/                 # SQL queries for sqlc
│   └── users.sql
│
└── sqlc/                    # Generated code (gitignored)
    ├── db.go
    ├── models.go
    └── queries.sql.go
```

**Migration naming:**
```
000001_create_users_table.up.sql        # Apply migration
000001_create_users_table.down.sql      # Rollback migration
000002_add_analytics_table.up.sql
000002_add_analytics_table.down.sql
```

---

## Frontend Structure (Next.js)

### `/app` - App Router (Next.js 15)

```
app/
├── layout.tsx               # Root layout (wraps all pages)
├── page.tsx                 # Home page (/)
├── loading.tsx              # Loading UI
├── error.tsx                # Error UI
├── not-found.tsx            # 404 page
│
├── (auth)/                  # Route group (URL not affected)
│   ├── login/
│   │   └── page.tsx         # /login
│   └── register/
│       └── page.tsx         # /register
│
└── dashboard/
    ├── layout.tsx           # Dashboard layout
    ├── page.tsx             # /dashboard
    ├── analytics/
    │   └── page.tsx         # /dashboard/analytics
    └── settings/
        └── page.tsx         # /dashboard/settings
```

**Route Groups `(name)`:**
- Organize routes without affecting URL
- Share layouts
- Example: `(auth)` groups login/register but doesn't add `/auth` to URL

**Special Files:**
- `layout.tsx` - Shared UI that persists across routes
- `page.tsx` - Unique UI for a route
- `loading.tsx` - Loading UI with Suspense
- `error.tsx` - Error boundary

---

### `/components` - React Components

```
components/
├── ui/                      # Shadcn components (generic, reusable)
│   ├── button.tsx
│   ├── card.tsx
│   ├── dialog.tsx
│   └── input.tsx
│
├── features/                # Feature-specific components
│   ├── auth/
│   │   ├── LoginForm.tsx
│   │   └── RegisterForm.tsx
│   ├── dashboard/
│   │   ├── MetricsCard.tsx
│   │   └── RecentActivity.tsx
│   └── analytics/
│       └── Chart.tsx
│
└── layout/                  # Layout components
    ├── Header.tsx
    ├── Sidebar.tsx
    └── Footer.tsx
```

**Organization Strategy:**
- `ui/` - Generic, reusable across entire app
- `features/` - Specific to a feature/domain
- `layout/` - Structural components

---

### `/lib` - Utilities

```typescript
// lib/api.ts - API client
import axios from 'axios';

export const api = axios.create({
  baseURL: process.env.NEXT_PUBLIC_API_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Interceptors for auth tokens
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});
```

```typescript
// lib/utils.ts - Utility functions
import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
```

---

### `/types` - TypeScript Types

```typescript
// types/user.ts
export interface User {
  id: number;
  email: string;
  name: string;
  createdAt: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  name: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface AuthResponse {
  user: User;
  accessToken: string;
}
```

**Organize by domain:**
- `types/user.ts`
- `types/analytics.ts`
- `types/api.ts`

---

### `/hooks` - Custom React Hooks

```typescript
// hooks/useAuth.ts
import { useQuery, useMutation } from '@tanstack/react-query';
import { api } from '@/lib/api';

export function useAuth() {
  return useQuery({
    queryKey: ['auth', 'me'],
    queryFn: () => api.get('/auth/me').then(res => res.data),
  });
}

export function useLogin() {
  return useMutation({
    mutationFn: (credentials: LoginRequest) =>
      api.post('/auth/login', credentials),
  });
}
```

---

## File Naming Conventions

### Backend (Go)

| Type | Convention | Example |
|------|-----------|---------|
| **Files** | snake_case | `user_service.go` |
| **Test files** | `*_test.go` | `user_service_test.go` |
| **Packages** | lowercase, singular | `user`, `service`, `handler` |
| **Interfaces** | PascalCase, verb | `UserRepository`, `Validator` |
| **Structs** | PascalCase | `User`, `CreateUserRequest` |
| **Functions** | PascalCase (exported), camelCase (private) | `CreateUser()`, `validateEmail()` |

```go
// user_service.go
package service

type UserService interface {  // PascalCase interface
    CreateUser()              // PascalCase method
}

type userService struct {}    // camelCase unexported

func (s *userService) CreateUser() {}

func validateEmail() {}       // camelCase private function
```

---

### Frontend (TypeScript/React)

| Type | Convention | Example |
|------|-----------|---------|
| **Components** | PascalCase.tsx | `LoginForm.tsx` |
| **Utilities** | camelCase.ts | `utils.ts`, `api.ts` |
| **Hooks** | use + PascalCase | `useAuth.ts` |
| **Types** | camelCase.ts | `user.ts`, `analytics.ts` |
| **Pages** | lowercase | `page.tsx`, `layout.tsx` |
| **Functions** | camelCase | `fetchUser()`, `validateEmail()` |
| **Types/Interfaces** | PascalCase | `User`, `LoginRequest` |

```typescript
// LoginForm.tsx
export function LoginForm() {}  // PascalCase component

// useAuth.ts
export function useAuth() {}    // use + PascalCase hook

// utils.ts
export function formatDate() {} // camelCase function

// types/user.ts
export interface User {}        // PascalCase interface
```

---

## Module Organization

### Organizing by Feature (Recommended for growth)

```
internal/
├── user/                    # User module
│   ├── domain.go            # User entity
│   ├── repository.go        # Interface + implementation
│   ├── service.go           # Business logic
│   └── handler.go           # HTTP handlers
│
├── analytics/               # Analytics module
│   ├── domain.go
│   ├── repository.go
│   ├── service.go
│   └── handler.go
│
└── auth/                    # Auth module
    ├── domain.go
    ├── service.go
    └── handler.go
```

**Benefits:**
- All user-related code together
- Easy to find everything about a feature
- Clear boundaries for extracting to microservices later
- Team can own entire vertical slice

---

## Summary

### Key Principles

1. **Separation of Concerns**
   - Each layer has a single responsibility
   - Clear boundaries between layers

2. **Dependency Rule**
   - Dependencies point inward
   - Domain has no dependencies
   - Infrastructure depends on domain (not vice versa)

3. **Testability**
   - Easy to mock interfaces
   - Business logic independent of HTTP/DB

4. **Scalability**
   - Modular structure enables future microservices
   - Clear boundaries make extraction easy

5. **Discoverability**
   - Consistent naming
   - Clear folder structure
   - Easy to find things

**Next Steps:**
- Use this structure as your starting point
- Adjust based on your needs
- Keep it simple initially, add complexity as needed
