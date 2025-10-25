// Package auth implements authentication-related endpoints.
package auth

import (
	"easyflow-oauth2-server/internal/server/config"
	"easyflow-oauth2-server/internal/shared/endpoint"
	_ "easyflow-oauth2-server/internal/shared/errors" // imported for swagger documentation
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
// @Summary Register a new user
// @Description Create a new user account with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body CreateUserRequest true "User registration details"
// @Success 201 {object} CreateUserResponse "User successfully created"
// @Failure 400 {object} errors.APIError "Invalid request payload"
// @Failure 409 {object} errors.APIError "Email already exists"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Router /auth/register [post].
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
// @Summary User login
// @Description Authenticate a user and create a session
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse "Login successful, session token set in cookie"
// @Failure 400 {object} errors.APIError "Invalid request payload"
// @Failure 401 {object} errors.APIError "Invalid credentials"
// @Failure 500 {object} errors.APIError "Internal server error"
// @Router /auth/login [post].
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
// @Summary User logout
// @Description Log out the current user and clear session
// @Tags Authentication
// @Accept json
// @Produce json
// @Security SessionToken
// @Success 204 "Logout successful"
// @Failure 401 {object} errors.APIError "Unauthorized"
// @Router /auth/logout [delete].
func (ctrl *Controller) Logout(c *gin.Context) {
	_, errs := endpoint.SetupEndpoint[any](c, endpoint.WithoutBody())
	if len(errs) > 0 {
		endpoint.SendSetupErrorResponse(c, errs)
		return
	}

	c.SetCookie(
		ctrl.service.Config.SessionCookieName,
		"",
		-1,
		"/",
		ctrl.service.Config.Domain,
		true,
		ctrl.service.Config.Environment == config.Production,
	)

	c.Status(http.StatusNoContent)
}
