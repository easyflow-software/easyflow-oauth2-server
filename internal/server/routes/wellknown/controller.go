// Package wellknown implements the .well-known routes
package wellknown

import (
	"easyflow-oauth2-server/internal/shared/endpoint"
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
func (ctrl *Controller) GetJWKS(c *gin.Context) {
	_, errs := endpoint.SetupEndpoint[any](c, endpoint.WithoutBody())
	if len(errs) > 0 {
		endpoint.SendSetupErrorResponse(c, errs)
		return
	}

	jwks := ctrl.service.GetJWKS(c.ClientIP())

	c.JSON(http.StatusOK, jwks)
}
