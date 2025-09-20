package tokens

import (
	"crypto/ed25519"
	"easyflow-oauth2-server/pkg/config"
	"easyflow-oauth2-server/pkg/database"
	"fmt"
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

func generateBasePayload(cfg *config.Config, userID string, client *database.GetOAuthClientByClientIDRow, sessionID string) JWTTokenPayload {
	return JWTTokenPayload{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:   cfg.JwtIssuer,
			Subject:  userID,
			Audience: []string{client.Name},
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ID:       sessionID,
		},
	}
}

func GenerateSessionToken(cfg *config.Config, key *ed25519.PrivateKey, userID string) (string, error) {
	var sessionTokenPayload = JWTTokenPayload{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    cfg.JwtIssuer,
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.JwtSessionTokenExpiryHours) * time.Hour)),
		},
		Type: SessionToken,
	}

	sessionToken, err := generateJWT(key, sessionTokenPayload)
	if err != nil {
		return "", fmt.Errorf("Failed to generate session token: %w", err)
	}

	return sessionToken, nil
}

func GenerateTokens(cfg *config.Config, key *ed25519.PrivateKey, userID string, client *database.GetOAuthClientByClientIDRow, scopes []string, sessionID string) (string, string, error) {
	var accessTokenPayload = generateBasePayload(cfg, userID, client, sessionID)
	accessTokenPayload.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Duration(cfg.JwtAccessTokenExpiryMinutes) * time.Minute))
	accessTokenPayload.Scopes = scopes

	accessToken, err := generateJWT(key, accessTokenPayload)
	if err != nil {
		return "", "", fmt.Errorf("Failed to generate access token: %w", err)
	}

	var refreshTokenPayload = generateBasePayload(cfg, userID, client, sessionID)
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
