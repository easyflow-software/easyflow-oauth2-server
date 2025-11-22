package middleware

import (
	"crypto/ed25519"
	"easyflow-oauth2-server/internal/server/config"
	"easyflow-oauth2-server/internal/tokens"
	"easyflow-oauth2-server/pkg/logger"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
)

func redirectToLogin(c *gin.Context, frontendURL string) {
	url, err := url.Parse(frontendURL + "/login")
	if err != nil {
		// This should never happen because the frontend URL is validated at startup
		panic("Invalid frontend URL")
	}
	q := url.Query()
	q.Set("next", c.Request.URL.String())
	url.RawQuery = q.Encode()

	c.Redirect(http.StatusFound, url.String())
	c.Abort()
}

// SessionTokenMiddleware is a Gin middleware that checks for a valid session token in the cookies.
func SessionTokenMiddleware(cfg *config.Config, key *ed25519.PrivateKey) gin.HandlerFunc {
	return func(c *gin.Context) {
		log := logger.NewLogger(os.Stdout, "SessionTokenMiddleware", cfg.LogLevel, c.ClientIP())
		// Get session token from cookies
		sessionToken, err := c.Cookie(cfg.SessionCookieName)
		if err != nil {
			redirectToLogin(c, cfg.FrontendURL)
			return
		}

		if sessionToken == "" {
			log.PrintfDebug("No session token provided")
			redirectToLogin(c, cfg.FrontendURL)
			return
		}

		payload, err := tokens.ValidateJwt(key, sessionToken)
		if err != nil {
			log.PrintfDebug("Error validating session token: %s", err.Error())
			redirectToLogin(c, cfg.FrontendURL)
			return
		}

		if payload.Type != tokens.SessionToken {
			log.PrintfDebug("Invalid session token type: %s", payload.Type)
			redirectToLogin(c, cfg.FrontendURL)
			return
		}

		c.Set("user", payload)
		c.Next()
	}
}
