# Technology Stack Explained

This document provides in-depth explanations of each technology in the stack, what problems they solve, and why they were chosen.

---

## Table of Contents

- [Frontend Technologies](#frontend-technologies)
- [Backend Technologies](#backend-technologies)
- [Database Technologies](#database-technologies)
- [Infrastructure & DevOps](#infrastructure--devops)

---

## Frontend Technologies

### Next.js 15

**What is it?**
Next.js is a React framework built by Vercel that adds powerful features on top of React for building production-ready web applications.

**What problems does it solve?**

1. **SEO Issues with Plain React**
   - React is client-side rendered (blank HTML until JS loads)
   - Next.js does Server-Side Rendering (SSR) - sends full HTML
   - Better for search engines and social media previews

2. **Routing Complexity**
   - React requires react-router setup and configuration
   - Next.js uses file-based routing - create `app/about/page.tsx` → `/about` route exists

3. **Performance**
   - Automatic code splitting
   - Image optimization
   - Font optimization
   - Static Site Generation (SSG) for fast pages

4. **Backend Integration**
   - Built-in API routes: `app/api/users/route.ts` → `/api/users` endpoint
   - Server Components can fetch data directly

**Example: File-Based Routing**
```
app/
  ├── page.tsx                    → /
  ├── about/
  │   └── page.tsx                → /about
  ├── dashboard/
  │   ├── page.tsx                → /dashboard
  │   ├── analytics/
  │   │   └── page.tsx            → /dashboard/analytics
  │   └── layout.tsx              → Shared layout for all /dashboard/* pages
  └── api/
      └── users/
          └── route.ts            → /api/users endpoint
```

**When to use plain React instead:**
- Internal tools where SEO doesn't matter
- Embedded widgets
- When you already have a complete separate backend

---

### TypeScript

**What is it?**
TypeScript is JavaScript with type annotations. It compiles to regular JavaScript but catches errors during development.

**What problems does it solve?**

```javascript
// JavaScript - no error until runtime
function calculateTotal(price, quantity) {
  return price * quantity;
}
calculateTotal("10", 5);  // Returns "1010101010" (string concatenation!)
```

```typescript
// TypeScript - error at compile time
function calculateTotal(price: number, quantity: number): number {
  return price * quantity;
}
calculateTotal("10", 5);  // ❌ Error: Argument of type 'string' is not assignable to parameter of type 'number'
```

**Key Benefits:**
1. **Catch bugs before runtime** - type errors show in your IDE
2. **Better autocomplete** - IDE knows what properties/methods exist
3. **Self-documenting code** - types show what data looks like
4. **Safer refactoring** - renaming/changing types updates everywhere
5. **Better collaboration** - team members know what data to pass

**Example: Type-Safe API Call**
```typescript
// Define the shape of your data
interface User {
  id: number;
  name: string;
  email: string;
  createdAt: Date;
}

// Function signature ensures type safety
async function getUser(id: number): Promise<User> {
  const response = await fetch(`/api/users/${id}`);
  return response.json();
}

// Now you have autocomplete and type checking
const user = await getUser(123);
console.log(user.name);    // ✓ IDE knows 'name' exists
console.log(user.age);     // ❌ Error: Property 'age' does not exist on type 'User'
```

---

### React Query (TanStack Query)

**What is it?**
A library for fetching, caching, and synchronizing server data in React applications.

**What problems does it solve?**

**Without React Query:**
```typescript
// You have to manage all this manually
function UserList() {
  const [users, setUsers] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  useEffect(() => {
    setLoading(true);
    fetch('/api/users')
      .then(res => res.json())
      .then(data => {
        setUsers(data);
        setLoading(false);
      })
      .catch(err => {
        setError(err);
        setLoading(false);
      });
  }, []);

  // Need to refetch when? How often? Manual refresh?
  // Cache where? How long? When to invalidate?

  if (loading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;
  return <div>{/* render users */}</div>;
}
```

**With React Query:**
```typescript
function UserList() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['users'],
    queryFn: () => fetch('/api/users').then(res => res.json()),
    staleTime: 5000,       // Data fresh for 5 seconds
    refetchInterval: 30000, // Auto-refetch every 30 seconds
  });

  // React Query handles loading, error, caching, refetching automatically
  if (isLoading) return <div>Loading...</div>;
  if (error) return <div>Error: {error.message}</div>;
  return <div>{/* render data */}</div>;
}
```

**Key Features:**
- **Automatic caching** - same query won't refetch unnecessarily
- **Background refetching** - keeps data fresh
- **Optimistic updates** - UI updates before server confirms
- **Pagination/infinite scroll** - built-in support
- **Perfect for dashboards** - always show fresh data

---

### Zustand

**What is it?**
A small, fast state management library for React - a simpler alternative to Redux.

**What problems does it solve?**

**Redux (complex):**
```typescript
// Redux requires: actions, reducers, store, providers, connect/useSelector
// Lots of boilerplate for simple state

// 1. Define action types
const INCREMENT = 'INCREMENT';

// 2. Define actions
const increment = () => ({ type: INCREMENT });

// 3. Define reducer
function counterReducer(state = 0, action) {
  switch (action.type) {
    case INCREMENT: return state + 1;
    default: return state;
  }
}

// 4. Create store
const store = createStore(counterReducer);

// 5. Wrap app in Provider
<Provider store={store}>
  <App />
</Provider>

// 6. Use in component
const count = useSelector(state => state);
const dispatch = useDispatch();
dispatch(increment());
```

**Zustand (simple):**
```typescript
// 1. Create store (that's it!)
const useStore = create((set) => ({
  count: 0,
  increment: () => set((state) => ({ count: state.count + 1 })),
}));

// 2. Use in component
function Counter() {
  const { count, increment } = useStore();
  return <button onClick={increment}>{count}</button>;
}
```

**When to use:**
- Theme preferences (dark/light mode)
- User session info
- UI state (modal open/closed, sidebar expanded)
- Shopping cart
- Any global state that's not server data (use React Query for that)

**Why not Redux?**
- Redux is overkill for most apps
- 90% less code with Zustand
- Easier to learn
- No Provider wrapper needed
- Better TypeScript support

**When to use Redux:**
- Very large teams (>20 developers)
- Extremely complex state logic
- Need time-travel debugging
- Legacy codebase already using Redux

---

### Shadcn/UI

**What is it?**
A collection of re-usable components that you copy into your project (not an npm package).

**What problems does it solve?**

**Traditional Component Libraries (MUI, Chakra):**
```bash
npm install @mui/material  # Adds 1MB+ to your bundle
```
- Large bundle size (include everything)
- Limited customization (styles are bundled)
- Dependency on package updates
- Can't modify component internals easily

**Shadcn/UI:**
```bash
npx shadcn@latest add button  # Copies button.tsx to your project
```
- Only include components you use
- Full source code in your project
- Complete customization (it's your code now)
- No dependency (can modify freely)
- Built on Radix UI (accessible) + Tailwind CSS (styling)

**Example: Adding a Button**
```bash
npx shadcn@latest add button
```

Creates `components/ui/button.tsx` in your project:
```typescript
// You now own this code - modify as needed
import { cn } from "@/lib/utils"

const Button = ({ className, ...props }) => {
  return (
    <button
      className={cn(
        "inline-flex items-center justify-center rounded-md bg-primary text-primary-foreground",
        className
      )}
      {...props}
    />
  )
}
```

---

## Backend Technologies

### Go (Golang)

**What is it?**
Go is a compiled, statically-typed programming language developed by Google.

**What problems does it solve?**

1. **Concurrency** (handling many requests at once)
   ```go
   // Goroutines make concurrent programming easy
   go handleRequest()   // Runs in separate lightweight thread
   go processData()     // Can run thousands concurrently
   ```

2. **Performance** (fast startup, low memory)
   - Compiles to single binary (no runtime dependencies)
   - Fast execution (compiled, not interpreted)
   - Low memory footprint

3. **Simplicity** (easier than Java/C++, safer than JavaScript)
   - Only 25 keywords (Java has 50+)
   - No classes, inheritance, generics complexity
   - Built-in testing, formatting, documentation

**Why Go for backend?**
- **Speed**: Compiles to native code, fast as C/C++
- **Concurrency**: Built-in goroutines and channels
- **Deployment**: Single binary, no dependencies (copy and run)
- **Standard Library**: HTTP server, JSON, crypto all included
- **Used by**: Google, Uber, Netflix, Dropbox, Docker, Kubernetes

**Example: Simple HTTP Server**
```go
package main

import (
    "encoding/json"
    "net/http"
)

type Response struct {
    Message string `json:"message"`
}

func main() {
    http.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(Response{Message: "Hello World"})
    })

    http.ListenAndServe(":8080", nil)
}

// Compile: go build -o server
// Run: ./server
// Single binary, no Node.js/Python runtime needed!
```

---

### Gin Web Framework

**What is it?**
Gin is a web framework for Go that makes building HTTP servers and REST APIs easier.

**What problems does it solve?**

**Without Gin (using standard library):**
```go
http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
    // 1. Manual method checking
    if r.Method != "GET" {
        http.Error(w, "Method not allowed", 405)
        return
    }

    // 2. Manual query param parsing
    id := r.URL.Query().Get("id")
    if id == "" {
        http.Error(w, "Missing id", 400)
        return
    }

    // 3. Manual JSON encoding
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
})
```

**With Gin:**
```go
router := gin.Default()

router.GET("/users/:id", func(c *gin.Context) {
    id := c.Param("id")  // Automatic param parsing

    var user User
    // Automatic JSON binding and validation
    if err := c.ShouldBindJSON(&user); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    c.JSON(200, user)  // Automatic JSON encoding
})
```

**Key Features:**

1. **Routing**
   ```go
   router.GET("/users", getUsers)
   router.POST("/users", createUser)
   router.PUT("/users/:id", updateUser)
   router.DELETE("/users/:id", deleteUser)

   // Route groups
   api := router.Group("/api/v1")
   {
       api.GET("/users", getUsers)
       api.GET("/analytics", getAnalytics)
   }
   ```

2. **Middleware** (auth, logging, CORS)
   ```go
   router.Use(Logger())
   router.Use(AuthMiddleware())
   router.Use(CORS())
   ```

3. **Validation**
   ```go
   type CreateUserRequest struct {
       Email    string `json:"email" binding:"required,email"`
       Password string `json:"password" binding:"required,min=8"`
   }

   var req CreateUserRequest
   if err := c.ShouldBindJSON(&req); err != nil {
       c.JSON(400, gin.H{"error": "Invalid request"})
       return
   }
   ```

**Alternatives:**

| Framework | Use Case |
|-----------|----------|
| **Gin** | General purpose, best for most apps |
| **Echo** | Enterprise apps, better routing organization |
| **Fiber** | Extreme performance needs, Express.js-like syntax |
| **Chi** | Minimal, close to stdlib, good for learning |
| **Stdlib** | Maximum control, educational |

---

### sqlc

**What is it?**
A tool that generates type-safe Go code from SQL queries.

**What problems does it solve?**

**Problem: Type safety with databases**

```go
// Traditional approach - no type safety
rows, err := db.Query("SELECT id, name, email FROM users WHERE id = $1", userID)

var id int
var name, email string
err = rows.Scan(&id, &name, &email)
// ❌ What if you change column order in SELECT?
// ❌ What if you scan in wrong order?
// ❌ What if column type changes?
// ❌ All errors happen at runtime!
```

**Solution: sqlc generates type-safe code**

**Step 1: Write SQL**
```sql
-- queries/users.sql

-- name: GetUser :one
SELECT id, name, email, created_at
FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT id, name, email, created_at
FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CreateUser :one
INSERT INTO users (name, email, password_hash)
VALUES ($1, $2, $3)
RETURNING id, name, email, created_at;
```

**Step 2: Run sqlc generate**
```bash
sqlc generate
```

**Step 3: Use generated type-safe code**
```go
// Generated structs (in db/models.go)
type User struct {
    ID        int64
    Name      string
    Email     string
    CreatedAt time.Time
}

// Generated functions (in db/queries.sql.go)
func (q *Queries) GetUser(ctx context.Context, id int64) (User, error) {
    // Implementation auto-generated
}

func (q *Queries) ListUsers(ctx context.Context, arg ListUsersParams) ([]User, error) {
    // Implementation auto-generated
}

// Usage in your code
user, err := queries.GetUser(ctx, 123)
if err != nil {
    return err
}
fmt.Println(user.Name)  // ✓ Type-safe! IDE autocomplete works!
```

**Benefits:**

1. **Compile-time type safety**
   - Change SQL → regenerate → Go code updates
   - Refactoring is safe
   - IDE catches errors before running

2. **Performance**
   - Zero runtime overhead
   - Just generates efficient Go code
   - As fast as hand-written code

3. **Learn SQL properly**
   - You write real SQL
   - No ORM magic hiding queries
   - Portable knowledge

4. **PostgreSQL-specific features**
   - Arrays, JSONB, enums all supported
   - Can use advanced SQL features

**Comparison with alternatives:**

**GORM (ORM):**
```go
// GORM - more automatic, but slower
db.Where("email = ?", email).First(&user)
// ❌ What SQL does this generate?
// ❌ Slower (reflection, runtime query building)
// ✓ Easier for complex associations
```

**sqlx (lightweight):**
```go
// sqlx - manual, less type-safe
err := db.Get(&user, "SELECT * FROM users WHERE id = $1", id)
// ✓ Lightweight
// ❌ No compile-time validation
// ❌ Easy to mess up struct tags
```

**When to use each:**
- **sqlc**: Best for most apps (recommended)
- **GORM**: Need complex associations, migrations, hooks
- **sqlx**: Want minimal abstraction, understand SQL well

---

### Clean/Hexagonal Architecture

**What is it?**
An architectural pattern that separates business logic from external dependencies (database, HTTP, etc.).

**What problems does it solve?**

**Problem: Tightly coupled code**
```go
// ❌ Handler directly uses database
func CreateUser(c *gin.Context) {
    var req CreateUserRequest
    c.BindJSON(&req)

    // Handler knows about database - hard to test!
    result := db.Exec("INSERT INTO users...")

    c.JSON(201, result)
}
```

**Solution: Clean Architecture with layers**

```
┌─────────────────────────────────────────┐
│         Handlers (HTTP/gRPC)            │  ← Delivery layer
│  - Parse requests                       │  ← Gin handlers
│  - Call services                        │  ← Convert HTTP ↔ Domain
│  - Format responses                     │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│         Services (Business Logic)       │  ← Application layer
│  - Validation                           │  ← Pure business rules
│  - Orchestration                        │  ← No HTTP/DB knowledge
│  - Business rules                       │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│         Repositories (Data Access)      │  ← Infrastructure layer
│  - Database queries                     │  ← PostgreSQL
│  - External APIs                        │  ← Can be swapped
│  - Caching                              │
└─────────────────────────────────────────┘
```

**Example:**

```go
// domain/user.go - Pure business entities
type User struct {
    ID        int64
    Email     string
    Name      string
    CreatedAt time.Time
}

// ports/repositories.go - Interfaces (contracts)
type UserRepository interface {
    Create(ctx context.Context, user *User) error
    GetByID(ctx context.Context, id int64) (*User, error)
    GetByEmail(ctx context.Context, email string) (*User, error)
}

// ports/services.go
type UserService interface {
    Register(ctx context.Context, email, password string) (*User, error)
}

// services/user_service.go - Business logic
type userService struct {
    repo UserRepository
}

func (s *userService) Register(ctx context.Context, email, password string) (*User, error) {
    // Business validation
    if !isValidEmail(email) {
        return nil, errors.New("invalid email")
    }

    // Check if exists
    existing, _ := s.repo.GetByEmail(ctx, email)
    if existing != nil {
        return nil, errors.New("email already exists")
    }

    // Create user
    user := &User{Email: email, Name: extractName(email)}
    return user, s.repo.Create(ctx, user)
}

// adapters/repositories/postgres/user_repo.go - Database implementation
type postgresUserRepo struct {
    queries *db.Queries
}

func (r *postgresUserRepo) Create(ctx context.Context, user *User) error {
    // Use sqlc generated code
    return r.queries.CreateUser(ctx, db.CreateUserParams{
        Email: user.Email,
        Name:  user.Name,
    })
}

// adapters/handlers/user_handler.go - HTTP layer
type UserHandler struct {
    service ports.UserService
}

func (h *UserHandler) Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    user, err := h.service.Register(c.Request.Context(), req.Email, req.Password)
    if err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }

    c.JSON(201, user)
}
```

**Benefits:**

1. **Testability**
   ```go
   // Mock repository for testing
   type mockUserRepo struct{}
   func (m *mockUserRepo) Create(...) error { return nil }

   // Test service without database
   service := NewUserService(&mockUserRepo{})
   ```

2. **Flexibility**
   - Swap PostgreSQL → MySQL without changing business logic
   - Swap Gin → Echo without changing services
   - Add gRPC alongside REST

3. **Microservices evolution**
   - Extract service into separate microservice easily
   - Clean boundaries already exist

---

## Database Technologies

### PostgreSQL 16

**What is it?**
PostgreSQL is a powerful open-source relational database.

**What problems does it solve?**

**Why PostgreSQL over others:**

| Feature | PostgreSQL | MySQL | MongoDB |
|---------|-----------|-------|---------|
| **SQL Support** | ✅ Full SQL standard | ⚠️ Partial | ❌ NoSQL |
| **JSON/JSONB** | ✅ Native, indexed | ⚠️ Basic | ✅ Native |
| **Complex Queries** | ✅ Excellent | ✅ Good | ⚠️ Limited |
| **ACID Compliance** | ✅ Full | ✅ Full | ⚠️ Eventual consistency |
| **Analytics** | ✅ Excellent (window functions) | ✅ Good | ⚠️ Limited |
| **Extensions** | ✅ Rich (PostGIS, TimescaleDB) | ⚠️ Limited | ❌ None |
| **Best For** | General purpose, analytics | Simple CRUD | Document storage |

**Key Features for Your App:**

1. **JSONB for flexible data**
   ```sql
   CREATE TABLE analytics_events (
       id SERIAL PRIMARY KEY,
       event_type VARCHAR(50),
       event_data JSONB  -- Store any JSON, query with SQL!
   );

   -- Query JSON fields
   SELECT *
   FROM analytics_events
   WHERE event_data->>'user_id' = '123'
     AND event_data->>'action' = 'click';

   -- Index JSON fields
   CREATE INDEX idx_event_user
   ON analytics_events ((event_data->>'user_id'));
   ```

2. **Partitioning for time-series data**
   ```sql
   CREATE TABLE events (
       id BIGSERIAL,
       created_at TIMESTAMPTZ,
       data JSONB,
       PRIMARY KEY (id, created_at)
   ) PARTITION BY RANGE (created_at);

   -- Automatically route queries to correct partition
   ```

3. **Window functions for analytics**
   ```sql
   -- Running totals, moving averages
   SELECT
       date,
       revenue,
       SUM(revenue) OVER (ORDER BY date) as running_total,
       AVG(revenue) OVER (ORDER BY date ROWS BETWEEN 6 PRECEDING AND CURRENT ROW) as week_avg
   FROM daily_sales;
   ```

4. **Materialized views for dashboards**
   ```sql
   CREATE MATERIALIZED VIEW daily_metrics AS
   SELECT
       DATE(created_at) as date,
       COUNT(*) as event_count,
       COUNT(DISTINCT user_id) as unique_users
   FROM events
   GROUP BY DATE(created_at);

   REFRESH MATERIALIZED VIEW daily_metrics;  -- Update periodically
   ```

---

### pgx/pgxpool

**What is it?**
pgx is a pure Go PostgreSQL driver. pgxpool provides connection pooling.

**Why use it over database/sql?**

```go
// database/sql (standard library)
// ✓ Works with any SQL database
// ❌ Slower (generic interface)
// ❌ No PostgreSQL-specific features

// pgx/pgxpool
// ✓ 30-50% faster for PostgreSQL
// ✓ Native support for arrays, JSONB, COPY
// ✓ Better connection pooling
// ✓ Context support built-in
```

**Example:**
```go
import (
    "github.com/jackc/pgx/v5/pgxpool"
)

// Create connection pool
pool, err := pgxpool.New(ctx, "postgres://user:pass@localhost/db")

// Configure pool
config, _ := pgxpool.ParseConfig(connString)
config.MaxConns = 25
config.MinConns = 5
pool, _ := pgxpool.NewWithConfig(ctx, config)

// Use with sqlc
queries := db.New(pool)
user, err := queries.GetUser(ctx, 123)
```

---

### golang-migrate

**What is it?**
A database migration tool for Go (and other languages).

**What problems does it solve?**

**Problem: Database schema changes**
- How do you version database schemas?
- How do you deploy schema changes?
- How do you rollback if something goes wrong?

**Solution: Migration files**

```bash
# Create migration
migrate create -ext sql -dir migrations -seq create_users_table
```

Creates two files:
```sql
-- 000001_create_users_table.up.sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 000001_create_users_table.down.sql
DROP TABLE users;
```

```bash
# Apply migrations
migrate -path migrations -database "postgresql://localhost/db" up

# Rollback
migrate -path migrations -database "postgresql://localhost/db" down 1
```

**Best practices:**
- Each migration is numbered sequentially
- Always create both up and down migrations
- Never modify existing migrations (create new ones)
- Test rollback migrations

---

## Infrastructure & DevOps

### Docker & docker-compose

**What is it?**
Docker packages your application with all dependencies into containers. docker-compose runs multiple containers together.

**What problems does it solve?**

**Without Docker:**
```
Developer 1: "Works on my machine!"
Developer 2: "I get errors..."
- Different PostgreSQL versions
- Different Go versions
- Missing environment variables
- Different OS (Windows/Mac/Linux)
```

**With Docker:**
```yaml
# docker-compose.yml
services:
  backend:
    image: golang:1.23
    # Everyone has same Go version

  postgres:
    image: postgres:16
    # Everyone has same PostgreSQL version

  frontend:
    image: node:20
    # Everyone has same Node version
```

**Example for your app:**
```yaml
version: '3.8'

services:
  backend:
    build: ./backend
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
    depends_on:
      - postgres

  frontend:
    build: ./frontend
    ports:
      - "3000:3000"

  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: myapp
      POSTGRES_PASSWORD: secret
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

```bash
# Start everything
docker-compose up

# Stop everything
docker-compose down
```

**Benefits:**
- Same environment for everyone
- Easy to set up for new developers
- Production-like environment locally
- Isolated (doesn't mess with your system)

---

### Air (Hot Reload)

**What is it?**
Tool for automatic reloading of Go applications during development.

**What problems does it solve?**

**Without Air:**
```bash
# 1. Make code change
# 2. Stop server (Ctrl+C)
# 3. Rebuild: go build
# 4. Run: ./app
# 5. Test change
# 6. Repeat...
```

**With Air:**
```bash
air  # Watches files, rebuilds and restarts automatically
```

```toml
# .air.toml
[build]
  cmd = "go build -o ./tmp/main ./cmd/api"
  bin = "tmp/main"
  include_ext = ["go"]
  exclude_dir = ["tmp", "vendor"]
```

Now when you save a `.go` file:
- Air detects change
- Rebuilds automatically
- Restarts server
- You just refresh browser

Similar to nodemon for Node.js or flask run --reload for Python.

---

## Summary

### The Complete Stack

```
Frontend:
  Next.js 15      → React framework with SSR
  TypeScript      → Type safety
  Shadcn/UI       → Component library
  React Query     → Server state management
  Zustand         → Client state management

Backend:
  Go              → Fast, compiled language
  Gin             → Web framework
  sqlc            → Type-safe database access
  Clean Arch      → Maintainable structure

Database:
  PostgreSQL 16   → Relational database
  pgx/pgxpool     → Fast Go driver
  golang-migrate  → Schema migrations

DevOps:
  Docker          → Containerization
  docker-compose  → Multi-container orchestration
  Air             → Hot reload
```

### Why This Stack?

1. **Type Safety**: TypeScript + sqlc catch errors early
2. **Performance**: Go + PostgreSQL handle scale
3. **Developer Experience**: Hot reload, great tooling, clear patterns
4. **Production Ready**: Battle-tested technologies
5. **Learning Value**: Transferable skills, deep understanding

Each technology solves specific problems and works well together to create a cohesive, modern web application stack.
