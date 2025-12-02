package repository

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestDB wraps a PostgreSQL testcontainer and provides test utilities
type TestDB struct {
	Container *postgres.PostgresContainer
	Pool      *pgxpool.Pool
	ctx       context.Context
}

// setupTestDB creates a PostgreSQL container for testing.
// Falls back to DATABASE_URL if container creation fails (e.g., Docker not available).
// Skips test if neither option is available.
func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	// Check if we should skip integration tests
	if os.Getenv("SKIP_INTEGRATION_TESTS") == "true" {
		t.Skip("Skipping integration test: SKIP_INTEGRATION_TESTS=true")
	}

	// Try testcontainers first (preferred for isolation)
	testDB, err := trySetupTestContainer(t)
	if err == nil {
		t.Cleanup(func() {
			testDB.Cleanup(t)
		})
		return testDB.Pool
	}

	// Log the container error for debugging
	t.Logf("Testcontainer unavailable: %v", err)

	// Fall back to DATABASE_URL if provided (for local development)
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL != "" {
		t.Log("WARNING: Falling back to DATABASE_URL for test database")
		t.Log("WARNING: Tests will NOT be isolated. Data may persist between test runs.")
		pool, err := pgxpool.New(context.Background(), databaseURL)
		if err != nil {
			t.Fatalf("Failed to connect to test database: %v", err)
		}
		t.Cleanup(func() {
			pool.Close()
		})
		return pool
	}

	// Neither option available - skip the test
	t.Skip("Skipping integration test: Docker not available and DATABASE_URL not set")
	return nil
}

// trySetupTestContainer attempts to create a PostgreSQL container.
// Returns an error if container creation fails (e.g., Docker not available).
func trySetupTestContainer(t *testing.T) (testDB *TestDB, err error) {
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

	// Start PostgreSQL container
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("taskflow_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start PostgreSQL container: %w", err)
	}

	// Get connection string
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		pgContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	// Create connection pool
	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		pgContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	testDB = &TestDB{
		Container: pgContainer,
		Pool:      pool,
		ctx:       ctx,
	}

	// Apply migrations
	if err := testDB.applyMigrations(t); err != nil {
		testDB.Cleanup(t)
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	return testDB, nil
}

// SetupTestDBContainer creates a new PostgreSQL container and applies migrations.
// Returns a TestDB instance that should be cleaned up with Cleanup().
// This is the public API for tests that need direct container access.
func SetupTestDBContainer(t *testing.T) *TestDB {
	t.Helper()

	testDB, err := trySetupTestContainer(t)
	if err != nil {
		t.Skipf("Skipping test: %v", err)
	}
	return testDB
}

// applyMigrations reads and executes all .up.sql migration files
func (tdb *TestDB) applyMigrations(t *testing.T) error {
	t.Helper()

	// Find migrations directory - try multiple locations
	migrationsDir := findMigrationsDir()
	if migrationsDir == "" {
		return fmt.Errorf("migrations directory not found")
	}

	// Read migration files
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("failed to read migrations directory: %w", err)
	}

	// Sort migration files by name to ensure correct order
	var upMigrations []string
	for _, entry := range entries {
		if strings.HasSuffix(entry.Name(), ".up.sql") {
			upMigrations = append(upMigrations, entry.Name())
		}
	}
	sort.Strings(upMigrations)

	// Execute each migration
	for _, migration := range upMigrations {
		migrationPath := filepath.Join(migrationsDir, migration)
		content, err := os.ReadFile(migrationPath)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", migration, err)
		}

		_, err = tdb.Pool.Exec(tdb.ctx, string(content))
		if err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", migration, err)
		}

		t.Logf("Applied migration: %s", migration)
	}

	return nil
}

// findMigrationsDir locates the migrations directory from various test locations
func findMigrationsDir() string {
	// Try common relative paths from test execution locations
	candidates := []string{
		"../../migrations",           // From internal/repository
		"../../../migrations",        // From internal/repository subdir
		"migrations",                 // From backend root
		"backend/migrations",         // From project root
		"../backend/migrations",      // From frontend root
	}

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

// Cleanup terminates the container and closes the connection pool
func (tdb *TestDB) Cleanup(t *testing.T) {
	t.Helper()

	if tdb.Pool != nil {
		tdb.Pool.Close()
	}

	if tdb.Container != nil {
		if err := tdb.Container.Terminate(tdb.ctx); err != nil {
			t.Logf("Warning: failed to terminate container: %v", err)
		}
	}
}

// TruncateTables clears all data from tables while preserving schema
func (tdb *TestDB) TruncateTables(t *testing.T) {
	t.Helper()

	// Truncate in order respecting foreign keys
	tables := []string{"task_history", "tasks", "users"}
	for _, table := range tables {
		_, err := tdb.Pool.Exec(tdb.ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Fatalf("Failed to truncate table %s: %v", table, err)
		}
	}
}

// cleanupTestDB closes the database connection (kept for compatibility)
func cleanupTestDB(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	// Pool cleanup is now handled by t.Cleanup() in setupTestDB
	// This function is kept for backward compatibility but does nothing
}

// createTestUser creates a test user in the database and returns the user ID
func createTestUser(t *testing.T, ctx context.Context, pool *pgxpool.Pool) string {
	t.Helper()
	userID := uuid.New().String()
	email := "test-" + uuid.New().String() + "@example.com"

	_, err := pool.Exec(ctx,
		"INSERT INTO users (id, email, name, password_hash, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		userID, email, "Test User", "hash", time.Now(), time.Now())
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return userID
}
