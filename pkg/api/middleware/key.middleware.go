package middleware

import (
	"crypto/ed25519"

	"github.com/gin-gonic/gin"
)

// KeyMiddlware adds the application configuration to the Gin context.
// It stores the config in the context for access by subsequent handlers.
func KeyMiddlware(key *ed25519.PrivateKey) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("key", key)
		c.Next()
	}
}
