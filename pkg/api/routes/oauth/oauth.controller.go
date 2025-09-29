package oauth

import (
	"easyflow-oauth2-server/pkg/api/errors"
	"easyflow-oauth2-server/pkg/api/middleware"
	"easyflow-oauth2-server/pkg/database"
	"easyflow-oauth2-server/pkg/endpoint"
	"easyflow-oauth2-server/pkg/tokens"
	"net/http"
	"net/url"
	"slices"

	"github.com/gin-gonic/gin"
)

// RegisterOAuthEndpoints sets up the OAuth2-related endpoints.
func RegisterOAuthEndpoints(r *gin.RouterGroup) {
	r.Use(middleware.LoggerMiddleware("OAuth"))

	r.GET("/authorize", middleware.SessionTokenMiddleware(), authorizeController)
	r.POST("/token", tokenController)

}

func redirectWithError(
	c *gin.Context,
	redirectURI *url.URL,
	errorCode, errorDescription, state string,
) {
	q := redirectURI.Query()
	q.Add("error", errorCode)
	if errorDescription != "" {
		q.Add("error_description", errorDescription)
	}
	if state != "" {
		q.Add("state", state)
	}
	redirectURI.RawQuery = q.Encode()
	c.Redirect(http.StatusFound, redirectURI.String())
}

func authorizeController(c *gin.Context) {
	utils, errs := endpoint.SetupEndpoint[any](c)
	if len(errs) > 0 {
		endpoint.SendSetupErrorResponse(c, errs)
		return
	}

	// Extract client_id first to get client info for redirect_uri validation
	clientID, ok := c.GetQuery("client_id")
	if !ok || clientID == "" {
		errors.SendErrorResponse(
			c,
			http.StatusBadRequest,
			errors.MissingClientID,
			"The client_id query parameter is required",
		)
		return
	}

	client, err := getClient(utils, clientID)
	if err != nil {
		c.JSON(err.Code, err)
		return
	}

	// Extract and validate redirect_uri early
	redirectURI, ok := c.GetQuery("redirect_uri")
	if (!ok || redirectURI == "") && len(client.RedirectUris) > 1 {
		errors.SendErrorResponse(
			c,
			http.StatusBadRequest,
			errors.MissingRedirectURI,
			"The redirect_uri query parameter is required",
		)
		return
	}

	// Use first registered redirect URI if none provided
	if !ok || redirectURI == "" {
		redirectURI = client.RedirectUris[0]
	} else if !slices.Contains(client.RedirectUris, redirectURI) {
		errors.SendErrorResponse(c, http.StatusBadRequest, errors.InvalidRedirectURI, "The provided redirect_uri is not registered for this client")
		return
	}

	uri, parseErr := url.ParseRequestURI(redirectURI)
	if parseErr != nil {
		errors.SendErrorResponse(
			c,
			http.StatusBadRequest,
			errors.InvalidRedirectURI,
			"The provided redirect_uri is not a valid URI",
		)
		return
	}

	state := c.Query("state")
	if state == "" {
		redirectWithError(c, uri, "invalid_request", "The state query parameter is required", "")
		return
	}
	if len(state) > 255 {
		redirectWithError(
			c,
			uri,
			"invalid_request",
			"The state query parameter must not exceed 255 characters",
			state,
		)
		return
	}

	responseType, ok := c.GetQuery("response_type")
	if !ok || responseType == "" {
		redirectWithError(
			c,
			uri,
			"invalid_request",
			"The response_type query parameter is required",
			state,
		)
		return
	}

	if responseType != "code" {
		redirectWithError(
			c,
			uri,
			"unsupported_response_type",
			"The /oauth/authorize endpoint only supports the 'code' response type",
			state,
		)
		return
	}

	codeChallenge, ok := c.GetQuery("code_challenge")
	if !ok || codeChallenge == "" {
		redirectWithError(
			c,
			uri,
			"invalid_request",
			"The code_challenge query parameter is required",
			state,
		)
		return
	}

	code, authErr := authorize(utils, client, codeChallenge)
	if authErr != nil {
		redirectWithError(c, uri, "server_error", "", state)
		return
	}

	q := uri.Query()
	q.Add("code", *code)
	q.Add("state", state)
	uri.RawQuery = q.Encode()

	c.Redirect(http.StatusFound, uri.String())
}

