package domain

import (
	"errors"
	"time"
)

var (
	ErrPreferencesNotFound = errors.New("user preferences not found")
	ErrCategoryPrefNotFound = errors.New("category preference not found")
)

// UserPreferences stores user-level settings
type UserPreferences struct {
	UserID                   string             `json:"user_id"`
	DefaultDueDateCalculation DueDateCalculation `json:"default_due_date_calculation"`
	CreatedAt                time.Time          `json:"created_at"`
	UpdatedAt                time.Time          `json:"updated_at"`
}

// CategoryPreference stores per-category settings
type CategoryPreference struct {
	ID                 string             `json:"id"`
	UserID             string             `json:"user_id"`
	Category           string             `json:"category"`
	DueDateCalculation DueDateCalculation `json:"due_date_calculation"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}

// UpdateUserPreferencesDTO is used for updating user preferences
type UpdateUserPreferencesDTO struct {
	DefaultDueDateCalculation *DueDateCalculation `json:"default_due_date_calculation,omitempty"`
}

// SetCategoryPreferenceDTO is used for setting a category preference
type SetCategoryPreferenceDTO struct {
	Category           string             `json:"category" binding:"required,max=50"`
	DueDateCalculation DueDateCalculation `json:"due_date_calculation" binding:"required"`
}

// Validate validates the DTO
func (dto *SetCategoryPreferenceDTO) Validate() error {
	if dto.Category == "" {
		return errors.New("category is required")
	}
	if len(dto.Category) > 50 {
		return errors.New("category must be 50 characters or less")
	}
	return dto.DueDateCalculation.Validate()
}

// AllPreferences combines user and category preferences for API response
type AllPreferences struct {
	UserPreferences      *UserPreferences      `json:"user_preferences"`
	CategoryPreferences  []CategoryPreference  `json:"category_preferences"`
}

// GetEffectiveDueDateCalculation returns the due date calculation mode
// based on user and category preferences (category takes precedence)
func GetEffectiveDueDateCalculation(
	userPref *UserPreferences,
	categoryPrefs []CategoryPreference,
	category *string,
) DueDateCalculation {
	// Default fallback
	defaultCalc := DueDateFromOriginal

	// Check category-specific preference first
	if category != nil {
		for _, cp := range categoryPrefs {
			if cp.Category == *category {
				return cp.DueDateCalculation
			}
		}
	}

	// Fall back to user preference
	if userPref != nil {
		return userPref.DefaultDueDateCalculation
	}

	return defaultCalc
}
