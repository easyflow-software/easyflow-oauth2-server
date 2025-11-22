// Package tokens provides functionalities for generating and validating JWT tokens.
package tokens

import (
	"crypto/ed25519"
	"crypto/rand"
	"easyflow-oauth2-server/internal/database"
	"easyflow-oauth2-server/internal/server/config"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Error definitions.
var (
	ErrFailedToSignToken            = errors.New("failed to sign token")
	ErrFailedToGenerateSessionToken = errors.New("failed to generate session token")
	ErrFailedToGenerateAccessToken  = errors.New("failed to generate access token")
	ErrFailedToGenerateRefreshToken = errors.New("failed to generate refresh token")
	ErrUnexpectedSigningMethod      = errors.New("unexpected signing method")
	ErrFailedToParseToken           = errors.New("failed to parse token")
	ErrInvalidToken                 = errors.New("invalid token")
)

// TokenType represents the type of JWT token.
type TokenType string

// Defined token types.
const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
	SessionToken TokenType = "session"
)

// JWTTokenPayload represents the payload of a JWT token, including standard claims and custom fields.
type JWTTokenPayload struct {
	jwt.RegisteredClaims
	Scopes []string  `json:"scopes"`
	Type   TokenType `json:"type,omitempty"`
}

// generates a JWT token using the provided Ed25519 private key and payload.
func generateJWT(secret *ed25519.PrivateKey, payload jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, payload)

	signedToken, err := token.SignedString(secret)
	if err != nil {
		return "", ErrFailedToSignToken
	}

	return signedToken, nil
}

func generateBasePayload(
	cfg *config.Config,
	userID string,
	client *database.GetOAuthClientByClientIDRow,
	sessionID string,
) JWTTokenPayload {
	return JWTTokenPayload{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:   cfg.BaseURL,
			Subject:  userID,
			Audience: []string{client.Name},
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ID:       sessionID,
		},
	}
}

// GenerateSessionToken generates a session token using the provided data.
// It creates a JWT token with appropriate claims and expiration time based on the configuration.
func GenerateSessionToken(
	cfg *config.Config,
	key *ed25519.PrivateKey,
	userID string,
) (string, error) {
	var sessionTokenPayload = JWTTokenPayload{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:   cfg.BaseURL,
			Subject:  userID,
			IssuedAt: jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(time.Duration(cfg.JwtSessionTokenExpiryHours) * time.Hour),
			),
		},
		Type: SessionToken,
	}

	sessionToken, err := generateJWT(key, sessionTokenPayload)
	if err != nil {
		return "", ErrFailedToGenerateSessionToken
	}

	return sessionToken, nil
}

// GenerateTokens generates an access token and a refresh token using the provided data.
// It creates JWT tokens with appropriate claims and expiration times based on the OAuth client settings.
func GenerateTokens(
	cfg *config.Config,
	key *ed25519.PrivateKey,
	userID string,
	client *database.GetOAuthClientByClientIDRow,
	scopes []string,
	sessionID string,
) (string, string, error) {
	var accessTokenPayload = generateBasePayload(cfg, userID, client, sessionID)
	accessTokenPayload.ExpiresAt = jwt.NewNumericDate(
		time.Now().Add(time.Duration(client.AccessTokenValidDuration) * time.Second),
	)
	accessTokenPayload.Scopes = scopes

	accessToken, err := generateJWT(key, accessTokenPayload)
	if err != nil {
		return "", "", ErrFailedToGenerateAccessToken
	}

	refreshToken := rand.Text() + rand.Text()

	return accessToken, refreshToken, nil
}

// ValidateJwt validates a JWT token using the provided Ed25519 private key and returns the token payload if valid.
// It checks the signing method and ensures the token is valid.
func ValidateJwt(key *ed25519.PrivateKey, token string) (*JWTTokenPayload, error) {
	parsedToken, err := jwt.ParseWithClaims(
		token,
		&JWTTokenPayload{},
		func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
				return nil, ErrUnexpectedSigningMethod
			}
			return key.Public(), nil
		},
	)
	if err != nil {
		return nil, ErrFailedToParseToken
	}

	claims, ok := parsedToken.Claims.(*JWTTokenPayload)
	if !ok || !parsedToken.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
