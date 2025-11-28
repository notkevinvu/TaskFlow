package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/notkevinvu/taskflow/backend/internal/domain"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// SanitizeText trims whitespace and validates text length and content
func SanitizeText(text string, maxLength int, fieldName string) (string, error) {
	// Trim leading/trailing whitespace
	text = strings.TrimSpace(text)

	// Check length in runes (Unicode characters, not bytes)
	runes := []rune(text)
	if len(runes) > maxLength {
		return "", domain.NewValidationError(fieldName,
			fmt.Sprintf("exceeds maximum length of %d characters", maxLength))
	}

	// Check for control characters (0x00-0x1F, 0x7F-0x9F)
	for i, r := range runes {
		if unicode.IsControl(r) {
			return "", domain.NewValidationError(fieldName,
				fmt.Sprintf("contains invalid control character at position %d", i))
		}
	}

	return text, nil
}

// ValidateRequiredText validates that text is not empty after sanitization
func ValidateRequiredText(text string, maxLength int, fieldName string) (string, error) {
	sanitized, err := SanitizeText(text, maxLength, fieldName)
	if err != nil {
		return "", err
	}

	if sanitized == "" {
		return "", domain.NewValidationError(fieldName, "is required")
	}

	return sanitized, nil
}

// ValidateOptionalText validates optional text fields
func ValidateOptionalText(text *string, maxLength int, fieldName string) (*string, error) {
	if text == nil || *text == "" {
		return nil, nil
	}

	sanitized, err := SanitizeText(*text, maxLength, fieldName)
	if err != nil {
		return nil, err
	}

	// Return nil if empty after sanitization
	if sanitized == "" {
		return nil, nil
	}

	return &sanitized, nil
}

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	email = strings.TrimSpace(strings.ToLower(email))

	if email == "" {
		return domain.NewValidationError("email", "is required")
	}

	if !emailRegex.MatchString(email) {
		return domain.NewValidationError("email", "invalid format")
	}

	if len(email) > 255 {
		return domain.NewValidationError("email", "exceeds maximum length of 255 characters")
	}

	return nil
}

// ValidatePassword validates password strength
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return domain.NewValidationError("password", "must be at least 8 characters")
	}

	if len(password) > 100 {
		return domain.NewValidationError("password", "exceeds maximum length of 100 characters")
	}

	// Check for uppercase, lowercase, and number
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)

	if !hasUpper || !hasLower || !hasNumber {
		return domain.NewValidationError("password",
			"must contain at least one uppercase letter, one lowercase letter, and one number")
	}

	return nil
}

// ValidatePriority validates user priority (1-10 scale)
func ValidatePriority(priority int) error {
	if priority < 1 || priority > 10 {
		return domain.NewValidationError("user_priority", "must be between 1 and 10")
	}
	return nil
}

// ValidateCategory validates and sanitizes category names
func ValidateCategory(category *string) (*string, error) {
	if category == nil || *category == "" {
		return nil, nil
	}

	sanitized, err := SanitizeText(*category, 50, "category")
	if err != nil {
		return nil, err
	}

	// Additional validation: no special characters except spaces, hyphens, underscores
	validCategory := regexp.MustCompile(`^[a-zA-Z0-9\s\-_]+$`)
	if !validCategory.MatchString(sanitized) {
		return nil, domain.NewValidationError("category",
			"can only contain letters, numbers, spaces, hyphens, and underscores")
	}

	return &sanitized, nil
}

// ValidateStringSlice validates and sanitizes a slice of strings
func ValidateStringSlice(slice []string, maxLength int, maxItems int, fieldName string) ([]string, error) {
	if len(slice) == 0 {
		return nil, nil
	}

	if len(slice) > maxItems {
		return nil, domain.NewValidationError(fieldName,
			fmt.Sprintf("cannot exceed %d items", maxItems))
	}

	sanitized := make([]string, 0, len(slice))
	for i, item := range slice {
		cleaned, err := SanitizeText(item, maxLength, fmt.Sprintf("%s[%d]", fieldName, i))
		if err != nil {
			return nil, err
		}

		// Skip empty strings after sanitization
		if cleaned != "" {
			sanitized = append(sanitized, cleaned)
		}
	}

	return sanitized, nil
}
