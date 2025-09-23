package oauth

import (
	"easyflow-oauth2-server/pkg/api/errors"
	"easyflow-oauth2-server/pkg/api/middleware"
	"easyflow-oauth2-server/pkg/database"
	"easyflow-oauth2-server/pkg/endpoint"
	"net/http"
	"net/url"
	"slices"

	"github.com/gin-gonic/gin"
)

func RegisterOAuthEndpoints(r *gin.RouterGroup) {
	r.Use(middleware.LoggerMiddleware("OAuth"))

	r.GET("/authorize", middleware.SessionTokenMiddleware(), authorizeController)
	r.POST("/token", tokenController)

}

func redirectWithError(c *gin.Context, redirectURI *url.URL, errorCode, errorDescription, state string) {
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
		errors.SendErrorResponse(c, http.StatusBadRequest, errors.MissingClientID, "The client_id query parameter is required")
		return
	}

	client, err := getClient(utils, clientID)
	if err != nil {
		c.JSON(err.Code, err)
		return
	}

	// Extract and validate redirect_uri early
	redirectUri, ok := c.GetQuery("redirect_uri")
	if (!ok || redirectUri == "") && len(client.RedirectUris) > 1 {
		errors.SendErrorResponse(c, http.StatusBadRequest, errors.MissingRedirectURI, "The redirect_uri query parameter is required")
		return
	}

	// Use first registered redirect URI if none provided
	if !ok || redirectUri == "" {
		redirectUri = client.RedirectUris[0]
	} else if !slices.Contains(client.RedirectUris, redirectUri) {
		errors.SendErrorResponse(c, http.StatusBadRequest, errors.InvalidRedirectURI, "The provided redirect_uri is not registered for this client")
		return
	}

	uri, parseErr := url.ParseRequestURI(redirectUri)
	if parseErr != nil {
		errors.SendErrorResponse(c, http.StatusBadRequest, errors.InvalidRedirectURI, "The provided redirect_uri is not a valid URI")
		return
	}

	state := c.Query("state")
	if state == "" {
		redirectWithError(c, uri, "invalid_request", "The state query parameter is required", "")
		return
	}
	if len(state) > 255 {
		redirectWithError(c, uri, "invalid_request", "The state query parameter must not exceed 255 characters", state)
		return
	}

	// check if client fully supports PKCE
	if !client.IsPublic.Valid || !client.IsPublic.Bool {
		redirectWithError(c, uri, "unauthorized_client", "The client is not a public client and cannot use PKCE", state)
		return
	}

	responseType, ok := c.GetQuery("response_type")
	if !ok || responseType == "" {
		redirectWithError(c, uri, "invalid_request", "The response_type query parameter is required", state)
		return
	}

	if responseType != "code" {
		redirectWithError(c, uri, "unsupported_response_type", "The /oauth/authorize endpoint only supports the 'code' response type", state)
		return
	}

	codeChallenge, ok := c.GetQuery("code_challenge")
	if !ok || codeChallenge == "" {
		redirectWithError(c, uri, "invalid_request", "The code_challenge query parameter is required", state)
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
		errors.SendErrorResponse(c, http.StatusBadRequest, errors.InvalidContentType, "The Content-Type header must be application/x-www-form-urlencoded")
		return
	}

	if err := c.Request.ParseForm(); err != nil {
		errors.SendErrorResponse(c, http.StatusBadRequest, errors.InvalidRequestBody, "Failed to parse request body")
		return
	}

	grantType := c.Request.FormValue("grant_type")
	if grantType == "" {
		errors.SendErrorResponse(c, http.StatusBadRequest, errors.MissingGrantType, "The grant_type parameter is required")
		return
	}

	switch grantType {
	case "authorization_code":
		clientID := c.Request.FormValue("client_id")
		if clientID == "" {
			errors.SendErrorResponse(c, http.StatusBadRequest, errors.MissingClientID, "The client_id parameter is required")
			return
		}
		client, err := getClient(utils, clientID)
		if err != nil {
			c.JSON(err.Code, err)
			return
		}
		if !client.IsPublic.Valid || !client.IsPublic.Bool {
			errors.SendErrorResponse(c, http.StatusBadRequest, errors.InvalidClientID, "The client is not a public client and cannot use PKCE")
			return
		}
		if !slices.Contains(client.GrantTypes, database.GrantTypesAuthorizationCode) {
			errors.SendErrorResponse(c, http.StatusBadRequest, errors.InvalidClientID, "The client is not authorized to use the authorization_code grant type")
		}

		code := c.Request.FormValue("code")
		if code == "" {
			errors.SendErrorResponse(c, http.StatusBadRequest, errors.MissingCode, "The code parameter is required")
			return
		}
		codeVerifier := c.Request.FormValue("code_verifier")
		if codeVerifier == "" {
			errors.SendErrorResponse(c, http.StatusBadRequest, errors.MissingCodeVerifier, "The code_verifier parameter is required")
			return
		}

		accessToken, refreshToken, scopes, err := authorizationCodeFlow(utils, client, code, codeVerifier)
		if err != nil {
			c.JSON(err.Code, err)
			return
		}

		c.JSON(http.StatusOK, AuthorizationCodeTokenResponse{
			AccessToken:           *accessToken,
			AccessTokenExpiresIn:  int(client.AccessTokenValidDuration),
			RefreshToken:          *refreshToken,
			RefreshTokenExpiresIn: int(client.RefreshTokenValidDuration),
			Scopes:                scopes,
		})

	case "client_credentials":
		c.JSON(http.StatusNotImplemented, gin.H{"message": "client_credentials grant type is not implemented yet"})
	case "refresh_token":
		c.JSON(http.StatusNotImplemented, gin.H{"message": "refresh_token grant type is not implemented yet"})
	default:
		errors.SendErrorResponse(c, http.StatusBadRequest, errors.InvalidGrantType, "The grant_type is not supported")
	}
}
