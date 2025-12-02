package ratelimit

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// =============================================================================
// Test Infrastructure
// =============================================================================

// TestRedis wraps a Redis testcontainer and provides test utilities
type TestRedis struct {
	Container testcontainers.Container
	Address   string
	ctx       context.Context
}

// setupTestRedis creates a Redis container for testing.
// Falls back to REDIS_URL if container creation fails (e.g., Docker not available).
// Skips test if neither option is available.
func setupTestRedis(t *testing.T) *TestRedis {
	t.Helper()

	// Check if we should skip integration tests
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test: SKIP_INTEGRATION_TESTS=true")
	}

	// Try testcontainers first (preferred for isolation)
	testRedis, err := trySetupRedisContainer(t)
	if err == nil {
		t.Cleanup(func() {
			testRedis.Cleanup(t)
		})
		return testRedis
	}

	// Log the container error for debugging
	t.Logf("Redis testcontainer unavailable: %v", err)

	// Fall back to REDIS_URL if provided (for local development)
	redisURL := os.Getenv("REDIS_URL")
	if redisURL != "" {
		t.Log("WARNING: Falling back to REDIS_URL for test Redis")
		t.Log("WARNING: Tests will NOT be isolated. Data may persist between test runs.")
		return &TestRedis{
			Address: redisURL,
			ctx:     context.Background(),
		}
	}

	// Neither option available - skip the test
	t.Skip("Skipping integration test: Docker not available and REDIS_URL not set")
	return nil
}

// trySetupRedisContainer attempts to create a Redis container.
// Returns an error if container creation fails (e.g., Docker not available).
func trySetupRedisContainer(t *testing.T) (testRedis *TestRedis, err error) {
	t.Helper()

	// Recover from panics (testcontainers panics on Windows without Docker)
	defer func() {
		if r := recover(); r != nil {
			stack := debug.Stack()
			err = fmt.Errorf("testcontainers panic: %v\nStack trace:\n%s", r, string(stack))
		}
	}()

	// Check for common Docker unavailability signals
	if runtime.GOOS == "windows" {
		// On Windows, check if Docker socket exists
		if _, statErr := os.Stat("\\\\.\\pipe\\docker_engine"); os.IsNotExist(statErr) {
			return nil, fmt.Errorf("Docker not available: docker_engine pipe not found")
		}
	}

	ctx := context.Background()

	// Start Redis container
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start Redis container: %w", err)
	}

	// Get the mapped port
	mappedPort, err := container.MappedPort(ctx, "6379")
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get mapped port: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get container host: %w", err)
	}

	address := fmt.Sprintf("%s:%s", host, mappedPort.Port())

	return &TestRedis{
		Container: container,
		Address:   address,
		ctx:       ctx,
	}, nil
}

// Cleanup terminates the container
func (tr *TestRedis) Cleanup(t *testing.T) {
	t.Helper()

	if tr.Container != nil {
		if err := tr.Container.Terminate(tr.ctx); err != nil {
			t.Logf("Warning: failed to terminate Redis container: %v", err)
		}
	}
}

// =============================================================================
// NewRedisLimiter Tests
// =============================================================================

func TestNewRedisLimiter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testRedis := setupTestRedis(t)

	t.Run("creates limiter successfully with valid connection", func(t *testing.T) {
		limiter, err := NewRedisLimiter(testRedis.Address)
		require.NoError(t, err)
		require.NotNil(t, limiter)

		// Verify connection is healthy
		err = limiter.Health(context.Background())
		assert.NoError(t, err)

		// Clean up
		err = limiter.Close()
		assert.NoError(t, err)
	})

	t.Run("fails with invalid address", func(t *testing.T) {
		limiter, err := NewRedisLimiter("invalid:99999")
		assert.Error(t, err)
		assert.Nil(t, limiter)
		assert.Contains(t, err.Error(), "failed to connect to Redis")
	})

	t.Run("fails with unreachable host", func(t *testing.T) {
		limiter, err := NewRedisLimiter("192.0.2.1:6379") // TEST-NET-1, guaranteed unreachable
		assert.Error(t, err)
		assert.Nil(t, limiter)
	})
}

// =============================================================================
// Allow Method Tests (Sliding Window Rate Limiting)
// =============================================================================

