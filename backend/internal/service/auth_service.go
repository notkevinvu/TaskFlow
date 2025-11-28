package service

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/middleware"
	"github.com/notkevinvu/taskflow/backend/internal/ports"
	"github.com/notkevinvu/taskflow/backend/internal/validation"
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo       ports.UserRepository
	jwtSecret      string
	jwtExpiryHours int
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo ports.UserRepository, jwtSecret string, jwtExpiryHours int) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		jwtSecret:      jwtSecret,
		jwtExpiryHours: jwtExpiryHours,
	}
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, dto *domain.CreateUserDTO) (*domain.AuthResponse, error) {
	// Validate email
	if err := validation.ValidateEmail(dto.Email); err != nil {
		return nil, err
	}

	// Validate name
	name, err := validation.ValidateRequiredText(dto.Name, 255, "name")
	if err != nil {
		return nil, err
	}
	dto.Name = name

	// Validate password
	if err := validation.ValidatePassword(dto.Password); err != nil {
		return nil, err
	}

	// Check if email already exists
	exists, err := s.userRepo.EmailExists(ctx, dto.Email)
	if err != nil {
		return nil, domain.NewInternalError("failed to check email existence", err)
	}
	if exists {
		return nil, domain.NewConflictError("user", "email already exists")
	}

	// Hash password
	passwordHash, err := domain.HashPassword(dto.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	now := time.Now()
	user := &domain.User{
		ID:           uuid.New().String(),
		Email:        dto.Email,
		Name:         dto.Name,
		PasswordHash: passwordHash,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Generate JWT token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		User:        *user,
		AccessToken: token,
	}, nil
}

// Login authenticates a user and returns a token
func (s *AuthService) Login(ctx context.Context, dto *domain.LoginDTO) (*domain.AuthResponse, error) {
	// Validate email
	if err := validation.ValidateEmail(dto.Email); err != nil {
		return nil, err
	}

	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, dto.Email)
	if err != nil {
		return nil, domain.NewInternalError("failed to find user", err)
	}
	if user == nil {
		return nil, domain.NewUnauthorizedError("invalid email or password")
	}

	// Check password
	if !domain.CheckPasswordHash(dto.Password, user.PasswordHash) {
		return nil, domain.NewUnauthorizedError("invalid email or password")
	}

	// Generate JWT token
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		User:        *user,
		AccessToken: token,
	}, nil
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(ctx context.Context, id string) (*domain.User, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, domain.NewInternalError("failed to find user", err)
	}
	if user == nil {
		return nil, domain.NewNotFoundError("user", id)
	}
	return user, nil
}

// generateToken creates a JWT token for a user
func (s *AuthService) generateToken(user *domain.User) (string, error) {
	expirationTime := time.Now().Add(time.Duration(s.jwtExpiryHours) * time.Hour)

	claims := &middleware.Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
