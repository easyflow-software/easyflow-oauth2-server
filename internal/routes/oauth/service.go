package oauth

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"easyflow-oauth2-server/internal/shared/errors"
	"easyflow-oauth2-server/internal/shared/service"
	"easyflow-oauth2-server/pkg/database"
	"easyflow-oauth2-server/pkg/scopes"
	"easyflow-oauth2-server/pkg/tokens"
	"encoding/base64"
	e "errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/fx"
)

// Service handles OAuth2 business logic.
type Service struct {
	*service.BaseService
	key *ed25519.PrivateKey
}

// ServiceParams holds dependencies for OAuthService.
type ServiceParams struct {
	fx.In
	service.BaseServiceParams
	Key *ed25519.PrivateKey
}

// NewOAuthService creates a new instance of OAuthService.
func NewOAuthService(deps ServiceParams) *Service {
	baseService := service.NewBaseService("OAuthService", deps.BaseServiceParams)
	return &Service{
		BaseService: baseService,
		key:         deps.Key,
	}
}

// GetClient retrieves an OAuth client by client ID.
func (s *Service) GetClient(
	ctx context.Context,
	clientID string,
	clientIP string,
) (*database.GetOAuthClientByClientIDRow, *errors.APIError) {
	logger := s.GetLogger(clientIP)
	client, err := s.Queries.GetOAuthClientByClientID(ctx, clientID)
	if err != nil {
		if e.Is(err, sql.ErrNoRows) {
			logger.PrintfDebug("Failed to retrieve client with client id: %s", clientID)
			return nil, &errors.APIError{
				Code:    http.StatusNotFound,
				Error:   errors.InvalidClientID,
				Details: "Client not found",
			}
		}
		return nil, &errors.APIError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to get client" + err.Error(),
		}
	}

	return &client, nil
}

// Authorize creates an authorization code for the OAuth flow.
func (s *Service) Authorize(
	ctx context.Context,
	client *database.GetOAuthClientByClientIDRow,
	codeChallenge string,
	userID string,
	clientIP string,
) (*string, *errors.APIError) {
	logger := s.GetLogger(clientIP)
	code := rand.Text()

	key := fmt.Sprintf("authorization-code:%s", code)

	values := map[string]string{
		"codeChallange": codeChallenge,
		"clientId":      client.ClientID,
		"userId":        userID,
		"scopes":        strings.Join(client.Scopes, " "),
	}

	if err := s.CacheHset(ctx, key, values, service.WithTTL(10*time.Minute)); err != nil {
		logger.PrintfError("Failed to store authorization code: %v", err)
		return nil, &errors.APIError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to store authorization code",
		}
	}
	return &code, nil
}

// AuthorizationCodeFlow handles the authorization code grant flow.
func (s *Service) AuthorizationCodeFlow(
	ctx context.Context,
	client *database.GetOAuthClientByClientIDRow,
	code, codeVerifier string,
	clientIP string,
) (*string, *string, []string, *errors.APIError) {
	logger := s.GetLogger(clientIP)
	key := fmt.Sprintf("authorization-code:%s", code)

	codeStore, err := s.CacheHgetall(ctx, key, service.WithExpiration(1*time.Minute))
	if err != nil {
		logger.PrintfError("Failed to get authorization code: %v", err)
		return nil, nil, []string{}, &errors.APIError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to get authorization code",
		}
	}

	if len(codeStore) == 0 {
		logger.PrintfWarning("Authorization code not found: %s", code)
		return nil, nil, []string{}, &errors.APIError{
			Code:    http.StatusBadRequest,
			Error:   errors.InvalidCode,
			Details: "Invalid authorization code",
		}
	}

	if codeStore["clientId"] != client.ClientID {
		logger.PrintfWarning("Client ID does not match authorization code: %s", code)
		return nil, nil, []string{}, &errors.APIError{
			Code:    http.StatusBadRequest,
			Error:   errors.InvalidClientID,
			Details: "Client ID does not match authorization code",
		}
	}

	hash := sha256.Sum256([]byte(codeVerifier))

	hashStr := base64.RawURLEncoding.EncodeToString(hash[:])

	if codeStore["codeChallange"] != hashStr {
		logger.PrintfWarning("Code verifier does not match code challenge: %s", code)
		return nil, nil, []string{}, &errors.APIError{
			Code:    http.StatusBadRequest,
			Error:   errors.InvalidCodeVerifier,
			Details: "Invalid code verifier",
		}
	}

	ID, err := uuid.Parse(codeStore["userId"])
	if err != nil {
		logger.PrintfError("Failed to parse user ID: %v", err)
		return nil, nil, []string{}, &errors.APIError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to parse user ID",
		}
	}

	user, err := s.Queries.GetUserWithRolesAndScopes(ctx, ID)
	if err != nil {
		if e.Is(err, sql.ErrNoRows) {
			logger.PrintfWarning("User not found: %s", codeStore["userId"])
			return nil, nil, []string{}, &errors.APIError{
				Code:    http.StatusNotFound,
				Error:   errors.NotFound,
				Details: "User not found",
			}
		}
		logger.PrintfError("Failed to get user: %v", err)
		return nil, nil, []string{}, &errors.APIError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to get user",
		}
	}
	logger.PrintfDebug("Found user with ID: %s", user.ID)

	userScopes := scopes.FilterScopes(user.Scopes, client.Scopes)

	sessionID := uuid.New()

	accessToken, refreshToken, err := tokens.GenerateTokens(
		s.Config,
		s.key,
		user.ID.String(),
		client,
		userScopes,
		sessionID.String(),
	)
	if err != nil {
		logger.PrintfError("Failed to generate tokens: %v", err)
		return nil, nil, []string{}, &errors.APIError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to generate tokens",
		}
	}

	sessionKey := fmt.Sprintf("session:%s", refreshToken)
	sessionData := map[string]string{
		"sessionID": sessionID.String(),
		"subject":   user.ID.String(),
		"scopes":    strings.Join(userScopes, ","),
	}

	if err := s.CacheHset(ctx, sessionKey, sessionData, service.WithTTL(time.Duration(client.RefreshTokenValidDuration)*time.Second)); err != nil {
		logger.PrintfError("Failed to store session: %v", err)
		return nil, nil, []string{}, &errors.APIError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to store session",
		}
	}
	logger.PrintfDebug("Stored session with ID: %s", sessionID.String())

	if err := s.CacheDel(ctx, key); err != nil {
		logger.PrintfError("Failed to delete authorization code: %v", err)
	}
	logger.PrintfDebug("Deleted authorization code: %s", code)

	return &accessToken, &refreshToken, userScopes, nil
}

