package oauth

import (
	"crypto/rand"
	"database/sql"
	"easyflow-oauth2-server/pkg/api/errors"
	"easyflow-oauth2-server/pkg/database"
	"easyflow-oauth2-server/pkg/endpoint"
	"fmt"
	"net/http"
	"slices"
	"strings"

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

	queries = append(queries, utils.Valkey.B().Hset().Key(key).FieldValue().FieldValue("codeChallange", codeChallange).FieldValue("clientId", client.ClientID).FieldValue("userId", utils.User.ID).FieldValue("scopes", strings.Join(client.Scopes, " ")).Build())
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
