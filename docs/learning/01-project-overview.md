# Module 01: Project Overview

## Learning Objectives

By the end of this module, you will:
- Understand the high-level architecture and why each technology was chosen
- Grasp the "modular monolith" philosophy and its benefits for MVP development
- Navigate the project structure confidently
- Identify the core problem being solved and how the solution is architected

---

## The Product Vision

### The Problem

Task management apps are everywhere, but most suffer from a common issue: **users have to manually prioritize tasks**. This leads to:

1. **Decision fatigue** - Constantly re-evaluating what's most important
2. **Forgotten tasks** - Old tasks get buried and never completed
3. **Deadline blindness** - Tasks don't feel urgent until it's too late
4. **Procrastination hiding** - No visibility into repeatedly delayed tasks

### The Solution

TaskFlow automatically calculates task priority using a **multi-factor algorithm**:

```
Score = (UserPriority × 0.4 + TimeDecay × 0.3 + DeadlineUrgency × 0.2 + BumpPenalty × 0.1) × EffortBoost
```

Each factor addresses a specific problem:

| Factor | Weight | Problem Solved |
|--------|--------|----------------|
| **User Priority** | 40% | Respects user intent as the strongest signal |
| **Time Decay** | 30% | Prevents old tasks from being forgotten |
| **Deadline Urgency** | 20% | Creates urgency as deadlines approach |
| **Bump Penalty** | 10% | Exposes procrastinated tasks |
| **Effort Boost** | 1.0-1.3x | Encourages completing small tasks |

**Key Differentiator:** The algorithm is **explainable**. Users can see exactly why each task is ranked where it is, building trust and understanding.

---

## Technology Stack

### The Full Stack

```
┌─────────────────────────────────────────────────────────────┐
│                        Frontend                              │
│  Next.js 16 + React 19 + TypeScript + Tailwind + shadcn/ui │
├─────────────────────────────────────────────────────────────┤
│                      API (HTTP/REST)                         │
├─────────────────────────────────────────────────────────────┤
│                        Backend                               │
│       Go 1.24 + Gin Framework + Clean Architecture          │
├─────────────────────────────────────────────────────────────┤
│                       Database                               │
│              PostgreSQL 16 (Supabase)                       │
└─────────────────────────────────────────────────────────────┘
```

### Why These Technologies?

#### Backend: Go

| Alternative | Why Go Wins Here |
|-------------|------------------|
| **Node.js** | Go has better CPU performance, built-in concurrency, and single binary deployment |
| **Python** | Go's static typing catches bugs at compile time, better performance |
| **Java/Spring** | Go is simpler, faster startup, smaller memory footprint |

**Go-specific benefits:**

```go
// Concurrency is trivial
go handleRequest()   // Runs in separate lightweight goroutine
go processData()     // Can run thousands concurrently

// Single binary deployment
go build -o server
./server  // Just run it - no runtime dependencies!

// Strong typing without ceremony
type Task struct {
    ID       string    `json:"id"`
    Title    string    `json:"title"`
    Priority int       `json:"priority"`
}
```

#### Frontend: Next.js 16

| Alternative | Why Next.js Wins Here |
|-------------|----------------------|
| **Create React App** | Next.js has SSR, file-based routing, better production defaults |
| **Vite + React** | Next.js App Router provides layouts, Server Components, streaming |
| **Vue/Nuxt** | React ecosystem is larger, better TypeScript support |

**Next.js-specific benefits:**

```typescript
// File-based routing - file structure IS the URL structure
app/
  (auth)/
    login/page.tsx      // → /login
    register/page.tsx   // → /register
  (dashboard)/
    dashboard/page.tsx  // → /dashboard
    analytics/page.tsx  // → /analytics

// Layout sharing - common UI wraps child pages
app/(dashboard)/layout.tsx  // Wraps all dashboard pages
```

#### Database: PostgreSQL (Supabase)

| Alternative | Why PostgreSQL Wins Here |
|-------------|-------------------------|
| **MongoDB** | Tasks have complex relationships (dependencies, subtasks) - SQL handles this better |
| **Firebase** | PostgreSQL has better query optimization, complex analytics |
| **MySQL** | PostgreSQL has better JSONB support, array types, window functions |

**PostgreSQL-specific benefits:**

```sql
-- Rich data types
CREATE TYPE task_status AS ENUM ('todo', 'in_progress', 'done', 'blocked');
CREATE TYPE task_effort AS ENUM ('small', 'medium', 'large', 'xlarge');

-- Full-text search built-in
CREATE INDEX idx_tasks_search ON tasks USING GIN(search_vector);
SELECT * FROM tasks WHERE search_vector @@ to_tsquery('important');

-- Complex analytics with window functions
SELECT title, priority_score,
       RANK() OVER (PARTITION BY category ORDER BY priority_score DESC)
FROM tasks;
```

#### State Management: React Query + Zustand

| Concern | Tool | Why |
|---------|------|-----|
| **Server State** (API data) | React Query | Automatic caching, background refetching, optimistic updates |
| **Client State** (UI/auth) | Zustand | Lightweight, simple API, no boilerplate |