func tokenController(c *gin.Context) {
	utils, errs := endpoint.SetupEndpoint[any](c)
	if len(errs) > 0 {
		endpoint.SendSetupErrorResponse(c, errs)
		return
	}

	contentType := c.GetHeader("Content-Type")
	if contentType != "application/x-www-form-urlencoded" {
		errors.SendErrorResponse(
			c,
			http.StatusBadRequest,
			errors.InvalidContentType,
			"The Content-Type header must be application/x-www-form-urlencoded",
		)
		return
	}

	if err := c.Request.ParseForm(); err != nil {
		errors.SendErrorResponse(
			c,
			http.StatusBadRequest,
			errors.InvalidRequestBody,
			"Failed to parse request body",
		)
		return
	}

	grantType := c.Request.FormValue("grant_type")
	if grantType == "" {
		errors.SendErrorResponse(
			c,
			http.StatusBadRequest,
			errors.MissingGrantType,
			"The grant_type parameter is required",
		)
		return
	}
	var clientSecret = ""
	clientID := c.Request.FormValue("client_id")
	if clientID == "" {
		// Try to get client_id and secret from Basic Auth
		var ok bool
		clientID, clientSecret, ok = c.Request.BasicAuth()
		if !ok {
			errors.SendErrorResponse(
				c,
				http.StatusBadRequest,
				errors.MissingClientID,
				"Basic auth header is required if client_id is not provided in the body or request is confidential",
			)
		}
		utils.Logger.PrintfDebug("%s", clientID)
		if clientID == "" {
			errors.SendErrorResponse(
				c,
				http.StatusBadRequest,
				errors.MissingClientID,
				"The client_id parameter is required",
			)
			return
		}
	}
	client, err := getClient(utils, clientID)
	if err != nil {
		c.JSON(err.Code, err)
		return
	}

	if !client.IsPublic {
		if clientSecret == "" {
			errors.SendErrorResponse(
				c,
				http.StatusBadRequest,
				errors.MissingClientSecret,
				"Client Secret is required for confidential clients",
			)
			return
		}

		if !client.ClientSecretHash.Valid ||
			!tokens.CompareClientSecretHash(clientSecret, client.ClientSecretHash.String) {
			errors.SendErrorResponse(
				c,
				http.StatusBadRequest,
				errors.InvalidClientSecret,
				"Invalid Client Secret",
			)
			return
		}
	}

	switch grantType {
	case "authorization_code":
		if !slices.Contains(client.GrantTypes, database.GrantTypesAuthorizationCode) {
			errors.SendErrorResponse(
				c,
				http.StatusBadRequest,
				errors.InvalidGrantType,
				"The client is not authorized to use the authorization_code grant type",
			)
		}

		code := c.Request.FormValue("code")
		if code == "" {
			errors.SendErrorResponse(
				c,
				http.StatusBadRequest,
				errors.MissingCode,
				"The code parameter is required",
			)
			return
		}
		codeVerifier := c.Request.FormValue("code_verifier")
		if codeVerifier == "" {
			errors.SendErrorResponse(
				c,
				http.StatusBadRequest,
				errors.MissingCodeVerifier,
				"The code_verifier parameter is required",
			)
			return
		}

		accessToken, refreshToken, scopes, err := authorizationCodeFlow(
			utils,
			client,
			code,
			codeVerifier,
		)
		if err != nil {
			c.JSON(err.Code, err)
			return
		}

		tokenRes := tokenResponse{
			AccessToken:           *accessToken,
			AccessTokenExpiresIn:  int(client.AccessTokenValidDuration),
			RefreshTokenExpiresIn: int(client.RefreshTokenValidDuration),
			Scopes:                scopes,
		}

		if slices.Contains(client.GrantTypes, database.GrantTypesRefreshToken) {
			tokenRes.RefreshToken = *refreshToken
		}

		c.JSON(http.StatusOK, tokenRes)

	case "client_credentials":
		if !slices.Contains(client.GrantTypes, database.GrantTypesClientCredentials) {
			errors.SendErrorResponse(
				c,
				http.StatusBadRequest,
				errors.InvalidGrantType,
				"The client is not authorized to use the client_credentials grant type",
			)
		}

		accessToken, scopes, err := clientCredentialsFlow(utils, client)
		if err != nil {
			c.JSON(err.Code, err)
			return
		}

		c.JSON(http.StatusOK, tokenResponse{
			AccessToken:          *accessToken,
			AccessTokenExpiresIn: int(client.AccessTokenValidDuration),
			Scopes:               scopes,
		})
	case "refresh_token":
		if !slices.Contains(client.GrantTypes, database.GrantTypesRefreshToken) {
			errors.SendErrorResponse(
				c,
				http.StatusBadRequest,
				errors.InvalidGrantType,
				"The client is not authorized to use the refresh_token grant type",
			)
		}

		refreshToken := c.Request.FormValue("refresh_token")
		if refreshToken == "" {
			errors.SendErrorResponse(
				c,
				http.StatusBadRequest,
				errors.MissingRefreshToken,
				"The refresh_token parameter is required",
			)
			return
		}

		newAccessToken, newRefreshToken, scopes, err := refreshTokenFlow(
			utils,
			client,
			refreshToken,
		)
		if err != nil {
			c.JSON(err.Code, err)
			return
		}

		c.JSON(http.StatusOK, tokenResponse{
			AccessToken:           *newAccessToken,
			AccessTokenExpiresIn:  int(client.AccessTokenValidDuration),
			RefreshToken:          *newRefreshToken,
			RefreshTokenExpiresIn: int(client.RefreshTokenValidDuration),
			Scopes:                scopes,
		})

	default:
		errors.SendErrorResponse(
			c,
			http.StatusBadRequest,
			errors.InvalidGrantType,
			"The grant_type is not supported",
		)
	}
}
