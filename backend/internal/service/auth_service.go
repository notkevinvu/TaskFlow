package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/notkevinvu/taskflow/backend/internal/domain"
	"github.com/notkevinvu/taskflow/backend/internal/middleware"
	"github.com/notkevinvu/taskflow/backend/internal/repository"
)

var (
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo       *repository.UserRepository
	jwtSecret      string
	jwtExpiryHours int
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo *repository.UserRepository, jwtSecret string, jwtExpiryHours int) *AuthService {
	return &AuthService{
		userRepo:       userRepo,
		jwtSecret:      jwtSecret,
		jwtExpiryHours: jwtExpiryHours,
	}
}

// Register creates a new user account
func (s *AuthService) Register(ctx context.Context, dto *domain.CreateUserDTO) (*domain.AuthResponse, error) {
	// Validate input
	if err := dto.Validate(); err != nil {
		return nil, err
	}

	// Check if email already exists
	exists, err := s.userRepo.EmailExists(ctx, dto.Email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrEmailAlreadyExists
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
	// Find user by email
	user, err := s.userRepo.FindByEmail(ctx, dto.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Check password
	if !domain.CheckPasswordHash(dto.Password, user.PasswordHash) {
		return nil, ErrInvalidCredentials
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
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
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
