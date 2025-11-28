package repository

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
)

func TestUserRepository_Create(t *testing.T) {
	pool := setupTestDB(t)
	defer cleanupTestDB(t, pool)

	repo := NewUserRepository(pool)
	ctx := context.Background()

	// Create a test user
	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        "test-" + uuid.New().String() + "@example.com",
		Name:         "Test User",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Test Create
	err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Cleanup
	_, err = pool.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	if err != nil {
		t.Logf("Warning: Failed to cleanup test user: %v", err)
	}
}

func TestUserRepository_FindByEmail(t *testing.T) {
	pool := setupTestDB(t)
	defer cleanupTestDB(t, pool)

	repo := NewUserRepository(pool)
	ctx := context.Background()

	// Create a test user
	email := "test-" + uuid.New().String() + "@example.com"
	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        email,
		Name:         "Test User",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	defer pool.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)

	// Test FindByEmail
	found, err := repo.FindByEmail(ctx, email)
	if err != nil {
		t.Fatalf("Failed to find user by email: %v", err)
	}

	if found == nil {
		t.Fatal("Expected to find user, got nil")
	}

	if found.Email != email {
		t.Errorf("Expected email %s, got %s", email, found.Email)
	}

	if found.Name != user.Name {
		t.Errorf("Expected name %s, got %s", user.Name, found.Name)
	}
}

func TestUserRepository_FindByID(t *testing.T) {
	pool := setupTestDB(t)
	defer cleanupTestDB(t, pool)

	repo := NewUserRepository(pool)
	ctx := context.Background()

	// Create a test user
	userID := uuid.New().String()
	user := &domain.User{
		ID:           userID,
		Email:        "test-" + uuid.New().String() + "@example.com",
		Name:         "Test User",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	defer pool.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)

	// Test FindByID
	found, err := repo.FindByID(ctx, userID)
	if err != nil {
		t.Fatalf("Failed to find user by ID: %v", err)
	}

	if found == nil {
		t.Fatal("Expected to find user, got nil")
	}

	if found.ID != userID {
		t.Errorf("Expected ID %s, got %s", userID, found.ID)
	}
}

func TestUserRepository_EmailExists(t *testing.T) {
	pool := setupTestDB(t)
	defer cleanupTestDB(t, pool)

	repo := NewUserRepository(pool)
	ctx := context.Background()

	// Test non-existent email
	exists, err := repo.EmailExists(ctx, "nonexistent-"+uuid.New().String()+"@example.com")
	if err != nil {
		t.Fatalf("Failed to check email existence: %v", err)
	}

	if exists {
		t.Error("Expected email to not exist, but it does")
	}

	// Create a test user
	email := "test-" + uuid.New().String() + "@example.com"
	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        email,
		Name:         "Test User",
		PasswordHash: "hashed_password",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	err = repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	defer pool.Exec(ctx, "DELETE FROM users WHERE id = $1", user.ID)

	// Test existing email
	exists, err = repo.EmailExists(ctx, email)
	if err != nil {
		t.Fatalf("Failed to check email existence: %v", err)
	}

	if !exists {
		t.Error("Expected email to exist, but it doesn't")
	}
}
