package domain

import (
	"errors"
	"regexp"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrWeakPassword    = errors.New("password must contain uppercase, lowercase, and number")
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// User represents a user in the system
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"` // Never expose password hash in JSON
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CreateUserDTO is used for user registration
type CreateUserDTO struct {
	Email    string `json:"email" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginDTO is used for user login
type LoginDTO struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse is returned after successful login/register
type AuthResponse struct {
	User        User   `json:"user"`
	AccessToken string `json:"access_token"`
}

// Validate validates the CreateUserDTO
func (dto *CreateUserDTO) Validate() error {
	if !emailRegex.MatchString(dto.Email) {
		return ErrInvalidEmail
	}

	if len(dto.Password) < 8 {
		return ErrPasswordTooShort
	}

	// Check for strong password (at least one uppercase, one lowercase, one number)
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(dto.Password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(dto.Password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(dto.Password)

	if !hasUpper || !hasLower || !hasNumber {
		return ErrWeakPassword
	}

	return nil
}

// HashPassword hashes the password using bcrypt
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPasswordHash compares password with hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
