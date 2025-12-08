package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// CanAccessFeature Tests
// =============================================================================

func TestCanAccessFeature_RegisteredUserHasAllAccess(t *testing.T) {
	// Registered users should have access to ALL features
	allFeatures := []Feature{
		// Core features
		FeatureTaskCRUD,
		FeatureTaskPriority,
		FeatureTaskCategories,
		FeatureTaskBump,
		FeatureAnalytics,
		FeatureInsights,
		// Advanced features
		FeatureSubtasks,
		FeatureDependencies,
		FeatureTemplates,
		FeatureGamification,
		FeatureRecurring,
	}

	for _, feature := range allFeatures {
		t.Run(string(feature), func(t *testing.T) {
			assert.True(t, CanAccessFeature(UserTypeRegistered, feature),
				"Registered users should have access to %s", feature)
		})
	}
}

func TestCanAccessFeature_AnonymousUserCoreFeatures(t *testing.T) {
	// Anonymous users should have access to core features
	coreFeatures := []Feature{
		FeatureTaskCRUD,
		FeatureTaskPriority,
		FeatureTaskCategories,
		FeatureTaskBump,
		FeatureAnalytics,
		FeatureInsights,
	}

	for _, feature := range coreFeatures {
		t.Run(string(feature), func(t *testing.T) {
			assert.True(t, CanAccessFeature(UserTypeAnonymous, feature),
				"Anonymous users should have access to core feature %s", feature)
		})
	}
}

func TestCanAccessFeature_AnonymousUserRestrictedFeatures(t *testing.T) {
	// Anonymous users should NOT have access to advanced features
	restrictedFeatures := []Feature{
		FeatureSubtasks,
		FeatureDependencies,
		FeatureTemplates,
		FeatureGamification,
		FeatureRecurring,
	}

	for _, feature := range restrictedFeatures {
		t.Run(string(feature), func(t *testing.T) {
			assert.False(t, CanAccessFeature(UserTypeAnonymous, feature),
				"Anonymous users should NOT have access to advanced feature %s", feature)
		})
	}
}

func TestCanAccessFeature_UnknownFeatureDefaultsToRestricted(t *testing.T) {
	// Unknown features should be restricted for anonymous users
	unknownFeature := Feature("unknown_feature")

	assert.True(t, CanAccessFeature(UserTypeRegistered, unknownFeature),
		"Registered users should have access to unknown features")
	assert.False(t, CanAccessFeature(UserTypeAnonymous, unknownFeature),
		"Anonymous users should NOT have access to unknown features")
}

// =============================================================================
// GetRestrictedFeatures Tests
// =============================================================================

func TestGetRestrictedFeatures_ReturnsCorrectFeatures(t *testing.T) {
	restricted := GetRestrictedFeatures()

	// Should contain exactly the advanced features
	expectedRestricted := map[Feature]bool{
		FeatureSubtasks:     true,
		FeatureDependencies: true,
		FeatureTemplates:    true,
		FeatureGamification: true,
		FeatureRecurring:    true,
	}

	assert.Len(t, restricted, len(expectedRestricted),
		"Should return exactly %d restricted features", len(expectedRestricted))

	for _, feature := range restricted {
		assert.True(t, expectedRestricted[feature],
			"Feature %s should be in the restricted list", feature)
	}
}

func TestGetRestrictedFeatures_DoesNotContainCoreFeatures(t *testing.T) {
	restricted := GetRestrictedFeatures()

	coreFeatures := []Feature{
		FeatureTaskCRUD,
		FeatureTaskPriority,
		FeatureTaskCategories,
		FeatureTaskBump,
		FeatureAnalytics,
		FeatureInsights,
	}

	restrictedMap := make(map[Feature]bool)
	for _, f := range restricted {
		restrictedMap[f] = true
	}

	for _, core := range coreFeatures {
		assert.False(t, restrictedMap[core],
			"Core feature %s should not be in restricted list", core)
	}
}

// =============================================================================
// GetAllowedFeatures Tests
// =============================================================================

