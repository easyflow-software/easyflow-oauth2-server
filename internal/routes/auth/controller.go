// Package auth implements authentication-related endpoints.
package auth

import (
	"easyflow-oauth2-server/internal/shared/endpoint"
	"easyflow-oauth2-server/pkg/config"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Controller handles authentication HTTP requests.
type Controller struct {
	service *Service
}

// NewAuthController creates a new instance of AuthController.
func NewAuthController(service *Service) *Controller {
	return &Controller{
		service: service,
	}
}

// RegisterRoutes sets up the authentication-related endpoints.
func (ctrl *Controller) RegisterRoutes(r *gin.RouterGroup) {
	r.POST("/register", ctrl.Register)
	r.POST("/login", ctrl.Login)
	r.DELETE("/logout", ctrl.Logout)
}

// Register handles user registration.
func (ctrl *Controller) Register(c *gin.Context) {
	utils, errs := endpoint.SetupEndpoint[CreateUserRequest](c)
	if len(errs) > 0 {
		endpoint.SendSetupErrorResponse(c, errs)
		return
	}

	user, err := ctrl.service.Register(c.Request.Context(), utils.Payload, c.ClientIP())
	if err != nil {
		c.JSON(err.Code, err)
		return
	}

	c.JSON(http.StatusCreated, user)
}

// Login handles user authentication.
func (ctrl *Controller) Login(c *gin.Context) {
	utils, errs := endpoint.SetupEndpoint[LoginRequest](c)
	if len(errs) > 0 {
		endpoint.SendSetupErrorResponse(c, errs)
		return
	}

	login, err := ctrl.service.Login(c.Request.Context(), utils.Payload, c.ClientIP())
	if err != nil {
		c.JSON(err.Code, err)
		return
	}

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		ctrl.service.Config.SessionCookieName,
		login.SessionToken,
		int(
			time.Duration(ctrl.service.Config.JwtSessionTokenExpiryHours)*time.Hour,
		)/int(
			time.Second,
		),
		"/",
		ctrl.service.Config.Domain,
		true,
		ctrl.service.Config.Environment == config.Production,
	)

	c.JSON(http.StatusOK, login)
}

// Logout handles user logout.
func (ctrl *Controller) Logout(c *gin.Context) {
	_, errs := endpoint.SetupEndpoint[any](c, endpoint.WithoutBody())
	if len(errs) > 0 {
		endpoint.SendSetupErrorResponse(c, errs)
		return
	}

	c.SetCookie(
		ctrl.service.Config.SessionCookieName,
		"value string",
		-1,
		"/",
		ctrl.service.Config.Domain,
		true,
		ctrl.service.Config.Environment == config.Production,
	)

	c.Status(http.StatusNoContent)
}
