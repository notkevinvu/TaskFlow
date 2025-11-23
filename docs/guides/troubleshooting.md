# Troubleshooting Guide

Solutions to common issues you might encounter.

---

## Frontend Issues (Next.js)

### Port Already in Use

**Error:** `Port 3000 is already in use`

**Solutions:**

```bash
# Option 1: Kill the process
# macOS/Linux
lsof -i :3000
kill -9 <PID>

# Windows
netstat -ano | findstr :3000
taskkill /PID <PID> /F

# Option 2: Use different port
# In package.json:
"dev": "next dev -p 3001"

# Or set PORT environment variable
PORT=3001 npm run dev
```

---

### Module Not Found

**Error:** `Module not found: Can't resolve '@/components/...'`

**Solutions:**

```bash
# 1. Delete and reinstall
rm -rf node_modules package-lock.json
npm install

# 2. Check tsconfig.json paths
{
  "compilerOptions": {
    "paths": {
      "@/*": ["./*"]
    }
  }
}

# 3. Restart dev server
# Stop (Ctrl+C) and run: npm run dev
```

---

### Shadcn Components Not Styling

**Error:** Components look unstyled or broken

**Solutions:**

```bash
# 1. Verify globals.css is imported in layout.tsx
import "./globals.css";

# 2. Check tailwind.config.ts content paths
module.exports = {
  content: [
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
  ],
}

# 3. Restart dev server
Ctrl+C
npm run dev

# 4. Reinstall Shadcn
npx shadcn@latest init --force
```

---

### Hydration Errors

**Error:** `Text content does not match server-rendered HTML`

**Solution:**

```typescript
// Don't use browser-only code in Server Components
// Wrap in 'use client' directive

'use client';  // Add this at top of file

export function MyComponent() {
  // Now can use useState, useEffect, window, etc.
}
```

---

### CORS Errors

**Error:** `Access to fetch blocked by CORS policy`

**Solution (Backend):**

```go
// In main.go
router.Use(func(c *gin.Context) {
    c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
    c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

    if c.Request.Method == "OPTIONS" {
        c.AbortWithStatus(204)
        return
    }

    c.Next()
})
```

---

## Backend Issues (Go)

### Cannot Find Package

**Error:** `package github.com/.../... is not in GOROOT`

**Solutions:**

```bash
# 1. Download dependencies
go mod download

# 2. Tidy dependencies
go mod tidy

# 3. If module name changed, update all imports
# Update go.mod:
module github.com/NEWNAME/webapp

# Then:
find . -type f -name '*.go' -exec sed -i 's/OLDNAME/NEWNAME/g' {} +
go mod tidy
```

---

### Database Connection Refused

**Error:** `dial tcp 127.0.0.1:5432: connect: connection refused`

**Solutions:**

```bash
# 1. Check if PostgreSQL is running
docker ps | grep postgres

# 2. Start PostgreSQL
docker-compose up -d postgres

# 3. Check connection string
# Should be: postgresql://postgres:postgres@localhost:5432/webapp_dev?sslmode=disable

# 4. Test connection
docker exec -it web-app-postgres psql -U postgres -d webapp_dev

# 5. Check logs
docker-compose logs postgres
```

---

### Migration Failed

**Error:** Migration fails or gets stuck

**Solutions:**

```bash
# 1. Check migration status
migrate -path db/migrations -database "postgresql://..." version

# 2. Force version (if stuck)
migrate -path db/migrations -database "postgresql://..." force <version>

# 3. Drop and recreate database
docker exec -it web-app-postgres psql -U postgres
DROP DATABASE webapp_dev;
CREATE DATABASE webapp_dev;
\q

# 4. Run migrations fresh
migrate -path db/migrations -database "postgresql://..." up
```

---

### sqlc Generate Fails

**Error:** `sqlc generate` fails

**Solutions:**

```bash
# 1. Check sqlc.yaml syntax
cat sqlc.yaml

# 2. Verify SQL syntax
# Run queries manually in psql to test

# 3. Update sqlc
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

# 4. Clear cache and regenerate
rm -rf db/sqlc
sqlc generate
```

---

### JWT Token Invalid

**Error:** `invalid token` or authentication fails

**Solutions:**

```go
// 1. Check JWT_SECRET is set
fmt.Println(os.Getenv("JWT_SECRET"))

// 2. Verify token format
// Should be: Bearer <token>

// 3. Check token expiration
claims, ok := token.Claims.(jwt.MapClaims)
exp := claims["exp"].(float64)
fmt.Println(time.Unix(int64(exp), 0)) // Should be in future

// 4. Debug token parsing
token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
    fmt.Println("Token:", token)  // Debug
    return []byte(jwtSecret), nil
})
```

---

## Database Issues (PostgreSQL)

### Too Many Connections

**Error:** `FATAL: sorry, too many clients already`

**Solution:**

```sql
-- Check current connections
SELECT count(*) FROM pg_stat_activity;

-- Kill idle connections
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE state = 'idle'
  AND pid <> pg_backend_pid();

-- Increase max connections (in postgresql.conf or docker)
max_connections = 100

-- Or use connection pooling
-- In Go: pgxpool with MaxConns
```

---

### Slow Queries

**Solution:**

```sql
-- 1. Check slow queries
SELECT query, mean_exec_time, calls
FROM pg_stat_statements
ORDER BY mean_exec_time DESC
LIMIT 10;

-- 2. Analyze query
EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'test@example.com';

-- 3. Add missing indexes
CREATE INDEX idx_users_email ON users(email);

-- 4. Update statistics
ANALYZE users;
```

---

### Migration Out of Sync

**Error:** Database schema doesn't match migrations

**Solution:**

