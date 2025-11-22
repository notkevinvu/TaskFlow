# TaskFlow Backend

Go backend service for the TaskFlow intelligent task prioritization system.

## Architecture

This backend follows **Clean Architecture** principles:

```
backend/
├── cmd/server/           # Application entry point
├── internal/
│   ├── domain/          # Business entities and logic
│   ├── repository/      # Database access layer
│   ├── service/         # Application services
│   ├── handler/         # HTTP handlers (Gin)
│   ├── middleware/      # HTTP middleware
│   └── config/          # Configuration management
├── migrations/          # Database migrations
└── tests/              # Unit and integration tests
```

## Features

- **JWT Authentication** - Email/password with secure bcrypt hashing
- **Smart Priority Calculation** - Multi-factor algorithm (user priority, time decay, deadline urgency, bump penalty)
- **Full-Text Search** - PostgreSQL tsvector for fast text search
- **Task History** - Complete audit log of all task changes
- **Rate Limiting** - 100 requests/minute per user
- **CORS** - Configured for frontend integration

## Tech Stack

- **Language:** Go 1.23
- **Framework:** Gin
- **Database:** PostgreSQL 16 with pgx driver
- **Auth:** JWT (golang-jwt/jwt/v5)
- **Migrations:** golang-migrate
- **Password Hashing:** bcrypt (golang.org/x/crypto)

## Prerequisites

- Go 1.23+
- PostgreSQL 16
- (Optional) golang-migrate for running migrations
- (Optional) Docker & Docker Compose

## Quick Start

### 1. Using Docker Compose (Recommended)

```bash
# From project root
docker-compose up -d

# Backend will be available at http://localhost:8080
# PostgreSQL at localhost:5432
# PgAdmin at http://localhost:5050
```

### 2. Local Development

#### Install Dependencies

```bash
cd backend
make deps
```

#### Set Up Environment

```bash
# Copy example env file
cp .env.example .env

# Edit .env and configure your settings
```

#### Run Database Migrations

```bash
# Make sure PostgreSQL is running
make migrate-up
```

#### Run the Server

```bash
make run

# Or build and run
make build
./bin/server
```

## API Endpoints

### Authentication

```
POST   /api/v1/auth/register   - Create new user account
POST   /api/v1/auth/login      - Login and get JWT token
GET    /api/v1/auth/me         - Get current user (requires auth)
```

### Tasks (All require authentication)

```
POST   /api/v1/tasks           - Create new task
GET    /api/v1/tasks           - List tasks (filtered, sorted by priority)
GET    /api/v1/tasks/:id       - Get single task
PUT    /api/v1/tasks/:id       - Update task
DELETE /api/v1/tasks/:id       - Delete task
POST   /api/v1/tasks/:id/bump  - Bump task (increment delay counter)
POST   /api/v1/tasks/:id/complete - Mark task as complete
```

### Query Parameters for GET /api/v1/tasks

```
?status=todo|in_progress|done  - Filter by status
?category=string               - Filter by category
?search=string                 - Full-text search
?limit=number                  - Limit results (default: 20)
?offset=number                 - Pagination offset
```

## Environment Variables

```bash
# Server
PORT=8080                       # Server port
GIN_MODE=debug|release          # Gin mode

# Database
DATABASE_URL=postgres://...     # PostgreSQL connection string

# JWT
JWT_SECRET=your-secret-key      # Min 32 characters
JWT_EXPIRY_HOURS=24            # Token expiration

# Rate Limiting
RATE_LIMIT_REQUESTS_PER_MINUTE=100

# CORS
ALLOWED_ORIGINS=http://localhost:3000
```

## Database Migrations

### Create New Migration

```bash
make migrate-create name=add_new_feature
```

### Run Migrations

```bash
make migrate-up
```

### Rollback Migrations

```bash
make migrate-down
```

## Testing

### Run Unit Tests

```bash
make test
```

### Run Tests with Coverage

```bash
make test-coverage
```

### Run Priority Calculator Tests

```bash
cd internal/domain/priority
go test -v -cover
```

## Priority Calculation Algorithm

The backend implements a sophisticated priority scoring system:

```
PriorityScore = (
    UserPriority × 0.4 +
    TimeDecay × 0.3 +
    DeadlineUrgency × 0.2 +
    BumpPenalty × 0.1
) × EffortBoost
```

**Components:**

- **UserPriority** (0-100): User-set importance
- **TimeDecay** (0-100): Linear increase over 30 days
- **DeadlineUrgency** (0-100): Quadratic urgency as due date approaches
- **BumpPenalty** (0-50): +10 points per bump, capped at 50
- **EffortBoost** (1.0-1.3x): Small tasks get 1.3x, large tasks 1.0x

**At-Risk Detection:**
- Bump count ≥ 3
- OR overdue by ≥ 3 days

See `internal/domain/priority/calculator.go` for implementation.

## Project Structure Details

### Domain Layer (`internal/domain/`)

Core business entities and validation logic:
- `user.go` - User entity, DTOs, password hashing
- `task.go` - Task entity, status, effort enums
- `task_history.go` - Audit log events
- `priority/calculator.go` - Priority scoring algorithm

### Repository Layer (`internal/repository/`)

Database access with pgx:
- `user_repository.go` - User CRUD operations
- `task_repository.go` - Task CRUD, filtering, full-text search
- `task_history_repository.go` - Audit log persistence

### Service Layer (`internal/service/`)

Business logic orchestration:
- `auth_service.go` - Registration, login, JWT generation
- `task_service.go` - Task management, priority calculation, history logging

### Handler Layer (`internal/handler/`)

HTTP request handling with Gin:
- `auth_handler.go` - Auth endpoints
- `task_handler.go` - Task endpoints with validation

### Middleware (`internal/middleware/`)

HTTP middleware:
- `auth.go` - JWT validation
- `cors.go` - CORS configuration
- `rate_limit.go` - Token bucket rate limiter

## Troubleshooting

### Database Connection Issues

```bash
# Check PostgreSQL is running
docker ps | grep postgres

# Test connection
psql -h localhost -U taskflow_user -d taskflow
```

### Migration Errors

```bash
# Reset database (CAUTION: deletes all data)
make migrate-down
make migrate-up
```

### JWT Token Errors

- Ensure `JWT_SECRET` is set and >= 32 characters
- Check token expiration time
- Verify `Authorization: Bearer <token>` header format

## Performance

- **Priority Calculation:** < 100ms per task
- **Full-Text Search:** Uses PostgreSQL GIN index for fast queries
- **Connection Pooling:** pgxpool for efficient database connections
- **Rate Limiting:** In-memory token bucket (100 req/min per user)

## Security

- **Password Hashing:** bcrypt with cost 12
- **JWT Signing:** HS256 with secret key
- **SQL Injection Prevention:** Parameterized queries via pgx
- **CORS:** Configurable allowed origins
- **Input Validation:** Gin binding with struct tags

## Future Enhancements

- [ ] Background job for auto-reprioritization (every 6 hours)
- [ ] Websockets for real-time priority updates
- [ ] Redis caching for frequently accessed tasks
- [ ] Prometheus metrics and Grafana dashboards
- [ ] Distributed tracing with OpenTelemetry

## Contributing

1. Follow Go best practices and idiomatic patterns
2. Write unit tests for new features
3. Update documentation when adding endpoints
4. Run `go fmt` before committing

## License

MIT
