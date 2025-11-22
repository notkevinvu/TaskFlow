# Getting Started with TaskFlow

Complete guide to get TaskFlow up and running on your local machine.

## Prerequisites Check

Before starting, ensure you have:

```bash
# Check Node.js (should be 20+)
node --version

# Check Docker
docker --version
docker compose version

# Check npm
npm --version
```

## Option 1: Full Stack with Docker (Recommended)

This is the fastest way to get everything running.

### Step 1: Clone and Install

```bash
# Clone the repository
git clone <repository-url>
cd TaskFlow

# Install frontend dependencies
cd frontend
npm install
cd ..
```

### Step 2: Start Backend Services

```bash
# From project root
docker compose up -d --build
```

This starts:
- PostgreSQL database (port 5432)
- Backend API (port 8080)
- PgAdmin (port 5050)

Wait 30-60 seconds for services to initialize.

### Step 3: Verify Services

```bash
# Check all services are running
docker compose ps

# Should see:
# - taskflow-postgres (healthy)
# - taskflow-backend (up)
# - taskflow-pgadmin (up)

# Test backend health
curl http://localhost:8080/health
# Should return: {"status":"healthy","timestamp":"..."}
```

### Step 4: Run Database Migrations

```bash
# Install golang-migrate (one-time setup)
# macOS
brew install golang-migrate

# Windows (via Chocolatey)
choco install migrate

# Linux
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.18.1/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# Run migrations
cd backend
make migrate-up

# Or without golang-migrate (using Docker)
# Note: migrate binary is included in Docker image
docker exec -it taskflow-backend sh
# Inside container:
migrate -path /root/migrations -database "postgres://taskflow_user:taskflow_dev_password@postgres:5432/taskflow?sslmode=disable" up
exit
```

### Step 5: Configure Frontend

```bash
cd frontend

# Copy environment file
cp .env.example .env

# Edit .env and set:
# NODE_ENV=production
# NEXT_PUBLIC_API_URL=http://localhost:8080

# Or use this one-liner:
echo "NEXT_PUBLIC_API_URL=http://localhost:8080" > .env
echo "NODE_ENV=production" >> .env
```

### Step 6: Start Frontend

```bash
npm run dev
```

### Step 7: Test the Application

Open browser: `http://localhost:3000`

1. **Register a new account:**
   - Click "Sign Up"
   - Email: `your@email.com`
   - Name: `Your Name`
   - Password: `Test1234` (must have uppercase, lowercase, number)

2. **Create your first task:**
   - Click "Quick Add" button
   - Title: `Test task`
   - Priority: 75 (High)
   - Click "Create Task"

3. **Verify priority calculation:**
   - Task should appear in list
   - Check priority score (calculated automatically)
   - Click task to see details sidebar

**Success!** Your TaskFlow instance is fully running.

---

## Option 2: Local Backend Development (Go Required)

For backend development without Docker.

### Prerequisites

- Go 1.23+
- PostgreSQL 16 running locally

### Step 1: Start PostgreSQL

```bash
# Using Docker for PostgreSQL only
docker run -d \
  --name taskflow-postgres \
  -e POSTGRES_USER=taskflow_user \
  -e POSTGRES_PASSWORD=taskflow_dev_password \
  -e POSTGRES_DB=taskflow \
  -p 5432:5432 \
  postgres:16-alpine
```

### Step 2: Set Up Backend

```bash
cd backend

# Create .env file
cp .env.example .env

# Install dependencies
go mod download

# Run migrations
make migrate-up

# Start backend
make run
```

Backend should start on `http://localhost:8080`.

### Step 3: Start Frontend

```bash
cd frontend
npm install

# Configure .env
echo "NEXT_PUBLIC_API_URL=http://localhost:8080" > .env
echo "NODE_ENV=production" >> .env

# Start
npm run dev
```

---

## Option 3: Frontend Only (Mock Data)

For UI development without backend.

```bash
cd frontend
npm install

# Configure for mock mode
echo "NODE_ENV=development" > .env

# Start
npm run dev
```

Open `http://localhost:3000` - you'll be auto-logged in with mock data.

---

## Verifying Your Setup

### Backend Verification

