# Observability Infrastructure Research

**Status:** Research Complete | **Decision:** Pending
**Last Updated:** 2024-12-06

---

## Executive Summary

TaskFlow currently uses Go's `log/slog` for structured logging, outputting JSON to stdout. This document evaluates options for centralizing logs for production observability, comparing self-hosted and cloud solutions.

**Recommendation:** Start with **Grafana Cloud Free Tier** for simplicity, with a migration path to self-hosted Loki if needed.

---

## Current State

### What We Have
- **Logger:** Go 1.21+ `log/slog` with structured JSON output
- **Output:** stdout/stderr (visible in terminal)
- **Fields:** Contextual data like `user_id`, `task_id`, `error`
- **Levels:** DEBUG, INFO, WARN, ERROR

### What's Missing
- Centralized log aggregation
- Log search and filtering UI
- Log retention and rotation
- Alerting on error patterns
- Correlation with traces/metrics

---

## Options Evaluated

### Option 1: Grafana Cloud Free Tier

**Architecture:**
```
┌─────────────┐     ┌─────────────────┐     ┌─────────────┐
│ Go Backend  │────▶│ Grafana Cloud   │────▶│  Grafana UI │
│ (HTTP push) │     │ (Loki hosted)   │     │  (hosted)   │
└─────────────┘     └─────────────────┘     └─────────────┘
```

**Free Tier Limits:**
| Resource | Limit |
|----------|-------|
| Logs | 50 GB/month |
| Metrics | 10,000 active series |
| Traces | 50 GB/month |
| Retention | 14 days |

**Implementation:**
```go
// Option A: Direct Loki client
import "github.com/grafana/loki-client-go/loki"

func setupLogging() {
    cfg := loki.Config{
        URL:       "https://logs-prod-us-central1.grafana.net/loki/api/v1/push",
        BatchWait: time.Second,
        BatchSize: 100 * 1024,
    }
    client, _ := loki.New(cfg)
    // Use as slog handler or io.Writer
}

// Option B: OpenTelemetry bridge (recommended for future tracing)
import "go.opentelemetry.io/contrib/bridges/otelslog"

handler := otelslog.NewHandler("taskflow-backend")
slog.SetDefault(slog.New(handler))
// Configure OTLP exporter to Grafana Cloud
```

**Pros:**
- Zero infrastructure to manage
- No Docker/Kubernetes required
- 50GB/month is plenty for dev/small production
- Same query experience as self-hosted Loki
- Easy upgrade path to paid tiers

**Cons:**
- 14-day retention only on free tier
- Requires internet connectivity
- Data leaves your infrastructure

**Cost:** $0 (free tier) → $0.50/GB after 50GB

**Best For:** Getting started quickly, small teams, development

---

### Option 2: PLG Stack (Promtail + Loki + Grafana)

**Architecture:**
```
┌─────────────┐     ┌───────────┐     ┌──────────┐     ┌─────────────┐
│ Go Backend  │────▶│  Promtail │────▶│   Loki   │────▶│   Grafana   │
│ (stdout)    │     │  (agent)  │     │ (storage)│     │    (UI)     │
└─────────────┘     └───────────┘     └──────────┘     └─────────────┘
     JSON logs       tail & push      label index       dashboards
```

**Docker Compose:**
```yaml
# docker-compose.observability.yml
version: '3.8'

services:
  loki:
    image: grafana/loki:2.9.0
    ports:
      - "3100:3100"
    volumes:
      - loki-data:/loki
    command: -config.file=/etc/loki/local-config.yaml

  promtail:
    image: grafana/promtail:2.9.0
    volumes:
      - /var/log:/var/log:ro
      - ./promtail-config.yml:/etc/promtail/config.yml
    command: -config.file=/etc/promtail/config.yml
    depends_on:
      - loki

  grafana:
    image: grafana/grafana:10.0.0
    ports:
      - "3001:3000"  # Avoid conflict with Next.js
    environment:
      - GF_AUTH_ANONYMOUS_ENABLED=true
      - GF_AUTH_ANONYMOUS_ORG_ROLE=Admin
    volumes:
      - grafana-data:/var/lib/grafana
    depends_on:
      - loki

volumes:
  loki-data:
  grafana-data:
```

