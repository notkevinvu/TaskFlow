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
		UserType:     domain.UserTypeRegistered,
		Email:        &dto.Email,
		Name:         &dto.Name,
		PasswordHash: &passwordHash,
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

	// Anonymous users cannot login with password (they have no password)
	if user.IsAnonymous() || user.PasswordHash == nil {
		return nil, domain.NewUnauthorizedError("invalid email or password")
	}

	// Check password
	if !domain.CheckPasswordHash(dto.Password, *user.PasswordHash) {
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

// CreateAnonymousUser creates a new anonymous guest user
func (s *AuthService) CreateAnonymousUser(ctx context.Context) (*domain.AuthResponse, error) {
	// Create anonymous user with expiry
	now := time.Now()
	expiresAt := domain.CalculateAnonymousExpiry()

	user := &domain.User{
		ID:        uuid.New().String(),
		UserType:  domain.UserTypeAnonymous,
		ExpiresAt: &expiresAt,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.userRepo.CreateAnonymous(ctx, user); err != nil {
		return nil, domain.NewInternalError("failed to create anonymous user", err)
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

// ConvertGuestToRegistered converts an anonymous user to a registered user
func (s *AuthService) ConvertGuestToRegistered(ctx context.Context, userID string, dto *domain.ConvertGuestDTO) (*domain.AuthResponse, error) {
	// First verify the user exists and is anonymous
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, domain.NewInternalError("failed to find user", err)
	}
	if user == nil {
		return nil, domain.NewNotFoundError("user", userID)
	}
	if !user.IsAnonymous() {
		return nil, domain.NewValidationError("user", "user is already registered")
	}

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
		return nil, domain.NewInternalError("failed to hash password", err)
	}

	// Convert the user
	if err := s.userRepo.ConvertToRegistered(ctx, userID, dto.Email, dto.Name, passwordHash); err != nil {
		return nil, domain.NewInternalError("failed to convert user", err)
	}

	// Fetch the updated user
	updatedUser, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, domain.NewInternalError("failed to fetch updated user", err)
	}

	// Generate new JWT token with registered status
	token, err := s.generateToken(updatedUser)
	if err != nil {
		return nil, err
	}

	return &domain.AuthResponse{
		User:        *updatedUser,
		AccessToken: token,
	}, nil
}

// generateToken creates a JWT token for a user
func (s *AuthService) generateToken(user *domain.User) (string, error) {
	expirationTime := time.Now().Add(time.Duration(s.jwtExpiryHours) * time.Hour)

	// Get email as string (empty for anonymous users)
	email := user.GetEmail()

	claims := &middleware.Claims{
		UserID:      user.ID,
		Email:       email,
		IsAnonymous: user.IsAnonymous(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
