package middleware

import (
	"crypto/ed25519"
	"easyflow-oauth2-server/internal/shared/errors"
	"easyflow-oauth2-server/pkg/config"
	"easyflow-oauth2-server/pkg/logger"
	"easyflow-oauth2-server/pkg/tokens"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// SessionTokenMiddleware is a Gin middleware that checks for a valid session token in the cookies.
func SessionTokenMiddleware(cfg *config.Config, key *ed25519.PrivateKey) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.NewLogger(os.Stdout, "SessionTokenMiddleware", cfg.LogLevel, c.ClientIP())
		// Get session token from cookies
		sessionToken, err := c.Cookie(cfg.SessionCookieName)
		if err != nil {
			log.PrintfDebug("Error while getting session token cookie: %s", err.Error())
			c.Redirect(
				http.StatusSeeOther,
				cfg.FrontendURL+"/login?next="+c.Request.URL.Path+"&error="+string(
					errors.MissingSessionToken,
				),
			)
			c.Abort()
			return
		}

		if sessionToken == "" {
			log.PrintfDebug("No session token provided")
			c.Redirect(
				http.StatusSeeOther,
				cfg.FrontendURL+"/login?next="+c.Request.URL.Path+"&error="+string(
					errors.InvalidSessionToken,
				),
			)
			c.Abort()
			return
		}

		payload, err := tokens.ValidateJwt(key, sessionToken)
		if err != nil {
			log.PrintfDebug("Error validating session token: %s", err.Error())
			c.Redirect(
				http.StatusSeeOther,
				cfg.FrontendURL+"/login?next="+c.Request.URL.Path+"&error="+string(
					errors.InvalidSessionToken,
				),
			)
			c.Abort()
			return
		}

		if payload.Type != tokens.SessionToken {
			log.PrintfDebug("Invalid session token type: %s", payload.Type)
			c.Redirect(
				http.StatusSeeOther,
				cfg.FrontendURL+"/login?next="+c.Request.URL.Path+"&error="+string(
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