```bash
# 1. Health check
curl http://localhost:8080/health
# Expected: {"status":"healthy",...}

# 2. Register test user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "name": "Test User",
    "password": "Test1234"
  }'
# Expected: {"user":{...},"access_token":"eyJ..."}

# Save the token from response
TOKEN="<paste-token-here>"

# 3. Create a task
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "My first task",
    "user_priority": 75,
    "description": "Testing the API",
    "estimated_effort": "small"
  }'
# Expected: {"id":"...","title":"My first task","priority_score":39,...}

# 4. List tasks
curl http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer $TOKEN"
# Expected: {"tasks":[...],"total_count":1}
```

### Database Verification

```bash
# Connect to PostgreSQL
docker exec -it taskflow-postgres psql -U taskflow_user -d taskflow

# Check tables exist
\dt
# Expected: users, tasks, task_history

# Check user count
SELECT COUNT(*) FROM users;

# Check task count
SELECT COUNT(*) FROM tasks;

# Exit
\q
```

### Frontend Verification

1. Open `http://localhost:3000`
2. Register with email/password
3. Create a task
4. Check task appears in list
5. Click task to open details sidebar
6. Try bumping the task
7. Try completing the task

---

## Common Issues

### Issue: "Connection refused" on backend

**Solution:**
```bash
# Check if backend is running
docker compose ps

# Check backend logs
docker compose logs backend

# Restart backend
docker compose restart backend
```

### Issue: "Database connection failed"

**Solution:**
```bash
# Ensure PostgreSQL is healthy
docker compose ps postgres

# Check PostgreSQL logs
docker compose logs postgres

# Verify database exists
docker exec -it taskflow-postgres psql -U taskflow_user -l
```

### Issue: "Migrations failed"

**Solution:**
```bash
# Check current migration version
cd backend
make migrate-version

# Reset and rerun
make migrate-down
make migrate-up
```

### Issue: "Frontend can't connect to backend"

**Solution:**
```bash
# 1. Verify backend is running
curl http://localhost:8080/health

# 2. Check frontend .env
cat frontend/.env

# 3. Should have:
# NEXT_PUBLIC_API_URL=http://localhost:8080
# NODE_ENV=production

# 4. Restart frontend dev server
cd frontend
npm run dev
```

### Issue: "CORS errors in browser"

**Solution:**
```bash
# Check backend ALLOWED_ORIGINS
docker compose logs backend | grep ALLOWED_ORIGINS

# Should include: http://localhost:3000

# Update docker-compose.yml if needed:
# backend:
#   environment:
#     ALLOWED_ORIGINS: http://localhost:3000

# Restart
docker compose restart backend
```

---

## Next Steps

Once you have TaskFlow running:

1. **Explore the Priority Algorithm**
   - Create tasks with different priorities (0-100)
   - Add due dates to see deadline urgency
   - Bump tasks to see how priority changes
   - Check `backend/internal/domain/priority/calculator.go` for formula

2. **Test Full-Text Search**
   - Create tasks with various titles/descriptions
   - Use search bar in dashboard
   - Search is powered by PostgreSQL tsvector

3. **View Task History**
   - Every action (create, update, bump, complete) is logged
   - Check `task_history` table in database

4. **Try the Analytics Page**
   - Visit `/analytics` in the app
   - See task distribution and completion rates

5. **Develop Custom Features**
   - Backend: `backend/internal/`
   - Frontend: `frontend/app/` and `frontend/components/`
   - Read `backend/README.md` for API details

---

## Quick Reference

### URLs
- **Frontend:** http://localhost:3000
- **Backend API:** http://localhost:8080
- **API Health:** http://localhost:8080/health
- **PgAdmin:** http://localhost:5050

### Database Credentials
- **Host:** localhost:5432
- **Database:** taskflow
- **User:** taskflow_user
- **Password:** taskflow_dev_password

### PgAdmin Credentials
- **Email:** admin@taskflow.dev
- **Password:** admin

### Docker Commands
```bash
docker compose up -d          # Start all services
docker compose down           # Stop all services
docker compose logs -f        # View logs (all)
docker compose logs backend   # View backend logs only
docker compose ps             # Check status
docker compose restart backend # Restart backend
```

### Backend Commands
```bash
cd backend
make help          # Show all commands
make run           # Run locally
make test          # Run tests
make migrate-up    # Run migrations
make migrate-down  # Rollback migrations
```

### Frontend Commands
```bash
cd frontend
npm run dev        # Start dev server
npm run build      # Build for production
npm run lint       # Lint code
```

---

## Support

- **Issues:** Create an issue on GitHub
- **Docs:** See `backend/README.md` and `README.md`
- **Migration Guide:** `backend/docs/SUPABASE_MIGRATION.md`

---

**Happy task prioritizing!** 🎯
