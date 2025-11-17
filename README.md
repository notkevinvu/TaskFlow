# TaskFlow - Intelligent Task Prioritization System

An intelligent task management system that automatically prioritizes tasks using multi-factor algorithms, helping you focus on what matters most.

## Features

- **Intelligent Prioritization** - Automatically calculates task priority based on multiple factors
- **Task Details Sidebar** - View comprehensive task information with smooth animations
- **Quick Task Creation** - Rapidly add new tasks with a streamlined modal interface
- **Analytics Dashboard** - Visualize your task patterns and completion rates
- **Responsive Design** - Works seamlessly on desktop, tablet, and mobile
- **Dark Mode Ready** - Built with light/dark mode support

## Tech Stack

### Frontend
- **Next.js 15** with App Router
- **React 19** with TypeScript
- **Tailwind CSS** for styling
- **shadcn/ui** component library
- **React Query** for data fetching
- **Zustand** for state management

### Backend
- **Go 1.23** with Gin framework
- **PostgreSQL 16** with pgx driver
- **JWT Authentication** (golang-jwt)
- **Clean Architecture** pattern
- **Docker** for containerization

## Getting Started

### Prerequisites

- **Node.js** 20+ and npm
- **Docker** and Docker Compose
- **(Optional) Go 1.23** for local backend development

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd TaskFlow
   ```

2. **Install frontend dependencies**
   ```bash
   cd frontend
   npm install
   ```

3. **Start the full stack with Docker Compose**
   ```bash
   # From project root
   docker compose up -d
   ```

   This starts:
   - **PostgreSQL** on `localhost:5432`
     - Database: `taskflow`
     - User: `taskflow_user`
     - Password: `taskflow_dev_password`
   - **Backend API** on `http://localhost:8080`
     - Health check: `http://localhost:8080/health`
   - **pgAdmin** on `http://localhost:5050`
     - Email: `admin@taskflow.dev`
     - Password: `admin`

4. **Run database migrations**
   ```bash
   # Option 1: Using Docker
   docker exec -it taskflow-backend sh -c "cd /root && ./migrate -path migrations -database \$DATABASE_URL up"

   # Option 2: Using local golang-migrate (if installed)
   cd backend
   make migrate-up
   ```

5. **Configure frontend environment**
   ```bash
   cd frontend
   cp .env.example .env
   # Edit .env and set NODE_ENV=production to use real backend
   ```

6. **Start the frontend development server**
   ```bash
   cd frontend
   npm run dev
   ```

7. **Open your browser**
   ```
   http://localhost:3000
   ```

### Quick Start (Docker Only)

For the fastest setup using Docker:

```bash
# 1. Start all services
docker compose up -d --build

# 2. Wait for services to be ready (check health)
docker compose ps

# 3. Run migrations
docker exec -it taskflow-backend sh -c "cd /root && ./migrate -path migrations -database \$DATABASE_URL up"

# 4. Frontend (separate terminal)
cd frontend && npm install && npm run dev

# Access: http://localhost:3000
```

### Development Modes

**Production Mode (Real Backend):**
```bash
# frontend/.env
NODE_ENV=production
NEXT_PUBLIC_API_URL=http://localhost:8080
```
- Full authentication required
- Real database integration
- Priority algorithm active

**Development Mode (Mock Data):**
```bash
# frontend/.env
NODE_ENV=development
```
- Mock task data
- Auto-login
- Frontend-only development

## Project Structure

```
TaskFlow/
├── frontend/                # Next.js frontend application
│   ├── app/                # Next.js App Router pages
│   │   ├── (auth)/        # Authentication pages
│   │   ├── (dashboard)/   # Dashboard pages
│   │   └── layout.tsx     # Root layout
│   ├── components/        # React components
│   │   ├── ui/           # shadcn/ui components
│   │   ├── CreateTaskDialog.tsx
│   │   └── TaskDetailsSidebar.tsx
│   ├── hooks/            # Custom React hooks
│   ├── lib/              # Utilities and API client
│   └── public/           # Static assets
├── backend/              # Go backend API
│   ├── cmd/server/      # Application entry point
│   ├── internal/        # Core application code
│   │   ├── domain/     # Business entities & priority algorithm
│   │   ├── repository/ # Database access layer
│   │   ├── service/    # Business logic
│   │   ├── handler/    # HTTP handlers
│   │   ├── middleware/ # Auth, CORS, rate limiting
│   │   └── config/     # Configuration
│   ├── migrations/     # Database migrations
│   ├── docs/          # Backend documentation
│   └── Makefile       # Build commands
├── docker-compose.yml  # Full stack orchestration
└── docs/              # Project documentation
```

## Available Scripts

### Frontend

```bash
cd frontend
npm run dev      # Start development server
npm run build    # Build for production
npm run start    # Start production server
npm run lint     # Run ESLint
```

### Backend

