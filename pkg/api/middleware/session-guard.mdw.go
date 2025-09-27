package middleware

import (
	"easyflow-oauth2-server/pkg/api/errors"
	"easyflow-oauth2-server/pkg/endpoint"
	"easyflow-oauth2-server/pkg/tokens"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SessionTokenMiddleware is a Gin middleware that checks for a valid session token in the cookies.
func SessionTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		utils, errs := endpoint.SetupEndpoint[any](c)
		if len(errs) > 0 {
			c.AbortWithStatusJSON(http.StatusInternalServerError, errors.APIError{
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
			c.Redirect(
				http.StatusSeeOther,
				utils.Config.FrontendURL+"/login?next="+c.Request.URL.Path+"&error="+string(
					errors.MissingSessionToken,
				),
			)
			c.Abort()
			return
		}

		if sessionToken == "" {
			utils.Logger.PrintfDebug("No session token provided")
			c.Redirect(
				http.StatusSeeOther,
				utils.Config.FrontendURL+"/login?next="+c.Request.URL.Path+"&error="+string(
					errors.InvalidSessionToken,
				),
			)
			c.Abort()
			return
		}

		payload, err := tokens.ValidateJwt(utils.Key, sessionToken)
		if err != nil {
			utils.Logger.PrintfDebug("Error validating session token: %s", err.Error())
			c.Redirect(
				http.StatusSeeOther,
				utils.Config.FrontendURL+"/login?next="+c.Request.URL.Path+"&error="+string(
					errors.InvalidSessionToken,
				),
			)
			c.Abort()
			return
		}

		if payload.Type != tokens.SessionToken {
			utils.Logger.PrintfDebug("Invalid session token type: %s", payload.Type)
			c.Redirect(
				http.StatusSeeOther,
				utils.Config.FrontendURL+"/login?next="+c.Request.URL.Path+"&error="+string(
					errors.InvalidSessionToken,
				),
			)
			c.Abort()
			return
		}

		c.Set("user", payload)
		c.Next()
	}
}
