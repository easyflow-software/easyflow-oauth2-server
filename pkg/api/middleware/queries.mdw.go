package middleware

import (
	"easyflow-oauth2-server/pkg/database"

	"github.com/gin-gonic/gin"
)

// QueriesMiddleware injects the database queries into the Gin context.
func QueriesMiddleware(queries *database.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("queries", queries)
		c.Next()
	}
}