```bash
# 1. Check migration status
migrate -path db/migrations -database "postgresql://..." version

# 2. See current schema
docker exec -it web-app-postgres psql -U postgres -d webapp_dev
\dt
\d users

# 3. Reset to clean state
migrate -path db/migrations -database "postgresql://..." down
migrate -path db/migrations -database "postgresql://..." up

# 4. Or force specific version
migrate -path db/migrations -database "postgresql://..." force 1
```

---

## Docker Issues

### Container Won't Start

**Error:** `Error starting userland proxy: listen tcp4 0.0.0.0:5432: bind: address already in use`

**Solution:**

```bash
# 1. Check what's using the port
lsof -i :5432  # macOS/Linux
netstat -ano | findstr :5432  # Windows

# 2. Stop conflicting process
# If local PostgreSQL is running:
brew services stop postgresql  # macOS
sudo service postgresql stop   # Linux

# 3. Or change port in docker-compose.yml
ports:
  - "5433:5432"  # Use 5433 on host

# 4. Remove and recreate
docker-compose down -v
docker-compose up -d
```

---

### Volume Permission Issues

**Error:** Permission denied errors

**Solution:**

```bash
# 1. Stop containers
docker-compose down

# 2. Remove volumes
docker volume ls
docker volume rm <volume_name>

# 3. Restart
docker-compose up -d

# 4. For Linux, check user permissions
sudo chown -R $USER:$USER ./
```

---

### Build Cache Issues

**Error:** Docker build using old code

**Solution:**

```bash
# 1. Build without cache
docker-compose build --no-cache

# 2. Remove images
docker image ls
docker image rm <image_id>

# 3. Prune everything
docker system prune -a --volumes

# 4. Rebuild
docker-compose up --build
```

---

## Development Workflow Issues

### Hot Reload Not Working

**Frontend (Next.js):**

```bash
# 1. Check if dev mode
npm run dev  # Not 'npm start'

# 2. Clear .next cache
rm -rf .next
npm run dev

# 3. Check file watcher limits (Linux)
echo fs.inotify.max_user_watches=524288 | sudo tee -a /etc/sysctl.conf
sudo sysctl -p
```

**Backend (Air):**

```bash
# 1. Check .air.toml exists
cat .air.toml

# 2. Verify Air is watching correct files
# In .air.toml:
include_ext = ["go", "tpl", "tmpl"]
exclude_dir = ["tmp", "vendor"]

# 3. Restart Air
pkill air
air
```

---

### Environment Variables Not Loading

**Solution:**

```bash
# 1. Check .env file exists
ls -la | grep .env

# 2. Verify .env is loaded
# Go:
godotenv.Load()
fmt.Println(os.Getenv("DATABASE_URL"))

# Next.js:
console.log(process.env.NEXT_PUBLIC_API_URL)

# 3. Check .env syntax (no quotes for simple values)
DATABASE_URL=postgresql://localhost:5432/db
PORT=8080

# 4. Restart servers after changing .env
```

---

## Testing Issues

### Tests Can't Connect to Database

**Solution:**

```go
//go:build integration

// Use testcontainers
postgresContainer, err := postgres.RunContainer(ctx,
    postgres.WithDatabase("testdb"),
    postgres.WithUsername("postgres"),
    postgres.WithPassword("postgres"),
)
```

---

### Tests Timeout

**Solution:**

```go
// Increase timeout
func TestSomething(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Your test...
}
```

---

## Performance Issues

### Slow API Responses

**Debug Steps:**

```go
// 1. Add timing middleware
func TimingMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        c.Next()
        duration := time.Since(start)

        if duration > 1*time.Second {
            log.Printf("Slow request: %s %s took %v",
                c.Request.Method,
                c.Request.URL.Path,
                duration,
            )
        }
    }
}

// 2. Profile code
import _ "net/http/pprof"

go func() {
    http.ListenAndServe("localhost:6060", nil)
}()

// Visit http://localhost:6060/debug/pprof/

// 3. Add indexes
CREATE INDEX idx_table_column ON table(column);

// 4. Use connection pooling
// Check pgxpool configuration
```

---

## General Debug Tips

### Enable Debug Logging

**Go:**

```go
// Development logger
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))
```

**Gin:**

```go
// Debug mode
gin.SetMode(gin.DebugMode)

// Or use Default() instead of New()
router := gin.Default()  // Includes logger and recovery
```

---

### Check All Services Are Running

```bash
# Quick health check script
#!/bin/bash

echo "Checking services..."

# Frontend
curl -f http://localhost:3000 > /dev/null 2>&1 && echo "✓ Frontend running" || echo "✗ Frontend down"

# Backend
curl -f http://localhost:8080/health > /dev/null 2>&1 && echo "✓ Backend running" || echo "✗ Backend down"

# PostgreSQL
docker exec web-app-postgres pg_isready -U postgres > /dev/null 2>&1 && echo "✓ Database running" || echo "✗ Database down"
```

---

## Getting More Help

If you're still stuck:

1. **Check logs:**
   ```bash
   # Frontend
   npm run dev  # Watch console output

   # Backend
   go run cmd/api/main.go  # Watch logs

   # Database
   docker-compose logs postgres
   ```

2. **Search error messages:**
   - Copy exact error to Google
   - Search Stack Overflow
   - Check GitHub issues

3. **Ask for help:**
   - Stack Overflow (with specific error + minimal code)
   - Reddit (r/golang, r/nextjs)
   - Discord communities
   - GitHub Discussions

4. **Debugging checklist:**
   - [ ] All services running?
   - [ ] Environment variables set?
   - [ ] Database migrated?
   - [ ] Dependencies installed?
   - [ ] Ports not conflicting?
   - [ ] Checked logs?

Remember: Most issues have been solved before - search first!
