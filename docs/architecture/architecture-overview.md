# Web Application Architecture Overview

**For:** CRUD + Analytics Dashboard Application
**Scale:** Medium/Startup MVP
**Priorities:** Learning Experience, Fast Development, Production-Ready Patterns

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Technology Stack](#technology-stack)
3. [Architecture Decisions](#architecture-decisions)
4. [System Design](#system-design)
5. [Evolution Path](#evolution-path)

---

## Executive Summary

This architecture is designed for a modern full-stack web application combining CRUD operations with analytics/dashboard capabilities. The stack balances learning opportunities (Go, PostgreSQL, Clean Architecture) with rapid development (Next.js, proven frameworks) and production-ready patterns.

### Key Architectural Principles

1. **Start Simple, Scale Smartly**: Begin with a modular monolith, evolve to microservices only when justified
2. **Type Safety Everywhere**: TypeScript on frontend, sqlc on backend
3. **Clean Boundaries**: Use Clean/Hexagonal Architecture to enable future refactoring
4. **Developer Experience**: Hot reload, clear patterns, excellent tooling
5. **Production-Ready from Day 1**: Auth, logging, testing, containerization

---

## Technology Stack

### Frontend

| Component | Technology | Why |
|-----------|-----------|-----|
| **Framework** | Next.js 15 (App Router) | SSR for dashboards, built-in API routes, file-based routing |
| **Language** | TypeScript | Type safety, better DX, fewer runtime errors |
| **UI Components** | Shadcn/UI | Full customization, copy into codebase, accessible |
| **Styling** | Tailwind CSS | Utility-first, fast development, small bundle |
| **Server State** | React Query (TanStack Query) | Caching, background updates, perfect for dashboards |
| **Client State** | Zustand | Simple, 90% lighter than Redux, SSR-friendly |
| **Charts/Analytics** | Recharts + TanStack Table | Type-safe, composable, production-ready |

### Backend

| Component | Technology | Why |
|-----------|-----------|-----|
| **Language** | Go 1.23+ | Performance, concurrency, learning goal |
| **Web Framework** | Gin | Mature, fast, great docs, large community |
| **Architecture** | Clean/Hexagonal | Testable, maintainable, enables future microservices |
| **API Style** | REST (gRPC later) | Simpler to start, add gRPC when needed |
| **Database Access** | sqlc | Type-safe, fast, learn SQL properly |
| **Validation** | Go validator v10 | Declarative struct tags |
| **Hot Reload** | Air | Automatic rebuild on file changes |

### Database & Infrastructure

| Component | Technology | Why |
|-----------|-----------|-----|
| **Database** | PostgreSQL 16 | Robust, JSON support, excellent for analytics |
| **Driver** | pgx/pgxpool | 30-50% faster than database/sql, native features |
| **Migrations** | golang-migrate | Mature, simple, language-agnostic |
| **Time-Series (Optional)** | TimescaleDB | PostgreSQL extension for advanced analytics |
| **Containerization** | Docker + docker-compose | Reproducible environments, easy deployment |
| **Configuration** | Viper | Env vars + config files, 12-factor compliant |

### Observability & DevOps

| Component | Technology | Why |
|-----------|-----------|-----|
| **Logging** | slog (Go stdlib) | Structured logging, zero dependencies |
| **Tracing** | OpenTelemetry | Industry standard, future-proof |
| **Metrics** | Prometheus | De facto standard for Go apps |
| **Visualization** | Grafana | Powerful dashboards for logs/metrics |
| **Testing** | testify + testcontainers | Comprehensive testing at all levels |

---

## Architecture Decisions

### 1. Next.js vs Plain React

**Decision: Next.js 15**

**Rationale:**
- **SSR Benefits**: Faster initial load for dashboards with data
- **API Routes**: Built-in BFF (Backend for Frontend) pattern
- **File-Based Routing**: Less configuration, clearer structure
- **App Router**: Modern patterns with Server Components
- **DX**: Zero-config TypeScript, fast refresh, excellent docs

**Trade-offs:**
- ✅ Better for SEO, performance, DX
- ✅ Integrated API layer
- ⚠️ Slightly more to learn than plain React
- ⚠️ Vendor lock-in to Vercel ecosystem (but can self-host)

### 2. Gin vs Other Go Frameworks

**Decision: Gin**

**Comparison:**

```
Gin:        Fast, mature, huge community, intuitive API
Echo:       Similar to Gin, better for enterprise scale
Fiber:      Fastest, but incompatible with standard middleware
Chi:        Minimal, stays close to stdlib
Stdlib:     Maximum control, but more verbose
```

**Rationale:**
- 78k+ GitHub stars, proven in production
- Excellent documentation and tutorials
- Gentle learning curve for Go beginners
- Performance: 40x faster than older frameworks
- Large middleware ecosystem

### 3. sqlc vs GORM vs sqlx

**Decision: sqlc**

**Comparison:**

| Feature | sqlc | GORM | sqlx |
|---------|------|------|------|
| **Type Safety** | ✅✅✅ Compile-time | ⚠️ Runtime only | ⚠️ Minimal |
| **Performance** | ✅✅✅ Zero overhead | ⚠️ 3-5x slower | ✅✅ Fast |
| **Learning Curve** | ✅✅ Learn real SQL | ⚠️ Learn ORM API | ✅ Medium |
| **Control** | ✅✅✅ Full control | ⚠️ Magic/hidden queries | ✅✅ Good control |
| **Features** | ⚠️ Basic | ✅✅✅ Rich (hooks, associations) | ✅ Medium |

**Rationale:**
- Learn SQL properly (valuable skill)
- Best performance for MVP and beyond
- Type errors caught at compile time
- Generated code is readable
- Switch to GORM later if you need ORM features

### 4. Monolith vs Microservices

**Decision: Modular Monolith (with microservices path)**

**Starting Architecture:**
```
Single Deployment
├── User Module (domain/ports/adapters/services)
├── Analytics Module (domain/ports/adapters/services)
└── Auth Module (domain/ports/adapters/services)
```

**Rationale:**
- **Faster development**: Single codebase, simpler debugging
- **Lower costs**: One server, one database
- **Easier testing**: No distributed system complexity
- **Learn first**: Understand domain boundaries before splitting

**Evolution to Microservices:**
Only when you have:
- Team size >10-15 developers
- Clear bounded contexts with different scaling needs
- Need for independent deployments
- Performance bottlenecks in specific domains

### 5. REST vs gRPC

**Decision: REST first, add gRPC later**

**Phase 1 (MVP): REST Only**
```
Frontend (Next.js) → REST API → Backend (Go)
```

**Phase 2 (If scaling): REST + gRPC with grpc-gateway**
```
Frontend → REST (grpc-gateway) → gRPC Services
                                      ├── User Service
                                      ├── Analytics Service
                                      └── Auth Service
```

**Rationale:**
- REST is simpler to start, debug, and test
- Frontend developers are more familiar with REST
- Add gRPC when you split services (better performance for service-to-service)
- grpc-gateway gives you both REST and gRPC from same .proto files

---

## System Design

### High-Level Architecture (MVP)

```
┌─────────────────────────────────────────────────┐
│                   Frontend                       │
│  Next.js 15 (App Router) + TypeScript           │
│  - Server Components (default)                   │
│  - Client Components (interactive UI)           │
│  - API Routes (BFF layer if needed)             │
└───────────────────┬─────────────────────────────┘
                    │ HTTP/REST
                    │ JSON
                    ▼
┌─────────────────────────────────────────────────┐
│              Backend API Server                  │
│  Go + Gin Framework                             │
│  ┌─────────────────────────────────────────┐   │
│  │   Clean Architecture Layers             │   │
│  │  ┌────────────────────────────────┐    │   │
│  │  │  Handlers (HTTP/REST)          │    │   │
│  │  └──────────┬─────────────────────┘    │   │
│  │             │                            │   │
│  │  ┌──────────▼─────────────────────┐    │   │
│  │  │  Services (Business Logic)     │    │   │
│  │  └──────────┬─────────────────────┘    │   │
│  │             │                            │   │
│  │  ┌──────────▼─────────────────────┐    │   │
│  │  │  Repositories (Data Access)    │    │   │
│  │  │  - sqlc generated code         │    │   │
│  │  └──────────┬─────────────────────┘    │   │
│  └─────────────┼────────────────────────────   │
└────────────────┼────────────────────────────────┘
                 │ SQL
                 ▼
┌─────────────────────────────────────────────────┐
│          PostgreSQL 16 Database                 │
│  - Relational tables for CRUD                   │
│  - Partitioned tables for analytics             │
│  - JSONB columns for flexible data              │
│  - Indexes optimized for queries                │
└─────────────────────────────────────────────────┘
```

### Request Flow Example

**Fetching Dashboard Analytics:**

```
1. User navigates to /dashboard/analytics
   ├─> Next.js Server Component fetches initial data
   └─> Renders HTML with initial state

2. Next.js makes API call: GET /api/v1/analytics?period=30d
   └─> Proxies to Go backend OR calls backend directly

3. Go Backend (Gin Handler)
   ├─> Validates request parameters
   ├─> Calls AnalyticsService.GetMetrics()
   └─> Service calls AnalyticsRepository.FetchMetrics()

4. Repository (sqlc generated)
   ├─> Executes optimized SQL query
   ├─> Uses connection pool (pgxpool)
   └─> Returns typed data structures

5. Response flows back
   ├─> Repository → Service → Handler
   ├─> Handler formats JSON response
   └─> Client receives data

6. React Query (on frontend)
   ├─> Caches response
   ├─> Automatically refetches on interval
   └─> Updates UI reactively
```

### Data Flow & State Management

```
Server State (API Data):
  React Query
  ├── Automatic caching
  ├── Background refetching
  ├── Optimistic updates
  └── Loading/error states

Client State (UI):
  Zustand
  ├── Theme preferences
  ├── UI state (modals, sidebars)
  └── User session info

URL State:
  Next.js Router
  ├── Current page/route
  ├── Query parameters
  └── Filters/pagination
```

### Database Schema Design Patterns

**CRUD Tables (Normalized):**
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
```

**Analytics Tables (Partitioned):**
```sql
CREATE TABLE analytics_events (
    id BIGSERIAL,
    user_id UUID REFERENCES users(id),
    event_type VARCHAR(50) NOT NULL,
    event_data JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Monthly partitions
CREATE TABLE analytics_events_2025_01
PARTITION OF analytics_events
FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');
```

**Materialized Views for Dashboards:**
```sql
CREATE MATERIALIZED VIEW daily_user_activity AS
SELECT
    user_id,
    DATE(created_at) as activity_date,
    COUNT(*) as event_count,
    COUNT(DISTINCT event_type) as unique_events
FROM analytics_events
WHERE created_at > NOW() - INTERVAL '90 days'
GROUP BY user_id, DATE(created_at);

-- Refresh periodically
CREATE INDEX ON daily_user_activity(activity_date DESC);
```

---

## Evolution Path

### Phase 1: MVP Foundation (Months 1-2)

**Focus:** Get to production with core features

```
Frontend:
  ✓ Next.js setup with TypeScript
  ✓ Basic UI with Shadcn/UI
  ✓ Authentication flow
  ✓ Main dashboard pages

Backend:
  ✓ Go project structure
  ✓ REST API with Gin
  ✓ Database integration with sqlc
  ✓ JWT authentication

Infrastructure:
  ✓ Docker development environment
  ✓ PostgreSQL with migrations
  ✓ Basic logging
```

### Phase 2: Production Hardening (Months 3-4)

**Focus:** Reliability and observability

```
Observability:
  ✓ Structured logging (slog)
  ✓ OpenTelemetry tracing
  ✓ Prometheus metrics
  ✓ Grafana dashboards

Testing:
  ✓ Unit tests (70% coverage)
  ✓ Integration tests with testcontainers
  ✓ E2E tests with Playwright

Performance:
  ✓ Query optimization
  ✓ Caching strategy (Redis)
  ✓ Database indexes
  ✓ Connection pooling tuning
```

### Phase 3: Scaling Preparation (Months 5-6)

**Focus:** Handle growth

```
Architecture:
  ✓ Evaluate microservices need
  ✓ Introduce gRPC for internal services
  ✓ Implement grpc-gateway
  ✓ API Gateway pattern

Infrastructure:
  ✓ Kubernetes setup
  ✓ CI/CD pipeline
  ✓ Multi-environment (dev/staging/prod)
  ✓ Database replication
```

### Phase 4: Advanced Features (Month 7+)

**Focus:** Scale and optimize

```
Microservices (if needed):
  ✓ Extract analytics service
  ✓ Extract auth service
  ✓ Service mesh (Istio/Linkerd)

Advanced Observability:
  ✓ Distributed tracing
  ✓ APM (Datadog/New Relic)
  ✓ Error tracking (Sentry)

Performance:
  ✓ Multi-region deployment
  ✓ CDN for static assets
  ✓ Database sharding (if needed)
```

### Decision Points

**When to introduce each technology:**

| Technology | Trigger | Benefit |
|------------|---------|---------|
| **Redis Cache** | API response time >500ms | 10-100x faster reads |
| **gRPC** | Split to microservices | Efficient service communication |
| **Kubernetes** | >3 services to orchestrate | Auto-scaling, self-healing |
| **Service Mesh** | >5 microservices | Observability, security, traffic management |
| **GraphQL** | Complex data requirements | Flexible queries, reduce over-fetching |
| **Event Streaming** | Need async processing | Kafka/NATS for event-driven architecture |

---

## Key Takeaways

1. **Start with a modular monolith** using Clean Architecture - this gives you flexibility to evolve
2. **Use proven technologies** (Next.js, Gin, PostgreSQL) rather than bleeding-edge for faster development
3. **Type safety everywhere** (TypeScript + sqlc) catches errors early
4. **Production patterns from day 1** (auth, logging, testing, Docker) but don't over-engineer
5. **Defer complexity** (gRPC, microservices, Kubernetes) until you have clear need

The architecture is designed to maximize your learning while building something production-ready. You'll learn Go, modern frontend development, database design, and system architecture - while having clear paths to evolve as requirements change.

---

**Next Steps:**
- Read `tech-stack-explained.md` for deep dives into each technology
- Follow `phase-1-weeks-1-2.md` to start implementation
- Refer to `common-patterns.md` for code examples
- Check `troubleshooting.md` when you hit issues
