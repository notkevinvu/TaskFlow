package config

import (
	"os"
	"testing"
)

func TestLoad_RequiredEnvVars(t *testing.T) {
	// Save original env vars
	origDB := os.Getenv("DATABASE_URL")
	origJWT := os.Getenv("JWT_SECRET")

	// Cleanup function to restore env vars
	defer func() {
		if origDB != "" {
			os.Setenv("DATABASE_URL", origDB)
		}
		if origJWT != "" {
			os.Setenv("JWT_SECRET", origJWT)
		}
	}()

	t.Run("panics when DATABASE_URL is missing", func(t *testing.T) {
		os.Unsetenv("DATABASE_URL")
		os.Setenv("JWT_SECRET", "test-secret")

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when DATABASE_URL is missing, but did not panic")
			}
		}()

		Load()
	})

	t.Run("panics when JWT_SECRET is missing", func(t *testing.T) {
		os.Setenv("DATABASE_URL", "postgres://test")
		os.Unsetenv("JWT_SECRET")

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when JWT_SECRET is missing, but did not panic")
			}
		}()

		Load()
	})

	t.Run("loads successfully when all required vars present", func(t *testing.T) {
		os.Setenv("DATABASE_URL", "postgres://test:5432/db")
		os.Setenv("JWT_SECRET", "test-secret-key-minimum-32-chars")

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Unexpected panic: %v", r)
			}
		}()

		cfg := Load()

		if cfg.DatabaseURL != "postgres://test:5432/db" {
			t.Errorf("Expected DATABASE_URL to be set, got: %s", cfg.DatabaseURL)
		}

		if cfg.JWTSecret != "test-secret-key-minimum-32-chars" {
			t.Errorf("Expected JWT_SECRET to be set, got: %s", cfg.JWTSecret)
		}
	})
}

func TestGetEnvRequired(t *testing.T) {
	t.Run("returns value when env var is set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "test-value")
		defer os.Unsetenv("TEST_VAR")

		value := getEnvRequired("TEST_VAR")
		if value != "test-value" {
			t.Errorf("Expected 'test-value', got '%s'", value)
		}
	})

	t.Run("panics when env var is not set", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when env var is missing, but did not panic")
			}
		}()

		getEnvRequired("NONEXISTENT_VAR")
	})
}