```bash
cd backend
make help         # Show all available commands
make run          # Run locally (requires Go 1.23)
make build        # Build binary
make test         # Run unit tests
make test-coverage # Run tests with coverage
make migrate-up   # Run database migrations
make migrate-down # Rollback migrations
```

### Docker

```bash
docker compose up -d         # Start all services
docker compose down          # Stop all services
docker compose logs -f       # View logs (all services)
docker compose logs backend  # View backend logs only
docker compose ps            # Check service status
docker compose restart backend  # Restart backend service
```

## Implementation Status

### ✅ Phase 1 Complete (Frontend)
- Next.js 15 frontend with TypeScript
- Authentication UI (login/register)
- Dashboard with priority visualization
- Task details sidebar with animations
- Quick task creation modal
- Analytics page with charts
- Responsive design

### ✅ Phase 2 Complete (Backend + Smart Prioritization)
- Go backend with Clean Architecture
- PostgreSQL database with full schema
- JWT authentication (email/password)
- **Smart Priority Algorithm:**
  - User Priority (40%)
  - Time Decay (30%)
  - Deadline Urgency (20%)
  - Bump Penalty (10%)
  - Effort Boost multiplier
- Full-text search (PostgreSQL tsvector)
- Task history/audit log
- Rate limiting (100 req/min)
- CORS middleware
- Docker Compose setup
- Unit tests for priority calculator

### 🚧 Phase 3 (Planned - Analytics & Advanced Features)
- Background job for auto-reprioritization
- Advanced analytics (estimation accuracy)
- Category breakdown charts
- Velocity tracking
- At-risk task alerts
- Design system improvements

## Design System

Currently using:
- **shadcn/ui** components with CSS variables
- **OKLCH color space** for modern color handling
- **Light/dark mode** ready
- **Semantic tokens** (primary, destructive, muted, etc.)

For future design system improvements and Figma integration plans, see the "Design System Improvements" section in `PRD.md`.

## Documentation

### Project Docs
- **PRD.md** - Product Requirements Document
- **architecture-overview.md** - Technical architecture
- **data-model.md** - Database schema
- **priority-algorithm.md** - Prioritization logic
- **phase-1-weeks-1-2.md** - Frontend implementation guide

### Backend Docs
- **backend/README.md** - Backend setup and API reference
- **backend/docs/SUPABASE_MIGRATION.md** - Migrating to Supabase cloud

## API Documentation

The backend exposes a REST API at `http://localhost:8080/api/v1`:

### Authentication
- `POST /auth/register` - Create account
- `POST /auth/login` - Login and get JWT token
- `GET /auth/me` - Get current user (requires auth)

### Tasks (all require authentication)
- `POST /tasks` - Create task
- `GET /tasks?status=&category=&search=&limit=&offset=` - List tasks
- `GET /tasks/:id` - Get single task
- `PUT /tasks/:id` - Update task
- `DELETE /tasks/:id` - Delete task
- `POST /tasks/:id/bump` - Bump task (increment delay counter)
- `POST /tasks/:id/complete` - Mark task complete

See `backend/README.md` for detailed API documentation.

## Testing

### Run Backend Unit Tests

```bash
cd backend
make test

# With coverage
make test-coverage
```

### Test Priority Calculator

```bash
cd backend/internal/domain/priority
go test -v
```

### Manual API Testing

```bash
# Health check
curl http://localhost:8080/health

# Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","name":"Test","password":"Test1234"}'

# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test1234"}'

# Create task (use token from login)
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{"title":"Test task","user_priority":75}'
```

## Contributing

This is a personal project, but feedback and suggestions are welcome!

## License

MIT

## Troubleshooting

### Backend won't start
```bash
# Check if PostgreSQL is running
docker ps | grep postgres

# Check backend logs
docker compose logs backend

# Ensure migrations ran
docker exec -it taskflow-postgres psql -U taskflow_user -d taskflow -c "\dt"
```

### Frontend can't connect to backend
```bash
# Verify backend is running
curl http://localhost:8080/health

# Check frontend .env
cat frontend/.env
# Should have: NEXT_PUBLIC_API_URL=http://localhost:8080
# Should have: NODE_ENV=production
```

### Database migration errors
```bash
# Check migration status
cd backend
make migrate-down
make migrate-up

# Or via Docker
docker exec -it taskflow-backend sh -c "./migrate -path migrations -database \$DATABASE_URL version"
```

## Production Deployment

1. **Update environment variables**
   - Change `JWT_SECRET` to a secure random value
   - Use strong database password
   - Update `ALLOWED_ORIGINS` for your frontend URL

2. **Database Migration**
   - For Supabase: See `backend/docs/SUPABASE_MIGRATION.md`
   - For other cloud providers: Similar pg_dump/restore process

3. **Build for production**
   ```bash
   # Frontend
   cd frontend && npm run build

   # Backend
   cd backend && make build
   ```

---

**Status:** Phase 2 Complete ✅
**Last Updated:** 2025-01-16