// ClientCredentialsFlow handles the client credentials grant flow.
func (s *Service) ClientCredentialsFlow(
	client *database.GetOAuthClientByClientIDRow,
	clientIP string,
) (*string, []string, *errors.APIError) {
	logger := s.GetLogger(clientIP)
	sessionToken := uuid.New()

	clientScopes := client.Scopes

	accessToken, _, err := tokens.GenerateTokens(
		s.Config,
		s.key,
		client.ClientID,
		client,
		clientScopes,
		sessionToken.String(),
	)
	if err != nil {
		logger.PrintfError("Failed to generate access token: %v", err)
		return nil, []string{}, &errors.APIError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to generate access token",
		}
	}

	return &accessToken, clientScopes, nil
}

// RefreshTokenFlow handles the refresh token grant flow.
func (s *Service) RefreshTokenFlow(
	ctx context.Context,
	client *database.GetOAuthClientByClientIDRow,
	refreshToken string,
	clientIP string,
) (*string, *string, []string, *errors.APIError) {
	logger := s.GetLogger(clientIP)
	sessionKey := fmt.Sprintf("session:%s", refreshToken)

	session, err := s.CacheHgetall(ctx, sessionKey, service.WithoutLocalCache())
	if err != nil {
		logger.PrintfError("Failed to get session: %v", err)
		return nil, nil, []string{}, &errors.APIError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to get session",
		}
	}

	if len(session) == 0 {
		logger.PrintfWarning("Session not found: %s", refreshToken)
		return nil, nil, []string{}, &errors.APIError{
			Code:    http.StatusBadRequest,
			Error:   errors.InvalidRefreshToken,
			Details: "Invalid refresh token",
		}
	}

	// TODO: Add check for changed session scopes if so refuse to issue new tokens
	sessionScopes := []string{}
	if session["scopes"] != "" {
		sessionScopes = strings.Split(session["scopes"], ",")
	}
	accessToken, newRefreshToken, err := tokens.GenerateTokens(
		s.Config,
		s.key,
		session["subject"],
		client,
		sessionScopes,
		session["sessionID"],
	)
	if err != nil {
		logger.PrintfError("Failed to generate tokens: %v", err)
		return nil, nil, []string{}, &errors.APIError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to generate tokens",
		}
	}

	newSessionKey := fmt.Sprintf("session:%s", newRefreshToken)
	newSessionData := map[string]string{
		"sessionID": session["sessionID"],
		"subject":   session["subject"],
		"scopes":    session["scopes"],
	}

	if err := s.CacheHset(ctx, newSessionKey, newSessionData, service.WithTTL(time.Duration(client.RefreshTokenValidDuration)*time.Second)); err != nil {
		logger.PrintfError("Failed to store new session: %v", err)
		return nil, nil, []string{}, &errors.APIError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to store new session",
		}
	}
	logger.PrintfDebug("Stored new session with refresh token: %s", newRefreshToken)

	if err := s.CacheDel(ctx, sessionKey); err != nil {
		logger.PrintfError("Failed to delete old session: %v", err)
	} else {
		logger.PrintfDebug("Deleted old session: %s", refreshToken)
	}

	return &accessToken, &newRefreshToken, sessionScopes, nil
}
