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

	r.GET("/authorize", authorizeController)

}

func authorizeController(c *gin.Context) {
	utils, errs := endpoint.SetupEndpoint[any](c)
	if len(errs) > 0 {
		endpoint.SendSetupErrorResponse(c, errs)
		return
	}

	state, ok := c.GetQuery("state")
	if !ok || state == "" {
		errors.SendErrorResponse(c, http.StatusBadRequest, errors.MissingState, "The state query parameter is required")
		return
	}

	if len(state) > 255 {
		errors.SendErrorResponse(c, http.StatusBadRequest, errors.InvalidState, "The state query parameter must not exceed 255 characters")
		return
	}

	clientID, ok := c.GetQuery("client_id")
	if !ok || clientID == "" {
		errors.SendErrorResponse(c, http.StatusBadRequest, errors.MissingClientID, "The client_id query parameter is required")
		return
	}

	responseType, ok := c.GetQuery("response_type")
	if !ok || responseType == "" {
		errors.SendErrorResponse(c, http.StatusBadRequest, errors.MissingResponseType, "The response_type query parameter is required")
		return
	}

	if responseType != "code" {
		errors.SendErrorResponse(c, http.StatusBadRequest, errors.UnsupportedResponseType, "The /oauth/authorize endpoint only supports the 'code' response type")
		return
	}

	codeChallenge, ok := c.GetQuery("code_challenge")
	if !ok || codeChallenge == "" {
		errors.SendErrorResponse(c, http.StatusBadRequest, errors.MissingCodeChallenge, "The code_challenge query parameter is required")
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

	redirectUri, ok := c.GetQuery("redirect_uri")
	if (!ok || redirectUri == "") && len(client.RedirectUris) > 1 {
		errors.SendErrorResponse(c, http.StatusBadRequest, errors.MissingRedirectURI, "The redirect_uri query parameter is required")
		return
	}

	uri, error := url.ParseRequestURI(redirectUri)
	if error != nil {
		errors.SendErrorResponse(c, http.StatusBadRequest, errors.InvalidRedirectURI, "The provided redirect_uri is not a valid URI")
		return
	}

	if ok && !slices.Contains(client.RedirectUris, redirectUri) {
		errors.SendErrorResponse(c, http.StatusBadRequest, errors.InvalidRedirectURI, "The provided redirect_uri is not registered for this client")
	}

	code, err := authorize(utils, client, codeChallenge)
	if err != nil {
		c.JSON(err.Code, err)
		return
	}
	
	q := uri.Query()
	
	q.Add("code", *code)
	q.Add("state", state)

	uri.RawQuery = q.Encode()

	c.Redirect(http.StatusFound, uri.String())
}
