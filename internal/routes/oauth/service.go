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
	"github.com/valkey-io/valkey-go"
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
) (*database.GetOAuthClientByClientIDRow, *errors.APIError) {
	client, err := s.Queries.GetOAuthClientByClientID(ctx, clientID)
	if err != nil {
		if e.Is(err, sql.ErrNoRows) {
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

	query := s.Valkey.B().
		Hset().
		Key(key).
		FieldValue().
		FieldValue("codeChallange", codeChallenge).
		FieldValue("clientId", client.ClientID).
		FieldValue("userId", userID).
		FieldValue("scopes", strings.Join(client.Scopes, " ")).
		Build()
	if s.Valkey.Do(ctx, query).Error() != nil {
		logger.PrintfError(
			"Failed to store authorization code: %v",
			s.Valkey.Do(ctx, query).Error(),
		)
		return nil, &errors.APIError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to store authorization code",
		}
	}

	query = s.Valkey.B().Expire().Key(key).Seconds(600).Build()
	if s.Valkey.Do(ctx, query).Error() != nil {
		logger.PrintfError(
			"Failed to set expiration for authorization code: %v",
			s.Valkey.Do(ctx, query).Error(),
		)
		return nil, &errors.APIError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to set expiration for authorization code",
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

	query := s.Valkey.B().Hgetall().Key(key).Cache()

	res := s.Valkey.DoCache(ctx, query, 1*time.Minute)
	if res.Error() != nil {
		logger.PrintfError("Failed to get authorization code: %v", res.Error())
		return nil, nil, []string{}, &errors.APIError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to get authorization code",
		}
	}

	codeStore, err := res.AsStrMap()
	if err != nil {
		logger.PrintfError("Failed to parse authorization code: %v", err)
		return nil, nil, []string{}, &errors.APIError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to parse authorization code",
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

	var queries valkey.Commands
	sessionKey := fmt.Sprintf("session:%s", refreshToken)
	// Convert to single commands to use auto-pipelining
	queries = append(
		queries,
		s.Valkey.B().
			Hset().
			Key(sessionKey).
			FieldValue().
			FieldValue("sessionID", sessionID.String()).
			FieldValue("subject", user.ID.String()).
			FieldValue("scopes", strings.Join(userScopes, ",")).
			Build(),
	)
	queries = append(
		queries,
		s.Valkey.B().
			Expire().
			Key(sessionKey).
			Seconds(int64(time.Duration(client.RefreshTokenValidDuration)*time.Second)).
			Build(),
	)
	multiRes := s.Valkey.DoMulti(ctx, queries...)
	for _, r := range multiRes {
		if r.Error() != nil {
			logger.PrintfError("Failed to store session: %v", r.Error())
			return nil, nil, []string{}, &errors.APIError{
				Code:    http.StatusInternalServerError,
				Error:   errors.InternalServerError,
				Details: "Failed to store session",
			}
		}
	}
	logger.PrintfDebug("Stored session with ID: %s", sessionID.String())

	destroyQuery := s.Valkey.B().Del().Key(key).Build()
	res = s.Valkey.Do(ctx, destroyQuery)
	if res.Error() != nil {
		logger.PrintfError("Failed to delete authorization code: %v", res.Error())
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

	query := s.Valkey.B().Hgetall().Key(sessionKey).Build()

	res := s.Valkey.Do(ctx, query)
	if res.Error() != nil {
		logger.PrintfError("Failed to get session: %v", res.Error())
		return nil, nil, []string{}, &errors.APIError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to get session",
		}
	}

	session, err := res.AsStrMap()
	if err != nil || len(session) == 0 {
		logger.PrintfWarning("Session not found: %s", refreshToken)
		return nil, nil, []string{}, &errors.APIError{
			Code:    http.StatusBadRequest,
			Error:   errors.InvalidRefreshToken,
			Details: "Invalid refresh token",
		}
	}

	// TODO: Add check for changed session scopes and if so reprompt for login
	sessionScopes := strings.Split(session["scopes"], ",")
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

	newSessionKey := fmt.Sprintf("session:%s", session["sessionID"])
	createQuery := s.Valkey.B().Set().Key(newSessionKey).Value(session["sessionID"]).Build()
	res = s.Valkey.Do(ctx, createQuery)
	if res.Error() != nil {
		logger.PrintfError("Failed to store new session: %v", res.Error())
		return nil, nil, []string{}, &errors.APIError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to store new session",
		}
	}
	logger.PrintfDebug("Stored new session with ID: %s", session["sessionID"])
	createQuery = s.Valkey.B().
		Expire().
		Key(newSessionKey).
		Seconds(int64(time.Duration(client.RefreshTokenValidDuration) * time.Second)).
		Build()
	res = s.Valkey.Do(ctx, createQuery)
	if res.Error() != nil {
		logger.PrintfError("Failed to set expiration for new session: %v", res.Error())
		return nil, nil, []string{}, &errors.APIError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to set expiration for new session",
		}
	}

	createQuery = s.Valkey.B().Del().Key(sessionKey).Build()
	res = s.Valkey.Do(ctx, createQuery)
	if res.Error() != nil {
		logger.PrintfError("Failed to delete old session: %v", res.Error())
	} else {
		logger.PrintfDebug("Deleted old session: %s", refreshToken)
	}

	return &accessToken, &newRefreshToken, sessionScopes, nil
}
