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
- **Next.js 16** with App Router and Turbopack
- **React 19** with TypeScript
- **Tailwind CSS 4** for styling
- **shadcn/ui** component library
- **React Query** for data fetching
- **Zustand** for state management

### Backend
- **Go 1.23** with Gin framework
- **Supabase** (managed PostgreSQL 16)
- **JWT Authentication** (golang-jwt)
- **Clean Architecture** pattern
- **Full-text search** with PostgreSQL tsvector

### Database
- **Supabase PostgreSQL** - Managed cloud database
- **Connection pooling** via PgBouncer
- **Automatic backups** and point-in-time recovery
- **Web dashboard** for database management

## Quick Start

### Prerequisites

- **Node.js** 20+ and npm
- **Go 1.23+**
- **Supabase Account** (free tier available)

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd TaskFlow
   ```

2. **Set up Supabase**
   - Create a free account at [supabase.com](https://supabase.com)
   - Create a new project
   - Get your database connection string from **Settings → Database**
   - Copy the **URI** format connection string

3. **Configure backend**
   ```bash
   cd backend
   cp .env.example .env
   # Edit .env and set DATABASE_URL to your Supabase connection string
   ```

4. **Run database migrations**
   ```bash
   cd backend
   go run cmd/server/main.go
   # Migrations run automatically on startup
   ```

5. **Install frontend dependencies**
   ```bash
   cd frontend
   npm install
   ```

6. **Start development servers**

   **Option 1: Using helper scripts (recommended)**

   Windows:
   ```bash
   scripts\start.bat
   ```

   Linux/Mac:
   ```bash
   chmod +x scripts/*.sh
   scripts/start.sh
   ```

   **Option 2: Manual start**

   Terminal 1 (Backend):
   ```bash
   cd backend
   go run cmd/server/main.go
   ```

   Terminal 2 (Frontend):
   ```bash
   cd frontend
   npm run dev
   ```

7. **Open your browser**
   ```
   http://localhost:3000
   ```

**What you get:**
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- Database: Managed by Supabase (view in Supabase dashboard)

## Project Structure

```
TaskFlow/
├── frontend/                # Next.js frontend application
│   ├── app/                # Next.js App Router pages
│   │   ├── (auth)/        # Authentication pages (login, register)
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
│   ├── migrations/     # Database migrations (auto-run)
│   └── Makefile       # Build commands
├── docs/              # Project documentation
│   ├── product/      # PRD, data model, priority algorithm
│   ├── architecture/ # System design and tech stack
│   ├── implementation/ # Phase plans and patterns
│   └── guides/       # Setup guides and troubleshooting
└── scripts/          # Development helper scripts
    ├── start.bat     # Start dev servers (Windows)
    ├── start.sh      # Start dev servers (Linux/Mac)
    ├── stop.bat      # Stop all services (Windows)
    └── stop.sh       # Stop all services (Linux/Mac)
```

## Available Scripts

### Development Scripts

**Windows:**
```bash
scripts\start.bat    # Start both backend and frontend
scripts\stop.bat     # Stop all services
```

**Linux/Mac:**
```bash
scripts/start.sh   # Start both backend and frontend (Ctrl+C to stop)
scripts/stop.sh    # Force stop all services
```

### Frontend Commands

```bash
cd frontend
npm run dev      # Start development server (Turbopack)
npm run build    # Build for production
npm run start    # Start production server
npm run lint     # Run ESLint
```

### Backend Commands

```bash
cd backend
make help         # Show all available commands
make run          # Run locally
make build        # Build binary
make test         # Run unit tests
make test-coverage # Run tests with coverage report
```

## Implementation Status

### ✅ Phase 1 Complete (Frontend Foundation)
- Next.js 16 frontend with TypeScript and Turbopack
- Authentication UI (login/register)
- Dashboard with task list and priority visualization
- Task details sidebar with smooth animations
- Quick task creation modal
- Responsive design with independent scrolling

### ✅ Phase 2 Complete (Backend + Smart Prioritization)
- Go backend with Clean Architecture
- Supabase PostgreSQL integration
- JWT authentication (email/password)
- **Smart Priority Algorithm:**
  - User Priority (40%)
  - Time Decay (30%)
  - Deadline Urgency (20%)
  - Bump Penalty (10%)
  - Effort Boost multiplier
- Full-text search (PostgreSQL tsvector + GIN indexes)
- Task history/audit log
- Rate limiting (100 req/min per user)
- CORS middleware
- Unit tests for priority calculator (100% coverage)
- Database migrations (auto-run on startup)

### 🚧 Phase 3 (Planned - Analytics & Advanced Features)
- Background job for auto-reprioritization
- Advanced analytics (estimation accuracy tracking)
- Category breakdown charts
- Velocity tracking over time
- At-risk task email alerts
- Design system token improvements

## API Documentation

The backend exposes a REST API at `http://localhost:8080/api/v1`:

### Authentication
- `POST /auth/register` - Create account
- `POST /auth/login` - Login and get JWT token
- `GET /auth/me` - Get current user info (requires auth)

### Tasks (all require authentication)
- `POST /tasks` - Create task (auto-calculates priority)
- `GET /tasks?status=&category=&search=&limit=&offset=` - List tasks
- `GET /tasks/:id` - Get single task details
- `PUT /tasks/:id` - Update task (recalculates priority)
- `DELETE /tasks/:id` - Delete task
- `POST /tasks/:id/bump` - Bump task (increment delay counter)
- `POST /tasks/:id/complete` - Mark task complete

For detailed API documentation and examples, see `backend/README.md`.

## Database Schema

### Tables
- **users** - User accounts with email/password auth
- **tasks** - Task records with priority scores
- **task_history** - Audit log of all task changes

### Key Features
- **Full-text search** using tsvector + GIN indexes
- **Automatic triggers** for search_vector updates
- **PostgreSQL enums** for task_status, task_effort, event_type
- **Optimized indexes** for priority sorting and filtering

See `docs/product/data-model.md` for complete schema details.

## Priority Algorithm

Tasks are automatically scored based on:

1. **User Priority (40%)** - Your explicit importance rating (0-100)
2. **Time Decay (30%)** - How long the task has existed (linear growth over 30 days)
3. **Deadline Urgency (20%)** - Quadratic urgency increase in final 7 days before due date
4. **Bump Penalty (10%)** - +10 points per delay/postponement (max 50)
5. **Effort Boost** - Small tasks get 1.3x multiplier, encouraging quick wins

**At-Risk Detection:**
- 3+ bumps (postponements)
- 3+ days overdue

For algorithm details and test cases, see `docs/product/priority-algorithm.md`.

## Documentation

All documentation is organized in the `docs/` folder:

### Product Documentation
- `docs/product/PRD.md` - Product Requirements Document
- `docs/product/data-model.md` - Database schema details
- `docs/product/priority-algorithm.md` - Prioritization logic

### Architecture Documentation
- `docs/architecture/architecture-overview.md` - System design
- `docs/architecture/tech-stack-explained.md` - Technology choices
- `docs/architecture/project-structure.md` - Codebase organization

### Implementation Plans
- `docs/implementation/phase-1-weeks-1-2.md` - Frontend implementation
- `docs/implementation/phase-2-weeks-3-4.md` - Backend implementation
- `docs/implementation/phase-3-weeks-5-6.md` - Analytics phase
- `docs/implementation/phase-4-month-2-plus.md` - Future enhancements
- `docs/implementation/common-patterns.md` - Code patterns and best practices

### Guides
- `docs/guides/quick-start.md` - Detailed setup instructions
- `docs/guides/troubleshooting.md` - Common issues and solutions
- `docs/guides/resources.md` - Learning resources
- `docs/guides/SECRETS_MANAGEMENT.md` - Security best practices
- `backend/docs/SUPABASE_MIGRATION.md` - Supabase setup guide

## Testing

### Run Backend Unit Tests

```bash
cd backend
make test

# With coverage report
make test-coverage
# Opens coverage.html in browser
```

### Test Priority Calculator

```bash
cd backend/internal/domain/priority
go test -v

# Expected: 8 scenarios pass, 100% coverage
```

### Manual API Testing

```bash
# Health check
curl http://localhost:8080/health

# Register user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","name":"Test User","password":"Test1234"}'

# Login (save the token)
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"Test1234"}'

# Create task (use token from login response)
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN_HERE" \
  -d '{"title":"Test task","user_priority":75}'
```

## Troubleshooting

### Backend won't start

```bash
# Check DATABASE_URL in backend/.env
cat backend/.env

# Verify Supabase connection
psql "YOUR_SUPABASE_CONNECTION_STRING"

# Check Go version
go version  # Should be 1.23+
```

### Frontend can't connect to backend

```bash
# Verify backend is running
curl http://localhost:8080/health
# Should return: {"status":"healthy"}

# Check if port 8080 is in use
# Windows: netstat -ano | findstr :8080
# Linux/Mac: lsof -i :8080
```

### Database migration errors

Migrations run automatically on backend startup. If you see errors:

```bash
# Check backend logs for migration status
cd backend
go run cmd/server/main.go
# Look for "Migration X applied successfully" messages

# Manually check database schema
psql "YOUR_SUPABASE_CONNECTION_STRING"
\dt  # List tables - should see: users, tasks, task_history
```

### Common Issues

**Port 3000 or 8080 already in use:**
```bash
# Use stop script to kill processes
scripts\stop.bat  # Windows
scripts/stop.sh   # Linux/Mac
```

**Supabase connection timeout:**
- Check firewall/network settings
- Verify connection string has `sslmode=require`
- Ensure Supabase project is active (not paused)

**Go dependencies not found:**
```bash
cd backend
go mod download
go mod tidy
```

For more troubleshooting tips, see `docs/guides/troubleshooting.md`.

## Architecture Decisions

### Why Local Development + Supabase?

- **Simplicity**: No Docker setup required, faster iteration
- **Cloud-Ready**: Database already in production environment
- **Free Tier**: Supabase free tier is generous (500MB DB, 2GB bandwidth)
- **Developer Experience**: Supabase dashboard for database management
- **Scalability**: Easy to add more Supabase features (Storage, Auth, Realtime)

### Why Go for Backend?

- **Performance**: Native concurrency, low memory footprint
- **Type Safety**: Strong typing without runtime overhead
- **Simplicity**: Easy deployment (single binary)
- **Ecosystem**: Excellent PostgreSQL drivers (pgx)

### Why Next.js for Frontend?

- **React 19**: Latest React features with Server Components
- **Turbopack**: Faster builds and hot reload
- **App Router**: Modern routing with layouts
- **TypeScript**: Type safety across the stack
- **Vercel**: Easy deployment when ready

## Contributing

This is a personal project, but feedback and suggestions are welcome! Feel free to:
- Open an issue for bugs or feature requests
- Submit a PR with improvements
- Share your thoughts on the architecture

## License

MIT

---

**Current Status:** Phase 2 Complete ✅
**Last Updated:** 2025-01-22
**Next Up:** Phase 3 - Analytics & Advanced Features
