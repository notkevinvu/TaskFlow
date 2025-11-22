# Migrating to Supabase

Guide for migrating your TaskFlow database from local PostgreSQL to Supabase.

## Why Supabase?

Supabase provides:
- **Managed PostgreSQL** - No server maintenance
- **Free tier** - 500MB database, good for MVP
- **Auto-backups** - Daily backups included
- **Dashboard** - Web UI for database management
- **Connection pooling** - Built-in PgBouncer
- **Realtime subscriptions** - PostgreSQL change data capture (future feature)

## Prerequisites

1. **Supabase Account** - Sign up at https://supabase.com
2. **Existing Local Database** - TaskFlow running locally
3. **Data to Migrate** (optional) - If you have test data to preserve

## Step 1: Create Supabase Project

1. Go to https://app.supabase.com
2. Click "New Project"
3. Enter:
   - **Name:** TaskFlow
   - **Database Password:** (save this securely)
   - **Region:** Choose closest to your users
4. Wait 2-3 minutes for project creation

## Step 2: Get Connection Details

1. In Supabase dashboard, go to **Settings → Database**
2. Find **Connection string** section
3. Copy the **URI** format:
   ```
   postgres://postgres:[YOUR-PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres
   ```
4. Replace `[YOUR-PASSWORD]` with your database password

## Step 3: Run Migrations on Supabase

### Option A: Using golang-migrate

```bash
# From backend directory
migrate -path migrations \
  -database "postgres://postgres:[PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres?sslmode=require" \
  up
```

### Option B: Using Supabase SQL Editor

1. Go to **SQL Editor** in Supabase dashboard
2. Copy contents of `migrations/000001_initial_schema.up.sql`
3. Paste and run in SQL Editor
4. Verify tables created in **Table Editor**

## Step 4: Update Backend Configuration

### Development Environment

Update `backend/.env`:

```bash
# OLD (Local)
DATABASE_URL=postgres://taskflow_user:taskflow_dev_password@localhost:5432/taskflow?sslmode=disable

# NEW (Supabase)
DATABASE_URL=postgres://postgres:[YOUR-PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres?sslmode=require
```

### Production Deployment (Docker/Cloud)

Update environment variables:

```bash
# Docker Compose
services:
  backend:
    environment:
      DATABASE_URL: postgres://postgres:[PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres?sslmode=require

# Or use .env file in production
```

## Step 5: Migrate Data (Optional)

If you have existing data in local database:

### Export from Local PostgreSQL

```bash
# Export data (no schema, just data)
pg_dump -h localhost \
  -U taskflow_user \
  -d taskflow \
  --data-only \
  --no-owner \
  --no-privileges \
  -f taskflow_data.sql
```

### Import to Supabase

```bash
# Import data
psql "postgres://postgres:[PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres?sslmode=require" \
  -f taskflow_data.sql
```

## Step 6: Test Connection

```bash
# Test backend connection
cd backend
make run

# Check logs for "Successfully connected to database"
```

## Step 7: Update Frontend (if needed)

Frontend doesn't need changes - it connects to backend, which now connects to Supabase.

```bash
# Just ensure backend URL is correct
# frontend/.env
NEXT_PUBLIC_API_URL=http://localhost:8080  # Local backend
# OR
NEXT_PUBLIC_API_URL=https://your-backend.com  # Deployed backend
```

## Step 8: Verify Migration

### Check Tables

```bash
# Connect to Supabase
psql "postgres://postgres:[PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres"

# List tables
\dt

# Should see:
# - users
# - tasks
# - task_history
```

### Check Extensions

```sql
SELECT * FROM pg_extension WHERE extname = 'uuid-ossp';
```

### Check Indexes

```sql
SELECT tablename, indexname FROM pg_indexes
WHERE schemaname = 'public'
ORDER BY tablename;
```

### Test API