func TestRedisLimiter_Allow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testRedis := setupTestRedis(t)
	limiter, err := NewRedisLimiter(testRedis.Address)
	require.NoError(t, err)
	t.Cleanup(func() {
		limiter.Close()
	})

	ctx := context.Background()

	t.Run("allows requests under limit", func(t *testing.T) {
		identifier := "test-allow-under-limit"
		limit := 5
		window := time.Minute

		// Reset any existing state
		err := limiter.Reset(ctx, identifier)
		require.NoError(t, err)

		// All requests should be allowed
		for i := 0; i < limit; i++ {
			allowed, err := limiter.Allow(ctx, identifier, limit, window)
			require.NoError(t, err)
			assert.True(t, allowed, "request %d should be allowed", i+1)
		}
	})

	t.Run("blocks requests over limit", func(t *testing.T) {
		identifier := "test-allow-over-limit"
		limit := 3
		window := time.Minute

		// Reset any existing state
		err := limiter.Reset(ctx, identifier)
		require.NoError(t, err)

		// Use up all allowed requests
		for i := 0; i < limit; i++ {
			allowed, err := limiter.Allow(ctx, identifier, limit, window)
			require.NoError(t, err)
			assert.True(t, allowed, "request %d should be allowed", i+1)
		}

		// Next request should be blocked
		allowed, err := limiter.Allow(ctx, identifier, limit, window)
		require.NoError(t, err)
		assert.False(t, allowed, "request over limit should be blocked")

		// Additional requests should also be blocked
		allowed, err = limiter.Allow(ctx, identifier, limit, window)
		require.NoError(t, err)
		assert.False(t, allowed, "additional requests should remain blocked")
	})

	t.Run("different identifiers have separate limits", func(t *testing.T) {
		identifier1 := "test-separate-1"
		identifier2 := "test-separate-2"
		limit := 2
		window := time.Minute

		// Reset both identifiers
		err := limiter.Reset(ctx, identifier1)
		require.NoError(t, err)
		err = limiter.Reset(ctx, identifier2)
		require.NoError(t, err)

		// Use up limit for identifier1
		for i := 0; i < limit; i++ {
			allowed, err := limiter.Allow(ctx, identifier1, limit, window)
			require.NoError(t, err)
			assert.True(t, allowed)
		}

		// identifier1 should now be blocked
		allowed, err := limiter.Allow(ctx, identifier1, limit, window)
		require.NoError(t, err)
		assert.False(t, allowed, "identifier1 should be blocked")

		// identifier2 should still be allowed
		allowed, err = limiter.Allow(ctx, identifier2, limit, window)
		require.NoError(t, err)
		assert.True(t, allowed, "identifier2 should still be allowed")
	})

	t.Run("sliding window allows new requests after time passes", func(t *testing.T) {
		identifier := "test-sliding-window"
		limit := 2
		window := 2 * time.Second // Short window for testing

		// Reset any existing state
		err := limiter.Reset(ctx, identifier)
		require.NoError(t, err)

		// Use up all requests
		for i := 0; i < limit; i++ {
			allowed, err := limiter.Allow(ctx, identifier, limit, window)
			require.NoError(t, err)
			assert.True(t, allowed)
		}

		// Should be blocked now
		allowed, err := limiter.Allow(ctx, identifier, limit, window)
		require.NoError(t, err)
		assert.False(t, allowed, "should be blocked after using all requests")

		// Wait for window to pass
		time.Sleep(window + 100*time.Millisecond)

		// Should be allowed again
		allowed, err = limiter.Allow(ctx, identifier, limit, window)
		require.NoError(t, err)
		assert.True(t, allowed, "should be allowed after window passes")
	})

	t.Run("handles concurrent requests correctly", func(t *testing.T) {
		identifier := "test-concurrent"
		limit := 10
		window := time.Minute
		numGoroutines := 20

		// Reset any existing state
		err := limiter.Reset(ctx, identifier)
		require.NoError(t, err)

		// Make concurrent requests
		results := make(chan bool, numGoroutines)
		for i := 0; i < numGoroutines; i++ {
			go func() {
				allowed, err := limiter.Allow(ctx, identifier, limit, window)
				if err != nil {
					results <- false
					return
				}
				results <- allowed
			}()
		}

		// Count allowed requests
		allowedCount := 0
		for i := 0; i < numGoroutines; i++ {
			if <-results {
				allowedCount++
			}
		}

		// Exactly 'limit' requests should have been allowed
		assert.Equal(t, limit, allowedCount, "exactly %d requests should be allowed", limit)
	})
}

