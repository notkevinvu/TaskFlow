# TaskFlow Deployment Testing Checklist

This document provides step-by-step instructions to test the new Docker deployment infrastructure.

## Prerequisites

Before testing, ensure you have:
- ✅ Docker Desktop installed and running
- ✅ At least 4GB RAM available for Docker
- ✅ Ports 3000, 5050, 5432, 8080 are not in use

Check Docker is running:
```bash
docker --version
docker compose version
```

---

## Test 1: Production Mode (Full Stack)

### Windows
```bash
# From project root
start.bat
```

### Linux/Mac
```bash
# Make scripts executable (first time only)
chmod +x *.sh

# Start the stack
./start.sh
```

### Expected Results

1. **Console Output:**
   ```
   ====================================
     TaskFlow - Starting Full Stack
   ====================================

   [1/4] Starting Docker services...
   [+] Running 4/4
    ✔ Container taskflow-postgres   Started
    ✔ Container taskflow-pgadmin    Started
    ✔ Container taskflow-backend    Started
    ✔ Container taskflow-frontend   Started

   [2/4] Waiting for services to be healthy...
   [3/4] Running database migrations...
   [4/4] All services started successfully!
   ```

2. **Browser Opens Automatically:**
   - Frontend should open at http://localhost:3000

3. **Test Each Service:**
   - **Frontend** (http://localhost:3000)
     - Should show login/register page
     - UI should be responsive

   - **Backend** (http://localhost:8080/health)
     - Should return: `{"status":"healthy"}`

   - **PgAdmin** (http://localhost:5050)
     - Login: admin@taskflow.dev / admin
     - Should show database connection

4. **Docker Status:**
   ```bash
   docker compose ps
   ```
   Should show all 4 services as "running" (healthy)

---

## Test 2: Development Mode (Hot Reload)

### Stop Previous Stack
```bash
# Windows
stop.bat

# Linux/Mac
./stop.sh
```

### Start Dev Mode

#### Windows
```bash
start-dev.bat
```

#### Linux/Mac
```bash
./start-dev.sh
```

### Expected Results

1. **Services Start with Hot Reload:**
   ```
   ====================================
     TaskFlow - Starting (Dev Mode)
   ====================================

   Frontend:  http://localhost:3000  (Hot Reload ✓)
   Backend:   http://localhost:8080  (Hot Reload ✓)
   ```

2. **Test Hot Reload (Frontend):**
   - Edit `frontend/app/(dashboard)/dashboard/page.tsx`
   - Change a text string (e.g., "Dashboard" → "Dashboard 🚀")
   - Save the file
   - Browser should auto-refresh within 2-3 seconds
   - **Expected:** Changes appear without rebuilding

3. **Test Hot Reload (Backend):**
   - Edit `backend/cmd/server/main.go`
   - Change the health check response
   - Save the file
   - Wait ~5 seconds
   - Visit http://localhost:8080/health
   - **Note:** Backend hot reload requires Air configuration (see docker-compose.dev.yml comments)

---

## Test 3: View Logs

### Windows
```bash
# All services
logs.bat

# Specific service
logs.bat frontend
logs.bat backend
logs.bat postgres
```

### Linux/Mac
```bash
# All services
./logs.sh

# Specific service
./logs.sh frontend
./logs.sh backend
./logs.sh postgres
```

### Expected Results

- Logs stream in real-time
- Press Ctrl+C to exit
- No error messages in logs (warnings are OK)

---

## Test 4: Database Migrations

### Windows
```bash
run-migrations.bat
```

### Linux/Mac
```bash
./run-migrations.sh
```

### Expected Results

```
Running database migrations...
Waiting for database to be ready...
1/u create_users_table (X.XXXs)
2/u create_tasks_table (X.XXXs)
3/u create_task_history_table (X.XXXs)

Migrations complete!
```

### Verify Database Schema

1. Open PgAdmin: http://localhost:5050
2. Login: admin@taskflow.dev / admin
3. Connect to taskflow database:
   - Host: postgres
   - Port: 5432
   - Username: taskflow_user
   - Password: taskflow_dev_password
4. Check tables exist:
   - users
   - tasks
   - task_history

---

## Test 5: Full Integration Test

### 1. Register a New User

1. Go to http://localhost:3000
2. Click "Register"
3. Fill in:
   - Name: Test User
   - Email: test@example.com
   - Password: Password123
4. Submit form

**Expected:** Redirected to dashboard with empty task list

### 2. Create a Task

1. Click "Quick Add" button
2. Fill in task details:
   - Title: Test Task
   - Description: This is a test task
   - Priority: High
   - Category: Work
3. Submit

**Expected:** Task appears in task list with priority score

### 3. View Task Details

1. Click on the task
2. **Expected:** Sidebar slides in from right with task details

### 4. Check Backend API

```bash
# Get auth token (replace with your actual token from login)
curl http://localhost:8080/api/v1/tasks

# Should return JSON array of tasks
```

---

## Test 6: Stop and Reset

### Stop Services (Keep Data)

#### Windows
```bash
stop.bat
```

#### Linux/Mac
```bash
./stop.sh
```

**Expected:**
- All containers stop
- Volumes are preserved
- Can restart with `start.bat` and data persists

### Reset Everything (Wipe Data)

#### Windows
```bash
reset.bat
```

#### Linux/Mac
```bash
./reset.sh
```

**Prompt:**
```
WARNING: This will delete ALL data!
Are you sure? (yes/no):
```

Type `yes` and press Enter.

**Expected:**
- All containers stopped
- All volumes deleted
- Docker images removed
- Database wiped clean

---

## Test 7: Supabase Cloud Database (Optional)

**Prerequisites:** Supabase account and project

### 1. Get Supabase Connection String

1. Go to https://supabase.com
2. Create a project
3. Go to Project Settings → Database
4. Copy "Connection string" (Transaction mode)

### 2. Set Environment Variable

#### Windows
```bash
set SUPABASE_DB_URL=postgres://postgres.PROJECT_REF:PASSWORD@aws-0-REGION.pooler.supabase.com:6543/postgres?sslmode=require
```

#### Linux/Mac
```bash
export SUPABASE_DB_URL="postgres://postgres.PROJECT_REF:PASSWORD@aws-0-REGION.pooler.supabase.com:6543/postgres?sslmode=require"
```

### 3. Start with Supabase

```bash
docker compose -f docker-compose.yml -f docker-compose.supabase.yml up -d
```

### 4. Expected Results

- Only 2 containers start: backend + frontend
- No local postgres or pgadmin
- Backend connects to Supabase cloud database
- Check Supabase dashboard to see tables created

---

## Troubleshooting

### Issue: Port Already in Use

**Error:**
```
Error: Ports are not available: exposing port TCP 0.0.0.0:3000
```

**Solution:**
```bash
# Find what's using the port
# Windows
netstat -ano | findstr :3000
taskkill /PID <PID> /F

# Linux/Mac
lsof -i :3000
kill -9 <PID>
```

### Issue: Docker Build Fails

**Error:**
```
failed to solve: failed to fetch
```

**Solution:**
1. Check internet connection
2. Restart Docker Desktop
3. Clear Docker cache:
   ```bash
   docker system prune -a
   ```

### Issue: Frontend Can't Connect to Backend

**Symptoms:**
- Frontend loads but shows connection errors
- API calls fail with CORS or network errors

**Solution:**
1. Check backend is running: `docker compose ps`
2. Check backend health: http://localhost:8080/health
3. Check CORS settings in `docker-compose.yml`:
   ```yaml
   ALLOWED_ORIGINS: http://localhost:3000
   ```
4. Check frontend env: `NEXT_PUBLIC_API_URL=http://localhost:8080`

### Issue: Database Migration Fails

**Error:**
```
error: pq: connection refused
```

**Solution:**
1. Wait longer for PostgreSQL to be ready:
   ```bash
   docker compose logs postgres
   ```
2. Check postgres is healthy:
   ```bash
   docker compose ps postgres
   ```
3. Manually retry migrations:
   ```bash
   run-migrations.bat  # or ./run-migrations.sh
   ```

---

## Success Criteria

All tests pass if:

✅ Production mode starts all 4 services successfully
✅ Frontend accessible at http://localhost:3000
✅ Backend API responds at http://localhost:8080
✅ PgAdmin accessible at http://localhost:5050
✅ Database migrations run successfully
✅ Can register user and create tasks
✅ Task details sidebar works
✅ Hot reload works in dev mode (frontend)
✅ Logs command shows service logs
✅ Stop command gracefully stops services
✅ Reset command wipes all data
✅ Restart after stop preserves data

---

## Performance Benchmarks

Expected timings on typical development machine:

- **Initial build:** 3-5 minutes (first time)
- **Subsequent starts:** 30-60 seconds
- **Hot reload (frontend):** 1-3 seconds
- **Database migration:** 5-10 seconds
- **Shutdown:** 5-10 seconds

---

## Next Steps

After successful testing:

1. **Commit changes:**
   ```bash
   git add .
   git commit -m "Add Docker deployment infrastructure"
   git push
   ```

2. **Update Phase status:**
   - Mark deployment infrastructure complete in PRD

3. **Production deployment:**
   - Follow `backend/docs/SUPABASE_MIGRATION.md` for cloud database
   - Configure CI/CD for automated deployments
   - Set up monitoring and logging

---

**Last Updated:** 2025-01-20
**Deployment Version:** Full Stack Docker