**Why this split?**

```typescript
// BAD: Everything in Redux
const state = {
  tasks: [...],           // From API - needs refetching, caching
  user: {...},            // From API - rarely changes
  theme: 'dark',          // Local - never sent to server
  sidebarOpen: false,     // Local - UI state only
}

// GOOD: Separate concerns
// React Query handles server state
const { data: tasks } = useQuery(['tasks'], fetchTasks);

// Zustand handles client state
const theme = useThemeStore(state => state.theme);
const sidebarOpen = useUIStore(state => state.sidebarOpen);
```

---

## Project Structure

### Directory Layout

```
TaskFlow/
├── backend/                    # Go API server (port 8080)
│   ├── cmd/server/             # Application entry point
│   │   └── main.go             # Wires everything together
│   ├── internal/               # Private application code
│   │   ├── domain/             # Business entities & rules
│   │   │   ├── task.go         # Task entity
│   │   │   ├── user.go         # User entity
│   │   │   └── priority/       # Priority algorithm
│   │   ├── ports/              # Interface definitions
│   │   │   ├── repositories.go # Data access contracts
│   │   │   └── services.go     # Business logic contracts
│   │   ├── service/            # Business logic implementation
│   │   ├── repository/         # Database access (sqlc)
│   │   ├── handler/            # HTTP request handlers
│   │   └── middleware/         # Auth, logging, rate limiting
│   ├── migrations/             # Database schema versions
│   └── internal/sqlc/          # Generated type-safe SQL
│
├── frontend/                   # Next.js app (port 3000)
│   ├── app/                    # Next.js App Router
│   │   ├── (auth)/             # Login, register pages
│   │   ├── (dashboard)/        # Main application pages
│   │   ├── layout.tsx          # Root layout (providers)
│   │   └── globals.css         # Design tokens
│   ├── components/             # React components
│   │   ├── ui/                 # shadcn/ui base components
│   │   ├── features/           # Feature-specific components
│   │   └── charts/             # Recharts visualizations
│   ├── hooks/                  # React Query & custom hooks
│   │   ├── useTasks.ts         # Task CRUD operations
│   │   ├── useAuth.ts          # Authentication
│   │   └── useAnalytics.ts     # Analytics data
│   └── lib/                    # Utilities
│       ├── api.ts              # Axios client + error handling
│       └── queryKeys.ts        # React Query key factory
│
└── docs/                       # Documentation
    ├── product/                # PRD, algorithm spec
    ├── architecture/           # System design docs
    ├── implementation/         # Phase plans
    └── learning/               # This curriculum!
```

### Key Insight: Modular Monolith

TaskFlow uses a **modular monolith** architecture:

```
                    ┌────────────────────┐
                    │    TaskFlow App    │
                    └────────────────────┘
                             │
            ┌────────────────┼────────────────┐
            ▼                ▼                ▼
    ┌──────────────┐ ┌──────────────┐ ┌──────────────┐
    │   Backend    │ │   Frontend   │ │   Database   │
    │  (Go + Gin)  │ │  (Next.js)   │ │ (PostgreSQL) │
    └──────────────┘ └──────────────┘ └──────────────┘
         │                │                │
         └────────────────┴────────────────┘
                    Shared database
```

**Why modular monolith over microservices?**

| Microservices | Modular Monolith |
|---------------|------------------|
| Network calls between services | Local function calls |
| Distributed transactions | ACID transactions |
| Complex deployment | Simple deployment |
| Service discovery needed | No discovery needed |
| Team scaling | Individual scaling |

**For an MVP:** Start with a modular monolith. The clean architecture boundaries make it possible to extract microservices later if needed.

---

## Request Flow Example

Let's trace a request from frontend to database:

### Creating a Task

```
1. User clicks "Create Task" button
   │
   ▼
2. Frontend: CreateTaskDialog component
   │  - Collects form data
   │  - Calls useCreateTask() mutation
   │
   ▼
3. Frontend: useTasks.ts hook
   │  - Axios POST to /api/v1/tasks
   │  - Adds JWT token to header
   │
   ▼
4. Backend: middleware stack
   │  - CORS validation
   │  - Rate limiting check
   │  - JWT authentication
   │
   ▼
5. Backend: TaskHandler.Create()
   │  - Parses JSON body
   │  - Validates input
   │
   ▼
6. Backend: TaskService.Create()
   │  - Business validation
   │  - Calculates priority score
   │  - Creates task history entry
   │
   ▼
7. Backend: TaskRepository.Create()
   │  - sqlc-generated SQL
   │  - INSERT INTO tasks
   │
   ▼
8. Database: PostgreSQL
   │  - Stores task
   │  - Updates search_vector trigger
   │
   ▼
9. Response flows back up
   │  - JSON task object
   │  - 201 Created status
   │
   ▼
10. Frontend: React Query
    - Invalidates task list cache
    - UI automatically updates
```

### Code Path References