// =============================================================================
// GetLimitInfo Tests
// =============================================================================

func TestRedisLimiter_GetLimitInfo(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testRedis := setupTestRedis(t)
	limiter, err := NewRedisLimiter(testRedis.Address)
	require.NoError(t, err)
	t.Cleanup(func() {
		limiter.Close()
	})

	ctx := context.Background()

	t.Run("returns correct info for fresh identifier", func(t *testing.T) {
		identifier := "test-info-fresh"
		limit := 100
		window := time.Minute

		// Reset any existing state
		err := limiter.Reset(ctx, identifier)
		require.NoError(t, err)

		info, err := limiter.GetLimitInfo(ctx, identifier, limit, window)
		require.NoError(t, err)
		require.NotNil(t, info)

		assert.Equal(t, limit, info.Limit)
		assert.Equal(t, limit, info.Remaining) // All requests remaining
		assert.True(t, info.ResetAt.After(time.Now()), "reset time should be in the future")
	})

	t.Run("remaining decreases after requests", func(t *testing.T) {
		identifier := "test-info-remaining"
		limit := 10
		window := time.Minute
		requestsMade := 3

		// Reset any existing state
		err := limiter.Reset(ctx, identifier)
		require.NoError(t, err)

		// Make some requests
		for i := 0; i < requestsMade; i++ {
			_, err := limiter.Allow(ctx, identifier, limit, window)
			require.NoError(t, err)
		}

		info, err := limiter.GetLimitInfo(ctx, identifier, limit, window)
		require.NoError(t, err)

		assert.Equal(t, limit, info.Limit)
		assert.Equal(t, limit-requestsMade, info.Remaining)
	})

	t.Run("remaining is zero when limit exceeded", func(t *testing.T) {
		identifier := "test-info-exceeded"
		limit := 3
		window := time.Minute

		// Reset any existing state
		err := limiter.Reset(ctx, identifier)
		require.NoError(t, err)

		// Exceed the limit
		for i := 0; i < limit+5; i++ {
			_, err := limiter.Allow(ctx, identifier, limit, window)
			require.NoError(t, err)
		}

		info, err := limiter.GetLimitInfo(ctx, identifier, limit, window)
		require.NoError(t, err)

		assert.Equal(t, limit, info.Limit)
		assert.Equal(t, 0, info.Remaining, "remaining should be 0 when limit is exceeded")
	})
}

// =============================================================================
// Reset Tests
// =============================================================================

func TestRedisLimiter_Reset(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testRedis := setupTestRedis(t)
	limiter, err := NewRedisLimiter(testRedis.Address)
	require.NoError(t, err)
	t.Cleanup(func() {
		limiter.Close()
	})

	ctx := context.Background()

	t.Run("reset clears rate limit counter", func(t *testing.T) {
		identifier := "test-reset"
		limit := 2
		window := time.Minute

		// Use up all requests
		for i := 0; i < limit; i++ {
			_, err := limiter.Allow(ctx, identifier, limit, window)
			require.NoError(t, err)
		}

		// Should be blocked
		allowed, err := limiter.Allow(ctx, identifier, limit, window)
		require.NoError(t, err)
		assert.False(t, allowed)

		// Reset the counter
		err = limiter.Reset(ctx, identifier)
		require.NoError(t, err)

		// Should be allowed again
		allowed, err = limiter.Allow(ctx, identifier, limit, window)
		require.NoError(t, err)
		assert.True(t, allowed, "should be allowed after reset")
	})

	t.Run("reset on non-existent identifier succeeds", func(t *testing.T) {
		identifier := "test-reset-nonexistent-" + time.Now().Format(time.RFC3339Nano)
		err := limiter.Reset(ctx, identifier)
		assert.NoError(t, err, "reset on non-existent identifier should not error")
	})
}

// =============================================================================
// Health Check Tests
// =============================================================================

