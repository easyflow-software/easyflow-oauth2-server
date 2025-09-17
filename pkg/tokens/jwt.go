package tokens

import (
	"crypto/ed25519"
	"easyflow-oauth2-server/pkg/config"
	"easyflow-oauth2-server/pkg/database"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
	SessionToken TokenType = "session"
)

type JWTTokenPayload struct {
	jwt.RegisteredClaims
	Scopes []string  `json:"scopes,omitempty"`
	Type   TokenType `json:"type,omitempty"`
}

// generates a JWT token using the provided Ed25519 private key and payload
func generateJWT(secret *ed25519.PrivateKey, payload jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, payload)

	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", fmt.Errorf("Failed to sign token: %w", err)
	}

	return signedToken, nil
}

// / filterScopes filters scopes so that only scopes present in both userScopes and clientScopes are returned.
//
// If user has general scopes (i.e "*", "api:*"), the specific scopes from clientScopes are returned
// this also works for multi-level scopes (i.e "api:read:*" will match "api:read:user").
//
// Important: if clientScopes has general scopes but the user does not it will be ignored
func filterScopes(userScopes, clientScopes []string) []string {
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

func generateBasePayload(cfg *config.Config, user database.GetUserByEmailRow, client database.GetOAuthClientByClientIDRow) JWTTokenPayload {
	return JWTTokenPayload{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:   cfg.JwtIssuer,
			Subject:  user.ID.String(),
			Audience: []string{client.Name},
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}
}

func GenerateSessionToken(cfg *config.Config, key *ed25519.PrivateKey, user database.GetUserByEmailRow) (string, error) {
	var sessionTokenPayload = JWTTokenPayload{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.JwtIssuer,
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.JwtSessionTokenExpiryHours) * time.Hour)),
		},
	}

	sessionToken, err := generateJWT(key, sessionTokenPayload)
	if err != nil {
		return "", fmt.Errorf("Failed to generate session token: %w", err)
	}

	return sessionToken, nil
}

func GenerateTokens(cfg *config.Config, key *ed25519.PrivateKey, user database.GetUserByEmailRow, client database.GetOAuthClientByClientIDRow) (string, string, error) {
	var accessTokenPayload = generateBasePayload(cfg, user, client)
	accessTokenPayload.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.JwtAccessTokenExpiryMinutes) * time.Minute))
	accessTokenPayload.Scopes = filterScopes(user.Scopes, client.Scopes)

	accessToken, err := generateJWT(key, accessTokenPayload)
	if err != nil {
		return "", "", fmt.Errorf("Failed to generate access token: %w", err)
	}

	var refreshTokenPayload = generateBasePayload(cfg, user, client)
	refreshTokenPayload.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.JwtRefreshTokenExpiryDays) * 24 * time.Hour))

	refreshToken, err := generateJWT(key, refreshTokenPayload)
	if err != nil {
		return "", "", fmt.Errorf("Failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

func ValidateJwt(key *ed25519.PrivateKey, token string) (*JWTTokenPayload, error) {
	parsedToken, err := jwt.ParseWithClaims(token, &JWTTokenPayload{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}
		return key.Public(), nil
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to parse token: %w", err)
	}

	claims, ok := parsedToken.Claims.(*JWTTokenPayload)
	if !ok || !parsedToken.Valid {
		return nil, fmt.Errorf("Invalid token")
	}

	return claims, nil
}
