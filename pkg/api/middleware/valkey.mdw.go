package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/valkey-io/valkey-go"
)

// ValkeyMiddleware injects the Valkey client into the Gin context for use in handlers.
func ValkeyMiddleware(valkeyClient valkey.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("valkey", valkeyClient)
		c.Next()
	}
}
