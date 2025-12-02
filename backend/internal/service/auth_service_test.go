package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const testJWTSecret = "test-secret-key-for-jwt-testing-minimum-32-chars"
const testJWTExpiryHours = 24

// Helper to create a valid user for testing
func createTestUser(id, email string) *domain.User {
	now := time.Now()
	passwordHash, err := domain.HashPassword("ValidPass123!")
	if err != nil {
		panic(fmt.Sprintf("Failed to hash password in test helper: %v", err))
	}
	return &domain.User{
		ID:           id,
		Email:        email,
		Name:         "Test User",
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
}

// =============================================================================
// AuthService.Register Tests
// =============================================================================

func TestAuthService_Register_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewAuthService(mockUserRepo, testJWTSecret, testJWTExpiryHours)

	dto := &domain.CreateUserDTO{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "ValidPass123!",
	}

	mockUserRepo.On("EmailExists", mock.Anything, "test@example.com").Return(false, nil)
	mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	response, err := service.Register(context.Background(), dto)

	require.NoError(t, err)
	require.NotNil(t, response)
	assert.Equal(t, "test@example.com", response.User.Email)
	assert.Equal(t, "Test User", response.User.Name)
	assert.NotEmpty(t, response.AccessToken)
	assert.NotEmpty(t, response.User.ID)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Register_InvalidEmail(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewAuthService(mockUserRepo, testJWTSecret, testJWTExpiryHours)

	dto := &domain.CreateUserDTO{
		Email:    "invalid-email",
		Name:     "Test User",
		Password: "ValidPass123!",
	}

	response, err := service.Register(context.Background(), dto)

	assert.Error(t, err)
	assert.Nil(t, response)
	var validationErr *domain.ValidationError
	assert.True(t, errors.As(err, &validationErr))
}

func TestAuthService_Register_EmptyName(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewAuthService(mockUserRepo, testJWTSecret, testJWTExpiryHours)

	dto := &domain.CreateUserDTO{
		Email:    "test@example.com",
		Name:     "",
		Password: "ValidPass123!",
	}

	response, err := service.Register(context.Background(), dto)

	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestAuthService_Register_WeakPassword(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewAuthService(mockUserRepo, testJWTSecret, testJWTExpiryHours)

	dto := &domain.CreateUserDTO{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "weak", // Too short
	}

	response, err := service.Register(context.Background(), dto)

	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestAuthService_Register_EmailAlreadyExists(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewAuthService(mockUserRepo, testJWTSecret, testJWTExpiryHours)

	dto := &domain.CreateUserDTO{
		Email:    "existing@example.com",
		Name:     "Test User",
		Password: "ValidPass123!",
	}

	mockUserRepo.On("EmailExists", mock.Anything, "existing@example.com").Return(true, nil)

	response, err := service.Register(context.Background(), dto)

	assert.Error(t, err)
	assert.Nil(t, response)
	var conflictErr *domain.ConflictError
	assert.True(t, errors.As(err, &conflictErr))
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Register_EmailCheckError(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewAuthService(mockUserRepo, testJWTSecret, testJWTExpiryHours)

	dto := &domain.CreateUserDTO{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "ValidPass123!",
	}

	mockUserRepo.On("EmailExists", mock.Anything, "test@example.com").
		Return(false, errors.New("database error"))

	response, err := service.Register(context.Background(), dto)

	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestAuthService_Register_CreateUserError(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewAuthService(mockUserRepo, testJWTSecret, testJWTExpiryHours)

	dto := &domain.CreateUserDTO{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "ValidPass123!",
	}

	mockUserRepo.On("EmailExists", mock.Anything, "test@example.com").Return(false, nil)
	mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).
		Return(errors.New("database error"))

	response, err := service.Register(context.Background(), dto)

	assert.Error(t, err)
	assert.Nil(t, response)
}

// =============================================================================
// AuthService.Login Tests
// =============================================================================

func TestAuthService_Login_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewAuthService(mockUserRepo, testJWTSecret, testJWTExpiryHours)

	existingUser := createTestUser("user-123", "test@example.com")

	dto := &domain.LoginDTO{
		Email:    "test@example.com",
		Password: "ValidPass123!",
	}

	mockUserRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(existingUser, nil)

	response, err := service.Login(context.Background(), dto)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "test@example.com", response.User.Email)
	assert.NotEmpty(t, response.AccessToken)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_Login_InvalidEmail(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewAuthService(mockUserRepo, testJWTSecret, testJWTExpiryHours)

	dto := &domain.LoginDTO{
		Email:    "invalid-email",
		Password: "ValidPass123!",
	}

	response, err := service.Login(context.Background(), dto)

	assert.Error(t, err)
	assert.Nil(t, response)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewAuthService(mockUserRepo, testJWTSecret, testJWTExpiryHours)

	dto := &domain.LoginDTO{
		Email:    "nonexistent@example.com",
		Password: "ValidPass123!",
	}

	mockUserRepo.On("FindByEmail", mock.Anything, "nonexistent@example.com").Return(nil, nil)

	response, err := service.Login(context.Background(), dto)

	assert.Error(t, err)
	assert.Nil(t, response)
	var unauthorizedErr *domain.UnauthorizedError
	assert.True(t, errors.As(err, &unauthorizedErr))
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewAuthService(mockUserRepo, testJWTSecret, testJWTExpiryHours)

	existingUser := createTestUser("user-123", "test@example.com")

	dto := &domain.LoginDTO{
		Email:    "test@example.com",
		Password: "WrongPassword123!",
	}

	mockUserRepo.On("FindByEmail", mock.Anything, "test@example.com").Return(existingUser, nil)

	response, err := service.Login(context.Background(), dto)

	assert.Error(t, err)
	assert.Nil(t, response)
	var unauthorizedErr *domain.UnauthorizedError
	assert.True(t, errors.As(err, &unauthorizedErr))
}

func TestAuthService_Login_FindUserError(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewAuthService(mockUserRepo, testJWTSecret, testJWTExpiryHours)

	dto := &domain.LoginDTO{
		Email:    "test@example.com",
		Password: "ValidPass123!",
	}

	mockUserRepo.On("FindByEmail", mock.Anything, "test@example.com").
		Return(nil, errors.New("database error"))

	response, err := service.Login(context.Background(), dto)

	assert.Error(t, err)
	assert.Nil(t, response)
}

// =============================================================================
// AuthService.GetUserByID Tests
// =============================================================================

func TestAuthService_GetUserByID_Success(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewAuthService(mockUserRepo, testJWTSecret, testJWTExpiryHours)

	existingUser := createTestUser("user-123", "test@example.com")

	mockUserRepo.On("FindByID", mock.Anything, "user-123").Return(existingUser, nil)

	user, err := service.GetUserByID(context.Background(), "user-123")

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "user-123", user.ID)
	assert.Equal(t, "test@example.com", user.Email)
	mockUserRepo.AssertExpectations(t)
}

func TestAuthService_GetUserByID_NotFound(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewAuthService(mockUserRepo, testJWTSecret, testJWTExpiryHours)

	mockUserRepo.On("FindByID", mock.Anything, "nonexistent").Return(nil, nil)

	user, err := service.GetUserByID(context.Background(), "nonexistent")

	assert.Error(t, err)
	assert.Nil(t, user)
	var notFoundErr *domain.NotFoundError
	assert.True(t, errors.As(err, &notFoundErr))
}

func TestAuthService_GetUserByID_RepoError(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewAuthService(mockUserRepo, testJWTSecret, testJWTExpiryHours)

	mockUserRepo.On("FindByID", mock.Anything, "user-123").
		Return(nil, errors.New("database error"))

	user, err := service.GetUserByID(context.Background(), "user-123")

	assert.Error(t, err)
	assert.Nil(t, user)
}

// =============================================================================
// AuthService.generateToken Tests (via Register/Login)
// =============================================================================

func TestAuthService_GenerateToken_ValidJWT(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	service := NewAuthService(mockUserRepo, testJWTSecret, testJWTExpiryHours)

	dto := &domain.CreateUserDTO{
		Email:    "test@example.com",
		Name:     "Test User",
		Password: "ValidPass123!",
	}

	mockUserRepo.On("EmailExists", mock.Anything, "test@example.com").Return(false, nil)
	mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	response, err := service.Register(context.Background(), dto)

	assert.NoError(t, err)
	assert.NotNil(t, response)

	// Token should be a valid JWT (3 parts separated by dots)
	token := response.AccessToken
	assert.NotEmpty(t, token)
	parts := strings.Split(token, ".")
	assert.Len(t, parts, 3, "JWT token should have 3 parts (header.payload.signature)")
	assert.Greater(t, len(token), 50, "JWT tokens are typically longer than 50 characters")
}
