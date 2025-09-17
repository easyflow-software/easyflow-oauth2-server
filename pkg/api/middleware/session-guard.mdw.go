package middleware

import (
	"easyflow-oauth2-server/pkg/api/errors"
	"easyflow-oauth2-server/pkg/endpoint"
	"easyflow-oauth2-server/pkg/tokens"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SessionTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		utils, errs := endpoint.SetupEndpoint[any](c)
		if len(errs) > 0 {
			c.AbortWithStatusJSON(http.StatusInternalServerError, errors.ApiError{
				Code:    http.StatusInternalServerError,
				Error:   errors.InternalServerError,
				Details: "Failed to setup endpoint",
			})
			return
		}

		// Get session token from cookies
		sessionToken, err := c.Cookie(utils.Config.SessionCookieName)
		if err != nil {
			utils.Logger.PrintfDebug("Error while getting session token cookie: %s", err.Error())
			c.AbortWithStatusJSON(http.StatusBadRequest, errors.ApiError{
				Code:  http.StatusBadRequest,
				Error: errors.InvalidSessionToken,
			})
			return
		}

		if sessionToken == "" {
			utils.Logger.PrintfDebug("No session token provided")
			c.AbortWithStatusJSON(http.StatusBadRequest, errors.ApiError{
				Code:  http.StatusBadRequest,
				Error: errors.InvalidSessionToken,
			})
			return
		}

		payload, err := tokens.ValidateJwt(utils.Key, sessionToken)
		if err != nil {
			utils.Logger.PrintfDebug("Error validating session token: %s", err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.ApiError{
				Code:    http.StatusUnauthorized,
				Error:   errors.InvalidSessionToken,
				Details: err.Error(),
			})
			return
		}

		if payload.Type != tokens.SessionToken {
			utils.Logger.PrintfDebug("Invalid session token type: %s", payload.Type)
			c.AbortWithStatusJSON(http.StatusUnauthorized, errors.ApiError{
				Code:    http.StatusBadRequest,
				Error:   errors.InvalidSessionToken,
				Details: "Cannot use other tokens than session tokens for this endpoint",
			})
			return
		}

		c.Set("user", payload)
		c.Next()
	}
}