func TestGetAllowedFeatures_ReturnsCorrectFeatures(t *testing.T) {
	allowed := GetAllowedFeatures()

	// Should contain exactly the core features
	expectedAllowed := map[Feature]bool{
		FeatureTaskCRUD:       true,
		FeatureTaskPriority:   true,
		FeatureTaskCategories: true,
		FeatureTaskBump:       true,
		FeatureAnalytics:      true,
		FeatureInsights:       true,
	}

	assert.Len(t, allowed, len(expectedAllowed),
		"Should return exactly %d allowed features", len(expectedAllowed))

	for _, feature := range allowed {
		assert.True(t, expectedAllowed[feature],
			"Feature %s should be in the allowed list", feature)
	}
}

func TestGetAllowedFeatures_DoesNotContainAdvancedFeatures(t *testing.T) {
	allowed := GetAllowedFeatures()

	advancedFeatures := []Feature{
		FeatureSubtasks,
		FeatureDependencies,
		FeatureTemplates,
		FeatureGamification,
		FeatureRecurring,
	}

	allowedMap := make(map[Feature]bool)
	for _, f := range allowed {
		allowedMap[f] = true
	}

	for _, adv := range advancedFeatures {
		assert.False(t, allowedMap[adv],
			"Advanced feature %s should not be in allowed list", adv)
	}
}

// =============================================================================
// NewFeatureGatedError Tests
// =============================================================================

func TestNewFeatureGatedError_CreatesCorrectError(t *testing.T) {
	tests := []struct {
		feature         Feature
		expectedMessage string
	}{
		{
			feature:         FeatureSubtasks,
			expectedMessage: "forbidden: cannot access 'subtasks' (requires registered account) subtasks",
		},
		{
			feature:         FeatureDependencies,
			expectedMessage: "forbidden: cannot access 'dependencies' (requires registered account) dependencies",
		},
		{
			feature:         FeatureTemplates,
			expectedMessage: "forbidden: cannot access 'templates' (requires registered account) templates",
		},
		{
			feature:         FeatureGamification,
			expectedMessage: "forbidden: cannot access 'gamification' (requires registered account) gamification",
		},
		{
			feature:         FeatureRecurring,
			expectedMessage: "forbidden: cannot access 'recurring' (requires registered account) recurring",
		},
	}

	for _, tt := range tests {
		t.Run(string(tt.feature), func(t *testing.T) {
			err := NewFeatureGatedError(tt.feature)

			assert.NotNil(t, err)
			assert.Equal(t, string(tt.feature), err.Resource)
			assert.Contains(t, err.Action, "requires registered account")
		})
	}
}

func TestNewFeatureGatedError_IsForbiddenError(t *testing.T) {
	err := NewFeatureGatedError(FeatureSubtasks)

	// Should be a ForbiddenError
	var forbiddenErr *ForbiddenError
	assert.IsType(t, forbiddenErr, err)
}

// =============================================================================
// Feature Constants Tests
// =============================================================================

func TestFeatureConstants_AreUnique(t *testing.T) {
	allFeatures := []Feature{
		FeatureTaskCRUD,
		FeatureTaskPriority,
		FeatureTaskCategories,
		FeatureTaskBump,
		FeatureAnalytics,
		FeatureInsights,
		FeatureSubtasks,
		FeatureDependencies,
		FeatureTemplates,
		FeatureGamification,
		FeatureRecurring,
	}

	seen := make(map[Feature]bool)
	for _, f := range allFeatures {
		assert.False(t, seen[f], "Feature %s is duplicated", f)
		seen[f] = true
	}
}

func TestFeatureConstants_AreNotEmpty(t *testing.T) {
	allFeatures := []Feature{
		FeatureTaskCRUD,
		FeatureTaskPriority,
		FeatureTaskCategories,
		FeatureTaskBump,
		FeatureAnalytics,
		FeatureInsights,
		FeatureSubtasks,
		FeatureDependencies,
		FeatureTemplates,
		FeatureGamification,
		FeatureRecurring,
	}

	for _, f := range allFeatures {
		assert.NotEmpty(t, string(f), "Feature constant should not be empty")
	}
}