**Promtail Config:**
```yaml
# promtail-config.yml
server:
  http_listen_port: 9080

positions:
  filename: /tmp/positions.yaml

clients:
  - url: http://loki:3100/loki/api/v1/push

scrape_configs:
  - job_name: taskflow-backend
    static_configs:
      - targets:
          - localhost
        labels:
          job: taskflow
          __path__: /var/log/taskflow/*.log
    pipeline_stages:
      - json:
          expressions:
            level: level
            msg: msg
            user_id: user_id
            error: error
      - labels:
          level:
          user_id:
```

**Resource Requirements:**
| Component | RAM | CPU | Disk |
|-----------|-----|-----|------|
| Loki | 256-512 MB | 0.5 core | 1-10 GB |
| Promtail | 64-128 MB | 0.1 core | minimal |
| Grafana | 256-512 MB | 0.5 core | 100 MB |
| **Total** | ~500 MB - 1 GB | ~1 core | ~10 GB |

**Pros:**
- Full control over data
- No external dependencies
- Unlimited retention (disk-limited)
- Can run offline

**Cons:**
- Requires Docker
- Operational overhead (updates, backups)
- Must manage disk space

**Cost:** $0 (self-hosted) + infrastructure costs

**Best For:** Privacy requirements, learning, full control

---

### Option 3: SigNoz (Unified Observability)

**Architecture:**
```
┌─────────────┐     ┌──────────────────────────────────────┐
│ Go Backend  │────▶│            SigNoz                    │
│ + OTel SDK  │     │  ┌──────┐  ┌────────┐  ┌─────────┐  │
└─────────────┘     │  │ Logs │  │ Traces │  │ Metrics │  │
                    │  └──┬───┘  └───┬────┘  └────┬────┘  │
                    │     └──────────┴───────────┘        │
                    │           ClickHouse                 │
                    └──────────────────────────────────────┘
```

**Docker Compose:**
```bash
# Clone and run SigNoz
git clone -b main https://github.com/SigNoz/signoz.git
cd signoz/deploy
docker-compose -f docker/clickhouse-setup/docker-compose.yaml up -d
```

**Go Integration:**
```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/exporters/otlp/otlptrace"
    "go.opentelemetry.io/contrib/bridges/otelslog"
)

func setupObservability() {
    // Configure OTLP exporter to SigNoz
    exporter, _ := otlptrace.New(ctx,
        otlptrace.WithEndpoint("localhost:4317"),
        otlptrace.WithInsecure(),
    )

    tp := trace.NewTracerProvider(
        trace.WithBatcher(exporter),
    )
    otel.SetTracerProvider(tp)

    // Bridge slog to OpenTelemetry
    handler := otelslog.NewHandler("taskflow")
    slog.SetDefault(slog.New(handler))
}
```

**Resource Requirements:**
| Component | RAM | CPU |
|-----------|-----|-----|
| ClickHouse | 2-4 GB | 2 cores |
| Query Service | 512 MB | 0.5 core |
| Frontend | 256 MB | 0.5 core |
| OTel Collector | 256 MB | 0.5 core |
| **Total** | ~3-5 GB | ~3-4 cores |

**Pros:**
- Logs, traces, AND metrics in one UI
- Correlate logs with request traces
- OpenTelemetry native (vendor-neutral)
- Fast ClickHouse queries
- Datadog-like experience, open-source

**Cons:**
- Heavy resource usage
- Complex deployment
- Steeper learning curve

**Cost:** $0 (self-hosted) | SigNoz Cloud available

**Best For:** Production systems, microservices, full observability

---

### Option 4: Cloud Alternatives

| Provider | Free Tier | Pros | Cons |
|----------|-----------|------|------|
| **Axiom** | 500 GB/month | Generous free tier | Less mature |
| **Better Stack** | 1 GB/month | Beautiful UI | Limited free tier |
| **Datadog** | 14-day trial | Industry standard | Expensive |
| **New Relic** | 100 GB/month | Full APM | Complex |

---

## Comparison Matrix

