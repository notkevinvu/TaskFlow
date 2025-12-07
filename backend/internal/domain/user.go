package domain

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// UserType represents the type of user account
type UserType string

const (
	// UserTypeRegistered is a fully registered user with credentials
	UserTypeRegistered UserType = "registered"
	// UserTypeAnonymous is a guest user without credentials (expires after 30 days)
	UserTypeAnonymous UserType = "anonymous"
)

// AnonymousExpiryDays is the number of days until an anonymous user expires
const AnonymousExpiryDays = 30

// User represents a user in the system
type User struct {
	ID           string     `json:"id"`
	UserType     UserType   `json:"user_type"`
	Email        *string    `json:"email,omitempty"`        // Nullable for anonymous users
	Name         *string    `json:"name,omitempty"`         // Nullable for anonymous users
	PasswordHash *string    `json:"-"`                      // Never expose, nullable for anonymous users
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`   // Set for anonymous users only
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// IsAnonymous returns true if the user is an anonymous/guest user
func (u *User) IsAnonymous() bool {
	return u.UserType == UserTypeAnonymous
}

// IsRegistered returns true if the user is a registered user
func (u *User) IsRegistered() bool {
	return u.UserType == UserTypeRegistered
}

// IsExpired returns true if the anonymous user has expired
func (u *User) IsExpired() bool {
	if !u.IsAnonymous() || u.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*u.ExpiresAt)
}

// GetDisplayName returns the user's name or a default for anonymous users
func (u *User) GetDisplayName() string {
	if u.Name != nil {
		return *u.Name
	}
	return "Guest User"
}

// GetEmail returns the user's email or empty string for anonymous users
func (u *User) GetEmail() string {
	if u.Email != nil {
		return *u.Email
	}
	return ""
}

// CalculateAnonymousExpiry returns the expiry time for a new anonymous user
func CalculateAnonymousExpiry() time.Time {
	return time.Now().Add(time.Duration(AnonymousExpiryDays) * 24 * time.Hour)
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

// ConvertGuestDTO is used to convert an anonymous user to a registered user
type ConvertGuestDTO struct {
	Email    string `json:"email" binding:"required"`
	Name     string `json:"name" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse is returned after successful login/register
type AuthResponse struct {
	User        User   `json:"user"`
	AccessToken string `json:"access_token"`
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