```bash
# Register a user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "name": "Test User",
    "password": "Test1234"
  }'

# Create a task (use token from register)
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "title": "Test task on Supabase",
    "user_priority": 75
  }'
```

## Supabase Features to Use

### 1. Connection Pooling

Supabase provides PgBouncer connection pooling:

```bash
# Use transaction mode for better performance
DATABASE_URL=postgres://postgres:[PASSWORD]@db.[PROJECT-REF].supabase.co:6543/postgres?sslmode=require

# Note the port change: 6543 instead of 5432
```

Update `backend/internal/config/config.go` if using pooler.

### 2. Database Dashboard

Access at: https://app.supabase.com/project/[PROJECT-REF]/database/tables

- View tables
- Edit data
- See relationships
- Monitor performance

### 3. SQL Editor

Run ad-hoc queries:
```sql
-- Get user task counts
SELECT u.email, COUNT(t.id) as task_count
FROM users u
LEFT JOIN tasks t ON u.id = t.user_id
GROUP BY u.email;

-- Find at-risk tasks
SELECT title, bump_count, due_date
FROM tasks
WHERE bump_count >= 3 OR due_date < NOW() - INTERVAL '3 days';
```

### 4. Backups

- **Automatic daily backups** (free tier: 7 days retention)
- **Point-in-time recovery** (paid plans)
- Manual backups via **Database → Backups**

## Common Issues

### Issue: Connection Timeout

**Solution:** Check firewall/network. Supabase requires internet access.

```bash
# Test connectivity
nc -zv db.[PROJECT-REF].supabase.co 5432
```

### Issue: SSL Required

**Solution:** Always use `sslmode=require` with Supabase.

```bash
# WRONG
DATABASE_URL=...?sslmode=disable

# CORRECT
DATABASE_URL=...?sslmode=require
```

### Issue: Permission Denied

**Solution:** Supabase user is `postgres`, not `taskflow_user`.

Make sure connection string uses:
- User: `postgres`
- Password: Your Supabase database password

### Issue: Too Many Connections

**Solution:** Use connection pooler (port 6543) instead of direct (port 5432).

## Cost Estimation

**Free Tier:**
- 500MB database
- 2GB bandwidth
- Unlimited API requests
- Good for: 1000-5000 tasks, 100-500 users

**Pro Tier ($25/month):**
- 8GB database
- 50GB bandwidth
- Daily backups for 30 days
- Good for: 50,000+ tasks, 5,000+ users

## Rollback to Local (if needed)

```bash
# Export from Supabase
pg_dump "postgres://postgres:[PASSWORD]@db.[PROJECT-REF].supabase.co:5432/postgres" \
  -f supabase_backup.sql

# Import to local
psql -U taskflow_user -d taskflow -f supabase_backup.sql

# Update .env back to local
DATABASE_URL=postgres://taskflow_user:taskflow_dev_password@localhost:5432/taskflow?sslmode=disable
```

## Production Deployment Checklist

- [ ] Database password is strong (20+ characters)
- [ ] JWT_SECRET is changed from development value
- [ ] Connection string uses `sslmode=require`
- [ ] Consider using connection pooler (port 6543)
- [ ] Enable Row Level Security (RLS) in Supabase (optional)
- [ ] Set up monitoring/alerts in Supabase dashboard
- [ ] Configure database backups retention
- [ ] Document connection details securely (password manager)

## Next Steps

After migration:
1. Monitor database performance in Supabase dashboard
2. Set up alerts for connection/query issues
3. Configure scheduled backups
4. Consider enabling Supabase Realtime for live updates
5. Explore Supabase Storage for file attachments (future feature)

## Support

- **Supabase Docs:** https://supabase.com/docs
- **Supabase Discord:** https://discord.supabase.com
- **PostgreSQL Docs:** https://www.postgresql.org/docs/

---

**Migration Time Estimate:** 15-30 minutes
**Downtime Required:** 0 minutes (run backend pointing to Supabase)
