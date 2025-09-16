package auth

import (
	"easyflow-oauth2-server/pkg/api/errors"
	"easyflow-oauth2-server/pkg/api/middleware"
	"easyflow-oauth2-server/pkg/config"
	"easyflow-oauth2-server/pkg/endpoint"
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	SESSION_COOKIE_NAME = "session_token"
)

func RegisterAuthEnpoints(r *gin.RouterGroup) {
	r.Use(middleware.LoggerMiddleware("Auth"))
	r.POST("/register", registerController)
	r.POST("/login", loginController)
	r.DELETE("/logout", logoutController)
}

func registerController(c *gin.Context) {
	utils, errs := endpoint.SetupEndpoint[CreateUserRequestDTO](c)
	if len(errs) > 0 {
		c.JSON(http.StatusInternalServerError, errors.ApiError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to setup endpoint",
		})
		return
	}
	user, err := register(utils)
	if err != nil {
		c.JSON(err.Code, err)
		return
	}

	c.JSON(http.StatusCreated, user)
}

func loginController(c *gin.Context) {
	utils, errs := endpoint.SetupEndpoint[LoginRequestDTO](c)
	if len(errs) > 0 {
		c.JSON(http.StatusInternalServerError, errors.ApiError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to setup endpoint",
		})
		return
	}

	login, err := login(utils)

	if err != nil {
		c.JSON(err.Code, err)
		return
	}

	c.SetCookie(SESSION_COOKIE_NAME, login.SessionToken, utils.Config.JwtSessionTokenExpiryHours, "/", utils.Config.Domain, true, utils.Config.Environment == config.Production)

	c.JSON(http.StatusOK, login)
}

func logoutController(c *gin.Context) {
	utils, errs := endpoint.SetupEndpoint[any](c)
	if len(errs) > 0 {
		c.JSON(http.StatusInternalServerError, errors.ApiError{
			Code:    http.StatusInternalServerError,
			Error:   errors.InternalServerError,
			Details: "Failed to setup endpoint",
		})
		return
	}

	c.SetCookie(SESSION_COOKIE_NAME, "value string", -1, "/", utils.Config.Domain, true, utils.Config.Environment == config.Production)

	c.Status(http.StatusNoContent)
}
