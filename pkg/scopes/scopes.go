package scopes

import (
	"slices"
	"strings"
)

// filters scopes so that only scopes present in both userScopes and clientScopes are returned.
//
// If user has general scopes (i.e "*", "api:*"), the specific scopes from clientScopes are returned
// this also works for multi-level scopes (i.e "api:read:*" will match "api:read:user").
//
// Important: if clientScopes has general scopes but the user does not it will be ignored
func FilterScopes(userScopes, clientScopes []string) []string {
	if len(userScopes) == 0 || len(clientScopes) == 0 {
		return []string{} // Return empty slice, not nil
	}

	var filteredScopes []string
	seen := make(map[string]bool) // Track duplicates

	for _, clientScope := range clientScopes {
		// Skip malformed scopes (empty parts, wildcards in wrong places)
		if !isValidScope(clientScope) {
			continue
		}

		// Check if user has permission for this client scope
		if userHasPermission(userScopes, clientScope) {
			if !seen[clientScope] {
				filteredScopes = append(filteredScopes, clientScope)
				seen[clientScope] = true
			}
		}
	}

	return filteredScopes
}

// isValidScope checks if a scope string is properly formatted.
// Valid scopes:
//   - "*" (ultimate admin scope)
//   - "scope" (simple scope)
//   - "scope:action" (scoped action)
//   - "scope:*" (general scope with wildcard at end)
//   - "scope:subscope:*" (multi-level general scope)
//
// Invalid scopes:
//   - Empty string
//   - Contains empty parts (e.g., "api:", ":read")
//   - Wildcard in wrong position (e.g., "api:*:read")
func isValidScope(scope string) bool {
	if scope == "" || scope == "*" {
		return scope == "*" // "*" is valid, empty string is not
	}

	parts := strings.Split(scope, ":")
	for i, part := range parts {
		if part == "" {
			return false // No empty parts allowed
		}
		// Wildcard only allowed at the end
		if part == "*" && i != len(parts)-1 {
			return false
		}
	}
	return true
}

// userHasPermission checks if a user has permission for a specific client scope.
// Permission is granted if:
//   - User has the exact scope
//   - User has ultimate admin scope ("*")
//   - User has a general scope that covers the specific scope
//     (e.g., "api:*" covers "api:read", "api:read:*" covers "api:read:user")
func userHasPermission(userScopes []string, clientScope string) bool {
	// Check for exact match first
	if slices.Contains(userScopes, clientScope) {
		return true
	}

	// Check for ultimate admin scope
	if slices.Contains(userScopes, "*") {
		return true
	}

	// Check for general scopes that would grant this specific scope
	parts := strings.Split(clientScope, ":")

	// Build all possible general scopes that would grant this permission
	// For "api:read:user" we check: "api:*", "api:read:*"
	for i := 1; i < len(parts); i++ {
		generalScope := strings.Join(parts[:i], ":") + ":*"
		if slices.Contains(userScopes, generalScope) {
			return true
		}
	}

	return false
}
