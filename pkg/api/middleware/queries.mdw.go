package middleware

import (
	"easyflow-oauth2-server/pkg/database"

	"github.com/gin-gonic/gin"
)

func QueriesMiddleware(queries *database.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("queries", queries)
		c.Next()
	}
}