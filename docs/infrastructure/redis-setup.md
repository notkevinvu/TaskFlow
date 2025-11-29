# Redis Setup for TaskFlow

## Overview

TaskFlow uses Redis for **optional** distributed rate limiting. If Redis is unavailable, the application automatically falls back to in-memory rate limiting (suitable for single-instance deployments).

## When Do You Need Redis?

- **Single Instance (Development/Small Scale)**: Redis is **optional**. In-memory rate limiting works fine.
- **Multiple Instances (Production/High Scale)**: Redis is **recommended** to share rate limit counters across instances.

## Development Setup

### Option 1: No Redis (Default)

The application works out of the box without Redis. Rate limiting is handled in-memory.

**Pros:**
- Zero setup required
- Works immediately

**Cons:**
- Rate limits are per-instance (not shared across multiple backend instances)
- Rate limit state is lost on server restart

### Option 2: Local Redis (Optional)

If you want to test Redis-backed rate limiting locally:

#### Windows (with Chocolatey)
```bash
choco install redis-64
redis-server
```

#### macOS (with Homebrew)
```bash
brew install redis
brew services start redis
```

#### Linux
```bash
sudo apt-get install redis-server
sudo systemctl start redis
```

#### Environment Variable
Add to `backend/.env`:
```
REDIS_URL=localhost:6379
```

## Production Setup

For production deployments with multiple backend instances, use a managed Redis service:

### Managed Redis Options

1. **Redis Cloud** (Free tier available)
   - https://redis.com/try-free/
   - 30MB free tier
   - Add to `.env`: `REDIS_URL=redis-xxxxx.cloud.redislabs.com:12345`

2. **Upstash** (Serverless Redis)
   - https://upstash.com/
   - Free tier with 10K commands/day
   - Add to `.env`: `REDIS_URL=xxxxx.upstash.io:6379`

3. **AWS ElastiCache** (if using AWS)
   - Managed Redis service
   - Add to `.env`: `REDIS_URL=xxxxx.cache.amazonaws.com:6379`

4. **Self-Hosted Redis** (Docker/VM)
   - Deferred to Phase 4 (Docker infrastructure planning)

## Verification

Check logs on startup:

**With Redis:**
```
Successfully connected to database
Successfully connected to Redis for rate limiting
```

**Without Redis (fallback):**
```
Successfully connected to database
Warning: Unable to connect to Redis: ... (rate limiting disabled)
```

## Rate Limiting Behavior

### With Redis
- Rate limits are shared across all backend instances
- Limits persist across server restarts
- Sliding window algorithm for accurate counting

### Without Redis (In-Memory)
- Rate limits are per-instance (each instance has its own counters)
- Limits are reset when server restarts
- Token bucket algorithm with cleanup

## Configuration

Rate limiting is configured via environment variables:

```bash
# In backend/.env
RATE_LIMIT_REQUESTS_PER_MINUTE=100  # Default: 100 requests per minute per user/IP
REDIS_URL=localhost:6379             # Optional - defaults to localhost:6379
```

## Troubleshooting

### Redis connection errors

**Symptom:** `Warning: Unable to connect to Redis` on startup

**Solution:** This is expected if Redis is not running. The application will use in-memory rate limiting. No action needed unless you want Redis.

### Rate limits not working across instances

**Symptom:** Each backend instance allows 100 req/min (total 200 req/min with 2 instances)

**Solution:** Set up Redis to share rate limit counters across instances.

## Future Enhancements (Phase 4)

- Docker Compose setup for local Redis
- Redis password authentication
- Redis Sentinel for high availability
- Redis Cluster for horizontal scaling
- Monitoring Redis metrics in observability dashboard
