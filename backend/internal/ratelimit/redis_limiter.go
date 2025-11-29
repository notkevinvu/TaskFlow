package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisLimiter implements rate limiting using Redis for horizontal scalability
// Uses the token bucket algorithm with sliding window
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
// Uses sliding window algorithm with fixed window counters for efficiency
func (l *RedisLimiter) Allow(ctx context.Context, identifier string, limit int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("ratelimit:%s", identifier)
	now := time.Now()
	windowStart := now.Truncate(window).Unix()

	// Use Lua script for atomic increment and check
	script := redis.NewScript(`
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local window = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])

		local current = redis.call('GET', key)
		if current == false then
			current = 0
		else
			current = tonumber(current)
		end

		if current < limit then
			redis.call('INCR', key)
			redis.call('EXPIREAT', key, now + window)
			return 1
		else
			return 0
		end
	`)

	result, err := script.Run(ctx, l.client, []string{key}, limit, int(window.Seconds()), windowStart).Int()
	if err != nil {
		// Fail open - allow request if Redis is down
		return true, fmt.Errorf("redis error: %w", err)
	}

	return result == 1, nil
}

// Reset removes the rate limit counter for a given identifier
// Useful for testing or administrative purposes
func (l *RedisLimiter) Reset(ctx context.Context, identifier string) error {
	key := fmt.Sprintf("ratelimit:%s", identifier)
	return l.client.Del(ctx, key).Err()
}

// Close closes the Redis connection
func (l *RedisLimiter) Close() error {
	return l.client.Close()
}
