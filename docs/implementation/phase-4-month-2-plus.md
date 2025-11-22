# Phase 4: Advanced Features & Scaling (Month 2+)

**Goal:** Add advanced features, optimize performance, and prepare for scale.

**Topics covered:**
- Analytics data collection and visualization
- React Query for data fetching
- Caching with Redis
- Background jobs
- Performance optimization
- Deployment strategies

---

## Analytics Implementation

### Backend: Analytics Events

**Migration:**

```sql
-- 000002_create_analytics_events.up.sql
CREATE TABLE analytics_events (
    id BIGSERIAL,
    user_id UUID REFERENCES users(id),
    event_type VARCHAR(50) NOT NULL,
    event_data JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

CREATE TABLE analytics_events_2025_01
PARTITION OF analytics_events
FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

CREATE INDEX idx_analytics_user_time ON analytics_events(user_id, created_at DESC);
CREATE INDEX idx_analytics_type ON analytics_events(event_type);
```

**sqlc Queries (`db/queries/analytics.sql`):**

```sql
-- name: RecordEvent :one
INSERT INTO analytics_events (user_id, event_type, event_data)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserEvents :many
SELECT * FROM analytics_events
WHERE user_id = $1
  AND created_at >= $2
ORDER BY created_at DESC
LIMIT $3;

-- name: GetEventCounts :one
SELECT
    COUNT(*) as total_events,
    COUNT(DISTINCT user_id) as unique_users,
    COUNT(DISTINCT event_type) as event_types
FROM analytics_events
WHERE created_at >= $1;
```

**Service Implementation:**

```go
type analyticsService struct {
    repo ports.AnalyticsRepository
}

func (s *analyticsService) RecordEvent(ctx context.Context, userID uuid.UUID, eventType string, data map[string]interface{}) error {
    return s.repo.RecordEvent(ctx, userID, eventType, data)
}

func (s *analyticsService) GetMetrics(ctx context.Context, period string) (*domain.Metrics, error) {
    // Calculate time range based on period
    startTime := time.Now().Add(-30 * 24 * time.Hour) // Last 30 days

    return s.repo.GetMetrics(ctx, startTime)
}
```

### Frontend: Charts with Recharts

**Install:**

```bash
npm install recharts
npm install @tanstack/react-query
```

**React Query Setup (`app/providers.tsx`):**

```typescript
'use client';

import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useState } from 'react';

export function Providers({ children }: { children: React.ReactNode }) {
  const [queryClient] = useState(() => new QueryClient());

  return (
    <QueryClientProvider client={queryClient}>
      {children}
    </QueryClientProvider>
  );
}
```

**Update Root Layout:**

```typescript
import { Providers } from './providers';

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html>
      <body>
        <Providers>
          {children}
        </Providers>
      </body>
    </html>
  );
}
```

**Chart Component:**

```typescript
'use client';

import { useQuery } from '@tanstack/react-query';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend } from 'recharts';
import { api } from '@/lib/api';

export function AnalyticsChart() {
  const { data, isLoading } = useQuery({
    queryKey: ['analytics', 'daily'],
    queryFn: () => api.get('/api/v1/analytics/daily').then(res => res.data),
    refetchInterval: 60000, // Refetch every minute
  });

  if (isLoading) return <div>Loading...</div>;

  return (
    <LineChart width={600} height={300} data={data}>
      <CartesianGrid strokeDasharray="3 3" />
      <XAxis dataKey="date" />
      <YAxis />
      <Tooltip />
      <Legend />
      <Line type="monotone" dataKey="events" stroke="#8884d8" />
      <Line type="monotone" dataKey="users" stroke="#82ca9d" />
    </LineChart>
  );
}
```

---

## Caching with Redis

### Setup Redis

**Add to docker-compose.yml:**

```yaml
redis:
  image: redis:7-alpine
  ports:
    - "6379:6379"
  volumes:
    - redis_data:/data

volumes:
  redis_data:
```

### Backend Integration

**Install:**

```bash
go get github.com/redis/go-redis/v9
```

**Cache Service:**

