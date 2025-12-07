package domain

import "fmt"

// Feature represents a TaskFlow feature that can be gated
type Feature string

const (
	// Core features available to all users
	FeatureTaskCRUD       Feature = "task_crud"
	FeatureTaskPriority   Feature = "task_priority"
	FeatureTaskCategories Feature = "task_categories"
	FeatureTaskBump       Feature = "task_bump"
	FeatureAnalytics      Feature = "analytics"
	FeatureInsights       Feature = "insights"

	// Advanced features restricted to registered users
	FeatureSubtasks      Feature = "subtasks"
	FeatureDependencies  Feature = "dependencies"
	FeatureTemplates     Feature = "templates"
	FeatureGamification  Feature = "gamification"
	FeatureRecurring     Feature = "recurring"
)

// anonymousAllowedFeatures defines which features are accessible to anonymous users
var anonymousAllowedFeatures = map[Feature]bool{
	// Core features - ALLOWED for anonymous users
	FeatureTaskCRUD:       true,
	FeatureTaskPriority:   true,
	FeatureTaskCategories: true,
	FeatureTaskBump:       true,
	FeatureAnalytics:      true,
	FeatureInsights:       true,

	// Advanced features - RESTRICTED to registered users
	FeatureSubtasks:     false,
	FeatureDependencies: false,
	FeatureTemplates:    false,
	FeatureGamification: false,
	FeatureRecurring:    false,
}

// CanAccessFeature checks if a user type can access a specific feature
func CanAccessFeature(userType UserType, feature Feature) bool {
	// Registered users have access to all features
	if userType == UserTypeRegistered {
		return true
	}

	// Anonymous users have restricted access
	allowed, exists := anonymousAllowedFeatures[feature]
	if !exists {
		// Unknown features are restricted by default
		return false
	}
	return allowed
}

// NewFeatureGatedError creates a standardized error for when a feature is gated
func NewFeatureGatedError(feature Feature) *ForbiddenError {
	return NewForbiddenError(
		"feature",
		fmt.Sprintf("The '%s' feature requires a registered account. Sign up to unlock!", feature),
	)
}

// GetRestrictedFeatures returns the list of features not available to anonymous users
func GetRestrictedFeatures() []Feature {
	restricted := []Feature{}
	for feature, allowed := range anonymousAllowedFeatures {
		if !allowed {
			restricted = append(restricted, feature)
		}
	}
	return restricted
}

// GetAllowedFeatures returns the list of features available to anonymous users
func GetAllowedFeatures() []Feature {
	allowed := []Feature{}
	for feature, isAllowed := range anonymousAllowedFeatures {
		if isAllowed {
			allowed = append(allowed, feature)
		}
	}
	return allowed
}
