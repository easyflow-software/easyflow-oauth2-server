package oauth

import (
	"easyflow-oauth2-server/pkg/api/errors"
	"easyflow-oauth2-server/pkg/api/middleware"
	"easyflow-oauth2-server/pkg/endpoint"
	"net/http"
	"net/url"
	"slices"

	"github.com/gin-gonic/gin"
)

func RegisterOAuthEndpoints(r *gin.RouterGroup) {
	r.Use(middleware.LoggerMiddleware("OAuth"))

	r.GET("/authorize", middleware.SessionTokenMiddleware(), authorizeController)

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
