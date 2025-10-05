package middleware

import (
	"database/sql"
	"easyflow-oauth2-server/pkg/database"

	"github.com/gin-gonic/gin"
)

// QueriesMiddleware injects the database queries into the Gin context.
func QueriesMiddleware(queries *database.Queries, db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("queries", queries)
		c.Set("db", db)
		c.Next()
	}
}
