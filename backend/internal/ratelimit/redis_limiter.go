package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisLimiter implements rate limiting using Redis for horizontal scalability
// Uses a true sliding window algorithm with sorted sets
type RedisLimiter struct {
	client *redis.Client
}

// NewRedisLimiter creates a new Redis-backed rate limiter
func NewRedisLimiter(redisURL string) (*RedisLimiter, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         redisURL,
		Password:     "", // No password by default
		DB:           0,  // Use default DB
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisLimiter{client: client}, nil
}

// Allow checks if a request from the given identifier should be allowed
// Uses a true sliding window algorithm with Redis sorted sets
func (l *RedisLimiter) Allow(ctx context.Context, identifier string, limit int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("ratelimit:%s", identifier)
	now := time.Now()
	nowMs := now.UnixMilli()
	nowNano := now.UnixNano() // Use nanoseconds for unique member IDs
	windowMs := window.Milliseconds()
	windowStartMs := nowMs - windowMs

	// Lua script for atomic sliding window check
	// Uses sorted set where score = timestamp in milliseconds
	// Member uses nanoseconds for uniqueness (prevents collisions when multiple
	// requests arrive in the same millisecond)
	script := redis.NewScript(`
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local window_start = tonumber(ARGV[2])
		local now_ms = tonumber(ARGV[3])
		local window_seconds = tonumber(ARGV[4])
		local now_nano = ARGV[5]

		-- Remove requests older than the window
		redis.call('ZREMRANGEBYSCORE', key, 0, window_start)

		-- Count requests in current window
		local current = redis.call('ZCARD', key)

		if current < limit then
			-- Add current request with millisecond timestamp as score
			-- Use nanosecond timestamp as member for uniqueness
			redis.call('ZADD', key, now_ms, now_nano)
			-- Set expiration on the key (not EXPIREAT which uses absolute timestamp)
			redis.call('EXPIRE', key, window_seconds)
			return {1, current + 1}
		else
			return {0, current}
		end
	`)

	result, err := script.Run(ctx, l.client, []string{key}, limit, windowStartMs, nowMs, int(window.Seconds()), nowNano).Int64Slice()
	if err != nil {
		// Fail open - allow request if Redis is down
		return true, fmt.Errorf("redis error: %w", err)
	}

	return result[0] == 1, nil
}

// LimitInfo contains rate limit information for HTTP headers
type LimitInfo struct {
	Limit     int       // Total requests allowed in window
	Remaining int       // Requests remaining in current window
	ResetAt   time.Time // When the window resets (next second)
}

// GetLimitInfo returns current rate limit information for an identifier
func (l *RedisLimiter) GetLimitInfo(ctx context.Context, identifier string, limit int, window time.Duration) (*LimitInfo, error) {
	key := fmt.Sprintf("ratelimit:%s", identifier)
	now := time.Now()
	nowMs := now.UnixMilli()
	windowMs := window.Milliseconds()
	windowStartMs := nowMs - windowMs

	// Lua script to get current count without modifying
	script := redis.NewScript(`
		local key = KEYS[1]
		local window_start = tonumber(ARGV[1])

		-- Remove requests older than the window
		redis.call('ZREMRANGEBYSCORE', key, 0, window_start)

		-- Count requests in current window
		local current = redis.call('ZCARD', key)

		return current
	`)

	current, err := script.Run(ctx, l.client, []string{key}, windowStartMs).Int()
	if err != nil {
		return nil, fmt.Errorf("redis error: %w", err)
	}

	remaining := limit - current
	if remaining < 0 {
		remaining = 0
	}

	// Reset time is the next second (simplified for 1-minute windows)
	resetAt := now.Add(time.Second).Truncate(time.Second)

	return &LimitInfo{
		Limit:     limit,
		Remaining: remaining,
		ResetAt:   resetAt,
	}, nil
}

// Reset removes the rate limit counter for a given identifier
// Useful for testing or administrative purposes
func (l *RedisLimiter) Reset(ctx context.Context, identifier string) error {
	key := fmt.Sprintf("ratelimit:%s", identifier)
	return l.client.Del(ctx, key).Err()
}

// Health checks if the Redis connection is healthy
func (l *RedisLimiter) Health(ctx context.Context) error {
	return l.client.Ping(ctx).Err()
}

// Close closes the Redis connection
func (l *RedisLimiter) Close() error {
	return l.client.Close()
}
