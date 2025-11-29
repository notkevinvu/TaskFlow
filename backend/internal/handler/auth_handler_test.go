package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService is a mock implementation of ports.AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(ctx context.Context, dto *domain.CreateUserDTO) (*domain.AuthResponse, error) {
	args := m.Called(ctx, dto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuthResponse), args.Error(1)
}

func (m *MockAuthService) Login(ctx context.Context, dto *domain.LoginDTO) (*domain.AuthResponse, error) {
	args := m.Called(ctx, dto)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.AuthResponse), args.Error(1)
}

func (m *MockAuthService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

// setupAuthTest creates a test Gin context and response recorder
func setupAuthTest() (*gin.Engine, *httptest.ResponseRecorder, *MockAuthService) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	router := gin.New()
	// Add error handler middleware to properly map errors to HTTP status codes
	router.Use(middleware.ErrorHandler())
	mockService := new(MockAuthService)
	return router, w, mockService
}

func TestAuthHandler_Register_Success(t *testing.T) {
	router, w, mockService := setupAuthTest()
	handler := NewAuthHandler(mockService)

	// Setup route
	router.POST("/register", handler.Register)

	// Mock expectations
	expectedResponse := &domain.AuthResponse{
		User: domain.User{
			ID:    "user-123",
			Email: "test@example.com",
			Name:  "Test User",
		},
		AccessToken: "test-token-123",
	}

	mockService.On("Register", mock.Anything, mock.MatchedBy(func(dto *domain.CreateUserDTO) bool {
		return dto.Email == "test@example.com" && dto.Name == "Test User" && dto.Password == "password123"
	})).Return(expectedResponse, nil)

	// Create request
	reqBody := map[string]string{
		"email":    "test@example.com",
		"name":     "Test User",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)

	// Parse response
	var response domain.AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "user-123", response.User.ID)
	assert.Equal(t, "test@example.com", response.User.Email)
	assert.Equal(t, "test-token-123", response.AccessToken)
}

func TestAuthHandler_Register_InvalidJSON(t *testing.T) {
	router, w, mockService := setupAuthTest()
	handler := NewAuthHandler(mockService)

	router.POST("/register", handler.Register)

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/register", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "Register")
}

func TestAuthHandler_Register_EmailAlreadyExists(t *testing.T) {
	router, w, mockService := setupAuthTest()
	handler := NewAuthHandler(mockService)

	router.POST("/register", handler.Register)

	// Mock expectations - email already exists
	mockService.On("Register", mock.Anything, mock.Anything).
		Return(nil, domain.NewConflictError("user", "email already exists"))

	// Create request
	reqBody := map[string]string{
		"email":    "existing@example.com",
		"name":     "Test User",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusConflict, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_Register_ValidationError(t *testing.T) {
	router, w, mockService := setupAuthTest()
	handler := NewAuthHandler(mockService)

	router.POST("/register", handler.Register)

	// Mock expectations - validation error
	mockService.On("Register", mock.Anything, mock.Anything).
		Return(nil, domain.NewValidationError("email", "invalid email format"))

	// Create request
	reqBody := map[string]string{
		"email":    "invalid-email",
		"name":     "Test User",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/register", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_Success(t *testing.T) {
	router, w, mockService := setupAuthTest()
	handler := NewAuthHandler(mockService)

	router.POST("/login", handler.Login)

	// Mock expectations
	expectedResponse := &domain.AuthResponse{
		User: domain.User{
			ID:    "user-123",
			Email: "test@example.com",
			Name:  "Test User",
		},
		AccessToken: "test-token-123",
	}

	mockService.On("Login", mock.Anything, mock.MatchedBy(func(dto *domain.LoginDTO) bool {
		return dto.Email == "test@example.com" && dto.Password == "password123"
	})).Return(expectedResponse, nil)

	// Create request
	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	// Parse response
	var response domain.AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "user-123", response.User.ID)
	assert.Equal(t, "test-token-123", response.AccessToken)
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	router, w, mockService := setupAuthTest()
	handler := NewAuthHandler(mockService)

	router.POST("/login", handler.Login)

	// Mock expectations - unauthorized error
	mockService.On("Login", mock.Anything, mock.Anything).
		Return(nil, domain.NewUnauthorizedError("invalid email or password"))

	// Create request
	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "wrongpassword",
	}
	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertExpectations(t)
}

func TestAuthHandler_Login_InvalidJSON(t *testing.T) {
	router, w, mockService := setupAuthTest()
	handler := NewAuthHandler(mockService)

	router.POST("/login", handler.Login)

	// Create request with invalid JSON
	req := httptest.NewRequest("POST", "/login", bytes.NewBufferString("{invalid}"))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	mockService.AssertNotCalled(t, "Login")
}

func TestAuthHandler_Me_Success(t *testing.T) {
	router, w, mockService := setupAuthTest()
	handler := NewAuthHandler(mockService)

	// Setup authenticated route (with user ID in context)
	router.GET("/me", func(c *gin.Context) {
		// Simulate auth middleware setting user ID
		c.Set("user_id", "user-123")
		handler.Me(c)
	})

	// Mock expectations
	expectedUser := &domain.User{
		ID:    "user-123",
		Email: "test@example.com",
		Name:  "Test User",
	}

	mockService.On("GetUserByID", mock.Anything, "user-123").
		Return(expectedUser, nil)

	// Create request
	req := httptest.NewRequest("GET", "/me", nil)

	// Execute request
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)

	// Parse response
	var user domain.User
	err := json.Unmarshal(w.Body.Bytes(), &user)
	assert.NoError(t, err)
	assert.Equal(t, "user-123", user.ID)
	assert.Equal(t, "test@example.com", user.Email)
}

func TestAuthHandler_Me_Unauthenticated(t *testing.T) {
	router, w, mockService := setupAuthTest()
	handler := NewAuthHandler(mockService)

	// Setup route WITHOUT user ID in context (unauthenticated)
	router.GET("/me", handler.Me)

	// Create request
	req := httptest.NewRequest("GET", "/me", nil)

	// Execute request
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertNotCalled(t, "GetUserByID")
}

func TestAuthHandler_Me_UserNotFound(t *testing.T) {
	router, w, mockService := setupAuthTest()
	handler := NewAuthHandler(mockService)

	// Setup authenticated route
	router.GET("/me", func(c *gin.Context) {
		c.Set("user_id", "nonexistent-user")
		handler.Me(c)
	})

	// Mock expectations - user not found
	mockService.On("GetUserByID", mock.Anything, "nonexistent-user").
		Return(nil, domain.NewNotFoundError("user", "nonexistent-user"))

	// Create request
	req := httptest.NewRequest("GET", "/me", nil)

	// Execute request
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}
