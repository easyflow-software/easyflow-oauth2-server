package oauth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"easyflow-oauth2-server/pkg/api/errors"
	"easyflow-oauth2-server/pkg/database"
	"easyflow-oauth2-server/pkg/endpoint"
	"easyflow-oauth2-server/pkg/scopes"
	"easyflow-oauth2-server/pkg/tokens"
	"encoding/hex"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/valkey-io/valkey-go"
)

func getClient(utils endpoint.EndpointUtils[any], clientId string) (*database.GetOAuthClientByClientIDRow, *errors.ApiError) {
	client, err := utils.Queries.GetOAuthClientByClientID(utils.RequestContext, clientId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &errors.ApiError{
				Code:    http.StatusNotFound,
				Error:   errors.InvalidClientID,
				Details: "Client not found",
			}
		}
		return nil, &errors.ApiError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to get client" + err.Error(),
		}
	}

	return &client, nil

}

func authorize(utils endpoint.EndpointUtils[any], client *database.GetOAuthClientByClientIDRow, codeChallange string) (*string, *errors.ApiError) {
	code := rand.Text()

	key := fmt.Sprintf("authorization-code:%s", code)

	var queries valkey.Commands

	queries = append(queries, utils.Valkey.B().Hset().Key(key).FieldValue().FieldValue("codeChallange", codeChallange).FieldValue("clientId", client.ClientID).FieldValue("userId", utils.User.Subject).FieldValue("scopes", strings.Join(client.Scopes, " ")).Build())
	queries = append(queries, utils.Valkey.B().Expire().Key(key).Seconds(600).Build())
	res := utils.Valkey.DoMulti(utils.RequestContext, queries...)
	if slices.ContainsFunc(res, func(r valkey.ValkeyResult) bool {
		return r.Error() != nil
	}) {
		return nil, &errors.ApiError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to store authorization code",
		}
	}
	return &code, nil
}

func authorizationCodeFlow(utils endpoint.EndpointUtils[any], client *database.GetOAuthClientByClientIDRow, code, codeVerifier string) (*string, *string, []string, *errors.ApiError) {
	key := fmt.Sprintf("authorization-code:%s", code)

	query := utils.Valkey.B().Hgetall().Key(key).Build()

	res := utils.Valkey.Do(utils.RequestContext, query)
	if res.Error() != nil {
		utils.Logger.PrintfError("Failed to get authorization code: %v", res.Error())
		return nil, nil, []string{}, &errors.ApiError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to get authorization code",
		}
	}

	codeStore, err := res.AsStrMap()
	if err != nil {
		utils.Logger.PrintfError("Failed to parse authorization code: %v", err)
		return nil, nil, []string{}, &errors.ApiError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to parse authorization code",
		}
	}

	if len(codeStore) == 0 {
		utils.Logger.PrintfWarning("Authorization code not found: %s", code)
		return nil, nil, []string{}, &errors.ApiError{
			Code:    http.StatusBadRequest,
			Error:   errors.InvalidCode,
			Details: "Invalid authorization code",
		}
	}

	if codeStore["clientId"] != client.ClientID {
		utils.Logger.PrintfWarning("Client ID does not match authorization code: %s", code)
		return nil, nil, []string{}, &errors.ApiError{
			Code:    http.StatusBadRequest,
			Error:   errors.InvalidClientID,
			Details: "Client ID does not match authorization code",
		}
	}

	hash := sha256.Sum256([]byte(codeVerifier))

	hashStr := hex.EncodeToString(hash[:])

	if codeStore["codeChallange"] != hashStr {
		utils.Logger.PrintfWarning("Code verifier does not match code challange: %s", code)
		return nil, nil, []string{}, &errors.ApiError{
			Code:    http.StatusBadRequest,
			Error:   errors.InvalidCodeVerifier,
			Details: "Invalid code verifier",
		}
	}

	ID, err := uuid.Parse(codeStore["userId"])
	if err != nil {
		utils.Logger.PrintfError("Failed to parse user ID: %v", err)
		return nil, nil, []string{}, &errors.ApiError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to parse user ID",
		}
	}

	user, err := utils.Queries.GetUserWithRolesAndScopes(utils.RequestContext, ID)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.Logger.PrintfWarning("User not found: %s", codeStore["userId"])
			return nil, nil, []string{}, &errors.ApiError{
				Code:    http.StatusNotFound,
				Error:   errors.NotFound,
				Details: "User not found",
			}
		}
		utils.Logger.PrintfError("Failed to get user: %v", err)
		return nil, nil, []string{}, &errors.ApiError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to get user",
		}
	}
	utils.Logger.PrintfDebug("Found user with ID: %s", user.ID)

	scopes := scopes.FilterScopes(user.Scopes, client.Scopes)

	sessionID := uuid.New()

	accessToken, refreshToken, err := tokens.GenerateTokens(utils.Config, utils.Key, user.ID.String(), client, scopes, sessionID.String())
	if err != nil {
		utils.Logger.PrintfError("Failed to generate tokens: %v", err)
		return nil, nil, []string{}, &errors.ApiError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to generate tokens",
		}
	}

	var queries valkey.Commands
	sessionKey := fmt.Sprintf("session:%s", sessionID.String())
	queries = append(queries, utils.Valkey.B().Set().Key(sessionKey).Value(sessionID.String()).Build())
	queries = append(queries, utils.Valkey.B().Expire().Key(sessionKey).Seconds(int64(time.Duration(utils.Config.JwtRefreshTokenExpiryDays)*24*time.Hour/time.Second)).Build())
	multiRes := utils.Valkey.DoMulti(utils.RequestContext, queries...)
	for _, r := range multiRes {
		if r.Error() != nil {
			utils.Logger.PrintfError("Failed to store session: %v", r.Error())
			return nil, nil, []string{}, &errors.ApiError{
				Code:    http.StatusInternalServerError,
				Error:   errors.InternalServerError,
				Details: "Failed to store session",
			}
		}
	}
	utils.Logger.PrintfDebug("Stored session with ID: %s", sessionID.String())

	query = utils.Valkey.B().Del().Key(key).Build()
	res = utils.Valkey.Do(utils.RequestContext, query)
	if res.Error() != nil {
		utils.Logger.PrintfError("Failed to delete authorization code: %v", res.Error())
	}
	utils.Logger.PrintfDebug("Deleted authorization code: %s", code)

	return &accessToken, &refreshToken, scopes, nil

}