| Step | File |
|------|------|
| Dialog UI | `frontend/components/features/CreateTaskDialog.tsx` |
| API Hook | `frontend/hooks/useTasks.ts` |
| API Client | `frontend/lib/api.ts` |
| Handler | `backend/internal/handler/task_handler.go` |
| Service | `backend/internal/service/task_service.go` |
| Repository | `backend/internal/repository/task_repository.go` |
| Priority Calc | `backend/internal/domain/priority/calculator.go` |

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────────────┐
│                              FRONTEND                                    │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                         Next.js App                              │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐              │   │
│  │  │  Dashboard  │  │  Analytics  │  │   Archive   │  (Pages)     │   │
│  │  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘              │   │
│  │         │                │                │                      │   │
│  │  ┌──────┴────────────────┴────────────────┴──────┐              │   │
│  │  │                React Query Hooks               │              │   │
│  │  │  useTasks()  useAnalytics()  useGamification() │              │   │
│  │  └──────────────────────┬────────────────────────┘              │   │
│  │                         │                                        │   │
│  │  ┌──────────────────────┴────────────────────────┐              │   │
│  │  │                  API Client                    │              │   │
│  │  │       axios + JWT interceptor + errors         │              │   │
│  │  └──────────────────────┬────────────────────────┘              │   │
│  └─────────────────────────┼─────────────────────────────────────────┘   │
└────────────────────────────┼─────────────────────────────────────────────┘
                             │ HTTP/REST
                             ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                              BACKEND                                     │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                       Middleware Stack                           │   │
│  │  CORS → Logging → Rate Limit → Compression → Auth               │   │
│  └──────────────────────────┬──────────────────────────────────────┘   │
│                              │                                          │
│  ┌──────────────────────────┴──────────────────────────────────────┐   │
│  │                         Handlers                                 │   │
│  │  AuthHandler  TaskHandler  AnalyticsHandler  RecurrenceHandler  │   │
│  └──────────────────────────┬──────────────────────────────────────┘   │
│                              │                                          │
│  ┌──────────────────────────┴──────────────────────────────────────┐   │
│  │                         Services                                 │   │
│  │  AuthService  TaskService  GamificationService  InsightsService │   │
│  └──────────────────────────┬──────────────────────────────────────┘   │
│                              │                                          │
│  ┌──────────────────────────┴──────────────────────────────────────┐   │
│  │                       Repositories                               │   │
│  │  UserRepository  TaskRepository  DependencyRepository  (sqlc)   │   │
│  └──────────────────────────┬──────────────────────────────────────┘   │
└────────────────────────────┼─────────────────────────────────────────────┘
                             │ SQL (pgx)
                             ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                            DATABASE                                      │
│  ┌─────────────────────────────────────────────────────────────────┐   │
│  │                    PostgreSQL 16 (Supabase)                      │   │
│  │                                                                   │   │
│  │  ┌─────────┐ ┌─────────┐ ┌────────────────┐ ┌─────────────────┐ │   │
│  │  │  users  │ │  tasks  │ │ task_history   │ │ task_dependencies│ │   │
│  │  └─────────┘ └─────────┘ └────────────────┘ └─────────────────┘ │   │
│  │                                                                   │   │
│  └─────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Exercises

### 🔰 Beginner: Explore the Project Structure

1. Clone the repository and navigate to it
2. List the contents of `backend/internal/` - what layers do you see?
3. List the contents of `frontend/app/` - how does the URL structure map to folders?
4. Open `backend/cmd/server/main.go` and identify where each layer is wired together

### 🎯 Intermediate: Trace a Request

1. Find the `useTasks()` hook in `frontend/hooks/useTasks.ts`
2. Trace which API endpoint it calls
3. Find the corresponding handler in `backend/internal/handler/`
4. Follow the handler to the service, then to the repository
5. Document the path with file names and key function names

### 🚀 Advanced: Evaluate the Stack

1. Research an alternative for one technology choice (e.g., Rust instead of Go)
2. Write a brief comparison covering:
   - Performance characteristics
   - Developer experience
   - Deployment complexity
   - When you'd choose each option

---

## Reflection Questions

1. **Why start with a monolith?** What signals would indicate it's time to extract a microservice?

2. **Why separate React Query and Zustand?** What problems would you face putting everything in one state store?

3. **Why PostgreSQL over a document database?** Looking at the features (subtasks, dependencies), why does relational make sense?

4. **Why Go over Node.js?** For a task management app, is the performance difference meaningful?

---

## Key Takeaways

1. **Technology choices should be intentional.** Each technology in TaskFlow solves a specific problem.

2. **Start simple, scale later.** The modular monolith provides clean boundaries without microservice complexity.

3. **Separate concerns early.** Server state (React Query) vs. client state (Zustand) prevents a tangled mess.

4. **Type safety everywhere.** TypeScript + Go + sqlc catch bugs at compile time.

---

## Next Module

Continue to **[Module 02: Backend Architecture](./02-backend-architecture.md)** to dive deep into Clean Architecture layers and dependency injection patterns.