```go
package cache

import (
    "context"
    "encoding/json"
    "time"

    "github.com/redis/go-redis/v9"
)

type RedisCache struct {
    client *redis.Client
}

func NewRedisCache(addr string) *RedisCache {
    client := redis.NewClient(&redis.Options{
        Addr: addr,
    })
    return &RedisCache{client: client}
}

func (c *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
    val, err := c.client.Get(ctx, key).Result()
    if err != nil {
        return err
    }
    return json.Unmarshal([]byte(val), dest)
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
    return c.client.Del(ctx, key).Err()
}
```

**Use in Service:**

```go
func (s *userService) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
    cacheKey := fmt.Sprintf("user:%s", id.String())

    // Try cache first
    var user domain.User
    if err := s.cache.Get(ctx, cacheKey, &user); err == nil {
        return &user, nil
    }

    // Cache miss, get from database
    user, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // Store in cache
    s.cache.Set(ctx, cacheKey, user, 5*time.Minute)

    return user, nil
}
```

---

## Background Jobs

**Worker Example:**

```go
package worker

import (
    "context"
    "log"
    "time"
)

type DailyMetricsWorker struct {
    analyticsService ports.AnalyticsService
}

func (w *DailyMetricsWorker) Run(ctx context.Context) {
    ticker := time.NewTicker(24 * time.Hour)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            if err := w.calculateDailyMetrics(ctx); err != nil {
                log.Printf("Error calculating daily metrics: %v", err)
            }
        case <-ctx.Done():
            return
        }
    }
}

func (w *DailyMetricsWorker) calculateDailyMetrics(ctx context.Context) error {
    // Calculate and store metrics
    // ...
    return nil
}
```

**Start Worker in main.go:**

```go
func main() {
    // ... existing setup ...

    // Start background workers
    ctx := context.Background()
    worker := worker.NewDailyMetricsWorker(analyticsService)
    go worker.Run(ctx)

    // Start server
    router.Run(":8080")
}
```

---

## Performance Optimization

### Database Connection Pooling

```go
config, _ := pgxpool.ParseConfig(dbURL)
config.MaxConns = 25
config.MinConns = 5
config.MaxConnLifetime = 1 * time.Hour
config.MaxConnIdleTime = 30 * time.Minute
config.HealthCheckPeriod = 1 * time.Minute

pool, _ := pgxpool.NewWithConfig(ctx, config)
```

### Frontend Optimizations

**Image Optimization:**

```typescript
import Image from 'next/image';

<Image
  src="/logo.png"
  alt="Logo"
  width={200}
  height={50}
  priority // For above-the-fold images
/>
```

**Code Splitting:**

```typescript
import dynamic from 'next/dynamic';

const Chart = dynamic(() => import('@/components/Chart'), {
  loading: () => <div>Loading chart...</div>,
  ssr: false, // Don't render on server
});
```

**React Query Optimizations:**

```typescript
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5 minutes
      cacheTime: 10 * 60 * 1000, // 10 minutes
      refetchOnWindowFocus: false,
    },
  },
});
```

---

## Deployment

### Environment Variables

**Production .env:**

```bash
# Backend
DATABASE_URL=postgresql://user:password@prod-db:5432/webapp_prod
JWT_SECRET=very-secure-random-string-here
REDIS_URL=redis://prod-redis:6379
PORT=8080
ENV=production

# Frontend
NEXT_PUBLIC_API_URL=https://api.yourdomain.com
```

### CI/CD with GitHub Actions

**.github/workflows/deploy.yml:**

```yaml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.23'
      - name: Run tests
        run: |
          cd backend
          go test ./...

  deploy:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Deploy to production
        run: |
          # Your deployment commands
          # docker build, push, deploy
```

### Monitoring with Prometheus

**Add metrics to backend:**

```go
import "github.com/prometheus/client_golang/prometheus/promhttp"

router.GET("/metrics", gin.WrapH(promhttp.Handler()))
```

---

## Summary

**You've learned:**
- Analytics implementation
- React Query for data fetching
- Redis caching
- Background workers
- Performance optimization
- Deployment strategies

**Your app is now production-ready!** 🚀

Continue exploring advanced topics like:
- Kubernetes deployment
- Microservices extraction
- GraphQL API
- WebSockets for real-time features
- Advanced monitoring (Datadog, New Relic)