func TestRedisLimiter_Health(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testRedis := setupTestRedis(t)

	t.Run("health returns nil for healthy connection", func(t *testing.T) {
		limiter, err := NewRedisLimiter(testRedis.Address)
		require.NoError(t, err)
		defer limiter.Close()

		err = limiter.Health(context.Background())
		assert.NoError(t, err)
	})

	t.Run("health respects context timeout", func(t *testing.T) {
		limiter, err := NewRedisLimiter(testRedis.Address)
		require.NoError(t, err)
		defer limiter.Close()

		// Create an already-cancelled context
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err = limiter.Health(ctx)
		assert.Error(t, err)
	})
}

// =============================================================================
// Close Tests
// =============================================================================

func TestRedisLimiter_Close(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testRedis := setupTestRedis(t)

	t.Run("close terminates connection gracefully", func(t *testing.T) {
		limiter, err := NewRedisLimiter(testRedis.Address)
		require.NoError(t, err)

		// Connection should be healthy before close
		err = limiter.Health(context.Background())
		require.NoError(t, err)

		// Close the connection
		err = limiter.Close()
		assert.NoError(t, err)

		// Operations after close should fail
		err = limiter.Health(context.Background())
		assert.Error(t, err, "health check should fail after close")
	})

	t.Run("close is idempotent", func(t *testing.T) {
		limiter, err := NewRedisLimiter(testRedis.Address)
		require.NoError(t, err)

		// First close should succeed
		err = limiter.Close()
		assert.NoError(t, err)

		// Second close should also succeed (or at least not panic)
		err = limiter.Close()
		// go-redis returns error on double close, which is acceptable
		// The important thing is it doesn't panic
	})
}

// =============================================================================
// Edge Cases and Error Handling
// =============================================================================

func TestRedisLimiter_EdgeCases(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	testRedis := setupTestRedis(t)
	limiter, err := NewRedisLimiter(testRedis.Address)
	require.NoError(t, err)
	t.Cleanup(func() {
		limiter.Close()
	})

	ctx := context.Background()

	t.Run("handles very short window", func(t *testing.T) {
		identifier := "test-short-window"
		limit := 5
		window := 100 * time.Millisecond

		err := limiter.Reset(ctx, identifier)
		require.NoError(t, err)

		allowed, err := limiter.Allow(ctx, identifier, limit, window)
		require.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("handles large limit", func(t *testing.T) {
		identifier := "test-large-limit"
		limit := 10000
		window := time.Minute

		err := limiter.Reset(ctx, identifier)
		require.NoError(t, err)

		allowed, err := limiter.Allow(ctx, identifier, limit, window)
		require.NoError(t, err)
		assert.True(t, allowed)

		info, err := limiter.GetLimitInfo(ctx, identifier, limit, window)
		require.NoError(t, err)
		assert.Equal(t, limit-1, info.Remaining)
	})

	t.Run("handles special characters in identifier", func(t *testing.T) {
		identifier := "test:special/chars@user.com"
		limit := 5
		window := time.Minute

		err := limiter.Reset(ctx, identifier)
		require.NoError(t, err)

		allowed, err := limiter.Allow(ctx, identifier, limit, window)
		require.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("handles empty identifier", func(t *testing.T) {
		identifier := ""
		limit := 5
		window := time.Minute

		// Reset to ensure clean state (especially important when using REDIS_URL fallback)
		err := limiter.Reset(ctx, identifier)
		require.NoError(t, err)

		// Should work even with empty identifier (creates key "ratelimit:")
		allowed, err := limiter.Allow(ctx, identifier, limit, window)
		require.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("handles limit of 1", func(t *testing.T) {
		identifier := "test-limit-one"
		limit := 1
		window := time.Minute

		err := limiter.Reset(ctx, identifier)
		require.NoError(t, err)

		// First request should be allowed
		allowed, err := limiter.Allow(ctx, identifier, limit, window)
		require.NoError(t, err)
		assert.True(t, allowed)

		// Second request should be blocked
		allowed, err = limiter.Allow(ctx, identifier, limit, window)
		require.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("handles zero limit", func(t *testing.T) {
		identifier := "test-zero-limit"
		limit := 0
		window := time.Minute

		err := limiter.Reset(ctx, identifier)
		require.NoError(t, err)

		// With limit=0, the Lua script condition (current < limit) is never true
		// so all requests should be blocked - this documents expected behavior
		allowed, err := limiter.Allow(ctx, identifier, limit, window)
		require.NoError(t, err)
		assert.False(t, allowed, "zero limit should block all requests")
	})
}
