// Package wellknown implements the .well-known routes
package wellknown

import (
	"easyflow-oauth2-server/internal/shared/endpoint"
	_ "easyflow-oauth2-server/internal/shared/errors" // imported for swagger documentation
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
)

// Controller handles well-known HTTP requests.
type Controller struct {
	service *Service
}

// ControllerParams holds dependencies for WellKnownController.
type ControllerParams struct {
	fx.In
	Service *Service
}

// NewWellKnownController creates a new instance of WellKnownController.
func NewWellKnownController(params ControllerParams) *Controller {
	return &Controller{
		service: params.Service,
	}
}

// RegisterRoutes sets up the .well-known endpoints.
func (ctrl *Controller) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/oauth-authorization-server", ctrl.GetOAuth2Metadata)
	r.GET("/openid-configuration", ctrl.GetOAuth2Metadata) // Alias for compatibility
	r.GET("/jwks.json", ctrl.GetJWKS)
}

// GetOAuth2Metadata handles the OAuth 2.0 Authorization Server Metadata endpoint.
// This implements RFC 8414 - OAuth 2.0 Authorization Server Metadata.
// @Summary Get OAuth2 Authorization Server Metadata
// @Description Returns OAuth 2.0 Authorization Server Metadata following RFC 8414
// @Tags Well-Known
// @Accept json
// @Produce json
// @Success 200 {object} OAuth2Metadata "OAuth2 server metadata"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Router /.well-known/oauth-authorization-server [get]
// @Router /.well-known/openid-configuration [get].
func (ctrl *Controller) GetOAuth2Metadata(c *gin.Context) {
	_, errs := endpoint.SetupEndpoint[any](c, endpoint.WithoutBody())
	if len(errs) > 0 {
		endpoint.SendSetupErrorResponse(c, errs)
		return
	}

	metadata := ctrl.service.GetOAuth2Metadata(c.Request.Context(), c.ClientIP())

	c.JSON(http.StatusOK, metadata)
}

// GetJWKS handles the JSON Web Key Set endpoint.
// This implements RFC 7517 - JSON Web Key (JWK).
// @Summary Get JSON Web Key Set
// @Description Returns the JSON Web Key Set (JWKS) for token verification following RFC 7517
// @Tags Well-Known
// @Accept json
// @Produce json
// @Success 200 {object} JWKSet "JSON Web Key Set"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Router /.well-known/jwks.json [get].
func (ctrl *Controller) GetJWKS(c *gin.Context) {
	_, errs := endpoint.SetupEndpoint[any](c, endpoint.WithoutBody())
	if len(errs) > 0 {
		endpoint.SendSetupErrorResponse(c, errs)
		return
	}

	jwks := ctrl.service.GetJWKS(c.ClientIP())

	c.JSON(http.StatusOK, jwks)
}
