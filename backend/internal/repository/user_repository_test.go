package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// UserRepository Integration Tests
// =============================================================================

func TestUserRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	repo := NewUserRepository(pool)
	ctx := context.Background()

	t.Run("creates user successfully", func(t *testing.T) {
		user := &domain.User{
			ID:           uuid.New().String(),
			Email:        "create-test-" + uuid.New().String() + "@example.com",
			Name:         "Test User",
			PasswordHash: "$2a$10$examplehash",
			CreatedAt:    time.Now().UTC().Truncate(time.Microsecond),
			UpdatedAt:    time.Now().UTC().Truncate(time.Microsecond),
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		// Verify user was created
		found, err := repo.FindByEmail(ctx, user.Email)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, user.Email, found.Email)
		assert.Equal(t, user.Name, found.Name)
		assert.Equal(t, user.PasswordHash, found.PasswordHash)
	})

	t.Run("fails with duplicate email", func(t *testing.T) {
		email := "duplicate-" + uuid.New().String() + "@example.com"

		user1 := &domain.User{
			ID:           uuid.New().String(),
			Email:        email,
			Name:         "First User",
			PasswordHash: "$2a$10$examplehash1",
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}

		user2 := &domain.User{
			ID:           uuid.New().String(),
			Email:        email, // Same email
			Name:         "Second User",
			PasswordHash: "$2a$10$examplehash2",
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}

		err := repo.Create(ctx, user1)
		require.NoError(t, err)

		err = repo.Create(ctx, user2)
		assert.Error(t, err, "should fail with duplicate email")
	})

	t.Run("fails with invalid UUID", func(t *testing.T) {
		user := &domain.User{
			ID:           "not-a-valid-uuid",
			Email:        "invalid-" + uuid.New().String() + "@example.com",
			Name:         "Invalid User",
			PasswordHash: "$2a$10$examplehash",
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}

		err := repo.Create(ctx, user)
		assert.Error(t, err, "should fail with invalid UUID")
	})

	t.Run("stores timestamps correctly", func(t *testing.T) {
		now := time.Now().UTC().Truncate(time.Microsecond)
		user := &domain.User{
			ID:           uuid.New().String(),
			Email:        "timestamps-" + uuid.New().String() + "@example.com",
			Name:         "Timestamp User",
			PasswordHash: "$2a$10$examplehash",
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, user.ID)
		require.NoError(t, err)
		require.NotNil(t, found)

		// PostgreSQL stores with microsecond precision
		assert.WithinDuration(t, now, found.CreatedAt, time.Second)
		assert.WithinDuration(t, now, found.UpdatedAt, time.Second)
	})
}

func TestUserRepository_FindByEmail(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	repo := NewUserRepository(pool)
	ctx := context.Background()

	t.Run("finds existing user", func(t *testing.T) {
		email := "findbyemail-" + uuid.New().String() + "@example.com"
		user := &domain.User{
			ID:           uuid.New().String(),
			Email:        email,
			Name:         "Find Me",
			PasswordHash: "$2a$10$examplehash",
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		found, err := repo.FindByEmail(ctx, email)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, user.ID, found.ID)
		assert.Equal(t, user.Name, found.Name)
	})

	t.Run("returns nil for non-existent email", func(t *testing.T) {
		found, err := repo.FindByEmail(ctx, "nonexistent-"+uuid.New().String()+"@example.com")
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("email lookup is case-sensitive", func(t *testing.T) {
		// PostgreSQL default is case-sensitive for VARCHAR comparisons
		email := "CaseSensitive-" + uuid.New().String() + "@example.com"
		user := &domain.User{
			ID:           uuid.New().String(),
			Email:        email,
			Name:         "Case Sensitive",
			PasswordHash: "$2a$10$examplehash",
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		// Exact match should work
		found, err := repo.FindByEmail(ctx, email)
		require.NoError(t, err)
		require.NotNil(t, found)

		// Different case should not match
		lowercaseEmail := "casesensitive-" + email[15:] // Lowercase version
		found, err = repo.FindByEmail(ctx, lowercaseEmail)
		require.NoError(t, err)
		assert.Nil(t, found)
	})
}

func TestUserRepository_FindByID(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	repo := NewUserRepository(pool)
	ctx := context.Background()

	t.Run("finds existing user", func(t *testing.T) {
		userID := uuid.New().String()
		user := &domain.User{
			ID:           userID,
			Email:        "findbyid-" + uuid.New().String() + "@example.com",
			Name:         "Find By ID",
			PasswordHash: "$2a$10$examplehash",
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		found, err := repo.FindByID(ctx, userID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, userID, found.ID)
		assert.Equal(t, user.Email, found.Email)
	})

	t.Run("returns nil for non-existent ID", func(t *testing.T) {
		nonExistentID := uuid.New().String()
		found, err := repo.FindByID(ctx, nonExistentID)
		require.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("returns error for invalid UUID", func(t *testing.T) {
		_, err := repo.FindByID(ctx, "not-a-uuid")
		assert.Error(t, err)
	})
}

func TestUserRepository_EmailExists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	pool := setupTestDB(t)
	repo := NewUserRepository(pool)
	ctx := context.Background()

	t.Run("returns true for existing email", func(t *testing.T) {
		email := "exists-" + uuid.New().String() + "@example.com"
		user := &domain.User{
			ID:           uuid.New().String(),
			Email:        email,
			Name:         "Exists",
			PasswordHash: "$2a$10$examplehash",
			CreatedAt:    time.Now().UTC(),
			UpdatedAt:    time.Now().UTC(),
		}

		err := repo.Create(ctx, user)
		require.NoError(t, err)

		exists, err := repo.EmailExists(ctx, email)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("returns false for non-existent email", func(t *testing.T) {
		exists, err := repo.EmailExists(ctx, "notexists-"+uuid.New().String()+"@example.com")
		require.NoError(t, err)
		assert.False(t, exists)
	})
}
