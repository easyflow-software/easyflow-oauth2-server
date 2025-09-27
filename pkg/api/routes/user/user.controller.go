package user

import (
	"easyflow-oauth2-server/pkg/api/middleware"

	"github.com/gin-gonic/gin"
)

// RegisterUserEndpoints sets up the user-related endpoints.
func RegisterUserEndpoints(r *gin.RouterGroup) {
	r.Use(middleware.LoggerMiddleware("User"))

}
