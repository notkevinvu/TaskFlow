# Quick Start Guide

Get up and running in 15 minutes!

---

## Prerequisites

```bash
# Install required software
- Node.js 20+ (https://nodejs.org/)
- Go 1.23+ (https://go.dev/dl/)
- Docker Desktop (https://www.docker.com/products/docker-desktop/)
- Git
```

---

## 1. Clone or Create Project

```bash
mkdir web-app && cd web-app
```

---

## 2. Frontend Setup (5 minutes)

```bash
# Create Next.js app
npx create-next-app@latest frontend
# Choose: TypeScript ✓, ESLint ✓, Tailwind ✓, App Router ✓

cd frontend

# Install Shadcn/UI
npx shadcn@latest init
# Choose: Default style, Slate color, CSS variables ✓

# Add components
npx shadcn@latest add button card input label

# Start dev server
npm run dev
```

Visit http://localhost:3000

---

## 3. Database Setup (2 minutes)

Create `docker-compose.yml` in project root:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: webapp_dev
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

```bash
# Start database
docker-compose up -d
```

---

## 4. Backend Setup (8 minutes)

```bash
cd ../
mkdir backend && cd backend

# Initialize Go module
go mod init github.com/yourusername/webapp

# Install dependencies
go get github.com/gin-gonic/gin
go get github.com/jackc/pgx/v5/pgxpool
go get github.com/golang-jwt/jwt/v5
go get golang.org/x/crypto/bcrypt

# Create structure
mkdir -p cmd/api internal/{domain,ports,adapters/handlers,services} config
```

Create `cmd/api/main.go`:

```go
package main

import (
    "github.com/gin-gonic/gin"
)

func main() {
    router := gin.Default()

    router.GET("/health", func(c *gin.Context) {
        c.JSON(200, gin.H{"status": "ok"})
    })

    router.Run(":8080")
}
```

```bash
# Run backend
go run cmd/api/main.go
```

Visit http://localhost:8080/health

---

## 5. Verify Everything Works

**Terminal 1 - Database:**
```bash
docker-compose up
```

**Terminal 2 - Backend:**
```bash
cd backend
go run cmd/api/main.go
```

**Terminal 3 - Frontend:**
```bash
cd frontend
npm run dev
```

**Check:**
- Frontend: http://localhost:3000 ✓
- Backend: http://localhost:8080/health ✓
- Database: `docker ps` shows postgres running ✓

---

## Next Steps

1. Follow `phase-1-weeks-1-2.md` for detailed frontend setup
2. Follow `phase-2-weeks-3-4.md` for backend implementation
3. Read `architecture-overview.md` for design decisions
4. Check `common-patterns.md` for code examples

---

## Useful Commands

```bash
# Frontend
npm run dev        # Start dev server
npm run build      # Build for production
npm run lint       # Run linter

# Backend
go run cmd/api/main.go  # Run server
go test ./...            # Run tests
go mod tidy              # Clean dependencies

# Database
docker-compose up -d        # Start in background
docker-compose down         # Stop
docker-compose logs postgres # View logs
docker exec -it web-app-postgres psql -U postgres -d webapp_dev  # Connect

# All together (from project root)
docker-compose up  # Terminal 1
cd backend && go run cmd/api/main.go  # Terminal 2
cd frontend && npm run dev  # Terminal 3
```

---

## Troubleshooting

**Port already in use:**
```bash
# Find and kill process
lsof -i :3000  # macOS/Linux
netstat -ano | findstr :3000  # Windows
```

**Database connection failed:**
```bash
docker-compose down -v  # Remove volumes
docker-compose up       # Restart
```

**Module not found:**
```bash
# Frontend
rm -rf node_modules package-lock.json && npm install

# Backend
go mod tidy && go mod download
```

---

You're ready to build! 🚀