| Criteria | Grafana Cloud | PLG Stack | SigNoz |
|----------|---------------|-----------|--------|
| **Setup Time** | 30 min | 2-3 hours | 3-4 hours |
| **Infrastructure** | None | Docker | Docker |
| **RAM Required** | 0 | ~1 GB | ~4 GB |
| **Tracing Support** | Add-on | Separate | Built-in |
| **Learning Curve** | Low | Medium | Medium |
| **Data Control** | Cloud | Full | Full |
| **Best For** | Quick start | Privacy | Full observability |

---

## Implementation Plan

### Phase 1: Quick Win (Week 1)
1. Sign up for Grafana Cloud free tier
2. Add Loki client to Go backend
3. Create basic dashboard for error monitoring
4. Set up alert for ERROR level logs

### Phase 2: Enhanced Logging (Week 2-3)
1. Add request tracing IDs to all logs
2. Create structured log format guidelines
3. Add log context middleware to Gin
4. Build dashboards for:
   - Error rate by endpoint
   - Slow requests
   - User activity patterns

### Phase 3: Full Observability (Optional, Week 4+)
1. Migrate to SigNoz or self-hosted Loki
2. Add OpenTelemetry tracing
3. Correlate logs with traces
4. Add custom metrics

---

## Code Changes Required

### 1. Add Logging Middleware (Gin)

```go
// internal/middleware/logging.go
func RequestLogger() gin.HandlerFunc {
    return func(c *gin.Context) {
        requestID := uuid.New().String()
        c.Set("request_id", requestID)

        start := time.Now()
        c.Next()
        duration := time.Since(start)

        slog.Info("request completed",
            "request_id", requestID,
            "method", c.Request.Method,
            "path", c.Request.URL.Path,
            "status", c.Writer.Status(),
            "duration_ms", duration.Milliseconds(),
            "user_id", c.GetString("user_id"),
        )
    }
}
```

### 2. Context-Aware Logging

```go
// internal/service/task_service.go
func (s *TaskService) Create(ctx context.Context, userID string, dto *CreateTaskDTO) (*Task, error) {
    logger := slog.With(
        "user_id", userID,
        "request_id", ctx.Value("request_id"),
    )

    logger.Info("creating task", "title", dto.Title)

    task, err := s.taskRepo.Create(ctx, task)
    if err != nil {
        logger.Error("failed to create task", "error", err)
        return nil, err
    }

    logger.Info("task created", "task_id", task.ID)
    return task, nil
}
```

### 3. Grafana Cloud Integration

```go
// cmd/server/main.go
import "github.com/grafana/loki-client-go/loki"

func setupLogging() *loki.Client {
    if os.Getenv("LOKI_URL") == "" {
        return nil // Fall back to stdout
    }

    cfg := loki.Config{
        URL:       os.Getenv("LOKI_URL"),
        BatchWait: 1 * time.Second,
        BatchSize: 100 * 1024,
        Client: loki.Client{
            BasicAuth: &loki.BasicAuth{
                Username: os.Getenv("LOKI_USER"),
                Password: os.Getenv("LOKI_API_KEY"),
            },
        },
    }

    client, err := loki.New(cfg)
    if err != nil {
        slog.Warn("failed to create Loki client", "error", err)
        return nil
    }

    return client
}
```

---

## Decision Criteria

| If you need... | Choose... |
|----------------|-----------|
| Fastest setup, no infra | Grafana Cloud Free |
| Full data control, privacy | Self-hosted PLG |
| Logs + Traces + Metrics | SigNoz |
| Enterprise features | Datadog/New Relic |

---

## Next Steps

1. [ ] **Decision:** Choose initial approach (recommend Grafana Cloud)
2. [ ] **Implementation:** Add logging middleware to Gin
3. [ ] **Integration:** Connect to chosen log aggregator
4. [ ] **Dashboards:** Create error monitoring dashboard
5. [ ] **Alerts:** Set up critical error alerts
6. [ ] **Documentation:** Update CLAUDE.md with logging guidelines

---

## References

- [Go slog Package Documentation](https://pkg.go.dev/log/slog)
- [Grafana Loki Documentation](https://grafana.com/docs/loki/latest/)
- [SigNoz Documentation](https://signoz.io/docs/)
- [OpenTelemetry Go SDK](https://opentelemetry.io/docs/instrumentation/go/)
- [Better Stack: Logging in Go with Slog](https://betterstack.com/community/guides/logging/logging-in-go/)
