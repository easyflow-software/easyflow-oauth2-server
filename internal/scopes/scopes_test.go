package scopes

import (
	"reflect"
	"testing"
)

func validateResult(result, expected []string) bool {
	if len(result) == 0 && len(expected) == 0 {
		return true
	}

	return reflect.DeepEqual(result, expected)
}

func TestFilterScopes(t *testing.T) {
	tests := []struct {
		name         string
		userScopes   []string
		clientScopes []string
		expected     []string
	}{
		// Basic exact matches
		{
			name:         "exact match single scope",
			userScopes:   []string{"api:read"},
			clientScopes: []string{"api:read"},
			expected:     []string{"api:read"},
		},
		{
			name:         "exact match multiple scopes",
			userScopes:   []string{"api:read", "api:write"},
			clientScopes: []string{"api:read", "api:write"},
			expected:     []string{"api:read", "api:write"},
		},
		{
			name:         "partial match",
			userScopes:   []string{"api:read"},
			clientScopes: []string{"api:read", "api:write"},
			expected:     []string{"api:read"},
		},
		{
			name:         "no match",
			userScopes:   []string{"api:read"},
			clientScopes: []string{"api:write"},
			expected:     []string{},
		},

		// Ultimate admin scope "*"
		{
			name:         "user has ultimate admin scope",
			userScopes:   []string{"*"},
			clientScopes: []string{"api:read", "api:write", "user:delete"},
			expected:     []string{"api:read", "api:write", "user:delete"},
		},
		{
			name:         "user has ultimate admin and specific scopes",
			userScopes:   []string{"*", "api:read"},
			clientScopes: []string{"api:read", "user:write"},
			expected:     []string{"api:read", "user:write"},
		},

		// General scopes with wildcard
		{
			name:         "user has general api scope",
			userScopes:   []string{"api:*"},
			clientScopes: []string{"api:read", "api:write", "user:delete"},
			expected:     []string{"api:read", "api:write"},
		},
		{
			name:         "user has multiple general scopes",
			userScopes:   []string{"api:*", "user:*"},
			clientScopes: []string{"api:read", "user:write", "admin:delete"},
			expected:     []string{"api:read", "user:write"},
		},

		// Multi-level scopes
		{
			name:         "user has multi-level general scope",
			userScopes:   []string{"api:read:*"},
			clientScopes: []string{"api:read:user", "api:read:posts", "api:write:user"},
			expected:     []string{"api:read:user", "api:read:posts"},
		},
		{
			name:         "user has multiple multi-level general scopes",
			userScopes:   []string{"api:read:*", "api:write:*"},
			clientScopes: []string{"api:read:user", "api:write:posts", "user:delete"},
			expected:     []string{"api:read:user", "api:write:posts"},
		},

		// Important: client general scopes ignored if user doesn't have them
		{
			name:         "client has general scope but user has specific",
			userScopes:   []string{"api:read"},
			clientScopes: []string{"api:*"},
			expected:     []string{},
		},
		{
			name:         "client has ultimate scope but user has specific",
			userScopes:   []string{"api:read"},
			clientScopes: []string{"*"},
			expected:     []string{},
		},

		// Mixed scenarios
		{
			name:         "mixed general and specific scopes",
			userScopes:   []string{"api:*", "user:read"},
			clientScopes: []string{"api:write", "user:read", "user:delete", "admin:read"},
			expected:     []string{"api:write", "user:read"},
		},
		{
			name:         "hierarchical scope matching",
			userScopes:   []string{"api:*"},
			clientScopes: []string{"api:read:user", "api:write:posts:meta"},
			expected:     []string{"api:read:user", "api:write:posts:meta"},
		},

		// Edge cases
		{
			name:         "empty user scopes",
			userScopes:   []string{},
			clientScopes: []string{"api:read"},
			expected:     []string{},
		},
		{
			name:         "empty client scopes",
			userScopes:   []string{"api:read"},
			clientScopes: []string{},
			expected:     []string{},
		},
		{
			name:         "both empty",
			userScopes:   []string{},
			clientScopes: []string{},
			expected:     []string{},
		},
		{
			name:         "nil slices",
			userScopes:   nil,
			clientScopes: nil,
			expected:     []string{},
		},

		// Complex hierarchical matching
		{
			name:       "deep nested scope matching",
			userScopes: []string{"api:read:user:*"},
			clientScopes: []string{
				"api:read:user:profile",
				"api:read:user:settings",
				"api:read:posts",
			},
			expected: []string{"api:read:user:profile", "api:read:user:settings"},
		},
		{
			name:       "multiple hierarchy levels",
			userScopes: []string{"api:*", "user:read:*"},
			clientScopes: []string{
				"api:write:posts:meta",
				"user:read:profile:public",
				"admin:delete",
			},
			expected: []string{"api:write:posts:meta", "user:read:profile:public"},
		},

		// Order independence
		{
			name:         "different order same result",
			userScopes:   []string{"user:read", "api:*"},
			clientScopes: []string{"api:write", "user:read", "admin:delete"},
			expected:     []string{"api:write", "user:read"},
		},

		// Duplicate handling
		{
			name:         "duplicate scopes in input",
			userScopes:   []string{"api:read", "api:read", "api:*"},
			clientScopes: []string{"api:read", "api:write", "api:read"},
			expected:     []string{"api:read", "api:write"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterScopes(tt.userScopes, tt.clientScopes)

			// Sort both slices to make comparison order-independent
			if !validateResult(result, tt.expected) {
				t.Errorf("filterScopes() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

// Additional edge case tests.
func TestFilterScopesEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		userScopes   []string
		clientScopes []string
		expected     []string
	}{
		{
			name:         "malformed scopes ignored",
			userScopes:   []string{"api:", ":read", "api:read"},
			clientScopes: []string{"api:read", "api:", ":write"},
			expected:     []string{"api:read"},
		},
		{
			name:         "single character scopes",
			userScopes:   []string{"*"},
			clientScopes: []string{"a", "b:c"},
			expected:     []string{"a", "b:c"},
		},
		{
			name:         "wildcard at wrong position ignored",
			userScopes:   []string{"api:read:*:invalid"},
			clientScopes: []string{"api:read:user"},
			expected:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FilterScopes(tt.userScopes, tt.clientScopes)
			if !validateResult(result, tt.expected) {
				t.Errorf("filterScopes() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
