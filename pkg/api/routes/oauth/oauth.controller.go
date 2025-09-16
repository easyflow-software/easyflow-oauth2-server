package oauth

import (
	"easyflow-oauth2-server/pkg/api/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterOAuthEndpoints(r *gin.RouterGroup) {
	r.Use(middleware.LoggerMiddleware("OAuth"))

}
